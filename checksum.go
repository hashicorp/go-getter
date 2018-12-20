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
	"os"
	"strings"
)

// fileChecksum helps verifying the checksum for a file.
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

// checksumsFromFile will download checksumFile that is a checksum for file
// behind src.
//
// checksumsFromFile will try to guess the hashing algorithm based on content
// of checksum file
//
// checksumsFromFile will only return checksums for files that match file
// behind src
func checksumsFromFile(file string) ([]*fileChecksum, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("Error opening checksum file: %s", err)
	}
	defer f.Close()
	rd := bufio.NewReader(f)
	res := []*fileChecksum{}
	for {
		line, err := rd.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				return nil, fmt.Errorf("Error reading checksum file: %s", err)
			}
			break
		}
		checksum, err := parseChecksumLine(line)
		if err != nil || checksum == nil {
			continue
		}
		res = append(res, checksum)
	}
	if len(res) == 0 {
		return nil, fmt.Errorf("no checksum found in: %s", file)
	}
	return res, nil
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
	case 0:
		return nil, nil // empty line
	default:
		return newChecksumFromValue(parts[0], "")
	}
}
