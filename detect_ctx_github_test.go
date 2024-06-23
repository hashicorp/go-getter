package getter

import (
	"testing"
)

func TestGitHubCtxDetector(t *testing.T) {
	cases := []struct {
		Input  string
		Output string
	}{
		// HTTP
		{"github.com/hashicorp/foo", "git::https://github.com/hashicorp/foo.git"},
		{"github.com/hashicorp/foo.git", "git::https://github.com/hashicorp/foo.git"},
		{
			"github.com/hashicorp/foo/bar",
			"git::https://github.com/hashicorp/foo.git//bar",
		},
		{
			"github.com/hashicorp/foo?foo=bar",
			"git::https://github.com/hashicorp/foo.git?foo=bar",
		},
		{
			"github.com/hashicorp/foo.git?foo=bar",
			"git::https://github.com/hashicorp/foo.git?foo=bar",
		},
	}

	pwd := "/pwd"
	forceToken := ""
	ctxSubDir := ""
	srcResolveFrom := ""

	f := new(GitHubCtxDetector)
	for i, tc := range cases {
		output, ok, err := f.CtxDetect(tc.Input, pwd, forceToken, ctxSubDir, srcResolveFrom)
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if !ok {
			t.Fatal("not ok")
		}

		if output != tc.Output {
			t.Fatalf("%d: bad: %#v", i, output)
		}
	}
}
