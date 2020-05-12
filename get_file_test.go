package getter

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	testing_helper "github.com/hashicorp/go-getter/v2/helper/testing"
	urlhelper "github.com/hashicorp/go-getter/v2/helper/url"
)

func TestFileGetter_impl(t *testing.T) {
	var _ Getter = new(FileGetter)
}

func TestFileGetter(t *testing.T) {
	g := new(FileGetter)
	dst := testing_helper.TempDir(t)
	ctx := context.Background()

	req := &Request{
		Dst: dst,
		u:   testModuleURL("basic"),
	}

	// With a dir that doesn't exist
	if err := g.Get(ctx, req); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the destination folder is a symlink
	fi, err := os.Lstat(dst)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Fatal("destination is not a symlink")
	}

	// Verify the main file exists
	mainPath := filepath.Join(dst, "main.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestFileGetter_sourceFile(t *testing.T) {
	g := new(FileGetter)
	dst := testing_helper.TempDir(t)
	ctx := context.Background()

	// With a source URL that is a path to a file
	u := testModuleURL("basic")
	u.Path += "/main.tf"

	req := &Request{
		Dst: dst,
		u:   u,
	}
	if err := g.Get(ctx, req); err == nil {
		t.Fatal("should error")
	}
}

func TestFileGetter_sourceNoExist(t *testing.T) {
	g := new(FileGetter)
	dst := testing_helper.TempDir(t)
	ctx := context.Background()

	// With a source URL that doesn't exist
	u := testModuleURL("basic")
	u.Path += "/main"

	req := &Request{
		Dst: dst,
		u:   u,
	}
	if err := g.Get(ctx, req); err == nil {
		t.Fatal("should error")
	}
}

func TestFileGetter_dir(t *testing.T) {
	g := new(FileGetter)
	dst := testing_helper.TempDir(t)
	ctx := context.Background()

	if err := os.MkdirAll(dst, 0755); err != nil {
		t.Fatalf("err: %s", err)
	}

	req := &Request{
		Dst: dst,
		u:   testModuleURL("basic"),
	}
	// With a dir that exists that isn't a symlink
	if err := g.Get(ctx, req); err == nil {
		t.Fatal("should error")
	}
}

func TestFileGetter_dirSymlink(t *testing.T) {
	g := new(FileGetter)
	dst := testing_helper.TempDir(t)
	ctx := context.Background()

	dst2 := testing_helper.TempDir(t)

	// Make parents
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		t.Fatalf("err: %s", err)
	}
	if err := os.MkdirAll(dst2, 0755); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Make a symlink
	if err := os.Symlink(dst2, dst); err != nil {
		t.Fatalf("err: %s", err)
	}

	req := &Request{
		Dst: dst,
		u:   testModuleURL("basic"),
	}

	// With a dir that exists that isn't a symlink
	if err := g.Get(ctx, req); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	mainPath := filepath.Join(dst, "main.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestFileGetter_GetFile(t *testing.T) {
	g := new(FileGetter)
	dst := testing_helper.TempTestFile(t)
	defer os.RemoveAll(filepath.Dir(dst))
	ctx := context.Background()

	req := &Request{
		Dst: dst,
		u:   testModuleURL("basic-file/foo.txt"),
	}

	// With a dir that doesn't exist
	if err := g.GetFile(ctx, req); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the destination folder is a symlink
	fi, err := os.Lstat(dst)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Fatal("destination is not a symlink")
	}

	// Verify the main file exists
	testing_helper.AssertContents(t, dst, "Hello\n")
}

func TestFileGetter_GetFile_Copy(t *testing.T) {
	g := new(FileGetter)

	dst := testing_helper.TempTestFile(t)
	defer os.RemoveAll(filepath.Dir(dst))
	ctx := context.Background()

	req := &Request{
		Dst:  dst,
		u:    testModuleURL("basic-file/foo.txt"),
		Copy: true,
	}

	// With a dir that doesn't exist
	if err := g.GetFile(ctx, req); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the destination folder is a symlink
	fi, err := os.Lstat(dst)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		t.Fatal("destination is a symlink")
	}

	// Verify the main file exists
	testing_helper.AssertContents(t, dst, "Hello\n")
}

// https://github.com/hashicorp/terraform/issues/8418
func TestFileGetter_percent2F(t *testing.T) {
	g := new(FileGetter)
	dst := testing_helper.TempDir(t)
	ctx := context.Background()

	req := &Request{
		Dst: dst,
		u:   testModuleURL("basic%2Ftest"),
	}

	// With a dir that doesn't exist
	if err := g.Get(ctx, req); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	mainPath := filepath.Join(dst, "main.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestFileGetter_Mode_notexist(t *testing.T) {
	g := new(FileGetter)
	ctx := context.Background()

	u := urlhelper.MustParse("nonexistent")
	if _, err := g.Mode(ctx, u); err == nil {
		t.Fatal("expect source file error")
	}
}

func TestFileGetter_Mode_file(t *testing.T) {
	g := new(FileGetter)
	ctx := context.Background()

	// Check the client mode when pointed at a file.
	mode, err := g.Mode(ctx, testModuleURL("basic-file/foo.txt"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if mode != ModeFile {
		t.Fatal("expect ModeFile")
	}
}

func TestFileGetter_Mode_dir(t *testing.T) {
	g := new(FileGetter)
	ctx := context.Background()

	// Check the client mode when pointed at a directory.
	mode, err := g.Mode(ctx, testModuleURL("basic"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if mode != ModeDir {
		t.Fatal("expect ModeDir")
	}
}

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

	f := new(FileGetter)
	for i, tc := range fileTests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			pwd := tc.pwd

			req := &Request{
				Src: tc.in,
				Pwd: pwd,
			}
			ok, err := Detect(req, f)
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

			if req.Src != expected {
				t.Fatalf("input: %q\npwd: %q\nexpected: %q\nbad output: %#v",
					tc.in, pwd, expected, req.Src)
			}
		})
	}
}

var noPwdFileTests = []fileTest{
	{in: "./foo", pwd: "", out: "./foo", err: true},
	{in: "foo", pwd: "", out: "foo", err: true},
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

	f := new(FileGetter)
	for i, tc := range noPwdFileTests {
		req := &Request{
			Src: tc.in,
			Pwd: tc.pwd,
		}
		ok, err := Detect(req, f)
		if err != nil != tc.err {
			t.Fatalf("%d: err: %s", i, err)
		}
		if !ok {
			t.Fatal("not ok")
		}

		if req.Src != tc.out {
			t.Fatalf("%d: bad: %#v", i, req.Src)
		}
	}
}

