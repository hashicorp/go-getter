package getter

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"cloud.google.com/go/storage"
)

// GCSGetter is a Getter implementation that will download a module from
// a GCS bucket.
type GCSGetter struct {
	getter
}

// ClientMode ...
// TODO: Check whether any files appear under the path
// by doing a bucket listing,
// or whether it shows as an exact match.
func (g *GCSGetter) ClientMode(u *url.URL) (ClientMode, error) {
	return ClientModeFile, nil
}

// Get ...
// TODO: might have to copy every file
func (g *GCSGetter) Get(dst string, u *url.URL) error {
	return nil
}

// GetFile ...
func (g *GCSGetter) GetFile(dst string, u *url.URL) error {
	ctx := g.Context()

	// Parse URL
	bucket, object, err := g.parseUrl(u)
	if err != nil {
		return err
	}

	sctx := context.Background()
	client, err := storage.NewClient(sctx)
	if err != nil {
		// TODO: Handle error.
	}
	rc, err := client.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
		return err
	}
	defer rc.Close()

	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = Copy(ctx, f, rc)
	return err
}

// https://www.googleapis.com/storage/v1/
func (g *GCSGetter) parseUrl(u *url.URL) (bucket, path string, err error) {
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
