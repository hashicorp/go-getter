package getter

import (
	"context"
	testing_helper "github.com/hashicorp/go-getter/v2/helper/testing"
	"os"
	"path/filepath"
	"testing"
)

const basicMainTFExpectedContents = `# Hello

module "foo" {
    source = "./foo"
}
`

func TestGitGetter_githubDirWithModeAny(t *testing.T) {
	if !testHasGit {
		t.Skip("git not found, skipping")
	}

	ctx := context.Background()
	dst := testing_helper.TempDir(t)
	defer os.RemoveAll(dst)

	req := &Request{
		Src:     "git::https://github.com/arikkfir/go-getter.git//testdata/basic?ref=v2",
		Dst:     dst,
		GetMode: ModeAny,
		Copy:    true,
	}
	client := Client{}
	result, err := client.Get(ctx, req)
	if err != nil {
		t.Fatalf("Failed fetching GitHub directory: %s", err)
	} else if stat, err := os.Stat(result.Dst); err != nil {
		t.Fatalf("Failed stat dst at '%s': %s", result.Dst, err)
	} else if !stat.IsDir() {
		t.Fatalf("Expected '%s' to be a directory", result.Dst)
	} else if entries, err := os.ReadDir(result.Dst); err != nil {
		t.Fatalf("Failed listing directory '%s': %s", result.Dst, err)
	} else if len(entries) != 3 {
		t.Fatalf("Expected dir '%s' to contain 3 items: %s", result.Dst, err)
	} else {
		testing_helper.AssertContents(t, filepath.Join(result.Dst, "main.tf"), basicMainTFExpectedContents)
	}
}

func TestGitGetter_githubFileWithModeAny(t *testing.T) {
	if !testHasGit {
		t.Skip("git not found, skipping")
	}

	ctx := context.Background()
	dst := testing_helper.TempDir(t)
	defer os.RemoveAll(dst)

	req := &Request{
		Src:     "git::https://github.com/arikkfir/go-getter.git//testdata/basic/main.tf?ref=v2",
		Dst:     dst,
		GetMode: ModeAny,
		Copy:    true,
	}
	client := Client{}
	result, err := client.Get(ctx, req)
	if err != nil {
		t.Fatalf("Failed fetching GitHub file: %s", err)
	} else if stat, err := os.Stat(result.Dst); err != nil {
		t.Fatalf("Failed stat dst at '%s': %s", result.Dst, err)
	} else if stat.IsDir() {
		t.Fatalf("Expected '%s' to be a file", result.Dst)
	} else {
		testing_helper.AssertContents(t, result.Dst, basicMainTFExpectedContents)
	}
}
