package getter

import (
	"fmt"
	"github.com/hashicorp/go-getter/helper/url"
	"path/filepath"
)

// Detect is a method used to detect if a Getter is a candidate for downloading an artifact
// by validating if a source string is detected to be of a known pattern,
// and transforming it to a known pattern when necessary.
func Detect(req *Request, getter Getter) (bool, error) {
	originalSrc := req.Src

	getForce, getSrc := getForcedGetter(req.Src)
	req.Forced = getForce

	// Separate out the subdir if there is one, we don't pass that to detect
	getSrc, subDir := SourceDirSubdir(getSrc)

	req.Src = getSrc
	ok, err := getter.Detect(req)
	if err != nil {
		return true, err
	}
	if !ok {
		// Write back the original source
		req.Src = originalSrc
		return ok, nil
	}

	result, detectSubdir := SourceDirSubdir(req.Src)

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
			return true, fmt.Errorf("Error parsing URL: %s", err)
		}
		u.Path += "//" + subDir

		// a subdir may contain wildcards, but in order to support them we
		// have to ensure the path isn't escaped.
		u.RawPath = u.Path

		result = u.String()
	}

	req.Src = result
	return true, nil
}
