//go:build windows
// +build windows

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package getter

import (
	"fmt"
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

// isWindowsJunctionPoint uses Windows API to reliably detect junction points
// by checking the reparse point tag specifically.
func isWindowsJunctionPoint(path string) (bool, error) {
	// DEBUG: Comprehensive debugging output
	fmt.Printf("=== JUNCTION DEBUG START ===\n")
	fmt.Printf("DEBUG: Checking path: %s\n", path)

	// Check basic file info first
	if fi, err := os.Lstat(path); err != nil {
		fmt.Printf("DEBUG: os.Lstat failed: %v\n", err)
	} else {
		fmt.Printf("DEBUG: os.Lstat - IsDir: %v, Mode: %v, Size: %d\n", fi.IsDir(), fi.Mode(), fi.Size())
		fmt.Printf("DEBUG: os.Lstat - ModeIrregular: %v\n", fi.Mode()&os.ModeIrregular != 0)
		fmt.Printf("DEBUG: os.Lstat - ModeSymlink: %v\n", fi.Mode()&os.ModeSymlink != 0)
	}

	if fi, err := os.Stat(path); err != nil {
		fmt.Printf("DEBUG: os.Stat failed: %v\n", err)
	} else {
		fmt.Printf("DEBUG: os.Stat - IsDir: %v, Mode: %v, Size: %d\n", fi.IsDir(), fi.Mode(), fi.Size())
	}

	// Convert path to UTF16 for Windows API
	pathPtr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		fmt.Printf("DEBUG: UTF16PtrFromString failed: %v\n", err)
		fmt.Printf("=== JUNCTION DEBUG END ===\n")
		return false, err
	}
	fmt.Printf("DEBUG: UTF16 conversion successful\n")

	// Get file attributes to check if it's a reparse point
	attrs := windows.Win32FileAttributeData{}
	err = windows.GetFileAttributesEx(
		pathPtr,
		windows.GetFileExInfoStandard,
		(*byte)(unsafe.Pointer(&attrs)),
	)
	if err != nil {
		fmt.Printf("DEBUG: GetFileAttributesEx failed: %v\n", err)
		fmt.Printf("=== JUNCTION DEBUG END ===\n")
		return false, err
	}

	fmt.Printf("DEBUG: FileAttributes: 0x%x\n", attrs.FileAttributes)
	fmt.Printf("DEBUG: FILE_ATTRIBUTE_REPARSE_POINT bit: %v\n", attrs.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0)
	fmt.Printf("DEBUG: FILE_ATTRIBUTE_DIRECTORY bit: %v\n", attrs.FileAttributes&windows.FILE_ATTRIBUTE_DIRECTORY != 0)

	// Check if this is a reparse point
	if attrs.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT == 0 {
		fmt.Printf("DEBUG: Not a reparse point - returning false\n")
		fmt.Printf("=== JUNCTION DEBUG END ===\n")
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
		fmt.Printf("DEBUG: CreateFile failed: %v\n", err)
		fmt.Printf("=== JUNCTION DEBUG END ===\n")
		return false, err
	}
	defer windows.CloseHandle(handle)
	fmt.Printf("DEBUG: CreateFile successful, handle: %v\n", handle)

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
		fmt.Printf("DEBUG: DeviceIoControl failed: %v\n", err)
		fmt.Printf("=== JUNCTION DEBUG END ===\n")
		return false, err
	}

	fmt.Printf("DEBUG: DeviceIoControl successful, bytes returned: %d\n", bytesReturned)

	// Parse the reparse tag from the buffer
	// The reparse tag is the first 4 bytes of the reparse data buffer
	if bytesReturned < 4 {
		fmt.Printf("DEBUG: Insufficient bytes returned: %d\n", bytesReturned)
		fmt.Printf("=== JUNCTION DEBUG END ===\n")
		return false, nil
	}

	reparseTag := *(*uint32)(unsafe.Pointer(&buffer[0]))
	fmt.Printf("DEBUG: Reparse tag: 0x%x\n", reparseTag)

	// IO_REPARSE_TAG_MOUNT_POINT indicates a junction point
	const IO_REPARSE_TAG_MOUNT_POINT = 0xA0000003
	// IO_REPARSE_TAG_SYMLINK indicates a symbolic link
	const IO_REPARSE_TAG_SYMLINK = 0xA000000C

	fmt.Printf("DEBUG: IO_REPARSE_TAG_MOUNT_POINT: 0x%x\n", IO_REPARSE_TAG_MOUNT_POINT)
	fmt.Printf("DEBUG: IO_REPARSE_TAG_SYMLINK: 0x%x\n", IO_REPARSE_TAG_SYMLINK)

	isJunction := reparseTag == IO_REPARSE_TAG_MOUNT_POINT
	fmt.Printf("DEBUG: Is junction point: %v\n", isJunction)
	fmt.Printf("=== JUNCTION DEBUG END ===\n")

	return isJunction, nil
}
