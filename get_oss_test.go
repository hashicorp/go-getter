// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package getter

import (
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

// Note for external contributors: In order to run the OSS test suite, you will only be able to be run
// in GitHub Actions when you open a PR.

func TestOSSGetter_impl(t *testing.T) {
	var _ Getter = new(OSSGetter)
}

func TestOSSGetter(t *testing.T) {
	g := new(OSSGetter)
	dst := tempDir(t)

	// With a dir exists
	err := g.Get(
		dst, testURL("https://hc-go-getter-test.oss-ap-southeast-1.aliyuncs.com/go-getter/folder"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	mainPath := filepath.Join(dst, "main.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestOSSGetter_subdir(t *testing.T) {
	g := new(OSSGetter)
	dst := tempDir(t)

	// With a dir exists
	err := g.Get(
		dst, testURL("https://hc-go-getter-test.oss-ap-southeast-1.aliyuncs.com/go-getter/folder/subfolder"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	subPath := filepath.Join(dst, "sub.tf")
	if _, err := os.Stat(subPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestOSSGetter_GetFile(t *testing.T) {
	g := new(OSSGetter)
	dst := tempTestFile(t)
	defer os.RemoveAll(filepath.Dir(dst))

	// Download
	err := g.GetFile(
		dst, testURL("https://hc-go-getter-test.oss-ap-southeast-1.aliyuncs.com/go-getter/folder/main.tf"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	if _, err := os.Stat(dst); err != nil {
		t.Fatalf("err: %s", err)
	}
	assertContents(t, dst, "# Main\n")
}

func TestOSSGetter_GetFile_notfound(t *testing.T) {
	g := new(OSSGetter)
	dst := tempTestFile(t)
	defer os.RemoveAll(filepath.Dir(dst))

	// Download
	err := g.GetFile(
		dst, testURL("https://hc-go-getter-test.oss-ap-southeast-1.aliyuncs.com/go-getter/folder/404.tf"))
	if err == nil {
		t.Fatalf("expected error, got none")
	}
}

func TestOSSGetter_ClientMode_dir(t *testing.T) {
	g := new(OSSGetter)

	// Check client mode on a key prefix with only a single key.
	mode, err := g.ClientMode(
		testURL("https://hc-go-getter-test.oss-ap-southeast-1.aliyuncs.com/go-getter/folder"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if mode != ClientModeDir {
		t.Fatal("expect ClientModeDir")
	}
}

func TestOSSGetter_ClientMode_file(t *testing.T) {
	g := new(OSSGetter)

	// Check client mode on a key prefix which contains sub-keys.
	mode, err := g.ClientMode(
		testURL("https://hc-go-getter-test.oss-ap-southeast-1.aliyuncs.com/go-getter/folder/main.tf"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if mode != ClientModeFile {
		t.Fatal("expect ClientModeFile")
	}
}

func TestOSSGetter_ClientMode_notfound(t *testing.T) {
	g := new(OSSGetter)

	// Check the client mode when a non-existent key is looked up. This does not
	// return an error, but rather should just return the file mode so that OSS
	// can return an appropriate error later on. This also checks that the
	// prefix is handled properly (e.g., "/fold" and "/folder" don't put the
	// client mode into "dir".
	mode, err := g.ClientMode(
		testURL("https://hc-go-getter-test.oss-ap-southeast-1.aliyuncs.com/go-getter/fold"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if mode != ClientModeFile {
		t.Fatal("expect ClientModeFile")
	}
}

func TestOSSGetter_ClientMode_collision(t *testing.T) {
	g := new(OSSGetter)

	// Check that the client mode is "file" if there is both an object and a
	// folder with a common prefix (i.e., a "collision" in the namespace).
	mode, err := g.ClientMode(
		testURL("https://hc-go-getter-test.oss-ap-southeast-1.aliyuncs.com/go-getter/collision/foo"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if mode != ClientModeFile {
		t.Fatal("expect ClientModeFile")
	}
}

func TestOSSGetter_Url(t *testing.T) {
	var OSStests = []struct {
		name        string
		url         string
		region      string
		bucket      string
		path        string
		version     string
		expectedErr string
	}{
		{
			name:    "OSSVhostDash",
			url:     "oss::https://bucket.oss-ap-southeast-1.aliyuncs.com/foo/bar.baz",
			region:  "ap-southeast-1",
			bucket:  "bucket",
			path:    "foo/bar.baz",
			version: "",
		},
		{
			name:    "OSSVhostDash",
			url:     "oss::https://bucket.oss-cn-hangzhou-internal.aliyuncs.com/foo/bar.baz",
			region:  "cn-hangzhou",
			bucket:  "bucket",
			path:    "foo/bar.baz",
			version: "",
		},
		{
			name:    "OSSVhostIPv6",
			url:     "oss::https://bucket.cn-hangzhou.oss.aliyuncs.com/foo/bar.baz",
			region:  "cn-hangzhou",
			bucket:  "bucket",
			path:    "foo/bar.baz",
			version: "",
		},
		{
			name:    "OSSv1234",
			url:     "oss::https://bucket.oss-ap-southeast-1.aliyuncs.com/foo/bar.baz?version=1234",
			region:  "ap-southeast-1",
			bucket:  "bucket",
			path:    "foo/bar.baz",
			version: "1234",
		},
		{
			name:        "malformed OSS url",
			url:         "oss::https://bucket-ap-southeast-1.aliyuncs.com/foo/bar.baz?version=1234",
			expectedErr: "URL is not a valid OSS URL",
		},
	}

	for i, pt := range OSStests {
		t.Run(pt.name, func(t *testing.T) {
			g := new(OSSGetter)
			forced, src := getForcedGetter(pt.url)
			u, err := url.Parse(src)

			if err != nil {
				t.Errorf("test %d: unexpected error: %s", i, err)
			}
			if forced != "OSS" {
				t.Fatalf("expected forced protocol to be OSS")
			}

			region, bucket, path, version, err := g.parseUrl(u)

			if err != nil {
				if pt.expectedErr == "" {
					t.Fatalf("err: %s", err)
				}
				if err.Error() != pt.expectedErr {
					t.Fatalf("expected %s, got %s", pt.expectedErr, err.Error())
				}
				return
			} else if pt.expectedErr != "" {
				t.Fatalf("expected error, got none")
			}
			if region != pt.region {
				t.Fatalf("expected %s, got %s", pt.region, region)
			}
			if bucket != pt.bucket {
				t.Fatalf("expected %s, got %s", pt.bucket, bucket)
			}
			if path != pt.path {
				t.Fatalf("expected %s, got %s", pt.path, path)
			}
			if version != pt.version {
				t.Fatalf("expected %s, got %s", pt.version, version)
			}
		})
	}
}
