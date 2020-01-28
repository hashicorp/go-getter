package getter

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGet_badSchema(t *testing.T) {
	ctx := context.Background()

	dst := tempDir(t)
	u := testModule("basic")
	u = strings.Replace(u, "file", "nope", -1)

	op, err := Get(ctx, dst, u)
	if err == nil {
		t.Fatal("should error")
	}
	if op != nil {
		t.Fatal("op should be nil")
	}
}

func TestGet_dir(t *testing.T) {
	ctx := context.Background()

	dst := tempDir(t)
	u := testModule("basic")

	op, err := Get(ctx, dst, u)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if diff := cmp.Diff(&Operation{Dst: dst}, op); diff != "" {
		t.Fatalf("unexpected op: %s", diff)
	}

	mainPath := filepath.Join(dst, "main.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

// https://github.com/hashicorp/terraform/issues/11438
func TestGet_fileDecompressorExt(t *testing.T) {
	ctx := context.Background()

	dst := tempDir(t)
	u := testModule("basic-tgz")

	op, err := Get(ctx, dst, u)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if diff := cmp.Diff(&Operation{Dst: dst}, op); diff != "" {
		t.Fatalf("unexpected op: %s", diff)
	}

	mainPath := filepath.Join(dst, "main.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

// https://github.com/hashicorp/terraform/issues/8418
func TestGet_filePercent2F(t *testing.T) {
	ctx := context.Background()

	dst := tempDir(t)
	u := testModule("basic%2Ftest")

	op, err := Get(ctx, dst, u)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if diff := cmp.Diff(&Operation{Dst: dst}, op); diff != "" {
		t.Fatalf("unexpected op: %s", diff)
	}

	mainPath := filepath.Join(dst, "main.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestGet_relativeDir(t *testing.T) {
	ctx := context.Background()

	dst := tempDir(t)
	u := filepath.Join(".", "test-fixtures", "basic")
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	req := &Request{
		Src:  u,
		Dst:  dst,
		Pwd:  pwd,
		Mode: ClientModeDir,
	}
	client := &Client{}

	if err := client.Configure(); err != nil {
		t.Fatalf("configure: %s", err)
	}

	op, err := client.Get(ctx, req)
	if err != nil {
		t.Fatalf("get: %s", err)
	}
	if diff := cmp.Diff(&Operation{Dst: dst}, op); diff != "" {
		t.Fatalf("unexpected op: %s", diff)
	}

	mainPath := filepath.Join(dst, "main.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("stat: %s", err)
	}
}

func TestGet_fileForced(t *testing.T) {
	ctx := context.Background()

	dst := tempDir(t)
	u := testModule("basic")
	u = "file::" + u

	op, err := Get(ctx, dst, u)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if diff := cmp.Diff(&Operation{Dst: dst}, op); diff != "" {
		t.Fatalf("unexpected op: %s", diff)
	}

	mainPath := filepath.Join(dst, "main.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestGet_fileSubdir(t *testing.T) {
	ctx := context.Background()

	dst := tempDir(t)
	u := testModule("basic//subdir")

	op, err := Get(ctx, dst, u)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if diff := cmp.Diff(&Operation{Dst: dst}, op); diff != "" {
		t.Fatalf("unexpected op: %s", diff)
	}

	mainPath := filepath.Join(dst, "sub.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestGet_archive(t *testing.T) {
	ctx := context.Background()

	dst := tempDir(t)
	u := filepath.Join("./test-fixtures", "archive.tar.gz")
	u, _ = filepath.Abs(u)

	op, err := Get(ctx, dst, u)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if diff := cmp.Diff(&Operation{Dst: dst}, op); diff != "" {
		t.Fatalf("unexpected op: %s", diff)
	}

	mainPath := filepath.Join(dst, "main.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestGetAny_archive(t *testing.T) {
	ctx := context.Background()

	dst := tempDir(t)
	u := filepath.Join("./test-fixtures", "archive.tar.gz")
	u, _ = filepath.Abs(u)

	op, err := GetAny(ctx, dst, u)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if diff := cmp.Diff(&Operation{Dst: dst}, op); diff != "" {
		t.Fatalf("unexpected op: %s", diff)
	}

	mainPath := filepath.Join(dst, "main.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestGet_archiveRooted(t *testing.T) {
	ctx := context.Background()

	dst := tempDir(t)
	u := testModule("archive-rooted/archive.tar.gz")
	op, err := Get(ctx, dst, u)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if diff := cmp.Diff(&Operation{Dst: dst}, op); diff != "" {
		t.Fatalf("unexpected op: %s", diff)
	}

	mainPath := filepath.Join(dst, "root", "hello.txt")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestGet_archiveSubdirWild(t *testing.T) {
	ctx := context.Background()

	dst := tempDir(t)
	u := testModule("archive-rooted/archive.tar.gz")
	u += "//*"
	op, err := Get(ctx, dst, u)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if diff := cmp.Diff(&Operation{Dst: dst}, op); diff != "" {
		t.Fatalf("unexpected op: %s", diff)
	}

	mainPath := filepath.Join(dst, "hello.txt")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestGet_archiveSubdirWildMultiMatch(t *testing.T) {
	ctx := context.Background()

	dst := tempDir(t)
	u := testModule("archive-rooted-multi/archive.tar.gz")
	u += "//*"
	op, err := Get(ctx, dst, u)
	switch err {
	case nil:
		t.Fatal("should error")
	default:
		if !strings.Contains(err.Error(), "multiple") {
			t.Fatalf("err: %s", err)
		}
		if op != nil {
			t.Fatal("operation should be nil")
		}
	}
}

func TestGetAny_file(t *testing.T) {
	ctx := context.Background()

	dst := tempDir(t)
	u := testModule("basic-file/foo.txt")
	mainPath := filepath.Join(dst, "foo.txt")

	op, err := GetAny(ctx, dst, u)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if diff := cmp.Diff(&Operation{Dst: mainPath}, op); diff != "" {
		t.Fatalf("unexpected op: %s", diff)
	}

	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestGetAny_dir(t *testing.T) {
	ctx := context.Background()

	dst := tempDir(t)
	u := filepath.Join("./test-fixtures", "basic")
	u, _ = filepath.Abs(u)

	op, err := GetAny(ctx, dst, u)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if diff := cmp.Diff(&Operation{Dst: dst}, op); diff != "" {
		t.Fatalf("unexpected op: %s", diff)
	}

	check := []string{
		"main.tf",
		"foo/main.tf",
	}

	for _, name := range check {
		mainPath := filepath.Join(dst, name)
		if _, err := os.Stat(mainPath); err != nil {
			t.Fatalf("err: %s", err)
		}
	}
}

func TestGetFile(t *testing.T) {
	ctx := context.Background()

	dst := tempTestFile(t)
	defer os.RemoveAll(filepath.Dir(dst))
	u := testModule("basic-file/foo.txt")

	op, err := GetFile(ctx, dst, u)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if diff := cmp.Diff(&Operation{Dst: dst}, op); diff != "" {
		t.Fatalf("unexpected op: %s", diff)
	}

	// Verify the main file exists
	assertContents(t, dst, "Hello\n")
}

func TestGetFile_archive(t *testing.T) {
	ctx := context.Background()

	dst := tempTestFile(t)
	defer os.RemoveAll(filepath.Dir(dst))
	u := testModule("basic-file-archive/archive.tar.gz")

	op, err := GetFile(ctx, dst, u)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if diff := cmp.Diff(&Operation{Dst: dst}, op); diff != "" {
		t.Fatalf("unexpected op: %s", diff)
	}

	// Verify the main file exists
	assertContents(t, dst, "Hello\n")
}

func TestGetFile_archiveChecksum(t *testing.T) {
	ctx := context.Background()

	dst := tempTestFile(t)
	defer os.RemoveAll(filepath.Dir(dst))
	u := testModule(
		"basic-file-archive/archive.tar.gz?checksum=md5:fbd90037dacc4b1ab40811d610dde2f0")

	op, err := GetFile(ctx, dst, u)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if diff := cmp.Diff(&Operation{Dst: dst}, op); diff != "" {
		t.Fatalf("unexpected op: %s", diff)
	}

	// Verify the main file exists
	assertContents(t, dst, "Hello\n")
}

func TestGetFile_archiveNoUnarchive(t *testing.T) {
	ctx := context.Background()

	dst := tempTestFile(t)
	defer os.RemoveAll(filepath.Dir(dst))
	u := testModule("basic-file-archive/archive.tar.gz")
	u += "?archive=false"

	op, err := GetFile(ctx, dst, u)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if diff := cmp.Diff(&Operation{Dst: dst}, op); diff != "" {
		t.Fatalf("unexpected op: %s", diff)
	}

	// Verify the main file exists
	actual := testMD5(t, dst)
	expected := "fbd90037dacc4b1ab40811d610dde2f0"
	if actual != expected {
		t.Fatalf("bad: %s", actual)
	}
}

func TestGetFile_checksum(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		Append string
		Err    bool
	}{
		{
			"",
			false,
		},

		// MD5
		{
			"?checksum=09f7e02f1290be211da707a266f153b3",
			false,
		},
		{
			"?checksum=md5:09f7e02f1290be211da707a266f153b3",
			false,
		},
		{
			"?checksum=md5:09f7e02f1290be211da707a266f153b4",
			true,
		},

		// SHA1
		{
			"?checksum=1d229271928d3f9e2bb0375bd6ce5db6c6d348d9",
			false,
		},
		{
			"?checksum=sha1:1d229271928d3f9e2bb0375bd6ce5db6c6d348d9",
			false,
		},
		{
			"?checksum=sha1:1d229271928d3f9e2bb0375bd6ce5db6c6d348d0",
			true,
		},

		// SHA256
		{
			"?checksum=66a045b452102c59d840ec097d59d9467e13a3f34f6494e539ffd32c1bb35f18",
			false,
		},
		{
			"?checksum=sha256:66a045b452102c59d840ec097d59d9467e13a3f34f6494e539ffd32c1bb35f18",
			false,
		},
		{
			"?checksum=sha256:66a045b452102c59d840ec097d59d9467e13a3f34f6494e539ffd32c1bb35f19",
			true,
		},

		// SHA512
		{
			"?checksum=c2bad2223811194582af4d1508ac02cd69eeeeedeeb98d54fcae4dcefb13cc882e7640328206603d3fb9cd5f949a9be0db054dd34fbfa190c498a5fe09750cef",
			false,
		},
		{
			"?checksum=sha512:c2bad2223811194582af4d1508ac02cd69eeeeedeeb98d54fcae4dcefb13cc882e7640328206603d3fb9cd5f949a9be0db054dd34fbfa190c498a5fe09750cef",
			false,
		},
		{
			"?checksum=sha512:c2bad2223811194582af4d1508ac02cd69eeeeedeeb98d54fcae4dcefb13cc882e7640328206603d3fb9cd5f949a9be0db054dd34fbfa190c498a5fe09750ced",
			true,
		},
	}

	for _, tc := range cases {
		u := testModule("basic-file/foo.txt") + tc.Append

		func() {
			dst := tempTestFile(t)
			defer os.RemoveAll(filepath.Dir(dst))
			op, err := GetFile(ctx, dst, u)
			if (err != nil) != tc.Err {
				t.Fatalf("append: %s\n\nerr: %s", tc.Append, err)
			}
			if err == nil {
				if diff := cmp.Diff(&Operation{Dst: dst}, op); diff != "" {
					t.Fatalf("unexpected dst: %s", diff)
				}
			}

			// Verify the main file exists
			assertContents(t, dst, "Hello\n")
		}()
	}
}

func TestGetFile_checksum_from_file(t *testing.T) {

	checksums := testModule("checksum-file")
	httpChecksums := httpTestModule("checksum-file")
	defer httpChecksums.Close()

	cases := []struct {
		Append       string
		WantTransfer bool
		WantErr      bool
	}{
		{
			"",
			true,
			false,
		},

		// md5
		{
			"?checksum=file:" + checksums + "/md5-p.sum",
			true,
			false,
		},
		{
			"?checksum=file:" + httpChecksums.URL + "/md5-bsd.sum",
			true,
			false,
		},
		{
			"?checksum=file:" + checksums + "/md5-bsd-bad.sum",
			false,
			true,
		},
		{
			"?checksum=file:" + httpChecksums.URL + "/md5-bsd-wrong.sum",
			true,
			true,
		},

		// sha1
		{
			"?checksum=file:" + checksums + "/sha1-p.sum",
			true,
			false,
		},
		{
			"?checksum=file:" + httpChecksums.URL + "/sha1.sum",
			true,
			false,
		},

		// sha256
		{
			"?checksum=file:" + checksums + "/sha256-p.sum",
			true,
			false,
		},

		// sha512
		{
			"?checksum=file:" + httpChecksums.URL + "/sha512-p.sum",
			true,
			false,
		},
	}

	for _, tc := range cases {
		u := checksums + "/content.txt" + tc.Append
		t.Run(tc.Append, func(t *testing.T) {
			ctx := context.Background()

			dst := tempTestFile(t)
			defer os.RemoveAll(filepath.Dir(dst))
			op, err := GetFile(ctx, dst, u)
			if (err != nil) != tc.WantErr {
				t.Fatalf("append: %s\n\nerr: %s", tc.Append, err)
			}
			if err == nil {
				if diff := cmp.Diff(&Operation{Dst: dst}, op); diff != "" {
					t.Fatalf("unexpected dst: %s", diff)
				}
			}

			if tc.WantTransfer {
				// Verify the main file exists
				assertContents(t, dst, "I am a file with some content\n")
			}
		})
	}
}

func TestGetFile_checksumURL(t *testing.T) {
	ctx := context.Background()

	dst := tempTestFile(t)
	defer os.RemoveAll(filepath.Dir(dst))
	u := testModule("basic-file/foo.txt") + "?checksum=md5:09f7e02f1290be211da707a266f153b3"

	getter := &MockGetter{Proxy: new(FileGetter)}
	req := &Request{
		Src:  u,
		Dst:  dst,
		Mode: ClientModeFile,
	}
	client := &Client{
		Getters: map[string]Getter{
			"file": getter,
		},
	}

	op, err := client.Get(ctx, req)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if diff := cmp.Diff(&Operation{Dst: dst}, op); diff != "" {
		t.Fatalf("unexpected op: %s", diff)
	}

	if v := getter.GetFileURL.Query().Get("checksum"); v != "" {
		t.Fatalf("bad: %s", v)
	}
}

func TestGetFile_filename(t *testing.T) {
	ctx := context.Background()

	dst := tempDir(t)
	u := testModule("basic-file/foo.txt")

	u += "?filename=bar.txt"
	realDst := filepath.Join(dst, "bar.txt")

	op, err := GetAny(ctx, dst, u)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if diff := cmp.Diff(realDst, op.Dst); diff != "" {
		t.Fatalf("unexpected op: %s", diff)
	}

	mainPath := realDst
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestGetFile_checksumSkip(t *testing.T) {
	ctx := context.Background()

	dst := tempTestFile(t)
	defer os.RemoveAll(filepath.Dir(dst))
	u := testModule("basic-file/foo.txt") + "?checksum=md5:09f7e02f1290be211da707a266f153b3"

	getter := &MockGetter{Proxy: new(FileGetter)}
	req := &Request{
		Src:  u,
		Dst:  dst,
		Mode: ClientModeFile,
	}
	client := &Client{
		Getters: map[string]Getter{
			"file": getter,
		},
	}

	// get the file
	op, err := client.Get(ctx, req)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if diff := cmp.Diff(&Operation{Dst: dst}, op); diff != "" {
		t.Fatalf("unexpected op: %s", diff)
	}

	if v := getter.GetFileURL.Query().Get("checksum"); v != "" {
		t.Fatalf("bad: %s", v)
	}

	// remove proxy file getter and reset GetFileCalled so that we can re-test.
	getter.Proxy = nil
	getter.GetFileCalled = false

	op, err = client.Get(ctx, req)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if diff := cmp.Diff(&Operation{Dst: dst}, op); diff != "" {
		t.Fatalf("unexpected op: %s", diff)
	}

	if getter.GetFileCalled {
		t.Fatalf("get should not have been called")
	}
}
