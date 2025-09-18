package getter

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
)

// GitCtxDetector implements CtxDetector to detect Git SSH URLs such as:
//
//     git@host.com:dir1/dir2
//
// and 'git::' forced filepaths such as:
//
//     git::../../path/to/some/git/repo//some/subdir?ref=v1.2.3
//
// and convert them to proper URLs that the GitGetter can understand.
//
// The go-getter force token for this CtxDetector implementation is 'git::'.
//
// (If using the sibling Detector interface rather than CtxDetector, see
// GitDetector, instead, which uses the same 'git::' force key).
//
// If a force token is provided but it is something other than 'git::', then
// this module's CtxDetect method will not attempt to interpret the 'src'
// string. In that case it is effectively a NOOP.
//
// If a 'git::' force token is provided, but this module is not able to
// interpet the 'src' string, then an error will be returned.
//
// If no force token is provided, then this module will attempt to determine
// whether the 'src' string can safely be determined to be a Git SSH URL (but
// no filepath detection can be performed).
//
// Details
// -------
// This CtxDetector implementation attempts to detect Git SSH URLs such as:
//
//     git@host.com:dir1/dir2
//
// and also filepath strings in the form:
//
//     git::<filepath>
//
// where <filepath> is either an absolute or relative filepath. For relative
// filepaths, be sure to see the notes on 'srcResolveFrom' below. Some
// examples:
//
// Absolute:
//
//     git::/path/to/some/git/repo
//
//     git::/path/to/some/git/repo//some/subdir
//
//     git::/path/to/some/git/repo//some/subdir?ref=v1.2.3
//
// Relative (subdir):
//
//     git::./path/to/some/git/repo//some/subdir?ref=v1.2.3
//
// Relative (parent dir):
//
//     git::../../path/to/some/git/repo//some/subdir?ref=v1.2.3
//
// Filepaths are transformed into 'file://' URIs that can be understood by
// GitGetter. The path will be properly encoded and will always be an absolute
// filepath. If an absolute filepath cannot be produced, then an error is
// emitted.
//
type GitCtxDetector struct{}

// CtxDetect is a method defined on the CtxDetector interface.
//
// Note that the 'git::' force token present in some of the above examples
// will have been stripped off of our 'src' param before this method is
// invoked. However, if it was provided, then we detect its presence here, by
// the provided 'forceToken' param, which will be contain just "git".
//
// In the specific case of detecting filepath strings, this module will only
// recognize them if they were specified with the 'git::' forcing
// token. Otherwise it would not be clear that they are supposed to be Git
// repos.
//
// Param Notes:
// ------------
//
// 'src' (required)
//
// 'pwd' (optional) Relative filepaths are resolved relative to this location
//       by default, but can be overridden by the 'srcResolveFrom' param.
//
// 'forceToken' (optional) Can be anything. If non-empty an "git", then src is
//       explicitly intended for us to process. If non-empty but is some value
//       other than "git", then src is intended for some other detector module
//       (so we won't interfere). If empty, then we will attempt to detect
//       what is in str. Note that filepath detection can only be performed if
//       forceToken is "git".
//
// 'srcResolveFrom' (optional) If non-empty, then relative filepaths are
//       resolved relative to this location, overriding the 'pwd' param for
//       that purpose. Must be a rooted value in order for path resolution to
//       succeed; will emit an error, otherwise.
//
// See also: GitDetector
//
func (d *GitCtxDetector) CtxDetect(src, pwd, forceToken, _, srcResolveFrom string) (string, bool, error) {

	// If the 'git::' force token was specified, then our work here is more
	// than just "best effort"; we must complain if we were not able to detect
	// how to parse a src string that was explicitly flagged for us to
	// handle. This flag tells our response handling how to adapt.
	//
	mustField := "git" == forceToken // Are we required to field this request, else bust?

	if len(src) == 0 {
		if mustField {
			return "", false, fmt.Errorf("forced 'git::' handling: src string must be non-empty")
		}
		return "", false, nil
	}

	if len(forceToken) > 0 {
		rslt, ok, err := detectGitForceFilepath(src, pwd, forceToken, srcResolveFrom)
		if err != nil {
			return "", false, err
		}
		if ok {
			return rslt, true, nil
		}
	}

	// The remainder of our detection can be done by GitDetector.Detect, as it
	// does not require the additional contextual information available
	// here. We'll delegate the processing to it in order to avoid duplicating
	// the logic here (even if doing so might allow us to provide more precise
	// error messages), but we'll intercept it's "not detected" responses and
	// change them to errors, if needed.
	//
	rslt, ok, err := (&GitDetector{}).Detect(src, pwd)
	if err != nil {
		return "", false, err
	}
	if ok {
		return rslt, true, nil
	}
	if mustField {
		return "", false, fmt.Errorf("forced 'git::' handling: was unable to interpret src string: %s", src)
	}
	return "", false, nil
}

// detectGitForceFilepath() allows the 'git::' force token to be used on a
// filepath to a Git repository. Both absolute and relative filepaths are
// supported.
//
// When in-effect, the returned string will contain:
//
//     "git::file://<abspath>"
// OR:
//     "git::file://<abspath>?<query_params>"
//
// where <abspath> is the provided src param expanded to an absolute filepath
// (if it wasn't already). If the provided src param is a relative filepath,
// then the expanded absolute file path is based on the 'pwd' param, unless
// 'srcResolveFrom' is non-empty; then the expanded path is based on
// 'srcResolveFrom'.
//
// For consistency with Client, this function will not expand filepaths based
// on the process's current working directory. See comment in client.go
//
// Note that detectGitForceFilepath() only handles filepaths that have been
// explicitly forced (via the 'git::' force token) for Git processing.
//
//
// Q: Why isn't this functionality in our GitDetector.Detect() implementation?
//
// A: It is only safe to do when the 'git::' force token is provided, and the
//    Detector interface does not currently support a mechanism for the caller
//    to supply that context. Consequently, the GitDetector implementation
//    cannot assume that a filepath is intended for Git processing (that is,
//    as representing a Git repository). Enter the CtxDetector interface, and
//    this function.
//
//
// Q: Why is the returned value embedded in a 'file://' URI?
//
// A. When specifying the 'git::' force token on a filepath, the intention is
//    to indicate that it is a path to a Git repository that can be cloned,
//    etc. Though Git will accept both relative and absolute file paths for
//    this purpose, we unlock more functionality by using a 'file://' URI. In
//    particular, that allows our GitGetter to use any provided query params
//    to specify a specific tag or commit hash, for example.
//
//
// Q: Why is a relative path expanded to an absolute path in the returned
//    'file://' URI?
//
// A: Relative paths are technically not "legal" in RFC 1738- and RFC 8089-
//    compliant 'file://' URIs, and (more importantly for our purposes here)
//    they are not accepted by Git. When generating a 'file://' URI as we're
//    doing here, using the absolute path is the only useful thing to do.
//
//
// Q: Why support this functionality at all? Why not just require that a
//    source location use an absolute path in a 'file://' URI explicitly if
//    that's what is needed?
//
// A: The primary reason is to allow support for relative filepaths to Git
//    repos. There are use cases in which the absolute path cannot be known in
//    advance, but a relative path to a Git repo is known. For example, when a
//    Terraform project (or any Git-based project) uses Git submodules, it
//    will know the relative location of the Git submodule repos, but cannot
//    know the absolute path in advance because that will vary based on where
//    the "superproject" repo is cloned. Nevertheless, those relative paths
//    should be usable as clonable Git repos, and this mechanism allows for
//    that.
//
//    Support for filepaths that are already absolute is provided mainly for
//    symmetry. It would be surprising if the feature worked for relative
//    filepaths and not for absolute filepaths.
//
//
// Param Notes
// -----------
// 'src' (required) Filepath detection only works if the 'src' param start
//       with './', '../', '/', or their Windows equivalents. That is, the
//       value must be clearly identifiable as a filepath reference without
//       requiring that we actually touch the filesystem itself.
//
//       Values such as 'foo', or 'foo/bar', which may or may not be
//       filepaths, are ignored (this function will not have an effect on
//       them).
//
// 'pwd' (optional) For compatibility with the Detector interface (which has a
//       'pwd' param but not a 'srcResolveFrom' param), this is used as the
//       default path against which to resolve a relative filepath in
//       'src'. May be overridden by 'srcResolveFrom'.
//
//       When used for filepath resolution, it is /required/ to be a rooted
//       filepath (that is, it must be absolute). This is a restriction of
//       this CtxDetector implementation that is more strict than what is
//       required by CtxDetector.CtxDetect.
//
// 'force' (optional) Indicates whether or not the 'git::' forcing token was
//       specified. If empty, or if non-empty, but contains a value other than
//       "git", then this function is effectively a NOOP.
//
// 'srcResolveFrom' (optional) If non-empty, then overrides 'pwd' as the
//       location from which a relative filepath value in 'src' will be
//       resolved. This is useful when the relative filepath in 'src' is
//       relative to a different location than the 'pwd' of the process
//       (example: Terraform modules in subdirectories -- the filepaths need
//       to be interpretted relative to the directory in which the module
//       lives).
//
//       As with 'pwd', when used for filepath resolution, 'srcResolveFrom' is
//       /required/ to be a rooted filepath. This is a restriction of this
//       CtxDetector implementation that is more strict than what is required
//       by CtxDetector.CtxDetect.
//
// Return value:
// -------------
// The filepath in our return string is Clean, in the filepath.Clean()
// sense. That means:
//
//     1. Even when 'src' was provided with an absolute filepath value, the
//        emitted cleaned value may be different.
//
//     2. Anything that looks like a go-getter "subdir" value
//        ('//some/subdir') in 'src' will not be distinguishable as such in
//        our result string (because '//' would get cleaned to just '/'). This
//        should not be a problem in practice, though as the CtxDetect()
//        dispatch function removes such values from 'src' prior to invoking
//        the 'CtxDetect()' method of the CtxDetector(s); this function should
//        only see such values in 'src' when running under our unit tests.
//
func detectGitForceFilepath(src, pwd, force, srcResolveFrom string) (string, bool, error) {

	// The full force key token is 'git::', but the Detect() dispatcher
	// function provides our CtxDetect() method with the parsed value (without
	// the trailing colons).
	//
	if force != "git" {
		return "", false, nil
	}

	if len(src) == 0 {
		// cannot be a filepath; not a value for this function
		return "", false, nil
	}

	var srcResolvedAbs string

	if filepath.IsAbs(src) {
		srcResolvedAbs = src
	} else {

		// For our purposes, a relative filepath MUST begin with './' or
		// '../', or the Windows equivalent.
		//
		if !isLocalSourceAddr(src) {
			// src is neither an absolute nor relative filepath (at least not
			// obviously so), so we'll treat it as "not for us".
			return "", false, nil
		}

		// Recall that the result of filepath.Join() is Cleaned
		if len(srcResolveFrom) > 0 {

			if !filepath.IsAbs(srcResolveFrom) {
				return "", false, fmt.Errorf("unable to resolve 'git::' forced filepath (%s)"+
					"; provided srcResolveFrom filepath (%s) is not rooted", src, srcResolveFrom)
			}

			srcResolvedAbs = filepath.Join(srcResolveFrom, src)

		} else if len(pwd) > 0 {

			if !filepath.IsAbs(pwd) {
				return "", false, fmt.Errorf("unable to resolve 'git::' forced filepath (%s)"+
					"; provided pwd filepath (%s) is not rooted", pwd, srcResolveFrom)
			}

			srcResolvedAbs = filepath.Join(pwd, src)
		} else {
			// There's no way to resolve a more complete filepath for 'src'
			// other than to do so relative to the current working directory
			// of the process (which go-getter won't do, by design; see
			// comment in Client).

			return "", false, fmt.Errorf("unable to resolve 'git::' force filepath (%s)"+
				"; neither 'pwd' nor 'srcResolveFrom' param was provided", src)
		}
	}

	srcResolvedAbs = filepath.Clean(srcResolvedAbs)

	// To make the filepath usable (hopefully) in a 'file://' URI, we may need
	// to flip Windows-style '\' to URI-style '/'.
	//
	// Note that filepath.ToSlash is effectively a NOOP on platforms where '/'
	// is the os.Separator; when running on Unix-like platforms, this WILL NOT
	// hose your Unix filepaths that just happen to have backslash characters
	// in them.
	//
	srcResolvedAbs = filepath.ToSlash(srcResolvedAbs)

	if !strings.HasPrefix(srcResolvedAbs, "/") {

		// An absolute file path on Unix will start with a '/', but that is
		// not true for all OS's. RFC 8089 makes the authority component
		// (including the '//') optional in a 'file:' URL, but git (at least
		// as of version 2.28.0) only recognizes the 'file://' form. In fact,
		// the git-clone(1) manpage is explicit that it wants the syntax to
		// be:
		//
		//     file:///path/to/repo.git/
		//
		// Some notes on the relevant RFCs:
		//
		// RFC 1738 (section 3.10, "FILES") documents a <host> and <path>
		// portion being separated by a '/' character:
		//
		//     file://<host>/<path>
		//
		// RFC 2396 (Appendix G.2, "Modifications from both RFC 1738 and RFC
		// 1808") refines the above by declaring that the '/' is actually part
		// of the path. It is still required to separate the "authority
		// portion" of the URI from the path portion, but is not a separate
		// component of the URI syntax.
		//
		// RFC 3986 (Section 3.2, "Authority") states that the authority
		// component of a URI "is terminated by the next slash ("/"), question
		// mark ("?"), or number sign ("#") character, or by the end of the
		// URI." However, for 'file:' URIs, only those terminated by a '/'
		// characters are supported by Git (as noted above).
		//
		// RFC 8089 (Appendix A, "Differences from Previous Specifications")
		// references the RFC 1738 form including the required '/' after the
		// <host>/authority component.
		//
		// Because it is the most compatible approach across the only
		// partially compatible RFC recommendations, and (more importantly)
		// because it is what Git requires for 'file:' URIs, we require that
		// our absolute path value start with a '/' character.
		//
		srcResolvedAbs = "/" + srcResolvedAbs
	}

	// Our 'srcResolvedAbs' value may have URI query parameters (e.g.,
	// "ref=v1.2.3"), and the path elements may have characters that would
	// need to be escaped in a proper URI. We'll leverage url.Parse() to deal
	// with all of that, and then down below the stringified version of it
	// will be properly encoded.
	//
	u, err := url.Parse("file://" + srcResolvedAbs)
	if err != nil {
		return "", false, fmt.Errorf("error parsing 'git::' force filepath (%s) to URL: %s", srcResolvedAbs, err)
	}

	rtn := fmt.Sprintf("%s::%s", force, u.String())

	return rtn, true, nil
}

// Borrowed from terraform/internal/initwd/getter.go
// (modified here to accept "." and "..", too, if exact, full matches)
var localSourcePrefixes = []string{
	`./`,
	`../`,
	`.\\`,
	`..\\`,
}
var localExactMatches = []string{
	`.`,
	`..`,
}

func isLocalSourceAddr(addr string) bool {
	for _, value := range localExactMatches {
		if value == addr {
			return true
		}
	}
	for _, prefix := range localSourcePrefixes {
		if strings.HasPrefix(addr, prefix) {
			return true
		}
	}
	return false
}
