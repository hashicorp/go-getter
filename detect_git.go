package getter

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var scpLikeURLPattern = regexp.MustCompile("^(?:([^@]+)@)?([^:]+):/?(.+)$")

// GitDetector implements Detector to detect Git URLs
// (HTTP, SSH, and SCP style) and turn them into URLs
// that the Git Getter can understand.
type GitDetector struct{}

func (d *GitDetector) Detect(src, _ string) (string, bool, error) {
	if len(src) == 0 {
		return "", false, nil
	}

	if strings.HasPrefix(src, "github.com/") {
		return d.detectHTTP(src)
	} else if _, err := url.Parse(src); err != nil && scpLikeURLPattern.MatchString(src) {
		// not valid URL syntax, and looks like an SCP style string
		return d.detectSSH(src)
	}

	return "", false, nil
}

func (d *GitDetector) detectHTTP(src string) (string, bool, error) {
	parts := strings.Split(src, "/")
	if len(parts) < 3 {
		return "", false, fmt.Errorf(
			"GitHub URLs should be github.com/username/repo")
	}

	urlStr := fmt.Sprintf("https://%s", strings.Join(parts[:3], "/"))
	url, err := url.Parse(urlStr)
	if err != nil {
		return "", true, fmt.Errorf("error parsing GitHub URL: %s", err)
	}

	if !strings.HasSuffix(url.Path, ".git") {
		url.Path += ".git"
	}

	if len(parts) > 3 {
		url.Path += "//" + strings.Join(parts[3:], "/")
	}

	return "git::" + url.String(), true, nil
}

func (d *GitDetector) detectSSH(src string) (string, bool, error) {
	matched := scpLikeURLPattern.FindStringSubmatch(src)
	if matched == nil {
		return "", false, fmt.Errorf("error matching SCP style URL")
	}

	user := matched[1]
	host := matched[2]
	path := matched[3]

	qidx := strings.Index(path, "?")
	if qidx == -1 {
		qidx = len(path)
	}

	var u url.URL
	u.Scheme = "ssh"
	u.User = url.User(user)
	u.Host = host
	u.Path = path[0:qidx]
	if qidx < len(path) {
		q, err := url.ParseQuery(path[qidx+1:])
		if err != nil {
			return "", true, fmt.Errorf("error parsing GitHub SSH URL: %s", err)
		}

		u.RawQuery = q.Encode()
	}

	return "git::" + u.String(), true, nil
}
