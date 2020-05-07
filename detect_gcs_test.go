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
			"www.googleapis.com/storage/v1/bucket/foo",
			"https://www.googleapis.com/storage/v1/bucket/foo",
		},
		{
			"www.googleapis.com/storage/v1/bucket/foo/bar",
			"https://www.googleapis.com/storage/v1/bucket/foo/bar",
		},
		{
			"www.googleapis.com/storage/v1/foo/bar.baz",
			"https://www.googleapis.com/storage/v1/foo/bar.baz",
		},
	}

	pwd := "/pwd"
	f := new(GCSGetter)
	for i, tc := range cases {
		req := &Request{
			Src: tc.Input,
			Pwd: pwd,
		}
		ok, err := Detect(req, f)
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if !ok {
			t.Fatal("not ok")
		}

		if req.Src != tc.Output {
			t.Fatalf("%d: bad: %#v", i, req.Src)
		}
	}
}
