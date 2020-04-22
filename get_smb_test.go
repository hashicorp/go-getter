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

func TestSmbGetter_Get(t *testing.T) {
	smbTestsPreCheck(t)

	tests := []struct {
		name    string
		rawURL  string
		file    string
		mounted bool
		fail    bool
	}{
		{
			"smbclient with registered authentication in private share",
			"smb://user:password@samba/private/subdir",
			"file.txt",
			false,
			false,
		},
		{
			"smbclient with registered authentication with file in private share",
			"smb://user:password@samba/private/subdir/file.txt",
			"file.txt",
			false,
			true,
		},
		{
			"smbclient with only registered username authentication in private share",
			"smb://user@samba/private/subdir",
			"file.txt",
			false,
			true,
		},
		{
			"smbclient with non registered username authentication in public share",
			"smb://username@samba/public/subdir",
			"file.txt",
			false,
			false,
		},
		{
			"smbclient without authentication in private share",
			"smb://samba/private/subdir",
			"file.txt",
			false,
			true,
		},
		{
			"smbclient without authentication in public share",
			"smb://samba/public/subdir",
			"file.txt",
			false,
			false,
		},
		{
			"local mounted smb shared file",
			"smb://mnt/file.txt",
			"file.txt",
			true,
			true,
		},
		{
			"local mounted smb shared directory",
			"smb://mnt/subdir",
			"file.txt",
			true,
			false,
		},
		{
			"non existent directory in private share",
			"smb://user:password@samba/private/invalid",
			"",
			false,
			true,
		},
		{
			"non existent directory in public share",
			"smb://samba/public/invalid",
			"",
			false,
			true,
		},
		{
			"no hostname provided",
			"smb://",
			"",
			false,
			true,
		},
		{
			"no filepath provided",
			"smb://samba",
			"",
			false,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
				if tt.mounted {
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
		name    string
		rawURL  string
		file    string
		mounted bool
		fail    bool
	}{
		{
			"smbclient with registered authentication in private share",
			"smb://user:password@samba/private/file.txt",
			"file.txt",
			false,
			false,
		},
		{
			"smbclient with registered authentication and subdirectory in private share",
			"smb://user:password@samba/private/subdir/file.txt",
			"file.txt",
			false,
			false,
		},
		{
			"smbclient with only registered username authentication in private share",
			"smb://user@samba/private/file.txt",
			"file.txt",
			false,
			true,
		},
		{
			"smbclient with non registered username authentication in public share",
			"smb://username@samba/public/file.txt",
			"file.txt",
			false,
			false,
		},
		{
			"smbclient without authentication in public share",
			"smb://samba/public/file.txt",
			"file.txt",
			false,
			false,
		},
		{
			"smbclient without authentication in private share",
			"smb://samba/private/file.txt",
			"file.txt",
			false,
			true,
		},
		{
			"smbclient get directory in private share",
			"smb://user:password@samba/private/subdir",
			"",
			false,
			true,
		},
		{
			"smbclient get directory in public share",
			"smb://samba/public/subdir",
			"",
			false,
			true,
		},
		{
			"local mounted smb shared file",
			"smb://mnt/file.txt",
			"file.txt",
			true,
			false,
		},
		{
			"local mounted smb shared directory",
			"smb://mnt/subdir",
			"",
			true,
			true,
		},
		{
			"non existent file in private share",
			"smb://user:password@samba/private/invalidfile.txt",
			"",
			false,
			true,
		},
		{
			"non existent file in public share",
			"smb://samba/public/invalidfile.txt",
			"",
			false,
			true,
		},
		{
			"no hostname provided",
			"smb://",
			"",
			false,
			true,
		},
		{
			"no filepath provided",
			"smb://samba",
			"",
			false,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst := tempDir(t)
			defer os.RemoveAll(dst)

			url, err := urlhelper.Parse(tt.rawURL)
			if err != nil {
				t.Fatalf("err: %s", err.Error())
			}
			req := &Request{
				Dst: filepath.Join(dst, tt.file),
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
				// Verify if the file was successfully downloaded
				// and exists at the destination folder
				mainPath := filepath.Join(dst, tt.file)
				if _, err := os.Stat(mainPath); err != nil {
					t.Fatalf("err: %s", err)
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
		fail         bool
	}{
		{
			"smbclient modefile for existing file in authenticated private share",
			"smb://user:password@samba/private/file.txt",
			ModeFile,
			false,
		},
		{
			"smbclient modedir for existing directory in authenticated private share",
			"smb://user:password@samba/private/subdir",
			ModeDir,
			false,
		},
		{
			"mode fail for non existent directory in authenticated private share",
			"smb://user:password@samba/private/invaliddir",
			0,
			true,
		},
		{
			"mode fail for non existent file in authenticated private share",
			"smb://user:password@samba/private/invalidfile.txt",
			0,
			true,
		},
		{
			"smbclient modefile for existing file in public share",
			"smb://samba/public/file.txt",
			ModeFile,
			false,
		},
		{
			"smbclient modedir for existing directory in public share",
			"smb://samba/public/subdir",
			ModeDir,
			false,
		},
		{
			"mode fail for non existent directory in public share",
			"smb://samba/public/invaliddir",
			0,
			true,
		},
		{
			"mode fail for non existent file in public share",
			"smb://samba/public/invalidfile.txt",
			0,
			true,
		},
		{
			"local mount modefile for existing file",
			"smb://mnt/file.txt",
			ModeFile,
			false,
		},
		{
			"local mount modedir for existing directory",
			"smb://mnt/subdir",
			ModeDir,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
