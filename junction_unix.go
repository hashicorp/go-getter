//go:build !windows
// +build !windows

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package getter

// isWindowsJunctionPoint is a no-op on non-Windows platforms
func isWindowsJunctionPoint(path string) (bool, error) {
	return false, nil
}
