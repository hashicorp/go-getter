package getter

import (
	"fmt"
	"path/filepath"

	"github.com/hashicorp/go-getter/helper/url"
)

// handleDetected is a helper function for the Detect(...) and CtxDetect(...)
// dispatch functions.
//
// Both dispatch functions work in the same general way:
//
//     * Each breaks-apart its input string to extract any provided 'force'
//       token and/or extract any '//some/subdir' element before supplying the
//       downstream {,Ctx}Detect methods with input to chew on.
//
//     * When a given detector indicates that it has processed the input
//       string, the dispatch function needs to re-introduce the previously
//       extracted bits before returning the reconstituted result string to
//       its caller.
//
// Given the originally extracted bits along with the result obtained from the
// detector, this function performs that reconstitution.
//
func handleDetected(detectedResult, srcGetForce, subDir string) (string, error) {
	var detectForce string
	detectForce, result := getForcedGetter(detectedResult)
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
			return "", fmt.Errorf("Error parsing URL: %s", err)
		}
		u.Path += "//" + subDir

		// a subdir may contain wildcards, but in order to support them we
		// have to ensure the path isn't escaped.
		u.RawPath = u.Path

		result = u.String()
	}

	// Preserve the forced getter if it exists. We try to use the
	// original set force first, followed by any force set by the
	// detector.
	if srcGetForce != "" {
		result = fmt.Sprintf("%s::%s", srcGetForce, result)
	} else if detectForce != "" {
		result = fmt.Sprintf("%s::%s", detectForce, result)
	}

	return result, nil
}
