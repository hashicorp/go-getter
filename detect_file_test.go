// Copyright IBM Corp. 2015, 2025
// SPDX-License-Identifier: MPL-2.0

package getter

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type fileTest struct {
	in, pwd, out string
	err          bool
}

var fileTests = []fileTest{
	{"./foo", "/pwd", "file:///pwd/foo", false},
	{"./foo?foo=bar", "/pwd", "file:///pwd/foo?foo=bar", false},
	{"foo", "/pwd", "file:///pwd/foo", false},
}

var unixFileTests = []fileTest{
	{"./foo", "testdata/detect-file-symlink-pwd/syml/pwd",
		"testdata/detect-file-symlink-pwd/real/foo", false},

	{"/foo", "/pwd", "file:///foo", false},
	{"/foo?bar=baz", "/pwd", "file:///foo?bar=baz", false},
}

var winFileTests = []fileTest{
	{"/foo", "/pwd", "file:///pwd/foo", false},
	{`C:\`, `/pwd`, `file://C:/`, false},
	{`C:\?bar=baz`, `/pwd`, `file://C:/?bar=baz`, false},
}

func TestFileDetector(t *testing.T) {
	if runtime.GOOS == "windows" {
		fileTests = append(fileTests, winFileTests...)
	} else {
		fileTests = append(fileTests, unixFileTests...)
	}

	// Get the pwd
	pwdRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	pwdRoot, err = filepath.Abs(pwdRoot)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	f := new(FileDetector)
	for i, tc := range fileTests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			pwd := tc.pwd

			out, ok, err := f.Detect(tc.in, pwd)
			if err != nil {
				t.Fatalf("err: %s", err)
			}
			if !ok {
				t.Fatal("not ok")
			}

			expected := tc.out
			if !strings.HasPrefix(expected, "file://") {
				expected = "file://" + filepath.Join(pwdRoot, expected)
			}

			if out != expected {
				t.Fatalf("input: %q\npwd: %q\nexpected: %q\nbad output: %#v",
					tc.in, pwd, expected, out)
			}
		})
	}
}

var noPwdFileTests = []fileTest{
	{in: "./foo", pwd: "", out: "", err: true},
	{in: "foo", pwd: "", out: "", err: true},
}

var noPwdUnixFileTests = []fileTest{
	{in: "/foo", pwd: "", out: "file:///foo", err: false},
}

var noPwdWinFileTests = []fileTest{
	{in: "/foo", pwd: "", out: "", err: true},
	{in: `C:\`, pwd: ``, out: `file://C:/`, err: false},
}

func TestFileDetector_percentInPath(t *testing.T) {
	dir := t.TempDir()
	name := "{% if foo %}bar{% endif %}"
	src := filepath.Join(dir, name)
	if err := os.Mkdir(src, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "test.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	got, ok, err := new(FileDetector).Detect(name, dir)
	if err != nil {
		t.Fatalf("detect: %s", err)
	}
	if !ok {
		t.Fatal("not ok")
	}

	u, err := url.Parse(got)
	if err != nil {
		t.Fatalf("parse %q: %s", got, err)
	}
	// FileGetter stats RawPath when it is set. A non-empty RawPath here
	// would be the encoded name, not the directory on disk.
	if u.RawPath != "" {
		t.Fatalf("RawPath = %q", u.RawPath)
	}
	if u.Path != src {
		t.Fatalf("path\n got: %q\nwant: %q", u.Path, src)
	}

	dst := filepath.Join(t.TempDir(), "out")
	client := &Client{
		Src:  name,
		Dst:  dst,
		Pwd:  dir,
		Mode: ClientModeDir,
	}
	if err := client.Get(); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(dst, "test.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "hello" {
		t.Fatalf("contents: %q", body)
	}

	// A literal %20 is a filename, not a space.
	literal := filepath.Join(dir, "pct%20name")
	if err := os.Mkdir(literal, 0755); err != nil {
		t.Fatal(err)
	}
	got, ok, err = new(FileDetector).Detect("pct%20name", dir)
	if err != nil || !ok {
		t.Fatalf("detect literal: ok=%v err=%v", ok, err)
	}
	u, err = url.Parse(got)
	if err != nil {
		t.Fatalf("parse literal %q: %s", got, err)
	}
	if u.RawPath != "" || u.Path != literal {
		t.Fatalf("literal path\n got: %#v\nwant: %q", u, literal)
	}

	// A query on the source string stays a query when the path contains %.
	got, ok, err = new(FileDetector).Detect("pct%20name?foo=bar", dir)
	if err != nil || !ok {
		t.Fatalf("detect query: ok=%v err=%v", ok, err)
	}
	u, err = url.Parse(got)
	if err != nil {
		t.Fatalf("parse query %q: %s", got, err)
	}
	if u.Path != literal || u.RawQuery != "foo=bar" || u.RawPath != "" {
		t.Fatalf("query url: %#v", u)
	}
}

func TestFileDetector_noPwd(t *testing.T) {
	if runtime.GOOS == "windows" {
		noPwdFileTests = append(noPwdFileTests, noPwdWinFileTests...)
	} else {
		noPwdFileTests = append(noPwdFileTests, noPwdUnixFileTests...)
	}

	f := new(FileDetector)
	for i, tc := range noPwdFileTests {
		out, ok, err := f.Detect(tc.in, tc.pwd)
		if err != nil != tc.err {
			t.Fatalf("%d: err: %s", i, err)
		}
		if !ok {
			t.Fatal("not ok")
		}

		if out != tc.out {
			t.Fatalf("%d: bad: %#v", i, out)
		}
	}
}
