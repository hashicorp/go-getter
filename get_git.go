// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package getter

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	urlhelper "github.com/hashicorp/go-getter/helper/url"
	safetemp "github.com/hashicorp/go-safetemp"
	version "github.com/hashicorp/go-version"
)

// GitGetter is a Getter implementation that will download a module from
// a git repository.
type GitGetter struct {
	getter
}

var lsRemoteSymRefRegexp = regexp.MustCompile(`ref: refs/heads/([^\s]+).*`)

func (g *GitGetter) ClientMode(_ context.Context, u *url.URL) (ClientMode, error) {
	return ClientModeDir, nil
}

func (g *GitGetter) Get(ctx context.Context, dst string, u *url.URL) error {

	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git must be available and on the PATH")
	}

	// The port number must be parseable as an integer. If not, the user
	// was probably trying to use a scp-style address, in which case the
	// ssh:// prefix must be removed to indicate that.
	//
	// This is not necessary in versions of Go which have patched
	// CVE-2019-14809 (e.g. Go 1.12.8+)
	if portStr := u.Port(); portStr != "" {
		if _, err := strconv.ParseUint(portStr, 10, 16); err != nil {
			return fmt.Errorf("invalid port number %q; if using the \"scp-like\" git address scheme where a colon introduces the path instead, remove the ssh:// portion and use just the git:: prefix", portStr)
		}
	}

	// Extract some query parameters we use
	var ref, sshKey, subdir string
	depth := 0 // 0 means "don't use shallow clone"

	q := u.Query()
	if len(q) > 0 {
		ref = q.Get("ref")
		q.Del("ref")

		sshKey = q.Get("sshkey")
		q.Del("sshkey")

		subdir = q.Get("subdir")
		q.Del("subdir")
		if subdir != "" {
			depth = 1
		}

		if n, err := strconv.Atoi(q.Get("depth")); err == nil {
			depth = n
		}
		q.Del("depth")

		// Copy the URL
		newU := *u
		u = &newU
		u.RawQuery = q.Encode()
	}

	var sshKeyFile string
	if sshKey != "" {
		// Check that the git version is sufficiently new.
		if err := checkGitVersion(ctx, "2.3"); err != nil {
			return fmt.Errorf("Error using ssh key: %v", err)
		}

		// We have an SSH key - decode it.
		raw, err := base64.StdEncoding.DecodeString(sshKey)
		if err != nil {
			return err
		}

		// Create a temp file for the key and ensure it is removed.
		fh, err := os.CreateTemp("", "go-getter")
		if err != nil {
			return err
		}
		sshKeyFile = fh.Name()
		defer func() { _ = os.Remove(sshKeyFile) }()

		// Set the permissions prior to writing the key material.
		if err := os.Chmod(sshKeyFile, 0600); err != nil {
			return err
		}

		// Write the raw key into the temp file.
		_, err = fh.Write(raw)
		_ = fh.Close()
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
		err = g.update(ctx, dst, sshKeyFile, u, ref, depth, subdir)
	} else {
		err = g.clone(ctx, dst, sshKeyFile, u, ref, depth, subdir)
	}
	if err != nil {
		// If git operations failed for a commit ID, try downloading via
		// the hosting platform's HTTP archive endpoint. This handles
		// orphaned commits that are unreachable via the git protocol.
		if gitCommitIDRegex.MatchString(ref) {
			if archiveErr := fetchArchive(ctx, dst, u, ref, subdir); archiveErr == nil {
				return nil
			} else {
				return errors.Join(err, archiveErr)
			}
		}
		return err
	}

	// Next: check out the proper tag/branch if it is specified, and checkout
	if ref != "" {
		if err := g.checkout(ctx, dst, ref); err != nil {
			return err
		}
	}

	// Lastly, download any/all submodules.
	return g.fetchSubmodules(ctx, dst, sshKeyFile, depth)
}

// GetFile for Git doesn't support updating at this time. It will download
// the file every time.
func (g *GitGetter) GetFile(ctx context.Context, dst string, u *url.URL) error {
	td, tdcloser, err := safetemp.Dir("", "getter")
	if err != nil {
		return err
	}
	defer func() { _ = tdcloser.Close() }()

	// Get the filename, and strip the filename from the URL so we can
	// just get the repository directly.
	filename := filepath.Base(u.Path)
	u.Path = filepath.Dir(u.Path)

	// Get the full repository
	if err := g.Get(ctx, td, u); err != nil {
		return err
	}

	// Copy the single file
	u, err = urlhelper.Parse(fmtFileURL(filepath.Join(td, filename)))
	if err != nil {
		return err
	}

	fg := &FileGetter{Copy: true}
	return fg.GetFile(ctx, dst, u)
}

func (g *GitGetter) checkout(ctx context.Context, dst string, ref string) error {
	cmd := exec.CommandContext(ctx, "git", "checkout", ref)
	cmd.Dir = dst
	return getRunCommand(cmd)
}

// gitCommitIDRegex is a pattern intended to match strings that seem
// "likely to be" git commit IDs, rather than named refs. This cannot be
// an exact decision because it's valid to name a branch or tag after a series
// of hexadecimal digits too.
//
// We require at least 7 digits here because that's the smallest size git
// itself will typically generate, and so it'll reduce the risk of false
// positives on short branch names that happen to also be "hex words".
var gitCommitIDRegex = regexp.MustCompile("^[0-9a-fA-F]{7,40}$")

func (g *GitGetter) clone(ctx context.Context, dst, sshKeyFile string, u *url.URL, ref string, depth int, subdir string) error {
	args := []string{"clone"}

	isCommitID := gitCommitIDRegex.MatchString(ref)

	originalRef := ref // we handle an unspecified ref differently than explicitly selecting the default branch below
	if ref == "" {
		ref = findRemoteDefaultBranch(ctx, u)
	}
	if depth > 0 {
		args = append(args, "--depth", strconv.Itoa(depth))
		if subdir == "" || !isCommitID {
			args = append(args, "--branch", ref)
		}
	}
	if subdir != "" {
		args = append(args, "--filter=blob:none")
		args = append(args, "--sparse")
		args = append(args, "--no-checkout")
		args = append(args, "--no-tags")
	}

	args = append(args, "--", u.String(), dst)

	cmd := exec.CommandContext(ctx, "git", args...)
	setupGitEnv(cmd, sshKeyFile)
	err := getRunCommand(cmd)
	if err != nil {
		_ = os.RemoveAll(dst)
		if depth > 0 && originalRef != "" {
			// If we're creating a shallow clone then the given ref must be
			// a named ref (branch or tag) rather than a commit directly.
			// We can't accurately recognize the resulting error here without
			// hard-coding assumptions about git's human-readable output, but
			// we can at least try a heuristic.
			if isCommitID {
				return fmt.Errorf("%w (note that setting 'depth' requires 'ref' to be a branch or tag name)", err)
			}
		}
		return err
	}

	if subdir != "" {
		cmd = exec.CommandContext(ctx, "git", "sparse-checkout", "set", subdir)
		cmd.Dir = dst
		err = getRunCommand(cmd)
		if err != nil {
			_ = os.RemoveAll(dst)
			return err
		}

		// If the commit is a long commit sha then we can fetch it
		if isCommitID && len(ref) == 40 {
			cmd = exec.CommandContext(ctx, "git", "fetch", "origin", ref, "--depth", "1", "--no-tags")
			cmd.Dir = dst
			err = getRunCommand(cmd)
			if err != nil {
				_ = os.RemoveAll(dst)
				return err
			}
		}

		// If the commit is a short commit sha then we will need to fetch the
		// commit graph to resolve it to a full hash. We use --filter=tree:0
		// to fetch only commit objects (no trees or blobs), which is much
		// smaller than --filter=blob:none. Once resolved, we fetch just that
		// single commit with its trees via sparse checkout.
		if isCommitID && len(ref) < 40 {
			cmd = exec.CommandContext(ctx, "git", "fetch", "--unshallow", "--filter=tree:0", "--no-tags")
			cmd.Dir = dst
			err = getRunCommand(cmd)
			if err != nil {
				_ = os.RemoveAll(dst)
				return err
			}

			// Resolve the short hash to a full hash
			cmd = exec.CommandContext(ctx, "git", "rev-parse", "--verify", ref)
			cmd.Dir = dst
			out, err := cmd.Output()
			if err != nil {
				_ = os.RemoveAll(dst)
				return err
			}
			fullRef := strings.TrimSpace(string(out))

			// Now fetch just that commit with depth 1 to get trees/blobs
			// for the sparse checkout
			cmd = exec.CommandContext(ctx, "git", "fetch", "origin", fullRef, "--depth", "1", "--no-tags")
			cmd.Dir = dst
			if err := getRunCommand(cmd); err != nil {
				_ = os.RemoveAll(dst)
				return err
			}
			ref = fullRef
		}

		if err := g.checkout(ctx, dst, ref); err != nil {
			_ = os.RemoveAll(dst)
			return err
		}
		return nil
	}

	if depth < 1 && originalRef != "" {
		// If we didn't add --depth and --branch above then we will now be
		// on the remote repository's default branch, rather than the selected
		// ref, so we'll need to fix that before we return.
		err := g.checkout(ctx, dst, originalRef)
		if err != nil {
			// Clean up git repository on disk
			_ = os.RemoveAll(dst)
			return err
		}
	}
	return nil
}

func (g *GitGetter) update(ctx context.Context, dst, sshKeyFile string, u *url.URL, ref string, depth int, subdir string) error {
	// Remove all variations of .git directories
	err := removeCaseInsensitiveGitDirectory(dst)
	if err != nil {
		return err
	}

	// Initialize the git repository
	cmd := exec.CommandContext(ctx, "git", "init")
	cmd.Dir = dst
	err = getRunCommand(cmd)
	if err != nil {
		return err
	}

	// Add the git remote
	cmd = exec.CommandContext(ctx, "git", "remote", "add", "origin", "--", u.String())
	cmd.Dir = dst
	err = getRunCommand(cmd)
	if err != nil {
		return err
	}

	// Fetch all tags so that tag-based refs can be resolved during checkout.
	// Skip this when depth > 0 because --tags fetches every tag reference
	// (e.g. 11k+ tags in large repos) regardless of --depth, and we already
	// fetch the specific ref we need below.
	if depth <= 0 {
		cmd = exec.CommandContext(ctx, "git", "fetch", "--tags")
		cmd.Dir = dst
		err = getRunCommand(cmd)
		if err != nil {
			return err
		}
	}

	// Fetch the remote ref
	fetchArgs := []string{"fetch", "origin"}
	if depth > 0 {
		fetchArgs = append(fetchArgs, "--depth", strconv.Itoa(depth))
	}
	if subdir != "" {
		fetchArgs = append(fetchArgs, "--filter=blob:none")
		fetchArgs = append(fetchArgs, "--no-tags")
	}
	fetchArgs = append(fetchArgs, "--", ref)
	cmd = exec.CommandContext(ctx, "git", fetchArgs...)
	cmd.Dir = dst
	err = getRunCommand(cmd)
	if err != nil {
		return err
	}

	// Reset the branch to the fetched ref
	cmd = exec.CommandContext(ctx, "git", "reset", "--hard", "FETCH_HEAD")
	cmd.Dir = dst
	err = getRunCommand(cmd)
	if err != nil {
		return err
	}

	// Set up sparse checkout if subdir is specified
	if subdir != "" {
		cmd = exec.CommandContext(ctx, "git", "sparse-checkout", "set", subdir)
		cmd.Dir = dst
		if err := getRunCommand(cmd); err != nil {
			return err
		}
	}

	// Checkout ref branch
	err = g.checkout(ctx, dst, ref)
	if err != nil {
		return err
	}

	// Pull the latest changes from the ref branch.
	// Skip this when subdir is set because we've already fetched the exact
	// ref we need above, and pull would re-fetch without --no-tags/--filter,
	// defeating our sparse/shallow optimisations.
	if subdir == "" {
		if depth > 0 {
			cmd = exec.CommandContext(ctx, "git", "pull", "origin", "--depth", strconv.Itoa(depth), "--ff-only", "--", ref)
		} else {
			cmd = exec.CommandContext(ctx, "git", "pull", "origin", "--ff-only", "--", ref)
		}

		cmd.Dir = dst
		setupGitEnv(cmd, sshKeyFile)
		return getRunCommand(cmd)
	}

	return nil
}

// fetchSubmodules downloads any configured submodules recursively.
func (g *GitGetter) fetchSubmodules(ctx context.Context, dst, sshKeyFile string, depth int) error {
	if g.client != nil {
		g.client.DisableSymlinks = true
	}
	args := []string{"submodule", "update", "--init", "--recursive"}
	if depth > 0 {
		args = append(args, "--depth", strconv.Itoa(depth))
	}
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dst
	setupGitEnv(cmd, sshKeyFile)
	return getRunCommand(cmd)
}

// findRemoteDefaultBranch checks the remote repo's HEAD symref to return the remote repo's
// default branch. "master" is returned if no HEAD symref exists.
func findRemoteDefaultBranch(ctx context.Context, u *url.URL) string {
	var stdoutbuf bytes.Buffer
	cmd := exec.CommandContext(ctx, "git", "ls-remote", "--symref", "--", u.String(), "HEAD")
	cmd.Stdout = &stdoutbuf
	err := cmd.Run()
	matches := lsRemoteSymRefRegexp.FindStringSubmatch(stdoutbuf.String())
	if err != nil || matches == nil {
		return "master"
	}
	return matches[len(matches)-1]
}

// setupGitEnv sets up the environment for the given command. This is used to
// pass configuration data to git and ssh and enables advanced cloning methods.
func setupGitEnv(cmd *exec.Cmd, sshKeyFile string) {
	// If there's no sshKeyFile argument to deal with, we can skip this
	// entirely.
	if sshKeyFile == "" {
		return
	}
	const gitSSHCommand = "GIT_SSH_COMMAND="
	var sshCmd []string

	// If we have an existing GIT_SSH_COMMAND, we need to append our options.
	// We will also remove our old entry to make sure the behavior is the same
	// with versions of Go < 1.9.
	env := os.Environ()
	for i, v := range env {
		if strings.HasPrefix(v, gitSSHCommand) && len(v) > len(gitSSHCommand) {
			sshCmd = []string{v}

			env[i], env[len(env)-1] = env[len(env)-1], env[i]
			env = env[:len(env)-1]
			break
		}
	}

	if len(sshCmd) == 0 {
		sshCmd = []string{gitSSHCommand + "ssh"}
	}

	// We have an SSH key temp file configured, tell ssh about this.
	if runtime.GOOS == "windows" {
		sshKeyFile = strings.ReplaceAll(sshKeyFile, `\`, `/`)
	}
	sshCmd = append(sshCmd, "-i", sshKeyFile)
	env = append(env, strings.Join(sshCmd, " "))

	cmd.Env = env
}

// checkGitVersion is used to check the version of git installed on the system
// against a known minimum version. Returns an error if the installed version
// is older than the given minimum.
func checkGitVersion(ctx context.Context, min string) error {
	want, err := version.NewVersion(min)
	if err != nil {
		return err
	}

	out, err := exec.CommandContext(ctx, "git", "version").Output()
	if err != nil {
		return err
	}

	fields := strings.Fields(string(out))
	if len(fields) < 3 {
		return fmt.Errorf("unexpected 'git version' output: %q", string(out))
	}
	v := fields[2]
	if runtime.GOOS == "windows" && strings.Contains(v, ".windows.") {
		// on windows, git version will return for example:
		// git version 2.20.1.windows.1
		// Which does not follow the semantic versionning specs
		// https://semver.org. We remove that part in order for
		// go-version to not error.
		v = v[:strings.Index(v, ".windows.")]
	}

	have, err := version.NewVersion(v)
	if err != nil {
		return err
	}

	if have.LessThan(want) {
		return fmt.Errorf("required git version = %s, have %s", want, have)
	}

	return nil
}

// removeCaseInsensitiveGitDirectory removes all .git directory variations
func removeCaseInsensitiveGitDirectory(dst string) error {
	files, err := os.ReadDir(dst)
	if err != nil {
		return fmt.Errorf("failed to read the destination directory %s during git update", dst)
	}
	for _, f := range files {
		if strings.EqualFold(f.Name(), ".git") && f.IsDir() {
			err := os.RemoveAll(filepath.Join(dst, f.Name()))
			if err != nil {
				return fmt.Errorf("failed to remove the .git directory in the destination directory %s during git update", dst)
			}
		}
	}
	return nil
}
