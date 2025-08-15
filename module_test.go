package getter

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	urlhelper "github.com/hashicorp/go-getter/v2/helper/url"
)

const fixtureDir = "./testdata"

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
