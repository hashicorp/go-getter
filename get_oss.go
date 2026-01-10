package getter

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"

	openapicred "github.com/aliyun/credentials-go/credentials"
)

// OSSGetter is a Getter implementation that will download a module from
// an OSS bucket.
type OSSGetter struct {
	getter

	// Timeout sets a deadline which all OSS operations should
	// complete within.
	//
	// The zero value means timeout.
	Timeout time.Duration
}

func (g *OSSGetter) ClientMode(u *url.URL) (ClientMode, error) {
	// Parse URL
	ctx := g.Context()

	if g.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, g.Timeout)
		defer cancel()
	}

	region, bucket, path, _, err := g.parseUrl(u)
	if err != nil {
		return 0, err
	}

	// Create client config
	client, err := g.newOSSClient(region, u)
	if err != nil {
		return 0, err
	}

	// List the object(s) at the given prefix
	request := &oss.ListObjectsV2Request{
		Bucket: oss.Ptr(bucket),
		Prefix: oss.Ptr(path),
	}

	p := client.NewListObjectsV2Paginator(request)

	var i int
	for p.HasNext() {
		i++
		page, err := p.NextPage(ctx)
		if err != nil {
			return 0, err
		}
		for _, obj := range page.Contents {
			// Use file mode on exact match.
			if *obj.Key == path {
				return ClientModeFile, nil
			}

			// Use dir mode if child keys are found.
			if strings.HasPrefix(*obj.Key, path+"/") {
				return ClientModeDir, nil
			}
		}
	}

	// There was no match, so just return file mode. The download is going
	// to fail but we will let OSS return the proper error later.
	return ClientModeFile, nil
}

func (g *OSSGetter) Get(dst string, u *url.URL) error {
	ctx := g.Context()

	if g.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, g.Timeout)
		defer cancel()
	}

	region, bucket, path, version, err := g.parseUrl(u)
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

	// Create client config
	client, err := g.newOSSClient(region, u)
	if err != nil {
		return err
	}

	// List files in path, keep listing until no more objects are found
	request := &oss.ListObjectsV2Request{
		Bucket: oss.Ptr(bucket),
		Prefix: oss.Ptr(path),
	}

	p := client.NewListObjectsV2Paginator(request)

	var i int
	for p.HasNext() {
		i++
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, obj := range page.Contents {
			objPath := *obj.Key

			// If the key ends with a backslash assume it is a directory and ignore
			if strings.HasSuffix(objPath, "/") {
				continue
			}

			// Get the object destination path
			objDst, err := filepath.Rel(path, objPath)
			if err != nil {
				return err
			}
			objDst = filepath.Join(dst, objDst)

			if err := g.getObject(ctx, client, objDst, bucket, objPath, version); err != nil {
				return err
			}

		}
	}

	return nil
}

func (g *OSSGetter) GetFile(dst string, u *url.URL) error {
	ctx := g.Context()

	if g.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, g.Timeout)
		defer cancel()
	}

	region, bucket, path, version, err := g.parseUrl(u)
	if err != nil {
		return err
	}

	client, err := g.newOSSClient(region, u)
	if err != nil {
		return err
	}

	return g.getObject(ctx, client, dst, bucket, path, version)
}

func (g *OSSGetter) getObject(ctx context.Context, client *oss.Client, dst, bucket, object, version string) error {
	request := &oss.GetObjectRequest{
		Bucket: oss.Ptr(bucket),
		Key:    oss.Ptr(object),
	}

	if version != "" {
		request.VersionId = oss.Ptr(version)
	}

	result, err := client.GetObject(ctx, request)
	if err != nil {
		return err
	}

	// Create all the parent directories
	if err := os.MkdirAll(filepath.Dir(dst), g.client.mode(0755)); err != nil {
		return err
	}

	body := result.Body

	// There is no limit set for the size of an object from OSS
	return copyReader(dst, body, 0666, g.client.umask(), 0)
}

func (g *OSSGetter) parseUrl(u *url.URL) (region, bucket, path, version string, err error) {
	if strings.Contains(u.Host, "aliyuncs.com") {
		hostParts := strings.Split(u.Host, ".")

		switch len(hostParts) {
		// path-style
		case 4:
			bucket = hostParts[0]
			region = strings.TrimPrefix(hostParts[1], "oss-")
			region = strings.TrimSuffix(region, "-internal")

		case 5:
			bucket = hostParts[0]
			region = hostParts[1]
		}

		pathParts := strings.SplitN(u.Path, "/", 2)
		if len(pathParts) != 2 {
			err = fmt.Errorf("URL is not a valid OSS URL")
			return
		}
		path = pathParts[1]

		if len(hostParts) < 4 || len(hostParts) > 5 {
			err = fmt.Errorf("URL is not a valid OSS URL")
			return
		}

		version = u.Query().Get("version")
	}
	return
}

func (g *OSSGetter) newOSSClient(region string, url *url.URL) (*oss.Client, error) {

	arnCredential, gerr := openapicred.NewCredential(nil)
	provider := credentials.CredentialsProviderFunc(func(ctx context.Context) (credentials.Credentials, error) {
		if gerr != nil {
			return credentials.Credentials{}, gerr
		}
		cred, err := arnCredential.GetCredential()
		if err != nil {
			return credentials.Credentials{}, err
		}
		return credentials.Credentials{
			AccessKeyID:     *cred.AccessKeyId,
			AccessKeySecret: *cred.AccessKeySecret,
			SecurityToken:   *cred.SecurityToken,
		}, nil
	})

	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(provider).
		WithRegion(region)

	client := oss.NewClient(cfg)

	return client, nil
}
