package getter

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
)

type AzureBlobGetter struct {
	getter

	// Timeout sets a deadline which all AzureBlob operations should
	// complete within. Zero value means no timeout.
	Timeout time.Duration
}

// Get downloads the given URL into the given directory. This always
// assumes that we're updating and gets the latest version that it can.
//
// The directory may already exist (if we're updating). If it is in a
// format that isn't understood, an error should be returned. Get shouldn't
// simply nuke the directory.
func (g *AzureBlobGetter) Get(dst string, url *url.URL) error {
	ctx := g.Context()

	if g.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, g.Timeout)
		defer cancel()
	}

	// Parse URL
	accountName, baseUrl, containerName, blobPath, _, err := g.parseUrl(url)
	if err != nil {
		return err
	}

	// Remove destination if it already exists
	_, err = os.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if err == nil {
		// Remove the destination
		if err := os.RemoveAll(dst); err != nil {
			return err
		}
	}

	// Create client config
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return err
	}
	containerUrl, err := url.Parse(fmt.Sprintf("https://%s.%s/%s", accountName, baseUrl, containerName))
	if err != nil {
		return err
	}

	client, err := g.newAzureContainerClient(containerUrl, cred)
	if err != nil {
		return err
	}

	pager := client.NewListBlobsFlatPager(&container.ListBlobsFlatOptions{
		Prefix: &blobPath,
	})

	for pager.More() {
		r, err := pager.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, item := range r.ListBlobsFlatSegmentResponse.Segment.BlobItems {
			blobFullName := *item.Name

			blobPathPart := filepath.Dir(blobFullName)
			blobName := filepath.Base(blobFullName)
			dstPathPart := strings.TrimPrefix(blobPathPart, blobPath)

			dst := strings.Join([]string{dst, dstPathPart}, "/")

			blobClient := client.NewBlobClient(blobFullName)
			err = os.MkdirAll(dst, os.ModeDir)
			if err != nil {
				return err
			}

			f, err := os.Create(dst + "/" + blobName)
			if err != nil {
				return err
			}
			defer f.Close()
			_, err = blobClient.DownloadFile(ctx, f, nil)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// GetFile downloads the give URL into the given path. The URL must
// reference a single file. If possible, the Getter should check if
// the remote end contains the same file and no-op this operation.
func (g *AzureBlobGetter) GetFile(dst string, url *url.URL) error {
	ctx := g.Context()

	if g.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, g.Timeout)
		defer cancel()
	}

	// Parse URL
	accountName, baseUrl, containerName, blobPath, _, err := g.parseUrl(url)
	if err != nil {
		return err
	}

	// Remove destination if it already exists
	_, err = os.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if err == nil {
		// Remove the destination
		if err := os.RemoveAll(dst); err != nil {
			return err
		}
	}

	// Create client config
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return err
	}
	containerUrl, err := url.Parse(fmt.Sprintf("https://%s.%s/%s", accountName, baseUrl, containerName))
	if err != nil {
		return err
	}

	client, err := g.newAzureContainerClient(containerUrl, cred)
	if err != nil {
		return err
	}

	blobClient := client.NewBlobClient(blobPath)
	err = os.MkdirAll(filepath.Dir(dst), os.ModeDir)
	if err != nil {
		return err
	}

	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = blobClient.DownloadFile(ctx, f, nil)
	if err != nil {
		return err
	}
	return nil
}

// ClientMode returns the mode based on the given URL. This is used to
// allow clients to let the getters decide which mode to use.
func (g *AzureBlobGetter) ClientMode(url *url.URL) (ClientMode, error) {
	ctx := g.Context()

	if g.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, g.Timeout)
		defer cancel()
	}

	// Parse URL
	accountName, baseUrl, containerName, blobPath, _, err := g.parseUrl(url)
	if err != nil {
		return ClientModeInvalid, err
	}
	if blobPath == "" {
		// Root Path so use DirMode
		return ClientModeDir, nil
	}

	// Create client config
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return ClientModeInvalid, err
	}
	containerUrl, err := url.Parse(fmt.Sprintf("https://%s.%s/%s", accountName, baseUrl, containerName))
	if err != nil {
		return ClientModeInvalid, err
	}

	client, err := g.newAzureContainerClient(containerUrl, cred)
	if err != nil {
		return ClientModeInvalid, err
	}

	pager := client.NewListBlobsFlatPager(&container.ListBlobsFlatOptions{
		Prefix: &blobPath,
	})

	for pager.More() {
		r, err := pager.NextPage(ctx)
		if err != nil {
			return ClientModeInvalid, err
		}
		blobs := r.ListBlobsFlatSegmentResponse.Segment.BlobItems
		if len(blobs) == 1 && *blobs[0].Name == blobPath {
			return ClientModeFile, nil
		} else {
			return ClientModeDir, nil
		}
	}
	return ClientModeInvalid, nil

}

func (g *AzureBlobGetter) newAzureContainerClient(url *url.URL, cred azcore.TokenCredential) (client *container.Client, err error) {

	client, err = container.NewClient(url.String(), cred, nil)
	return
}

func (g *AzureBlobGetter) parseUrl(u *url.URL) (accountName, baseURL, container, blobPath, accessKey string, err error) {
	// Expected host style: accountname.blob.core.windows.net.
	// The last 3 parts will be different across environments.
	hostParts := strings.SplitN(u.Host, ".", 2)
	if len(hostParts) != 2 {
		err = fmt.Errorf("URL is not a valid Azure Blob URL: %v", hostParts)
		return
	}

	accountName = hostParts[0]
	baseURL = hostParts[1]

	pathParts := strings.SplitN(strings.TrimPrefix(u.Path, "/"), "/", 2)
	if len(pathParts) < 1 {
		err = fmt.Errorf("URL is not a valid Azure Blob URL: %v", pathParts)
		return
	}

	container = pathParts[0]
	if len(pathParts) > 1 {
		blobPath = pathParts[1]
	} else {
		blobPath = ""
	}

	accessKey = u.Query().Get("access_key")

	return
}
