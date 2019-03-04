package getter

import (
	"path/filepath"
	"testing"
	"time"
)

func TestTar(t *testing.T) {
	mtime := time.Unix(0, 0)
	cases := []TestDecompressCase{
		{
			"extended_header.tar",
			true,
			false,
			[]string{"directory/", "directory/a", "directory/b"},
			nil,
			"",
			nil,
		},
		{
			"implied_dir.tar",
			true,
			false,
			[]string{"directory/", "directory/sub/", "directory/sub/a", "directory/sub/b"},
			nil,
			"",
			nil,
		},
		{
			"unix_time_0.tar",
			true,
			false,
			[]string{"directory/", "directory/sub/", "directory/sub/a", "directory/sub/b"},
			nil,
			"",
			&mtime,
		},
		{
			"with-symlinks.tar",
			true,
			false,
			[]string{"baz", "foo"},
			map[string]string{"bar": "baz"},
			"",
			&mtime,
		},
	}

	for i, tc := range cases {
		cases[i].Input = filepath.Join("./test-fixtures", "decompress-tar", tc.Input)
	}

	TestDecompressor(t, new(tarDecompressor), cases)
}
