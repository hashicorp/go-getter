package getter

import (
	"context"
	urlhelper "github.com/hashicorp/go-getter/helper/url"
	"os"
	"path/filepath"
	"testing"
)

func TestSmbGetter_impl(t *testing.T) {
	var _ Getter = new(SmbGetter)
}

// TODO:
// write higher level tests
// save tests results on circleci
// write docs of how to run tests locally (makefile?)
// update readme

func TestSmbGetter_Get(t *testing.T) {
	smbTestsPreCheck(t)

	tests := []struct {
		name      string
		rawURL    string
		file      string
		createDir string
		fail      bool
	}{
		{
			"smbclient with authentication",
			"smb://username:password@samba/shared/subdir",
			"file.txt",
			"",
			false,
		},
		{
			"smbclient with authentication with file",
			"smb://username:password@samba/shared/subdir/file.txt",
			"file.txt",
			"",
			true,
		},
		{
			"smbclient with only username authentication",
			"smb://username@samba/shared/subdir",
			"file.txt",
			"",
			false,
		},
		{
			"smbclient without authentication",
			"smb://samba/shared/subdir",
			"file.txt",
			"",
			false,
		},
		{
			"local mounted smb shared file",
			"smb://mnt/shared/file.txt",
			"file.txt",
			"/mnt/shared",
			true,
		},
		{
			"local mounted smb shared directory",
			"smb://mnt/shared/subdir",
			"file.txt",
			"/mnt/shared/subdir",
			false,
		},
		{
			"no hostname provided",
			"smb://",
			"",
			"",
			true,
		},
		{
			"no filepath provided",
			"smb://samba",
			"",
			"",
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.createDir != "" {
				// mock mounted folder by creating one
				err := os.MkdirAll(tt.createDir, 0755)
				if err != nil {
					t.Fatalf("err: %s", err.Error())
				}

				if tt.file != "" {
					f, err := os.Create(filepath.Join(tt.createDir, tt.file))
					if err != nil {
						t.Fatalf("err: %s", err.Error())
					}
					defer f.Close()

					// Write content to assert later
					_, err = f.WriteString("Hello\n")
					if err != nil {
						t.Fatalf("err: %s", err.Error())
					}
					f.Sync()
				}

				defer os.RemoveAll(tt.createDir)
			}

			dst := tempDir(t)
			defer os.RemoveAll(dst)

			url, err := urlhelper.Parse(tt.rawURL)
			if err != nil {
				t.Fatalf("err: %s", err.Error())
			}
			req := &Request{
				Dst: dst,
				u:   url,
			}

			g := new(SmbGetter)
			err = g.Get(context.Background(), req)

			fail := err != nil
			if tt.fail != fail {
				if fail {
					t.Fatalf("err: unexpected error %s", err.Error())
				}
				t.Fatalf("err: expecting to fail but it did not")
			}

			if !tt.fail {
				if tt.createDir != "" {
					// Verify the destination folder is a symlink to the mounted one
					fi, err := os.Lstat(dst)
					if err != nil {
						t.Fatalf("err: %s", err)
					}
					if fi.Mode()&os.ModeSymlink == 0 {
						t.Fatal("destination is not a symlink")
					}
					// Verify the file exists
					assertContents(t, filepath.Join(dst, tt.file), "Hello\n")
				} else {
					// Verify if the file was successfully downloaded
					// and exists at the destination folder
					mainPath := filepath.Join(dst, tt.file)
					if _, err := os.Stat(mainPath); err != nil {
						t.Fatalf("err: %s", err)
					}
				}
			}
		})
	}
}

func TestSmbGetter_GetFile(t *testing.T) {
	smbTestsPreCheck(t)

	tests := []struct {
		name      string
		rawURL    string
		file      string
		createDir string
		fail      bool
	}{
		{
			"smbclient with authentication",
			"smb://username:password@samba/shared/file.txt",
			"file.txt",
			"",
			false,
		},
		{
			"smbclient with authentication and subdirectory",
			"smb://username:password@samba/shared/subdir/file.txt",
			"file.txt",
			"",
			false,
		},
		{
			"smbclient with only username authentication",
			"smb://username@samba/shared/file.txt",
			"file.txt",
			"",
			false,
		},
		{
			"smbclient without authentication",
			"smb://samba/shared/file.txt",
			"file.txt",
			"",
			false,
		},
		{
			"smbclient get directory",
			"smb://username:password@samba/shared/subdir",
			"",
			"",
			true,
		},
		{
			"local mounted smb shared file",
			"smb://mnt/shared/file.txt",
			"file.txt",
			"/mnt/shared",
			false,
		},
		{
			"local mounted smb shared directory",
			"smb://mnt/shared/subdir",
			"",
			"/mnt/shared/subdir",
			true,
		},
		{
			"non existent file",
			"smb://username:password@samba/shared/invalidfile.txt",
			"",
			"",
			true,
		},
		{
			"non existent directory",
			"smb://username:password@samba/shared/invaliddir",
			"",
			"",
			true,
		},
		{
			"no hostname provided",
			"smb://",
			"",
			"",
			true,
		},
		{
			"no filepath provided",
			"smb://samba",
			"",
			"",
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.createDir != "" {
				// mock mounted folder by creating one
				err := os.MkdirAll(tt.createDir, 0755)
				if err != nil {
					t.Fatalf("err: %s", err.Error())
				}

				if tt.file != "" {
					f, err := os.Create(filepath.Join(tt.createDir, tt.file))
					if err != nil {
						t.Fatalf("err: %s", err.Error())
					}
					defer f.Close()

					// Write content to assert later
					_, err = f.WriteString("Hello\n")
					if err != nil {
						t.Fatalf("err: %s", err.Error())
					}
					f.Sync()
				}

				defer os.RemoveAll(tt.createDir)
			}

			dst := tempDir(t)
			defer os.RemoveAll(dst)

			url, err := urlhelper.Parse(tt.rawURL)
			if err != nil {
				t.Fatalf("err: %s", err.Error())
			}
			req := &Request{
				Dst: dst,
				u:   url,
			}

			g := new(SmbGetter)
			err = g.GetFile(context.Background(), req)

			fail := err != nil
			if tt.fail != fail {
				if fail {
					t.Fatalf("err: unexpected error %s", err.Error())
				}
				t.Fatalf("err: expecting to fail but it did not")
			}

			if !tt.fail {
				if tt.createDir != "" {
					// Verify the destination folder is a symlink to the mounted one
					fi, err := os.Lstat(dst)
					if err != nil {
						t.Fatalf("err: %s", err)
					}
					if fi.Mode()&os.ModeSymlink == 0 {
						t.Fatal("destination is not a symlink")
					}
					// Verify the file exists
					assertContents(t, dst, "Hello\n")
				} else {
					// Verify if the file was successfully downloaded
					// and exists at the destination folder
					mainPath := filepath.Join(dst, tt.file)
					if _, err := os.Stat(mainPath); err != nil {
						t.Fatalf("err: %s", err)
					}
				}
			}
		})
	}
}

func TestSmbGetter_Mode(t *testing.T) {
	smbTestsPreCheck(t)

	tests := []struct {
		name         string
		rawURL       string
		expectedMode Mode
		file         string
		createDir    string
		fail         bool
	}{
		{
			"smbclient modefile for existing file",
			"smb://username:password@samba/shared/file.txt",
			ModeFile,
			"file.txt",
			"",
			false,
		},
		{
			"smbclient modedir for existing directory",
			"smb://username:password@samba/shared/subdir",
			ModeDir,
			"",
			"",
			false,
		},
		{
			"mode fail for non existent directory",
			"smb://username:password@samba/shared/invaliddir",
			0,
			"",
			"",
			true,
		},
		{
			"mode fail for non existent file",
			"smb://username:password@samba/shared/invalidfile.txt",
			0,
			"",
			"",
			true,
		},
		{
			"local mount modefile for existing file",
			"smb://mnt/shared/file.txt",
			ModeFile,
			"file.txt",
			"/mnt/shared",
			false,
		},
		{
			"local mount modedir for existing directory",
			"smb://mnt/shared/subdir",
			ModeDir,
			"",
			"/mnt/shared/subdir",
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.createDir != "" {
				// mock mounted folder by creating one
				err := os.MkdirAll(tt.createDir, 0755)
				if err != nil {
					t.Fatalf("err: %s", err.Error())
				}

				if tt.file != "" {
					f, err := os.Create(filepath.Join(tt.createDir, tt.file))
					if err != nil {
						t.Fatalf("err: %s", err.Error())
					}
					defer f.Close()
				}

				defer os.RemoveAll(tt.createDir)
			}

			url, err := urlhelper.Parse(tt.rawURL)
			if err != nil {
				t.Fatalf("err: %s", err.Error())
			}

			g := new(SmbGetter)
			mode, err := g.Mode(context.Background(), url)

			fail := err != nil
			if tt.fail != fail {
				if fail {
					t.Fatalf("err: unexpected error %s", err.Error())
				}
				t.Fatalf("err: expecting to fail but it did not")
			}

			if mode != tt.expectedMode {
				t.Fatalf("err: expeting mode %d, actual mode %d", tt.expectedMode, mode)
			}
		})
	}
}

func smbTestsPreCheck(t *testing.T) {
	r := os.Getenv("ACC_SMB_TEST")
	if r != "1" {
		t.Skip("smb getter tests won't run. ACC_SMB_TEST not set")
	}
}
