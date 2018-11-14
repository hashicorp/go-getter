package getter

import (
	"fmt"
	"testing"
)

func TestGitDetector(t *testing.T) {
	cases := []struct {
		Input  string
		Output string
	}{
		// HTTP
		{
			"github.com/hashicorp/foo",
			"git::https://github.com/hashicorp/foo.git",
		},
		{
			"github.com/hashicorp/foo.git",
			"git::https://github.com/hashicorp/foo.git",
		},
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

		// SSH
		{
			"git@github.com:hashicorp/foo.git",
			"git::ssh://git@github.com/hashicorp/foo.git",
		},
		{
			"git@github.com:hashicorp/foo.git//bar",
			"git::ssh://git@github.com/hashicorp/foo.git//bar",
		},
		{
			"git@github.com:hashicorp/foo.git?foo=bar",
			"git::ssh://git@github.com/hashicorp/foo.git?foo=bar",
		},
		// non-github SSH
		{
			"git@ssh.dev.azure.com:v3/paul0787/terraform-null-test/terraform-null-test",
			"git::ssh://git@ssh.dev.azure.com/v3/paul0787/terraform-null-test/terraform-null-test",
		},
		{
			"git@my.custom.git:dir1/dir2",
			"git::ssh://git@my.custom.git/dir1/dir2",
		},
	}

	pwd := "/pwd"
	f := new(GitDetector)
	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d %s", i, tc.Input), func(t *testing.T) {
			output, ok, err := f.Detect(tc.Input, pwd)
			if err != nil {
				t.Fatalf("err: %s", err)
			}
			if !ok {
				t.Fatal("not ok")
			}

			if output != tc.Output {
				t.Fatalf("%d: bad: %#v", i, output)
			}
		})
	}
}
