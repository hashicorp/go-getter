// +build go_getter_nogcs

package getter

import (
	"net/url"
	"errors"
)

// GCSGetter is a Getter implementation that will download a module from
// a GCS bucket.
type GCSGetter struct {
	getter
}

func (g *GCSGetter) ClientMode(u *url.URL) (ClientMode, error) {
	return 0, errors.New("not available")
}

func (g *GCSGetter) Get(dst string, u *url.URL) error {
	return errors.New("not available")
}

func (g *GCSGetter) GetFile(dst string, u *url.URL) error {
	return errors.New("not available")
}
