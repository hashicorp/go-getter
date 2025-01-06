// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package getter

import (
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"
	"unicode"
)

// GCSDetector implements Detector to detect GCS URLs and turn
// them into URLs that the GCSGetter can understand.
type GCSDetector struct{}

func (d *GCSDetector) Detect(src, _ string) (string, bool, error) {
	if len(src) == 0 {
		return "", false, nil
	}

	if strings.Contains(src, ".googleapis.com/") {
		return d.detectHTTP(src)
	}

	return "", false, nil
}

func (d *GCSDetector) detectHTTP(src string) (string, bool, error) {
	src = path.Clean(src)

	parts := strings.Split(src, "/")
	if len(parts) < 5 {
		return "", false, fmt.Errorf(
			"URL is not a valid GCS URL")
	}

	version := parts[2]
	if !isValidGCSVersion(version) {
		return "", false, fmt.Errorf(
			"GCS URL version is not valid")
	}

	bucket := parts[3]
	if !isValidGCSBucketName(bucket) {
		return "", false, fmt.Errorf(
			"GCS URL bucket name is not valid")
	}

	object := strings.Join(parts[4:], "/")
	if !isValidGCSObjectName(object) {
		return "", false, fmt.Errorf(
			"GCS URL object name is not valid")
	}

	url, err := url.Parse(fmt.Sprintf("https://www.googleapis.com/storage/%s/%s/%s",
		version, bucket, object))
	if err != nil {
		return "", false, fmt.Errorf("error parsing GCS URL: %s", err)
	}

	return "gcs::" + url.String(), true, nil
}

func isValidGCSVersion(version string) bool {
	versionPattern := `^v\d+$`
	if matched, _ := regexp.MatchString(versionPattern, version); !matched {
		return false
	}
	return true
}

// Validate the bucket name using the following rules: https://cloud.google.com/storage/docs/naming-buckets
func isValidGCSBucketName(bucket string) bool {
	// Rule 1: Must be between 3 and 63 characters (or up to 222 if it contains dots, each component up to 63 chars)
	if len(bucket) < 3 || len(bucket) > 63 {
		if len(bucket) > 63 && len(bucket) <= 222 {
			// If it contains dots, each segment between dots must be <= 63 chars
			components := strings.Split(bucket, ".")
			for _, component := range components {
				if len(component) > 63 {
					return false
				}
			}
		} else {
			return false
		}
	}

	// Rule 2: Bucket name cannot start or end with a hyphen, dot, or underscore
	if bucket[0] == '-' || bucket[0] == '.' || bucket[len(bucket)-1] == '-' || bucket[len(bucket)-1] == '.' || bucket[len(bucket)-1] == '_' {
		return false
	}

	// Rule 3: Bucket name cannot contain spaces
	if strings.Contains(bucket, " ") {
		return false
	}

	// Rule 4: Bucket name cannot be an IP address (only digits and dots, e.g., 192.168.5.4)
	ipPattern := `^(\d{1,3}\.){3}\d{1,3}$`
	if matched, _ := regexp.MatchString(ipPattern, bucket); matched {
		return false
	}

	// Rule 5: Bucket name cannot start with "goog"
	if strings.HasPrefix(bucket, "goog") {
		return false
	}

	// Rule 6: Bucket name cannot contain "google" or common misspellings like "g00gle"
	googlePattern := `google|g00gle`
	if matched, _ := regexp.MatchString(googlePattern, bucket); matched {
		return false
	}

	// Rule 7: Bucket name can only contain lowercase letters, digits, dashes, underscores, and dots
	bucketPattern := `^[a-z0-9\-_\.]+$`
	if matched, _ := regexp.MatchString(bucketPattern, bucket); !matched {
		return false
	}

	return true
}

// Validate the object name using the following rules: https://cloud.google.com/storage/docs/naming-objects
func isValidGCSObjectName(object string) bool {
	// Rule 1: Object names cannot contain Carriage Return (\r) or Line Feed (\n) characters
	if strings.Contains(object, "\r") || strings.Contains(object, "\n") {
		return false
	}

	// Rule 2: Object names cannot start with '.well-known/acme-challenge/'
	if strings.HasPrefix(object, ".well-known/acme-challenge/") {
		return false
	}

	// Rule 3: Object names cannot be exactly '.' or '..'
	if object == "." || object == ".." {
		return false
	}

	// Rule 4: Ensure that the object name contains only valid Unicode characters
	// (for simplicity, let's ensure it's not empty and does not contain any forbidden control characters)
	for _, r := range object {
		if !unicode.IsPrint(r) && !unicode.IsSpace(r) && r != '.' && r != '-' && r != '/' {
			return false
		}
	}

	return true
}
