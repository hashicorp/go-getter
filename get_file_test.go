// Copyright IBM Corp. 2015, 2025
// SPDX-License-Identifier: MPL-2.0

package getter

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestFileGetter_impl(t *testing.T) {
	var _ Getter = new(FileGetter)
}

func TestFileGetter(t *testing.T) {
	g := new(FileGetter)
	dst := filepath.Join(t.TempDir(), "target")

	// With a dir that doesn't exist
	if err := g.Get(dst, testModuleURL("basic")); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	mainPath := filepath.Join(dst, "main.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify it's a symlink on Unix or junction point on Windows
	fi, err := os.Lstat(dst)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if runtime.GOOS == "windows" {
		isJunction, junctionErr := isWindowsJunctionPoint(dst)
		if junctionErr != nil {
			t.Fatalf("failed to check if destination is a junction point: %s", junctionErr)
		}
		if !isJunction {
			t.Fatal("destination is not a junction point")
		}
		// Additional verification: should be accessible as a directory
		if dirInfo, err := os.Stat(dst); err != nil || !dirInfo.IsDir() {
			t.Fatal("destination junction point is not accessible as a directory")
		}
	} else {
		// On Unix, verify it's a traditional symlink
		if fi.Mode()&os.ModeSymlink == 0 {
			t.Fatal("destination is not a symlink")
		}
	}
}

func TestFileGetter_sourceFile(t *testing.T) {
	g := new(FileGetter)
	dst := filepath.Join(t.TempDir(), "target")

	// With a source URL that is a path to a file
	u := testModuleURL("basic")
	u.Path += "/main.tf"
	if err := g.Get(dst, u); err == nil {
		t.Fatal("should error")
	}
}

func TestFileGetter_sourceNoExist(t *testing.T) {
	g := new(FileGetter)
	dst := filepath.Join(t.TempDir(), "target")

	// With a source URL that doesn't exist
	u := testModuleURL("basic")
	u.Path += "/main"
	if err := g.Get(dst, u); err == nil {
		t.Fatal("should error")
	}
}

func TestFileGetter_dir(t *testing.T) {
	g := new(FileGetter)
	dst := filepath.Join(t.TempDir(), "target")

	if err := os.MkdirAll(dst, 0755); err != nil {
		t.Fatalf("err: %s", err)
	}

	// With a dir that exists that isn't a symlink
	if err := g.Get(dst, testModuleURL("basic")); err == nil {
		t.Fatal("should error")
	}
}

func TestFileGetter_dirSymlink(t *testing.T) {
	g := new(FileGetter)
	tempBase := t.TempDir()
	dst := filepath.Join(tempBase, "dst")
	dst2 := filepath.Join(tempBase, "dst2")

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

	// With a dir that exists that isn't a symlink
	if err := g.Get(dst, testModuleURL("basic")); err != nil {
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
	dst := filepath.Join(t.TempDir(), "test-file")

	// With a dir that doesn't exist
	if err := g.GetFile(dst, testModuleURL("basic-file/foo.txt")); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	assertContents(t, dst, "Hello\n")

	// On Unix, verify it's a symlink; on Windows, just verify it works
	if runtime.GOOS != "windows" {
		fi, err := os.Lstat(dst)
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if fi.Mode()&os.ModeSymlink == 0 {
			t.Fatal("destination is not a symlink")
		}
	}
}

func TestFileGetter_GetFile_Copy(t *testing.T) {
	g := new(FileGetter)
	g.Copy = true

	dst := filepath.Join(t.TempDir(), "test-file")

	// With a dir that doesn't exist
	if err := g.GetFile(dst, testModuleURL("basic-file/foo.txt")); err != nil {
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
	assertContents(t, dst, "Hello\n")
}

// https://github.com/hashicorp/terraform/issues/8418
func TestFileGetter_percent2F(t *testing.T) {
	g := new(FileGetter)
	dst := filepath.Join(t.TempDir(), "target")

	// With a dir that doesn't exist
	if err := g.Get(dst, testModuleURL("basic%2Ftest")); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	mainPath := filepath.Join(dst, "main.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestFileGetter_ClientMode_notexist(t *testing.T) {
	g := new(FileGetter)

	u := testURL("nonexistent")
	if _, err := g.ClientMode(u); err == nil {
		t.Fatal("expect source file error")
	}
}

func TestFileGetter_ClientMode_file(t *testing.T) {
	g := new(FileGetter)

	// Check the client mode when pointed at a file.
	mode, err := g.ClientMode(testModuleURL("basic-file/foo.txt"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if mode != ClientModeFile {
		t.Fatal("expect ClientModeFile")
	}
}

func TestFileGetter_ClientMode_dir(t *testing.T) {
	g := new(FileGetter)

	// Check the client mode when pointed at a directory.
	mode, err := g.ClientMode(testModuleURL("basic"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if mode != ClientModeDir {
		t.Fatal("expect ClientModeDir")
	}
}
