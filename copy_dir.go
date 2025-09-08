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
	resolved, err := resolveSymlinks(src)
	if err != nil {
		return err
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
			return err
		}

		if disableSymlinks {
			fileInfo, err := os.Lstat(path)
			if err != nil {
				return fmt.Errorf("failed to check copy file source for symlinks: %w", err)
			}
			if fileInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
				return ErrSymlinkCopy
			}
		}

		if path == resolved {
			return nil
		}

		if ignoreDot && strings.HasPrefix(filepath.Base(path), ".") {
			// Skip any dot files
			if info.IsDir() {
				return filepath.SkipDir
			} else {
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
				return nil
			}
			if err := os.MkdirAll(dstPath, mode(0755, umask)); err != nil {
				return err
			}

			return nil
		}

		// If we have a file, copy the contents.
		_, err = copyFile(ctx, dstPath, path, disableSymlinks, info.Mode(), umask)
		return err
	}

	return filepath.Walk(resolved, walkFn)
}

// resolveSymlinks resolves symlinks with special handling for Windows junction points in Go 1.23+.
// This function provides a more robust fallback for Windows path resolution issues.
func resolveSymlinks(src string) (string, error) {
	resolved, err := filepath.EvalSymlinks(src)
	if err == nil {
		return resolved, nil
	}

	// On non-Windows platforms, EvalSymlinks should work properly
	if runtime.GOOS != "windows" {
		return "", err
	}

	// On Windows, first check if the path actually exists
	if _, statErr := os.Stat(src); statErr != nil {
		// Path doesn't exist, return the original EvalSymlinks error
		return "", err
	}

	// Path exists but EvalSymlinks failed - likely a junction point or similar
	// Check if this is a junction point
	if isJunction, junctionErr := isWindowsJunctionPointWinAPI(src); junctionErr == nil && isJunction {
		// Confirmed junction point - use original path
		return src, nil
	}

	// If junction detection failed or it's not a junction, but the path exists,
	// fall back to using the original path on Windows for compatibility
	// This maintains the original behavior of falling back when EvalSymlinks fails
	return src, nil
}
