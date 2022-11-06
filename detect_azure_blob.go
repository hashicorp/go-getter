package getter

import (
	"fmt"
	"net/url"
	"strings"
)

type AzureBlobDetector struct{}

var azureStorageSuffixes = []string{
	"blob.core.windows.net",
	"blob.core.usgovcloudapi.net",
	"blob.core.chinacloudapi.cn",
}

func (d *AzureBlobDetector) Detect(src, pwd string) (string, bool, error) {
	if len(src) == 0 {
		return "", false, nil
	}

	for _, s := range azureStorageSuffixes {
		if strings.Contains(src, s) {
			return d.detectURL(src)
		}
	}

	return "", false, nil
}

func (d *AzureBlobDetector) detectURL(src string) (string, bool, error) {
	u, err := url.Parse(src)
	if err != nil {
		return "", false, err
	}

	parts := strings.Split(u.Path, "/")
	if len(parts) < 2 {
		return "", false, fmt.Errorf("path to blob must not be empty")
	}

	u.Scheme = "https"

	return fmt.Sprintf("azureblob::%s", u.String()), true, nil
}
