package getter

import (
	"net/http"
	"strings"
	"testing"
)

const testCtxBBUrl = "https://bitbucket.org/hashicorp/tf-test-git"

func TestBitBucketCtxDetector(t *testing.T) {
	t.Parallel()

	if _, err := http.Get(testCtxBBUrl); err != nil {
		t.Log("internet may not be working, skipping BB tests")
		t.Skip()
	}

	cases := []struct {
		Input  string
		Output string
	}{
		// HTTP
		{
			"bitbucket.org/hashicorp/tf-test-git",
			"git::https://bitbucket.org/hashicorp/tf-test-git.git",
		},
		{
			"bitbucket.org/hashicorp/tf-test-git.git",
			"git::https://bitbucket.org/hashicorp/tf-test-git.git",
		},
		{
			"bitbucket.org/hashicorp/tf-test-hg",
			"hg::https://bitbucket.org/hashicorp/tf-test-hg",
		},
	}

	pwd := "/pwd"
	forceToken := ""
	ctxSubDir := ""
	srcResolveFrom := ""

	f := new(BitBucketCtxDetector)
	for i, tc := range cases {
		var err error
		for i := 0; i < 3; i++ {
			var output string
			var ok bool
			output, ok, err = f.CtxDetect(tc.Input, pwd, forceToken, ctxSubDir, srcResolveFrom)
			if err != nil {
				if strings.Contains(err.Error(), "invalid character") {
					continue
				}

				t.Fatalf("err: %s", err)
			}
			if !ok {
				t.Fatal("not ok")
			}

			if output != tc.Output {
				t.Fatalf("%d: bad: %#v", i, output)
			}

			break
		}
		if i >= 3 {
			t.Fatalf("failure from bitbucket: %s", err)
		}
	}
}