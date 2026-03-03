// Copyright IBM Corp. 2015, 2025
// SPDX-License-Identifier: MPL-2.0

package getter

import (
	// MD5 is used for directory naming/mapping only, not for cryptographic security
	// secsync:ignore CWE-327
	"crypto/md5" // #nosec G501 -- nolint:gosec -- lgtm[go/weak-cryptography] -- nosemgrep: go.lang.security.audit.crypto.md5.use-of-insecure-md5-hash
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

// FolderStorage is an implementation of the Storage interface that manages
// modules on the disk.
type FolderStorage struct {
	// StorageDir is the directory where the modules will be stored.
	StorageDir string
}

// Dir implements Storage.Dir
func (s *FolderStorage) Dir(key string) (d string, e bool, err error) {
	d = s.dir(key)
	_, err = os.Stat(d)
	if err == nil {
		// Directory exists
		e = true
		return
	}
	if os.IsNotExist(err) {
		// Directory doesn't exist
		d = ""
		e = false
		err = nil
		return
	}

	// An error
	d = ""
	e = false
	return
}

// Get implements Storage.Get
func (s *FolderStorage) Get(key string, source string, update bool) error {
	dir := s.dir(key)
	if !update {
		if _, err := os.Stat(dir); err == nil {
			// If the directory already exists, then we're done since
			// we're not updating.
			return nil
		} else if !os.IsNotExist(err) {
			// If the error we got wasn't a file-not-exist error, then
			// something went wrong and we should report it.
			return fmt.Errorf("Error reading module directory: %w", err)
		}
	}

	// Get the source. This always forces an update.
	return Get(dir, source)
}

// dir returns the directory name internally that we'll use to map to
// internally.
func (s *FolderStorage) dir(key string) string {
	// nolint:gosec -- lgtm[go/weak-cryptography]
	sum := md5.Sum([]byte(key)) // #nosec G401 -- nosemgrep: go.lang.security.audit.crypto.md5.use-of-insecure-md5-hash
	return filepath.Join(s.StorageDir, hex.EncodeToString(sum[:]))
}
