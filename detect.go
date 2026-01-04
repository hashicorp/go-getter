// Copyright IBM Corp. 2015, 2025
// SPDX-License-Identifier: MPL-2.0

package getter

import (
	"fmt"

	"github.com/hashicorp/go-getter/helper/url"
)

// Detector defines the interface that an invalid URL or a URL with a blank
// scheme is passed through in order to determine if its shorthand for
// something else well-known.
type Detector interface {
	// Detect will detect whether the string matches a known pattern to
	// turn it into a proper URL.
	Detect(string, string) (string, bool, error)
}

// Detectors is the list of detectors that are tried on an invalid URL.
// This is also the order they're tried (index 0 is first).
var Detectors []Detector

func init() {
	Detectors = []Detector{
		new(GitHubDetector),
		new(GitLabDetector),
		new(GitDetector),
		new(BitBucketDetector),
		new(S3Detector),
		new(GCSDetector),
		new(FileDetector),
	}
}

// Detect turns a source string into another source string if it is
// detected to be of a known pattern.
//
// The third parameter should be the list of detectors to use in the
// order to try them. If you don't want to configure this, just use
// the global Detectors variable.
//
// This is safe to be called with an already valid source string: Detect
// will just return it.
func Detect(src string, pwd string, ds []Detector) (string, error) {
	getForce, getSrc := getForcedGetter(src)

	// Separate out the subdir if there is one, we don't pass that to detect
	getSrc, subDir := SourceDirSubdir(getSrc)

	u, err := url.Parse(getSrc)
	if err == nil && u.Scheme != "" {
		// Valid URL
		return src, nil
	}

	for _, d := range ds {
		result, ok, err := d.Detect(getSrc, pwd)
		if err != nil {
			return "", err
		}
		if !ok {
			continue
		}

		result, err = handleDetected(result, getForce, subDir)
		if err != nil {
			return "", err
		}

		return result, nil
	}

	return "", fmt.Errorf("invalid source string: %s", src)
}
