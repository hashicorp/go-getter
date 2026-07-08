package s3

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/hashicorp/go-getter/v2"
)

// Getter is a Getter implementation that will download a module from
// a S3 bucket.
type Getter struct {

	// Timeout sets a deadline which all S3 operations should
	// complete within.
	//
	// The zero value means no timeout.
	Timeout time.Duration
}

func (g *Getter) Mode(ctx context.Context, u *url.URL) (getter.Mode, error) {

	if g.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, g.Timeout)
		defer cancel()
	}

	// Parse URL
	region, bucket, path, _, creds, err := g.parseUrl(u)
	if err != nil {
		return 0, err
	}

	// Create client config
	client, err := g.newS3Client(ctx, region, u, creds)
	if err != nil {
		return 0, err
	}

	// List the object(s) at the given prefix
	req := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(path),
	}
	paginator := s3.NewListObjectsV2Paginator(client, req)
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return 0, err
		}

		for _, o := range output.Contents {
			// Use file mode on exact match.
			if aws.ToString(o.Key) == path {
				return getter.ModeFile, nil
			}

			// Use dir mode if child keys are found.
			if strings.HasPrefix(aws.ToString(o.Key), path+"/") {
				return getter.ModeDir, nil
			}
		}
	}

	// There was no match, so just return file mode. The download is going
	// to fail but we will let S3 return the proper error later.
	return getter.ModeFile, nil
}

func (g *Getter) Get(ctx context.Context, req *getter.Request) error {

	if g.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, g.Timeout)
		defer cancel()
	}

	// Parse URL
	region, bucket, path, _, creds, err := g.parseUrl(req.URL())
	if err != nil {
		return err
	}

	// Remove destination if it already exists
	_, err = os.Stat(req.Dst)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if err == nil {
		// Remove the destination
		if err := os.RemoveAll(req.Dst); err != nil {
			return err
		}
	}

	// Create all the parent directories
	if err := os.MkdirAll(filepath.Dir(req.Dst), req.Mode(0755)); err != nil {
		return err
	}

	client, err := g.newS3Client(ctx, region, req.URL(), creds)
	if err != nil {
		return err
	}

	// List files in path, keep listing until no more objects are found
	s3Req := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(path),
	}
	paginator := s3.NewListObjectsV2Paginator(client, s3Req)
	for paginator.HasMorePages() {
		resp, err := paginator.NextPage(ctx)
		if err != nil {
			return err
		}

		// Get each object storing each file relative to the destination path
		for _, object := range resp.Contents {
			objPath := aws.ToString(object.Key)

			// If the key ends with a backslash assume it is a directory and ignore
			if strings.HasSuffix(objPath, "/") {
				continue
			}

			// Get the object destination path
			objDst, err := filepath.Rel(path, objPath)
			if err != nil {
				return err
			}
			objDst = filepath.Join(req.Dst, objDst)

			if err := g.getObject(ctx, client, req, objDst, bucket, objPath, ""); err != nil {
				return err
			}
		}
	}

	return nil
}

func (g *Getter) GetFile(ctx context.Context, req *getter.Request) error {

	if g.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, g.Timeout)
		defer cancel()
	}

	region, bucket, path, version, creds, err := g.parseUrl(req.URL())
	if err != nil {
		return err
	}

	client, err := g.newS3Client(ctx, region, req.URL(), creds)
	if err != nil {
		return err
	}

	return g.getObject(ctx, client, req, req.Dst, bucket, path, version)
}

func (g *Getter) getObject(ctx context.Context, client *s3.Client, req *getter.Request, dst, bucket, key, version string) error {
	s3req := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	if version != "" {
		s3req.VersionId = aws.String(version)
	}

	resp, err := client.GetObject(ctx, s3req)
	if err != nil {
		return err
	}

	// Create all the parent directories
	if err := os.MkdirAll(filepath.Dir(dst), req.Mode(0755)); err != nil {
		return err
	}

	// There is no limit set for the size of an object from S3
	return req.CopyReader(dst, resp.Body, 0666, 0)
}

func (g *Getter) getAWSConfig(ctx context.Context, region string, url *url.URL, staticCreds *credentials.StaticCredentialsProvider) (conf aws.Config, err error) {
	var loadOptions []func(*config.LoadOptions) error
	var creds aws.CredentialsProvider

	metadataURLOverride := os.Getenv("AWS_METADATA_URL")
	if staticCreds == nil && metadataURLOverride != "" {
		creds = ec2rolecreds.New(func(o *ec2rolecreds.Options) {
			o.Client = imds.New(imds.Options{
				Endpoint:          metadataURLOverride,
				ClientEnableState: imds.ClientEnabled,
			})
		})
	} else if staticCreds != nil {
		creds = staticCreds
	}

	if creds != nil {
		loadOptions = append(loadOptions,
			config.WithEC2IMDSClientEnableState(imds.ClientEnabled),
			config.WithCredentialsProvider(creds))
	}

	if region != "" {
		loadOptions = append(loadOptions, config.WithRegion(region))
	}

	return config.LoadDefaultConfig(ctx, loadOptions...)
}

func (g *Getter) parseUrl(u *url.URL) (region, bucket, path, version string, creds *credentials.StaticCredentialsProvider, err error) {
	// This just check whether we are dealing with S3 or
	// any other S3 compliant service. S3 has a predictable
	// url as others do not
	if strings.Contains(u.Host, "amazonaws.com") {
		// Expected host style: s3.amazonaws.com or s3-region.amazonaws.com or bucket.s3.region.amazonaws.com
		// They can have different formats
		hostParts := strings.Split(u.Host, ".")
		
		// Handle different S3 URL formats
		if len(hostParts) >= 3 {
			// Check if it's bucket.s3.region.amazonaws.com (vhost-style)
			if len(hostParts) == 4 && hostParts[1] == "s3" {
				region = hostParts[2]
				bucket = hostParts[0]
				pathParts := strings.SplitN(u.Path, "/", 2)
				if len(pathParts) < 2 {
					err = fmt.Errorf("URL is not a valid S3 URL")
					return
				}
				path = pathParts[1]
				version = u.Query().Get("version")
				return
			}
			
			// Handle s3.amazonaws.com or s3-region.amazonaws.com (path-style)
			if hostParts[0] == "s3" || strings.HasPrefix(hostParts[0], "s3-") {
				// Parse the region out of the first part of the host
				region = strings.TrimPrefix(strings.TrimPrefix(hostParts[0], "s3-"), "s3")
				if region == "" {
					region = "us-east-1"
				}

				pathParts := strings.SplitN(u.Path, "/", 3)
				if len(pathParts) != 3 {
					err = fmt.Errorf("URL is not a valid S3 URL")
					return
				}

				bucket = pathParts[1]
				path = pathParts[2]
				version = u.Query().Get("version")
				return
			}
		}
		
		err = fmt.Errorf("URL is not a valid S3 URL")
		return
	} else {
		// S3-compatible service (like Minio)
		pathParts := strings.SplitN(u.Path, "/", 3)
		if len(pathParts) != 3 {
			err = fmt.Errorf("URL is not a valid S3 compliant URL")
			return
		}
		bucket = pathParts[1]
		path = pathParts[2]
		version = u.Query().Get("version")
		region = u.Query().Get("region")
		if region == "" {
			region = "us-east-1"
		}
	}

	_, hasAwsId := u.Query()["aws_access_key_id"]
	_, hasAwsSecret := u.Query()["aws_access_key_secret"]
	_, hasAwsToken := u.Query()["aws_access_token"]
	if hasAwsId || hasAwsSecret || hasAwsToken {
		provider := credentials.NewStaticCredentialsProvider(
			u.Query().Get("aws_access_key_id"),
			u.Query().Get("aws_access_key_secret"),
			u.Query().Get("aws_access_token"),
		)
		creds = &provider
	}

	return
}

func (g *Getter) Detect(req *getter.Request) (bool, error) {
	src := req.Src
	if len(src) == 0 {
		return false, nil
	}

	if req.Forced != "" {
		// There's a getter being forced
		if !g.validScheme(req.Forced) {
			// Current getter is not the forced one
			// Don't use it to try to download the artifact
			return false, nil
		}
	}
	isForcedGetter := req.Forced != "" && g.validScheme(req.Forced)

	u, err := url.Parse(src)
	if err == nil && u.Scheme != "" {
		if isForcedGetter {
			// Is the forced getter and source is a valid url
			return true, nil
		}
		if g.validScheme(u.Scheme) {
			return true, nil
		}
		// Valid url with a scheme that is not valid for current getter
		return false, nil
	}

	src, ok, err := new(Detector).Detect(src, req.Pwd)
	if err != nil {
		return ok, err
	}
	if ok {
		req.Src = src
	}

	return ok, nil
}

func (g *Getter) validScheme(scheme string) bool {
	return scheme == "s3"
}

func (g *Getter) newS3Client(
	ctx context.Context, region string, url *url.URL, creds *credentials.StaticCredentialsProvider,
) (*s3.Client, error) {
	var cfg aws.Config

	if profile := url.Query().Get("aws_profile"); profile != "" {
		var err error
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithSharedConfigProfile(profile),
		)
		if err != nil {
			return nil, err
		}
	} else {
		var err error
		cfg, err = g.getAWSConfig(ctx, region, url, creds)
		if err != nil {
			return nil, err
		}
	}

	clientOptions := func(opts *s3.Options) {
		opts.UsePathStyle = true
		// Check if this is a custom S3-compatible endpoint (not AWS)
		if !strings.Contains(url.Host, "amazonaws.com") {
			scheme := url.Scheme
			if scheme == "" {
				scheme = "https"
			}
			opts.BaseEndpoint = aws.String(scheme + "://" + url.Host)
		}
	}

	return s3.NewFromConfig(cfg, clientOptions), nil
}
