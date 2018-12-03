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
	"io/ioutil"
	"net/url"
	"path/filepath"
	"strings"

	urlhelper "github.com/hashicorp/go-getter/helper/url"
)

var checksummers = map[string]func() hash.Hash{
	"md5":    md5.New,
	"sha1":   sha1.New,
	"sha256": sha256.New,
	"sha512": sha512.New,
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
// For GNU-style checksum files; it is very common that the hashing algorithm identifier
// is in the filename; so the name of every supported hashing algorithm is compared
// against checksum_url for a match/guess.
// In case a different hashing algorithm is in the filename of checksum_url
// it is recommended to explicitly set hashing algorithm instead.
func checksumHashAndValue(u *url.URL) (checksumHash hash.Hash, checksumValue []byte, err error) {
	q := u.Query()
	v := q.Get("checksum")

	if v == "" {
		return nil, nil, nil
	}

	// Determine the checksum hash type
	checksumType := ""
	idx := len(v)
	if i := strings.Index(v, ":"); i > -1 {
		idx = i
	}
	checksumType = v[:idx]
	if fn, found := checksummers[checksumType]; found {
		checksumHash = fn()
		// Get the remainder of the value and parse it into bytes
		checksumValue, err = hex.DecodeString(v[idx+1:])
		return
	}

	if checksumType != "file" {
		return nil, nil, fmt.Errorf(
			"unsupported checksum type: %s", checksumType)
	}
	file := v[idx+1:]

	checkSums, err := checksumsFromFile(file, u)
	if err != nil {
		return nil, nil, err
	}

	var checksumValueString string
	for checksumType, checksumValueString = range checkSums {
		if fn, found := checksummers[checksumType]; found {
			checksumHash = fn()
			checksumValue, err = hex.DecodeString(checksumValueString)
			return checksumHash, checksumValue, err
		}
	}

	return nil, nil, fmt.Errorf(
		"Could not find/guess checksum in %s: %s", file, checksumType)
}

// checksumsFromFile will download checksumFile that is a checksum for file
// behind src.
//
// checksumsFromFile will try to guess the hashing algorithm based on content
// of or name of checksum file
func checksumsFromFile(checksumFile string, src *url.URL) (checkSums map[string]string, err error) {

	checksumFileURL, err := urlhelper.Parse(checksumFile)
	if err != nil {
		return nil, err
	}

	f, err := ioutil.TempFile("", filepath.Base(checksumFileURL.Path))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if err = GetFile(f.Name(), checksumFile); err != nil {
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
		relpath,        // dir/ubuntu-14.04.1-server-amd64.iso
		"./" + relpath, // ./dir/ubuntu-14.04.1-server-amd64.iso
		absPath,        // fullpath; set if local
	}

	rd := bufio.NewReader(f)
	res := map[string]string{}
	for {

		line, err := rd.ReadString('\n')
		if err != nil && line == "" {
			break
		}
		checksumType, checksumValue, filename, ok := parseChecksumLine(checksumFile, line)
		if !ok {
			continue
		}
		for _, option := range options {
			// filename matches src ?
			if filename == option {
				res[checksumType] = checksumValue
				break
			}
		}

	}
	if len(res) == 0 {
		err = fmt.Errorf("Could not find a checksum for %s in %s", filename, checksumFile)
	}
	return res, err
}

// parseChecksumLine takes a line from a checksum file and returns
// checksumType, checksumValue and filename
// For GNU-style entries parseChecksumLine will try to guess checksumType
// based on checksumFile which should usually contain the checksum type.
// checksumType will be lowercase.
//
// BSD-style checksum:
//  MD5 (file1) = <checksum>
//  MD5 (file2) = <checksum>
//
// GNU-style:
//  <checksum>  file1
//  <checksum> *file2
func parseChecksumLine(checksumFilename, line string) (checksumType, checksumValue, filename string, ok bool) {
	parts := strings.Fields(line)

	ok = true
	switch len(parts) {
	case 4: // BSD-style
		if len(parts[1]) <= 2 || parts[1][0] != '(' || parts[1][len(parts[1])-1] != ')' {
			return "", "", "", false
		}
		checksumType = strings.ToLower(parts[0])
		filename = parts[1][1 : len(parts[1])-1]
		checksumValue = parts[3]
	case 2: // GNU-style

		for ctype := range checksummers {
			// guess checksum type from filename
			if strings.Contains(strings.ToLower(checksumFilename), ctype) {
				checksumType = ctype
			}
		}
		checksumValue = parts[0]
		filename = parts[1]
	default:
		ok = false
	}

	return
}
