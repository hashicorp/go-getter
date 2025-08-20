// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package getter

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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
	// We can safely evaluate the symlinks here, even if disabled, because they
	// will be checked before actual use in walkFn and copyFile
	var err error
	fmt.Println("copyDir - 27", src, dst)
	resolved, err := filepath.EvalSymlinks(src)
	if err != nil {
		if runtime.GOOS == "windows" {
			resolved = src
		} else {
			return err
		}
	}

	// Check if the resolved path tries to escape upward from the original
	if disableSymlinks {
		rel, err := filepath.Rel(filepath.Dir(src), resolved)
		fmt.Println("copyDir - 36", rel, filepath.Dir(src), resolved)
		if err != nil || filepath.IsAbs(rel) || containsDotDot(rel) {
			return ErrSymlinkCopy
		}
	}

	walkFn := func(path string, info os.FileInfo, err error) error {
		fmt.Println("copyDir - 43", path)
		if err != nil {
			return err
		}

		if disableSymlinks {
			fileInfo, err := os.Lstat(path)
			fmt.Println("copyDir - 50", fileInfo, err)
			if err != nil {
				return fmt.Errorf("failed to check copy file source for symlinks: %w", err)
			}
			fmt.Println("copyDir - 54", fileInfo.Mode())
			if fileInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
				return ErrSymlinkCopy
			}
		}

		fmt.Println("copyDir - 58", path, resolved)
		if path == resolved {
			return nil
		}

		if ignoreDot && strings.HasPrefix(filepath.Base(path), ".") {
			// Skip any dot files
			fmt.Println("copyDir - 66", path)
			if info.IsDir() {
				fmt.Println("copyDir - 69", path, "is a dir, skipping")
				return filepath.SkipDir
			} else {
				fmt.Println("copyDir - 72", path, "is a file, skipping")
				return nil
			}
		}

		// The "path" has the src prefixed to it. We need to join our
		// destination with the path without the src on it.
		dstPath := filepath.Join(dst, path[len(resolved):])

		// If we have a directory, make that subdirectory, then continue
		// the walk.
		if info.IsDir() {
			if path == filepath.Join(resolved, dst) {
				// dst is in src; don't walk it.
				fmt.Println("copyDir - 84", path, "is the dst, skipping")
				return nil
			}
			if err := os.MkdirAll(dstPath, mode(0755, umask)); err != nil {
				fmt.Println("copyDir - 88", err)
				return err
			}

			fmt.Println("copyDir - 92", dstPath, "is a dir, continuing walk")
			return nil
		}

		// If we have a file, copy the contents.
		fmt.Println("copyDir - 97", dstPath, "is a file, copying")
		_, err = copyFile(ctx, dstPath, path, disableSymlinks, info.Mode(), umask)
		fmt.Println("copyDir - 100", err)
		return err
	}

	fmt.Println("copyDir - 103", resolved)
	return filepath.Walk(resolved, walkFn)
}
