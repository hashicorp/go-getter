package getter

import (
	"path/filepath"
	"testing"
)

func TestGzipDecompressor(t *testing.T) {
	cases := []decompressTestCase{
		{
			"single.gz",
			false,
			false,
			nil,
			"d3b07384d113edec49eaa6238ad5ff00",
		},

		{
			"single.gz",
			true,
			true,
			nil,
			"",
		},
	}

	for i, tc := range cases {
		cases[i].Input = filepath.Join("./test-fixtures", "decompress-gz", tc.Input)
	}

	decompressorTest(t, new(GzipDecompressor), cases)
}
