package getter

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// MvnGetter is a Getter implementation that will download an artifact from maven repository, e.g. Sonatype Nexus,
// uri format: mvn::http://[username@]hostname[:port]/directoryname[?options]
type MvnGetter struct {
	HttpGet HttpGetter
}

func (g *MvnGetter) ClientMode(u *url.URL) (ClientMode, error) {
	return ClientModeFile, nil
}

// GetFilename returns filename in format of '<artifactId>-<version>[-<classifier>].<type>'
func (g *MvnGetter) GetFilename(u *url.URL) (string, error) {
	q := u.Query()
	artifactId := q.Get("artifactId")
	if artifactId == "" {
		return "", fmt.Errorf("query parameter 'artifactId' is required.")
	}

	version := q.Get("version")
	if version == "" {
		return "", fmt.Errorf("query parameter 'version' is required.")
	}

	classifier := q.Get("classifier")

	artType := q.Get("type")
	if artType == "" {
		artType = "jar"
	}

	filename := artifactId + "-" + version
	if classifier != "" {
		filename += "-" + classifier
	}
	filename += "." + artType
	filename = filepath.Clean(filename)

	return filename, nil
}

func (g *MvnGetter) Get(dstDir string, u *url.URL) error {
	filename, err := g.GetFilename(u)
	if err != nil {
		return err
	}

	dstFile := filepath.Join(dstDir, filename)
	return g.GetFile(dstFile, u)
}

// Get the artifact from remote maven repo.
//
// If the version is a snapshot version, it will get the latest snapshot artifact.
// Query parameters:
//   - groupId: the group id
//   - artifactId: the artifact id
//   - version: the artifact version
//   - type: the artifact type, default as 'jar'
// example url: mvn::http://username@host/mavan/repo/path?groupId=org.example&artifactId=test&version=1.0.0-SNAPSHOT
func (g *MvnGetter) GetFile(dst string, u *url.URL) error {
	q := u.Query()
	groupId := q.Get("groupId")
	if groupId == "" {
		return fmt.Errorf("query parameter 'groupId' is required.")
	}
	artifactId := q.Get("artifactId")
	if artifactId == "" {
		return fmt.Errorf("query parameter 'artifactId' is required.")
	}
	// the artifact version, Ex., 6.13.1 or 6.13-SNAPSHOT
	version := q.Get("version")
	if version == "" {
		return fmt.Errorf("query parameter 'version' is required.")
	}
	classifier := q.Get("classifier")
	artType := q.Get("type")
	if artType == "" {
		artType = "jar"
	}

	// construct the real url hits the maven repo
	artifactUrl, err := url.Parse(u.String())
	if err != nil {
		return err
	}
	artifactUrl.RawQuery = ""
	artifactUrl.Path = path.Join(artifactUrl.Path, fmt.Sprintf("/%s/%s/%s", strings.Replace(groupId, ".", "/", -1), artifactId, version))

	// the artifact file version.
	//   when the artifact version is a snapshot version, the artifact file version will be expanded to the latest snapshot version, Ex., '6.13-20171126.202552-6'
	artifactFileVer := version
	if strings.HasSuffix(version, "-SNAPSHOT") {
		// get the latest snapshot
		snapshotVer, err := g.parseLastestSnapshotVersion(artifactUrl)
		if err != nil {
			return err
		}

		artifactFileVer = snapshotVer
	}

	filename := artifactId + "-" + artifactFileVer
	if classifier != "" {
		filename += "-" + classifier
	}
	filename += "." + artType
	artifactUrl.Path = path.Join(artifactUrl.Path, filename)

	return g.HttpGet.GetFile(dst, artifactUrl)
}

// get the latest snapshot version by parsig the maven-metadata.xml from remote maven repo.
//   - artifactVerUrl the url to the artifact version, Ex., 'https://repo1.maven.org/maven2/org/testng/testng/6.13.1/'
func (g *MvnGetter) parseLastestSnapshotVersion(artifactVerUrl *url.URL) (string, error) {
	mvnMetaUrl, err := url.Parse(artifactVerUrl.String())
	if err != nil {
		return "", err
	}
	mvnMetaUrl.Path = path.Join(mvnMetaUrl.Path, "maven-metadata.xml")

	mvnMetaFile, err := ioutil.TempFile("", "maven-metadata")
	if err != nil {
		return "", err
	}
	defer os.Remove(mvnMetaFile.Name())

	if err := g.HttpGet.GetFile(mvnMetaFile.Name(), mvnMetaUrl); err != nil {
		return "", err
	}

	mvnMetaXml, err := ioutil.ReadFile(mvnMetaFile.Name())
	if err != nil {
		return "", err
	}

	var meta Metadata
	if err := xml.Unmarshal(mvnMetaXml, &meta); err != nil {
		return "", err
	}
	vers := meta.Versioning.SnapshotVersions.VersionList
	if len(vers) == 0 {
		return "", fmt.Errorf("no snapshot versions in the %s", mvnMetaUrl)
	}
	return vers[0].Value, nil
}

type Metadata struct {
	GroupId    string            `xml:"groupId"`
	ArtifactId string            `xml:"artifactId"`
	Version    string            `xml:"version"`
	Versioning SnapshotVerioning `xml:"versioning"`
}
type SnapshotVerioning struct {
	SnapshotVersions SnapshotVersions `xml:"snapshotVersions"`
}
type SnapshotVersions struct {
	VersionList []SnapshotVersion `xml:"snapshotVersion"`
}
type SnapshotVersion struct {
	Extension string `xml:"extension"`
	Value     string `xml:"value"`
}
