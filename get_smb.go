package getter

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

// SmbGetter is a Getter implementation that will download a module from
// a shared folder using samba scheme.
type SmbGetter struct {
	getter
}

const basePathError = "samba path should contain valid Host and filepath (smb://<host>/<file_path>)"

func (g *SmbGetter) Mode(ctx context.Context, u *url.URL) (Mode, error) {
	// TODO: validate mode from smb path instead of stat
	return ModeFile, nil
}

// TODO: also copy directory
func (g *SmbGetter) Get(ctx context.Context, req *Request) error {
	hostPath, filePath, err := g.findHostAndFilePath(req)

	if err == nil {
		err = g.smbclientGetFile(hostPath, filePath, req)
		if err == nil {
			return nil
		}
	}
	os.RemoveAll(req.Dst)

	if err.Error() == basePathError {
		return err
	}

	// Look for local mount of shared folder
	if runtime.GOOS == "linux" {
		hostPath = strings.TrimPrefix(hostPath, "/")
	}
	err = get(hostPath, req)
	if err == nil {
		return nil
	}

	// TODO throw error msg to install smbclient or mount shared folder
	return err
}

func (g *SmbGetter) GetFile(ctx context.Context, req *Request) error {
	hostPath, filePath, err := g.findHostAndFilePath(req)
	if err == nil {
		err = g.smbclientGetFile(hostPath, filePath, req)
	}

	if err != nil && err.Error() == basePathError {
		return err
	}

	// Look for local mount of shared folder
	if err != nil && runtime.GOOS == "windows" {
		err = getFile(hostPath, req, ctx)
	}

	// throw error msg to install smbclient or mount shared folder
	return err
}

func (g *SmbGetter) findHostAndFilePath(req *Request) (string, string, error) {
	if req.u.Host == "" || req.u.Path == "" {
		return "", "", fmt.Errorf(basePathError)
	}
	// Host path
	hostPath := "//" + req.u.Host

	// Get shared directory
	path := strings.TrimPrefix(req.u.Path, "/")
	splt := regexp.MustCompile(`/`)
	directories := splt.Split(path, 2)

	if len(directories) > 0 {
		hostPath = hostPath + "/" + directories[0]
	}

	// Check file path
	if len(directories) <= 1 || directories[1] == "" {
		return "", "", fmt.Errorf("can not find file path and/or name in the smb url")
	}

	return hostPath, directories[1], nil
}

func (g *SmbGetter) smbclientGetFile(hostPath string, fileDir string, req *Request) error {
	file := ""
	if strings.Contains(fileDir, "/") {
		i := strings.LastIndex(fileDir, "/")
		file = fileDir[i+1:]
		fileDir = fileDir[:i]
	} else {
		file = fileDir
		fileDir = "."
	}

	smbcmd := "smbclient -N"

	// Get auth user and password
	auth := req.u.User.Username()
	if auth != "" {
		if password, ok := req.u.User.Password(); ok {
			auth = auth + "%" + password
		}
		smbcmd = smbcmd + " -U " + auth
	}

	getFile := fmt.Sprintf("'get %s'", file)
	smbcmd = smbcmd + " " + hostPath + " --directory " + fileDir + " --command " + getFile
	cmd := exec.Command("bash", "-c", smbcmd)

	if req.Dst != "" {
		_, err := os.Lstat(req.Dst)
		if err != nil {
			if os.IsNotExist(err) {
				// Create destination folder if it doesn't exists
				if err := os.MkdirAll(req.Dst, os.ModePerm); err != nil {
					return fmt.Errorf("failed to creat destination path: %s", err.Error())
				}
			} else {
				return err
			}
		}
		cmd.Dir = req.Dst
	}

	// Execute smbclient command
	return getRunCommand(cmd)
}
