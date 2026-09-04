// Copyright IBM Corp. 2015, 2026
// SPDX-License-Identifier: MPL-2.0

package version

import (
	"bytes"
	"fmt"
)

var (
	Version           = "1.8.10"
	VersionPrerelease = "dev"
	VersionMetadata   = ""
	// PluginVersion removed to avoid import cycle
)

func FullVersionString() string {
	var versionString bytes.Buffer

	fmt.Fprintf(&versionString, "Go-Getter v%s", Version)
	if VersionPrerelease != "" {
		fmt.Fprintf(&versionString, "-%s", VersionPrerelease)
	}

	if VersionMetadata != "" {
		fmt.Fprintf(&versionString, "+%s", VersionMetadata)
	}

	return versionString.String()
}
