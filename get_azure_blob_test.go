package getter

import (
	"log"
	"os"
	"path/filepath"
	"testing"
)

func TestAzureBlobGetter_Get(t *testing.T) {
	t.Parallel()

	ab := new(AzureBlobGetter)
	dst := t.TempDir()

	err := ab.Get(dst, testURL("https://stgogettertest.blob.core.windows.net/test"))
	if err != nil {
		log.Fatal(err)
	}

	// Verify the main file exists
	mainPath := filepath.Join(dst, "main.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatal(err)
	}

}

func TestAzureBlobGet_subdir(t *testing.T) {
	t.Parallel()

	ab := new(AzureBlobGetter)
	dst := t.TempDir()

	err := ab.Get(dst, testURL("https://stgogettertest.blob.core.windows.net/test/subdir"))
	if err != nil {
		t.Fatal(err)
	}

	// Verify the main file exists
	mainPath := filepath.Join(dst, "sub.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatal(err)
	}

}

func TestAzureBlobGetFile(t *testing.T) {
	t.Parallel()

	ab := new(AzureBlobGetter)
	dst := tempTestFile(t)

	err := ab.GetFile(dst, testURL("https://stgogettertest.blob.core.windows.net/test/main.tf"))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(dst); err != nil {
		t.Fatal(err)
	}
	assertContents(t, dst, "# Main")
}

func TestAzureBlobClientMode_dir(t *testing.T) {
	t.Parallel()

	g := new(AzureBlobGetter)

	// Check client mode on a key prefix with only a single key.
	mode, err := g.ClientMode(
		testURL("https://stgogettertest.blob.core.windows.net/test"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if mode != ClientModeDir {
		t.Fatal("expect ClientModeDir")
	}
}

func TestAzureBlobClientMode_dir_subdir(t *testing.T) {
	t.Parallel()

	g := new(AzureBlobGetter)

	// Check client mode on a key prefix with only a single key.
	mode, err := g.ClientMode(
		testURL("https://stgogettertest.blob.core.windows.net/test/subdir"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if mode != ClientModeDir {
		t.Fatal("expect ClientModeDir")
	}
}

func TestAzureBlobClientMode_file(t *testing.T) {
	t.Parallel()

	g := new(AzureBlobGetter)

	// Check client mode on a key prefix with only a single key.
	mode, err := g.ClientMode(
		testURL("https://stgogettertest.blob.core.windows.net/test/main.tf"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if mode != ClientModeFile {
		t.Fatal("expect ClientModeFile")
	}
}

func TestAzureBlobClientMode_file_subdir_file(t *testing.T) {
	t.Parallel()

	g := new(AzureBlobGetter)

	// Check client mode on a key prefix with only a single key.
	mode, err := g.ClientMode(
		testURL("https://stgogettertest.blob.core.windows.net/test/subdir/sub.tf"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if mode != ClientModeFile {
		t.Fatal("expect ClientModeFile")
	}
}
