package getter

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
)

// SmbDetector implements Detector to detect smb paths with // for Windows OS.
type SmbDetector struct{}

func (d *SmbDetector) Detect(src, pwd string) (string, bool, error) {
	if len(src) == 0 {
		return "", false, nil
	}

	if windowsSmbPath(src) {
		// This is a valid smb path for Windows and will be checked in the SmbGetter
		// by the file system using the FileGetter, if available.
		src = filepath.ToSlash(src)
		return fmt.Sprintf("smb:%s", src), true, nil
	}
	return "", false, nil
}

func windowsSmbPath(path string) bool {
	return runtime.GOOS == "windows" && (strings.HasPrefix(path, "\\\\") || strings.HasPrefix(path, "//"))
}
