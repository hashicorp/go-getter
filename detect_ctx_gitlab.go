package getter

// GitLabCtxDetector implements CtxDetector to detect GitLab URLs and turn
// them into URLs that the Git Getter can understand.
//
type GitLabCtxDetector struct{}

func (d *GitLabCtxDetector) CtxDetect(src, pwd, _, _, _ string) (string, bool, error) {

	// Currently not taking advantage of the extra contextual data available
	// to us. For now, we just delegate to GitLabDetector.Detect.
	//
	return (&GitLabDetector{}).Detect(src, pwd)
}
