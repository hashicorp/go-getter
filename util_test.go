package getter

import (
	"os"
	"testing"
)

// tempEnv sets the env var temporarily and returns a function that should
// be deferred to clean it up.
func tempEnv(t *testing.T, k, v string) func() {
	old := os.Getenv(k)

	// Set env
	if err := os.Setenv(k, v); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Easy cleanup
	return func() {
		if err := os.Setenv(k, old); err != nil {
			t.Fatalf("err: %s", err)
		}
	}
}
