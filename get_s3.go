package getter

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

// S3Getter is a Getter implementation that will download a module from
// a S3 bucket.
type S3Getter struct{}

func (g *S3Getter) Get(dst string, u *url.URL) error {
	// Parse URL
	region, bucket, path, err := g.parseUrl(u)
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

	client := s3.New(&aws.Config{Region: aws.String(region)})

	// List files in path, keep listing until no more objects are found
	lastMarker := ""
	hasMore := true
	for hasMore {
		req := &s3.ListObjectsInput{
			Bucket: aws.String(bucket),
			Prefix: aws.String(path),
		}

		if lastMarker != "" {
			req.Marker = aws.String(lastMarker)
		}

		resp, err := client.ListObjects(req)
		if err != nil {
			return err
		}

		hasMore = aws.BoolValue(resp.IsTruncated)

		for _, object := range resp.Contents {
			lastMarker = aws.StringValue(object.Key)
			objPath := aws.StringValue(object.Key)

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

			if err := g.getObject(client, objDst, bucket, objPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func (g *S3Getter) GetFile(dst string, u *url.URL) error {
	region, bucket, path, err := g.parseUrl(u)
	if err != nil {
		return err
	}

	client := s3.New(&aws.Config{Region: aws.String(region)})

	return g.getObject(client, dst, bucket, path)
}

func (g *S3Getter) getObject(client *s3.S3, dst, bucket, key string) error {
	resp, err := client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
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

	_, err = io.Copy(f, resp.Body)
	return err
}

func (g *S3Getter) parseUrl(u *url.URL) (region, bucket, path string, err error) {
	hostParts := strings.Split(u.Host, ".")

	if len(hostParts) != 3 {
		return "", "", "", fmt.Errorf("URL is not a valid S3 URL")
	}
	region = strings.TrimPrefix(strings.TrimPrefix(hostParts[0], "s3-"), "s3")

	pathParts := strings.Split(u.Path, "/")
	bucket = pathParts[1]
	path = strings.Join(pathParts[2:], "/")

	return
}
