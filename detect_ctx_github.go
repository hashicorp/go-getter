package getter

// GitHubCtxDetector implements CtxDetector to detect GitHub URLs and turn
// them into URLs that the Git Getter can understand.
//
type GitHubCtxDetector struct{}

func (d *GitHubCtxDetector) CtxDetect(src, pwd, _, _, _ string) (string, bool, error) {

	// Currently not taking advantage of the extra contextual data available
	// to us. For now, we just delegate to GitHubDetector.Detect.
	//
	return (&GitHubDetector{}).Detect(src, pwd)
}
