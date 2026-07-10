// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package getter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func tmpFile(dir, pattern string) (string, error) {
	f, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return "", err
	}
	_ = f.Close()
	return f.Name(), nil
}

func objectDestination(dst, prefix, key string) (string, error) {
	rel, err := filepath.Rel(prefix, key)
	if err != nil {
		return "", err
	}
	if filepath.IsAbs(rel) || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("object key %q escapes prefix %q", key, prefix)
	}
	return filepath.Join(dst, rel), nil
}
