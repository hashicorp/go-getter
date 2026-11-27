package getter

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
)

// Reproducer for #649: when resolveCheckoutRef fails for an environmental
// reason (here: dst is not a git repository), the returned error should surface
// the underlying git message, not just a generic "invalid ref".
func TestResolveCheckoutRef_SurfacesGitError(t *testing.T) {
	if !testHasGit {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	// Stop git from finding a repository above the temp dir (TMPDIR may be
	// inside a checkout in CI or dev containers).
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(dir))

	_, err := resolveCheckoutRef(context.Background(), dir, "v1.2.3")
	if err == nil {
		t.Fatal("expected an error resolving a ref in a non-repository dir")
	}
	t.Logf("error: %v", err)
	// Before the fix the message is exactly `invalid ref: "v1.2.3"` with no
	// underlying cause. After the fix it also carries git's own message
	// (e.g. "not a git repository").
	low := strings.ToLower(err.Error())
	if !strings.Contains(low, "git") && !strings.Contains(low, "repository") {
		t.Fatalf("error does not surface the underlying git cause: %v", err)
	}
}

// Errors that are not a git exit status, such as a cancelled context, should
// be surfaced too.
func TestResolveCheckoutRef_SurfacesNonExitError(t *testing.T) {
	if !testHasGit {
		t.Skip("git not available")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := resolveCheckoutRef(ctx, t.TempDir(), "v1.2.3")
	if err == nil {
		t.Fatal("expected an error with a cancelled context")
	}
	if !strings.Contains(err.Error(), context.Canceled.Error()) {
		t.Fatalf("error does not surface the cancelled context: %v", err)
	}
}
