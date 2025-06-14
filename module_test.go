// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package getter

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	urlhelper "github.com/hashicorp/go-getter/helper/url"
)

const fixtureDir = "./testdata"

func tempDir(t *testing.T) string {
	dir, err := ioutil.TempDir("", "tf")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if err := os.RemoveAll(dir); err != nil {
		t.Fatalf("err: %s", err)
	}

	return dir
}

func tempTestFile(t *testing.T) string {
	dir := tempDir(t)
	return filepath.Join(dir, "foo")
}

func testModule(n string) string {
	p := filepath.Join(fixtureDir, n)
	p, err := filepath.Abs(p)
	if err != nil {
		panic(err)
	}
	return fmtFileURL(p)
}
func httpTestModule(n string) *httptest.Server {
	p := filepath.Join(fixtureDir, n)
	p, err := filepath.Abs(p)
	if err != nil {
		panic(err)
	}

	return httptest.NewServer(http.FileServer(http.Dir(p)))
}

func testModuleURL(n string) *url.URL {
	n, subDir := SourceDirSubdir(n)
	u, err := urlhelper.Parse(testModule(n))
	if err != nil {
		panic(err)
	}
	if subDir != "" {
		u.Path += "//" + subDir
		u.RawPath = u.Path
	}

	return u
}

func testURL(s string) *url.URL {
	u, err := urlhelper.Parse(s)
	if err != nil {
		panic(err)
	}

	return u
}

func assertContents(t *testing.T, path string, contents string) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !reflect.DeepEqual(data, []byte(contents)) {
		t.Fatalf("bad. expected:\n\n%s\n\nGot:\n\n%s", contents, string(data))
	}
}
