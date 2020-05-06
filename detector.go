package getter

import (
	"fmt"
	"github.com/hashicorp/go-getter/helper/url"
	"path/filepath"
)

// Detect is a method used to detect Getters by validating if
// a source string is detected to be of a known pattern,
// and to transform it to a known pattern when necessary.
//
// The result is a list of possible Getters to download an artifact.
func Detect(src, pwd string, gs []Getter) (string, []Getter, error) {
	resultSrc := src
	var validGetters []Getter
	for _, getter := range gs {
		gList := []Getter{getter}
		getForce, getSrc := getForcedGetter(resultSrc)
		isForcedGetter := getForce != "" && getter.ValidScheme(getForce)

		// Separate out the subdir if there is one, we don't pass that to detect
		getSrc, subDir := SourceDirSubdir(getSrc)

		u, err := url.Parse(getSrc)
		if err == nil && u.Scheme != "" {
			if !isForcedGetter && !getter.ValidScheme(u.Scheme) {
				// Not forced getter and not valid scheme
				continue
			}
			if !isForcedGetter && getter.ValidScheme(u.Scheme) {
				// Not forced but a valid scheme for current getter
				validGetters = append(validGetters, getter)
				continue
			}
			if isForcedGetter {
				// With a forced getter and another scheme, we want to try only the force getter
				return getSrc, gList, nil
			}
		}

		result, ok, err := getter.Detect(getSrc, pwd)
		if err != nil {
			return "", nil, err
		}
		if !ok {
			if isForcedGetter {
				// Adds the forced getter to the list but keep transforming the source string
				_, resultSrc = getForcedGetter(resultSrc)
				validGetters = append(validGetters, getter)
			}
			continue
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
				return "", nil, fmt.Errorf("Error parsing URL: %s", err)
			}
			u.Path += "//" + subDir

			// a subdir may contain wildcards, but in order to support them we
			// have to ensure the path isn't escaped.
			u.RawPath = u.Path

			result = u.String()
		}

		resultSrc = result
		if getForce != "" && !isForcedGetter {
			// If there's a forced getter and it's not the current one
			// We don't append current getter to the list and try next getter
			resultSrc = fmt.Sprintf("%s::%s", getForce, result)
			continue
		}

		// this is valid by getter detection
		validGetters = append(validGetters, getter)
	}

	if len(validGetters) == 0 {
		return "", nil, fmt.Errorf("couldn't find any available getter")
	}

	_, getSrc := getForcedGetter(resultSrc)
	return getSrc, validGetters, nil
}
