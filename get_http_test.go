package getter

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	cleanhttp "github.com/hashicorp/go-cleanhttp"
	testing_helper "github.com/hashicorp/go-getter/v2/helper/testing"
)

func TestHttpGetter_impl(t *testing.T) {
	var _ Getter = new(HttpGetter)
}

func TestHttpGetter_header(t *testing.T) {
	ln := testHttpServer(t)
	defer ln.Close()
	ctx := context.Background()

	g := new(HttpGetter)
	dst := testing_helper.TempDir(t)
	defer os.RemoveAll(dst)

	var u url.URL
	u.Scheme = "http"
	u.Host = ln.Addr().String()
	u.Path = "/header"

	req := &Request{
		Dst:     dst,
		Src:     u.String(),
		u:       &u,
		GetMode: ModeDir,
	}

	// Get it, which should error because it uses the file protocol.
	err := g.Get(ctx, req)
	if !strings.Contains(err.Error(), "download not supported for scheme") {
		t.Fatalf("unexpected error: %v", err)
	}
	// But, using a wrapper client with a file getter will work.
	c := &Client{
		Getters: []Getter{
			g,
			new(FileGetter),
		},
	}

	if _, err = c.Get(ctx, req); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	mainPath := filepath.Join(dst, "main.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestHttpGetter_requestHeader(t *testing.T) {
	ln := testHttpServer(t)
	defer ln.Close()
	ctx := context.Background()

	g := new(HttpGetter)
	g.Header = make(http.Header)
	g.Header.Add("X-Foobar", "foobar")
	dst := testing_helper.TempDir(t)
	defer os.RemoveAll(dst)

	var u url.URL
	u.Scheme = "http"
	u.Host = ln.Addr().String()
	u.Path = "/expect-header"
	u.RawQuery = "expected=X-Foobar"

	req := &Request{
		Dst: dst,
		u:   &u,
	}

	// Get it!
	if err := g.GetFile(ctx, req); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	if _, err := os.Stat(dst); err != nil {
		t.Fatalf("err: %s", err)
	}
	testing_helper.AssertContents(t, dst, "Hello\n")
}

func TestHttpGetter_meta(t *testing.T) {
	ln := testHttpServer(t)
	defer ln.Close()
	ctx := context.Background()

	g := new(HttpGetter)
	dst := testing_helper.TempDir(t)
	defer os.RemoveAll(dst)

	var u url.URL
	u.Scheme = "http"
	u.Host = ln.Addr().String()
	u.Path = "/meta"

	req := &Request{
		Dst:     dst,
		Src:     u.String(),
		u:       &u,
		GetMode: ModeDir,
	}

	// Get it, which should error because it uses the file protocol.
	err := g.Get(ctx, req)
	if !strings.Contains(err.Error(), "download not supported for scheme") {
		t.Fatalf("unexpected error: %v", err)
	}
	// But, using a wrapper client with a file getter will work.
	c := &Client{
		Getters: []Getter{
			g,
			new(FileGetter),
		},
	}

	if _, err = c.Get(ctx, req); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	mainPath := filepath.Join(dst, "main.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestHttpGetter_metaSubdir(t *testing.T) {
	ln := testHttpServer(t)
	defer ln.Close()
	ctx := context.Background()

	g := new(HttpGetter)
	dst := testing_helper.TempDir(t)
	defer os.RemoveAll(dst)

	var u url.URL
	u.Scheme = "http"
	u.Host = ln.Addr().String()
	u.Path = "/meta-subdir"

	req := &Request{
		Dst:     dst,
		Src:     u.String(),
		u:       &u,
		GetMode: ModeDir,
	}

	// Get it, which should error because it uses the file protocol.
	err := g.Get(ctx, req)
	if !strings.Contains(err.Error(), "error downloading") {
		t.Fatalf("unexpected error: %v", err)
	}
	// But, using a wrapper client with a file getter will work.
	c := &Client{
		Getters: []Getter{
			g,
			new(FileGetter),
		},
	}

	if _, err = c.Get(ctx, req); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	mainPath := filepath.Join(dst, "sub.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestHttpGetter_metaSubdirGlob(t *testing.T) {
	ln := testHttpServer(t)
	defer ln.Close()
	ctx := context.Background()

	g := new(HttpGetter)
	dst := testing_helper.TempDir(t)
	defer os.RemoveAll(dst)

	var u url.URL
	u.Scheme = "http"
	u.Host = ln.Addr().String()
	u.Path = "/meta-subdir-glob"

	req := &Request{
		Dst:     dst,
		Src:     u.String(),
		u:       &u,
		GetMode: ModeDir,
	}

	// Get it, which should error because it uses the file protocol.
	err := g.Get(ctx, req)
	if !strings.Contains(err.Error(), "error downloading") {
		t.Fatalf("unexpected error: %v", err)
	}
	// But, using a wrapper client with a file getter will work.
	c := &Client{
		Getters: []Getter{
			g,
			new(FileGetter),
		},
	}

	if _, err = c.Get(ctx, req); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	mainPath := filepath.Join(dst, "sub.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestHttpGetter_none(t *testing.T) {
	ln := testHttpServer(t)
	defer ln.Close()
	ctx := context.Background()

	g := new(HttpGetter)
	dst := testing_helper.TempDir(t)
	defer os.RemoveAll(dst)

	var u url.URL
	u.Scheme = "http"
	u.Host = ln.Addr().String()
	u.Path = "/none"

	req := &Request{
		Dst: dst,
		u:   &u,
	}

	// Get it!
	if err := g.Get(ctx, req); err == nil {
		t.Fatal("should error")
	}
}

func TestHttpGetter_resume(t *testing.T) {
	load := []byte(testHttpMetaStr)
	sha := sha256.New()
	if n, err := sha.Write(load); n != len(load) || err != nil {
		t.Fatalf("sha write failed: %d, %s", n, err)
	}
	checksum := hex.EncodeToString(sha.Sum(nil))
	downloadFrom := len(load) / 2

	ln := testHttpServer(t)
	defer ln.Close()

	dst := testing_helper.TempDir(t)
	defer os.RemoveAll(dst)

	dst = filepath.Join(dst, "..", "range")
	f, err := os.Create(dst)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if n, err := f.Write(load[:downloadFrom]); n != downloadFrom || err != nil {
		t.Fatalf("partial file write failed: %d, %s", n, err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close failed: %s", err)
	}

	u := url.URL{
		Scheme:   "http",
		Host:     ln.Addr().String(),
		Path:     "/range",
		RawQuery: "checksum=" + checksum,
	}
	t.Logf("url: %s", u.String())
	ctx := context.Background()

	// Finish getting it!
	if _, err := GetFile(ctx, dst, u.String()); err != nil {
		t.Fatalf("finishing download should not error: %v", err)
	}

	b, err := ioutil.ReadFile(dst)
	if err != nil {
		t.Fatalf("readfile failed: %v", err)
	}

	if string(b) != string(load) {
		t.Fatalf("file differs: got:\n%s\n expected:\n%s\n", string(b), string(load))
	}

	// Get it again
	if _, err := GetFile(ctx, dst, u.String()); err != nil {
		t.Fatalf("should not error: %v", err)
	}
}

// The server may support Byte-Range, but has no size for the requested object
func TestHttpGetter_resumeNoRange(t *testing.T) {
	load := []byte(testHttpMetaStr)
	sha := sha256.New()
	if n, err := sha.Write(load); n != len(load) || err != nil {
		t.Fatalf("sha write failed: %d, %s", n, err)
	}
	checksum := hex.EncodeToString(sha.Sum(nil))
	downloadFrom := len(load) / 2

	ln := testHttpServer(t)
	defer ln.Close()

	dst := testing_helper.TempDir(t)
	defer os.RemoveAll(dst)

	dst = filepath.Join(dst, "..", "range")
	f, err := os.Create(dst)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if n, err := f.Write(load[:downloadFrom]); n != downloadFrom || err != nil {
		t.Fatalf("partial file write failed: %d, %s", n, err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close failed: %s", err)
	}

	u := url.URL{
		Scheme:   "http",
		Host:     ln.Addr().String(),
		Path:     "/no-range",
		RawQuery: "checksum=" + checksum,
	}
	t.Logf("url: %s", u.String())
	ctx := context.Background()

	// Finish getting it!
	if _, err := GetFile(ctx, dst, u.String()); err != nil {
		t.Fatalf("finishing download should not error: %v", err)
	}

	b, err := ioutil.ReadFile(dst)
	if err != nil {
		t.Fatalf("readfile failed: %v", err)
	}

	if string(b) != string(load) {
		t.Fatalf("file differs: got:\n%s\n expected:\n%s\n", string(b), string(load))
	}
}

func TestHttpGetter_file(t *testing.T) {
	ln := testHttpServer(t)
	defer ln.Close()
	ctx := context.Background()

	g := new(HttpGetter)
	dst := testing_helper.TempTestFile(t)
	defer os.RemoveAll(filepath.Dir(dst))

	var u url.URL
	u.Scheme = "http"
	u.Host = ln.Addr().String()
	u.Path = "/file"

	req := &Request{
		Dst: dst,
		u:   &u,
	}

	// Get it!
	if err := g.GetFile(ctx, req); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	if _, err := os.Stat(dst); err != nil {
		t.Fatalf("err: %s", err)
	}
	testing_helper.AssertContents(t, dst, "Hello\n")
}

func TestHttpGetter_auth(t *testing.T) {
	ln := testHttpServer(t)
	defer ln.Close()
	ctx := context.Background()

	g := new(HttpGetter)
	dst := testing_helper.TempDir(t)
	defer os.RemoveAll(dst)

	var u url.URL
	u.Scheme = "http"
	u.Host = ln.Addr().String()
	u.Path = "/meta-auth"
	u.User = url.UserPassword("foo", "bar")

	req := &Request{
		Dst:     dst,
		Src:     u.String(),
		u:       &u,
		GetMode: ModeDir,
	}

	// Get it, which should error because it uses the file protocol.
	err := g.Get(ctx, req)
	if !strings.Contains(err.Error(), "download not supported for scheme") {
		t.Fatalf("unexpected error: %v", err)
	}
	// But, using a wrapper client with a file getter will work.
	c := &Client{
		Getters: []Getter{
			g,
			new(FileGetter),
		},
	}

	if _, err = c.Get(ctx, req); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	mainPath := filepath.Join(dst, "main.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestHttpGetter_authNetrc(t *testing.T) {
	ln := testHttpServer(t)
	defer ln.Close()
	ctx := context.Background()

	g := new(HttpGetter)
	dst := testing_helper.TempDir(t)
	defer os.RemoveAll(dst)

	var u url.URL
	u.Scheme = "http"
	u.Host = ln.Addr().String()
	u.Path = "/meta"

	// Write the netrc file
	path, closer := testing_helper.TempFileWithContent(t, fmt.Sprintf(testHttpNetrc, ln.Addr().String()))
	defer closer()
	defer tempEnv(t, "NETRC", path)()

	req := &Request{
		Dst:     dst,
		Src:     u.String(),
		u:       &u,
		GetMode: ModeDir,
	}

	// Get it, which should error because it uses the file protocol.
	err := g.Get(ctx, req)
	if !strings.Contains(err.Error(), "download not supported for scheme") {
		t.Fatalf("unexpected error: %v", err)
	}
	// But, using a wrapper client with a file getter will work.
	c := &Client{
		Getters: []Getter{
			g,
			new(FileGetter),
		},
	}

	if _, err = c.Get(ctx, req); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	mainPath := filepath.Join(dst, "main.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

// test round tripper that only returns an error
type errRoundTripper struct{}

func (errRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return nil, errors.New("test round tripper")
}

// verify that the default httpClient no longer comes from http.DefaultClient
func TestHttpGetter_cleanhttp(t *testing.T) {
	ln := testHttpServer(t)
	defer ln.Close()

	// break the default http client
	http.DefaultClient.Transport = errRoundTripper{}
	defer func() {
		http.DefaultClient.Transport = http.DefaultTransport
	}()
	ctx := context.Background()

	g := new(HttpGetter)
	dst := testing_helper.TempDir(t)
	defer os.RemoveAll(dst)

	var u url.URL
	u.Scheme = "http"
	u.Host = ln.Addr().String()
	u.Path = "/header"

	req := &Request{
		Dst:     dst,
		Src:     u.String(),
		u:       &u,
		GetMode: ModeDir,
	}

	// Get it, which should error because it uses the file protocol.
	err := g.Get(ctx, req)
	if !strings.Contains(err.Error(), "download not supported for scheme") {
		t.Fatalf("unexpected error: %v", err)
	}
	// But, using a wrapper client with a file getter will work.
	c := &Client{
		Getters: []Getter{
			g,
			new(FileGetter),
		},
	}

	if _, err = c.Get(ctx, req); err != nil {
		t.Fatalf("err: %s", err)
	}

}

func TestHttpGetter__RespectsContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	ln := testHttpServer(t)

	var u url.URL
	u.Scheme = "http"
	u.Host = ln.Addr().String()
	u.Path = "/file"

	dst := testing_helper.TempDir(t)
	defer os.RemoveAll(dst)

	rt := hookableHTTPRoundTripper{
		before: func(req *http.Request) {
			err := req.Context().Err()
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("Expected http.Request with canceled.Context, got: %v", err)
			}
		},
		RoundTripper: http.DefaultTransport,
	}

	g := new(HttpGetter)
	g.Client = &http.Client{
		Transport: &rt,
	}

	req := Request{
		Dst: dst,
		u:   &u,
	}
	err := g.Get(ctx, &req)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}
}

func TestHttpGetter__XTerraformGetLimit(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ln := testHttpServerWithXTerraformGetLoop(t)

	var u url.URL
	u.Scheme = "http"
	u.Host = ln.Addr().String()
	u.Path = "/loop"

	dst := testing_helper.TempDir(t)
	defer os.RemoveAll(dst)

	g := new(HttpGetter)
	g.XTerraformGetLimit = 10
	g.Client = &http.Client{}

	req := Request{
		Dst:     dst,
		u:       &u,
		GetMode: ModeDir,
	}

	err := g.Get(ctx, &req)
	if !strings.Contains(err.Error(), "too many X-Terraform-Get redirects") {
		t.Fatalf("too many X-Terraform-Get redirects, got: %v", err)
	}
}

func TestHttpGetter__XTerraformGetDisabled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ln := testHttpServerWithXTerraformGetLoop(t)

	var u url.URL
	u.Scheme = "http"
	u.Host = ln.Addr().String()
	u.Path = "/loop"
	dst := testing_helper.TempDir(t)

	g := new(HttpGetter)
	g.XTerraformGetDisabled = true
	g.Client = &http.Client{}

	req := Request{
		Dst:     dst,
		u:       &u,
		GetMode: ModeDir,
	}

	err := g.Get(ctx, &req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
func TestHttpGetter__XTerraformGetProxyBypass(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ln := testHttpServerWithXTerraformGetProxyBypass(t)

	proxyLn := testHttpServerProxy(t, ln.Addr().String())

	t.Logf("starting malicious server on: %v", ln.Addr().String())
	t.Logf("starting proxy on: %v", proxyLn.Addr().String())

	var u url.URL
	u.Scheme = "http"
	u.Host = ln.Addr().String()
	u.Path = "/start"
	dst := testing_helper.TempDir(t)

	proxy, err := url.Parse(fmt.Sprintf("http://%s/", proxyLn.Addr().String()))
	if err != nil {
		t.Fatalf("failed to parse proxy URL: %v", err)
	}

	transport := cleanhttp.DefaultTransport()
	transport.Proxy = http.ProxyURL(proxy)

	g := new(HttpGetter)
	g.XTerraformGetLimit = 10
	g.Client = &http.Client{
		Transport: transport,
	}

	client := &Client{
		Getters: []Getter{g},
	}

	req := Request{
		Dst: dst,
		Src: u.String(),
	}

	_, err = client.Get(ctx, &req)
	if err != nil {
		t.Logf("client get error: %v", err)
	}
}

func TestHttpGetter__XTerraformGetConfiguredGettersBypass(t *testing.T) {
	tc := []struct {
		name              string
		configuredGetters []Getter
		errExpected       bool
	}{
		{name: "configured getter for git protocol switch", configuredGetters: []Getter{new(GitGetter)}, errExpected: false},
		{name: "configured getter for multiple protocol switch", configuredGetters: []Getter{new(HgGetter), new(GitGetter), new(FileGetter)}, errExpected: false},
		{name: "configured getter for file protocol switch", configuredGetters: []Getter{new(FileGetter)}, errExpected: true},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ln := testHttpServerWithXTerraformGetConfiguredGettersBypass(t)

	var u url.URL
	u.Scheme = "http"
	u.Host = ln.Addr().String()
	u.Path = "/start"

	for _, tt := range tc {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			dst := testing_helper.TempDir(t)

			rt := hookableHTTPRoundTripper{
				before: func(req *http.Request) {
					t.Logf("making request")
				},
				RoundTripper: http.DefaultTransport,
			}

			g := new(HttpGetter)
			g.XTerraformGetLimit = 10
			g.Client = &http.Client{
				Transport: &rt,
			}

			client := &Client{
				Getters: []Getter{g},
			}
			client.Getters = append(client.Getters, tt.configuredGetters...)

			t.Logf("%v", u.String())

			req := Request{
				Dst:     dst,
				Src:     u.String(),
				GetMode: ModeDir,
			}

			_, err := client.Get(ctx, &req)
			if tt.errExpected && err == nil {
				t.Fatalf("error expected")
			}
			if err != nil {
				if !strings.Contains(err.Error(), "download not supported for scheme") {
					t.Fatalf("expected download not supported for scheme, got: %v", err)
				}
			}
		})
	}
}

func TestHttpGetter__endless_body(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ln := testHttpServerWithEndlessBody(t)

	var u url.URL
	u.Scheme = "http"
	u.Host = ln.Addr().String()
	u.Path = "/"
	dst := testing_helper.TempDir(t)

	g := new(HttpGetter)
	g.MaxBytes = 10
	g.DoNotCheckHeadFirst = true

	client := &Client{
		Getters: []Getter{g},
	}

	t.Logf("%v", u.String())

	req := Request{
		Dst:     dst,
		Src:     u.String(),
		GetMode: ModeFile,
	}

	_, err := client.Get(ctx, &req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHttpGetter_subdirLink(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ln := testHttpServerSubDir(t)
	defer ln.Close()

	dst, err := ioutil.TempDir("", "tf")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	t.Logf("dst: %q", dst)

	var u url.URL
	u.Scheme = "http"
	u.Host = ln.Addr().String()
	u.Path = "/regular-subdir//meta-subdir"

	g := new(HttpGetter)
	client := &Client{
		Getters: []Getter{g},
	}

	t.Logf("url: %q", u.String())

	req := Request{
		Dst:     dst,
		Src:     u.String(),
		GetMode: ModeAny,
	}

	_, err = client.Get(ctx, &req)
	if err != nil {
		t.Fatalf("get err: %v", err)
	}
}

func testHttpServerWithXTerraformGetLoop(t *testing.T) net.Listener {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	header := fmt.Sprintf("http://%v:%v", ln.Addr().String(), "/loop")

	mux := http.NewServeMux()
	mux.HandleFunc("/loop", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Terraform-Get", header)
		t.Logf("serving loop")
	})

	var server http.Server
	server.Handler = mux
	go server.Serve(ln)

	return ln
}

func testHttpServerWithXTerraformGetProxyBypass(t *testing.T) net.Listener {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	header := fmt.Sprintf("http://%v/bypass", ln.Addr().String())

	mux := http.NewServeMux()
	mux.HandleFunc("/start/start", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Terraform-Get", header)
		t.Logf("serving start")
	})

	mux.HandleFunc("/bypass", func(w http.ResponseWriter, r *http.Request) {
		t.Fail()
		t.Logf("bypassed proxy")
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("serving HTTP server path: %v", r.URL.Path)
	})

	var server http.Server
	server.Handler = mux
	go server.Serve(ln)

	return ln
}

func testHttpServerWithXTerraformGetConfiguredGettersBypass(t *testing.T) net.Listener {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	header := fmt.Sprintf("git::http://%v/some/repository.git", ln.Addr().String())

	mux := http.NewServeMux()
	mux.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Terraform-Get", header)
		t.Logf("serving start")
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("serving git HTTP server path: %v", r.URL.Path)
	})

	var server http.Server
	server.Handler = mux
	go server.Serve(ln)

	return ln
}

func TestHttpGetter_XTerraformWithClientFromContext(t *testing.T) {
	tc := []struct {
		name        string
		client      *Client
		errExpected bool
	}{
		{
			name: "default getters",
			client: &Client{
				Getters: Getters,
			},
			errExpected: false,
		},
		{
			name: "client configured with needed getters",
			client: &Client{
				Getters: []Getter{
					new(HttpGetter),
					new(FileGetter),
				},
			},
			errExpected: false,
		},
		{
			name:        "nil client",
			errExpected: true,
		},
	}

	for _, tt := range tc {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ln := testHttpServer(t)
			defer ln.Close()
			ctx := context.Background()

			g := new(HttpGetter)
			dst := testing_helper.TempDir(t)
			defer os.RemoveAll(dst)

			var u url.URL
			u.Scheme = "http"
			u.Host = ln.Addr().String()
			u.Path = "/header"

			req := &Request{
				Dst:     dst,
				Src:     u.String(),
				u:       &u,
				GetMode: ModeDir,
			}

			// Using a client stored in the ctx with a file getter should work
			ctx = NewContextWithClient(ctx, tt.client)

			err := g.Get(ctx, req)
			if tt.errExpected && err == nil {
				t.Fatalf("error expected")
			}

			if err != nil {
				if !strings.Contains(err.Error(), "download not supported for scheme") {
					t.Fatalf("expected download not supported for scheme, got: %v", err)
				}
				return
			}

			// Verify the main file exists
			mainPath := filepath.Join(dst, "main.tf")
			if _, err := os.Stat(mainPath); err != nil {
				t.Fatalf("err: %s", err)
			}
		})
	}
}

func testHttpServerProxy(t *testing.T, upstreamHost string) net.Listener {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("serving proxy: %v: %#+v", r.URL.Path, r.Header)
		// create the reverse proxy
		proxy := httputil.NewSingleHostReverseProxy(r.URL)
		// Note that ServeHttp is non blocking & uses a go routine under the hood
		proxy.ServeHTTP(w, r)
	})

	var server http.Server
	server.Handler = mux
	go server.Serve(ln)

	return ln
}

func testHttpServer(t *testing.T) net.Listener {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/expect-header", testHttpHandlerExpectHeader)
	mux.HandleFunc("/file", testHttpHandlerFile)
	mux.HandleFunc("/header", testHttpHandlerHeader)
	mux.HandleFunc("/meta", testHttpHandlerMeta)
	mux.HandleFunc("/meta-auth", testHttpHandlerMetaAuth)
	mux.HandleFunc("/meta-subdir", testHttpHandlerMetaSubdir)
	mux.HandleFunc("/meta-subdir-glob", testHttpHandlerMetaSubdirGlob)
	mux.HandleFunc("/range", testHttpHandlerRange)
	mux.HandleFunc("/no-range", testHttpHandlerNoRange)

	var server http.Server
	server.Handler = mux
	go server.Serve(ln)

	return ln
}

func testHttpServerWithEndlessBody(t *testing.T) net.Listener {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		for {
			w.Write([]byte(".\n"))
		}
	})

	var server http.Server
	server.Handler = mux
	go server.Serve(ln)

	return ln
}

func testHttpHandlerExpectHeader(w http.ResponseWriter, r *http.Request) {
	if expected, ok := r.URL.Query()["expected"]; ok {
		if r.Header.Get(expected[0]) != "" {
			w.Write([]byte("Hello\n"))
			return
		}
	}

	w.WriteHeader(400)
}

func testHttpHandlerFile(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello\n"))
}

func testHttpHandlerHeader(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("X-Terraform-Get", testModuleURL("basic").String())
	w.WriteHeader(200)
}

func testHttpHandlerMeta(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(fmt.Sprintf(testHttpMetaStr, testModuleURL("basic").String())))
}

func testHttpHandlerMetaAuth(w http.ResponseWriter, r *http.Request) {
	user, pass, ok := r.BasicAuth()
	if !ok {
		w.WriteHeader(401)
		return
	}

	if user != "foo" || pass != "bar" {
		w.WriteHeader(401)
		return
	}

	w.Write([]byte(fmt.Sprintf(testHttpMetaStr, testModuleURL("basic").String())))
}

func testHttpHandlerMetaSubdir(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(fmt.Sprintf(testHttpMetaStr, testModuleURL("basic//subdir").String())))
}

func testHttpHandlerMetaSubdirGlob(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(fmt.Sprintf(testHttpMetaStr, testModuleURL("basic//sub*").String())))
}

func testHttpHandlerNone(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(testHttpNoneStr))
}

func testHttpHandlerRange(w http.ResponseWriter, r *http.Request) {
	load := []byte(testHttpMetaStr)
	switch r.Method {
	case "HEAD":
		w.Header().Add("accept-ranges", "bytes")
		w.Header().Add("content-length", strconv.Itoa(len(load)))
	default:
		// request should have header "Range: bytes=0-1023"
		// or                         "Range: bytes=123-"
		rangeHeaderValue := strings.Split(r.Header.Get("Range"), "=")[1]
		rng, _ := strconv.Atoi(strings.Split(rangeHeaderValue, "-")[0])
		if rng < 1 || rng > len(load) {
			http.Error(w, "", http.StatusBadRequest)
		}
		w.Write(load[rng:])
	}
}

func testHttpHandlerNoRange(w http.ResponseWriter, r *http.Request) {
	load := []byte(testHttpMetaStr)
	switch r.Method {
	case "HEAD":
		// we support range, but the object size isn't known
		w.Header().Add("accept-ranges", "bytes")
	default:
		if r.Header.Get("Range") != "" {
			http.Error(w, "range not supported", http.StatusBadRequest)
		}
		w.Write(load)
	}
}

func testHttpServerSubDir(t *testing.T) net.Listener {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			t.Logf("serving: %v: %v: %#+[1]v", r.Method, r.URL.String(), r.Header)
		}
	})

	var server http.Server
	server.Handler = mux
	go server.Serve(ln)

	return ln
}

const testHttpMetaStr = `
<html>
<head>
<meta name="terraform-get" content="%s">
</head>
</html>
`

const testHttpNoneStr = `
<html>
<head>
</head>
</html>
`

const testHttpNetrc = `
machine %s
login foo
password bar
`

type hookableHTTPRoundTripper struct {
	before func(req *http.Request)
	http.RoundTripper
}

func (m *hookableHTTPRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.before != nil {
		m.before(req)
	}
	return m.RoundTripper.RoundTrip(req)
}
