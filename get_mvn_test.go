package getter

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMvnGetter_impl(t *testing.T) {
	var _ Getter = new(MvnGetter)
}

func TestMvnGetter_artifact_from_central_maven_repo(t *testing.T) {
	mvnGetter_artifact(t, "https://repo1.maven.org/maven2", "6.13.1", "")
}

func TestMvnGetter_artifact_from_jcenter_repo(t *testing.T) {
	mvnGetter_artifact(t, "http://jcenter.bintray.com", "6.13.1", "")
}

func TestMvnGetter_artifact_with_classifier(t *testing.T) {
	mvnGetter_artifact(t, "https://repo1.maven.org/maven2", "6.13.1", "sources")
}

func TestMvnGetter_snapshot_artifact(t *testing.T) {
	mvnGetter_artifact(t, "https://oss.sonatype.org/content/repositories/snapshots", "6.13-SNAPSHOT", "")
}

func mvnGetter_artifact(t *testing.T, repo, version, classifier string) {
	mvnGetter := new(MvnGetter)
	groupId := "org.testng"
	artifactId := "testng"
	filename := artifactId + "-" + version
	if classifier != "" {
		filename += "-" + classifier
	}
	filename += ".jar"
	dstDir := tempDir(t)
	dst := filepath.Join(dstDir, filename)
	defer os.RemoveAll(dstDir)

	var artifactUrlStr string
	if classifier == "" {
		artifactUrlStr = fmt.Sprintf("%s?groupId=%s&artifactId=%s&version=%s", repo, groupId, artifactId, version)
	} else {
		artifactUrlStr = fmt.Sprintf("%s?groupId=%s&artifactId=%s&version=%s&classifier=%s", repo, groupId, artifactId, version, classifier)
	}
	artifactUrl, err := url.Parse(artifactUrlStr)
	if err != nil {
		t.Fatalf("parse artifact url failed: %s", err)
	}

	// Get the artifact
	if err := mvnGetter.GetFile(dst, artifactUrl); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	if _, err := os.Stat(dst); err != nil {
		t.Fatalf("err: %s", err)
	}

	artifactFileVer := version
	if strings.HasSuffix(version, "-SNAPSHOT") {
		snapshotUrlStr := fmt.Sprintf("%s/org/testng/%s/%s", repo, artifactId, version)
		snapshotUrl, err := url.Parse(snapshotUrlStr)
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		snapshotVer, err := mvnGetter.parseLastestSnapshotVersion(snapshotUrl)
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		artifactFileVer = snapshotVer
	}

	// verify the jar file md5
	httpGetter := new(HttpGetter)
	var md5UrlStr string
	if classifier == "" {
		md5UrlStr = fmt.Sprintf("%s/org/testng/%s/%s/%s-%s.jar.md5", repo, artifactId, version, artifactId, artifactFileVer)
	} else {
		md5UrlStr = fmt.Sprintf("%s/org/testng/%s/%s/%s-%s-%s.jar.md5", repo, artifactId, version, artifactId, artifactFileVer, classifier)
	}
	md5Url, err := url.Parse(md5UrlStr)
	if err != nil {
		t.Fatalf("parse artifact md5 file url failed: %s", err)
	}
	md5File := tempFile(t)
	defer os.Remove(md5File)
	if err := httpGetter.GetFile(md5File, md5Url); err != nil {
		t.Fatalf("download artifact md5 file failed: %s", err)
	}

	actualMd5, err := hashFileMd5(dst)
	if err != nil {
		t.Fatalf("compute artifact md5 failed: %s", err)
	}

	md5Bytes, err := ioutil.ReadFile(md5File)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	expectedMd5 := string(md5Bytes)
	if actualMd5 != expectedMd5 {
		t.Errorf("the hashing not match: %s != %s", actualMd5, expectedMd5)
	}
}

func hashFileMd5(filePath string) (string, error) {
	//Initialize variable returnMD5String now in case an error has to be returned
	var returnMD5String string

	//Open the passed argument and check for any error
	file, err := os.Open(filePath)
	if err != nil {
		return returnMD5String, err
	}

	//Tell the program to call the following function when the current function returns
	defer file.Close()

	//Open a new hash interface to write to
	hash := md5.New()

	//Copy the file in the hash interface and check for any error
	if _, err := io.Copy(hash, file); err != nil {
		return returnMD5String, err
	}

	//Get the 16 bytes hash
	hashInBytes := hash.Sum(nil)[:16]

	//Convert the bytes to a string
	returnMD5String = hex.EncodeToString(hashInBytes)

	return returnMD5String, nil
}
