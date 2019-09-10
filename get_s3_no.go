// +build go_getter_nos3

package getter

import (
	"net/url"
	"errors"
)

// S3Getter is a Getter implementation that will download a module from
// a S3 bucket.
type S3Getter struct {
	getter
}

func (g *S3Getter) ClientMode(u *url.URL) (ClientMode, error) {
	return 0, errors.New("not available")
}

func (g *S3Getter) Get(dst string, u *url.URL) error {
	return errors.New("not available")
}

func (g *S3Getter) GetFile(dst string, u *url.URL) error {
	return errors.New("not available")
}
