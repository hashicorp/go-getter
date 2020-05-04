package getter

import (
	"fmt"
	"github.com/hashicorp/go-getter/helper/url"
	"path/filepath"
)

type GetterDetector struct {
	getters []Getter
}

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
	// Start chain of detection to allow detecting multiple getters for a src
	result, getters, err := g.getters[0].Detect(src, pwd)
	g.getters = getters
	return result, err
}

//func (g *GetterDetector) Mode(ctx context.Context, u *url.URL) (Mode, error) {
//	var result *multierror.Error
//	for _, getter := range g.getters {
//		mode, err := getter.Mode(ctx, u)
//		if err == nil {
//			// will return the first mode that is found and
//			// the current getter will be used for downloading the object
//			g.getters = []Getter{getter}
//			return mode, nil
//		}
//		result = multierror.Append(result, err)
//	}
//	return 0, result
//}
//
//func (g *GetterDetector) Get(ctx context.Context, req *Request) error {
//	// Mode will always return the first mode that is found successfully
//	// At this point, the getters list will always contain only 1 getter
//	return g.getters[0].Get(ctx, req)
//}
//
//func (g *GetterDetector) GetFile(ctx context.Context, req *Request) error {
//	// Mode will always return the first mode that is found successfully
//	// At this point, the getters list will always contain only 1 getter
//	return g.getters[0].GetFile(ctx, req)
//}

// Detect turns a source string into another source string if it is
// detected to be of a known pattern.
//
// The third parameter should be the list of detectors to use in the
// order to try them. If you don't want to configure this, just use
// the global Detectors variable.
//
// This is safe to be called with an already valid source string: Detect
// will just return it.
func Detect(src string, pwd string, g Getter) (string, []Getter, error) {
	gList := []Getter{g}
	getForce, getSrc := getForcedGetter(src)
	isForcedGetter := getForce != "" && g.ValidScheme(getForce)

	// Separate out the subdir if there is one, we don't pass that to detect
	getSrc, subDir := SourceDirSubdir(getSrc)

	u, err := url.Parse(getSrc)
	if err == nil && u.Scheme != "" {
		if !isForcedGetter && !g.ValidScheme(u.Scheme) {
			// Not forced and no valid scheme
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
		if r == "" && err == nil {
			r = getSrc
		}
		if isForcedGetter && err == nil {
			// Remove forced getter from the path
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
	// no forced getter, no scheme, and this is valid by detection
	return result, gList, nil
}

func tryNextGetter(src string, pwd string, g Getter) (string, []Getter, error) {
	if g.Next() != nil {
		return g.Next().Detect(src, pwd)
	}
	return "", nil, nil
}
