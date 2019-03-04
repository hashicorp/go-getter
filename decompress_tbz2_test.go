package getter

import (
	"path/filepath"
	"testing"
)

func TestTarBzip2Decompressor(t *testing.T) {
	orderingPaths := []string{"workers/", "workers/mq/", "workers/mq/__init__.py"}

	cases := []TestDecompressCase{
		{
			"empty.tar.bz2",
			false,
			true,
			nil,
			nil,
			"",
			nil,
		},

		{
			"single.tar.bz2",
			false,
			false,
			nil,
			nil,
			"d3b07384d113edec49eaa6238ad5ff00",
			nil,
		},

		{
			"single.tar.bz2",
			true,
			false,
			[]string{"file"},
			nil,
			"",
			nil,
		},

		{
			"multiple.tar.bz2",
			true,
			false,
			[]string{"file1", "file2"},
			nil,
			"",
			nil,
		},

		{
			"multiple.tar.bz2",
			false,
			true,
			nil,
			nil,
			"",
			nil,
		},

		// Tests when the file is listed before the parent folder
		{
			"ordering.tar.bz2",
			true,
			false,
			orderingPaths,
			nil,
			"",
			nil,
		},
	}

	for i, tc := range cases {
		cases[i].Input = filepath.Join("./test-fixtures", "decompress-tbz2", tc.Input)
	}

	TestDecompressor(t, new(TarBzip2Decompressor), cases)
}
