package getter

import (
	"context"
	"io"
	"os"
	"path/filepath"
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
