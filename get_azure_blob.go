package getter

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"log"
	"bytes"
	//
	"github.com/Azure/azure-storage-blob-go/azblob"
)
// TODO: https://docs.microsoft.com/en-us/azure/storage/blobs/storage-quickstart-blobs-go?tabs=windows#understand-the-sample-code


// AzureBlobGetter is a Getter implementation that will download a module from
// an Azure Blob Storage Account.
type AzureBlobGetter struct {
	getter
}

func (g *AzureBlobGetter) ClientMode(u *url.URL) (ClientMode, error) {
	//// Parse URL
	//accountName, baseURL, containerName, blobPath, accessKey, err := g.parseUrl(u)
	//if err != nil {
	//	return 0, err
	//}
	//
	//client, err := g.getBobClient(accountName, baseURL, accessKey)
	//if err != nil {
	//	return 0, err
	//}
	//
	//container := client.GetContainerReference(containerName)
	//
	//containerReference := storage.GetContainerReference(containerName)
	//blobReference := containerReference.GetBlobReference(c.keyName)
	//options := &storage.GetBlobOptions{}
	//
	//// List the object(s) at the given prefix
	//params := storage.ListBlobsParameters{
	//	Prefix: blobPath,
	//}
	//resp, err := container.ListBlobs(params)
	//if err != nil {
	//	return 0, err
	//}
	//
	//for _, b := range resp.Blobs {
	//	// Use file mode on exact match.
	//	if b.Name == blobPath {
	//		return ClientModeFile, nil
	//	}
	//
	//	// Use dir mode if child keys are found.
	//	if strings.HasPrefix(b.Name, blobPath+"/") {
	//		return ClientModeDir, nil
	//	}
	//}
	//
	//// There was no match, so just return file mode. The download is going
	//// to fail but we will let Azure return the proper error later.
	//return ClientModeFile, nil
	//ClientModeFile := nil

	// From the Azure portal, get your storage account name and key and set environment variables.
	//accountName, accountKey := os.Getenv("AZURE_STORAGE_ACCOUNT"), os.Getenv("AZURE_STORAGE_ACCESS_KEY")
	//if len(accountName) == 0 || len(accountKey) == 0 {
	//	log.Fatal("Either the AZURE_STORAGE_ACCOUNT or AZURE_STORAGE_ACCESS_KEY environment variable is not set")
	//}
	//
	//// Create a default request pipeline using your storage account name and account key.
	//credential, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	//if err != nil {
	//	log.Fatal("Invalid credentials with error: " + err.Error())
	//}
	//p := azblob.NewPipeline(credential, azblob.PipelineOptions{})
	//
	//// Create a random string for the quick start container
	//containerName := fmt.Sprintf("quickstart-%s", randomString())
	//
	//// From the Azure portal, get your storage account blob service URL endpoint.
	//URL, _ := url.Parse(
	//	fmt.Sprintf("https://%s.blob.core.windows.net/%s", accountName, containerName))
	//
	//// Create a ContainerURL object that wraps the container URL and a request
	//// pipeline to make requests.
	//containerURL := azblob.NewContainerURL(*URL, p)
	//
	//// Create the container
	//fmt.Printf("Creating a container named %s\n", containerName)
	//ctx := context.Background() // This example uses a never-expiring context
	//_, err = containerURL.Create(ctx, azblob.Metadata{}, azblob.PublicAccessNone)
	//handleErrors(err)
	//
	//
	//
	//
	//// Here's how to download the blob
	//downloadResponse, err := blobURL.Download(ctx, 0, azblob.CountToEnd, azblob.BlobAccessConditions{}, false)
	//
	//// NOTE: automatically retries are performed if the connection fails
	//bodyStream := downloadResponse.Body(azblob.RetryReaderOptions{MaxRetryRequests: 20})
	//
	//// read the body into a buffer
	//downloadedData := bytes.Buffer{}
	//_, err = downloadedData.ReadFrom(bodyStream)
	//handleErrors(err)





	return ClientModeFile, nil
}

func (g *AzureBlobGetter) Get(dst string, u *url.URL) error {
	// Parse URL
	//accountName, baseURL, containerName, blobPath, accessKey, err := g.parseUrl(u)
	//if err != nil {
	//	return err
	//}
	//
	//// Remove destination if it already exists
	//_, err = os.Stat(dst)
	//if err != nil && !os.IsNotExist(err) {
	//	return err
	//}
	//
	//if err == nil {
	//	// Remove the destination
	//	if err := os.RemoveAll(dst); err != nil {
	//		return err
	//	}
	//}
	//
	//// Create all the parent directories
	//if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
	//	return err
	//}
	//
	//client, err := g.getBobClient(accountName, baseURL, accessKey)
	//if err != nil {
	//	return err
	//}
	//
	//container := client.GetContainerReference(containerName)
	//
	//// List files in path, keep listing until no more objects are found
	//lastMarker := ""
	//hasMore := true
	//for hasMore {
	//	params := storage.ListBlobsParameters{
	//		Prefix: blobPath,
	//	}
	//	if lastMarker != "" {
	//		params.Marker = lastMarker
	//	}
	//
	//	resp, err := container.ListBlobs(params)
	//	if err != nil {
	//		return err
	//	}
	//
	//	hasMore = resp.NextMarker != ""
	//	lastMarker = resp.NextMarker
	//
	//	// Get each object storing each file relative to the destination path
	//	for _, object := range resp.Blobs {
	//		objPath := object.Name
	//
	//		// If the key ends with a backslash assume it is a directory and ignore
	//		if strings.HasSuffix(objPath, "/") {
	//			continue
	//		}
	//
	//		// Get the object destination path
	//		objDst, err := filepath.Rel(blobPath, objPath)
	//		if err != nil {
	//			return err
	//		}
	//		objDst = filepath.Join(dst, objDst)
	//
	//		if err := g.getObject(client, objDst, containerName, objPath); err != nil {
	//			return err
	//		}
	//	}
	//}

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

func (g *AzureBlobGetter) getObject(serviceURL azblob.ServiceURL, dst, container, blobName string) error {
	// All HTTP operations allow you to specify a Go context.Context object to control cancellation/timeout.
	//ctx := context.Background() // This example uses a never-expiring context.

	// This example shows several common operations just to get you started.

	// Create a URL that references a to-be-created container in your Azure Storage account.
	// This returns a ContainerURL object that wraps the container's URL and a request pipeline (inherited from serviceURL)
	containerURL := serviceURL.NewContainerURL(container) // Container names require lowercase

	//// Create the container on the service (with no metadata and no public access)
	//_, err = containerURL.Create(ctx, azblob.Metadata{}, azblob.PublicAccessNone)
	//if err != nil {
	//	log.Fatal(err)
	//}

	// Create a URL that references a to-be-created blob in your Azure Storage account's container.
	// This returns a BlockBlobURL object that wraps the blob's URL and a request pipeline (inherited from containerURL)
	blobURL := containerURL.NewBlockBlobURL(blobName) // Blob names can be mixed case

	// Download the blob's contents and verify that it worked correctly
	get, err := blobURL.Download(nil, 0, 0, azblob.BlobAccessConditions{}, false)
	if err != nil {
		log.Fatal(err)
	}

	downloadedData := &bytes.Buffer{}
	reader := get.Body(azblob.RetryReaderOptions{})
	downloadedData.ReadFrom(reader)
	reader.Close() // The client must close the response body when finished with it

	//r, err := client(container, blobName)
	//if err != nil {
	//	return err
	//}
	//

	// Create all the parent directories
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, downloadedData)
	return err
}

func (g *AzureBlobGetter) getBobClient(accountName string, baseURL string, accountKey string) (azblob.ServiceURL, error) {
	// Use your Storage account's name and key to create a credential object; this is used to access your account.
	credential, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		log.Fatal(err)
	}

	// Create a request pipeline that is used to process HTTP(S) requests and responses. It requires
	// your account credentials. In more advanced scenarios, you can configure telemetry, retry policies,
	// logging, and other options. Also, you can configure multiple request pipelines for different scenarios.
	p := azblob.NewPipeline(credential, azblob.PipelineOptions{})

	// From the Azure portal, get your Storage account blob service URL endpoint.
	// The URL typically looks like this:
	u, _ := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net", accountName))

	// Create an ServiceURL object that wraps the service URL and a request pipeline.
	serviceURL := azblob.NewServiceURL(*u, p)

	return serviceURL, nil
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
