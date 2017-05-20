package getter

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"regexp"
	"strconv"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// SftpGetter is a Getter implementation that will download a file through sftp
// uri format: sftp://[username@]hostname[:port]/directoryname[?options]
// see also: http://camel.apache.org/ftp2.html
type SftpGetter struct{}

func (g *SftpGetter) ClientMode(u *url.URL) (ClientMode, error) {
	sftp, err := g.createSftpClient(u)
	if err != nil {
		return ClientModeInvalid, err
	}
	defer sftp.Close()

	fileInfo, err := sftp.Stat(u.Path)
	if err != nil {
		return ClientModeInvalid, err
	}
	if fileInfo.IsDir() {
		return ClientModeDir, nil
	}
	return ClientModeFile, nil
}

// Get the files under the remote dir.
// Note: recursively download is not supported at the moment.
// Query parameters:
//   - fileName: the name of the file to download, support regex
//   - preservePermissions: true to preserve the file permissions on local file, default as false
// example url: sftp://username@host/the/remote/dir?fileName=f1.txt&fileName=f2.txt&fileName=.*\.txt
func (g *SftpGetter) Get(dst string, u *url.URL) error {
	sftp, err := g.createSftpClient(u)
	if err != nil {
		return err
	}
	defer sftp.Close()

	rmtDir := u.Path
	rmtFiles, err := sftp.ReadDir(rmtDir)
	if err != nil {
		return err
	}

	var targetFiles []string
	if fileNames, hasFileName := u.Query()["fileName"]; hasFileName {
		for _, fileName := range fileNames {
			re, err := regexp.Compile(fileName)
			if err != nil {
				continue
			}
			for _, rmtFile := range rmtFiles {
				if !rmtFile.IsDir() && re.FindString(rmtFile.Name()) != "" {
					targetFiles = append(targetFiles, rmtFile.Name())
				}
			}
		}
	} else {
		for _, rmtFile := range rmtFiles {
			if !rmtFile.IsDir() {
				targetFiles = append(targetFiles, rmtFile.Name())
			}
		}
	}

	preservePerm, _ := strconv.ParseBool(u.Query().Get("preservePermissions"))
	for _, file := range targetFiles {
		if err := g.getFile(sftp, dst+"/"+file, rmtDir+"/"+file, preservePerm); err != nil {
			return err
		}
	}

	return nil
}

// Get the remote file.
// Query parameters:
//   - preservePermissions: true to preserve the file permissions on local file, default as false
// example url: sftp://username@host/the/remote/file.txt
func (g *SftpGetter) GetFile(dst string, u *url.URL) error {
	sftp, err := g.createSftpClient(u)
	if err != nil {
		return err
	}
	defer sftp.Close()

	preservePerm, _ := strconv.ParseBool(u.Query().Get("preservePermissions"))
	return g.getFile(sftp, dst, u.Path, preservePerm)
}

func (g *SftpGetter) createSftpClient(u *url.URL) (*sftp.Client, error) {
	var authMethods []ssh.AuthMethod
	idRsaFile, _ := homedir.Expand("~/.ssh/id_rsa")
	idDsaFile, _ := homedir.Expand("~/.ssh/id_dsa")
	potentialKeyFiles := []string{u.Query().Get("privateKeyFile"), idRsaFile, idDsaFile}
	for _, keyFile := range potentialKeyFiles {
		if keyFile != "" && exists(keyFile) {
			key, err := g.getKeyFile(keyFile)
			if err != nil {
				log.Printf("failed to parse private key [%s]: %v", keyFile, err)
			} else {
				authMethods = append(authMethods, ssh.PublicKeys(key))
			}
		}
	}
	if passwd, hasPasswd := u.User.Password(); hasPasswd {
		authMethods = append(authMethods, ssh.Password(passwd))
	} else if passwd := u.Query().Get("password"); passwd != "" {
		authMethods = append(authMethods, ssh.Password(passwd))
	}

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("either password or private key is required for ssh auth.")
	}

	user := u.User.Username()
	config := &ssh.ClientConfig{
		User: user,
		Auth: authMethods,
	}
	var port = u.Port()
	if port == "" {
		port = "22"
	}
	client, err := ssh.Dial("tcp", u.Hostname()+":"+port, config)
	if err != nil {
		return nil, err
	}

	sftp, err := sftp.NewClient(client)
	return sftp, err
}

func (g *SftpGetter) getKeyFile(file string) (key ssh.Signer, err error) {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	key, err = ssh.ParsePrivateKey(buffer)
	return key, err
}

func (g *SftpGetter) getFile(sftp *sftp.Client, dst, src string, preservePerm bool) error {
	rmtFile, err := sftp.Open(src)
	if err != nil {
		return err
	}
	defer rmtFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}

	log.Printf("Downloading remote %s to local %s", src, dst)
	_, err = rmtFile.WriteTo(dstFile)
	dstFile.Close()
	if err != nil {
		return err
	}

	if preservePerm {
		// Chmod the file
		rmtFileInfo, err := rmtFile.Stat()
		if err != nil {
			return err
		}
		if err := os.Chmod(dst, rmtFileInfo.Mode()); err != nil {
			return err
		}
	}

	return nil
}

func exists(filePath string) (exists bool) {
	exists = true

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		exists = false
	}

	return
}
