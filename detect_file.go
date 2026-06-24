// Copyright IBM Corp. 2015, 2025
// SPDX-License-Identifier: MPL-2.0

package getter

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// FileDetector implements Detector to detect file paths.
type FileDetector struct{}

func (d *FileDetector) Detect(src, pwd string) (string, bool, error) {
	if len(src) == 0 {
		return "", false, nil
	}

	if !filepath.IsAbs(src) {
		if pwd == "" {
			return "", true, fmt.Errorf(
				"relative paths require a module with a pwd")
		}

		// Stat the pwd to determine if its a symbolic link. If it is,
		// then the pwd becomes the original directory. Otherwise,
		// `filepath.Join` below does some weird stuff.
		//
		// We just ignore if the pwd doesn't exist. That error will be
		// caught later when we try to use the URL.
		if fi, err := os.Lstat(pwd); !os.IsNotExist(err) {
			if err != nil {
				return "", true, err
			}
			if fi.Mode()&os.ModeSymlink != 0 {
				pwd, err = filepath.EvalSymlinks(pwd)
				if err != nil {
					return "", true, err
				}

				// The symlink itself might be a relative path, so we have to
				// resolve this to have a correctly rooted URL.
				pwd, err = filepath.Abs(pwd)
				if err != nil {
					return "", true, err
				}
			}
		}

		src = filepath.Join(pwd, src)
	}

	return fmtFileURL(src), true, nil
}

func fmtFileURL(path string) string {
	if runtime.GOOS == "windows" {
		// Make sure we're using "/" on Windows. URLs are "/"-based.
		path = filepath.ToSlash(path)
		return fmt.Sprintf("file://%s", escapeBarePercent(path))
	}

	// Make sure that we don't start with "/" since we add that below.
	if path[0] == '/' {
		path = path[1:]
	}
	return fmt.Sprintf("file:///%s", escapeBarePercent(path))
}

// escapeBarePercent escapes any '%' in a filesystem path that is not already
// the start of a valid percent-encoding (%XX). Such a '%' (e.g. in a path
// like "{% if foo %}bar") is a literal filesystem character, but once the
// path is embedded in a file:// URL it would be parsed as an invalid escape
// and url.Parse would fail with "invalid URL escape". Escaping it to "%25"
// lets the URL round-trip back to the original path. Valid existing escapes
// and other URL metacharacters (such as a "?" query suffix that go-getter
// supports) are left untouched.
func escapeBarePercent(path string) string {
	if !strings.Contains(path, "%") {
		return path
	}
	var b strings.Builder
	b.Grow(len(path))
	for i := 0; i < len(path); i++ {
		if path[i] == '%' && !(i+2 < len(path) && isHex(path[i+1]) && isHex(path[i+2])) {
			b.WriteString("%25")
			continue
		}
		b.WriteByte(path[i])
	}
	return b.String()
}

func isHex(c byte) bool {
	return ('0' <= c && c <= '9') || ('a' <= c && c <= 'f') || ('A' <= c && c <= 'F')
}
