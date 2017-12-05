package getter

import (
	"fmt"
	"path/filepath"

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
		new(BitBucketDetector),
		new(S3Detector),
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
	fmt.Printf("Megan getSrc is %s, subDir is %s\n", getSrc, subDir)

	u, err := url.Parse(getSrc)
	if err == nil && u.Scheme != "" {
		// Valid URL
		return src, nil
	}

	for _, d := range ds {
		result, ok, err := d.Detect(getSrc, pwd)
		fmt.Printf("Megan result is %s, ok is %s, err is %s\n\n\n", result, ok, err)
		if err != nil {
			return "", err
		}
		if !ok {
			continue
		}

		var detectForce string
		detectForce, result = getForcedGetter(result)
		fmt.Printf("Megan detectForce is %s, result is %s\n", detectForce, result)
		result, detectSubdir := SourceDirSubdir(result)

		// If we have a subdir from the detection, then prepend it to our
		// requested subdir.
		if detectSubdir != "" {
			if subDir != "" {
				subDir = filepath.Join(detectSubdir, subDir)
				fmt.Printf("Megan here 1\n")
			} else {
				subDir = detectSubdir
				fmt.Printf("Megan here 2\n")
			}
			fmt.Printf("Megan here 3\n")
		}
		fmt.Printf("Megan here 4\n")

		if subDir != "" {
			fmt.Printf("Megan here 5\n")
			u, err := url.Parse(result)
			if err != nil {
				return "", fmt.Errorf("Error parsing URL: %s", err)
			}
			fmt.Printf("Megan u.Path is %s\n", u.Path)
			fmt.Printf("Megan subDir is %s\n", subDir)
			u.Path += "//" + subDir
			fmt.Printf("Megan u.Path is %s\n", u.Path)

			// a subdir may contain wildcards, but in order to support them we
			// have to ensure the path isn't escaped.
			u.RawPath = u.Path

			result = u.String()
			fmt.Printf("Megan result is %s\n", result)
		}

		// Preserve the forced getter if it exists. We try to use the
		// original set force first, followed by any force set by the
		// detector.
		if getForce != "" {
			result = fmt.Sprintf("%s::%s", getForce, result)
		} else if detectForce != "" {
			result = fmt.Sprintf("%s::%s", detectForce, result)
		}

		return result, nil
	}

	return "", fmt.Errorf("invalid source string: %s", src)
}
