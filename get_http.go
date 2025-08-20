// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package getter

import (
	"context"
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/go-cleanhttp"
	safetemp "github.com/hashicorp/go-safetemp"
)

// HttpGetter is a Getter implementation that will download from an HTTP
// endpoint.
//
// For file downloads, HTTP is used directly.
//
// The protocol for downloading a directory from an HTTP endpoint is as follows:
//
// An HTTP GET request is made to the URL with the additional GET parameter
// "terraform-get=1". This lets you handle that scenario specially if you
// wish. The response must be a 2xx.
//
// First, a header is looked for "X-Terraform-Get" which should contain
// a source URL to download. This source must use one of the configured
// protocols and getters for the client, or "http"/"https" if using
// the HttpGetter directly.
//
// If the header is not present, then a meta tag is searched for named
// "terraform-get" and the content should be a source URL.
//
// The source URL, whether from the header or meta tag, must be a fully
// formed URL. The shorthand syntax of "github.com/foo/bar" or relative
// paths are not allowed.
type HttpGetter struct {
	getter

	// Netrc, if true, will lookup and use auth information found
	// in the user's netrc file if available.
	Netrc bool

	// Client is the http.Client to use for Get requests.
	// This defaults to a cleanhttp.DefaultClient if left unset.
	Client *http.Client

	// Header contains optional request header fields that should be included
	// with every HTTP request. Note that the zero value of this field is nil,
	// and as such it needs to be initialized before use, via something like
	// make(http.Header).
	Header http.Header

	// DoNotCheckHeadFirst configures the client to NOT check if the server
	// supports HEAD requests.
	DoNotCheckHeadFirst bool

	// HeadFirstTimeout configures the client to enforce a timeout when
	// the server supports HEAD requests.
	//
	// The zero value means no timeout.
	HeadFirstTimeout time.Duration

	// ReadTimeout configures the client to enforce a timeout when
	// making a request to an HTTP server and reading its response body.
	//
	// The zero value means no timeout.
	ReadTimeout time.Duration

	// MaxBytes limits the number of bytes that will be ready from an HTTP
	// response body returned from a server. The zero value means no limit.
	MaxBytes int64

	// XTerraformGetLimit configures how many times the client with follow
	// the " X-Terraform-Get" header value.
	//
	// The zero value means no limit.
	XTerraformGetLimit int

	// XTerraformGetDisabled disables the client's usage of the "X-Terraform-Get"
	// header value.
	XTerraformGetDisabled bool
}

func (g *HttpGetter) ClientMode(u *url.URL) (ClientMode, error) {
	if strings.HasSuffix(u.Path, "/") {
		return ClientModeDir, nil
	}
	return ClientModeFile, nil
}

type contextKey int

const (
	xTerraformGetDisable           contextKey = 0
	xTerraformGetLimit             contextKey = 1
	xTerraformGetLimitCurrentValue contextKey = 2
	httpClientValue                contextKey = 3
	httpMaxBytesValue              contextKey = 4
)

func xTerraformGetDisabled(ctx context.Context) bool {
	value, ok := ctx.Value(xTerraformGetDisable).(bool)
	if !ok {
		return false
	}
	return value
}

func xTerraformGetLimitCurrentValueFromContext(ctx context.Context) int {
	value, ok := ctx.Value(xTerraformGetLimitCurrentValue).(int)
	if !ok {
		return 1
	}
	return value
}

func xTerraformGetLimiConfiguredtFromContext(ctx context.Context) int {
	value, ok := ctx.Value(xTerraformGetLimit).(int)
	if !ok {
		return 0
	}
	return value
}

func httpClientFromContext(ctx context.Context) *http.Client {
	value, ok := ctx.Value(httpClientValue).(*http.Client)
	if !ok {
		return nil
	}
	return value
}

func httpMaxBytesFromContext(ctx context.Context) int64 {
	value, ok := ctx.Value(httpMaxBytesValue).(int64)
	if !ok {
		return 0 // no limit
	}
	return value
}

type limitedWrappedReaderCloser struct {
	underlying io.Reader
	closeFn    func() error
}

func (l *limitedWrappedReaderCloser) Read(p []byte) (n int, err error) {
	return l.underlying.Read(p)
}

func (l *limitedWrappedReaderCloser) Close() (err error) {
	return l.closeFn()
}

func newLimitedWrappedReaderCloser(r io.ReadCloser, limit int64) io.ReadCloser {
	return &limitedWrappedReaderCloser{
		underlying: io.LimitReader(r, limit),
		closeFn:    r.Close,
	}
}

func (g *HttpGetter) Get(dst string, u *url.URL) error {
	ctx := g.Context()

	// Log the start of the Get operation
	fmt.Printf("[DEBUG] HttpGetter.Get: Starting download from %s to %s\n", u.String(), dst)

	// Optionally disable any X-Terraform-Get redirects. This is reccomended for usage of
	// this client outside of Terraform's. This feature is likely not required if the
	// source server can provider normal HTTP redirects.
	if g.XTerraformGetDisabled {
		fmt.Printf("[DEBUG] HttpGetter.Get: X-Terraform-Get redirects disabled\n")
		ctx = context.WithValue(ctx, xTerraformGetDisable, g.XTerraformGetDisabled)
	}

	// Optionally enforce a limit on X-Terraform-Get redirects. We check this for every
	// invocation of this function, because the value is not passed down to subsequent
	// client Get function invocations.
	if g.XTerraformGetLimit > 0 {
		fmt.Printf("[DEBUG] HttpGetter.Get: X-Terraform-Get redirect limit set to %d\n", g.XTerraformGetLimit)
		ctx = context.WithValue(ctx, xTerraformGetLimit, g.XTerraformGetLimit)
	}

	// If there was a limit on X-Terraform-Get redirects, check what the current count value.
	//
	// If the value is greater than the limit, return an error. Otherwise, increment the value,
	// and include it in the the context to be passed along in all the subsequent client
	// Get function invocations.
	if limit := xTerraformGetLimiConfiguredtFromContext(ctx); limit > 0 {
		currentValue := xTerraformGetLimitCurrentValueFromContext(ctx)
		fmt.Printf("[DEBUG] HttpGetter.Get: Current X-Terraform-Get redirect count: %d, limit: %d\n", currentValue, limit)

		if currentValue > limit {
			fmt.Printf("[ERROR] HttpGetter.Get: Too many X-Terraform-Get redirects: %d\n", currentValue)
			return fmt.Errorf("too many X-Terraform-Get redirects: %d", currentValue)
		}

		currentValue++
		fmt.Printf("[DEBUG] HttpGetter.Get: Incrementing X-Terraform-Get redirect count to %d\n", currentValue)

		ctx = context.WithValue(ctx, xTerraformGetLimitCurrentValue, currentValue)
	}

	// Optionally enforce a maxiumum HTTP response body size.
	if g.MaxBytes > 0 {
		fmt.Printf("[DEBUG] HttpGetter.Get: Setting max response body size to %d bytes\n", g.MaxBytes)
		ctx = context.WithValue(ctx, httpMaxBytesValue, g.MaxBytes)
	}

	// Copy the URL so we can modify it
	newU := *u
	u = &newU
	fmt.Printf("[DEBUG] HttpGetter.Get: URL copied for modification\n")

	if g.Netrc {
		fmt.Printf("[DEBUG] HttpGetter.Get: Adding auth from netrc file\n")
		// Add auth from netrc if we can
		if err := addAuthFromNetrc(u); err != nil {
			fmt.Printf("[ERROR] HttpGetter.Get: Failed to add auth from netrc: %v\n", err)
			return err
		}
		fmt.Printf("[DEBUG] HttpGetter.Get: Successfully added auth from netrc\n")
	}

	fmt.Println("Part 1")

	// If the HTTP client is nil, check if there is one available in the context,
	// otherwise create one using cleanhttp's default transport.
	if g.Client == nil {
		fmt.Printf("[DEBUG] HttpGetter.Get: HTTP client is nil, checking context or creating new one\n")
		if client := httpClientFromContext(ctx); client != nil {
			fmt.Printf("[DEBUG] HttpGetter.Get: Using HTTP client from context\n")
			g.Client = client
		} else {
			fmt.Printf("[DEBUG] HttpGetter.Get: Creating new HTTP client\n")
			client := httpClient
			if g.client != nil && g.client.Insecure {
				fmt.Printf("[DEBUG] HttpGetter.Get: Configuring insecure transport (TLS verification disabled)\n")
				insecureTransport := cleanhttp.DefaultTransport()
				insecureTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
				client.Transport = insecureTransport
			}
			g.Client = client
		}
	} else {
		fmt.Printf("[DEBUG] HttpGetter.Get: Using existing HTTP client\n")
	}

	// Pass along the configured HTTP client in the context for usage with the X-Terraform-Get feature.
	ctx = context.WithValue(ctx, httpClientValue, g.Client)
	fmt.Printf("[DEBUG] HttpGetter.Get: HTTP client stored in context\n")

	// Add terraform-get to the parameter.
	q := u.Query()
	q.Add("terraform-get", "1")
	u.RawQuery = q.Encode()
	fmt.Printf("[DEBUG] HttpGetter.Get: Added terraform-get=1 parameter to URL: %s\n", u.String())

	readCtx := ctx

	if g.ReadTimeout > 0 {
		fmt.Printf("[DEBUG] HttpGetter.Get: Setting read timeout to %v\n", g.ReadTimeout)
		var cancel context.CancelFunc
		readCtx, cancel = context.WithTimeout(ctx, g.ReadTimeout)
		defer cancel()
	}

	fmt.Println("Part 2")
	// Get the URL
	fmt.Printf("[DEBUG] HttpGetter.Get: Creating HTTP GET request\n")
	req, err := http.NewRequestWithContext(readCtx, "GET", u.String(), nil)
	if err != nil {
		fmt.Printf("[ERROR] HttpGetter.Get: Failed to create HTTP request: %v\n", err)
		return err
	}

	if g.Header != nil {
		fmt.Printf("[DEBUG] HttpGetter.Get: Cloning custom headers to request\n")
		req.Header = g.Header.Clone()
	}

	fmt.Printf("[DEBUG] HttpGetter.Get: Executing HTTP request to %s\n", u.String())
	resp, err := g.Client.Do(req)
	if err != nil {
		fmt.Printf("[ERROR] HttpGetter.Get: HTTP request failed: %v\n", err)
		return err
	}
	defer func() {
		fmt.Printf("[DEBUG] HttpGetter.Get: Closing response body\n")
		_ = resp.Body.Close()
	}()

	fmt.Printf("[DEBUG] HttpGetter.Get: Received HTTP response with status code: %d\n", resp.StatusCode)

	body := resp.Body

	if maxBytes := httpMaxBytesFromContext(ctx); maxBytes > 0 {
		fmt.Printf("[DEBUG] HttpGetter.Get: Wrapping response body with %d byte limit\n", maxBytes)
		body = newLimitedWrappedReaderCloser(body, maxBytes)
	}

	fmt.Println("Part 3")
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Printf("[ERROR] HttpGetter.Get: Bad response code: %d\n", resp.StatusCode)
		return fmt.Errorf("bad response code: %d", resp.StatusCode)
	}

	if disabled := xTerraformGetDisabled(ctx); disabled {
		fmt.Printf("[DEBUG] HttpGetter.Get: X-Terraform-Get is disabled, returning early\n")
		return nil
	}

	// Extract the source URL,
	var source string
	if v := resp.Header.Get("X-Terraform-Get"); v != "" {
		fmt.Printf("[DEBUG] HttpGetter.Get: Found X-Terraform-Get header: %s\n", v)
		source = v
	} else {
		fmt.Printf("[DEBUG] HttpGetter.Get: No X-Terraform-Get header found, parsing meta tags\n")
		source, err = g.parseMeta(readCtx, body)
		if err != nil {
			fmt.Printf("[ERROR] HttpGetter.Get: Failed to parse meta tags: %v\n", err)
			return err
		}
		if source != "" {
			fmt.Printf("[DEBUG] HttpGetter.Get: Found terraform-get meta tag: %s\n", source)
		}
	}
	if source == "" {
		fmt.Printf("[ERROR] HttpGetter.Get: No source URL found in headers or meta tags\n")
		return fmt.Errorf("no source URL was returned")
	}

	// If there is a subdir component, then we download the root separately
	// into a temporary directory, then copy over the proper subdir.
	source, subDir := SourceDirSubdir(source)
	if subDir != "" {
		fmt.Printf("[DEBUG] HttpGetter.Get: Detected subdirectory: %s, main source: %s\n", subDir, source)
	} else {
		fmt.Printf("[DEBUG] HttpGetter.Get: No subdirectory detected, source: %s\n", source)
	}

	var opts []ClientOption

	// Check if the protocol was switched to one which was not configured.
	if g.client != nil && g.client.Getters != nil {
		fmt.Printf("[DEBUG] HttpGetter.Get: Using client detectors to process source URL\n")
		// We must first use the Detectors provided, because `X-Terraform-Get does
		// not necessarily return a valid URL. We can replace the source string
		// here, since the detectors would have been called immediately during the
		// next Get anyway.
		source, err = Detect(source, g.client.Pwd, g.client.Detectors)
		if err != nil {
			fmt.Printf("[ERROR] HttpGetter.Get: Failed to detect source protocol: %v\n", err)
			return err
		}
		fmt.Printf("[DEBUG] HttpGetter.Get: Detected source after processing: %s\n", source)

		protocol := ""
		// X-Terraform-Get allows paths relative to the previous request too,
		// which won't have a protocol.
		if !relativeGet(source) {
			protocol = strings.Split(source, ":")[0]
			fmt.Printf("[DEBUG] HttpGetter.Get: Detected protocol: %s\n", protocol)
		} else {
			fmt.Printf("[DEBUG] HttpGetter.Get: Relative path detected, no protocol check needed\n")
		}

		// Otherwise, all default getters are allowed.
		if protocol != "" {
			_, allowed := g.client.Getters[protocol]
			if !allowed {
				fmt.Printf("[ERROR] HttpGetter.Get: Protocol %s not allowed by client configuration\n", protocol)
				return fmt.Errorf("no getter available for X-Terraform-Get source protocol: %q", protocol)
			}
			fmt.Printf("[DEBUG] HttpGetter.Get: Protocol %s is allowed by client configuration\n", protocol)
		}
	}

	// Add any getter client options.
	if g.client != nil {
		fmt.Printf("[DEBUG] HttpGetter.Get: Adding client options\n")
		opts = g.client.Options
	}

	fmt.Println("Part 4")
	// If the client is nil, we know we're using the HttpGetter directly. In
	// this case, we don't know exactly which protocols are configured, but we
	// can make a good guess.
	//
	// This prevents all default getters from being allowed when only using the
	// HttpGetter directly. To enable protocol switching, a client "wrapper" must
	// be used.
	if g.client == nil {
		fmt.Printf("[DEBUG] HttpGetter.Get: Client is nil, configuring getters for direct usage\n")
		switch {
		case subDir != "":
			// If there's a subdirectory, we will also need a file getter to
			// unpack it.
			fmt.Printf("[DEBUG] HttpGetter.Get: Subdirectory detected, adding file, http, and https getters\n")
			opts = append(opts, WithGetters(map[string]Getter{
				"file":  new(FileGetter),
				"http":  g,
				"https": g,
			}))
		default:
			fmt.Printf("[DEBUG] HttpGetter.Get: No subdirectory, adding http and https getters only\n")
			opts = append(opts, WithGetters(map[string]Getter{
				"http":  g,
				"https": g,
			}))
		}
	} else {
		fmt.Printf("[DEBUG] HttpGetter.Get: Using existing client configuration\n")
	}

	// Ensure we pass along the context we constructed in this function.
	//
	// This is especially important to enforce a limit on X-Terraform-Get redirects
	// which could be setup, if configured, at the top of this function.
	fmt.Printf("[DEBUG] HttpGetter.Get: Adding context to client options\n")
	opts = append(opts, WithContext(ctx))

	if subDir != "" {
		// We have a subdir, time to jump some hoops
		fmt.Printf("[DEBUG] HttpGetter.Get: Processing subdirectory download: %s\n", subDir)
		return g.getSubdir(ctx, dst, source, subDir, opts...)
	}

	fmt.Printf("[DEBUG] HttpGetter.Get: Delegating to generic Get function with source: %s\n", source)

	// Note: this allows the protocol to be switched to another configured getters.
	result := Get(dst, source, opts...)
	fmt.Println("result is", result)
	return result
}

// GetFile fetches the file from src and stores it at dst.
// If the server supports Accept-Range, HttpGetter will attempt a range
// request. This means it is the caller's responsibility to ensure that an
// older version of the destination file does not exist, else it will be either
// falsely identified as being replaced, or corrupted with extra bytes
// appended.
func (g *HttpGetter) GetFile(dst string, src *url.URL) error {
	ctx := g.Context()

	// Optionally enforce a maxiumum HTTP response body size.
	if g.MaxBytes > 0 {
		ctx = context.WithValue(ctx, httpMaxBytesValue, g.MaxBytes)
	}

	if g.Netrc {
		// Add auth from netrc if we can
		if err := addAuthFromNetrc(src); err != nil {
			return err
		}
	}
	// Create all the parent directories if needed
	if err := os.MkdirAll(filepath.Dir(dst), g.client.mode(0755)); err != nil {
		return err
	}

	f, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE, g.client.mode(0666))
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	if g.Client == nil {
		g.Client = httpClient
		if g.client != nil && g.client.Insecure {
			insecureTransport := cleanhttp.DefaultTransport()
			insecureTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
			g.Client.Transport = insecureTransport
		}
	}

	var (
		currentFileSize int64
		req             *http.Request
	)

	if !g.DoNotCheckHeadFirst {
		headCtx := ctx

		if g.HeadFirstTimeout > 0 {
			var cancel context.CancelFunc

			headCtx, cancel = context.WithTimeout(ctx, g.HeadFirstTimeout)
			defer cancel()
		}

		// We first make a HEAD request so we can check
		// if the server supports range queries. If the server/URL doesn't
		// support HEAD requests, we just fall back to GET.
		req, err = http.NewRequestWithContext(headCtx, "HEAD", src.String(), nil)
		if err != nil {
			return err
		}
		if g.Header != nil {
			req.Header = g.Header.Clone()
		}
		headResp, err := g.Client.Do(req)
		if err == nil {
			_ = headResp.Body.Close()
			if headResp.StatusCode == 200 {
				// If the HEAD request succeeded, then attempt to set the range
				// query if we can.
				if headResp.Header.Get("Accept-Ranges") == "bytes" && headResp.ContentLength >= 0 {
					if fi, err := f.Stat(); err == nil {
						if _, err = f.Seek(0, io.SeekEnd); err == nil {
							currentFileSize = fi.Size()
							if currentFileSize >= headResp.ContentLength {
								// file already present
								return nil
							}
						}
					}
				}
			}
		}
	}

	readCtx := ctx

	if g.ReadTimeout > 0 {
		var cancel context.CancelFunc
		readCtx, cancel = context.WithTimeout(ctx, g.ReadTimeout)
		defer cancel()
	}

	req, err = http.NewRequestWithContext(readCtx, "GET", src.String(), nil)
	if err != nil {
		return err
	}
	if g.Header != nil {
		req.Header = g.Header.Clone()
	}
	if currentFileSize > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", currentFileSize))
	}

	resp, err := g.Client.Do(req)
	if err != nil {
		return err
	}
	switch resp.StatusCode {
	case http.StatusOK, http.StatusPartialContent:
		// all good
	default:
		_ = resp.Body.Close()
		return fmt.Errorf("bad response code: %d", resp.StatusCode)
	}

	body := resp.Body

	if maxBytes := httpMaxBytesFromContext(ctx); maxBytes > 0 {
		body = newLimitedWrappedReaderCloser(body, maxBytes)
	}

	if g.client != nil && g.client.ProgressListener != nil {
		// track download
		fn := filepath.Base(src.EscapedPath())
		body = g.client.ProgressListener.TrackProgress(fn, currentFileSize, currentFileSize+resp.ContentLength, resp.Body)
	}
	defer func() { _ = body.Close() }()

	n, err := Copy(readCtx, f, body)
	if err == nil && n < resp.ContentLength {
		err = io.ErrShortWrite
	}
	return err
}

// getSubdir downloads the source into the destination, but with
// the proper subdir.
func (g *HttpGetter) getSubdir(ctx context.Context, dst, source, subDir string, opts ...ClientOption) error {
	// Create a temporary directory to store the full source. This has to be
	// a non-existent directory.
	td, tdcloser, err := safetemp.Dir("", "getter")
	if err != nil {
		return err
	}
	defer func() { _ = tdcloser.Close() }()

	// Download that into the given directory
	if err := Get(td, source, opts...); err != nil {
		return err
	}

	// Process any globbing
	sourcePath, err := SubdirGlob(td, subDir)
	if err != nil {
		return err
	}

	// Make sure the subdir path actually exists
	if _, err := os.Stat(sourcePath); err != nil {
		return fmt.Errorf(
			"Error downloading %s: %s", source, err)
	}

	// Copy the subdirectory into our actual destination.
	if err := os.RemoveAll(dst); err != nil {
		return err
	}

	// Make the final destination
	if err := os.MkdirAll(dst, g.client.mode(0755)); err != nil {
		return err
	}

	var disableSymlinks bool

	if g.client != nil && g.client.DisableSymlinks {
		disableSymlinks = true
	}

	return copyDir(ctx, dst, sourcePath, false, disableSymlinks, g.client.umask())
}

// parseMeta looks for the first meta tag in the given reader that
// will give us the source URL.
func (g *HttpGetter) parseMeta(ctx context.Context, r io.Reader) (string, error) {
	d := xml.NewDecoder(r)
	d.CharsetReader = charsetReader
	d.Strict = false
	var err error
	var t xml.Token
	for {
		if ctx.Err() != nil {
			return "", fmt.Errorf("context error while parsing meta tag: %w", ctx.Err())
		}

		t, err = d.Token()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return "", err
		}
		if e, ok := t.(xml.StartElement); ok && strings.EqualFold(e.Name.Local, "body") {
			return "", nil
		}
		if e, ok := t.(xml.EndElement); ok && strings.EqualFold(e.Name.Local, "head") {
			return "", nil
		}
		e, ok := t.(xml.StartElement)
		if !ok || !strings.EqualFold(e.Name.Local, "meta") {
			continue
		}
		if attrValue(e.Attr, "name") != "terraform-get" {
			continue
		}
		if f := attrValue(e.Attr, "content"); f != "" {
			return f, nil
		}
	}
}

// X-Terraform-Get allows paths relative to the previous request
var relativeGet = regexp.MustCompile(`^\.{0,2}/`).MatchString

// attrValue returns the attribute value for the case-insensitive key
// `name', or the empty string if nothing is found.
func attrValue(attrs []xml.Attr, name string) string {
	for _, a := range attrs {
		if strings.EqualFold(a.Name.Local, name) {
			return a.Value
		}
	}
	return ""
}

// charsetReader returns a reader for the given charset. Currently
// it only supports UTF-8 and ASCII. Otherwise, it returns a meaningful
// error which is printed by go get, so the user can find why the package
// wasn't downloaded if the encoding is not supported. Note that, in
// order to reduce potential errors, ASCII is treated as UTF-8 (i.e. characters
// greater than 0x7f are not rejected).
func charsetReader(charset string, input io.Reader) (io.Reader, error) {
	switch strings.ToLower(charset) {
	case "ascii":
		return input, nil
	default:
		return nil, fmt.Errorf("can't decode XML document using charset %q", charset)
	}
}
