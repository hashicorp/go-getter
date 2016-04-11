package getter

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestTarDecompressor(t *testing.T) {

	multiplePaths := []string{"dir/", "dir/test2", "test1"}
	if runtime.GOOS == "windows" {
		multiplePaths = []string{"dir/", "dir\\test2", "test1"}
	}

	cases := []TestDecompressCase{
		{
			"empty.tar",
			false,
			true,
			nil,
			"",
		},

		{
			"single.tar",
			false,
			false,
			nil,
			"d3b07384d113edec49eaa6238ad5ff00",
		},

		{
			"single.tar",
			true,
			false,
			[]string{"file"},
			"",
		},

		{
			"multiple.tar",
			true,
			false,
			[]string{"file1", "file2"},
			"",
		},

		{
			"multiple.tar",
			false,
			true,
			nil,
			"",
		},

		{
			"multiple_dir.tar",
			true,
			false,
			multiplePaths,
			"",
		},
	}

	for i, tc := range cases {
		cases[i].Input = filepath.Join("./test-fixtures", "decompress-tar", tc.Input)
	}

	TestDecompressor(t, new(TarDecompressor), cases)
}
