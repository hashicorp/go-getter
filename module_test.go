package getter

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"

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
