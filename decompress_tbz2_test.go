// Copyright IBM Corp. 2015, 2025
// SPDX-License-Identifier: MPL-2.0

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
			"",
			nil,
		},

		{
			"single.tar.bz2",
			false,
			false,
			nil,
			"b5bb9d8014a0f9b1d61e21e796d78dccdf1352f23cd32812f4850b878ae4944c",
			nil,
		},

		{
			"single.tar.bz2",
			true,
			false,
			[]string{"file"},
			"",
			nil,
		},

		{
			"multiple.tar.bz2",
			true,
			false,
			[]string{"file1", "file2"},
			"",
			nil,
		},

		{
			"multiple.tar.bz2",
			false,
			true,
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
			"",
			nil,
		},
	}

	for i, tc := range cases {
		cases[i].Input = filepath.Join("./testdata", "decompress-tbz2", tc.Input)
	}

	TestDecompressor(t, new(TarBzip2Decompressor), cases)
}
