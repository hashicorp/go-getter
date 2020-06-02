package getter

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

type fileTest struct {
	in, pwd, out string
	symlink, err bool
}

var fileTests = []fileTest{
	{"./foo", "/pwd", "/pwd/foo", false, false},
	{"./foo?foo=bar", "/pwd", "/pwd/foo?foo=bar", false, false},
	{"foo", "/pwd", "/pwd/foo", false, false},
}

var unixFileTests = []fileTest{
	{"./foo", "testdata/detect-file-symlink-pwd/syml/pwd",
		"testdata/detect-file-symlink-pwd/real/foo", true, false},

	{"/foo", "/pwd", "/foo", false, false},
	{"/foo?bar=baz", "/pwd", "/foo?bar=baz", false, false},
}

var winFileTests = []fileTest{
	{"/foo", "/pwd", "/pwd/foo", false, false},
	{`C:\`, `/pwd`, `C:/`, false, false},
	{`C:\?bar=baz`, `/pwd`, `C:/?bar=baz`, false, false},
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
			if tc.symlink {
				expected = filepath.Join(pwdRoot, expected)
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
	{in: "/foo", pwd: "", out: "/foo", err: false},
}

var noPwdWinFileTests = []fileTest{
	{in: "/foo", pwd: "", out: "", err: true},
	{in: `C:\`, pwd: ``, out: `C:/`, err: false},
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
