package getter

import (
	"testing"

	"path/filepath"
)

func TestCtxGitDetector(t *testing.T) {
	cases := []struct {
		Input  string
		Output string
	}{
		{
			"git@github.com:hashicorp/foo.git",
			"git::ssh://git@github.com/hashicorp/foo.git",
		},
		{
			"git@github.com:org/project.git?ref=test-branch",
			"git::ssh://git@github.com/org/project.git?ref=test-branch",
		},
		{
			"git@github.com:hashicorp/foo.git//bar",
			"git::ssh://git@github.com/hashicorp/foo.git//bar",
		},
		{
			"git@github.com:hashicorp/foo.git?foo=bar",
			"git::ssh://git@github.com/hashicorp/foo.git?foo=bar",
		},
		{
			"git@github.xyz.com:org/project.git",
			"git::ssh://git@github.xyz.com/org/project.git",
		},
		{
			"git@github.xyz.com:org/project.git?ref=test-branch",
			"git::ssh://git@github.xyz.com/org/project.git?ref=test-branch",
		},
		{
			"git@github.xyz.com:org/project.git//module/a",
			"git::ssh://git@github.xyz.com/org/project.git//module/a",
		},
		{
			"git@github.xyz.com:org/project.git//module/a?ref=test-branch",
			"git::ssh://git@github.xyz.com/org/project.git//module/a?ref=test-branch",
		},
		{
			// Already in the canonical form, so no rewriting required
			// When the ssh: protocol is used explicitly, we recognize it as
			// URL form rather than SCP-like form, so the part after the colon
			// is a port number, not part of the path.
			"git::ssh://git@git.example.com:2222/hashicorp/foo.git",
			"git::ssh://git@git.example.com:2222/hashicorp/foo.git",
		},
	}

	pwd := "/pwd"
	f := new(GitCtxDetector)
	ds := []CtxDetector{f}
	for _, tc := range cases {
		t.Run(tc.Input, func(t *testing.T) {
			output, err := CtxDetect(tc.Input, pwd, "", ds)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if output != tc.Output {
				t.Errorf("wrong result\ninput: %s\ngot:   %s\nwant:  %s", tc.Input, output, tc.Output)
			}
		})
	}
}

// The empty string value for the 'pwd' and/or 'srcResolveFrom' params is
// interpretted as "no value provided". When neither is provided, there will
// be no filepath against which to resolve the 'src' param.
//
// If 'src' is already an unambiguous absolute filepath, then Git filepath
// detection will still work (because neither 'pwd' nor 'srcResolveFrom' was
// needed to do so).
//
func Test_detectGitForceFilepath_directly_pos_pwd_and_srcResolveFrom_both_emtpy(t *testing.T) {

	posCases := []struct {
		Input  string
		Output string
	}{
		{
			"/somedir",
			"git::file:///somedir",
		},
		{
			"/somedir/two",
			"git::file:///somedir/two",
		},
		{
			"/somedir/two with space",
			"git::file:///somedir/two%20with%20space",
		},
		{
			"/somedir/two/three",
			"git::file:///somedir/two/three",
		},
		{
			"/somedir/two with space/three",
			"git::file:///somedir/two%20with%20space/three",
		},

		{ // subdir-looking thing IS NOT retained here; is "cleaned away" (good)
			"/somedir/two//three",
			"git::file:///somedir/two/three",
		},

		{
			"/somedir/two/three?ref=v4.5.6",
			"git::file:///somedir/two/three?ref=v4.5.6",
		},

		{
			"/somedir/two with space/three?ref=v4.5.6",
			"git::file:///somedir/two%20with%20space/three?ref=v4.5.6",
		},
	}

	pwd := ""
	srcResolveFrom := ""
	force := "git" // parsed form of magic 'git::' force token

	for _, tc := range posCases {
		t.Run(tc.Input, func(t *testing.T) {
			output, ok, err := detectGitForceFilepath(tc.Input, pwd, force, srcResolveFrom)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if !ok {
				t.Fatalf("unexpected !ok")
			}

			if output != tc.Output {
				t.Errorf("wrong result\ninput: %s\ngot:   %s\nwant:  %s", tc.Input, output, tc.Output)
			}
		})
	}
}

// The empty string value for the 'pwd' and/or 'srcResolveFrom' params is
// interpretted as "no value provided". When neither is provided, there will
// be no filepath against which to resolve the 'src' param.
//
// Regardless of whether the whether or not 'pwd' and/or 'srcResolveFrom'
// params are provided, if the 'src' param does not look unambiguously like a
// filepath, then detectGitForceFilepath(...) should not process the input
// string.
//
// For these negative cases, anything that looks like a relative filepath will
// not be processable. Unlike cases in which detectGitForceFilepath() ignores
// values that are not intended for it, in this case the 'git::' force token
// coupled with the relative filepath means the item is definitely intended
// for the function. But it simply cannot be processed because no rooted path
// value was provided against which to resolve 'src'. The function therefore
// emits an error directly.
//
func Test_detectGitForceFilepath_directly_neg_pwd_and_srcResolveFrom_both_emtpy(t *testing.T) {

	// These are cases that should be ignored because the provided string does
	// not look like a filepath. These will not be processed, but no error
	// should be emitted for them.
	//
	// See also negErrorCases below.
	//
	negIgnoreCases := []struct {
		Input  string
		Output string
	}{
		{
			"",
			"",
		},
		{
			"foo",
			"",
		},
		{
			"foo/bar",
			"",
		},
		{
			"foo/bar/",
			"",
		},
		{
			"git@github.com:hashicorp/foo.git",
			"",
		},
		{
			"git@github.com:org/project.git?ref=test-branch",
			"",
		},
		{
			"git@github.com:hashicorp/foo.git//bar",
			"",
		},
		{
			"git@github.com:hashicorp/foo.git?foo=bar",
			"",
		},
		{
			"git@github.xyz.com:org/project.git",
			"",
		},
		{
			"git@github.xyz.com:org/project.git?ref=test-branch",
			"",
		},
		{
			"git@github.xyz.com:org/project.git//module/a",
			"",
		},
		{
			"git@github.xyz.com:org/project.git//module/a?ref=test-branch",
			"",
		},
		{
			"ssh://git@git.example.com:2222/hashicorp/foo.git",
			"",
		},
	}

	// These are cases that should cause an error to be emitted directly by
	// the function because the provided string looks like a relative
	// filepath, but no rooted 'pwd' or 'srcResolveFrom' value was provided
	// against which to resolve them.
	//
	// See also negIgnoreCases above.
	//
	negErrorCases := []struct {
		Input  string
		Output string
	}{
		{
			".",
			"",
		},
		{
			"./",
			"",
		},
		{
			"./.",
			"",
		},
		{
			"././",
			"",
		},
		{
			"././.",
			"",
		},
		{
			"./././",
			"",
		},
		{
			"..",
			"",
		},
		{
			"../",
			"",
		},
		{
			"../..",
			"",
		},
		{
			"../../",
			"",
		},
		{
			"../../..",
			"",
		},
		{
			"../../../",
			"",
		},

		{
			"./somedir",
			"",
		},
		{
			"./somedir/two",
			"",
		},
		{
			"./somedir/two with space",
			"",
		},
		{
			"./somedir/two/three",
			"",
		},
		{
			"./somedir/two//three",
			"",
		},
		{
			"./somedir/two/three?ref=v4.5.6",
			"",
		},

		{
			"../some-parent-dir",
			"",
		},
		{
			"../some-parent-dir/two",
			"",
		},

		{
			"../some-parent-dir/two/three",
			"",
		},
		{
			"../some-parent-dir/two//three",
			"",
		},

		{
			"../../some-grandparent-dir",
			"",
		},
		{
			"../../some-grandparent-dir?ref=v1.2.3",
			"",
		},

		{
			"../../../some-great-grandparent-dir",
			"",
		},
		{
			"../../../some-great-grandparent-dir?ref=v1.2.3",
			"",
		},

		{
			"../../../../some-2x-great-grandparent-dir",
			"",
		},
		{
			"../../../../some-2x-great-grandparent-dir?ref=v1.2.3",
			"",
		},

		{
			"../../../../../some-3x-great-grandparent-dir",
			"",
		},
		{
			"../../../../../some-3x-great-grandparent-dir?ref=v1.2.3",
			"",
		},
	}

	pwd := ""
	srcResolveFrom := ""
	force := "git"

	for _, tc := range negIgnoreCases {
		t.Run(tc.Input, func(t *testing.T) {
			output, ok, err := detectGitForceFilepath(tc.Input, pwd, force, srcResolveFrom)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if ok {
				t.Errorf("unexpected ok on input: %s", tc.Input)
			}
			if output != "" {
				t.Errorf("unexpected non-empty output string; input: %s; output: %s", tc.Input, output)
			}
		})
	}

	for _, tc := range negErrorCases {
		t.Run(tc.Input, func(t *testing.T) {
			output, ok, err := detectGitForceFilepath(tc.Input, pwd, force, srcResolveFrom)
			if err == nil {
				t.Errorf("unexpected non-error; input: %s; output: %s; ok: %t", tc.Input, output, ok)
			}
		})
	}
}

// When the 'srcResolveFrom' param is empty, then filepath resolution should
// be performed against the 'pwd' param (unless it, too, is empty). In this
// function, srcResolveFrom is always empty, and pwd is always non-empty and
// rooted.
//
func Test_detectGitForceFilepath_directly_pos_empty_srcResolveFrom(t *testing.T) {

	// Test case input values represent the parsed inputs, similar to what
	// would be presented by the 'CtxDetect' dispatch function.
	//
	// CAREFUL: Recall that the 'CtxDetect' dispatch function does subdir
	//          parsing before attempting to invoke the 'CtxDetect' methods of
	//          the CtxDetectors. The subdir value is passed as a separate
	//          'subDir' param to the 'CtxDetect' methods, and in the
	//          particular case of our GitCtxDetector that subDir param is not
	//          provided to the 'detectGitForceFilepath' function. So in
	//          practice 'detectGitForceFilepath' should never see a ('//')
	//          subdir in its 'src' param. Consequently, our positive tests
	//          below that fake one up, anyway, result in "cleaned" paths (in
	//          the filepath.Clean() sense). The function is behaving
	//          correctly by /not/ doing explicit subdir handling, and by
	//          "cleaning away" filepaths that happen to have subdir-looking
	//          sections in them.

	pwd := "/some/caller-provided/abs/path"
	pwdMinusOne := filepath.Dir(pwd)            // dirname
	pwdMinusTwo := filepath.Dir(pwdMinusOne)    // likewise
	pwdMinusThree := filepath.Dir(pwdMinusTwo)  // likewise
	pwdMinusFour := filepath.Dir(pwdMinusThree) // likewise; is root ('/')

	// Sanity check our "directory math"
	if pwdMinusThree == "/" {
		t.Fatalf("pwdMinusThree is unexpectedly the top-level root ('/') directory")
	}
	if pwdMinusFour != "/" {
		t.Fatalf("pwdMinusFour is not the top-level root ('/') directory; is: %s", pwdMinusFour)
	}

	// The 'CtxDetect' protocol, ultimately implemented here in our
	// 'detectGitForceFilepath' function, is that 'src' is resolved relative
	// to 'pwd', unless the provided 'srcResolveFrom' param is non-empty; then
	// the 'src' is resolved relative to 'srcResolveFrom'.
	//
	// Because our 'pwd' param is non-empty in all of these cases, these tests
	// demonstrate the above protocol working.
	//
	posCases := []struct {
		Input  string
		Output string
	}{
		{
			".",
			"git::file://" + pwd,
		},
		{
			"./",
			"git::file://" + pwd,
		},
		{
			"./.",
			"git::file://" + pwd,
		},
		{
			"././",
			"git::file://" + pwd,
		},
		{
			"././.",
			"git::file://" + pwd,
		},
		{
			"./././",
			"git::file://" + pwd,
		},

		{
			"..",
			"git::file://" + pwdMinusOne,
		},
		{
			"../",
			"git::file://" + pwdMinusOne,
		},
		{
			"../.",
			"git::file://" + pwdMinusOne,
		},
		{
			".././",
			"git::file://" + pwdMinusOne,
		},

		{
			"../../",
			"git::file://" + pwdMinusTwo,
		},

		{
			"../../../",
			"git::file://" + pwdMinusThree,
		},

		{
			"../../../../",
			"git::file://" + pwdMinusFour,
		},

		{ // Same output as previous, since we only have 2x
			// great-grandparents before we hit the top-level root directory
			"../../../../../",
			"git::file://" + pwdMinusFour,
		},

		{
			"/somedir",
			"git::file:///somedir",
		},
		{
			"./somedir",
			"git::file://" + pwd + "/somedir",
		},

		{
			"/somedir/two",
			"git::file:///somedir/two",
		},
		{
			"./somedir/two",
			"git::file://" + pwd + "/somedir/two",
		},

		{
			"/somedir/two/three",
			"git::file:///somedir/two/three",
		},
		{
			"./somedir/two/three",
			"git::file://" + pwd + "/somedir/two/three",
		},

		{ // subdir-looking thing IS NOT retained here; is "cleaned away" (good)
			"/somedir/two//three",
			"git::file:///somedir/two/three",
		},
		{
			"./somedir/two//three",
			"git::file://" + pwd + "/somedir/two/three", // no subdir here (good)
		},

		{
			"/somedir/two/three?ref=v4.5.6",
			"git::file:///somedir/two/three?ref=v4.5.6",
		},
		{
			"./somedir/two/three?ref=v4.5.6",
			"git::file://" + pwd + "/somedir/two/three?ref=v4.5.6",
		},

		{
			"../some-parent-dir",
			"git::file://" + pwdMinusOne + "/some-parent-dir",
		},
		{
			"../some-parent-dir/two",
			"git::file://" + pwdMinusOne + "/some-parent-dir/two",
		},

		{
			"../some-parent-dir/two/three",
			"git::file://" + pwdMinusOne + "/some-parent-dir/two/three",
		},
		{
			"../some-parent-dir/two//three",
			"git::file://" + pwdMinusOne + "/some-parent-dir/two/three", // no subdir here (okay)
		},

		{
			"../../some-grandparent-dir",
			"git::file://" + pwdMinusTwo + "/some-grandparent-dir",
		},
		{
			"../../some-grandparent-dir?ref=v1.2.3",
			"git::file://" + pwdMinusTwo + "/some-grandparent-dir?ref=v1.2.3",
		},

		{
			"../../../some-great-grandparent-dir",
			"git::file://" + pwdMinusThree + "/some-great-grandparent-dir",
		},
		{
			"../../../some-great-grandparent-dir?ref=v1.2.3",
			"git::file://" + pwdMinusThree + "/some-great-grandparent-dir?ref=v1.2.3",
		},

		{
			"../../../../some-2x-great-grandparent-dir",
			"git::file://" + pwdMinusFour + "some-2x-great-grandparent-dir",
			// .............................^
			// CAREFUL: No explicit leading '/' here because 'pwdMinusFour' is
			//          our top-level root directory; is just '/'; we do not
			//          expect 'file:////some-2x-...'
		},
		{
			"../../../../some-2x-great-grandparent-dir?ref=v1.2.3",
			"git::file://" + pwdMinusFour + "some-2x-great-grandparent-dir?ref=v1.2.3",
		},

		{ // Same output as previous, since we only have 2x
			// great-grandparents before we hit the top-level root directory
			"../../../../../some-3x-great-grandparent-dir",
			"git::file://" + pwdMinusFour + "some-3x-great-grandparent-dir",
		},
		{
			"../../../../../some-3x-great-grandparent-dir?ref=v1.2.3",
			"git::file://" + pwdMinusFour + "some-3x-great-grandparent-dir?ref=v1.2.3",
		},
	}

	force := "git" // parsed form of magic 'git::' force token

	srcAbsResolveFrom := ""

	for _, tc := range posCases {
		t.Run(tc.Input, func(t *testing.T) {
			output, ok, err := detectGitForceFilepath(tc.Input, pwd, force, srcAbsResolveFrom)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if !ok {
				t.Fatalf("unexpected !ok")
			}

			if output != tc.Output {
				t.Errorf("wrong result\ninput: %s\ngot:   %s\nwant:  %s", tc.Input, output, tc.Output)
			}
		})
	}
}

func Test_detectGitForceFilepath_directly_pos_nonempty_srcResolveFrom(t *testing.T) {

	// Test case input values represent the parsed inputs, similar to what
	// would be presented by the 'CtxDetect' dispatch function.
	//
	// CAREFUL: Recall that the 'CtxDetect' dispatch function does subdir
	//          parsing before attempting to invoke the 'CtxDetect' methods of
	//          the CtxDetectors. The subdir value is passed as a separate
	//          'subDir' param to the 'CtxDetect' methods, and in the
	//          particular case of our GitCtxDetector that subDir param is not
	//          provided to the 'detectGitForceFilepath' function. So in
	//          practice 'detectGitForceFilepath' should never see a ('//')
	//          subdir in its 'src' param. Consequently, our positive tests
	//          below that fake one up, anyway, result in "cleaned" paths (in
	//          the filepath.Clean() sense). The function is behaving
	//          correctly by /not/ doing explicit subdir handling, and by
	//          "cleaning away" filepaths that happen to have subdir-looking
	//          sections in them.
	//
	//          Compare with:
	//              Test_detectGitForceFilepath_indirectly()
	//          which leverages the higher-level processing of CtxDetect(), as
	//          well.

	srcAbsResolveFrom := "/some/caller-provided/abs/path"
	srcAbsResolveFromMinusOne := filepath.Dir(srcAbsResolveFrom)            // dirname
	srcAbsResolveFromMinusTwo := filepath.Dir(srcAbsResolveFromMinusOne)    // likewise
	srcAbsResolveFromMinusThree := filepath.Dir(srcAbsResolveFromMinusTwo)  // likewise
	srcAbsResolveFromMinusFour := filepath.Dir(srcAbsResolveFromMinusThree) // likewise; is root ('/')

	// Sanity check our "directory math"
	if srcAbsResolveFromMinusThree == "/" {
		t.Fatalf("srcAbsResolvFromMinusThree is unexpectedly the top-level root ('/') directory")
	}
	if srcAbsResolveFromMinusFour != "/" {
		t.Fatalf("srcAbsResolvFromMinusFour is not the top-level root ('/') directory; is: %s", srcAbsResolveFromMinusFour)
	}

	// Our test loop below runs one iteration, with a hard-coded, non-empty
	// 'pwd' value.
	//
	// Our non-empty srcAbsResolveFrom values for the 'srcResolveFrom' param
	// are provided in addition to the value provided for the 'pwd' param.
	//
	// The 'CtxDetect' protocol, ultimately implemented here in our
	// 'detectGitForceFilepath' function, is that 'src' is resolved relative
	// to 'pwd', unless the provided 'srcResolveFrom' param is non-empty; then
	// the 'src' is resolved relative to 'srcResolveFrom'.
	//
	// Because our 'pwd' param is non-empty in all of these cases, these tests
	// demonstrate the above protocol working.
	//
	posCases := []struct {
		Input  string
		Output string
	}{
		{
			"/somedir",
			"git::file:///somedir",
		},
		{
			"./somedir",
			"git::file://" + srcAbsResolveFrom + "/somedir",
		},

		{
			"/somedir/two",
			"git::file:///somedir/two",
		},
		{
			"./somedir/two",
			"git::file://" + srcAbsResolveFrom + "/somedir/two",
		},

		{
			"/somedir/two/three",
			"git::file:///somedir/two/three",
		},
		{
			"./somedir/two/three",
			"git::file://" + srcAbsResolveFrom + "/somedir/two/three",
		},

		{ // subdir-looking thing IS NOT retained here; is "cleaned away" (good)
			"/somedir/two//three",
			"git::file:///somedir/two/three",
		},
		{
			"./somedir/two//three",
			"git::file://" + srcAbsResolveFrom + "/somedir/two/three", // no subdir here (good)
		},

		{
			"/somedir/two/three?ref=v4.5.6",
			"git::file:///somedir/two/three?ref=v4.5.6",
		},
		{
			"./somedir/two/three?ref=v4.5.6",
			"git::file://" + srcAbsResolveFrom + "/somedir/two/three?ref=v4.5.6",
		},

		{
			"../some-parent-dir",
			"git::file://" + srcAbsResolveFromMinusOne + "/some-parent-dir",
		},
		{
			"../some-parent-dir/two",
			"git::file://" + srcAbsResolveFromMinusOne + "/some-parent-dir/two",
		},

		{
			"../some-parent-dir/two/three",
			"git::file://" + srcAbsResolveFromMinusOne + "/some-parent-dir/two/three",
		},
		{
			"../some-parent-dir/two//three",
			"git::file://" + srcAbsResolveFromMinusOne + "/some-parent-dir/two/three", // no subdir here (okay)
		},

		{
			"../../some-grandparent-dir",
			"git::file://" + srcAbsResolveFromMinusTwo + "/some-grandparent-dir",
		},
		{
			"../../some-grandparent-dir?ref=v1.2.3",
			"git::file://" + srcAbsResolveFromMinusTwo + "/some-grandparent-dir?ref=v1.2.3",
		},

		{
			"../../../some-great-grandparent-dir",
			"git::file://" + srcAbsResolveFromMinusThree + "/some-great-grandparent-dir",
		},
		{
			"../../../some-great-grandparent-dir?ref=v1.2.3",
			"git::file://" + srcAbsResolveFromMinusThree + "/some-great-grandparent-dir?ref=v1.2.3",
		},

		{
			"../../../../some-2x-great-grandparent-dir",
			"git::file://" + srcAbsResolveFromMinusFour + "some-2x-great-grandparent-dir",
			// ...........................................^
			// CAREFUL: No explicit leading '/' here because
			//          'srcAbsResolveFromMinusFour' is our top-level root
			//          directory; is just '/'; we do not expect 'file:////some-2x-...'
		},
		{
			"../../../../some-2x-great-grandparent-dir?ref=v1.2.3",
			"git::file://" + srcAbsResolveFromMinusFour + "some-2x-great-grandparent-dir?ref=v1.2.3",
		},

		{ // Same output as previous, since we only have 2x
			// great-grandparents before we hit the top-level root directory
			"../../../../../some-3x-great-grandparent-dir",
			"git::file://" + srcAbsResolveFromMinusFour + "some-3x-great-grandparent-dir",
		},
		{
			"../../../../../some-3x-great-grandparent-dir?ref=v1.2.3",
			"git::file://" + srcAbsResolveFromMinusFour + "some-3x-great-grandparent-dir?ref=v1.2.3",
		},
	}

	pwd := "/pwd-ignored"
	force := "git" // parsed form of magic 'git::' force token

	for _, tc := range posCases {
		t.Run(tc.Input, func(t *testing.T) {
			output, ok, err := detectGitForceFilepath(tc.Input, pwd, force, srcAbsResolveFrom)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if !ok {
				t.Fatalf("unexpected !ok")
			}

			if output != tc.Output {
				t.Errorf("wrong result\ninput: %s\ngot:   %s\nwant:  %s", tc.Input, output, tc.Output)
			}
		})
	}
}

func Test_detectGitForceFilepath_directly_neg(t *testing.T) {

	// Test case input values represent the parsed inputs, similar to what
	// would be presented by the CtxDetect() dispatch function.
	//
	// These are negative tests; the input represent values that
	// detectGitForceFilepath() should effectively ignore. So most outputs are
	// expected to be a !ok flag coupled with an empty result string.
	//
	// Recall that detectGitForceFilepath() considers as relative only those
	// paths that begin with './' or '../', or their Windows
	// equivalents. Paths in the form 'parent-dir/child-dir' are not
	// considered relative.

	negCases := []struct {
		Input  string
		Output string
	}{
		{
			"",
			"",
		},
		{
			"somedir",
			"",
		},
		{
			"somedir/two",
			"",
		},
		{
			"somedir/two//three",
			"",
		},
		{
			"somedir/two/three?ref=v4.5.6",
			"",
		},
	}

	pwd := "/pwd-ignored"
	srcResolveFrom := "/some/absolute/file/path/ignored"

	// We'll loop over our tests multiple times, with 'force' set to different
	// values. The tests in which the value is not 'git' should short circuit
	// quickly because the function should only have an effect when the force
	// token is in effect. For these iterations, the input string values
	// should be irrelevant.
	//
	// Those tests in which 'force' is set to "git" should also result in
	// negative results here, but exercise a different part of the
	// logic. These simulate the force token being in effect, and the function
	// interpretting the string input values to decide whether to
	// respond. Check the coverage report.
	//
	// The parsed form of the magic 'git::' force token is "git". All of the
	// other values (including "git::") should be ignored by
	// detectGitForceFilepath().
	//
	forceVals := []string{"", "blah:", "blah::", "git:", "git::", "git"}

	for _, force := range forceVals {

		for _, tc := range negCases {
			t.Run(tc.Input, func(t *testing.T) {
				output, ok, err := detectGitForceFilepath(tc.Input, pwd, force, srcResolveFrom)
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
				if ok {
					t.Errorf("unexpected ok on input: %s", tc.Input)
				}
				if output != "" {
					t.Errorf("unexpected non-empty output string; input: %s; output: %s", tc.Input, output)
				}
			})
		}
	}
}

func Test_detectGitForceFilepath_indirectly_pos(t *testing.T) {

	// Test case input values represent the raw input provided to the
	// CtxDetect() dispatch function, which will parse them prior to invoking
	// GitCtxDetector's CtxDetect() method (which in turn invokes
	// detectGitForceFilepath()).

	srcAbsResolveFrom := "/some/caller-provided/abs/path"
	srcAbsResolveFromMinusOne := filepath.Dir(srcAbsResolveFrom)            // dirname
	srcAbsResolveFromMinusTwo := filepath.Dir(srcAbsResolveFromMinusOne)    // likewise
	srcAbsResolveFromMinusThree := filepath.Dir(srcAbsResolveFromMinusTwo)  // likewise
	srcAbsResolveFromMinusFour := filepath.Dir(srcAbsResolveFromMinusThree) // likewise; is root ('/')

	// Sanity check our "directory math"
	if srcAbsResolveFromMinusThree == "/" {
		t.Fatalf("srcAbsResolvFromMinusThree is unexpectedly the top-level root ('/') directory")
	}
	if srcAbsResolveFromMinusFour != "/" {
		t.Fatalf("srcAbsResolvFromMinusFour is not the top-level root ('/') directory; is: %s", srcAbsResolveFromMinusFour)
	}

	posCases := []struct {
		Input  string
		Output string
	}{
		{
			"git::/somedir",
			"git::file:///somedir",
		},
		{
			"git::./somedir",
			"git::file://" + srcAbsResolveFrom + "/somedir",
		},

		{
			"git::/somedir/two",
			"git::file:///somedir/two",
		},
		{
			"git::./somedir/two",
			"git::file://" + srcAbsResolveFrom + "/somedir/two",
		},

		{
			"git::/somedir/two/three",
			"git::file:///somedir/two/three",
		},
		{
			"git::./somedir/two/three",
			"git::file://" + srcAbsResolveFrom + "/somedir/two/three",
		},

		{
			"git::/somedir/two//three",
			"git::file:///somedir/two//three", // subdir is preserved
		},
		{
			"git::./somedir/two//three",
			"git::file://" + srcAbsResolveFrom + "/somedir/two//three", // subdir is preserved
		},

		{
			"git::/somedir/two/three?ref=v4.5.6",
			"git::file:///somedir/two/three?ref=v4.5.6",
		},
		{
			"git::./somedir/two/three?ref=v4.5.6",
			"git::file://" + srcAbsResolveFrom + "/somedir/two/three?ref=v4.5.6",
		},

		{
			"git::/somedir/two//three?ref=v4.5.6",
			"git::file:///somedir/two//three?ref=v4.5.6", // subdir is preserved
		},
		{
			"git::./somedir/two//three?ref=v4.5.6",
			"git::file://" + srcAbsResolveFrom + "/somedir/two//three?ref=v4.5.6", // subdir is preserved
		},

		{
			"git::/somedir/two//three?ref=v4.5.6",
			"git::file:///somedir/two//three?ref=v4.5.6", // subdir is preserved
		},
		{
			"git::./somedir/two//three?ref=v4.5.6",
			"git::file://" + srcAbsResolveFrom + "/somedir/two//three?ref=v4.5.6", // subdir is preserved
		},

		{
			"git::../some-parent-dir",
			"git::file://" + srcAbsResolveFromMinusOne + "/some-parent-dir",
		},
		{
			"git::../some-parent-dir/two",
			"git::file://" + srcAbsResolveFromMinusOne + "/some-parent-dir/two",
		},
		{
			"git::../some-parent-dir/two/three",
			"git::file://" + srcAbsResolveFromMinusOne + "/some-parent-dir/two/three",
		},
		{
			"git::../some-parent-dir/two//three",
			"git::file://" + srcAbsResolveFromMinusOne + "/some-parent-dir/two//three", // subdir is preserved
		},
		{
			"git::../../some-grandparent-dir",
			"git::file://" + srcAbsResolveFromMinusTwo + "/some-grandparent-dir",
		},
		{
			"git::../../some-grandparent-dir?ref=v1.2.3",
			"git::file://" + srcAbsResolveFromMinusTwo + "/some-grandparent-dir?ref=v1.2.3",
		},
		{ // subdir is preserved on output
			"git::../../some-grandparent-dir/childdir//moduledir?ref=v1.2.3",
			"git::file://" + srcAbsResolveFromMinusTwo + "/some-grandparent-dir/childdir//moduledir?ref=v1.2.3",
		},
	}

	pwd := "/pwd-ignored"

	f := new(GitCtxDetector)
	ds := []CtxDetector{f}

	for _, tc := range posCases {
		t.Run(tc.Input, func(t *testing.T) {

			output, err := CtxDetect(tc.Input, pwd, srcAbsResolveFrom, ds)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if output != tc.Output {
				t.Errorf("wrong result\ninput: %s\ngot:   %s\nwant:  %s", tc.Input, output, tc.Output)
			}
		})
	}
}

func Test_detectGitForceFilepath_indirectly_neg(t *testing.T) {

	// Test case input values represent the path string input provided to the
	// CtxDetect() dispatch function, which will parse them prior to invoking
	// the CtxDetect() method on GitCtxDetector (which internally invokes
	// detectGitForceFilepath()).
	//
	// We loop over the list several times with different force key values to
	// trigger different paths through the code. All of these tests should
	// produce negative results either due to the absence of a legit 'git::'
	// force key, or because the path value does not represent a relative in
	// the form required.
	//
	// When the 'git::' force key is specified for these values, then the
	// GitCtxDetector will yield an error because the fields are clearly
	// flagged for processing by that detector, but it is unable to do so
	// because the input is not meaningful to it.

	negCases := []struct {
		Input  string
		Output string
	}{
		// FIXME: This first case (which is pathological, to be sure)
		//        incorrectly succeeds when joined with the 'git::'
		//        prefix. The reason is because the getForcedGetter function
		//        (in get.go) that is used to identify the force tokens is
		//        expecting at least one character after the '::' separator;
		//        at the time of writing it is using this regex:
		//
		//            `^([A-Za-z0-9]+)::(.+)$`
		//
		//        Since the length of the token on the left of the '::' is not
		//        size restricted, how to fix that without accidentally
		//        breaking somebody's use case?
		// {
		// 	"",
		// 	"",
		// },
		{
			"somedir",
			"",
		},
		{
			"somedir/two",
			"",
		},
		{
			"somedir/two/three",
			"",
		},
		{
			"somedir/two//three",
			"",
		},
		{
			"somedir/two/three?ref=v4.5.6",
			"",
		},
		{
			"somedir/two//three?ref=v4.5.6",
			"",
		},
	}

	pwd := "/pwd-ignored"
	srcResolveFrom := "/src-resolve-from-ignored"

	f := new(GitCtxDetector)
	ds := []CtxDetector{f}

	// Empty-string force will fail because no CtxDetector handles these
	// quasi-relative filepaths.
	//
	// The 'git::' force will fail because neither GitCtxDetector.CtxDetect()
	// nor detectGitForceFilepath() recognize the quasi-relative filepaths
	// (and neither do any other Detector implemenations, though they are not
	// invoked in this configuration).
	//
	forceVals := []string{"", "git::"}

	for _, force := range forceVals {
		for _, tc := range negCases {
			t.Run(tc.Input, func(t *testing.T) {

				output, err := CtxDetect(force+tc.Input, pwd, srcResolveFrom, ds)
				if err == nil {
					// When force is "", the error would be an "invalid source
					// string" error from the CtxDetect dispatch function.
					//
					// When force is "git::", the error would be something
					// from the GitCtxDetector's CtxDetect method complaining
					// that the string value was force tagged for it to
					// process, but it is not able to.
					//
					t.Fatalf("was expecting error, but call succeeded: output: %s (force is: %s)", output, force)
				}

				if output != "" {
					t.Errorf("unexpected non-empty output string; input: %s; output: %s (force is: %s)", tc.Input, output, force)
				}
			})
		}
	}

}
