package getter

import (
	"context"
	"fmt"
	"net/url"
)

// SmbGetter is a Getter implementation that will download a module from
// a shared folder using samba scheme.
type SmbGetter struct {
	getter
}

const pathError = "samba path should contain valid Host and filepath (smb://<host>/<file_path>)"

func (g *SmbGetter) Mode(ctx context.Context, u *url.URL) (Mode, error) {
	if u.Host == "" || u.Path == "" {
		return ModeFile, fmt.Errorf(pathError)
	}
	path := "//" + u.Host + u.Path
	if u.RawPath != "" {
		path = u.RawPath
	}
	return mode(path)
}

func (g *SmbGetter) Get(ctx context.Context, req *Request) error {
	if req.u.Host == "" || req.u.Path == "" {
		return fmt.Errorf(pathError)
	}
	path := "//" + req.u.Host + req.u.Path
	if req.u.RawPath != "" {
		path = req.u.RawPath
	}
	return get(path, req)
}

func (g *SmbGetter) GetFile(ctx context.Context, req *Request) error {
	if req.u.Host == "" || req.u.Path == "" {
		return fmt.Errorf(pathError)
	}
	path := "//" + req.u.Host + req.u.Path
	if req.u.RawPath != "" {
		path = req.u.RawPath
	}
	return getFile(path, req, ctx)
}
