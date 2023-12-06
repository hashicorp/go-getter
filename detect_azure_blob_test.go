package getter

import (
	"testing"
)

func TestAzureBlobDetector(t *testing.T) {
	cases := []struct {
		Input  string
		Output string
	}{
		{
			"account.blob.core.windows.net/foo/bar",
			"azureblob::https://account.blob.core.windows.net/foo/bar",
		},
		{
			"account.blob.core.usgovcloudapi.net/foo/bar",
			"azureblob::https://account.blob.core.usgovcloudapi.net/foo/bar",
		},
		{
			"account.blob.core.chinacloudapi.cn/foo/bar",
			"azureblob::https://account.blob.core.chinacloudapi.cn/foo/bar",
		},
		// Misc tests
		{
			"account.blob.core.windows.net/foo/bar?version=1234",
			"azureblob::https://account.blob.core.windows.net/foo/bar?version=1234",
		},
	}

	pwd := "/pwd"
	f := new(AzureBlobDetector)
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
