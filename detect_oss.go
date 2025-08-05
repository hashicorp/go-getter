package getter

import (
	"fmt"
	"net/url"
	"strings"
)

// OSSDetector implements Detector to detect OSS URLs and turn
// them into URLs that the OSSGetter can understand.
type OSSDetector struct{}

func (d *OSSDetector) Detect(src, _ string) (string, bool, error) {
	if len(src) == 0 {
		return "", false, nil
	}

	if strings.Contains(src, ".aliyuncs.com/") {
		return d.detectHTTP(src)
	}

	return "", false, nil
}

func (d *OSSDetector) detectHTTP(src string) (string, bool, error) {
	parts := strings.Split(src, "/")
	if len(parts) < 2 {
		return "", false, fmt.Errorf(
			"URL is not a valid OSS URL")
	}

	urlStr := fmt.Sprintf("https://%s", strings.Join(parts, "/"))
	url, err := url.Parse(urlStr)
	if err != nil {
		return "", true, fmt.Errorf("error parsing OSS URL: %s", err)
	}

	return "oss::" + url.String(), true, nil
}
