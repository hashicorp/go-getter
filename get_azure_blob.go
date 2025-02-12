package getter

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/go-azure-sdk/sdk/auth"
	"github.com/hashicorp/go-azure-sdk/sdk/environments"
	"github.com/jackofallops/giovanni/storage/2020-08-04/blob/containers"
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

func (g *AzureBlobGetter) ClientMode(u *url.URL) (ClientMode, error) {
	// Parse URL
	ctx := g.Context()

	if g.Timeout > 0 {
		var cancel context.CancelFunc
		_, cancel = context.WithTimeout(ctx, g.Timeout)
		defer cancel()
	}

	// fileName
	containerName, _, creds, err := g.parseUrl(u)
	if err != nil {
		return 0, err
	}

	client, err := g.newContainersClient(u, creds)
	if err != nil {
		return 0, err
	}

	// List the object(s) at the given prefix
	listBlobs := containers.ListBlobsInput{}
	if _, err := client.ListBlobs(ctx, containerName, listBlobs); err != nil {
		return 0, err
	}

	// TODO get if it's file or dir mode

	return ClientModeFile, nil
}

func (g *AzureBlobGetter) Get(dst string, u *url.URL) error {
	ctx := g.Context()

	if g.Timeout > 0 {
		var cancel context.CancelFunc
		_, cancel = context.WithTimeout(ctx, g.Timeout)
		defer cancel()
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

func (g *AzureBlobGetter) parseUrl(u *url.URL) (containerName, fileName, creds *auth.NewSharedKeyAuthorizer, err error) {
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

	environment := environments.Public
	credentials := auth.Credentials{
		Environment:                       environment,
		EnableAuthenticatingUsingAzureCLI: true,
	}
	creds, err = auth.NewAuthorizerFromCredentials(ctx, credentials, environment.MSGraph)
	if err != nil {
		log.Fatalf("building authorizer from credentials: %+v", err)
	}

	return
}

func (g *AzureBlobGetter) newContainersClient(url *url.URL, creds auth.NewSharedKeyAuthorizer) (*containers.Client, error) {

	containersClient, err := containers.NewWithBaseUri(fmt.Sprintf("https://%s", url.Host))
	if err != nil {
		return nil, fmt.Errorf("building client for environment: %v", err)
	}

	containersClient.Client.SetAuthorizer(creds)

	return containersClient, nil
}
