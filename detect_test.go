package getter

import (
	"fmt"
	"testing"
)

func TestDetect(t *testing.T) {
	cases := []struct {
		Input  string
		Pwd    string
		Output string
		g Getter
		Err    bool
	}{
	{"./foo", "/foo", "/foo/foo", new(FileGetter), false},
	//{"git::./foo", "/foo", "/foo/foo", new(GitGetter),false}, // TODO @sylviamoss understand this test. Is this a real situation?
		{
			"git::github.com/hashicorp/foo",
			"",
			"https://github.com/hashicorp/foo.git",
			new(GitGetter),
			false,
		},
		{
			"./foo//bar",
			"/foo",
			"/foo/foo//bar",
			new(FileGetter),
			false,
		},
		{
			"git::github.com/hashicorp/foo//bar",
			"",
			"https://github.com/hashicorp/foo.git//bar",
			new(GitGetter),
			false,
		},
		{
			"git::https://github.com/hashicorp/consul.git",
			"",
			"https://github.com/hashicorp/consul.git",
			new(GitGetter),
			false,
		},
		{
			"git::https://person@someothergit.com/foo/bar", // this
			"",
			"https://person@someothergit.com/foo/bar",
			new(GitGetter),
			false,
		},
		{
			"git::https://person@someothergit.com/foo/bar", // this
			"/bar",
			"https://person@someothergit.com/foo/bar",
			new(GitGetter),
			false,
		},
		{
			"./foo/archive//*",
			"/bar",
			"/bar/foo/archive//*",
			new(FileGetter),
			false,
		},

		// https://github.com/hashicorp/go-getter/pull/124
		{
			"git::ssh://git@my.custom.git/dir1/dir2",
			"",
			"ssh://git@my.custom.git/dir1/dir2",
			new(GitGetter),
			false,
		},
		{
			"git::git@my.custom.git:dir1/dir2",
			"/foo",
			"ssh://git@my.custom.git/dir1/dir2",
			new(GitGetter),
			false,
		},
		{
			"git::git@my.custom.git:dir1/dir2",
			"",
			"ssh://git@my.custom.git/dir1/dir2",
			new(GitGetter),
			false,
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d %s", i, tc.Input), func(t *testing.T) {
			output, ok, err := Detect(tc.Input, tc.Pwd, tc.g)
			if err != nil != tc.Err {
				t.Fatalf("%d: bad err: %s", i, err)
			}
			if !ok {
				t.Fatalf("%s url expected to valid", tc.Input)
			}
			if output != tc.Output {
				t.Fatalf("%d: bad output: %s\nexpected: %s", i, output, tc.Output)
			}
		})
	}
}
