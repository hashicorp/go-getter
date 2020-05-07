package getter

import (
	"fmt"
	"github.com/hashicorp/go-getter/helper/url"
	"path/filepath"
)

// Detect is a method used to detect if a Getter is a candidate for downloading and artifact
// by validating if a source string is detected to be of a known pattern,
// and to transform it to a known pattern when necessary.
func Detect(req *Request, getter Getter) (string, bool, error) {
	originalSrc := req.Src

	getForce, getSrc := getForcedGetter(req.Src)

	if getForce != "" {
		// There's a getter being forced
		if !getter.ValidScheme(getForce) {
			// Current getter is not the forced one
			// Don't use it to try to download the artifact
			return "", false, nil
		}
	}

	isForcedGetter := getForce != "" && getter.ValidScheme(getForce)

	// Separate out the subdir if there is one, we don't pass that to detect
	getSrc, subDir := SourceDirSubdir(getSrc)

	u, err := url.Parse(getSrc)
	if err == nil && u.Scheme != "" {
		if isForcedGetter {
			// Is the forced getter and source is a valid url
			return getSrc, true, nil
		}
		if getter.ValidScheme(u.Scheme) {
			return getSrc, true, nil
		} else {
			// Valid url with a scheme that is not valid for current getter
			return "", false, nil
		}
	}
	req.Src = getSrc
	result, ok, err := getter.Detect(req)
	if err != nil {
		return "", true, err
	}
	if !ok {
		if isForcedGetter {
			// Is the forced getter then should be used to download the artifact
			if req.Pwd != "" && !filepath.IsAbs(getSrc) {
				// Make sure to add pwd to relative paths
				getSrc = filepath.Join(req.Pwd, getSrc)
			}
			return getSrc, true, nil
		}
		// Write back the original source
		req.Src = originalSrc
		return "", ok, nil
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
			return "", true, fmt.Errorf("Error parsing URL: %s", err)
		}
		u.Path += "//" + subDir

		// a subdir may contain wildcards, but in order to support them we
		// have to ensure the path isn't escaped.
		u.RawPath = u.Path

		result = u.String()
	}

	return result, true, nil
}
