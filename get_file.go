package getter

import "net/url"

// FileGetter is a Getter implementation that will download a module from
// a file scheme.
type FileGetter struct {
	// Copy, if set to true, will copy data instead of using a symlink
	Copy bool
}

func (g *FileGetter) ClientMode(_ *url.URL) ClientMode {
	return ClientModeFile
}
