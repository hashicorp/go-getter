//go:build !windows
// +build !windows

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package getter

import (
	"os"
	"path/filepath"
	"testing"
)

// TestIsWindowsJunctionPoint tests that the function always returns false on Unix
func TestIsWindowsJunctionPoint(t *testing.T) {
	tempDir := t.TempDir()

	// Test regular directory
	regularDir := filepath.Join(tempDir, "regular")
	if err := os.Mkdir(regularDir, 0755); err != nil {
		t.Fatalf("Failed to create regular directory: %v", err)
	}

	isJunction, err := isWindowsJunctionPoint(regularDir)
	if err != nil {
		t.Errorf("Expected no error on non-Windows, got: %v", err)
	}
	if isJunction {
		t.Error("Expected false for junction check on non-Windows")
	}

	// Test symlink directory (Unix)
	symlinkDir := filepath.Join(tempDir, "symlink")
	if err := os.Symlink(regularDir, symlinkDir); err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	isJunction, err = isWindowsJunctionPoint(symlinkDir)
	if err != nil {
		t.Errorf("Expected no error on non-Windows, got: %v", err)
	}
	if isJunction {
		t.Error("Expected false for symlink on non-Windows")
	}

	// Test non-existent path
	nonExistent := filepath.Join(tempDir, "nonexistent")
	isJunction, err = isWindowsJunctionPoint(nonExistent)
	if err != nil {
		t.Errorf("Expected no error on non-Windows, got: %v", err)
	}
	if isJunction {
		t.Error("Expected false for non-existent path on non-Windows")
	}
}

// TestResolveSymlinks tests Unix-specific symlink resolution
func TestResolveSymlinks(t *testing.T) {
	tempDir := t.TempDir()

	// Create target directory
	targetDir := filepath.Join(tempDir, "target")
	if err := os.Mkdir(targetDir, 0755); err != nil {
		t.Fatalf("Failed to create target directory: %v", err)
	}

	// Create symlink
	symlinkDir := filepath.Join(tempDir, "symlink")
	if err := os.Symlink(targetDir, symlinkDir); err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	// Resolve symlink
	resolved, err := resolveSymlinks(symlinkDir)
	if err != nil {
		t.Errorf("Unexpected error resolving symlink: %v", err)
	}

	// On macOS, /var is a symlink to /private/var, so we need to resolve
	// the expected path through the same mechanism to compare properly
	expectedResolved, err := filepath.EvalSymlinks(targetDir)
	if err != nil {
		// If EvalSymlinks fails, fall back to cleaned path
		expectedResolved = filepath.Clean(targetDir)
	}

	if resolved != expectedResolved {
		t.Errorf("Expected %s, got %s", expectedResolved, resolved)
	}
}

// TestUnixSymlinks_Integration tests Unix symlink functionality
func TestUnixSymlinks_Integration(t *testing.T) {
	tempDir := t.TempDir()

	// Create target directory with content
	targetDir := filepath.Join(tempDir, "target")
	if err := os.Mkdir(targetDir, 0755); err != nil {
		t.Fatalf("Failed to create target directory: %v", err)
	}

	testFile := filepath.Join(targetDir, "test.txt")
	testContent := "symlink test content"
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create symlink directory
	symlinkDir := filepath.Join(tempDir, "symlink")
	if err := os.Symlink(targetDir, symlinkDir); err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	// Verify isWindowsJunctionPoint always returns false on Unix
	isJunction, err := isWindowsJunctionPoint(symlinkDir)
	if err != nil {
		t.Errorf("Unexpected error on Unix: %v", err)
	}
	if isJunction {
		t.Error("isWindowsJunctionPoint should always return false on Unix")
	}

	// Test resolveSymlinks works correctly
	resolved, err := resolveSymlinks(symlinkDir)
	if err != nil {
		t.Errorf("Error resolving symlink: %v", err)
	}

	// On macOS, resolve the expected path through the same mechanism
	// to handle /var -> /private/var symlink
	expectedTarget, err := filepath.EvalSymlinks(targetDir)
	if err != nil {
		// If EvalSymlinks fails, fall back to cleaned path
		expectedTarget = filepath.Clean(targetDir)
	}
	actualTarget := filepath.Clean(resolved)
	if actualTarget != expectedTarget {
		t.Errorf("Symlink resolution mismatch: expected %s, got %s", expectedTarget, actualTarget)
	}

	// Verify symlink works functionally
	symlinkTestFile := filepath.Join(symlinkDir, "test.txt")
	content, err := os.ReadFile(symlinkTestFile)
	if err != nil {
		t.Errorf("Could not read file through symlink: %v", err)
	}
	if string(content) != testContent {
		t.Errorf("File content mismatch through symlink: expected %s, got %s", testContent, string(content))
	}
}

// TestUnixSymlinks_ChainedSymlinks tests multiple levels of symlinks
func TestUnixSymlinks_ChainedSymlinks(t *testing.T) {
	tempDir := t.TempDir()

	// Create target directory
	targetDir := filepath.Join(tempDir, "target")
	if err := os.Mkdir(targetDir, 0755); err != nil {
		t.Fatalf("Failed to create target directory: %v", err)
	}

	// Create first symlink
	symlink1 := filepath.Join(tempDir, "symlink1")
	if err := os.Symlink(targetDir, symlink1); err != nil {
		t.Fatalf("Failed to create first symlink: %v", err)
	}

	// Create second symlink pointing to first symlink
	symlink2 := filepath.Join(tempDir, "symlink2")
	if err := os.Symlink(symlink1, symlink2); err != nil {
		t.Fatalf("Failed to create second symlink: %v", err)
	}

	// Resolve the chained symlinks
	resolved, err := resolveSymlinks(symlink2)
	if err != nil {
		t.Errorf("Error resolving chained symlinks: %v", err)
	}

	// On macOS, resolve the expected path through the same mechanism
	// to handle /var -> /private/var symlink
	expectedTarget, err := filepath.EvalSymlinks(targetDir)
	if err != nil {
		// If EvalSymlinks fails, fall back to cleaned path
		expectedTarget = filepath.Clean(targetDir)
	}
	actualTarget := filepath.Clean(resolved)
	if actualTarget != expectedTarget {
		t.Errorf("Chained symlink resolution mismatch: expected %s, got %s", expectedTarget, actualTarget)
	}

	// Test junction detection on all symlinks
	for _, path := range []string{symlink1, symlink2} {
		isJunction, err := isWindowsJunctionPoint(path)
		if err != nil {
			t.Errorf("Unexpected error checking %s: %v", path, err)
		}
		if isJunction {
			t.Errorf("isWindowsJunctionPoint should return false for symlink %s", path)
		}
	}
}

// TestUnixSymlinks_BrokenSymlinks tests handling of broken symlinks
func TestUnixSymlinks_BrokenSymlinks(t *testing.T) {
	tempDir := t.TempDir()

	// Create symlink to non-existent target
	brokenSymlink := filepath.Join(tempDir, "broken")
	nonExistentTarget := filepath.Join(tempDir, "nonexistent")
	if err := os.Symlink(nonExistentTarget, brokenSymlink); err != nil {
		t.Fatalf("Failed to create broken symlink: %v", err)
	}

	// isWindowsJunctionPoint should still work (return false)
	isJunction, err := isWindowsJunctionPoint(brokenSymlink)
	if err != nil {
		t.Errorf("Unexpected error on broken symlink: %v", err)
	}
	if isJunction {
		t.Error("Broken symlink should not be detected as junction")
	}

	// resolveSymlinks should fail appropriately
	_, err = resolveSymlinks(brokenSymlink)
	if err == nil {
		t.Error("Expected error resolving broken symlink")
	}
}

// TestUnixSymlinks_FileSymlinks tests file symlinks (not just directory symlinks)
func TestUnixSymlinks_FileSymlinks(t *testing.T) {
	tempDir := t.TempDir()

	// Create target file
	targetFile := filepath.Join(tempDir, "target.txt")
	testContent := "file symlink test"
	if err := os.WriteFile(targetFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create target file: %v", err)
	}

	// Create file symlink
	symlinkFile := filepath.Join(tempDir, "symlink.txt")
	if err := os.Symlink(targetFile, symlinkFile); err != nil {
		t.Fatalf("Failed to create file symlink: %v", err)
	}

	// Test junction detection (should be false)
	isJunction, err := isWindowsJunctionPoint(symlinkFile)
	if err != nil {
		t.Errorf("Unexpected error on file symlink: %v", err)
	}
	if isJunction {
		t.Error("File symlink should not be detected as junction")
	}

	// Test symlink resolution
	resolved, err := resolveSymlinks(symlinkFile)
	if err != nil {
		t.Errorf("Error resolving file symlink: %v", err)
	}

	// On macOS, resolve the expected path through the same mechanism
	// to handle /var -> /private/var symlink
	expectedTarget, err := filepath.EvalSymlinks(targetFile)
	if err != nil {
		// If EvalSymlinks fails, fall back to cleaned path
		expectedTarget = filepath.Clean(targetFile)
	}
	actualTarget := filepath.Clean(resolved)
	if actualTarget != expectedTarget {
		t.Errorf("File symlink resolution mismatch: expected %s, got %s", expectedTarget, actualTarget)
	}
}

// TestUnixSymlinks_PermissionEdgeCases tests permission-related edge cases
func TestUnixSymlinks_PermissionEdgeCases(t *testing.T) {
	tempDir := t.TempDir()

	// Create a directory and remove read permissions
	targetDir := filepath.Join(tempDir, "noperm")
	if err := os.Mkdir(targetDir, 0755); err != nil {
		t.Fatalf("Failed to create target directory: %v", err)
	}

	// Create symlink to the directory
	symlinkDir := filepath.Join(tempDir, "symlink")
	if err := os.Symlink(targetDir, symlinkDir); err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	// Remove permissions from target
	if err := os.Chmod(targetDir, 0000); err != nil {
		t.Fatalf("Failed to remove permissions: %v", err)
	}

	// Restore permissions at test end
	defer func() {
		os.Chmod(targetDir, 0755)
	}()

	// Junction detection should still work
	isJunction, err := isWindowsJunctionPoint(symlinkDir)
	if err != nil {
		t.Errorf("Unexpected error with permission-restricted symlink: %v", err)
	}
	if isJunction {
		t.Error("Permission-restricted symlink should not be detected as junction")
	}

	// Symlink resolution might fail due to permissions, but should handle gracefully
	_, err = resolveSymlinks(symlinkDir)
	// We don't strictly require this to succeed or fail - depends on the system
	t.Logf("resolveSymlinks on permission-restricted target: %v", err)
}
