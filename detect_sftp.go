package getter

// SftpDetector implements Detector to detect Sftp URLs and turn
// them into URLs that the Sftp getter can understand.
type SftpDetector struct{}

func (d *SftpDetector) Detect(src, _ string) (string, bool, error) {
	return "", false, nil
}
