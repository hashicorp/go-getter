package getter

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

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

	var sum string
	sumRaw := u.Query()["checksum"]
	if len(sumRaw) == 1 {
		sum = sumRaw[0]
	}

	// If we're not downloading a directory, then just download the file
	// and return.
	if !c.Dir {
		if sum == "" {
			return g.GetFile(dst, u)
		}

		err := g.GetFile(dst, u)
		if err != nil {
			return err
		}

		return checksum(dst, sum)
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

	return checksum(realDst, sum)
}

// checksum is a simple method to compute the SHA256 checksum of a source (file
// or dir) and compare it to a given sum.
func checksum(source, sum string) error {
	if sum == "" {
		return nil
	}
	// compute and check checksum
	log.Printf("[DEBUG] Running checksum on (%s)", source)
	hasher := sha256.New()
	file, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("Failed to open file for checksum: %s", err)
	}

	defer file.Close()
	io.Copy(hasher, file)

	computed := hex.EncodeToString(hasher.Sum(nil))
	if sum != computed {
		return fmt.Errorf(
			"Checksums did not match.\nExpected (%s), got (%s)",
			sum,
			computed)
	}

	return nil
}
