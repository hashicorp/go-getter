package getter

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	testing_helper "github.com/hashicorp/go-getter/v2/helper/testing"
)

var testHasHg bool

func init() {
	if _, err := exec.LookPath("hg"); err == nil {
		testHasHg = true
	}
}

func TestHgGetter_impl(t *testing.T) {
	var _ Getter = new(HgGetter)
}

func TestHgGetter(t *testing.T) {
	if !testHasHg {
		t.Log("hg not found, skipping")
		t.Skip()
	}
	ctx := context.Background()

	g := new(HgGetter)
	dst := testing_helper.TempDir(t)

	req := &Request{
		Dst: dst,
		u:   testModuleURL("basic-hg"),
	}

	// With a dir that doesn't exist
	if err := g.Get(ctx, req); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	mainPath := filepath.Join(dst, "main.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestHgGetter_branch(t *testing.T) {
	if !testHasHg {
		t.Log("hg not found, skipping")
		t.Skip()
	}
	ctx := context.Background()

	g := new(HgGetter)
	dst := testing_helper.TempDir(t)

	url := testModuleURL("basic-hg")
	q := url.Query()
	q.Add("rev", "test-branch")
	url.RawQuery = q.Encode()

	req := &Request{
		Dst: dst,
		u:   url,
	}

	if err := g.Get(ctx, req); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	mainPath := filepath.Join(dst, "main_branch.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Get again should work
	if err := g.Get(ctx, req); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	mainPath = filepath.Join(dst, "main_branch.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestHgGetter_GetFile(t *testing.T) {
	if !testHasHg {
		t.Log("hg not found, skipping")
		t.Skip()
	}
	ctx := context.Background()

	g := new(HgGetter)
	dst := testing_helper.TempTestFile(t)
	defer os.RemoveAll(filepath.Dir(dst))

	req := &Request{
		Dst: dst,
		u:   testModuleURL("basic-hg/foo.txt"),
	}

	// Download
	if err := g.GetFile(ctx, req); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	if _, err := os.Stat(dst); err != nil {
		t.Fatalf("err: %s", err)
	}
	testing_helper.AssertContents(t, dst, "Hello\n")
}

const testBBUrl = "https://bitbucket.org/hashicorp/tf-test-git"

func TestHgGetter_DetectBitBucketDetector(t *testing.T) {
	t.Parallel()

	if _, err := http.Get(testBBUrl); err != nil {
		t.Log("internet may not be working, skipping BB tests")
		t.Skip()
	}

	cases := []struct {
		Input  string
		Output string
	}{
		{
			"bitbucket.org/hashicorp/tf-test-hg",
			"https://bitbucket.org/hashicorp/tf-test-hg",
		},
	}

	pwd := "/pwd"
	for i, tc := range cases {
		var err error
		for i := 0; i < 3; i++ {
			var ok bool
			req := &Request{
				Src: tc.Input,
				Pwd: pwd,
			}
			ok, err = Detect(req, new(HgGetter))
			if err != nil {
				if strings.Contains(err.Error(), "invalid character") {
					continue
				}

				t.Fatalf("err: %s", err)
			}
			if !ok {
				t.Fatal("not ok")
			}

			if req.Src != tc.Output {
				t.Fatalf("%d: bad: %#v", i, req.Src)
			}

			break
		}
		if i >= 3 {
			t.Fatalf("failure from bitbucket: %s", err)
		}
	}
}