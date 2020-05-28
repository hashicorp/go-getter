package getter

import (
	"fmt"
	"testing"
)

func TestDetect(t *testing.T) {
	gitGetter := &GitGetter{[]Detector{
		new(GitDetector),
		new(BitBucketDetector),
		new(GitHubDetector),
	},
	}
	cases := []struct {
		Input  string
		Pwd    string
		Output string
		Err    bool
		getter Getter
	}{
		{"./foo", "/foo", "/foo/foo", false, new(FileGetter)},
		{"git::./foo", "/foo", "/foo/foo", false, gitGetter},
		{
			"git::github.com/hashicorp/foo",
			"",
			"https://github.com/hashicorp/foo.git",
			false,
			gitGetter,
		},
		{
			"./foo",
			"/foo",
			"/foo/foo",
			false,
			new(FileGetter),
		},
		{
			"git::https://github.com/hashicorp/consul.git",
			"",
			"https://github.com/hashicorp/consul.git",
			false,
			gitGetter,
		},
		{
			"git::https://person@someothergit.com/foo/bar",
			"",
			"https://person@someothergit.com/foo/bar",
			false,
			gitGetter,
		},
		{
			"git::https://person@someothergit.com/foo/bar",
			"/bar",
			"https://person@someothergit.com/foo/bar",
			false,
			gitGetter,
		},

		// https://github.com/hashicorp/go-getter/pull/124
		{
			"git::ssh://git@my.custom.git/dir1/dir2",
			"",
			"ssh://git@my.custom.git/dir1/dir2",
			false,
			gitGetter,
		},
		{
			"git::git@my.custom.git:dir1/dir2",
			"/foo",
			"ssh://git@my.custom.git/dir1/dir2",
			false,
			gitGetter,
		},
		{
			"git::git@my.custom.git:dir1/dir2",
			"",
			"ssh://git@my.custom.git/dir1/dir2",
			false,
			gitGetter,
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d %s", i, tc.Input), func(t *testing.T) {
			req := &Request{
				Src: tc.Input,
				Pwd: tc.Pwd,
			}
			ok, err := Detect(req, tc.getter)
			if err != nil != tc.Err {
				t.Fatalf("%d: bad err: %s", i, err)
			}

			if !tc.Err && !ok {
				t.Fatalf("%d: should be ok", i)
			}

			if req.Src != tc.Output {
				t.Fatalf("%d: bad output: %s\nexpected: %s", i, req.Src, tc.Output)
			}
		})
	}
}
