package getter

import (
	"fmt"
	"io/ioutil"
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
	//
	// Pwd is the working directory for detection. If this isn't set, some
	// detection may fail. Client will not default pwd to the current
	// working directory for security reasons.
	Src string
	Dst string
	Pwd string

	// Dir, if true, tells the Client it is downloading a directory (versus
	// a single file). This distinction is necessary since filenames and
	// directory names follow the same format so disambiguating is impossible
	// without knowing ahead of time.
	Dir bool

	// Detectors is the list of detectors that are tried on the source.
	// If this is nil, then the default Detectors will be used.
	Detectors []Detector

	// Getters is the map of protocols supported by this client. If this
	// is nil, then the default Getters variable will be used.
	Getters map[string]Getter
}

// Get downloads the configured source to the destination.
func (c *Client) Get() error {
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

	getters := c.Getters
	if getters == nil {
		getters = Getters
	}

	g, ok := getters[force]
	if !ok {
		return fmt.Errorf(
			"download not supported for scheme '%s'", force)
	}

	// If we're not downloading a directory, then just download the file
	// and return.
	if !c.Dir {
		return g.GetFile(dst, u)
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
