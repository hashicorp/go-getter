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
		Err    bool
		getter Getter
	}{
		{"./foo", "/foo", "/foo/foo", false, new(FileGetter)},
		{"git::./foo", "/foo", "/foo/foo", false, new(GitGetter)},
		{
			"git::github.com/hashicorp/foo",
			"",
			"https://github.com/hashicorp/foo.git",
			false,
			new(GitGetter),
		},
		{
			"./foo//bar",
			"/foo",
			"/foo/foo//bar",
			false,
			new(FileGetter),
		},
		{
			"git::github.com/hashicorp/foo//bar",
			"",
			"https://github.com/hashicorp/foo.git//bar",
			false,
			new(GitGetter),
		},
		{
			"git::https://github.com/hashicorp/consul.git",
			"",
			"https://github.com/hashicorp/consul.git",
			false,
			new(GitGetter),
		},
		{
			"git::https://person@someothergit.com/foo/bar", // this
			"",
			"https://person@someothergit.com/foo/bar",
			false,
			new(GitGetter),
		},
		{
			"git::https://person@someothergit.com/foo/bar", // this
			"/bar",
			"https://person@someothergit.com/foo/bar",
			false,
			new(GitGetter),
		},
		{
			"./foo/archive//*",
			"/bar",
			"/bar/foo/archive//*",
			false,
			new(FileGetter),
		},

		// https://github.com/hashicorp/go-getter/pull/124
		{
			"git::ssh://git@my.custom.git/dir1/dir2",
			"",
			"ssh://git@my.custom.git/dir1/dir2",
			false,
			new(GitGetter),
		},
		{
			"git::git@my.custom.git:dir1/dir2",
			"/foo",
			"ssh://git@my.custom.git/dir1/dir2",
			false,
			new(GitGetter),
		},
		{
			"git::git@my.custom.git:dir1/dir2",
			"",
			"ssh://git@my.custom.git/dir1/dir2",
			false,
			new(GitGetter),
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d %s", i, tc.Input), func(t *testing.T) {
			req := &Request{
				Src: tc.Input,
				Pwd: tc.Pwd,
			}
			_, err := Detect(req, tc.getter)
			if err != nil != tc.Err {
				t.Fatalf("%d: bad err: %s", i, err)
			}
			if req.Src != tc.Output {
				t.Fatalf("%d: bad output: %s\nexpected: %s", i, req.Src, tc.Output)
			}
		})
	}
}
