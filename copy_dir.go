// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package getter

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// mode returns the file mode masked by the umask
func mode(mode, umask os.FileMode) os.FileMode {
	return mode & ^umask
}

// copyDir copies the src directory contents into dst. Both directories
// should already exist.
//
// If ignoreDot is set to true, then dot-prefixed files/folders are ignored.
func copyDir(ctx context.Context, dst string, src string, ignoreDot bool, disableSymlinks bool, umask os.FileMode) error {
	fmt.Printf("[DEBUG] copyDir: dst=%q, src=%q, ignoreDot=%v, disableSymlinks=%v\n", dst, src, ignoreDot, disableSymlinks)

	// We can safely evaluate the symlinks here, even if disabled, because they
	// will be checked before actual use in walkFn and copyFile
	var err error
	fmt.Printf("[DEBUG] copyDir: About to call filepath.EvalSymlinks(%q)\n", src)
	resolved, err := filepath.EvalSymlinks(src)
	if err != nil {
		fmt.Printf("[DEBUG] copyDir: filepath.EvalSymlinks failed: %v\n", err)
		// On Windows with Go 1.24+, EvalSymlinks may fail due to enhanced symlink handling.
		// Fall back to using the original src path if symlink evaluation fails.
		// This maintains compatibility while avoiding the godebug override.
		fmt.Printf("[DEBUG] copyDir: Using original src path as fallback: %q\n", src)
		resolved = src
	} else {
		fmt.Printf("[DEBUG] copyDir: filepath.EvalSymlinks succeeded: resolved=%q\n", resolved)
	}

	// Check if the resolved path tries to escape upward from the original
	if disableSymlinks {
		rel, err := filepath.Rel(filepath.Dir(src), resolved)
		if err != nil || filepath.IsAbs(rel) || containsDotDot(rel) {
			return ErrSymlinkCopy
		}
	}

	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("[DEBUG] copyDir walkFn: got error for path=%q: %v\n", path, err)
			return err
		}

		fmt.Printf("[DEBUG] copyDir walkFn: processing path=%q, isDir=%v\n", path, info.IsDir())

		if disableSymlinks {
			fileInfo, err := os.Lstat(path)
			if err != nil {
				fmt.Printf("[DEBUG] copyDir walkFn: os.Lstat(%q) failed: %v\n", path, err)
				return fmt.Errorf("failed to check copy file source for symlinks: %w", err)
			}
			if fileInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
				fmt.Printf("[DEBUG] copyDir walkFn: detected symlink at %q, returning ErrSymlinkCopy\n", path)
				return ErrSymlinkCopy
			}
		}

		if path == resolved {
			fmt.Printf("[DEBUG] copyDir walkFn: skipping resolved root path=%q\n", path)
			return nil
		}

		if ignoreDot && strings.HasPrefix(filepath.Base(path), ".") {
			// Skip any dot files
			fmt.Printf("[DEBUG] copyDir walkFn: skipping dot file/dir=%q\n", path)
			if info.IsDir() {
				return filepath.SkipDir
			} else {
				return nil
			}
		}

		// The "path" has the src prefixed to it. We need to join our
		// destination with the path without the src on it.
		dstPath := filepath.Join(dst, path[len(resolved):])
		fmt.Printf("[DEBUG] copyDir walkFn: calculated dstPath=%q from path=%q (len(resolved)=%d)\n", dstPath, path, len(resolved))

		// If we have a directory, make that subdirectory, then continue
		// the walk.
		if info.IsDir() {
			if path == filepath.Join(resolved, dst) {
				// dst is in src; don't walk it.
				fmt.Printf("[DEBUG] copyDir walkFn: skipping dst in src: path=%q\n", path)
				return nil
			}
			fmt.Printf("[DEBUG] copyDir walkFn: creating directory dstPath=%q\n", dstPath)
			if err := os.MkdirAll(dstPath, mode(0755, umask)); err != nil {
				fmt.Printf("[DEBUG] copyDir walkFn: os.MkdirAll(%q) failed: %v\n", dstPath, err)
				return err
			}
			fmt.Printf("[DEBUG] copyDir walkFn: directory created successfully\n")

			return nil
		}

		// If we have a file, copy the contents.
		fmt.Printf("[DEBUG] copyDir walkFn: copying file from %q to %q\n", path, dstPath)
		_, err = copyFile(ctx, dstPath, path, disableSymlinks, info.Mode(), umask)
		if err != nil {
			fmt.Printf("[DEBUG] copyDir walkFn: copyFile failed: %v\n", err)
		} else {
			fmt.Printf("[DEBUG] copyDir walkFn: copyFile succeeded\n")
		}
		return err
	}

	fmt.Printf("[DEBUG] copyDir: About to call filepath.Walk(%q, walkFn)\n", resolved)
	err = filepath.Walk(resolved, walkFn)
	if err != nil {
		fmt.Printf("[DEBUG] copyDir: filepath.Walk failed: %v\n", err)
	} else {
		fmt.Printf("[DEBUG] copyDir: filepath.Walk succeeded\n")
	}
	return err
}
