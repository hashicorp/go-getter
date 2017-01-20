package getter

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"

	urlhelper "github.com/hashicorp/go-getter/helper/url"
)

// GitGetter is a Getter implementation that will download a module from
// a git repository.
type GitGetter struct{}

func (g *GitGetter) ClientMode(_ *url.URL) (ClientMode, error) {
	return ClientModeDir, nil
}

func (g *GitGetter) Get(dst string, u *url.URL) error {
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git must be available and on the PATH")
	}

	// Extract some query parameters we use
	var ref, sshKey string
	q := u.Query()
	if len(q) > 0 {
		ref = q.Get("ref")
		q.Del("ref")

		sshKey = q.Get("sshkey")
		q.Del("sshkey")

		// Copy the URL
		var newU url.URL = *u
		u = &newU
		u.RawQuery = q.Encode()
	}

	var sshKeyFile string
	if sshKey != "" {
		// We have an SSH key - decode it.
		raw, err := base64.StdEncoding.DecodeString(sshKey)
		if err != nil {
			return err
		}

		// Create a temp file for the key and ensure it is removed.
		fh, err := ioutil.TempFile("", "go-getter")
		if err != nil {
			return err
		}
		sshKeyFile = fh.Name()
		defer os.Remove(sshKeyFile)

		// Set the permissions prior to writing the key material.
		if err := os.Chmod(sshKeyFile, 0600); err != nil {
			return err
		}

		// Write the raw key into the temp file.
		_, err = fh.Write(raw)
		fh.Close()
		if err != nil {
			return err
		}
	}

	// Clone or update the repository
	_, err := os.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err == nil {
		err = g.update(dst, sshKeyFile, ref)
	} else {
		err = g.clone(dst, sshKeyFile, u)
	}
	if err != nil {
		return err
	}

	// Next: check out the proper tag/branch if it is specified, and checkout
	if ref == "" {
		return nil
	}

	return g.checkout(dst, ref)
}

// GetFile for Git doesn't support updating at this time. It will download
// the file every time.
func (g *GitGetter) GetFile(dst string, u *url.URL) error {
	td, err := ioutil.TempDir("", "getter-git")
	if err != nil {
		return err
	}
	if err := os.RemoveAll(td); err != nil {
		return err
	}

	// Get the filename, and strip the filename from the URL so we can
	// just get the repository directly.
	filename := filepath.Base(u.Path)
	u.Path = filepath.Dir(u.Path)

	// Get the full repository
	if err := g.Get(td, u); err != nil {
		return err
	}

	// Copy the single file
	u, err = urlhelper.Parse(fmtFileURL(filepath.Join(td, filename)))
	if err != nil {
		return err
	}

	fg := &FileGetter{Copy: true}
	return fg.GetFile(dst, u)
}

func (g *GitGetter) checkout(dst string, ref string) error {
	cmd := exec.Command("git", "checkout", ref)
	cmd.Dir = dst
	return getRunCommand(cmd)
}

func (g *GitGetter) clone(dst, sshKeyFile string, u *url.URL) error {
	cmd := exec.Command("git", "clone", u.String(), dst)
	addSSHKeyFile(cmd, sshKeyFile)
	return getRunCommand(cmd)
}

func (g *GitGetter) update(dst, sshKeyFile, ref string) error {
	// Determine if we're a branch. If we're NOT a branch, then we just
	// switch to master prior to checking out
	cmd := exec.Command("git", "show-ref", "-q", "--verify", "refs/heads/"+ref)
	cmd.Dir = dst

	if getRunCommand(cmd) != nil {
		// Not a branch, switch to master. This will also catch non-existent
		// branches, in which case we want to switch to master and then
		// checkout the proper branch later.
		ref = "master"
	}

	// We have to be on a branch to pull
	if err := g.checkout(dst, ref); err != nil {
		return err
	}

	cmd = exec.Command("git", "pull", "--ff-only")
	cmd.Dir = dst
	addSSHKeyFile(cmd, sshKeyFile)
	return getRunCommand(cmd)
}

// addSSHKeyFile sets up the given SSH private key file such that it will
// be used by the "git" command during authentication. This is accomplished
// using a special environment variable, which is set on the provided cmd.
// If the sshKeyFile is empty, this is a noop.
func addSSHKeyFile(cmd *exec.Cmd, sshKeyFile string) {
	if sshKeyFile == "" {
		return
	}
	cmd.Env = append(os.Environ(), "GIT_SSH_COMMAND=ssh "+
		"-o StrictHostKeyChecking=no "+
		"-i "+sshKeyFile)
}
