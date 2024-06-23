package getter

// BitBucketCtxDetector implements CtxDetector to detect BitBucket URLs and
// turn them into URLs that the Git or Hg Getter can understand.
//
type BitBucketCtxDetector struct{}

func (d *BitBucketCtxDetector) CtxDetect(src, pwd, _, _, _ string) (string, bool, error) {

	// Currently not taking advantage of the extra contextual data available
	// to us. For now, we just delegate to BitBucketDetector.Detect.
	//
	return (&BitBucketDetector{}).Detect(src, pwd)
}
