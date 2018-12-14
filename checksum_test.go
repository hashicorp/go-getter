package getter

import (
	"crypto/md5"
	"encoding/hex"
	"hash"
	"net/url"
	"path/filepath"
	"testing"
)

func u(t *testing.T, in string) *url.URL {
	u, err := url.Parse(in)
	if err != nil {
		t.Fatalf("cannot parse %s: %v", in, err)
	}
	return u
}

func Test_fileChecksum_checksum(t *testing.T) {

	type fields struct {
		Type     string
		Hash     hash.Hash
		Value    string
		Filename string
	}
	type args struct {
		source string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"correct checksum",
			fields{"md5", md5.New(), "074729f0ccb41a391fb646c38f86ea54", "content.txt"},
			args{filepath.Join(fixtureDir, "checksum-file", "content.txt")},
			false,
		},
		{"wrong file",
			fields{"md5", md5.New(), "074729f0ccb41a391fb646c38f86ea54", "content.txt"},
			args{filepath.Join(fixtureDir, "checksum-file", "sha1-p.sum")},
			true,
		},
		{"file not fount",
			fields{"md5", md5.New(), "074729f0ccb41a391fb646c38f86ea54", "content.txt"},
			args{filepath.Join(fixtureDir, "checksum-file", "not-such-file-or-directory.txt")},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &fileChecksum{
				Type:     tt.fields.Type,
				Hash:     tt.fields.Hash,
				Filename: tt.fields.Filename,
			}
			var err error
			if c.Value, err = hex.DecodeString(tt.fields.Value); err != nil {
				panic(err)
			}
			if err := c.checksum(tt.args.source); (err != nil) != tt.wantErr {
				t.Errorf("fileChecksum.checksum() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
