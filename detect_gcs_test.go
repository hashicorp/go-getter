package getter

import (
	"testing"
)

func TestGCSDetector(t *testing.T) {
	cases := []struct {
		Input  string
		Output string
	}{
		{
			"https://www.googleapis.com/storage/v1/test/test.tar.gz",
			"gcs::https://www.googleapis.com/storage/v1/test/test.tar.gz",
		},
		{
			"www.googleapis.com/storage/v1/test/test.tar.gz",
			"gcs::https://www.googleapis.com/storage/v1/test/test.tar.gz",
		},
		{
			"googleapis.com/storage/v1/test/test.tar.gz",
			"gcs::https://www.googleapis.com/storage/v1/test/test.tar.gz",
		},
	}

	pwd := "/pwd"
	f := new(GCSDetector)
	for i, tc := range cases {
		output, ok, err := f.Detect(tc.Input, pwd)
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if !ok {
			t.Fatal("not ok")
		}
		if output != tc.Output {
			t.Fatalf("%d: bad: %#v", i, output)
		}
	}
}
