package getter

import (
	"context"
	"testing"
)

func TestClient_ChecksumFromFile(t *testing.T) {
	ctx := context.TODO()
	checksumURL := "http://archive.ubuntu.com/ubuntu/dists/bionic-updates/main/installer-amd64/current/images/SHA256SUMS"
	isoURL := "http://archive.ubuntu.com/ubuntu/dists/bionic-updates/main/installer-amd64/current/images/netboot/mini.iso"

	client := Client{}
	file, err := client.ChecksumFromFile(ctx, checksumURL, isoURL)

	if err != nil {
		t.Fatalf("bad: should not have errored: %s", err.Error())
	}
	if file.Filename != "./netboot/mini.iso" {
		t.Fatalf("bad: not expected filename: %s", file.Filename)
	}
}

