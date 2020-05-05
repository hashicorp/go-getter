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
	next Getter
}

func (g *SmbMountGetter) Mode(ctx context.Context, u *url.URL) (Mode, error) {
	if u.Host == "" || u.Path == "" {
		return 0, new(smbPathError)
	}

	prefix, path := g.findPrefixAndPath(u)
	path = prefix + path

	if u.RawPath != "" {
		path = u.RawPath
	}

	fi, err := os.Stat(path)
	if err != nil {
		return 0, err
	}

	// Check if the source is a directory.
	if fi.IsDir() {
		return ModeDir, nil
	}

	return ModeFile, nil
}

func (g *SmbMountGetter) Get(ctx context.Context, req *Request) error {
	if req.u.Host == "" || req.u.Path == "" {
		return new(smbPathError)
	}

	prefix, path := g.findPrefixAndPath(req.u)
	path = prefix + path

	if req.u.RawPath != "" {
		path = req.u.RawPath
	}

	// The source path must exist and be a directory to be usable.
	if fi, err := os.Stat(path); err != nil {
		return fmt.Errorf("source path error: %s", err)
	} else if !fi.IsDir() {
		return fmt.Errorf("source path must be a directory")
	}

	fi, err := os.Lstat(req.Dst)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if req.Inplace {
		req.Dst = path
		return nil
	}

	// If the destination already exists, it must be a symlink
	if err == nil {
		mode := fi.Mode()
		if mode&os.ModeSymlink == 0 {
			return fmt.Errorf("destination exists and is not a symlink")
		}

		// Remove the destination
		if err := os.Remove(req.Dst); err != nil {
			return err
		}
	}

	// Create all the parent directories
	if err := os.MkdirAll(filepath.Dir(req.Dst), 0755); err != nil {
		return err
	}

	return SymlinkAny(path, req.Dst)
}

func (g *SmbMountGetter) GetFile(ctx context.Context, req *Request) error {
	if req.u.Host == "" || req.u.Path == "" {
		return new(smbPathError)
	}

	prefix, path := g.findPrefixAndPath(req.u)
	path = prefix + path

	if req.u.RawPath != "" {
		path = req.u.RawPath
	}

	// The source path must exist and be a file to be usable.
	if fi, err := os.Stat(path); err != nil {
		return fmt.Errorf("source path error: %s", err)
	} else if fi.IsDir() {
		return fmt.Errorf("source path must be a file")
	}

	if req.Inplace {
		req.Dst = path
		return nil
	}

	_, err := os.Lstat(req.Dst)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// If the destination already exists, it must be a symlink
	if err == nil {
		// Remove the destination
		if err := os.Remove(req.Dst); err != nil {
			return err
		}
	}

	// Create all the parent directories
	if err := os.MkdirAll(filepath.Dir(req.Dst), 0755); err != nil {
		return err
	}

	// If we're not copying, just symlink and we're done
	if !req.Copy {
		if err = os.Symlink(path, req.Dst); err == nil {
			return err
		}
		lerr, ok := err.(*os.LinkError)
		if !ok {
			return err
		}
		switch lerr.Err {
		case ErrUnauthorized:
			// On windows this  means we don't have
			// symlink privilege, let's
			// fallback to a copy to avoid an error.
			break
		default:
			return err
		}
	}

	// Copy
	srcF, err := os.Open(path)
	if err != nil {
		return err
	}
	defer srcF.Close()

	dstF, err := os.Create(req.Dst)
	if err != nil {
		return err
	}
	defer dstF.Close()

	_, err = Copy(ctx, dstF, srcF)
	return err
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

func (g *SmbMountGetter) DetectGetter(src, pwd string) (string, bool, error) {
	if len(src) == 0 {
		return "", false, nil
	}

	u, err := url.Parse(src)
	if err == nil && u.Scheme == "smb" {
		// Valid URL
		return src, true, nil
	}

	return "", false, nil
}

func (g *SmbMountGetter) ValidScheme(scheme string) bool {
	return scheme == "smb"
}

func (g *SmbMountGetter) Detect(src, pwd string) (string, []Getter, error) {
	return Detect(src, pwd, g)
}

func (g *SmbMountGetter) Next() Getter {
	return g.next
}

func (g *SmbMountGetter) SetNext(next Getter) {
	g.next = next
}
