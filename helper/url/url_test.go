package url

import (
	"testing"
)

type parseTest struct {
	rawURL string
	scheme string
	host   string
	path   string
	str    string
	err    bool
}

var parseTests = []parseTest{
	{
		rawURL: "/foo/bar",
		scheme: "",
		host:   "",
		path:   "/foo/bar",
		str:    "/foo/bar",
		err:    false,
	},
	{
		rawURL: "file:///dir/",
		scheme: "file",
		host:   "",
		path:   "/dir/",
		str:    "file:///dir/",
		err:    false,
	},
}

var winParseTests = []parseTest{
	{
		rawURL: `C:\`,
		scheme: `file`,
		host:   ``,
		path:   `C:/`,
		str:    `file://C:/`,
		err:    false,
	},
	{
		rawURL: `file://C:\`,
		scheme: `file`,
		host:   ``,
		path:   `C:/`,
		str:    `file://C:/`,
		err:    false,
	},
	{
		rawURL: `file:///C:\`,
		scheme: `file`,
		host:   ``,
		path:   `C:/`,
		str:    `file://C:/`,
		err:    false,
	},
}

func TestParse(t *testing.T) {
	for i, pt := range append(parseTests, winParseTests...) {
		url, err := Parse(pt.rawURL)
		if err != nil && !pt.err {
			t.Errorf("test %d: unexpected error: %s", i, err)
		}
		if err == nil && pt.err {
			t.Errorf("test %d: expected an error", i)
		}
		if url == nil {
			continue
		}
		if url.Scheme != pt.scheme {
			t.Errorf("test %d: expected Scheme = %q, got %q", i, pt.scheme, url.Scheme)
		}
		if url.Host != pt.host {
			t.Errorf("test %d: expected Host = %q, got %q", i, pt.host, url.Host)
		}
		if url.Path != pt.path {
			t.Errorf("test %d: expected Path = %q, got %q", i, pt.path, url.Path)
		}
		if url.String() != pt.str {
			t.Errorf("test %d: expected url.String() = %q, got %q", i, pt.str, url.String())
		}
	}
}
