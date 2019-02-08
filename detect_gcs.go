package getter

import (
	"fmt"
	"net/url"
	"strings"
)

// GCSDetector implements Detector to detect GCS URLs and turn
// them into URLs that the GCSGetter can understand.
type GCSDetector struct{}

func (d *GCSDetector) Detect(src, _ string) (string, bool, error) {
	if len(src) == 0 {
		return "", false, nil
	}

	if strings.Contains(src, "googleapis.com/") {
		return d.detectHTTP(fmt.Sprintf("https://www.googleapis.com/%s",
			strings.SplitN(src, "googleapis.com/", 2)[1]))
	}

	return "", false, nil
}

func (d *GCSDetector) detectHTTP(src string) (string, bool, error) {
	url, err := url.Parse(src)
	if err != nil {
		return "", false, fmt.Errorf("error parsing GCS URL: %s", err)
	}

	pathParts := strings.SplitN(url.Path, "/", 5)
	if len(pathParts) != 5 {
		return "", false, fmt.Errorf("URL is not a valid GCS URL")
	}
	return "gcs::" + url.String(), true, nil
}
