package getter

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

// copyDir copies the src directory contents into dst. Both directories
// should already exist.
//
// If ignoreDot is set to true, then dot-prefixed files/folders are ignored.
func copyDir(ctx context.Context, dst string, src string, ignoreDot bool, disableSymlinks bool, umask os.FileMode) error {
	resolved, err := filepath.EvalSymlinks(src)
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

		if disableSymlinks {
			if info.Mode()&os.ModeSymlink == os.ModeSymlink {
				return ErrSymlinkCopy
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
