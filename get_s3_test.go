package getter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestS3Getter_impl(t *testing.T) {
	var _ Getter = new(S3Getter)
}

func TestS3Getter(t *testing.T) {
	g := new(S3Getter)
	dst := tempDir(t)

	// With a dir that doesn't exist
	if err := g.Get(dst, testURL("https://s3-eu-west-1.amazonaws.com/hailo-s3-test")); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	mainPath := filepath.Join(dst, "main.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestS3Getter_Subdir(t *testing.T) {
	g := new(S3Getter)
	dst := tempDir(t)

	// With a dir that doesn't exist
	if err := g.Get(dst, testURL("https://s3-eu-west-1.amazonaws.com/hailo-s3-test/subdir")); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	subPath := filepath.Join(dst, "sub.tf")
	if _, err := os.Stat(subPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestS3Getter_GetFile(t *testing.T) {
	g := new(S3Getter)
	dst := tempFile(t)

	// Download
	if err := g.GetFile(dst, testURL("https://s3-eu-west-1.amazonaws.com/hailo-s3-test/foo.txt")); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	if _, err := os.Stat(dst); err != nil {
		t.Fatalf("err: %s", err)
	}
	assertContents(t, dst, "Hello\n")
}
