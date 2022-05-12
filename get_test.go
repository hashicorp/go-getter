package getter

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	testing_helper "github.com/hashicorp/go-getter/v2/helper/testing"
)

func TestGet_badSchema(t *testing.T) {
	ctx := context.Background()

	dst := testing_helper.TempDir(t)
	u := testModule("basic")
	u = "nope::" + u

	op, err := Get(ctx, dst, u)
	if err == nil {
		t.Fatal("should error")
	}
	if op != nil {
		t.Fatal("op should be nil")
	}
}

func TestGet_file(t *testing.T) {
	ctx := context.Background()

	dst := testing_helper.TempDir(t)
	u := testModule("basic")

	op, err := Get(ctx, dst, u)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if diff := cmp.Diff(&GetResult{Dst: dst}, op); diff != "" {
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

	dst := testing_helper.TempDir(t)
	u := testModule("basic-tgz")

	op, err := Get(ctx, dst, u)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if diff := cmp.Diff(&GetResult{Dst: dst}, op); diff != "" {
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

	dst := testing_helper.TempDir(t)
	u := testModule("basic%2Ftest")

	op, err := Get(ctx, dst, u)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if diff := cmp.Diff(&GetResult{Dst: dst}, op); diff != "" {
		t.Fatalf("unexpected op: %s", diff)
	}

	mainPath := filepath.Join(dst, "main.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestGet_fileDetect(t *testing.T) {
	ctx := context.Background()

	dst := testing_helper.TempDir(t)
	u := filepath.Join(".", "testdata", "basic")
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	req := &Request{
		Src:     u,
		Dst:     dst,
		Pwd:     pwd,
		GetMode: ModeAny,
	}
	client := &Client{}

	if err := client.configure(); err != nil {
		t.Fatalf("configure: %s", err)
	}

	op, err := client.Get(ctx, req)
	if err != nil {
		t.Fatalf("get: %s", err)
	}
	if diff := cmp.Diff(&GetResult{Dst: dst}, op); diff != "" {
		t.Fatalf("unexpected op: %s", diff)
	}

	mainPath := filepath.Join(dst, "main.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("stat: %s", err)
	}
}

func TestGet_fileForced(t *testing.T) {
	ctx := context.Background()

	dst := testing_helper.TempDir(t)
	u := testModule("basic")
	u = "file::" + u

	op, err := Get(ctx, dst, u)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if diff := cmp.Diff(&GetResult{Dst: dst}, op); diff != "" {
		t.Fatalf("unexpected op: %s", diff)
	}

	mainPath := filepath.Join(dst, "main.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestGet_fileSubdir(t *testing.T) {
	ctx := context.Background()

	dst := testing_helper.TempDir(t)
	u := testModule("basic//subdir")

	op, err := Get(ctx, dst, u)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if diff := cmp.Diff(&GetResult{Dst: dst}, op); diff != "" {
		t.Fatalf("unexpected op: %s", diff)
	}

	mainPath := filepath.Join(dst, "sub.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestGet_archive(t *testing.T) {
	ctx := context.Background()

	dst := testing_helper.TempDir(t)
	u := filepath.Join("./testdata", "archive.tar.gz")
	u, _ = filepath.Abs(u)

	op, err := Get(ctx, dst, u)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if diff := cmp.Diff(&GetResult{Dst: dst}, op); diff != "" {
		t.Fatalf("unexpected op: %s", diff)
	}

	mainPath := filepath.Join(dst, "main.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestGetAny_archive(t *testing.T) {
	ctx := context.Background()

	dst := testing_helper.TempDir(t)
	u := filepath.Join("./testdata", "archive.tar.gz")
	u, _ = filepath.Abs(u)

	op, err := GetAny(ctx, dst, u)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if diff := cmp.Diff(&GetResult{Dst: dst}, op); diff != "" {
		t.Fatalf("unexpected op: %s", diff)
	}

	mainPath := filepath.Join(dst, "main.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestGet_archiveRooted(t *testing.T) {
	ctx := context.Background()

	dst := testing_helper.TempDir(t)
	u := testModule("archive-rooted/archive.tar.gz")
	op, err := Get(ctx, dst, u)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if diff := cmp.Diff(&GetResult{Dst: dst}, op); diff != "" {
		t.Fatalf("unexpected op: %s", diff)
	}

	mainPath := filepath.Join(dst, "root", "hello.txt")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestGet_archiveSubdirWild(t *testing.T) {
	ctx := context.Background()

	dst := testing_helper.TempDir(t)
	u := testModule("archive-rooted/archive.tar.gz")
	u += "//*"
	op, err := Get(ctx, dst, u)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if diff := cmp.Diff(&GetResult{Dst: dst}, op); diff != "" {
		t.Fatalf("unexpected op: %s", diff)
	}

	mainPath := filepath.Join(dst, "hello.txt")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestGet_archiveSubdirWildMultiMatch(t *testing.T) {
	ctx := context.Background()

	dst := testing_helper.TempDir(t)
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
			t.Fatal("GetResult should be nil")
		}
	}
}

func TestGetAny_file(t *testing.T) {
	ctx := context.Background()

	dst := testing_helper.TempDir(t)
	u := testModule("basic-file/foo.txt")

	if _, err := GetAny(ctx, dst, u); err != nil {
		t.Fatalf("err: %s", err)
	}

	mainPath := filepath.Join(dst, "foo.txt")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestGetAny_dir(t *testing.T) {
	ctx := context.Background()

	dst := testing_helper.TempDir(t)
	u := filepath.Join("./testdata", "basic")
	u, _ = filepath.Abs(u)

	op, err := GetAny(ctx, dst, u)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if diff := cmp.Diff(&GetResult{Dst: dst}, op); diff != "" {
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

	dst := testing_helper.TempTestFile(t)
	defer os.RemoveAll(filepath.Dir(dst))
	u := testModule("basic-file/foo.txt")

	op, err := GetFile(ctx, dst, u)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if diff := cmp.Diff(&GetResult{Dst: dst}, op); diff != "" {
		t.Fatalf("unexpected op: %s", diff)
	}

	// Verify the main file exists
	testing_helper.AssertContents(t, dst, "Hello\n")
}

func TestGetFile_archive(t *testing.T) {
	ctx := context.Background()

	dst := testing_helper.TempTestFile(t)
	defer os.RemoveAll(filepath.Dir(dst))
	u := testModule("basic-file-archive/archive.tar.gz")

	op, err := GetFile(ctx, dst, u)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if diff := cmp.Diff(&GetResult{Dst: dst}, op); diff != "" {
		t.Fatalf("unexpected op: %s", diff)
	}

	// Verify the main file exists
	testing_helper.AssertContents(t, dst, "Hello\n")
}
func TestGetFile_filename_path_traversal(t *testing.T) {
	dst := testing_helper.TempDir(t)
	u := testModule("basic-file/foo.txt")

	u += "?filename=../../../../../../../../../../../../../tmp/bar.txt"

	ctx := context.Background()
	op, err := GetAny(ctx, dst, u)

	if op != nil {
		t.Fatalf("unexpected op: %v", op)
	}

	if err == nil {
		t.Fatalf("expected error")
	}

	if !strings.Contains(err.Error(), "filename query parameter contain path traversal") {
		t.Fatalf("unexpected err: %s", err)
	}
}

func TestGetFile_archiveChecksum(t *testing.T) {
	ctx := context.Background()

	dst := testing_helper.TempTestFile(t)
	defer os.RemoveAll(filepath.Dir(dst))
	u := testModule(
		"basic-file-archive/archive.tar.gz?checksum=md5:fbd90037dacc4b1ab40811d610dde2f0")

	op, err := GetFile(ctx, dst, u)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if diff := cmp.Diff(&GetResult{Dst: dst}, op); diff != "" {
		t.Fatalf("unexpected op: %s", diff)
	}

	// Verify the main file exists
	testing_helper.AssertContents(t, dst, "Hello\n")
}

func TestGetFile_archiveNoUnarchive(t *testing.T) {
	ctx := context.Background()

	dst := testing_helper.TempTestFile(t)
	defer os.RemoveAll(filepath.Dir(dst))
	u := testModule("basic-file-archive/archive.tar.gz")
	u += "?archive=false"

	op, err := GetFile(ctx, dst, u)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if diff := cmp.Diff(&GetResult{Dst: dst}, op); diff != "" {
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
			dst := testing_helper.TempTestFile(t)
			defer os.RemoveAll(filepath.Dir(dst))
			op, err := GetFile(ctx, dst, u)
			if (err != nil) != tc.Err {
				t.Fatalf("append: %s\n\nerr: %s", tc.Append, err)
			}
			if err == nil {
				if diff := cmp.Diff(&GetResult{Dst: dst}, op); diff != "" {
					t.Fatalf("unexpected dst: %s", diff)
				}
			}

			// Verify the main file exists
			testing_helper.AssertContents(t, dst, "Hello\n")
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

		// sha512
		{
			"?checksum=file:" + checksums + "/CHECKSUM_sha256_gpg",
			true,
			false,
		},
	}

	for _, tc := range cases {
		u := checksums + "/content.txt" + tc.Append
		t.Run(tc.Append, func(t *testing.T) {
			ctx := context.Background()

			dst := testing_helper.TempTestFile(t)
			defer os.RemoveAll(filepath.Dir(dst))
			op, err := GetFile(ctx, dst, u)
			if (err != nil) != tc.WantErr {
				t.Fatalf("append: %s\n\nerr: %s", tc.Append, err)
			}
			if err == nil {
				if diff := cmp.Diff(&GetResult{Dst: dst}, op); diff != "" {
					t.Fatalf("unexpected dst: %s", diff)
				}
			}

			if tc.WantTransfer {
				// Verify the main file exists
				testing_helper.AssertContents(t, dst, "I am a file with some content\n")
			}
		})
		return
	}
}

func TestGetFile_checksumURL(t *testing.T) {
	ctx := context.Background()

	dst := testing_helper.TempTestFile(t)
	defer os.RemoveAll(filepath.Dir(dst))
	u := testModule("basic-file/foo.txt") + "?checksum=md5:09f7e02f1290be211da707a266f153b3"

	getter := &MockGetter{Proxy: new(FileGetter)}
	req := &Request{
		Src:     u,
		Dst:     dst,
		GetMode: ModeFile,
	}
	client := &Client{
		Getters: []Getter{getter},
	}

	op, err := client.Get(ctx, req)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if diff := cmp.Diff(&GetResult{Dst: dst}, op); diff != "" {
		t.Fatalf("unexpected op: %s", diff)
	}

	if v := getter.GetFileURL.Query().Get("checksum"); v != "" {
		t.Fatalf("bad: %s", v)
	}
}

func TestGetFile_filename(t *testing.T) {
	ctx := context.Background()

	dst := testing_helper.TempDir(t)
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

	dst := testing_helper.TempTestFile(t)
	defer os.RemoveAll(filepath.Dir(dst))
	u := testModule("basic-file/foo.txt") + "?checksum=md5:09f7e02f1290be211da707a266f153b3"

	getter := &MockGetter{Proxy: new(FileGetter)}
	req := &Request{
		Src:     u,
		Dst:     dst,
		GetMode: ModeFile,
	}
	client := &Client{
		Getters: []Getter{getter},
	}

	// get the file
	op, err := client.Get(ctx, req)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if diff := cmp.Diff(&GetResult{Dst: dst}, op); diff != "" {
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
	if diff := cmp.Diff(&GetResult{Dst: dst}, op); diff != "" {
		t.Fatalf("unexpected op: %s", diff)
	}

	if getter.GetFileCalled {
		t.Fatalf("get should not have been called")
	}
}

func TestGetFile_inplace(t *testing.T) {
	ctx := context.Background()

	dst := testing_helper.TempTestFile(t)
	defer os.RemoveAll(filepath.Dir(dst))
	src := testModule("basic-file/foo.txt")

	getter := &MockGetter{Proxy: new(FileGetter)}
	req := &Request{
		Src:     src + "?checksum=md5:09f7e02f1290be211da707a266f153b3",
		Dst:     dst,
		GetMode: ModeFile,
		Inplace: true,
	}
	client := &Client{
		Getters: []Getter{getter},
	}

	// get the file
	op, err := client.Get(ctx, req)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if diff := cmp.Diff(&GetResult{Dst: strings.ReplaceAll(src, "file://", "")}, op); diff != "" {
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
	if diff := cmp.Diff(&GetResult{Dst: strings.ReplaceAll(src, "file://", "")}, op); diff != "" {
		t.Fatalf("unexpected op: %s", diff)
	}

	if getter.GetFileCalled {
		t.Fatalf("get should not have been called")
	}
}

func TestGetFile_inplace_badChecksum(t *testing.T) {
	ctx := context.Background()

	dst := testing_helper.TempTestFile(t)
	defer os.RemoveAll(filepath.Dir(dst))
	src := testModule("basic-file/foo.txt")

	getter := &MockGetter{Proxy: new(FileGetter)}
	req := &Request{
		Src:     src + "?checksum=md5:09f7e02f1290be211da707a266f153b4",
		Dst:     dst,
		GetMode: ModeFile,
		Inplace: true,
	}
	client := &Client{
		Getters: []Getter{getter},
	}

	// get the file
	op, err := client.Get(ctx, req)
	if err == nil {
		t.Fatalf("err is nil")
	}
	if _, ok := err.(*ChecksumError); !ok {
		t.Fatalf("err is not a checksum error: %v", err)
	}
	if op != nil {
		t.Fatalf("op is not nil")
	}
}

func TestgetForcedGetter(t *testing.T) {
	type args struct {
		src string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 string
	}{
		{"s3 AWSv1234",
			args{src: "s3::https://s3-eu-west-1.amazonaws.com/bucket/foo/bar.baz?version=1234"},
			"s3", "https://s3-eu-west-1.amazonaws.com/bucket/foo/bar.baz?version=1234",
		},
		{"s3 localhost-1",
			args{src: "s3::http://127.0.0.1:9000/test-bucket/hello.txt?aws_access_key_id=TESTID&aws_access_key_secret=TestSecret&region=us-east-2&version=1"},
			"s3", "http://127.0.0.1:9000/test-bucket/hello.txt?aws_access_key_id=TESTID&aws_access_key_secret=TestSecret&region=us-east-2&version=1",
		},
		{"s3 localhost-2",
			args{src: "s3::http://127.0.0.1:9000/test-bucket/hello.txt?aws_access_key_id=TESTID&aws_access_key_secret=TestSecret&version=1"},
			"s3", "http://127.0.0.1:9000/test-bucket/hello.txt?aws_access_key_id=TESTID&aws_access_key_secret=TestSecret&version=1",
		},
		{"s3 localhost-3",
			args{src: "s3::http://127.0.0.1:9000/test-bucket/hello.txt?aws_access_key_id=TESTID&aws_access_key_secret=TestSecret"},
			"s3", "http://127.0.0.1:9000/test-bucket/hello.txt?aws_access_key_id=TESTID&aws_access_key_secret=TestSecret",
		},

		{
			"gcs test1",
			args{"gcs::https://www.googleapis.com/storage/v1/go-getter-test/go-getter/foo/null.zip"},
			"gcs", "https://www.googleapis.com/storage/v1/go-getter-test/go-getter/foo/null.zip",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := getForcedGetter(tt.args.src)
			if got != tt.want {
				t.Errorf("getForcedGetter() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("getForcedGetter() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
