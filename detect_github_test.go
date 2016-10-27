package getter

import (
	"testing"
)

func TestGitHubDetector(t *testing.T) {
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
		{
			"github.xyz.com/hashicorp/terraform.git",
			"git::https://github.xyz.com/hashicorp/terraform.git",
		},
		{
			"github.xyz.com/hashicorp/terraform",
			"git::https://github.xyz.com/hashicorp/terraform.git",
		},
		{
			"github.xyz.com/hashicorp/terraform.git?ref=test-branch",
			"git::https://github.xyz.com/hashicorp/terraform.git?ref=test-branch",
		},
		{
			"github.xyz.com/hashicorp/terraform.git//modules/a",
			"git::https://github.xyz.com/hashicorp/terraform.git///modules/a",
		},
		{
			"github.xyz.com/hashicorp/terraform//modules/a",
			"git::https://github.xyz.com/hashicorp/terraform.git///modules/a",
		},

		// SSH
		{"git@github.com:hashicorp/foo.git", "git::ssh://git@github.com/hashicorp/foo.git"},
		{
			"git@github.com:org/project.git?ref=test-branch",
			"git::ssh://git@github.com/org/project.git?ref=test-branch",
		},
		{
			"git@github.com:hashicorp/foo.git//bar",
			"git::ssh://git@github.com/hashicorp/foo.git//bar",
		},
		{
			"git@github.com:org/project.git//module/a?ref=test-branch",
			"git::ssh://git@github.com/org/project.git//module/a?ref=test-branch",
		},
		{
			"git@github.xyz.com:org/project.git", 
			"git::ssh://git@github.xyz.com/org/project.git", 
		},
		{
			"git@github.xyz.com:org/project.git?ref=test-branch",
			"git::ssh://git@github.xyz.com/org/project.git?ref=test-branch",
		},
		{ 
			"git@github.xyz.com:org/project.git//module/a",
			"git::ssh://git@github.xyz.com/org/project.git//module/a",
		},
		{
			"git@github.xyz.com:org/project.git//module/a?ref=test-branch", 
			"git::ssh://git@github.xyz.com/org/project.git//module/a?ref=test-branch", 
		},
	}

	pwd := "/pwd"
	f := new(GitHubDetector)
	for i, tc := range cases {
		output, ok, err := f.Detect(tc.Input, pwd)
		if err != nil {
			t.Fatalf("Idx: %d input: '%s' expected '%s' err: %s", i, tc.Input, tc.Output, err)
		}
		if !ok {
			t.Fatalf("Idx: %d input: '%s' expected '%s' not ok", i, tc.Input, tc.Output,)
		}

		if output != tc.Output {
			t.Fatalf("Idx: %d input: '%s' expected '%s', got '%s'", i, tc.Input, tc.Output, output)
		}
	}
}
