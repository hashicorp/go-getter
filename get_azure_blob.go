package getter

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	storage "github.com/Azure/azure-storage-go"
)

// AzureBlobGetter is a Getter implementation that will download a module from
// an Azure Blob Storage Account.
type AzureBlobGetter struct{}

func (g *AzureBlobGetter) ClientMode(u *url.URL) (ClientMode, error) {
	// Parse URL
	accountName, baseURL, containerName, blobPath, accessKey, err := g.parseUrl(u)
	if err != nil {
		return 0, err
	}

	client, err := g.getBobClient(accountName, baseURL, accessKey)
	if err != nil {
		return 0, err
	}

	container := client.GetContainerReference(containerName)

	// List the object(s) at the given prefix
	params := storage.ListBlobsParameters{
		Prefix: blobPath,
	}
	resp, err := container.ListBlobs(params)
	if err != nil {
		return 0, err
	}

	for _, b := range resp.Blobs {
		// Use file mode on exact match.
		if b.Name == blobPath {
			return ClientModeFile, nil
		}

		// Use dir mode if child keys are found.
		if strings.HasPrefix(b.Name, blobPath+"/") {
			return ClientModeDir, nil
		}
	}

	// There was no match, so just return file mode. The download is going
	// to fail but we will let Azure return the proper error later.
	return ClientModeFile, nil
}

func (g *AzureBlobGetter) Get(dst string, u *url.URL) error {
	// Parse URL
	accountName, baseURL, containerName, blobPath, accessKey, err := g.parseUrl(u)
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

	// Create all the parent directories
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	client, err := g.getBobClient(accountName, baseURL, accessKey)
	if err != nil {
		return err
	}

	container := client.GetContainerReference(containerName)

	// List files in path, keep listing until no more objects are found
	lastMarker := ""
	hasMore := true
	for hasMore {
		params := storage.ListBlobsParameters{
			Prefix: blobPath,
		}
		if lastMarker != "" {
			params.Marker = lastMarker
		}

		resp, err := container.ListBlobs(params)
		if err != nil {
			return err
		}

		hasMore = resp.NextMarker != ""
		lastMarker = resp.NextMarker

		// Get each object storing each file relative to the destination path
		for _, object := range resp.Blobs {
			objPath := object.Name

			// If the key ends with a backslash assume it is a directory and ignore
			if strings.HasSuffix(objPath, "/") {
				continue
			}

			// Get the object destination path
			objDst, err := filepath.Rel(blobPath, objPath)
			if err != nil {
				return err
			}
			objDst = filepath.Join(dst, objDst)

			if err := g.getObject(client, objDst, containerName, objPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func (g *AzureBlobGetter) GetFile(dst string, u *url.URL) error {
	accountName, baseURL, containerName, blobPath, accessKey, err := g.parseUrl(u)
	if err != nil {
		return err
	}

	client, err := g.getBobClient(accountName, baseURL, accessKey)
	if err != nil {
		return err
	}

	return g.getObject(client, dst, containerName, blobPath)
}

func (g *AzureBlobGetter) getObject(client storage.BlobStorageClient, dst, container, blobName string) error {
	r, err := client.GetBlob(container, blobName)
	if err != nil {
		return err
	}

	// Create all the parent directories
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, r)
	return err
}

func (g *AzureBlobGetter) getBobClient(accountName string, baseURL string, accountKey string) (storage.BlobStorageClient, error) {
	var b storage.BlobStorageClient

	if accountKey == "" {
		accountKey = os.Getenv("ARM_ACCESS_KEY")
	}

	c, err := storage.NewClient(accountName, accountKey, baseURL, storage.DefaultAPIVersion, true)
	if err != nil {
		return b, err
	}

	b = c.GetBlobService()

	return b, nil
}

func (g *AzureBlobGetter) parseUrl(u *url.URL) (accountName, baseURL, container, blobPath, accessKey string, err error) {
	// Expected host style: accountname.blob.core.windows.net.
	// The last 3 parts will be different across environments.
	hostParts := strings.SplitN(u.Host, ".", 3)
	if len(hostParts) != 3 {
		err = fmt.Errorf("URL is not a valid Azure Blob URL")
		return
	}

	accountName = hostParts[0]
	baseURL = hostParts[2]

	pathParts := strings.SplitN(strings.TrimPrefix(u.Path, "/"), "/", 2)
	if len(pathParts) != 2 {
		err = fmt.Errorf("URL is not a valid Azure Blob URL")
		return
	}

	container = pathParts[0]
	blobPath = pathParts[1]

	accessKey = u.Query().Get("access_key")

	return
}
