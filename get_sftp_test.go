package getter

import (
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSftpGetter_impl(t *testing.T) {
	var _ Getter = new(SftpGetter)
}

func TestSftpGetter_file(t *testing.T) {
	g := new(SftpGetter)
	dst := filepath.Join(tempDir(t), "readme.txt")

	u, err := url.Parse("sftp://demo:password@test.rebex.net:22/readme.txt")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Get it!
	if err := g.GetFile(dst, u); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	if _, err := os.Stat(dst); err != nil {
		t.Fatalf("err: %s", err)
	}

	data, err := ioutil.ReadFile(dst)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	content := string(data)
	if !strings.Contains(content, "you are connected to an FTP or SFTP server used for testing purposes by Rebex FTP/SSL or Rebex SFTP sample code.") {
		t.Fatalf("download file does not contain expected content")
	}
}
