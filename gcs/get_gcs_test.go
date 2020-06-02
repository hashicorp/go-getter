package gcs

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/go-getter/v2"

	testing_helper "github.com/hashicorp/go-getter/v2/helper/testing"
	urlhelper "github.com/hashicorp/go-getter/v2/helper/url"
)

// initGCPCredentials writes a temporary GCS credentials file if necessary and
// returns the path and a function to clean it up. allAuthenticatedUsers can
// access go-getter-test with read only access.
func initGCPCredentials(t *testing.T) func() {
	if gc := os.Getenv("GOOGLE_CREDENTIALS"); gc != "" &&
		os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		file, cleanup := testing_helper.TempFileWithContent(t, gc)
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", file)
		return func() {
			os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "")
			cleanup()
		}
	}
	return func() {}
}

func TestGetter_impl(t *testing.T) {
	var _ getter.Getter = new(Getter)
}

func TestGetter(t *testing.T) {
	defer initGCPCredentials(t)()

	g := new(Getter)
	dst := testing_helper.TempDir(t)
	ctx := context.Background()

	req := &getter.Request{
		Src: "www.googleapis.com/storage/v1/go-getter-test/go-getter/folder",
		Dst: dst,
	}

	c := getter.Client{
		Getters: []getter.Getter{g},
	}

	// With a dir that doesn't exist
	_, err := c.Get(ctx, req)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	mainPath := filepath.Join(dst, "main.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestGetter_subdir(t *testing.T) {
	defer initGCPCredentials(t)()

	g := new(Getter)
	dst := testing_helper.TempDir(t)
	ctx := context.Background()

	req := &getter.Request{
		Src: "www.googleapis.com/storage/v1/go-getter-test/go-getter/folder/subfolder",
		Dst: dst,
	}

	c := getter.Client{
		Getters: []getter.Getter{g},
	}

	// With a dir that doesn't exist
	_, err := c.Get(ctx, req)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the sub file exists
	subPath := filepath.Join(dst, "sub.tf")
	if _, err := os.Stat(subPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestGetter_GetFile(t *testing.T) {
	defer initGCPCredentials(t)()

	g := new(Getter)
	dst := testing_helper.TempTestFile(t)
	defer os.RemoveAll(filepath.Dir(dst))
	ctx := context.Background()

	req := &getter.Request{
		Src:  "www.googleapis.com/storage/v1/go-getter-test/go-getter/folder/main.tf",
		Dst:  dst,
		Mode: getter.ModeFile,
	}

	c := getter.Client{
		Getters: []getter.Getter{g},
	}

	// Download
	_, err := c.Get(ctx, req)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	if _, err := os.Stat(dst); err != nil {
		t.Fatalf("err: %s", err)
	}
	testing_helper.AssertContents(t, dst, "# Main\n")
}

func TestGetter_GetFile_notfound(t *testing.T) {
	g := new(Getter)
	dst := testing_helper.TempTestFile(t)
	defer os.RemoveAll(filepath.Dir(dst))
	ctx := context.Background()

	req := &getter.Request{
		Src: "https://www.googleapis.com/storage/v1/go-getter-test/go-getter/folder/404.tf",
		Dst: dst,
	}

	c := getter.Client{
		Getters: []getter.Getter{g},
	}

	// Download
	_, err := c.Get(ctx, req)
	if err == nil {
		t.Fatalf("expected error, got none")
	}
}

func TestGetter_Mode_dir(t *testing.T) {
	defer initGCPCredentials(t)()

	g := new(Getter)
	ctx := context.Background()

	// Check client mode on a key prefix with only a single key.
	mode, err := g.Mode(ctx,
		urlhelper.MustParse("https://www.googleapis.com/storage/v1/go-getter-test/go-getter/folder/subfolder"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if mode != getter.ModeDir {
		t.Fatal("expect ModeDir")
	}
}

func TestGetter_Mode_file(t *testing.T) {
	defer initGCPCredentials(t)()

	g := new(Getter)
	ctx := context.Background()

	// Check client mode on a key prefix which contains sub-keys.
	mode, err := g.Mode(ctx,
		urlhelper.MustParse("https://www.googleapis.com/storage/v1/go-getter-test/go-getter/folder/subfolder/sub.tf"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if mode != getter.ModeFile {
		t.Fatal("expect ModeFile")
	}
}

func TestGetter_Mode_notfound(t *testing.T) {
	defer initGCPCredentials(t)()

	g := new(Getter)
	ctx := context.Background()

	// Check the client mode when a non-existent key is looked up. This does not
	// return an error, but rather should just return the file mode.
	mode, err := g.Mode(ctx,
		urlhelper.MustParse("https://www.googleapis.com/storage/v1/go-getter-test/go-getter/foobar"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if mode != getter.ModeFile {
		t.Fatal("expect ModeFile")
	}
}

func TestGetter_Url(t *testing.T) {
	defer initGCPCredentials(t)()

	var gcstests = []struct {
		name   string
		url    string
		bucket string
		path   string
	}{
		{
			name:   "test1",
			url:    "https://www.googleapis.com/storage/v1/go-getter-test/go-getter/foo/null.zip",
			bucket: "go-getter-test",
			path:   "go-getter/foo/null.zip",
		},
	}

	for i, pt := range gcstests {
		t.Run(pt.name, func(t *testing.T) {
			g := new(Getter)
			src := pt.url
			u, err := url.Parse(src)

			if err != nil {
				t.Errorf("test %d: unexpected error: %s", i, err)
			}

			bucket, path, err := g.parseURL(u)

			if err != nil {
				t.Fatalf("err: %s", err)
			}

			if bucket != pt.bucket {
				t.Fatalf("expected %s, got %s", pt.bucket, bucket)
			}
			if path != pt.path {
				t.Fatalf("expected %s, got %s", pt.path, path)
			}
		})
	}
}