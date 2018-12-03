package getter

import "testing"

func Test_parseChecksumLine(t *testing.T) {
	type args struct {
		checksumFile string
		line         string
	}
	tests := []struct {
		name              string
		args              args
		wantChecksumType  string
		wantChecksumValue string
		wantFilename      string
		wantOk            bool
	}{
		{"gnu SHA256SUMS",
			args{
				"http://old-releases.ubuntu.com/releases/14.04.1/SHA512SUMS",
				"d9a217e80fb6dc2576d5ccca4c44376c25e6016de47a48e07345678d660fac51 *ubuntu-14.04-desktop-amd64+mac.iso",
			},
			"sha512",
			"d9a217e80fb6dc2576d5ccca4c44376c25e6016de47a48e07345678d660fac51",
			"*ubuntu-14.04-desktop-amd64+mac.iso",
			true,
		},
		{"bsd CHECKSUM.SHA256",
			args{
				"ftp://ftp.freebsd.org/pub/FreeBSD/snapshots/VM-IMAGES/10.4-STABLE/i386/Latest/CHECKSUM.SHA256",
				"SHA256 (FreeBSD-10.4-STABLE-i386-20181012-r339297.qcow2.xz) = cedf5203525ef1c7048631d7d26ca54b81f224fccf6b9185eab2cf4b894e8651",
			},
			"sha256",
			"cedf5203525ef1c7048631d7d26ca54b81f224fccf6b9185eab2cf4b894e8651",
			"FreeBSD-10.4-STABLE-i386-20181012-r339297.qcow2.xz",
			true,
		},
		{"potato",
			args{
				"blip",
				"potato chips 3 4",
			},
			"",
			"",
			"",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotChecksumType, gotChecksumValue, gotFilename, gotOk := parseChecksumLine(tt.args.checksumFile, tt.args.line)
			if gotChecksumType != tt.wantChecksumType {
				t.Errorf("parseChecksumLine() gotChecksumType = %v, want %v", gotChecksumType, tt.wantChecksumType)
			}
			if gotChecksumValue != tt.wantChecksumValue {
				t.Errorf("parseChecksumLine() gotChecksumValue = %v, want %v", gotChecksumValue, tt.wantChecksumValue)
			}
			if gotFilename != tt.wantFilename {
				t.Errorf("parseChecksumLine() gotFilename = %v, want %v", gotFilename, tt.wantFilename)
			}
			if gotOk != tt.wantOk {
				t.Errorf("parseChecksumLine() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}
