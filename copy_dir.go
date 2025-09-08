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
	resolved, err := resolveSourcePath(src)
	if err != nil {
		return err
	}

	if err := validateSymlinkSafety(src, resolved, disableSymlinks); err != nil {
		return err
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

// resolveSourcePath resolves symlinks in the source path, with special handling
// for Windows junction points in Go 1.24+.
func resolveSourcePath(src string) (string, error) {
	// We can safely evaluate the symlinks here, even if disabled, because they
	// will be checked before actual use in walkFn and copyFile
	resolved, err := filepath.EvalSymlinks(src)
	if err != nil {
		return handleEvalSymlinksError(src, err)
	}
	return resolved, nil
}

// handleEvalSymlinksError handles EvalSymlinks failures, with special logic
// for Windows junction points in Go 1.24+.
func handleEvalSymlinksError(src string, originalErr error) (string, error) {
	// On Windows with Go 1.24+, EvalSymlinks may fail specifically for junction points
	// due to enhanced symlink handling. Check if this is a junction point before falling back.
	if runtime.GOOS != "windows" {
		// On non-Windows platforms, EvalSymlinks should work properly
		return "", originalErr
	}

	isJunction, junctionErr := isWindowsJunctionPoint(src)
	if junctionErr != nil || !isJunction {
		// Not a junction point or detection failed - propagate the original error
		// This ensures real errors (permissions, network, etc.) are reported properly
		return "", originalErr
	}

	// This is a junction point that EvalSymlinks can't handle in Go 1.24+
	// Use the original path since junctions are safe directory links
	return src, nil
}

// validateSymlinkSafety checks if the resolved path tries to escape upward
// from the original when symlinks are disabled.
func validateSymlinkSafety(src, resolved string, disableSymlinks bool) error {
	if !disableSymlinks {
		return nil
	}

	rel, err := filepath.Rel(filepath.Dir(src), resolved)
	if err != nil || filepath.IsAbs(rel) || containsDotDot(rel) {
		return ErrSymlinkCopy
	}
	return nil
}

// isWindowsJunctionPoint detects Windows junction points for cross-platform compatibility.
// This is a simplified version that works across different Go versions.
func isWindowsJunctionPoint(path string) (bool, error) {
	if runtime.GOOS != "windows" {
		return false, nil
	}

	// Check if it's a directory with irregular mode bits (Go 1.24+ detection)
	fi, err := os.Lstat(path)
	if err != nil {
		return false, err
	}

	// In Go 1.24+, junction points report as ModeIrregular
	if fi.Mode()&os.ModeIrregular != 0 {
		// Additional check: junctions should also appear as directories when stat'd
		// Use os.Stat (not Lstat) to follow the junction and check if target is a directory
		if dirInfo, err := os.Stat(path); err == nil && dirInfo.IsDir() {
			return true, nil
		}
	}

	return false, nil
}
