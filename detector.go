package getter

import (
	"fmt"
	"github.com/hashicorp/go-getter/helper/url"
	"path/filepath"
)

// GetterDetector detects the possible Getters in the list of available Getters to download an artifact
type GetterDetector struct {
	// List of possible getters that will be used for trying to download the artifact
	getters []Getter
}

// NewGetterDetector creates a chain of Getters that will be used for detection
func NewGetterDetector(getters []Getter) *GetterDetector {
	for i, g := range getters {
		if i == len(getters)-1 {
			break
		}
		g.SetNext(getters[i+1])
	}
	return &GetterDetector{getters}
}

func (g *GetterDetector) Detect(src, pwd string) (string, error) {
	// Start chain of detection to allow detecting multiple getters for an object
	if len(g.getters) > 0 {
		result, getters, err := g.getters[0].Detect(src, pwd)
		g.getters = getters
		return result, err
	}
	return "", fmt.Errorf("couldn't find any available getter")
}

// Detect is a shared method used by all of the Getters to validate if
// a source string is detected to be of a known pattern,
// and to transform it to a known pattern when necessary.
//
// This is the logic of the getters chain of detection.
// When a getter detects or not a valid source string, it will
// call the next getter that will then use this method to do the same.
func Detect(src string, pwd string, g Getter) (string, []Getter, error) {
	gList := []Getter{g}
	getForce, getSrc := getForcedGetter(src)
	isForcedGetter := getForce != "" && g.ValidScheme(getForce)

	// Separate out the subdir if there is one, we don't pass that to detect
	getSrc, subDir := SourceDirSubdir(getSrc)

	u, err := url.Parse(getSrc)
	if err == nil && u.Scheme != "" {
		if !isForcedGetter && !g.ValidScheme(u.Scheme) {
			// Not forced getter and not valid scheme
			// Keep going through the chain to find valid getters for this scheme
			return tryNextGetter(src, pwd, g)
		}
		if !isForcedGetter && g.ValidScheme(u.Scheme) {
			// Not forced but a valid scheme for current getter
			// Keep going through the chain to find other valid getters for this scheme
			_, gs, err := tryNextGetter(src, pwd, g)
			if gs != nil && err == nil {
				// append this getter to the list of valid getters
				return getSrc, append(gList, gs...), nil
			}
			return getSrc, gList, nil
		}
		if isForcedGetter {
			// With a forced getter and another scheme, we want to try only the force getter
			return getSrc, gList, nil
		}
	}

	result, ok, err := g.DetectGetter(getSrc, pwd)
	if err != nil {
		return "", nil, err
	}
	if !ok {
		// Try next of the chain
		r, gs, err := tryNextGetter(src, pwd, g)
		if isForcedGetter && err == nil {
			// Remove any forced getter from the path
			_, r = getForcedGetter(r)
			return r, gList, nil
		}
		return r, gs, err
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
			return "", gList, fmt.Errorf("Error parsing URL: %s", err)
		}
		u.Path += "//" + subDir

		// a subdir may contain wildcards, but in order to support them we
		// have to ensure the path isn't escaped.
		u.RawPath = u.Path

		result = u.String()
	}

	if getForce != "" && !isForcedGetter {
		// If there's a forced getter and it's not the current one
		// We don't append current getter to the list and try next getter
		result = fmt.Sprintf("%s::%s", getForce, result)
		_, gs, err := tryNextGetter(result, pwd, g)
		return result, gs, err
	}

	// this is valid by getter detection
	r, gs, err := tryNextGetter(result, pwd, g)
	if err != nil && gs != nil {
		return r, append(gList, gs...), nil
	}
	return result, gList, nil
}

func tryNextGetter(src string, pwd string, g Getter) (string, []Getter, error) {
	if g.Next() != nil {
		return g.Next().Detect(src, pwd)
	}
	return src, nil, nil
}
