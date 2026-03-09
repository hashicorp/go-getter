package getter

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// archiveURLOverride, when non-empty, replaces the URL that fetchArchive
// downloads from. This exists so that tests can point at a local httptest
// server without needing a real hosting platform.
var archiveURLOverride string

// fetchArchive downloads a tarball archive of the given commit from the
// hosting platform's HTTP API and extracts it to dst. This is used as a
// fallback when git-fetch cannot retrieve a commit (e.g. orphaned commits
// that are unreachable from any ref).
//
// Authentication is resolved from URL userinfo, netrc, or environment
// variables (GH_TOKEN, GITLAB_TOKEN, etc.). SSH keys cannot be used here
// since this is an HTTP download — if the user only has SSH credentials
// configured, this fallback will fail for private repositories.
//
// If subdir is non-empty, only files under that subdirectory are placed in
// dst. The resulting directory is NOT a git repository.
func fetchArchive(ctx context.Context, dst string, u *url.URL, ref string, subdir string) error {
	aURL := archiveURLOverride
	if aURL == "" {
		var err error
		aURL, err = archiveURL(u, ref)
		if err != nil {
			return err
		}
	}

	// Parse the archive URL so we can attach credentials.
	archiveParsed, err := url.Parse(aURL)
	if err != nil {
		return err
	}

	// Carry over credentials from the original git URL if present,
	// otherwise fall back to the user's netrc file. Skip the common SSH
	// placeholder user "git" since it isn't a real credential.
	if u.User != nil && u.User.Username() != "" && u.User.Username() != "git" {
		archiveParsed.User = u.User
	} else if err := addAuthFromNetrc(archiveParsed); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, archiveParsed.String(), nil)
	if err != nil {
		return err
	}

	if archiveParsed.User != nil {
		password, _ := archiveParsed.User.Password()
		req.SetBasicAuth(archiveParsed.User.Username(), password)
	} else if token := tokenFromEnv(u.Host); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download archive from %s: %w", aURL, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download archive (%s): HTTP %d", aURL, resp.StatusCode)
	}

	gzipR, err := gzip.NewReader(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to decompress archive: %w", err)
	}
	defer func() { _ = gzipR.Close() }()

	if err := extractArchive(gzipR, dst, subdir); err != nil {
		return fmt.Errorf("failed to extract archive: %w", err)
	}

	return nil
}

// extractArchive reads a tar stream and extracts its contents into dst. The
// archive is expected to contain a single top-level directory (e.g.
// "repo-sha/") which is stripped from all paths. If subdir is non-empty, only
// entries under that subdirectory are extracted, and the subdir path is
// preserved relative to dst.
//
// This does not reuse the shared untar helper because hosting-platform
// archives require stripping the top-level directory and filtering by subdir,
// neither of which untar supports. Adding those concerns to untar would
// complicate a function shared by all tar-based decompressors.
func extractArchive(r io.Reader, dst string, subdir string) error {
	tarR := tar.NewReader(r)
	topDir := ""
	found := false

	for {
		hdr, err := tarR.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if hdr.Typeflag == tar.TypeXGlobalHeader || hdr.Typeflag == tar.TypeXHeader {
			continue
		}

		// Disallow parent traversal.
		if containsDotDot(hdr.Name) {
			return fmt.Errorf("entry contains '..': %s", hdr.Name)
		}

		// Discover and strip the top-level directory.
		if topDir == "" {
			topDir = strings.SplitN(hdr.Name, "/", 2)[0] + "/"
		}
		rel := strings.TrimPrefix(hdr.Name, topDir)
		if rel == "" {
			// This is the top-level directory entry itself; skip it.
			continue
		}

		// If a subdir filter is set, skip entries outside it.
		if subdir != "" {
			subdirPrefix := strings.TrimRight(subdir, "/") + "/"
			if !strings.HasPrefix(rel, subdirPrefix) {
				continue
			}
		}

		found = true
		outPath := filepath.Join(dst, filepath.FromSlash(rel))

		if hdr.FileInfo().IsDir() {
			if err := os.MkdirAll(outPath, 0755); err != nil {
				return err
			}
			continue
		}

		// Ensure parent directory exists.
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return err
		}

		if err := copyReader(outPath, tarR, hdr.FileInfo().Mode(), 0, 0); err != nil {
			return err
		}
	}

	if subdir != "" && !found {
		return fmt.Errorf("path %q not found in archive", subdir)
	}

	return nil
}

// archiveURL constructs a tarball download URL for the given ref based on the
// hosting platform detected from u's hostname. The API endpoints are used
// rather than the web URLs because they resolve short commit SHAs.
func archiveURL(u *url.URL, ref string) (string, error) {
	owner, repo, err := parseOwnerRepo(u.Path)
	if err != nil {
		return "", err
	}

	host := strings.ToLower(u.Host)
	// Strip port if present (e.g. "github.com:443" → "github.com").
	if i := strings.LastIndex(host, ":"); i != -1 {
		host = host[:i]
	}

	switch {
	case host == "github.com" || strings.HasSuffix(host, ".github.com"):
		return fmt.Sprintf("https://api.github.com/repos/%s/%s/tarball/%s", owner, repo, ref), nil
	case host == "gitlab.com" || strings.HasSuffix(host, ".gitlab.com"):
		return fmt.Sprintf("https://gitlab.com/api/v4/projects/%s%%2F%s/repository/archive.tar.gz?sha=%s", owner, repo, ref), nil
	case host == "bitbucket.org" || strings.HasSuffix(host, ".bitbucket.org"):
		return fmt.Sprintf("https://bitbucket.org/%s/%s/get/%s.tar.gz", owner, repo, ref), nil
	default:
		return "", fmt.Errorf("unsupported git hosting platform %q for archive fallback", u.Host)
	}
}

// tokenFromEnv returns an API token from well-known environment variables
// for the given host. It returns an empty string if no token is found.
func tokenFromEnv(host string) string {
	host = strings.ToLower(host)
	// Strip port if present.
	if i := strings.LastIndex(host, ":"); i != -1 {
		host = host[:i]
	}

	switch {
	case host == "github.com" || strings.HasSuffix(host, ".github.com"):
		// GH_TOKEN is the newer GitHub CLI convention; GITHUB_TOKEN is the
		// widely-used CI/Actions variable.
		if t := os.Getenv("GH_TOKEN"); t != "" {
			return t
		}
		return os.Getenv("GITHUB_TOKEN")
	case host == "gitlab.com" || strings.HasSuffix(host, ".gitlab.com"):
		if t := os.Getenv("GITLAB_TOKEN"); t != "" {
			return t
		}
		return os.Getenv("GL_TOKEN")
	case host == "bitbucket.org" || strings.HasSuffix(host, ".bitbucket.org"):
		return os.Getenv("BITBUCKET_TOKEN")
	default:
		return ""
	}
}

// parseOwnerRepo extracts the owner and repository name from a URL path
// like "/owner/repo.git" or "/owner/repo".
func parseOwnerRepo(rawPath string) (owner, repo string, err error) {
	path := strings.TrimPrefix(rawPath, "/")
	path = strings.TrimSuffix(path, ".git")
	parts := strings.SplitN(path, "/", 3)
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("cannot parse owner/repo from path %q", rawPath)
	}
	return parts[0], parts[1], nil
}
