// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build windows
// +build windows

package getter

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

func (g *FileGetter) Get(dst string, u *url.URL) error {
	ctx := g.Context()
	path := u.Path
	if u.RawPath != "" {
		path = u.RawPath
	}

	// The source path must exist and be a directory to be usable.
	if fi, err := os.Stat(path); err != nil {
		return fmt.Errorf("source path error: %s", err)
	} else if !fi.IsDir() {
		return fmt.Errorf("source path must be a directory")
	}

	fi, err := os.Lstat(dst)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// If the destination already exists, it must be a symlink
	if err == nil {
		mode := fi.Mode()
		if mode&os.ModeSymlink == 0 {
			return fmt.Errorf("destination exists and is not a symlink")
		}

		// Remove the destination
		if err := os.Remove(dst); err != nil {
			return err
		}
	}

	// Create all the parent directories
	if err := os.MkdirAll(filepath.Dir(dst), g.client.mode(0755)); err != nil {
		return err
	}

	sourcePath := toBackslash(path)

	// Use mklink to create a junction point
	fmt.Printf("[DEBUG] FileGetter.Get: About to run mklink /J %q %q\n", dst, sourcePath)
	output, err := exec.CommandContext(ctx, "cmd", "/c", "mklink", "/J", dst, sourcePath).CombinedOutput()
	if err != nil {
		fmt.Printf("[DEBUG] FileGetter.Get: mklink failed: %v, output: %q\n", err, output)
		return fmt.Errorf("failed to run mklink %v %v: %v %q", dst, sourcePath, err, output)
	}
	fmt.Printf("[DEBUG] FileGetter.Get: mklink succeeded, output: %q\n", output)

	// Verify what was created for debugging
	if fi, err := os.Lstat(dst); err != nil {
		fmt.Printf("[DEBUG] FileGetter.Get: os.Lstat failed after mklink: %v\n", err)
	} else {
		fmt.Printf("[DEBUG] FileGetter.Get: Created link mode: %v, ModeSymlink=%v, ModeDir=%v, ModeIrregular=%v\n",
			fi.Mode(), fi.Mode()&os.ModeSymlink != 0, fi.Mode()&os.ModeDir != 0, fi.Mode()&os.ModeIrregular != 0)
	}

	// Test our junction point detection
	if isJunction, err := isJunctionPoint(dst); err != nil {
		fmt.Printf("[DEBUG] FileGetter.Get: isJunctionPoint failed: %v\n", err)
	} else {
		fmt.Printf("[DEBUG] FileGetter.Get: isJunctionPoint result: %v\n", isJunction)
	}

	return nil
}

func (g *FileGetter) GetFile(dst string, u *url.URL) error {
	ctx := g.Context()
	path := u.Path
	if u.RawPath != "" {
		path = u.RawPath
	}

	// The source path must exist and be a directory to be usable.
	if fi, err := os.Stat(path); err != nil {
		return fmt.Errorf("source path error: %s", err)
	} else if fi.IsDir() {
		return fmt.Errorf("source path must be a file")
	}

	_, err := os.Lstat(dst)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// If the destination already exists, it must be a symlink
	if err == nil {
		// Remove the destination
		if err := os.Remove(dst); err != nil {
			return err
		}
	}

	// Create all the parent directories
	if err := os.MkdirAll(filepath.Dir(dst), g.client.mode(0755)); err != nil {
		return err
	}

	// If we're not copying, just symlink and we're done
	if !g.Copy {
		if err = os.Symlink(path, dst); err == nil {
			return err
		}
		lerr, ok := err.(*os.LinkError)
		if !ok {
			return err
		}
		switch lerr.Err {
		case syscall.ERROR_PRIVILEGE_NOT_HELD:
			// no symlink privilege, let's
			// fallback to a copy to avoid an error.
			break
		default:
			return err
		}
	}

	var disableSymlinks bool

	if g.client != nil && g.client.DisableSymlinks {
		disableSymlinks = true
	}

	// Copy
	_, err = copyFile(ctx, dst, path, disableSymlinks, 0666, g.client.umask())
	return err
}

// toBackslash returns the result of replacing each slash character
// in path with a backslash ('\') character. Multiple separators are
// replaced by multiple backslashes.
func toBackslash(path string) string {
	return strings.Replace(path, "/", "\\", -1)
}

// isJunctionPoint checks if the given path is a Windows junction point.
// Junction points are directory symbolic links on Windows, but they are not
// detected as os.ModeSymlink by Go's os.Lstat(). This function provides
// Windows-specific detection for junction points.
func isJunctionPoint(path string) (bool, error) {
	// First try the simple approach using Go 1.24+ ModeIrregular detection
	fi, err := os.Lstat(path)
	if err != nil {
		return false, err
	}

	// In Go 1.24+, junctions report as ModeIrregular
	if fi.Mode()&os.ModeIrregular != 0 {
		// Additional check: junctions should also appear as directories when stat'd
		if dirInfo, err := os.Stat(path); err == nil && dirInfo.IsDir() {
			return true, nil
		}
	}

	// Fallback to Windows API for more precise detection
	// Use GetFileAttributes Windows API to check for reparse point
	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return false, err
	}
	attrs, err := syscall.GetFileAttributes(pathPtr)
	if err != nil {
		return false, err
	}

	// Check if FILE_ATTRIBUTE_REPARSE_POINT is set
	// Junction points are reparse points with specific characteristics
	const FILE_ATTRIBUTE_REPARSE_POINT = 0x400
	const FILE_ATTRIBUTE_DIRECTORY = 0x10

	isReparsePoint := (attrs & FILE_ATTRIBUTE_REPARSE_POINT) != 0
	isDirectory := (attrs & FILE_ATTRIBUTE_DIRECTORY) != 0

	// Junction points are reparse points that are also directories
	return isReparsePoint && isDirectory, nil
}
