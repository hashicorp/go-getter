package getter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGet_badSchema(t *testing.T) {
	dst := tempDir(t)
	u := testModule("basic")
	u = strings.Replace(u, "file", "nope", -1)

	if err := Get(dst, u); err == nil {
		t.Fatal("should error")
	}
}

func TestGet_file(t *testing.T) {
	dst := tempDir(t)
	u := testModule("basic")

	if err := Get(dst, u); err != nil {
		t.Fatalf("err: %s", err)
	}

	mainPath := filepath.Join(dst, "main.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestGet_fileDetect(t *testing.T) {
	dst := tempDir(t)
	u := filepath.Join("./test-fixtures", "basic")
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	client := &Client{
		Src: u,
		Dst: dst,
		Pwd: pwd,
		Dir: true,
	}

	if err := client.Get(); err != nil {
		t.Fatalf("err: %s", err)
	}

	mainPath := filepath.Join(dst, "main.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestGet_fileForced(t *testing.T) {
	dst := tempDir(t)
	u := testModule("basic")
	u = "file::" + u

	if err := Get(dst, u); err != nil {
		t.Fatalf("err: %s", err)
	}

	mainPath := filepath.Join(dst, "main.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestGet_fileSubdir(t *testing.T) {
	dst := tempDir(t)
	u := testModule("basic//subdir")

	if err := Get(dst, u); err != nil {
		t.Fatalf("err: %s", err)
	}

	mainPath := filepath.Join(dst, "sub.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}
