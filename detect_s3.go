package getter

import (
	"fmt"
	"net/url"
	"strings"
)

// S3Detector implements Detector to detect S3 URLs and turn
// them into URLs that the S3 getter can understand.
type S3Detector struct{}

var amazonAWSHosts = []string{
	"amazonaws.com",
	"amazonaws.com.cn", // China region's hostname
}

func (d *S3Detector) Detect(src, _ string) (string, bool, error) {
	if len(src) == 0 {
		return "", false, nil
	}

	for _, hstnm := range amazonAWSHosts {
		if strings.Contains(src, fmt.Sprintf(".%s/", hstnm)) {
			return d.detectHTTP(
				d.removeHostnameFromPath(src, hstnm),
				hstnm)
		}
	}

	return "", false, nil
}

func (d S3Detector) removeHostnameFromPath(path, hstnm string) string {
	return strings.Replace(path, fmt.Sprintf(".%s", hstnm), "", -1)
}

func (d *S3Detector) detectHTTP(src, hstnm string) (string, bool, error) {
	parts := strings.Split(src, "/")
	if len(parts) < 2 {
		return "", false, fmt.Errorf(
			"URL is not a valid S3 URL")
	}

	hostParts := strings.Split(parts[0], ".")
	if isPathStyle := strings.HasPrefix(parts[0], "s3"); isPathStyle {
		return d.detectPathStyle(hstnm, parts[0], parts[1:])
	}

	region := strings.Join(hostParts[1:], ".")
	return d.detectVhostStyle(
		hstnm, region, hostParts[0], parts[1:])
}

func (d *S3Detector) detectPathStyle(hstnm, region string, parts []string) (string, bool, error) {
	urlStr := fmt.Sprintf("https://%s.%s/%s", region, hstnm, strings.Join(parts, "/"))
	url, err := url.Parse(urlStr)
	if err != nil {
		return "", false, fmt.Errorf("error parsing S3 URL: %s", err)
	}

	return "s3::" + url.String(), true, nil
}

func (d *S3Detector) detectVhostStyle(hstnm, region, bucket string, parts []string) (string, bool, error) {
	urlStr := fmt.Sprintf("https://%s.%s/%s/%s", region, hstnm, bucket, strings.Join(parts, "/"))
	url, err := url.Parse(urlStr)
	if err != nil {
		return "", false, fmt.Errorf("error parsing S3 URL: %s", err)
	}

	return "s3::" + url.String(), true, nil
}
