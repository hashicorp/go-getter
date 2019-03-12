package getter

import (
	"path/filepath"
	"testing"
)

func TestZipDecompressor(t *testing.T) {
	cases := []TestDecompressCase{
		{
			"empty.zip",
			false,
			true,
			nil,
			nil,
			"",
			nil,
		},

		{
			"single.zip",
			false,
			false,
			nil,
			nil,
			"d3b07384d113edec49eaa6238ad5ff00",
			nil,
		},

		{
			"single.zip",
			true,
			false,
			[]string{"file"},
			nil,
			"",
			nil,
		},

		{
			"multiple.zip",
			true,
			false,
			[]string{"file1", "file2"},
			nil,
			"",
			nil,
		},

		{
			"multiple.zip",
			false,
			true,
			nil,
			nil,
			"",
			nil,
		},

		{
			"subdir.zip",
			true,
			false,
			[]string{"file1", "subdir/", "subdir/child"},
			nil,
			"",
			nil,
		},

		{
			"subdir_empty.zip",
			true,
			false,
			[]string{"file1", "subdir/"},
			nil,
			"",
			nil,
		},

		{
			"subdir_missing_dir.zip",
			true,
			false,
			[]string{"file1", "subdir/", "subdir/child"},
			nil,
			"",
			nil,
		},

		// Tests that a zip can't contain references with "..".
		{
			"outside_parent.zip",
			true,
			true,
			nil,
			nil,
			"",
			nil,
		},
	}

	for i, tc := range cases {
		cases[i].Input = filepath.Join("./test-fixtures", "decompress-zip", tc.Input)
	}

	TestDecompressor(t, new(ZipDecompressor), cases)
}
