package testing

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TempDir(t *testing.T) string {
	dir, err := ioutil.TempDir("", "tf")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if err := os.RemoveAll(dir); err != nil {
		t.Fatalf("err: %s", err)
	}

	return dir
}

func TempTestFile(t *testing.T) string {
	dir := TempDir(t)
	return filepath.Join(dir, "foo")
}

func AssertContents(t *testing.T, path string, contents string) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !reflect.DeepEqual(data, []byte(contents)) {
		t.Fatalf("bad. expected:\n\n%q\n\nGot:\n\n%q", contents, string(data))
	}
}

// TempFileWithContent writes a temporary file and returns the path and a function
// to clean it up.
func TempFileWithContent(t *testing.T, contents string) (string, func()) {
	tf, err := ioutil.TempFile("", "getter")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if _, err := io.Copy(tf, strings.NewReader(contents)); err != nil {
		t.Fatalf("err: %s", err)
	}

	tf.Close()

	path := tf.Name()
	return path, func() {
		if err := os.Remove(path); err != nil {
			t.Fatalf("err: %s", err)
		}
	}
}
