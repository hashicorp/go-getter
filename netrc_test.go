package getter

import (
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

func TestAddAuthFromNetrc(t *testing.T) {
	defer tempEnv(t, "NETRC", "./testdata/netrc/basic")()

	u, err := url.Parse("http://example.com")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if err := addAuthFromNetrc(u); err != nil {
		t.Fatalf("err: %s", err)
	}

	expected := "http://foo:bar@example.com"
	actual := u.String()
	if expected != actual {
		t.Fatalf("Mismatch: %q != %q", actual, expected)
	}
}

func TestAddAuthFromNetrc_secret(t *testing.T) {
	dir := tempDir(t)
	if err := os.Mkdir(dir, 0755); err != nil {
		t.Fatalf("err: %s", err)
	}
	netrc := filepath.Join(dir, ".netrc")
	if _, err := os.Create(netrc); err != nil {
		t.Fatalf("err: %s", err)
	}
	if err := os.Chmod(netrc, 0000); err != nil {
		t.Fatalf("err: %s", err)
	}
	defer tempEnv(t, "NETRC", netrc)()
	defer os.RemoveAll(dir)

	u, err := url.Parse("http://example.com")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if err := addAuthFromNetrc(u); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestAddAuthFromNetrc_hasAuth(t *testing.T) {
	defer tempEnv(t, "NETRC", "./testdata/netrc/basic")()

	u, err := url.Parse("http://username:password@example.com")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	expected := u.String()
	if err := addAuthFromNetrc(u); err != nil {
		t.Fatalf("err: %s", err)
	}

	actual := u.String()
	if expected != actual {
		t.Fatalf("Mismatch: %q != %q", actual, expected)
	}
}

func TestAddAuthFromNetrc_hasUsername(t *testing.T) {
	defer tempEnv(t, "NETRC", "./testdata/netrc/basic")()

	u, err := url.Parse("http://username@example.com")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	expected := u.String()
	if err := addAuthFromNetrc(u); err != nil {
		t.Fatalf("err: %s", err)
	}

	actual := u.String()
	if expected != actual {
		t.Fatalf("Mismatch: %q != %q", actual, expected)
	}
}
