package getter

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// SmbMountGetter is a Getter implementation that will download a module from
// a shared folder using the file system.
type SmbMountGetter struct {
}

func (g *SmbMountGetter) Mode(ctx context.Context, u *url.URL) (Mode, error) {
	if u.Host == "" || u.Path == "" {
		return 0, new(smbPathError)
	}

	prefix, path := g.findPrefixAndPath(u)
	u.Path = prefix + path

	return new(FileGetter).Mode(ctx, u)
}

func (g *SmbMountGetter) Get(ctx context.Context, req *Request) error {
	if req.u.Host == "" || req.u.Path == "" {
		return new(smbPathError)
	}

	prefix, path := g.findPrefixAndPath(req.u)
	req.u.Path = prefix + path

	return new(FileGetter).Get(ctx, req)
}

func (g *SmbMountGetter) GetFile(ctx context.Context, req *Request) error {
	if req.u.Host == "" || req.u.Path == "" {
		return new(smbPathError)
	}

	prefix, path := g.findPrefixAndPath(req.u)
	req.u.Path = prefix + path

	return new(FileGetter).GetFile(ctx, req)
}

func (g *SmbMountGetter) findPrefixAndPath(u *url.URL) (string, string) {
	var prefix, path string
	switch runtime.GOOS {
	case "windows":
		prefix = string(os.PathSeparator) + string(os.PathSeparator)
		path = filepath.Join(u.Host, u.Path)
	case "darwin":
		prefix = string(os.PathSeparator)
		path = filepath.Join("Volumes", u.Path)
	case "linux":
		prefix = string(os.PathSeparator)
		share := g.findShare(u)
		pwd := fmt.Sprintf("run/user/1000/gvfs/smb-share:server=%s,share=%s", u.Host, share)
		path = filepath.Join(pwd, u.Path)
	}
	return prefix, path
}

func (g *SmbMountGetter) findShare(u *url.URL) string {
	// Get shared directory
	path := strings.TrimPrefix(u.Path, "/")
	splt := regexp.MustCompile(`/`)
	directories := splt.Split(path, 2)

	if len(directories) > 0 {
		return directories[0]
	}

	return "."
}

func (g *SmbMountGetter) Detect(req *Request) (string, bool, error) {
	src := req.Src
	if len(src) == 0 {
		return "", false, nil
	}

	if req.Forced != "" {
		// There's a getter being Forced
		if !g.validScheme(req.Forced) {
			// Current getter is not the Forced one
			// Don't use it to try to download the artifact
			return "", false, nil
		}
	}
	isForcedGetter := req.Forced != "" && g.validScheme(req.Forced)

	u, err := url.Parse(src)
	if err == nil && u.Scheme != "" {
		if isForcedGetter {
			// Is the Forced getter and source is a valid url
			return src, true, nil
		}
		if g.validScheme(u.Scheme) {
			return src, true, nil
		}
		// Valid url with a scheme that is not valid for current getter
		return "", false, nil
	}

	if isForcedGetter {
		// Is the Forced getter and should be used to download the artifact
		if req.Pwd != "" && !filepath.IsAbs(src) {
			// Make sure to add pwd to relative paths
			src = filepath.Join(req.Pwd, src)
		}
		// Make sure we're using "/" on Windows. URLs are "/"-based.
		return filepath.ToSlash(src), true, nil
	}

	return "", false, nil
}

func (g *SmbMountGetter) validScheme(scheme string) bool {
	return scheme == "smb"
}
