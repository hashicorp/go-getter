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
	return fmt.Errorf("Operation is unsupported")
}

func (g *S3Getter) GetFile(dst string, u *url.URL) error {
	region, bucket, path, err := g.parseUrl(u)
	if err != nil {
		return err
	}

	client := s3.New(&aws.Config{Region: aws.String(region)})
	resp, err := client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(path),
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
