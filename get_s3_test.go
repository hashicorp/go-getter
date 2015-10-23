package getter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awserr"
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

func TestS3Getter_subdir(t *testing.T) {
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

func TestS3Getter_GetFile_params(t *testing.T) {
	g := new(S3Getter)
	dst := tempFile(t)

	// Download
	err := g.GetFile(dst, testURL("https://s3-eu-west-1.amazonaws.com/hailo-s3-test/foo.txt?aws_access_key_id=foo&aws_access_key_secret=bar&aws_access_token=baz"))
	if err == nil {
		t.Fatalf("expected error, got none")
	}

	if reqerr, ok := err.(awserr.RequestFailure); !ok || reqerr.StatusCode() != 403 {
		t.Fatalf("expected InvalidAccessKeyId error")
	}
}
