package getter

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

func TestArchiveURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		ref     string
		want    string
		wantErr bool
	}{
		{
			name: "github with .git suffix",
			url:  "https://github.com/hashicorp/terraform.git",
			ref:  "abc1234",
			want: "https://api.github.com/repos/hashicorp/terraform/tarball/abc1234",
		},
		{
			name: "github without .git suffix",
			url:  "https://github.com/hashicorp/terraform",
			ref:  "abc1234",
			want: "https://api.github.com/repos/hashicorp/terraform/tarball/abc1234",
		},
		{
			name: "gitlab",
			url:  "https://gitlab.com/myorg/myrepo.git",
			ref:  "def5678",
			want: "https://gitlab.com/api/v4/projects/myorg%2Fmyrepo/repository/archive.tar.gz?sha=def5678",
		},
		{
			name: "bitbucket",
			url:  "https://bitbucket.org/myorg/myrepo.git",
			ref:  "aaa1111",
			want: "https://bitbucket.org/myorg/myrepo/get/aaa1111.tar.gz",
		},
		{
			name: "github via ssh",
			url:  "ssh://git@github.com/hashicorp/terraform.git",
			ref:  "abc1234",
			want: "https://api.github.com/repos/hashicorp/terraform/tarball/abc1234",
		},
		{
			name: "github via http",
			url:  "http://github.com/hashicorp/terraform.git",
			ref:  "abc1234",
			want: "https://api.github.com/repos/hashicorp/terraform/tarball/abc1234",
		},
		{
			name: "github via git scheme",
			url:  "git://github.com/hashicorp/terraform.git",
			ref:  "abc1234",
			want: "https://api.github.com/repos/hashicorp/terraform/tarball/abc1234",
		},
		{
			name: "gitlab via ssh",
			url:  "ssh://git@gitlab.com/myorg/myrepo.git",
			ref:  "def5678",
			want: "https://gitlab.com/api/v4/projects/myorg%2Fmyrepo/repository/archive.tar.gz?sha=def5678",
		},
		{
			name: "bitbucket via ssh",
			url:  "ssh://git@bitbucket.org/myorg/myrepo.git",
			ref:  "aaa1111",
			want: "https://bitbucket.org/myorg/myrepo/get/aaa1111.tar.gz",
		},
		{
			name:    "unsupported host",
			url:     "https://example.com/owner/repo.git",
			ref:     "abc1234",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := url.Parse(tt.url)
			if err != nil {
				t.Fatalf("failed to parse URL %q: %v", tt.url, err)
			}

			got, err := archiveURL(u, tt.ref)
			if (err != nil) != tt.wantErr {
				t.Fatalf("archiveURL() err = %v, wantErr = %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("archiveURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseOwnerRepo(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		wantOwner string
		wantRepo  string
		wantErr   bool
	}{
		{
			name:      "with .git suffix",
			path:      "/hashicorp/terraform.git",
			wantOwner: "hashicorp",
			wantRepo:  "terraform",
		},
		{
			name:      "without .git suffix",
			path:      "/hashicorp/terraform",
			wantOwner: "hashicorp",
			wantRepo:  "terraform",
		},
		{
			name:      "extra path segments are ignored",
			path:      "/org/repo/extra/path",
			wantOwner: "org",
			wantRepo:  "repo",
		},
		{
			name:    "single segment is invalid",
			path:    "/onlyone",
			wantErr: true,
		},
		{
			name:    "empty path is invalid",
			path:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, err := parseOwnerRepo(tt.path)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseOwnerRepo(%q): err=%v, wantErr=%v", tt.path, err, tt.wantErr)
			}
			if owner != tt.wantOwner || repo != tt.wantRepo {
				t.Errorf("parseOwnerRepo(%q) = (%q, %q), want (%q, %q)", tt.path, owner, repo, tt.wantOwner, tt.wantRepo)
			}
		})
	}
}

func TestTokenFromEnv(t *testing.T) {
	tests := []struct {
		name string
		host string
		vars map[string]string
		want string
	}{
		{
			name: "GH_TOKEN",
			host: "github.com",
			vars: map[string]string{"GH_TOKEN": "ghtoken123"},
			want: "ghtoken123",
		},
		{
			name: "GITHUB_TOKEN",
			host: "github.com",
			vars: map[string]string{"GITHUB_TOKEN": "ghtoken456"},
			want: "ghtoken456",
		},
		{
			name: "GH_TOKEN takes precedence over GITHUB_TOKEN",
			host: "github.com",
			vars: map[string]string{"GH_TOKEN": "primary", "GITHUB_TOKEN": "secondary"},
			want: "primary",
		},
		{
			name: "GITLAB_TOKEN",
			host: "gitlab.com",
			vars: map[string]string{"GITLAB_TOKEN": "gltoken123"},
			want: "gltoken123",
		},
		{
			name: "GL_TOKEN",
			host: "gitlab.com",
			vars: map[string]string{"GL_TOKEN": "gltoken456"},
			want: "gltoken456",
		},
		{
			name: "GITLAB_TOKEN takes precedence over GL_TOKEN",
			host: "gitlab.com",
			vars: map[string]string{"GITLAB_TOKEN": "primary", "GL_TOKEN": "secondary"},
			want: "primary",
		},
		{
			name: "BITBUCKET_TOKEN",
			host: "bitbucket.org",
			vars: map[string]string{"BITBUCKET_TOKEN": "bbtoken"},
			want: "bbtoken",
		},
		{
			name: "unsupported host",
			host: "example.com",
			vars: map[string]string{},
			want: "",
		},
		{
			name: "github with port",
			host: "github.com:443",
			vars: map[string]string{"GH_TOKEN": "tok"},
			want: "tok",
		},
	}

	// Clear all relevant env vars before each subtest.
	envVars := []string{"GH_TOKEN", "GITHUB_TOKEN", "GITLAB_TOKEN", "GL_TOKEN", "BITBUCKET_TOKEN"}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, k := range envVars {
				t.Setenv(k, "")
			}
			for k, v := range tt.vars {
				t.Setenv(k, v)
			}

			got := tokenFromEnv(tt.host)
			if got != tt.want {
				t.Errorf("tokenFromEnv(%q) = %q, want %q", tt.host, got, tt.want)
			}
		})
	}
}

// buildTestTarGz creates a .tar.gz archive in memory with the given files
// nested under a top-level directory (mimicking GitHub/GitLab/Bitbucket
// archive layout). The files map is relative path → content.
func buildTestTarGz(t *testing.T, topDir string, files map[string]string) []byte {
	t.Helper()

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	// Write top-level directory entry.
	if err := tw.WriteHeader(&tar.Header{
		Name:     topDir + "/",
		Typeflag: tar.TypeDir,
		Mode:     0755,
	}); err != nil {
		t.Fatal(err)
	}

	for path, content := range files {
		fullPath := topDir + "/" + path

		// Create parent directory entries.
		dir := filepath.Dir(fullPath)
		if dir != topDir {
			if err := tw.WriteHeader(&tar.Header{
				Name:     dir + "/",
				Typeflag: tar.TypeDir,
				Mode:     0755,
			}); err != nil {
				t.Fatal(err)
			}
		}

		if err := tw.WriteHeader(&tar.Header{
			Name:     fullPath,
			Size:     int64(len(content)),
			Typeflag: tar.TypeReg,
			Mode:     0644,
		}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}

	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// setArchiveOverride sets archiveURLOverride for the duration of the test
// and restores it when the test completes.
func setArchiveOverride(t *testing.T, url string) {
	t.Helper()
	old := archiveURLOverride
	archiveURLOverride = url
	t.Cleanup(func() { archiveURLOverride = old })
}

func TestFetchArchive(t *testing.T) {
	archive := buildTestTarGz(t, "repo-abc1234", map[string]string{
		"main.tf":       "resource \"null\" \"a\" {}",
		"modules/m.tf":  "resource \"null\" \"b\" {}",
		"README.md":     "hello",
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/gzip")
		_, _ = w.Write(archive)
	}))
	defer srv.Close()

	setArchiveOverride(t, srv.URL+"/test/repo/archive/abc1234.tar.gz")

	u, _ := url.Parse("https://github.com/test/repo.git")
	dst := filepath.Join(t.TempDir(), "dst")

	err := fetchArchive(context.Background(), dst, u, "abc1234", "")
	if err != nil {
		t.Fatalf("fetchArchive() error: %v", err)
	}

	// Verify all files were extracted.
	for _, file := range []string{"main.tf", "modules/m.tf", "README.md"} {
		path := filepath.Join(dst, file)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected file %q not found: %v", file, err)
		}
	}

	// Verify content.
	got, err := os.ReadFile(filepath.Join(dst, "main.tf"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "resource \"null\" \"a\" {}" {
		t.Errorf("main.tf content = %q, want %q", got, "resource \"null\" \"a\" {}")
	}
}

func TestFetchArchive_subdir(t *testing.T) {
	archive := buildTestTarGz(t, "repo-abc1234", map[string]string{
		"main.tf":       "root",
		"modules/m.tf":  "module",
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/gzip")
		_, _ = w.Write(archive)
	}))
	defer srv.Close()

	setArchiveOverride(t, srv.URL+"/test/repo/archive/abc1234.tar.gz")

	u, _ := url.Parse("https://github.com/test/repo.git")
	dst := filepath.Join(t.TempDir(), "dst")

	err := fetchArchive(context.Background(), dst, u, "abc1234", "modules")
	if err != nil {
		t.Fatalf("fetchArchive() error: %v", err)
	}

	// The subdir path should be preserved relative to dst.
	got, err := os.ReadFile(filepath.Join(dst, "modules", "m.tf"))
	if err != nil {
		t.Fatalf("expected modules/m.tf under dst: %v", err)
	}
	if string(got) != "module" {
		t.Errorf("m.tf content = %q, want %q", got, "module")
	}

	// Files outside the subdir should not be extracted.
	if _, err := os.Stat(filepath.Join(dst, "main.tf")); err == nil {
		t.Error("main.tf should not exist in dst when subdir is set")
	}
}

func TestFetchArchive_subdir_not_found(t *testing.T) {
	archive := buildTestTarGz(t, "repo-abc1234", map[string]string{
		"main.tf": "root",
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/gzip")
		_, _ = w.Write(archive)
	}))
	defer srv.Close()

	setArchiveOverride(t, srv.URL+"/test/repo/archive/abc1234.tar.gz")

	u, _ := url.Parse("https://github.com/test/repo.git")
	dst := filepath.Join(t.TempDir(), "dst")

	err := fetchArchive(context.Background(), dst, u, "abc1234", "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent subdir")
	}
}

func TestFetchArchive_http_error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	setArchiveOverride(t, srv.URL+"/test/repo/archive/abc1234.tar.gz")

	u, _ := url.Parse("https://github.com/test/repo.git")
	dst := filepath.Join(t.TempDir(), "dst")

	err := fetchArchive(context.Background(), dst, u, "abc1234", "")
	if err == nil {
		t.Fatal("expected error for HTTP 404")
	}
}

func TestFetchArchive_bearer_token(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		// Return a valid archive so fetchArchive doesn't error on extraction.
		archive := buildTestTarGz(t, "repo-abc1234", map[string]string{
			"main.tf": "hello",
		})
		w.Header().Set("Content-Type", "application/gzip")
		_, _ = w.Write(archive)
	}))
	defer srv.Close()

	setArchiveOverride(t, srv.URL+"/test/repo/archive/abc1234.tar.gz")
	t.Setenv("GH_TOKEN", "test-token-123")

	u, _ := url.Parse("https://github.com/test/repo.git")
	dst := filepath.Join(t.TempDir(), "dst")

	err := fetchArchive(context.Background(), dst, u, "abc1234", "")
	if err != nil {
		t.Fatalf("fetchArchive() error: %v", err)
	}

	want := "Bearer test-token-123"
	if gotAuth != want {
		t.Errorf("Authorization header = %q, want %q", gotAuth, want)
	}
}

func TestFetchArchive_basic_auth_from_url(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		archive := buildTestTarGz(t, "repo-abc1234", map[string]string{
			"main.tf": "hello",
		})
		w.Header().Set("Content-Type", "application/gzip")
		_, _ = w.Write(archive)
	}))
	defer srv.Close()

	setArchiveOverride(t, srv.URL+"/test/repo/archive/abc1234.tar.gz")

	u, _ := url.Parse("https://myuser:mytoken@github.com/test/repo.git")
	dst := filepath.Join(t.TempDir(), "dst")

	err := fetchArchive(context.Background(), dst, u, "abc1234", "")
	if err != nil {
		t.Fatalf("fetchArchive() error: %v", err)
	}

	if gotAuth == "" {
		t.Fatal("expected Authorization header to be set from URL userinfo")
	}
}

func TestFetchArchive_skips_git_ssh_user(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		archive := buildTestTarGz(t, "repo-abc1234", map[string]string{
			"main.tf": "hello",
		})
		w.Header().Set("Content-Type", "application/gzip")
		_, _ = w.Write(archive)
	}))
	defer srv.Close()

	setArchiveOverride(t, srv.URL+"/test/repo/archive/abc1234.tar.gz")
	// Clear env tokens so we can verify "git" user was skipped and no auth is set.
	t.Setenv("GH_TOKEN", "")
	t.Setenv("GITHUB_TOKEN", "")

	u, _ := url.Parse("ssh://git@github.com/test/repo.git")
	dst := filepath.Join(t.TempDir(), "dst")

	err := fetchArchive(context.Background(), dst, u, "abc1234", "")
	if err != nil {
		t.Fatalf("fetchArchive() error: %v", err)
	}

	if gotAuth != "" {
		t.Errorf("expected no Authorization header for git@ SSH user, got %q", gotAuth)
	}
}