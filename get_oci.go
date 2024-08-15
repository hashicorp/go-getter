package getter

import (
	"fmt"
	"net/url"
	"os"
	"path"

	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry/remote"
)

// OCIGetter is responsible for handling OCI repositories
type OCIGetter struct {
	getter
}

// ClientMode returns the client mode directory
func (g *OCIGetter) ClientMode(u *url.URL) (ClientMode, error) {
	return ClientModeDir, nil
}

// Get gets the repository as the specified url
func (g *OCIGetter) Get(path string, u *url.URL) error {
	ctx := g.Context()

	src, err := g.getRepository(u)
	if err != nil {
		return err
	}

	reference := src.Reference.Reference

	if reference == "" {
		reference = "latest"
	}

	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return fmt.Errorf("make directory for OCI storage: %w", err)
	}

	dst, err := file.New(path)
	if err != nil {
		return fmt.Errorf("cannot create file destination OCIGetter: %w", err)
	}
	defer dst.Close()

	_, err = oras.Copy(ctx, src, reference, dst, reference, oras.DefaultCopyOptions)
	if err != nil {
		return fmt.Errorf("unable to copy OCI artifact: %w", err)
	}

	return nil
}

func (g *OCIGetter) getRepository(u *url.URL) (*remote.Repository, error) {
	repository, err := remote.NewRepository(getReferenceFromURL(u))
	if err != nil {
		return nil, fmt.Errorf("invalid OCI URL: %w", err)
	}

	return repository, nil
}

func getReferenceFromURL(u *url.URL) (string) {
	return path.Join(u.Host, u.Path)
}

// GetFile is currently a NOOP
func (g *OCIGetter) GetFile(dst string, u *url.URL) error {
	return nil
}