// Copyright IBM Corp. 2015, 2025
// SPDX-License-Identifier: MPL-2.0

package getter

import (
	"path/filepath"
	"testing"
)

func TestTarXzDecompressor(t *testing.T) {

	multiplePaths := []string{"dir/", "dir/test2", "test1"}
	orderingPaths := []string{"workers/", "workers/mq/", "workers/mq/__init__.py"}

	cases := []TestDecompressCase{
		{
			"empty.tar.xz",
			false,
			true,
			nil,
			"",
			nil,
		},

		{
			"single.tar.xz",
			false,
			false,
			nil,
			"b5bb9d8014a0f9b1d61e21e796d78dccdf1352f23cd32812f4850b878ae4944c",
			nil,
		},

		{
			"single.tar.xz",
			true,
			false,
			[]string{"file"},
			"",
			nil,
		},

		{
			"multiple.tar.xz",
			true,
			false,
			[]string{"file1", "file2"},
			"",
			nil,
		},

		{
			"multiple.tar.xz",
			false,
			true,
			nil,
			"",
			nil,
		},

		{
			"multiple_dir.tar.xz",
			true,
			false,
			multiplePaths,
			"",
			nil,
		},

		// Tests when the file is listed before the parent folder
		{
			"ordering.tar.xz",
			true,
			false,
			orderingPaths,
			"",
			nil,
		},
	}

	for i, tc := range cases {
		cases[i].Input = filepath.Join("./testdata", "decompress-txz", tc.Input)
	}

	TestDecompressor(t, new(TarXzDecompressor), cases)
}
