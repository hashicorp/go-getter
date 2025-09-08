// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package getter

import (
	"os"
)

// isSymlinkOrReparsePoint checks if the file mode indicates a symlink or any other
// type of reparse point (Windows-specific). This is more robust than checking only
// os.ModeSymlink as there are other types of reparse points that can link to
// directories, and os.ModeSymlink was only catching symlinks and mount points.
// By testing for both os.ModeSymlink and os.ModeIrregular, we can catch all cases.
// Ref: https://github.com/golang/go/issues/73827
func isSymlinkOrReparsePoint(mode os.FileMode) bool {
	return mode&os.ModeSymlink != 0 || mode&os.ModeIrregular != 0
}

func tmpFile(dir, pattern string) (string, error) {
	f, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return "", err
	}
	_ = f.Close()
	return f.Name(), nil
}
