package getter

import (
	"fmt"
	"net/url"
	"strings"
)

// GitLabDetector implements Detector to detect GitLab URLs and turn
// them into URLs that the Git Getter can understand.
type GitLabDetector struct{}

func (d *GitLabDetector) Detect(src, _ string) (string, bool, error) {
	if len(src) == 0 {
		return "", false, nil
	}

	if strings.HasPrefix(src, "gitlab.com/") {
		return d.detectHTTP(src)
	}

	return "", false, nil
}

func (d *GitLabDetector) detectHTTP(src string) (string, bool, error) {
	repoUrl, err := url.Parse(fmt.Sprintf("https://%s", src))
	if err != nil {
		return "", true, fmt.Errorf("error parsing GitLab URL: %s", err)
	}

	parts := strings.Split(repoUrl.Path, "//")

	if len(strings.Split(parts[0], "/")) < 3 {
		return "", false, fmt.Errorf(
			"GitLab URLs should be gitlab.com/username/repo " +
				"or gitlab.com/organization/project/repo")
	}

	if len(parts) > 2 {
		return "", false, fmt.Errorf(
			"URL malformed: \"//\" can only used once in path")
	}

	if !strings.HasSuffix(parts[0], ".git") {
		parts[0] += ".git"
	}

	repoUrl.Path = fmt.Sprintf("%s", strings.Join(parts, "//"))

	return "git::" + repoUrl.String(), true, nil
}
