// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package getter

import (
	"fmt"
	"net/url"
	"strings"
)

// GitHubDetector implements Detector to detect GitHub URLs and turn
// them into URLs that the Git Getter can understand.
type GitHubDetector struct{}

func (d *GitHubDetector) Detect(src, _ string) (string, bool, error) {
	if len(src) == 0 {
		return "", false, nil
	}

	if strings.HasPrefix(src, "github.com/") {
		return d.detectHTTP(src)
	}

	return "", false, nil
}

func (d *GitHubDetector) detectHTTP(src string) (string, bool, error) {
	parts := strings.Split(src, "?")
	if len(parts) > 2 {
		return "", false, fmt.Errorf("there is more than 1 '?' in the URL")
	}
	hostAndPath := parts[0]
	hostAndPathParts := strings.Split(hostAndPath, "/")
	if len(hostAndPathParts) < 3 {
		return "", false, fmt.Errorf(
			"GitHub URLs should be github.com/username/repo")
	}
	urlStr := fmt.Sprintf("https://%s", strings.Join(hostAndPathParts[:3], "/"))
	url, err := url.Parse(urlStr)
	if err != nil {
		return "", true, fmt.Errorf("error parsing GitHub URL: %s", err)
	}

	if !strings.HasSuffix(url.Path, ".git") {
		url.Path += ".git"
	}

	if len(hostAndPathParts) > 3 {
		url.Path += "//" + strings.Join(hostAndPathParts[3:], "/")
	}

	if len(parts) == 2 {
		url.RawQuery = parts[1]
	}

	return "git::" + url.String(), true, nil
}
