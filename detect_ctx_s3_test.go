package getter

import (
	"testing"
)

func TestS3CtxDetector(t *testing.T) {
	cases := []struct {
		Input  string
		Output string
	}{
		// Virtual hosted style
		{
			"bucket.s3.amazonaws.com/foo",
			"s3::https://s3.amazonaws.com/bucket/foo",
		},
		{
			"bucket.s3.amazonaws.com/foo/bar",
			"s3::https://s3.amazonaws.com/bucket/foo/bar",
		},
		{
			"bucket.s3.amazonaws.com/foo/bar.baz",
			"s3::https://s3.amazonaws.com/bucket/foo/bar.baz",
		},
		{
			"bucket.s3-eu-west-1.amazonaws.com/foo",
			"s3::https://s3-eu-west-1.amazonaws.com/bucket/foo",
		},
		{
			"bucket.s3-eu-west-1.amazonaws.com/foo/bar",
			"s3::https://s3-eu-west-1.amazonaws.com/bucket/foo/bar",
		},
		{
			"bucket.s3-eu-west-1.amazonaws.com/foo/bar.baz",
			"s3::https://s3-eu-west-1.amazonaws.com/bucket/foo/bar.baz",
		},
		// Path style
		{
			"s3.amazonaws.com/bucket/foo",
			"s3::https://s3.amazonaws.com/bucket/foo",
		},
		{
			"s3.amazonaws.com/bucket/foo/bar",
			"s3::https://s3.amazonaws.com/bucket/foo/bar",
		},
		{
			"s3.amazonaws.com/bucket/foo/bar.baz",
			"s3::https://s3.amazonaws.com/bucket/foo/bar.baz",
		},
		{
			"s3-eu-west-1.amazonaws.com/bucket/foo",
			"s3::https://s3-eu-west-1.amazonaws.com/bucket/foo",
		},
		{
			"s3-eu-west-1.amazonaws.com/bucket/foo/bar",
			"s3::https://s3-eu-west-1.amazonaws.com/bucket/foo/bar",
		},
		{
			"s3-eu-west-1.amazonaws.com/bucket/foo/bar.baz",
			"s3::https://s3-eu-west-1.amazonaws.com/bucket/foo/bar.baz",
		},
		// Misc tests
		{
			"s3-eu-west-1.amazonaws.com/bucket/foo/bar.baz?version=1234",
			"s3::https://s3-eu-west-1.amazonaws.com/bucket/foo/bar.baz?version=1234",
		},
	}

	pwd := "/pwd"
	forceToken := ""
	ctxSubDir := ""
	srcResolveFrom := ""

	f := new(S3CtxDetector)
	for i, tc := range cases {
		output, ok, err := f.CtxDetect(tc.Input, pwd, forceToken, ctxSubDir, srcResolveFrom)
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
