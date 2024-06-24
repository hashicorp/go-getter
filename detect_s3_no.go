// +build go_getter_nos3

package getter

// S3Detector implements Detector to detect S3 URLs and turn
// them into URLs that the S3 getter can understand.
type S3Detector struct{}

func (d *S3Detector) Detect(src, _ string) (string, bool, error) {
	return "", false, nil
}
