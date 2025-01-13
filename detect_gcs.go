// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package getter

import (
	"fmt"
	"strings"

	urlhelper "github.com/hashicorp/go-getter/helper/url"
)

// GCSDetector implements Detector to detect GCS URLs and turn
// them into URLs that the GCSGetter can understand.
type GCSDetector struct{}

func (d *GCSDetector) Detect(src, _ string) (string, bool, error) {
	if len(src) == 0 {
		return "", false, nil
	}

	if !strings.HasPrefix(src, "http://") && !strings.HasPrefix(src, "https://") {
		src = "https://" + src
	}

	parsedURL, err := urlhelper.Parse(src)
	if err != nil {
		return "", false, fmt.Errorf("error parsing GCS URL")
	}

	if strings.HasSuffix(parsedURL.Host, ".googleapis.com") {
		src = strings.TrimPrefix(src, "https://")
		src = strings.TrimPrefix(src, "http://")
		return d.detectHTTP(src)
	}

	return "", false, nil
}

func (d *GCSDetector) detectHTTP(src string) (string, bool, error) {

	parts := strings.Split(src, "/")
	if len(parts) < 5 {
		return "", false, fmt.Errorf(
			"URL is not a valid GCS URL")
	}
	version := parts[2]
	bucket := parts[3]
	object := strings.Join(parts[4:], "/")

	url, err := urlhelper.Parse(fmt.Sprintf("https://www.googleapis.com/storage/%s/%s/%s",
		version, bucket, object))
	if err != nil {
		return "", false, fmt.Errorf("error parsing GCS URL: %s", err)
	}

	return "gcs::" + url.String(), true, nil
}
