//go:build windows
// +build windows

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package getter

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestIsWindowsJunctionPoint tests basic junction point detection
func TestIsWindowsJunctionPoint(t *testing.T) {
	tempDir := t.TempDir()

	// Test regular directory
	regularDir := filepath.Join(tempDir, "regular")
	if err := os.Mkdir(regularDir, 0755); err != nil {
		t.Fatalf("Failed to create regular directory: %v", err)
	}

	isJunction, err := isWindowsJunctionPoint(regularDir)
	if err != nil {
		t.Errorf("Unexpected error checking regular directory: %v", err)
	}
	if isJunction {
		t.Error("Regular directory should not be detected as junction")
	}

	// Test actual junction point creation and detection
	targetDir := filepath.Join(tempDir, "target")
	if err := os.Mkdir(targetDir, 0755); err != nil {
		t.Fatalf("Failed to create target directory: %v", err)
	}

	junctionDir := filepath.Join(tempDir, "junction")

	// Create junction using mklink (requires admin or Developer Mode)
	// Note: This might fail in CI without proper permissions
	if err := createJunctionForTest(junctionDir, targetDir); err != nil {
		t.Logf("Could not create junction (this may be expected in CI): %v", err)
		t.Skip("Skipping junction test - unable to create junction point")
	}

	// Verify junction point detection
	isJunction, err = isWindowsJunctionPoint(junctionDir)
	if err != nil {
		t.Errorf("Unexpected error checking junction point: %v", err)
	}
	if !isJunction {
		t.Error("Junction point should be detected as junction")
	}

	// Test symbolic link (different from junction)
	symlinkDir := filepath.Join(tempDir, "symlink")
	if err := os.Symlink(targetDir, symlinkDir); err != nil {
		t.Logf("Could not create symlink (this may be expected): %v", err)
	} else {
		// Verify symlink is NOT detected as junction
		isJunction, err = isWindowsJunctionPoint(symlinkDir)
		if err != nil {
			t.Errorf("Unexpected error checking symlink: %v", err)
		}
		if isJunction {
			t.Error("Symbolic link should NOT be detected as junction")
		}
	}

	// Test non-existent path
	nonExistent := filepath.Join(tempDir, "nonexistent")
	isJunction, err = isWindowsJunctionPoint(nonExistent)
	if err == nil {
		t.Error("Expected error for non-existent path")
	}
	if isJunction {
		t.Error("Non-existent path should not be detected as junction")
	}

	// Test file (not directory)
	testFile := filepath.Join(tempDir, "testfile.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	isJunction, err = isWindowsJunctionPoint(testFile)
	if err != nil {
		t.Errorf("Unexpected error checking file: %v", err)
	}
	if isJunction {
		t.Error("Regular file should not be detected as junction")
	}
}

// TestResolveSymlinks tests Windows-specific symlink resolution
func TestResolveSymlinks(t *testing.T) {
	tempDir := t.TempDir()

	// Test regular directory (common case where EvalSymlinks might fail)
	regularDir := filepath.Join(tempDir, "regular")
	if err := os.Mkdir(regularDir, 0755); err != nil {
		t.Fatalf("Failed to create regular directory: %v", err)
	}

	resolved, err := resolveSymlinks(regularDir)
	if err != nil {
		t.Errorf("Unexpected error resolving regular directory: %v", err)
	}

	expected := filepath.Clean(regularDir)
	if resolved != expected {
		t.Errorf("Expected %s, got %s", expected, resolved)
	}

	// Test junction point if we can create one
	targetDir := filepath.Join(tempDir, "target")
	if err := os.Mkdir(targetDir, 0755); err != nil {
		t.Fatalf("Failed to create target directory: %v", err)
	}

	junctionDir := filepath.Join(tempDir, "junction")
	if err := createJunctionForTest(junctionDir, targetDir); err != nil {
		t.Logf("Could not create junction: %v", err)
		t.Skip("Skipping junction resolution test - unable to create junction point")
	}

	// Resolve junction point
	resolved, err = resolveSymlinks(junctionDir)
	if err != nil {
		t.Errorf("Unexpected error resolving junction: %v", err)
	}

	// Should resolve to target directory
	expected = filepath.Clean(targetDir)
	if resolved != expected {
		t.Errorf("Expected %s, got %s", expected, resolved)
	}
}

// createJunctionForTest attempts to create a Windows junction point for testing
// This might fail without admin privileges or Developer Mode enabled
func createJunctionForTest(junctionPath, targetPath string) error {
	// Try using os.Symlink first (works in newer Windows with Developer Mode)
	if err := os.Symlink(targetPath, junctionPath); err == nil {
		return nil
	}

	// If that fails, we can't easily create junctions in tests without external tools
	// In a real-world scenario, we'd use mklink /J or Windows APIs
	return os.ErrPermission // Indicate we couldn't create the junction
}

// TestWindowsJunctionPoint_Integration tests junction point functionality
// using actual Windows junction creation when possible
func TestWindowsJunctionPoint_Integration(t *testing.T) {
	tempDir := t.TempDir()

	// Create target directory
	targetDir := filepath.Join(tempDir, "target")
	if err := os.Mkdir(targetDir, 0755); err != nil {
		t.Fatalf("Failed to create target directory: %v", err)
	}

	// Create a test file in target
	testFile := filepath.Join(targetDir, "test.txt")
	testContent := "junction test content"
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	junctionDir := filepath.Join(tempDir, "junction")

	// Try to create junction using mklink command
	if err := createJunctionWithMklink(junctionDir, targetDir); err != nil {
		t.Logf("Could not create junction with mklink: %v", err)
		t.Skip("Skipping integration test - mklink not available or insufficient permissions")
	}

	// Test junction detection
	isJunction, err := isWindowsJunctionPoint(junctionDir)
	if err != nil {
		t.Errorf("Error detecting junction: %v", err)
	}
	if !isJunction {
		t.Error("Created junction not detected as junction point")
	}

	// Test junction target resolution
	target, err := resolveJunctionTarget(junctionDir)
	if err != nil {
		t.Errorf("Error resolving junction target: %v", err)
	}

	// Normalize paths for comparison (handle different path formats)
	expectedTarget := filepath.Clean(targetDir)
	actualTarget := filepath.Clean(target)

	if !strings.EqualFold(actualTarget, expectedTarget) {
		t.Errorf("Junction target mismatch: expected %s, got %s", expectedTarget, actualTarget)
	}

	// Test that junction works functionally
	junctionTestFile := filepath.Join(junctionDir, "test.txt")
	content, err := os.ReadFile(junctionTestFile)
	if err != nil {
		t.Errorf("Could not read file through junction: %v", err)
	}
	if string(content) != testContent {
		t.Errorf("File content mismatch through junction: expected %s, got %s", testContent, string(content))
	}

	// Test resolveSymlinks with junction
	resolved, err := resolveSymlinks(junctionDir)
	if err != nil {
		t.Errorf("Error resolving symlinks for junction: %v", err)
	}
	if !strings.EqualFold(filepath.Clean(resolved), expectedTarget) {
		t.Errorf("resolveSymlinks mismatch: expected %s, got %s", expectedTarget, resolved)
	}
}

// TestWindowsJunctionPoint_ErrorCases tests various error conditions
func TestWindowsJunctionPoint_ErrorCases(t *testing.T) {
	// Test with invalid path characters
	invalidPath := "C:\\invalid<>path"
	isJunction, err := isWindowsJunctionPoint(invalidPath)
	if err == nil {
		t.Error("Expected error for invalid path characters")
	}
	if isJunction {
		t.Error("Invalid path should not be detected as junction")
	}

	// Test with very long path
	longPath := "C:\\" + strings.Repeat("very_long_directory_name_", 20) + "\\test"
	isJunction, err = isWindowsJunctionPoint(longPath)
	if err == nil {
		t.Log("Long path test completed (may succeed on newer Windows)")
	}
	if isJunction {
		t.Error("Non-existent long path should not be detected as junction")
	}

	// Test resolveJunctionTarget with non-junction
	tempDir := t.TempDir()
	regularDir := filepath.Join(tempDir, "regular")
	if err := os.Mkdir(regularDir, 0755); err != nil {
		t.Fatalf("Failed to create regular directory: %v", err)
	}

	target, err := resolveJunctionTarget(regularDir)
	if err == nil {
		t.Errorf("Expected error resolving non-junction as junction, got target: %s", target)
	}
}

// TestWindowsJunctionPoint_PermissionTests tests permission-related scenarios
func TestWindowsJunctionPoint_PermissionTests(t *testing.T) {
	// Test with system directories (should not fail, just return false)
	systemDirs := []string{
		"C:\\Windows",
		"C:\\Program Files",
		"C:\\System Volume Information", // This one might fail due to permissions
	}

	for _, dir := range systemDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue // Skip if directory doesn't exist
		}

		isJunction, err := isWindowsJunctionPoint(dir)
		if err != nil {
			t.Logf("Permission error checking %s (expected): %v", dir, err)
			continue
		}

		t.Logf("System directory %s: isJunction=%v", dir, isJunction)
	}
}

// createJunctionWithMklink creates a junction using the Windows mklink command
func createJunctionWithMklink(junctionPath, targetPath string) error {
	cmd := exec.Command("cmd", "/c", "mklink", "/J", junctionPath, targetPath)
	_, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	// Check if junction was actually created
	if _, err := os.Lstat(junctionPath); err != nil {
		return err
	}

	return nil
}
