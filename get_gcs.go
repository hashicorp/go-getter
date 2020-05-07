package getter

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

// GCSGetter is a Getter implementation that will download a module from
// a GCS bucket.
type GCSGetter struct {
}

func (g *GCSGetter) Mode(ctx context.Context, u *url.URL) (Mode, error) {

	// Parse URL
	bucket, object, err := g.parseURL(u)
	if err != nil {
		return 0, err
	}

	client, err := storage.NewClient(ctx)
	if err != nil {
		return 0, err
	}
	iter := client.Bucket(bucket).Objects(ctx, &storage.Query{Prefix: object})
	for {
		obj, err := iter.Next()
		if err != nil && err != iterator.Done {
			return 0, err
		}

		if err == iterator.Done {
			break
		}
		if strings.HasSuffix(obj.Name, "/") {
			// A directory matched the prefix search, so this must be a directory
			return ModeDir, nil
		} else if obj.Name != object {
			// A file matched the prefix search and doesn't have the same name
			// as the query, so this must be a directory
			return ModeDir, nil
		}
	}
	// There are no directories or subdirectories, and if a match was returned,
	// it was exactly equal to the prefix search. So return File mode
	return ModeFile, nil
}

func (g *GCSGetter) Get(ctx context.Context, req *Request) error {
	// Parse URL
	bucket, object, err := g.parseURL(req.u)
	if err != nil {
		return err
	}

	// Remove destination if it already exists
	_, err = os.Stat(req.Dst)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err == nil {
		// Remove the destination
		if err := os.RemoveAll(req.Dst); err != nil {
			return err
		}
	}

	// Create all the parent directories
	if err := os.MkdirAll(filepath.Dir(req.Dst), 0755); err != nil {
		return err
	}

	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}

	// Iterate through all matching objects.
	iter := client.Bucket(bucket).Objects(ctx, &storage.Query{Prefix: object})
	for {
		obj, err := iter.Next()
		if err != nil && err != iterator.Done {
			return err
		}
		if err == iterator.Done {
			break
		}

		if !strings.HasSuffix(obj.Name, "/") {
			// Get the object destination path
			objDst, err := filepath.Rel(object, obj.Name)
			if err != nil {
				return err
			}
			objDst = filepath.Join(req.Dst, objDst)
			// Download the matching object.
			err = g.getObject(ctx, client, objDst, bucket, obj.Name)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *GCSGetter) GetFile(ctx context.Context, req *Request) error {
	// Parse URL
	bucket, object, err := g.parseURL(req.u)
	if err != nil {
		return err
	}

	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	return g.getObject(ctx, client, req.Dst, bucket, object)
}

func (g *GCSGetter) getObject(ctx context.Context, client *storage.Client, dst, bucket, object string) error {
	rc, err := client.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
		return err
	}
	defer rc.Close()

	// Create all the parent directories
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = Copy(ctx, f, rc)
	return err
}

func (g *GCSGetter) parseURL(u *url.URL) (bucket, path string, err error) {
	if strings.Contains(u.Host, "googleapis.com") {
		hostParts := strings.Split(u.Host, ".")
		if len(hostParts) != 3 {
			err = fmt.Errorf("URL is not a valid GCS URL")
			return
		}

		pathParts := strings.SplitN(u.Path, "/", 5)
		if len(pathParts) != 5 {
			err = fmt.Errorf("URL is not a valid GCS URL")
			return
		}
		bucket = pathParts[3]
		path = pathParts[4]
	}
	return
}

func (g *GCSGetter) Detect(req *Request) (string, bool, error) {
	src := req.Src
	if len(src) == 0 {
		return "", false, nil
	}

	if req.forced != "" {
		// There's a getter being forced
		if !g.validScheme(req.forced) {
			// Current getter is not the forced one
			// Don't use it to try to download the artifact
			return "", false, nil
		}
	}
	isForcedGetter := req.forced != "" && g.validScheme(req.forced)

	u, err := url.Parse(src)
	if err == nil && u.Scheme != "" {
		if isForcedGetter {
			// Is the forced getter and source is a valid url
			return src, true, nil
		}
		if g.validScheme(u.Scheme) {
			return src, true, nil
		}
		// Valid url with a scheme that is not valid for current getter
		return "", false, nil
	}

	if strings.Contains(src, "googleapis.com/") {
		return g.detectHTTP(src)
	}

	if isForcedGetter {
		// Is the forced getter and should be used to download the artifact
		if req.Pwd != "" && !filepath.IsAbs(src) {
			// Make sure to add pwd to relative paths
			src = filepath.Join(req.Pwd, src)
		}
		// Make sure we're using "/" on Windows. URLs are "/"-based.
		return filepath.ToSlash(src), true, nil
	}

	return "", false, nil
}

func (g *GCSGetter) detectHTTP(src string) (string, bool, error) {

	parts := strings.Split(src, "/")
	if len(parts) < 5 {
		return "", false, fmt.Errorf(
			"URL is not a valid GCS URL")
	}
	version := parts[2]
	bucket := parts[3]
	object := strings.Join(parts[4:], "/")

	url, err := url.Parse(fmt.Sprintf("https://www.googleapis.com/storage/%s/%s/%s",
		version, bucket, object))
	if err != nil {
		return "", false, fmt.Errorf("error parsing GCS URL: %s", err)
	}

	return url.String(), true, nil
}

func (g *GCSGetter) validScheme(scheme string) bool {
	return scheme == "gcs"
}
