//go:build windows
// +build windows

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package getter

import (
	"fmt"
	"os"
	"path/filepath"
	"unsafe"

	"golang.org/x/sys/windows"
)

// resolveSymlinks resolves symlinks with special handling for Windows junction points in Go 1.23+.
// This function provides a more robust fallback for Windows path resolution issues.
func resolveSymlinks(src string) (string, error) {
	resolved, err := filepath.EvalSymlinks(src)
	if err == nil {
		return resolved, nil
	}

	// On Windows, first check if the path actually exists and what type it is
	lstatInfo, lstatErr := os.Lstat(src)
	if lstatErr != nil {
		// Path doesn't exist or can't be accessed, return the original EvalSymlinks error
		return "", err
	}

	statInfo, statErr := os.Stat(src)
	if statErr != nil {
		// Path can't be followed, return the original EvalSymlinks error
		return "", err
	}

	// Path exists but EvalSymlinks failed on Windows
	// Check what type of path this is

	// Case 1: Check if this is a junction point
	var isJunction bool
	var junctionErr error
	if isJunction, junctionErr = isWindowsJunctionPoint(src); junctionErr == nil && isJunction {
		// Confirmed junction point - return cleaned path
		return filepath.Clean(src), nil
	}

	// Case 2: Check if this is a regular directory or file
	isRegularDir := lstatInfo.IsDir() && statInfo.IsDir() &&
		lstatInfo.Mode()&os.ModeIrregular == 0 &&
		lstatInfo.Mode()&os.ModeSymlink == 0

	if isRegularDir {
		// I'm just a dir, yes I'm only a dir,
		// EvalSymlinks tried but it stopped right here,
		// So do what EvalSymlinks would normally do:
		// Return the cleaned path, plain and true.
		return filepath.Clean(src), nil
	}

	// If we get here, it's some other type of special file/link that we don't understand
	// Return the original EvalSymlinks error
	if junctionErr != nil {
		return "", fmt.Errorf("failed to resolve symlinks (checking whether junction point: %s): %w", junctionErr, err)
	}
	return "", fmt.Errorf("failed to resolve symlinks: %w", err)
}

// isWindowsJunctionPoint uses Windows API to reliably detect junction points
// by checking the reparse point tag specifically.
func isWindowsJunctionPoint(path string) (bool, error) {
	// Convert path to UTF16 for Windows API
	pathPtr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return false, err
	}

	// Get file attributes to check if it's a reparse point
	attrs := windows.Win32FileAttributeData{}
	err = windows.GetFileAttributesEx(
		pathPtr,
		windows.GetFileExInfoStandard,
		(*byte)(unsafe.Pointer(&attrs)),
	)
	if err != nil {
		return false, err
	}

	// Check if this is a reparse point
	if attrs.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT == 0 {
		return false, nil
	}

	// Open the file to get reparse point information
	handle, err := windows.CreateFile(
		pathPtr,
		0, // No access needed, just query reparse data
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT,
		0,
	)
	if err != nil {
		return false, err
	}
	defer windows.CloseHandle(handle)

	// Query the reparse point data
	const FSCTL_GET_REPARSE_POINT = 0x900a8
	const MAXIMUM_REPARSE_DATA_BUFFER_SIZE = 16 * 1024

	buffer := make([]byte, MAXIMUM_REPARSE_DATA_BUFFER_SIZE)
	var bytesReturned uint32

	err = windows.DeviceIoControl(
		handle,
		FSCTL_GET_REPARSE_POINT,
		nil,
		0,
		&buffer[0],
		uint32(len(buffer)),
		&bytesReturned,
		nil,
	)
	if err != nil {
		return false, err
	}

	// Parse the reparse tag from the buffer
	// The reparse tag is the first 4 bytes of the reparse data buffer
	if bytesReturned < 4 {
		return false, nil
	}

	reparseTag := *(*uint32)(unsafe.Pointer(&buffer[0]))

	// IO_REPARSE_TAG_MOUNT_POINT indicates a junction point
	const IO_REPARSE_TAG_MOUNT_POINT = 0xA0000003
	// IO_REPARSE_TAG_SYMLINK indicates a symbolic link
	const IO_REPARSE_TAG_SYMLINK = 0xA000000C

	isJunction := reparseTag == IO_REPARSE_TAG_MOUNT_POINT
	return isJunction, nil
}
