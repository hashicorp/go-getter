package getter

import (
	"fmt"
	"os/exec"
	"strings"
	"syscall"
)

func SymlinkAny(oldname, newname string) error {
	sourcePath := toBackslash(oldname)

	// Use mklink to create a junction point
	output, err := exec.Command("cmd", "/c", "mklink", "/J", newname, sourcePath).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run mklink %v %v: %v %q", newname, sourcePath, err, output)
	}
	return nil
}

var ErrUnauthorized = syscall.ERROR_PRIVILEGE_NOT_HELD

// toBackslash returns the result of replacing each slash character
// in path with a backslash ('\') character. Multiple separators are
// replaced by multiple backslashes.
func toBackslash(path string) string {
	return strings.Replace(path, "/", "\\", -1)
}
