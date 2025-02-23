package getter

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/go-azure-sdk/sdk/auth"
	"github.com/hashicorp/go-azure-sdk/sdk/environments"

	"github.com/jackofallops/giovanni/storage/2023-11-03/blob/blobs"
	"github.com/jackofallops/giovanni/storage/2023-11-03/blob/containers"
)

type AzureBlobGetter struct {
	getter

	Timeout time.Duration
}

// Interesting links
// https://pkg.go.dev/github.com/redtenant/go-azure-sdk/sdk/auth#section-readme
// https://github.com/jackofallops/giovanni/tree/main/storage/2023-11-03/blob/containers
// https://github.com/jackofallops/giovanni/blob/3916641df25097d26ec240814c9cdd2b6d89ba31/storage/2023-11-03/blob/containers/list_blobs.go
// https://github.com/hashicorp/go-getter/pull/395
// https://github.com/manicminer/hamilton/blob/main/example/example.go

func (g *AzureBlobGetter) ClientMode(u *url.URL) (ClientMode, error) {

	log.Println("standard logger")

	// Parse URL
	ctx := g.Context()

	if g.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, g.Timeout)
		defer cancel()
	}

	// Parse URL
	containerName, fileName, creds, err := g.parseUrl(u)
	if err != nil {
		return 0, err
	}

	// Create client config
	client, err := g.newContainersClient(u, creds)
	if err != nil {
		return 0, err
	}

	// List the object(s) at the given prefix
	listBlobs := containers.ListBlobsInput{
		Prefix: &fileName,
	}
	resp, err := client.ListBlobs(ctx, containerName, listBlobs)
	if err != nil {
		log.Fatal(err)
		return 0, err
	}

	for _, o := range resp.Blobs.Blobs {

		if o.Name == fileName {
			return ClientModeFile, nil
		}

		// Use dir mode if child keys are found.
		if strings.HasPrefix(o.Name, fileName) {
			return ClientModeDir, nil
		}
	}

	return ClientModeDir, nil
}

func (g *AzureBlobGetter) Get(dst string, u *url.URL) error {
	ctx := g.Context()

	if g.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, g.Timeout)
		defer cancel()
	}

	// Parse URL
	containerName, fileName, creds, err := g.parseUrl(u)
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
	if err := os.MkdirAll(filepath.Dir(dst), g.client.mode(0755)); err != nil {
		return err
	}

	containerClient, err := g.newContainersClient(u, creds)
	if err != nil {
		return err
	}

	blobClient, err := g.newBlobClient(u, creds)
	if err != nil {
		return err
	}

	// List files in path, keep listing until no more objects are found
	var lastMarker *string
	hasMore := true
	max := 1
	for hasMore {
		listBlobs := containers.ListBlobsInput{
			Prefix:     &fileName,
			MaxResults: &max,
		}
		if lastMarker != nil && *lastMarker != "" {
			listBlobs.Marker = lastMarker
		}

		resp, err := containerClient.ListBlobs(ctx, containerName, listBlobs)
		if err != nil {
			log.Fatal(err)
			return err
		}

		hasMore = *resp.NextMarker != ""
		lastMarker = resp.NextMarker
		for _, object := range resp.Blobs.Blobs {
			objPath := object.Name

			// If the key ends with a backslash assume it is a directory and ignore
			if strings.HasSuffix(objPath, "/") {
				continue
			}

			// Get the object destination path
			objDst, err := filepath.Rel(fileName, objPath)
			if err != nil {
				return err
			}
			objDst = filepath.Join(dst, objDst)

			if err := g.getObject(ctx, blobClient, containerName, objDst, objPath); err != nil {
				return err
			}
		}

	}

	return nil
}

func (g *AzureBlobGetter) GetFile(dst string, u *url.URL) error {
	ctx := g.Context()

	if g.Timeout > 0 {
		var cancel context.CancelFunc
		_, cancel = context.WithTimeout(ctx, g.Timeout)
		defer cancel()
	}

	return nil
}

func (g *AzureBlobGetter) parseUrl(u *url.URL) (containerName, fileName string, creds auth.Authorizer, err error) {
	ctx := g.Context()

	if u == nil {
		err = fmt.Errorf("invalid URL: nil value provided")
		return
	}

	hostParts := strings.Split(u.Host, ".")
	if len(hostParts) < 4 || hostParts[1] != "blob" || hostParts[2] != "core" {
		err = fmt.Errorf("invalid Azure Blob Storage hostname: %s", u.Host)
		return
	}

	pathParts := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	if len(pathParts) < 1 || pathParts[0] == "" {
		err = fmt.Errorf("URL path must contain at least a container name")
		return
	}

	containerName = pathParts[0]
	if len(pathParts) > 1 {
		fileName = strings.Join(pathParts[1:], "/") // Remaining part as file name
	}

	environment := environments.AzurePublic()
	credentials := auth.Credentials{
		Environment:                       *environment,
		EnableAuthenticatingUsingAzureCLI: true,
	}
	creds, err = auth.NewAuthorizerFromCredentials(ctx, credentials, environment.Storage)
	if err != nil {
		log.Fatalf("building authorizer from credentials: %+v", err)
	}

	return
}

func (g *AzureBlobGetter) newContainersClient(url *url.URL, creds auth.Authorizer) (*containers.Client, error) {

	containersClient, err := containers.NewWithBaseUri(fmt.Sprintf("https://%s", url.Host))
	if err != nil {
		return nil, fmt.Errorf("building client for environment: %v", err)
	}

	containersClient.Client.SetAuthorizer(creds)

	return containersClient, nil
}

func (g *AzureBlobGetter) newBlobClient(url *url.URL, creds auth.Authorizer) (*blobs.Client, error) {

	blobClient, err := blobs.NewWithBaseUri(fmt.Sprintf("https://%s", url.Host))
	if err != nil {
		return nil, fmt.Errorf("building client for environment: %v", err)
	}

	blobClient.Client.SetAuthorizer(creds)

	return blobClient, nil
}

func (g *AzureBlobGetter) getObject(ctx context.Context, client *blobs.Client, containerName, dst, key string) error {
	req := blobs.GetInput{}

	resp, err := client.Get(ctx, containerName, key, req)
	if err != nil {
		return err
	}

	if resp.Contents == nil {
		return nil
	}

	// Create all the parent directories
	if err := os.MkdirAll(filepath.Dir(dst), g.client.mode(0755)); err != nil {
		return err
	}

	reader := bytes.NewReader(*resp.Contents)
	return copyReader(dst, reader, 0666, g.client.umask(), 0)
}
