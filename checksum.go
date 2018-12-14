package getter

import (
	"bufio"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	urlhelper "github.com/hashicorp/go-getter/helper/url"
)

// checksum helps verifying the checksum for
// a file.
type fileChecksum struct {
	Type     string
	Hash     hash.Hash
	Value    []byte
	Filename string
}

func newChecksum(checksumValue, filename string) (*fileChecksum, error) {
	c := &fileChecksum{
		Filename: filename,
	}
	var err error
	c.Value, err = hex.DecodeString(checksumValue)
	if err != nil {
		return nil, fmt.Errorf("invalid checksum: %s", err)
	}
	return c, nil
}

func newChecksumFromType(checksumType, checksumValue, filename string) (*fileChecksum, error) {
	c, err := newChecksum(checksumValue, filename)
	if err != nil {
		return nil, err
	}

	c.Type = strings.ToLower(checksumType)
	switch c.Type {
	case "md5":
		c.Hash = md5.New()
	case "sha1":
		c.Hash = sha1.New()
	case "sha256":
		c.Hash = sha256.New()
	case "sha512":
		c.Hash = sha512.New()
	default:
		return nil, fmt.Errorf(
			"unsupported checksum type: %s", checksumType)
	}

	return c, nil
}

func newChecksumFromValue(checksumValue, filename string) (*fileChecksum, error) {
	c, err := newChecksum(checksumValue, filename)
	if err != nil {
		return nil, err
	}

	switch len(c.Value) {
	case md5.Size:
		c.Hash = md5.New()
		c.Type = "md5"
	case sha1.Size:
		c.Hash = sha1.New()
		c.Type = "sha1"
	case sha256.Size:
		c.Hash = sha256.New()
		c.Type = "sha256"
	case sha512.Size:
		c.Hash = sha512.New()
		c.Type = "sha512"
	default:
		return nil, fmt.Errorf("Unknown type for checksum %s", checksumValue)
	}

	return c, nil
}

// checksumHashAndValue will return checksum based on checksum parameter of u
// ex:
//  http://hashicorp.com/terraform?checksum=<checksumType>:<checksumValue>
//  http://hashicorp.com/terraform?checksum=file:<checksum_url>
// when checksumming from a file checksumHashAndValue will go get checksum_url
// in a temporary directory and parse the content of the file.
// Content of files are expected to be BSD style or GNU style.
//
// BSD-style checksum:
//  MD5 (file1) = <checksum>
//  MD5 (file2) = <checksum>
//
// GNU-style:
//  <checksum>  file1
//  <checksum> *file2
//
// see parseChecksumLine for more detail.
func checksumHashAndValue(u *url.URL) (*fileChecksum, error) {
	q := u.Query()
	v := q.Get("checksum")

	if v == "" {
		return nil, nil
	}

	vs := strings.SplitN(v, ":", 2)
	if len(vs) != 2 {
		return nil, fmt.Errorf("non explicit checksum type: %s", v)
	}

	checksumType, checksumValue := vs[0], vs[1]

	switch checksumType {
	case "file":
		return checksumFromFile(checksumValue, u)
	default:
		return newChecksumFromType(checksumType, checksumValue, filepath.Base(u.EscapedPath()))
	}
}

// checksumsFromFile will download checksumFile that is a checksum for file
// behind src.
//
// checksumsFromFile will try to guess the hashing algorithm based on content
// of checksum file
//
// checksumsFromFile will only return checksums for files that match file
// behind src
func checksumFromFile(checksumFile string, src *url.URL) (*fileChecksum, error) {
	checksumFileURL, err := urlhelper.Parse(checksumFile)
	if err != nil {
		return nil, err
	}

	tempfile, err := tmpFile("", filepath.Base(checksumFileURL.Path))
	if err != nil {
		return nil, err
	}
	defer func() {
		os.Remove(tempfile)
	}()

	if err = GetFile(tempfile, checksumFile); err != nil {
		return nil, fmt.Errorf(
			"Error downloading checksum file: %s", err)
	}

	filename := filepath.Base(src.Path)
	absPath, _ := filepath.Abs(src.Path)
	relpath, _ := filepath.Rel(filepath.Dir(checksumFileURL.Path), absPath)

	// possible file identifiers:
	options := []string{
		filename,       // ubuntu-14.04.1-server-amd64.iso
		"*" + filename, // *ubuntu-14.04.1-server-amd64.iso  Standard checksum
		"?" + filename, // ?ubuntu-14.04.1-server-amd64.iso  shasum -p
		relpath,        // dir/ubuntu-14.04.1-server-amd64.iso
		"./" + relpath, // ./dir/ubuntu-14.04.1-server-amd64.iso
		absPath,        // fullpath; set if local
	}

	f, err := os.Open(tempfile)
	if err != nil {
		return nil, fmt.Errorf(
			"Error opening downloaded file: %s", err)
	}
	defer f.Close()
	rd := bufio.NewReader(f)
	for {
		line, err := rd.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				return nil, fmt.Errorf(
					"Error reading checksum file: %s", err)
			}
			break
		}
		checksum, err := parseChecksumLine(line)
		if err != nil {
			continue
		}
		// make sure the checksum is for the right file
		for _, option := range options {
			if checksum.Filename == option {
				// any checksum will work so we return the first one
				return checksum, nil
			}
		}
	}
	return nil, fmt.Errorf("no checksum found in: %s", checksumFile)
}

// parseChecksumLine takes a line from a checksum file and returns
// checksumType, checksumValue and filename parseChecksumLine guesses the style
// of the checksum BSD vs GNU by splitting the line and by counting the parts.
// of a line.
// for BSD type sums parseChecksumLine guesses the hashing algorithm
// by checking the length of the checksum.
func parseChecksumLine(line string) (*fileChecksum, error) {
	parts := strings.Fields(line)

	switch len(parts) {
	case 4:
		// BSD-style checksum:
		//  MD5 (file1) = <checksum>
		//  MD5 (file2) = <checksum>
		if len(parts[1]) <= 2 ||
			parts[1][0] != '(' || parts[1][len(parts[1])-1] != ')' {
			return nil, fmt.Errorf(
				"Unexpected BSD-style-checksum filename format: %s", line)
		}
		filename := parts[1][1 : len(parts[1])-1]
		return newChecksumFromType(parts[0], parts[3], filename)
	case 2:
		// GNU-style:
		//  <checksum>  file1
		//  <checksum> *file2
		return newChecksumFromValue(parts[0], parts[1])
	default:
		return nil, fmt.Errorf("Unexpected checksum line format: %s", line)
	}
}
