package getter

import (
	"fmt"

	"github.com/hashicorp/go-getter/helper/url"
)

// CtxDetector (read: "Contextual Detector"), the evil twin of Detector.
//
// Like its Detector sibling, CtxDetector defines an interface that an invalid
// URL or a URL with a blank scheme can be passed through in order to
// determine if its shorthand for something else well-known.
//
// CtxDetector expands on the capabilities of Detector in the following ways:
//
//    * A CtxDetector allows the caller to provide more information about its
//      invocation context to the CtxDetect dispatch function. This allows for
//      some types of useful detections and transformations that were not
//      previously possible.
//
//    * The CtxDetect dispatch function provides the CtxDetector
//      implementations with all of the context information it has available
//      to it, including the force token flag (e.g., "git::"). This allows the
//      implementations to safely take (or avoid taking) actions that would be
//      unsafe otherwise.
//
// A CtxDetector is slightly more cumbersome to use than Detector. Callers can
// (and should) continue to use Detector unless the enhanced capabilities of
// one or more of the CtxDetector implementation is needed. At the time of
// writing (2020-08), the only CtxDetector with such extra mojo is
// GitCtxDetector (q.v.).
//
type CtxDetector interface {

	// CtxDetect will detect whether the string matches a known pattern to
	// turn it into a proper URL
	//
	// 'src' (required) is the input string to be interpretted. In the common
	// case this value will have been preparsed by the CtxDetect dispatch
	// function; its forcing token (if any) will have been removed; same for
	// any 'go-getter' subdir portion('//some/subdir'). Some examples:
	//
	//     "s3-eu-west-1.amazonaws.com/bucket/foo/bar.baz?version=1234"
	//
	//     "github.com/hashicorp/foo.git"
	//
	//     "git@github.com:hashicorp/foo.git?foo=bar"
	//
	//     "../../git-submods/tf-mods/some-tf-module?ref=v1.2.3"
	//
	// 'pwd' (optional, sometimes) is the filepath that should be taken as the
	// current working directory (mainly for the purpose of resolving
	// filesystem paths; may be overridden for that purpose by
	// 'srcResolveFrom'). Some CtxDetector implementation may require this
	// path to be an abosolute filepath.
	//
	// 'forceToken' (optional) is the forcing token, if any, extracted from
	// the input string submitted to the CtxDetect dispatch function. It is
	// provided as a param to the CtxDetect method so that CtxDetector
	// implementations may recognize 'src' strings intended for them. This
	// removes ambiguity when a given 'src' value could be legitimately
	// processed by more than one CtxDetector implementation.
	//
	// 'ctxSubDir' (optional) is the 'go-getter' subdir portion (if any)
	// pre-extracted from the source string (as noted above). It is provided
	// to the CtxDetector implementation only for contextual awareness, which
	// conceivably could inform its decision-making process. It should not be
	// incorporated into the result returned by the CtxDetector impl.
	//
	// 'srcResolveFrom' (optional, sometimes) A caller-provided filepath to be
	// used as the directory from which any relative filepath in 'src' should
	// be resolved, instead of relative to 'pwd'. An individual CtxDetector
	// implementation may require that this value be absolute.
	//
	// Protocol: Where they need to be resolved, relative filepath values in
	//           'src' will be resolved relative to 'pwd', unless
	//           'srcResolveFrom' is non-empty; then they will be resolved
	//           relative to 'srcResolveFrom'.
	//
	//           Note that some CtxDetector impls. (FileCtxDetector,
	//           GitCtxDetector) can only produce meaningful results in some
	//           circumstances if they have an absolute directory to resolve
	//           to. For best results, when 'srcResolveFrom' is non-empty,
	//           provide an absolute filepath.
	//
	//           The CtxDetect interface itself does not require that either
	//           'pwd' or 'srcResolveFrom' be absolute filepaths, but that
	//           might be required by a particular CtxDetector implementation.
	//           Know that RFC-compliant use of 'file://' URIs (which some
	//           CtxDetector impls. emit) permit only absolute filepaths, and
	//           tools (such as Git) expect this. Providing relative filepaths
	//           for 'pwd' and/or 'srcResolveFrom' may result in the
	//           generation of non-legit 'file://' URIs with relative paths in
	//           them, and a CtxDetector implementation is permitted to reject
	//           them with an error if it requires an absolute path.
	//
	CtxDetect(src, pwd, forceToken, ctxSubDir, srcResolveFrom string) (string, bool, error)
}

// ContextualDetectors is the list of detectors that are tried on an invalid URL.
// This is also the order they're tried (index 0 is first).
var ContextualDetectors []CtxDetector

func init() {
	ContextualDetectors = []CtxDetector{
		new(GitHubCtxDetector),
		new(GitLabCtxDetector),
		new(GitCtxDetector),
		new(BitBucketCtxDetector),
		new(S3CtxDetector),
		new(GCSCtxDetector),
		new(FileCtxDetector),
	}
}

// CtxDetect turns a source string into another source string if it is
// detected to be of a known pattern.
//
// An empty-string value provided for 'pwd' is interpretted as "not
// provided". Likewise for 'srcResolveFrom'.
//
// The (optional) 'srcResolveFrom' parameter allows the caller to provide a
// directory from which any relative filepath in 'src' should be resolved,
// instead of relative to 'pwd'. This supports those use cases (e.g.,
// Terraform modules with relative 'source' filepaths) where the caller
// context for path resolution may be different than the pwd. For best result,
// the provided value should be an absolute filepath. If unneeded, use specify
// the empty string.
//
// The 'cds' []CtxDetector parameter should be the list of detectors to use in
// the order to try them. If you don't want to configure this, just use the
// global ContextualDetectors variable.
//
// This is safe to be called with an already valid source string: CtxDetect
// will just return it.
//
func CtxDetect(src, pwd, srcResolveFrom string, cds []CtxDetector) (string, error) {

	getForce, getSrc := getForcedGetter(src)

	// Separate out the subdir if there is one, we don't pass that to detect
	getSrc, subDir := SourceDirSubdir(getSrc)

	u, err := url.Parse(getSrc)
	if err == nil && u.Scheme != "" {
		// Valid URL
		return src, nil
	}

	for _, d := range cds {
		result, ok, err := d.CtxDetect(getSrc, pwd, getForce, subDir, srcResolveFrom)
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
