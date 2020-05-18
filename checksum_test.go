package getter

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestClient_ChecksumFromFileWithSubFolder(t *testing.T) {
	httpChecksums := httpTestModule("checksum-file")
	defer httpChecksums.Close()
	ctx := context.TODO()
	isoURL := "http://hashicorp.com/ubuntu/dists/bionic-updates/main/installer-amd64/current/images/netboot/mini.iso"

	client := Client{}
	file, err := client.checksumFromFile(ctx, httpChecksums.URL+"/sha256-subfolder.sum", isoURL, "")

	if err != nil {
		t.Fatalf("bad: should not have error: %s", err.Error())
	}
	if file.Filename != "./netboot/mini.iso" {
		t.Fatalf("bad: expecting filename ./netboot/mini.iso but was: %s", file.Filename)
	}
}

func TestClient_GetChecksum(t *testing.T) {
	// Creates checksum file in local dir
	p := filepath.Join(fixtureDir, "checksum-file/sha256-subfolder.sum")
	source, err := os.Open(p)
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer source.Close()
	destination, err := os.Create("local.sum")
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer os.Remove("local.sum")
	defer destination.Close()
	if _, err := io.Copy(destination, source); err != nil {
		t.Fatalf(err.Error())
	}

	client := Client{}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf(err.Error())
	}
	req := &Request{
		Src: "http://hashicorp.com/ubuntu/dists/bionic-updates/main/installer-amd64/current/images/netboot/mini.iso?checksum=file:./local.sum",
		Pwd: wd,
	}
	file, err := client.GetChecksum(context.TODO(), req)

	if err != nil {
		t.Fatalf("bad: should not have error: %s", err.Error())
	}
	if file.Filename != "./netboot/mini.iso" {
		t.Fatalf("bad: expecting filename ./netboot/mini.iso but was: %s", file.Filename)
	}
}

func TestFileChecksum_String(t *testing.T) {
	type fields struct {
		checksum string
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf(err.Error())
	}
	tests := []struct {
		fields fields
		want   string
	}{
		{fields{"090992ba9fd140077b0661cb75f7ce13"}, "md5:090992ba9fd140077b0661cb75f7ce13"},
		{fields{"ebfb681885ddf1234c18094a45bbeafd91467911"}, "sha1:ebfb681885ddf1234c18094a45bbeafd91467911"},
		{fields{"sha256:ed363350696a726b7932db864dda019bd2017365c9e299627830f06954643f93"}, "sha256:ed363350696a726b7932db864dda019bd2017365c9e299627830f06954643f93"},
		{fields{"file:" + filepath.Join(wd, fixtureDir, "checksum-file", "sha1.sum")}, "sha1:e2c7dc83ac8aa7f181314387f6dfb132cd117e3a"},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			req := &Request{
				Src: "http://example.dev?checksum=" + tt.fields.checksum,
			}
			c, err := DefaultClient.GetChecksum(context.TODO(), req)
			if err != nil {
				t.Fatalf("GetChecksum: %v", err)
			}

			if got := c.String(); got != tt.want {
				t.Errorf("FileChecksum.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
