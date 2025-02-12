// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package getter

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
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Getter is a Getter implementation that will download a module from
// a S3 bucket.
type S3Getter struct {
	getter

	// Timeout sets a deadline which all S3 operations should
	// complete within.
	//
	// The zero value means timeout.
	Timeout time.Duration
}

func (g *S3Getter) ClientMode(u *url.URL) (ClientMode, error) {
	// Parse URL
	ctx := g.Context()

	if g.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, g.Timeout)
		defer cancel()
	}

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
	resp, err := client.ListObjectsV2(ctx, req)
	if err != nil {
		return 0, err
	}

	for _, o := range resp.Contents {
		// Use file mode on exact match.
		if *o.Key == path {
			return ClientModeFile, nil
		}

		// Use dir mode if child keys are found.
		if strings.HasPrefix(*o.Key, path+"/") {
			return ClientModeDir, nil
		}
	}

	// There was no match, so just return file mode. The download is going
	// to fail but we will let S3 return the proper error later.
	return ClientModeFile, nil
}

func (g *S3Getter) Get(dst string, u *url.URL) error {
	ctx := g.Context()

	if g.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, g.Timeout)
		defer cancel()
	}

	// Parse URL
	region, bucket, path, _, creds, err := g.parseUrl(u)
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

	client, err := g.newS3Client(ctx, region, u, creds)
	if err != nil {
		return err
	}

	// List files in path, keep listing until no more objects are found
	continuationToken := ""
	hasMore := true
	for hasMore {
		req := &s3.ListObjectsV2Input{
			Bucket: aws.String(bucket),
			Prefix: aws.String(path),
		}

		if continuationToken != "" {
			req.ContinuationToken = aws.String(continuationToken)
		}

		resp, err := client.ListObjectsV2(ctx, req)
		if err != nil {
			return err
		}

		hasMore = aws.ToBool(resp.IsTruncated)

		// Get each object storing each file relative to the destination path
		for _, object := range resp.Contents {
			continuationToken = aws.ToString(resp.NextContinuationToken)
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
			objDst = filepath.Join(dst, objDst)

			if err := g.getObject(ctx, client, objDst, bucket, objPath, ""); err != nil {
				return err
			}
		}
	}

	return nil
}

func (g *S3Getter) GetFile(dst string, u *url.URL) error {
	ctx := g.Context()

	if g.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, g.Timeout)
		defer cancel()
	}

	region, bucket, path, version, creds, err := g.parseUrl(u)
	if err != nil {
		return err
	}

	client, err := g.newS3Client(ctx, region, u, creds)
	if err != nil {
		return err
	}

	return g.getObject(ctx, client, dst, bucket, path, version)
}

func (g *S3Getter) getObject(ctx context.Context, client *s3.Client, dst, bucket, key, version string) error {
	req := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	if version != "" {
		req.VersionId = aws.String(version)
	}

	resp, err := client.GetObject(ctx, req)
	if err != nil {
		return err
	}

	// Create all the parent directories
	if err := os.MkdirAll(filepath.Dir(dst), g.client.mode(0755)); err != nil {
		return err
	}

	body := resp.Body

	if g.client != nil && g.client.ProgressListener != nil {
		fn := filepath.Base(key)
		body = g.client.ProgressListener.TrackProgress(fn, 0, *resp.ContentLength, resp.Body)
	}
	defer body.Close()

	// There is no limit set for the size of an object from S3
	return copyReader(dst, body, 0666, g.client.umask(), 0)
}

func (g *S3Getter) parseUrl(u *url.URL) (region, bucket, path, version string, creds *credentials.StaticCredentialsProvider, err error) {
	// This just check whether we are dealing with S3 or
	// any other S3 compliant service. S3 has a predictable
	// url as others do not
	if strings.HasSuffix(u.Host, ".amazonaws.com") {
		// Amazon S3 supports both virtual-hosted–style and path-style URLs to access a bucket, although path-style is deprecated
		// In both cases few older regions supports dash-style region indication (s3-Region) even if AWS discourages their use.
		// The same bucket could be reached with:
		// bucket.s3.region.amazonaws.com/path
		// bucket.s3-region.amazonaws.com/path
		// s3.amazonaws.com/bucket/path
		// s3-region.amazonaws.com/bucket/path

		hostParts := strings.Split(u.Host, ".")
		switch len(hostParts) {
		// path-style
		case 3:
			// Parse the region out of the first part of the host
			region = strings.TrimPrefix(strings.TrimPrefix(hostParts[0], "s3-"), "s3")
			if region == "" {
				region = "us-east-1"
			}
			pathParts := strings.SplitN(u.Path, "/", 3)
			if len(pathParts) < 3 {
				err = fmt.Errorf("URL is not a valid S3 URL")
				return
			}
			bucket = pathParts[1]
			path = pathParts[2]
		// vhost-style, dash region indication
		case 4:
			// Parse the region out of the second part of the host
			region = strings.TrimPrefix(strings.TrimPrefix(hostParts[1], "s3-"), "s3")
			if region == "" {
				err = fmt.Errorf("URL is not a valid S3 URL")
				return
			}
			pathParts := strings.SplitN(u.Path, "/", 2)
			if len(pathParts) < 2 {
				err = fmt.Errorf("URL is not a valid S3 URL")
				return
			}
			bucket = hostParts[0]
			path = pathParts[1]
		//vhost-style, dot region indication
		case 5:
			region = hostParts[2]
			pathParts := strings.SplitN(u.Path, "/", 2)
			if len(pathParts) < 2 {
				err = fmt.Errorf("URL is not a valid S3 URL")
				return
			}
			bucket = hostParts[0]
			path = pathParts[1]

		}
		if len(hostParts) < 3 || len(hostParts) > 5 {
			err = fmt.Errorf("URL is not a valid S3 URL")
			return
		}
		version = u.Query().Get("version")

	} else {
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
		credentialProvider := credentials.NewStaticCredentialsProvider(
			u.Query().Get("aws_access_key_id"),
			u.Query().Get("aws_access_key_secret"),
			u.Query().Get("aws_access_token"),
		)
		creds = &credentialProvider
	}

	return
}

func (g *S3Getter) newS3Client(ctx context.Context, region string, url *url.URL, creds *credentials.StaticCredentialsProvider) (*s3.Client, error) {
	var cfg aws.Config
	var err error

	// We first check if the AWS_METADATA_URL is set, if it is we use it to load the config.
	if creds == nil && os.Getenv("AWS_METADATA_URL") != "" {
		cfg, err = config.LoadDefaultConfig(
			ctx,
			config.WithRegion(region),
			config.WithEC2IMDSEndpoint(os.Getenv("AWS_METADATA_URL")),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to load AWS config with metadata URL override: %w", err)
		}
		return s3.NewFromConfig(cfg), nil
	}

	// Otherwise, we load the config with the default settings.
	cfg, err = config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// If a credential provider is provided, override the default credentials;
	// otherwise, let the SDK automatically resolve them.
	if creds != nil {
		cfg.Credentials = creds
	}

	// Create the client
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if creds != nil {
			o.UsePathStyle = true
			o.BaseEndpoint = &url.Host

			if url.Scheme == "http" {
				o.EndpointOptions.DisableHTTPS = true
			}
		}
	})

	return client, nil
}
