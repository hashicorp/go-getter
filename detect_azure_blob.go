package getter

import (
	"fmt"
	"net/url"
	"strings"
)

type AzureBlobDetector struct{}

func (d *AzureBlobDetector) Detect(src, pwd string) (string, bool, error) {
	if len(src) == 0 {
		return "", false, nil
	}

	if strings.Contains(src, ".blob.core.windows.net") {
		return d.detectURL(src)
	}

	return "", false, nil
}

func (d *AzureBlobDetector) detectURL(src string) (string, bool, error) {
	u, err := url.Parse(src)
	if err != nil {
		return "", false, err
	}

	if err := validateScheme(u.Scheme); err != nil {
		return "", false, err
	}

	if err := validateAzureBlobHost(u.Host); err != nil {
		return "", false, err
	}

	if err := validateBlobPath(u.Path); err != nil {
		return "", false, err
	}

	u.Scheme = "https"

	return fmt.Sprintf("azureblob::%s", u.String()), true, nil
}

func validateScheme(scheme string) error {
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("invalid scheme: %s, must be http or https", scheme)
	}
	return nil
}

func validateAzureBlobHost(host string) error {
	hostParts := strings.Split(host, ".")
	if len(hostParts) != 4 || hostParts[1] != "blob" || hostParts[2] != "core" || hostParts[3] != "windows.net" {
		return fmt.Errorf("invalid Azure Blob Storage hostname: %s", host)
	}
	return nil
}

func validateBlobPath(path string) error {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) < 1 || parts[0] == "" {
		return fmt.Errorf("path to blob must contain at least a container name")
	}
	return nil
}
