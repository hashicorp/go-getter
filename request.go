package getter

import "net/url"

type Request struct {
	// Src is the source URL to get.
	//
	// Dst is the path to save the downloaded thing as. If Dir is set to
	// true, then this should be a directory. If the directory doesn't exist,
	// it will be created for you.
	//
	// Pwd is the working directory for detection. If this isn't set, some
	// detection may fail. Client will not default pwd to the current
	// working directory for security reasons.
	Src string
	Dst string
	Pwd string

	// Mode is the method of download the client will use. See ClientMode
	// for documentation.
	Mode ClientMode

	// Copy, in local file mode if set to true, will copy data instead of using
	// a symlink. If false, attempts to symlink to speed up the operation and
	// to lower the disk space usage. If the symlink fails, may attempt to copy
	// on windows.
	Copy bool

	// Dir, if true, tells the Client it is downloading a directory (versus
	// a single file). This distinction is necessary since filenames and
	// directory names follow the same format so disambiguating is impossible
	// without knowing ahead of time.
	//
	// WARNING: deprecated. If Mode is set, that will take precedence.
	Dir bool

	// ProgressListener allows to track file downloads.
	// By default a no op progress listener is used.
	ProgressListener ProgressTracker

	u *url.URL
}
