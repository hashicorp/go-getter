package getter

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	urlhelper "github.com/hashicorp/go-getter/helper/url"
	"github.com/hashicorp/go-safetemp"
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
	//
	// Pwd is the working directory for detection. If this isn't set, some
	// detection may fail. Client will not default pwd to the current
	// working directory for security reasons.
	Src string
	Dst string
	Pwd string

	// Mode is the method of download the client will use. See ClientMode
	// for documentation.
	Mode ClientMode

	// Detectors is the list of detectors that are tried on the source.
	// If this is nil, then the default Detectors will be used.
	Detectors []Detector

	// Decompressors is the map of decompressors supported by this client.
	// If this is nil, then the default value is the Decompressors global.
	Decompressors map[string]Decompressor

	// Getters is the map of protocols supported by this client. If this
	// is nil, then the default Getters variable will be used.
	Getters map[string]Getter

	// Dir, if true, tells the Client it is downloading a directory (versus
	// a single file). This distinction is necessary since filenames and
	// directory names follow the same format so disambiguating is impossible
	// without knowing ahead of time.
	//
	// WARNING: deprecated. If Mode is set, that will take precedence.
	Dir bool
}

// Get downloads the configured source to the destination.
func (c *Client) Get() error {
	// Store this locally since there are cases we swap this
	mode := c.Mode
	if mode == ClientModeInvalid {
		if c.Dir {
			mode = ClientModeDir
		} else {
			mode = ClientModeFile
		}
	}

	// Default decompressor value
	decompressors := c.Decompressors
	if decompressors == nil {
		decompressors = Decompressors
	}

	// Detect the URL. This is safe if it is already detected.
	detectors := c.Detectors
	if detectors == nil {
		detectors = Detectors
	}
	src, err := Detect(c.Src, c.Pwd, detectors)
	if err != nil {
		return err
	}

	// Determine if we have a forced protocol, i.e. "git::http://..."
	force, src := getForcedGetter(src)

	// If there is a subdir component, then we download the root separately
	// and then copy over the proper subdir.
	var realDst string
	dst := c.Dst
	src, subDir := SourceDirSubdir(src)
	if subDir != "" {
		td, tdcloser, err := safetemp.Dir("", "getter")
		if err != nil {
			return err
		}
		defer tdcloser.Close()

		realDst = dst
		dst = td
	}

	u, err := urlhelper.Parse(src)
	if err != nil {
		return err
	}
	if force == "" {
		force = u.Scheme
	}

	getters := c.Getters
	if getters == nil {
		getters = Getters
	}

	g, ok := getters[force]
	if !ok {
		return fmt.Errorf(
			"download not supported for scheme '%s'", force)
	}

	// We have magic query parameters that we use to signal different features
	q := u.Query()

	// Determine if we have an archive type
	archiveV := q.Get("archive")
	if archiveV != "" {
		// Delete the paramter since it is a magic parameter we don't
		// want to pass on to the Getter
		q.Del("archive")
		u.RawQuery = q.Encode()

		// If we can parse the value as a bool and it is false, then
		// set the archive to "-" which should never map to a decompressor
		if b, err := strconv.ParseBool(archiveV); err == nil && !b {
			archiveV = "-"
		}
	}
	if archiveV == "" {
		// We don't appear to... but is it part of the filename?
		matchingLen := 0
		for k, _ := range decompressors {
			if strings.HasSuffix(u.Path, "."+k) && len(k) > matchingLen {
				archiveV = k
				matchingLen = len(k)
			}
		}
	}

	// If we have a decompressor, then we need to change the destination
	// to download to a temporary path. We unarchive this into the final,
	// real path.
	var decompressDst string
	var decompressDir bool
	decompressor := decompressors[archiveV]
	if decompressor != nil {
		// Create a temporary directory to store our archive. We delete
		// this at the end of everything.
		td, err := ioutil.TempDir("", "getter")
		if err != nil {
			return fmt.Errorf(
				"Error creating temporary directory for archive: %s", err)
		}
		defer os.RemoveAll(td)

		// Swap the download directory to be our temporary path and
		// store the old values.
		decompressDst = dst
		decompressDir = mode != ClientModeFile
		dst = filepath.Join(td, "archive")
		mode = ClientModeFile
	}

	// Determine checksum if we have one
	checksum, err := extractChecksum(u)
	if err != nil {
		return fmt.Errorf("invalid checksum: %s", err)
	}

	// Delete the query parameter if we have it.
	q.Del("checksum")
	u.RawQuery = q.Encode()

	if mode == ClientModeAny {
		// Ask the getter which client mode to use
		mode, err = g.ClientMode(u)
		if err != nil {
			return err
		}

		// Destination is the base name of the URL path in "any" mode when
		// a file source is detected.
		if mode == ClientModeFile {
			filename := filepath.Base(u.Path)

			// Determine if we have a custom file name
			if v := q.Get("filename"); v != "" {
				// Delete the query parameter if we have it.
				q.Del("filename")
				u.RawQuery = q.Encode()

				filename = v
			}

			dst = filepath.Join(dst, filename)
		}
	}

	// If we're not downloading a directory, then just download the file
	// and return.
	if mode == ClientModeFile {
		getFile := true
		if checksum != nil {
			if err := checksum.checksum(dst); err == nil {
				// don't get the file if the checksum of dst is correct
				getFile = false
			}
		}
		if getFile {
			err := g.GetFile(dst, u)
			if err != nil {
				return err
			}

			if checksum != nil {
				if err := checksum.checksum(dst); err != nil {
					return err
				}
			}
		}

		if decompressor != nil {
			// We have a decompressor, so decompress the current destination
			// into the final destination with the proper mode.
			err := decompressor.Decompress(decompressDst, dst, decompressDir)
			if err != nil {
				return err
			}

			// Swap the information back
			dst = decompressDst
			if decompressDir {
				mode = ClientModeAny
			} else {
				mode = ClientModeFile
			}
		}

		// We check the dir value again because it can be switched back
		// if we were unarchiving. If we're still only Get-ing a file, then
		// we're done.
		if mode == ClientModeFile {
			return nil
		}
	}

	// If we're at this point we're either downloading a directory or we've
	// downloaded and unarchived a directory and we're just checking subdir.
	// In the case we have a decompressor we don't Get because it was Get
	// above.
	if decompressor == nil {
		// If we're getting a directory, then this is an error. You cannot
		// checksum a directory. TODO: test
		if checksum != nil {
			return fmt.Errorf(
				"checksum cannot be specified for directory download")
		}

		// We're downloading a directory, which might require a bit more work
		// if we're specifying a subdir.
		err := g.Get(dst, u)
		if err != nil {
			err = fmt.Errorf("error downloading '%s': %s", src, err)
			return err
		}
	}

	// If we have a subdir, copy that over
	if subDir != "" {
		if err := os.RemoveAll(realDst); err != nil {
			return err
		}
		if err := os.MkdirAll(realDst, 0755); err != nil {
			return err
		}

		// Process any globs
		subDir, err := SubdirGlob(dst, subDir)
		if err != nil {
			return err
		}

		return copyDir(realDst, subDir, false)
	}

	return nil
}

// checksum is a simple method to compute the checksum of a source file
// and compare it to the given expected value.
func (c *fileChecksum) checksum(source string) error {
	f, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("Failed to open file for checksum: %s", err)
	}
	defer f.Close()

	if _, err := io.Copy(c.Hash, f); err != nil {
		return fmt.Errorf("Failed to hash: %s", err)
	}

	if actual := c.Hash.Sum(nil); !bytes.Equal(actual, c.Value) {
		return fmt.Errorf(
			"Checksums did not match.\nExpected: %s\nGot: %s",
			hex.EncodeToString(c.Value),
			hex.EncodeToString(actual))
	}

	return nil
}
