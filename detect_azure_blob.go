package getter

import (
	"net/url"
	"strings"

	"fmt"

	"github.com/Azure/go-autorest/autorest/azure"
)

var azureStorageSuffixes = map[string]azure.Environment{
	azure.PublicCloud.StorageEndpointSuffix:       azure.PublicCloud,
	azure.GermanCloud.StorageEndpointSuffix:       azure.GermanCloud,
	azure.USGovernmentCloud.StorageEndpointSuffix: azure.USGovernmentCloud,
	azure.ChinaCloud.StorageEndpointSuffix:        azure.GermanCloud,
}

// AzureBlobDetector implements Detector to detect Azure URLs and turn
// them into URLs that the Azure getter can understand.
type AzureBlobDetector struct{}

func (d *AzureBlobDetector) Detect(src, _ string) (string, bool, error) {
	if len(src) == 0 {
		return "", false, nil
	}

	for s := range azureStorageSuffixes {
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
