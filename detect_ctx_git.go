package getter

// GitCtxDetector implements CtxDetector to detect Git SSH URLs such as
// git@host.com:dir1/dir2 and converts them to proper URLs.
//
type GitCtxDetector struct{}

func (d *GitCtxDetector) CtxDetect(src, pwd, _, _, _ string) (string, bool, error) {

	// Currently not taking advantage of the extra contextual data available
	// to us. For now, we just delegate to GitDetector.Detect.
	//
	return (&GitDetector{}).Detect(src, pwd)
}
