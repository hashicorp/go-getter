package getter

import (
	"context"
	urlhelper "github.com/hashicorp/go-getter/helper/url"
	testing_helper "github.com/hashicorp/go-getter/v2/helper/testing"
	"os"
	"path/filepath"
	"testing"
)

func TestSmb_GetterImpl(t *testing.T) {
	var _ Getter = new(SmbClientGetter)
}

func TestSmb_GetterGet(t *testing.T) {
	smbTestsPreCheck(t)

	tests := []struct {
		name   string
		rawURL string
		file   string
		fail   bool
	}{
		{
			"smbclient with registered authentication in private share",
			"smb://user:password@samba/private/subdir",
			"file.txt",
			false,
		},
		{
			"smbclient with registered authentication with file in private share",
			"smb://user:password@samba/private/subdir/file.txt",
			"file.txt",
			true,
		},
		{
			"smbclient with only registered username authentication in private share",
			"smb://user@samba/private/subdir",
			"file.txt",
			true,
		},
		{
			"smbclient with non registered username authentication in public share",
			"smb://username@samba/public/subdir",
			"file.txt",
			false,
		},
		{
			"smbclient without authentication in private share",
			"smb://samba/private/subdir",
			"file.txt",
			true,
		},
		{
			"smbclient without authentication in public share",
			"smb://samba/public/subdir",
			"file.txt",
			false,
		},
		{
			"non existent directory in private share",
			"smb://user:password@samba/private/invalid",
			"",
			true,
		},
		{
			"non existent directory in public share",
			"smb://samba/public/invalid",
			"",
			true,
		},
		{
			"no hostname provided",
			"smb://",
			"",
			true,
		},
		{
			"no filepath provided",
			"smb://samba",
			"",
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst := testing_helper.TempDir(t)
			defer os.RemoveAll(dst)

			url, err := urlhelper.Parse(tt.rawURL)
			if err != nil {
				t.Fatalf("err: %s", err.Error())
			}
			req := &Request{
				Dst: dst,
				u:   url,
			}

			g := new(SmbClientGetter)
			err = g.Get(context.Background(), req)

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
				testing_helper.AssertContents(t, filepath.Join(dst, tt.file), "Hello\n")
			}
		})
	}
}

func TestSmb_GetterGetFile(t *testing.T) {
	smbTestsPreCheck(t)

	tests := []struct {
		name   string
		rawURL string
		file   string
		fail   bool
	}{
		{
			"smbclient with registered authentication in private share",
			"smb://user:password@samba/private/file.txt",
			"file.txt",
			false,
		},
		{
			"smbclient with registered authentication and subdirectory in private share",
			"smb://user:password@samba/private/subdir/file.txt",
			"file.txt",
			false,
		},
		{
			"smbclient with only registered username authentication in private share",
			"smb://user@samba/private/file.txt",
			"file.txt",
			true,
		},
		{
			"smbclient with non registered username authentication in public share",
			"smb://username@samba/public/file.txt",
			"file.txt",
			false,
		},
		{
			"smbclient without authentication in public share",
			"smb://samba/public/file.txt",
			"file.txt",
			false,
		},
		{
			"smbclient without authentication in private share",
			"smb://samba/private/file.txt",
			"file.txt",
			true,
		},
		{
			"smbclient get directory in private share",
			"smb://user:password@samba/private/subdir",
			"",
			true,
		},
		{
			"smbclient get directory in public share",
			"smb://samba/public/subdir",
			"",
			true,
		},
		{
			"non existent file in private share",
			"smb://user:password@samba/private/invalidfile.txt",
			"",
			true,
		},
		{
			"non existent file in public share",
			"smb://samba/public/invalidfile.txt",
			"",
			true,
		},
		{
			"no hostname provided",
			"smb://",
			"",
			true,
		},
		{
			"no filepath provided",
			"smb://samba",
			"",
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst := testing_helper.TempDir(t)
			defer os.RemoveAll(dst)

			url, err := urlhelper.Parse(tt.rawURL)
			if err != nil {
				t.Fatalf("err: %s", err.Error())
			}
			req := &Request{
				Dst: filepath.Join(dst, tt.file),
				u:   url,
			}

			g := new(SmbClientGetter)
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
				testing_helper.AssertContents(t, filepath.Join(dst, tt.file), "Hello\n")
			}
		})
	}
}

func TestSmb_GetterMode(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, err := urlhelper.Parse(tt.rawURL)
			if err != nil {
				t.Fatalf("err: %s", err.Error())
			}

			g := new(SmbClientGetter)
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
