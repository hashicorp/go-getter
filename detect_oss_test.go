package getter

import (
	"testing"
)

func TestOSSDetector(t *testing.T) {
	cases := []struct {
		Input  string
		Output string
	}{
		{
			"go-getter.oss-ap-southeast-1.aliyuncs.com/foo",
			"oss::https://go-getter.oss-ap-southeast-1.aliyuncs.com/foo",
		},
		{
			"go-getter.oss-ap-southeast-1.aliyuncs.com/foo/bar",
			"oss::https://go-getter.oss-ap-southeast-1.aliyuncs.com/foo/bar",
		},
		{
			"go-getter.oss-ap-southeast-1.aliyuncs.com/foo/bar.baz",
			"oss::https://go-getter.oss-ap-southeast-1.aliyuncs.com/foo/bar.baz",
		},
	}

	pwd := "/pwd"
	f := new(OSSDetector)
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
