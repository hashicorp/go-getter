package getter

import (
	"context"
	"net/url"
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
func TestHgGetter_HgArgumentsNotAllowed(t *testing.T) {
	if !testHasHg {
		t.Log("hg not found, skipping")
		t.Skip()
	}
	ctx := context.Background()

	tc := []struct {
		name   string
		req    Request
		errChk func(testing.TB, error)
	}{
		{
			// If arguments are allowed in the destination, this request to Get will fail
			name: "arguments allowed in destination",
			req: Request{
				Dst: "--config=alias.clone=!touch ./TEST",
				u:   testModuleURL("basic-hg"),
			},
			errChk: func(t testing.TB, err error) {
				if err != nil {
					t.Errorf("Expected no err, got: %s", err)
				}
			},
		},
		{
			// Test arguments passed into the `rev` parameter
			// This clone call will fail regardless, but an exit code of 1 indicates
			// that the `false` command executed
			// We are expecting an hg parse error
			name: "arguments passed into rev parameter",
			req: Request{
				u: testModuleURL("basic-hg?rev=--config=alias.update=!false"),
			},
			errChk: func(t testing.TB, err error) {
				if err == nil {
					return
				}

				if !strings.Contains(err.Error(), "hg: parse error") {
					t.Errorf("Expected no err, got: %s", err)
				}
			},
		},
		{
			// Test arguments passed in the repository URL
			// This Get call will fail regardless, but it should fail
			// because the repository can't be found.
			// Other failures indicate that hg interpreted the argument passed in the URL
			name: "arguments passed in the repository URL",
			req: Request{
				u: &url.URL{Path: "--config=alias.clone=false"},
			},
			errChk: func(t testing.TB, err error) {
				if err == nil {
					return
				}

				if !strings.Contains(err.Error(), "repository --config=alias.clone=false not found") {
					t.Errorf("Expected no err, got: %s", err)
				}
			},
		},
	}
	for _, tt := range tc {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			g := new(HgGetter)

			if tt.req.Dst == "" {
				dst := testing_helper.TempDir(t)
				tt.req.Dst = dst
			}

			defer os.RemoveAll(tt.req.Dst)
			err := g.Get(ctx, &tt.req)
			tt.errChk(t, err)
		})
	}
}
