package getter

// S3CtxDetector implements CtxDetector to detect S3 URLs and turn them into
// URLs that the S3 getter can understand.
//
type S3CtxDetector struct{}

func (d *S3CtxDetector) CtxDetect(src, pwd, _, _, _ string) (string, bool, error) {

	// Currently not taking advantage of the extra contextual data available
	// to us. For now, we just delegate to S3Detector.Detect.
	//
	return (&S3Detector{}).Detect(src, pwd)
}
