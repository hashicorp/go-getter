package getter

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type ctxFileTest struct {
	in, pwd, out string
	err          bool
}

var ctxFileTests = []ctxFileTest{
	{"./foo", "/pwd", "file:///pwd/foo", false},
	{"./foo?foo=bar", "/pwd", "file:///pwd/foo?foo=bar", false},
	{"foo", "/pwd", "file:///pwd/foo", false},
}

var unixCtxFileTests = []ctxFileTest{
	{"./foo", "testdata/detect-file-symlink-pwd/syml/pwd",
		"testdata/detect-file-symlink-pwd/real/foo", false},

	{"/foo", "/pwd", "file:///foo", false},
	{"/foo?bar=baz", "/pwd", "file:///foo?bar=baz", false},
}

var winCtxFileTests = []ctxFileTest{
	{"/foo", "/pwd", "file:///pwd/foo", false},
	{`C:\`, `/pwd`, `file://C:/`, false},
	{`C:\?bar=baz`, `/pwd`, `file://C:/?bar=baz`, false},
}

func TestFileCtxDetector(t *testing.T) {
	if runtime.GOOS == "windows" {
		ctxFileTests = append(ctxFileTests, winCtxFileTests...)
	} else {
		ctxFileTests = append(ctxFileTests, unixCtxFileTests...)
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

	f := new(FileCtxDetector)

	forceToken := ""
	ctxSubDir := ""
	srcResolveFrom := ""

	for i, tc := range ctxFileTests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			pwd := tc.pwd

			out, ok, err := f.CtxDetect(tc.in, pwd, forceToken, ctxSubDir, srcResolveFrom)
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

var noPwdCtxFileTests = []ctxFileTest{
	{in: "./foo", pwd: "", out: "", err: true},
	{in: "foo", pwd: "", out: "", err: true},
}

var noPwdUnixCtxFileTests = []ctxFileTest{
	{in: "/foo", pwd: "", out: "file:///foo", err: false},
}

var noPwdWinCtxFileTests = []ctxFileTest{
	{in: "/foo", pwd: "", out: "", err: true},
	{in: `C:\`, pwd: ``, out: `file://C:/`, err: false},
}

func TestCtxFileCtxDetector_noPwd(t *testing.T) {
	if runtime.GOOS == "windows" {
		noPwdCtxFileTests = append(noPwdCtxFileTests, noPwdWinCtxFileTests...)
	} else {
		noPwdCtxFileTests = append(noPwdCtxFileTests, noPwdUnixCtxFileTests...)
	}

	f := new(FileCtxDetector)

	forceToken := ""
	ctxSubDir := ""
	srcResolveFrom := ""

	for i, tc := range noPwdCtxFileTests {
		out, ok, err := f.CtxDetect(tc.in, tc.pwd, forceToken, ctxSubDir, srcResolveFrom)
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
