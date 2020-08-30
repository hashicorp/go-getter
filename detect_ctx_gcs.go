package getter

// GCSCtxDetector implements CtxDetector to detect GCS URLs and turn them into
// URLs that the GCSGetter can understand.
//
type GCSCtxDetector struct{}

func (d *GCSCtxDetector) CtxDetect(src, pwd, _, _, _ string) (string, bool, error) {

	// Currently not taking advantage of the extra contextual data available
	// to us. For now, we just delegate to GCSDetector.Detect.
	//
	return (&GCSDetector{}).Detect(src, pwd)
}
