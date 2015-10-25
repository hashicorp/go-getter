package getter

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	urlhelper "github.com/hashicorp/terraform/helper/url"
)

// Client is a client for downloading things.
//
// Top-level functions such as Get are shortcuts for interacting with a client.
// Using a client directly allows more fine-grained control over how downloading
// is done, as well as customizing the protocols supported.
type Client struct {
	// Src is the source URL to get.
	//
	// Dst is the path to save the downloaded thing as. If Dir is set to
	// true, then this should be a directory. If the directory doesn't exist,
	// it will be created for you.
	Src string
	Dst string

	// Dir, if true, tells the Client it is downloading a directory (versus
	// a single file). This distinction is necessary since filenames and
	// directory names follow the same format so disambiguating is impossible
	// without knowing ahead of time.
	Dir bool

	// Getters is the map of protocols supported by this client. Use
	// the global Getters variable for the built-in defaults.
	Getters map[string]Getter
}

// Get downloads the configured source to the destination.
func (c *Client) Get() error {
	force, src := getForcedGetter(c.Src)

	// If there is a subdir component, then we download the root separately
	// and then copy over the proper subdir.
	var realDst string
	dst := c.Dst
	src, subDir := SourceDirSubdir(src)
	if subDir != "" {
		tmpDir, err := ioutil.TempDir("", "tf")
		if err != nil {
			return err
		}
		if err := os.RemoveAll(tmpDir); err != nil {
			return err
		}
		defer os.RemoveAll(tmpDir)

		realDst = dst
		dst = tmpDir
	}

	u, err := urlhelper.Parse(src)
	if err != nil {
		return err
	}
	if force == "" {
		force = u.Scheme
	}

	g, ok := c.Getters[force]
	if !ok {
		return fmt.Errorf(
			"download not supported for scheme '%s'", force)
	}

	// Determine if we have a checksum
	var checksumHash hash.Hash
	var checksumValue []byte
	if v := u.Query().Get("checksum"); v != "" {
		// Delete the query parameter if we have it.
		u.Query().Del("checksum")

		// If we're getting a directory, then this is an error. You cannot
		// checksum a directory. TODO: test
		if c.Dir {
			return fmt.Errorf(
				"checksum cannot be specified for directory download")
		}

		// Determine the checksum hash type
		checksumType := ""
		idx := strings.Index(v, ":")
		if idx > -1 {
			checksumType = v[:idx]
		}
		switch checksumType {
		case "md5":
			checksumHash = md5.New()
		case "sha1":
			checksumHash = sha1.New()
		case "sha256":
			checksumHash = sha256.New()
		case "sha512":
			checksumHash = sha512.New()
		default:
			return fmt.Errorf(
				"unsupported checksum type: %s", checksumType)
		}

		// Get the remainder of the value and parse it into bytes
		b, err := hex.DecodeString(v[idx+1:])
		if err != nil {
			return fmt.Errorf("invalid checksum: %s", err)
		}

		// Set our value
		checksumValue = b
	}

	// If we're not downloading a directory, then just download the file
	// and return.
	if !c.Dir {
		err := g.GetFile(dst, u)
		if err != nil {
			return err
		}

		if checksumHash != nil {
			return checksum(dst, checksumHash, checksumValue)
		}

		return nil
	}

	// We're downloading a directory, which might require a bit more work
	// if we're specifying a subdir.
	err = g.Get(dst, u)
	if err != nil {
		err = fmt.Errorf("error downloading '%s': %s", src, err)
		return err
	}

	// If we have a subdir, copy that over
	if subDir != "" {
		if err := os.RemoveAll(realDst); err != nil {
			return err
		}
		if err := os.MkdirAll(realDst, 0755); err != nil {
			return err
		}

		return copyDir(realDst, filepath.Join(dst, subDir), false)
	}

	return nil
}

// checksum is a simple method to compute the checksum of a source file
// and compare it to the given expected value.
func checksum(source string, h hash.Hash, v []byte) error {
	f, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("Failed to open file for checksum: %s", err)
	}
	defer f.Close()

	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("Failed to hash: %s", err)
	}

	if actual := h.Sum(nil); !bytes.Equal(actual, v) {
		return fmt.Errorf(
			"Checksums did not match.\nExpected (%s), got (%s)",
			hex.EncodeToString(v),
			hex.EncodeToString(actual))
	}

	return nil
}
