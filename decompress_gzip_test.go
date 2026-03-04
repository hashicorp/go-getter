// Copyright IBM Corp. 2015, 2025
// SPDX-License-Identifier: MPL-2.0

package getter

import (
	"path/filepath"
	"testing"
)

func TestGzipDecompressor(t *testing.T) {
	cases := []TestDecompressCase{
		{
			"single.gz",
			false,
			false,
			nil,
			"b5bb9d8014a0f9b1d61e21e796d78dccdf1352f23cd32812f4850b878ae4944c",
			nil,
		},

		{
			"single.gz",
			true,
			true,
			nil,
			"",
			nil,
		},
	}

	for i, tc := range cases {
		cases[i].Input = filepath.Join("./testdata", "decompress-gz", tc.Input)
	}

	TestDecompressor(t, new(GzipDecompressor), cases)
}
