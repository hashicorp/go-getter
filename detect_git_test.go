package getter

import (
	"testing"
)

func TestGitDetector(t *testing.T) {
	cases := []struct {
		Input  string
		Output string
	}{
		{
			"git@github.com:hashicorp/foo.git",
			"ssh://git@github.com/hashicorp/foo.git",
		},
		{
			"git@github.com:org/project.git?ref=test-branch",
			"ssh://git@github.com/org/project.git?ref=test-branch",
		},
		{
			"git@github.com:hashicorp/foo.git//bar",
			"ssh://git@github.com/hashicorp/foo.git//bar",
		},
		{
			"git@github.com:hashicorp/foo.git?foo=bar",
			"ssh://git@github.com/hashicorp/foo.git?foo=bar",
		},
		{
			"git@github.xyz.com:org/project.git",
			"ssh://git@github.xyz.com/org/project.git",
		},
		{
			"git@github.xyz.com:org/project.git?ref=test-branch",
			"ssh://git@github.xyz.com/org/project.git?ref=test-branch",
		},
		{
			"git@github.xyz.com:org/project.git//module/a",
			"ssh://git@github.xyz.com/org/project.git//module/a",
		},
		{
			"git@github.xyz.com:org/project.git//module/a?ref=test-branch",
			"ssh://git@github.xyz.com/org/project.git//module/a?ref=test-branch",
		},
		{
			// Already in the canonical form, so no rewriting required
			// When the ssh: protocol is used explicitly, we recognize it as
			// URL form rather than SCP-like form, so the part after the colon
			// is a port number, not part of the path.

			// TODO @sylviamoss rewrite comment
			// No need to set git scheme anymore
			"git::ssh://git@git.example.com:2222/hashicorp/foo.git",
			"ssh://git@git.example.com:2222/hashicorp/foo.git",
		},
	}

	pwd := "/pwd"
	for _, tc := range cases {
		t.Run(tc.Input, func(t *testing.T) {
			detector := NewGetterDetector([]Getter{new(GitGetter)})
			output, err := detector.Detect(tc.Input, pwd)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if output != tc.Output {
				t.Errorf("wrong result\ninput: %s\ngot:   %s\nwant:  %s", tc.Input, output, tc.Output)
			}
		})
	}
}
