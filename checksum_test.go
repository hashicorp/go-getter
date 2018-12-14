package getter

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"hash"
	"net/url"
	"path/filepath"
	"reflect"
	"testing"
)

func u(t *testing.T, in string) *url.URL {
	u, err := url.Parse(in)
	if err != nil {
		t.Fatalf("cannot parse %s: %v", in, err)
	}
	return u
}

func Test_checksumHashAndValue(t *testing.T) {
	checksums := testModule("checksum-file")

	type args struct {
		u *url.URL
	}
	type want struct {
		Filename string
		Checksum string
		Type     string
		Hash     hash.Hash
		Err      bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{"sha1",
			args{
				u(t, checksums+"/content.txt?checksum=sha1:e2c7dc83ac8aa7f181314387f6dfb132cd117e3a"),
			},
			want{
				"content.txt",
				"e2c7dc83ac8aa7f181314387f6dfb132cd117e3a",
				"sha1",
				sha1.New(),
				false,
			},
		},
		{"file sha1",
			args{
				u(t, checksums+"/content.txt?checksum=file:"+checksums+"/sha1-p.sum"),
			},
			want{
				"?content.txt",
				"e2c7dc83ac8aa7f181314387f6dfb132cd117e3a",
				"sha1",
				sha1.New(),
				false,
			},
		},
		{"file sha1 BSD",
			args{
				u(t, checksums+"/content.txt?checksum=file:"+checksums+"/sha1-bsd.sum"),
			},
			want{
				"content.txt",
				"e2c7dc83ac8aa7f181314387f6dfb132cd117e3a",
				"sha1",
				sha1.New(),
				false,
			},
		},
		{"sha256",
			args{
				u(t, checksums+"/content.txt?checksum=sha256:47afcdfff05a6e5d9db5f6c6df2140f04a6e7422d7ad7f6a7006a4f5a78570e4"),
			},
			want{
				"content.txt",
				"47afcdfff05a6e5d9db5f6c6df2140f04a6e7422d7ad7f6a7006a4f5a78570e4",
				"sha256",
				sha256.New(),
				false,
			},
		},
		{"file sha256",
			args{
				u(t, checksums+"/content.txt?checksum=file:"+checksums+"/sha256-p.sum"),
			},
			want{
				"?content.txt",
				"47afcdfff05a6e5d9db5f6c6df2140f04a6e7422d7ad7f6a7006a4f5a78570e4",
				"sha256",
				sha256.New(),
				false,
			},
		},
		{"file sha256 BSD",
			args{
				u(t, checksums+"/content.txt?checksum=file:"+checksums+"/sha256-bsd.sum"),
			},
			want{
				"content.txt",
				"47afcdfff05a6e5d9db5f6c6df2140f04a6e7422d7ad7f6a7006a4f5a78570e4",
				"sha256",
				sha256.New(),
				false,
			},
		},
		{"sha512",
			args{
				u(t, checksums+"/content.txt?checksum=sha512:060a8cc41c501e41b4537029661090597aeb4366702ac3cae8959f24b2c49005d6bd339833ebbeb481b127ac822d70b937c1637c8d0eaf81b6979d4c1d75d0e1"),
			},
			want{
				"content.txt",
				"060a8cc41c501e41b4537029661090597aeb4366702ac3cae8959f24b2c49005d6bd339833ebbeb481b127ac822d70b937c1637c8d0eaf81b6979d4c1d75d0e1",
				"sha512",
				sha512.New(),
				false,
			},
		},
		{"file sha512",
			args{
				u(t, checksums+"/content.txt?checksum=file:"+checksums+"/sha512-p.sum"),
			},
			want{
				"?content.txt",
				"060a8cc41c501e41b4537029661090597aeb4366702ac3cae8959f24b2c49005d6bd339833ebbeb481b127ac822d70b937c1637c8d0eaf81b6979d4c1d75d0e1",
				"sha512",
				sha512.New(),
				false,
			},
		},
		{"file sha512 BSD",
			args{
				u(t, checksums+"/content.txt?checksum=file:"+checksums+"/sha512-bsd.sum"),
			},
			want{
				"content.txt",
				"060a8cc41c501e41b4537029661090597aeb4366702ac3cae8959f24b2c49005d6bd339833ebbeb481b127ac822d70b937c1637c8d0eaf81b6979d4c1d75d0e1",
				"sha512",
				sha512.New(),
				false,
			},
		},
		{"md5",
			args{
				u(t, checksums+"/content.txt?checksum=md5:074729f0ccb41a391fb646c38f86ea54"),
			},
			want{
				"content.txt",
				"074729f0ccb41a391fb646c38f86ea54",
				"md5",
				md5.New(),
				false,
			},
		},
		{"file md5",
			args{
				u(t, checksums+"/content.txt?checksum=file:"+checksums+"/md5-p.sum"),
			},
			want{
				"content.txt",
				"074729f0ccb41a391fb646c38f86ea54",
				"md5",
				md5.New(),
				false,
			},
		},
		{"file md5 BSD",
			args{
				u(t, checksums+"/content.txt?checksum=file:"+checksums+"/md5-bsd.sum"),
			},
			want{
				"content.txt",
				"074729f0ccb41a391fb646c38f86ea54",
				"md5",
				md5.New(),
				false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFileChecksum, err := checksumHashAndValue(tt.args.u)
			if (err != nil) != tt.want.Err {
				t.Errorf("checksumHashAndValue() error = %v , wantErr %v", err, tt.want.Err)
				return
			}
			wantFileChecksum := &fileChecksum{
				Filename: tt.want.Filename,
				Hash:     tt.want.Hash,
				Type:     tt.want.Type,
			}
			if tt.want.Checksum != "" {
				if wantFileChecksum.Value, err = hex.DecodeString(tt.want.Checksum); err != nil {
					panic(err)
				}
			}
			if !reflect.DeepEqual(gotFileChecksum, wantFileChecksum) {
				t.Errorf("checksumHashAndValue() gotFileChecksum = %#v, want %#v", gotFileChecksum, wantFileChecksum)
			}

		})
	}
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
