package getter

import (
	"fmt"
	"path/filepath"

	"github.com/hashicorp/go-getter/helper/url"
)

// Detect turns a source string into another source string if it is
// detected to be of a known pattern.
//
// The third parameter should be the list of detectors to use in the
// order to try them. If you don't want to configure this, just use
// the global Detectors variable.
//
// This is safe to be called with an already valid source string: Detect
// will just return it.
func Detect(src string, pwd string, g Getter) (string, bool, error) {
	getForce, getSrc := getForcedGetter(src)

	if getForce != "" && !g.ValidScheme(getForce){
			// Another getter is being forced
			return "", false, nil
	}

	isForcedGetter := getForce != "" && g.ValidScheme(getForce)

	// Separate out the subdir if there is one, we don't pass that to detect
	getSrc, subDir := SourceDirSubdir(getSrc)

	if !isForcedGetter {
		u, err := url.Parse(getSrc)
		if err == nil && u.Scheme != "" {
			return getSrc, g.ValidScheme(u.Scheme), nil
		}
	}

	result, ok, err := g.Detect(getSrc, pwd)
	if err != nil {
		return "", false, err
	}
	if !ok {
		// If is this is the forced getter we return true even the detection wasn't valid
		return getSrc, isForcedGetter, nil
	}

	result, detectSubdir := SourceDirSubdir(result)

	// If we have a subdir from the detection, then prepend it to our
	// requested subdir.
	if detectSubdir != "" {
		if subDir != "" {
			subDir = filepath.Join(detectSubdir, subDir)
		} else {
			subDir = detectSubdir
		}
	}

	if subDir != "" {
		u, err := url.Parse(result)
		if err != nil {
			return "", false, fmt.Errorf("Error parsing URL: %s", err)
		}
		u.Path += "//" + subDir

		// a subdir may contain wildcards, but in order to support them we
		// have to ensure the path isn't escaped.
		u.RawPath = u.Path

		result = u.String()
	}

	return result, true, nil
}
