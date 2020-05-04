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
	}{
		{"./foo", "/foo", "/foo/foo", false},
		{"git::./foo", "/foo", "/foo/foo", false},
		{
			"git::github.com/hashicorp/foo",
			"",
			"https://github.com/hashicorp/foo.git",
			false,
		},
		{
			"./foo//bar",
			"/foo",
			"/foo/foo//bar",
			false,
		},
		{
			"git::github.com/hashicorp/foo//bar",
			"",
			"https://github.com/hashicorp/foo.git//bar",
			false,
		},
		{
			"git::https://github.com/hashicorp/consul.git",
			"",
			"https://github.com/hashicorp/consul.git",
			false,
		},
		{
			"git::https://person@someothergit.com/foo/bar", // this
			"",
			"https://person@someothergit.com/foo/bar",
			false,
		},
		{
			"git::https://person@someothergit.com/foo/bar", // this
			"/bar",
			"https://person@someothergit.com/foo/bar",
			false,
		},
		{
			"./foo/archive//*",
			"/bar",
			"/bar/foo/archive//*",
			false,
		},

		// https://github.com/hashicorp/go-getter/pull/124
		{
			"git::ssh://git@my.custom.git/dir1/dir2",
			"",
			"ssh://git@my.custom.git/dir1/dir2",
			false,
		},
		{
			"git::git@my.custom.git:dir1/dir2",
			"/foo",
			"ssh://git@my.custom.git/dir1/dir2",
			false,
		},
		{
			"git::git@my.custom.git:dir1/dir2",
			"",
			"ssh://git@my.custom.git/dir1/dir2",
			false,
		},
	}

	for i, tc := range cases {
		httpGetter := &HttpGetter{
			Netrc: true,
		}
		Getters = []Getter{
			new(GitGetter),
			new(HgGetter),
			new(S3Getter),
			new(GCSGetter),
			new(FileGetter),
			new(SmbGetter),
			httpGetter,
		}
		t.Run(fmt.Sprintf("%d %s", i, tc.Input), func(t *testing.T) {
			detector := NewGetterDetector(Getters)
			output, err := detector.Detect(tc.Input, tc.Pwd)
			if err != nil != tc.Err {
				t.Fatalf("%d: bad err: %s", i, err)
			}
			if output != tc.Output {
				t.Fatalf("%d: bad output: %s\nexpected: %s", i, output, tc.Output)
			}
		})
	}
}
