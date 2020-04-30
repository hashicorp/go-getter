package getter

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// SmbGetter is a Getter implementation that will download a module from
// a shared folder using smbclient cli for Unix and local file system for Windows.
type SmbGetterWindows struct {
}

func (g *SmbGetterWindows) Mode(ctx context.Context, u *url.URL) (Mode, error) {
	if u.Host == "" || u.Path == "" {
		return 0, new(smbPathError)
	}

	if runtime.GOOS == "windows" {
		prefix := string(os.PathSeparator) + string(os.PathSeparator)
		u.Path = prefix + filepath.Join(u.Host, u.Path)
		return new(FileGetter).Mode(ctx, u)
	}

	return 0, nil
}

func (g *SmbGetterWindows) Get(ctx context.Context, req *Request) error {
	if req.u.Host == "" || req.u.Path == "" {
		return new(smbPathError)
	}

	if runtime.GOOS == "windows" {
		prefix := string(os.PathSeparator) + string(os.PathSeparator)
		req.u.Path = prefix + filepath.Join(req.u.Host, req.u.Path)
		return new(FileGetter).Get(ctx, req)
	}

	return nil
}

func (g *SmbGetterWindows) GetFile(ctx context.Context, req *Request) error {
	if req.u.Host == "" || req.u.Path == "" {
		return new(smbPathError)
	}

	if runtime.GOOS == "windows" {
		prefix := string(os.PathSeparator) + string(os.PathSeparator)
		req.u.Path = prefix + filepath.Join(req.u.Host, req.u.Path)
		return new(FileGetter).GetFile(ctx, req)
	}

	return nil
}

func (g *SmbGetterWindows) Detect(src, pwd string) (string, bool, error) {
	if len(src) == 0 {
		return "", false, nil
	}

	// Don't even try SmbGetterWindows if is not Windows
	if runtime.GOOS != "windows" {
		return "", false, nil
	}

	u, err := url.Parse(src)
	if err == nil && u.Scheme == "smb" {
		// Valid URL
		return src, true, nil
	}

	if windowsSmbPath(src) {
		// This is a valid smb path for Windows and will be checked in the SmbGetter
		// by the file system using the FileGetter, if available.
		return filepath.ToSlash(src), true, nil
	}

	return "", false, nil
}

func windowsSmbPath(path string) bool {
	return runtime.GOOS == "windows" && (strings.HasPrefix(path, "\\\\") || strings.HasPrefix(path, "//"))
}

func (g *SmbGetterWindows) ValidScheme(scheme string) bool {
	return scheme == "smb"
}
