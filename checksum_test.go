package getter

import (
	"encoding/hex"
	"hash"
	"net/url"
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
	tests := []struct {
		name              string
		args              args
		wantChecksumHash  hash.Hash
		wantChecksumValue string
		wantErr           bool
	}{
		{"shasum -a 256 -p",
			args{u(t, checksums+"/content.txt?checksum=file:"+checksums+"/sha256-p.sum")},
			checksummers["sha256"](),
			"47afcdfff05a6e5d9db5f6c6df2140f04a6e7422d7ad7f6a7006a4f5a78570e4",
			false,
		},
		{"not properly named shasum -a 256 -p",
			args{u(t, checksums+"/content.txt?checksum=file:"+checksums+"/sha2FiveSixError.sum")},
			nil,
			"",
			true,
		},
		{"md5",
			args{u(t, checksums+"/content.txt?checksum=file:"+checksums+"/sha256-p.sum")},
			checksummers["sha256"](),
			"47afcdfff05a6e5d9db5f6c6df2140f04a6e7422d7ad7f6a7006a4f5a78570e4",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotChecksumHash, gotChecksumValue, err := checksumHashAndValue(tt.args.u)
			if (err != nil) != tt.wantErr {
				t.Errorf("checksumHashAndValue() error = %v , wantErr %v", err, tt.wantErr)
				return
			}
			var wantChecksumValue []byte
			if tt.wantChecksumValue != "" {
				if wantChecksumValue, err = hex.DecodeString(tt.wantChecksumValue); err != nil {
					panic(err)
				}
			}
			if !reflect.DeepEqual(gotChecksumHash, tt.wantChecksumHash) {
				t.Errorf("checksumHashAndValue() gotChecksumHash = %v, want %v", gotChecksumHash, tt.wantChecksumHash)
			}
			if !reflect.DeepEqual(gotChecksumValue, wantChecksumValue) {
				t.Errorf("checksumHashAndValue() gotChecksumValue = %v, want %v", gotChecksumValue, tt.wantChecksumValue)
			}
		})
	}
}
