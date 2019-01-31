package url

import (
	"fmt"
	"net/url"
	"strings"
)

// Parse parses rawURL into a URL structure.
// The rawURL may be relative or absolute.
//
// Parse is a wrapper for the Go stdlib net/url Parse function, but returns
// "safe" URLs on Windows and Unix platforms.
func Parse(rawURL string) (*url.URL, error) {
	// Make sure we're using "/" since net/url URLs are "/"-based. "/" is a
	// valid path separator for all golang known systems: See
	// https://golang.org/src/os/path_[windows|unix|plan9].go#IsPathSeparator
	// for more info.
	rawURL = strings.Replace(rawURL, string(`\`), `/`, -1)
	rawURL = strings.Replace(rawURL, string(`\\`), `/`, -1)

	var u *url.URL
	var err error

	if len(rawURL) > 1 && rawURL[1] == ':' {
		// Assume we're dealing with a windows drive letter. In which case we
		// force the 'file' scheme to avoid "net/url" URL.String() prepending
		// our url with "./".
		rawURL = "file://" + rawURL
	}

	u, err = url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	if len(u.Host) > 1 && u.Host[1] == ':' && strings.HasPrefix(rawURL, "file://") {
		// Assume we're dealing with a drive letter file path where the drive
		// letter has been parsed into the URL Host.
		u.Path = fmt.Sprintf("%s%s", u.Host, u.Path)
		u.Host = ""
	}

	// Remove leading slash for absolute file paths.
	if len(u.Path) > 2 && u.Path[0] == '/' && u.Path[2] == ':' {
		u.Path = u.Path[1:]
	}

	return u, err
}
