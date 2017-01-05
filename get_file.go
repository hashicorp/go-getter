package getter

import (
	"net/url"
	"os"
)

// FileGetter is a Getter implementation that will download a module from
// a file scheme.
type FileGetter struct {
	// Copy, if set to true, will copy data instead of using a symlink
	Copy bool
}

func (g *FileGetter) ClientMode(u *url.URL) (ClientMode, error) {
	path := u.Path
	if u.RawPath != "" {
		path = u.RawPath
	}

	// Check if the source is a directory.
	if fi, err := os.Stat(path); err == nil && fi.IsDir() {
		return ClientModeDir, nil
	}

	return ClientModeFile, nil
}
