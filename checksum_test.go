package getter

import (
	"context"
	"testing"
)

func TestClient_ChecksumFromFileWithSubFolder(t *testing.T) {
	httpChecksums := httpTestModule("checksum-file")
	defer httpChecksums.Close()
	ctx := context.TODO()
	isoURL := "http://hashicorp.com/ubuntu/dists/bionic-updates/main/installer-amd64/current/images/netboot/mini.iso"

	client := Client{}
	file, err := client.ChecksumFromFile(ctx, httpChecksums.URL+"/sha256-subfolder.sum", isoURL)

	if err != nil {
		t.Fatalf("bad: should not have error: %s", err.Error())
	}
	if file.Filename != "./netboot/mini.iso" {
		t.Fatalf("bad: expecting filename ./netboot/mini.iso but was: %s", file.Filename)
	}
}
