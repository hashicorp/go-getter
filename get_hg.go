package getter

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	urlhelper "github.com/hashicorp/go-getter/v2/helper/url"
	safetemp "github.com/hashicorp/go-safetemp"
)

// HgGetter is a Getter implementation that will download a module from
// a Mercurial repository.
type HgGetter struct {
}

func (g *HgGetter) Mode(ctx context.Context, _ *url.URL) (Mode, error) {
	return ModeDir, nil
}

func (g *HgGetter) Get(ctx context.Context, req *Request) error {
	if _, err := exec.LookPath("hg"); err != nil {
		return fmt.Errorf("hg must be available and on the PATH")
	}

	newURL, err := urlhelper.Parse(req.URL.String())
	if err != nil {
		return err
	}
	if fixWindowsDrivePath(newURL) {
		// See valid file path form on http://www.selenic.com/hg/help/urls
		newURL.Path = fmt.Sprintf("/%s", newURL.Path)
	}

	// Extract some query parameters we use
	var rev string
	q := newURL.Query()
	if len(q) > 0 {
		rev = q.Get("rev")
		q.Del("rev")

		newURL.RawQuery = q.Encode()
	}

	_, err = os.Stat(req.Dst)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err != nil {
		if err := g.clone(req.Dst, newURL); err != nil {
			return err
		}
	}

	if err := g.pull(req.Dst, newURL); err != nil {
		return err
	}

	return g.update(ctx, req.Dst, newURL, rev)
}

// GetFile for Hg doesn't support updating at this time. It will download
// the file every time.
func (g *HgGetter) GetFile(ctx context.Context, req *Request) error {
	// Create a temporary directory to store the full source. This has to be
	// a non-existent directory.
	td, tdcloser, err := safetemp.Dir("", "getter")
	if err != nil {
		return err
	}
	defer tdcloser.Close()

	// Get the filename, and strip the filename from the URL so we can
	// just get the repository directly.
	filename := filepath.Base(req.URL.Path)
	req.URL.Path = filepath.Dir(req.URL.Path)
	dst := req.Dst
	req.Dst = td

	// If we're on Windows, we need to set the host to "localhost" for hg
	if runtime.GOOS == "windows" {
		req.URL.Host = "localhost"
	}

	// Get the full repository
	if err := g.Get(ctx, req); err != nil {
		return err
	}

	// Copy the single file
	req.URL, err = urlhelper.Parse(fmtFileURL(filepath.Join(td, filename)))
	if err != nil {
		return err
	}

	fg := &FileGetter{}
	req.Copy = true
	req.Dst = dst
	return fg.GetFile(ctx, req)
}

func (g *HgGetter) clone(dst string, u *url.URL) error {
	cmd := exec.Command("hg", "clone", "-U", u.String(), dst)
	return getRunCommand(cmd)
}

func (g *HgGetter) pull(dst string, u *url.URL) error {
	cmd := exec.Command("hg", "pull")
	cmd.Dir = dst
	return getRunCommand(cmd)
}

func (g *HgGetter) update(ctx context.Context, dst string, u *url.URL, rev string) error {
	args := []string{"update"}
	if rev != "" {
		args = append(args, rev)
	}

	cmd := exec.CommandContext(ctx, "hg", args...)
	cmd.Dir = dst
	return getRunCommand(cmd)
}

func fixWindowsDrivePath(u *url.URL) bool {
	// hg assumes a file:/// prefix for Windows drive letter file paths.
	// (e.g. file:///c:/foo/bar)
	// If the URL Path does not begin with a '/' character, the resulting URL
	// path will have a file:// prefix. (e.g. file://c:/foo/bar)
	// See http://www.selenic.com/hg/help/urls and the examples listed in
	// http://selenic.com/repo/hg-stable/file/1265a3a71d75/mercurial/util.py#l1936
	return runtime.GOOS == "windows" && u.Scheme == "file" &&
		len(u.Path) > 1 && u.Path[0] != '/' && u.Path[1] == ':'
}
