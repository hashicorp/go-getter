package getter

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"syscall"
)

// SmbGetter is a Getter implementation that will download a module from
// a shared folder using samba scheme.
type SmbGetter struct {
	getter
}

const basePathError = "samba path should contain valid host, filepath, and authentication if necessary (smb://<user>:<password>@<host>/<file_path>)"

func (g *SmbGetter) Mode(ctx context.Context, u *url.URL) (Mode, error) {
	if u.Host == "" || u.Path == "" {
		return 0, fmt.Errorf(basePathError)
	}
	// Look in possible local mount of shared folder
	path := "/" + u.Host + u.Path
	if runtime.GOOS == "windows" {
		path = "/" + path
	}
	if m, err := mode(path); err == nil {
		return m, nil
	}

	// TODO write propper error msg
	return g.smbClientMode(u)
}

func (g *SmbGetter) smbClientMode(u *url.URL) (Mode, error) {
	hostPath, fileDir, err := g.findHostAndFilePath(u)
	if err != nil {
		return 0, err
	}
	file := ""
	// Get file and subdirectory name when existent
	if strings.Contains(fileDir, "/") {
		i := strings.LastIndex(fileDir, "/")
		file = fileDir[i+1:]
		fileDir = fileDir[:i]
	} else {
		file = fileDir
		fileDir = "."
	}

	getFileCmd := "smbclient -N"

	// Append auth user and password to cmd
	auth := u.User.Username()
	if auth != "" {
		if password, ok := u.User.Password(); ok {
			auth = auth + "%" + password
		}
		getFileCmd = getFileCmd + " -U " + auth
	}

	baseCmd := getFileCmd + " " + hostPath + " --directory " + fileDir

	// check file exists in the smb shared folder to check the mode
	isDir, err := isDirectory(baseCmd, file)
	if err != nil {
		return 0, err
	}
	if isDir {
		return ModeDir, nil
	}
	return ModeFile, nil
}

// TODO: copy directory
func (g *SmbGetter) Get(ctx context.Context, req *Request) error {
	if req.u.Host == "" || req.u.Path == "" {
		return fmt.Errorf(basePathError)
	}
	dstExisted := false
	if req.Dst != "" {
		if _, err := os.Lstat(req.Dst); err == nil {
			dstExisted = true
		}
	}

	// First look in  a possible local mount of the shared folder
	path := "/" + req.u.Host + req.u.Path
	if runtime.GOOS == "windows" {
		path = "/" + path
	}
	if err := get(path, req); err == nil {
		return nil
	}

	// Look for the file using smbclient cli
	err := g.smbclientGetFile(req)
	if err == nil {
		return nil
	}

	err = fmt.Errorf("one of the options should be available: \n 1. local mount of the smb shared folder or; \n 2. smbclient cli installed. \n err: %s", err.Error())

	if !dstExisted {
		// Remove the destination created for smbclient files
		if rerr := os.Remove(req.Dst); rerr != nil {
			err = fmt.Errorf("%s \n failed to remove created destination folder: %s", err.Error(), rerr.Error())
		}
	}

	return err
}

func (g *SmbGetter) GetFile(ctx context.Context, req *Request) error {
	if req.u.Host == "" || req.u.Path == "" {
		return fmt.Errorf(basePathError)
	}
	dstExisted := false
	if req.Dst != "" {
		if _, err := os.Lstat(req.Dst); err == nil {
			dstExisted = true
		}
	}

	// First look in a possible local mount of the shared folder
	path := "/" + req.u.Host + req.u.Path
	if runtime.GOOS == "windows" {
		path = "/" + path
	}
	if err := getFile(path, req, ctx); err == nil {
		return nil
	}

	// Look for the file using smbclient cli
	err := g.smbclientGetFile(req)
	if err == nil {
		return nil
	}

	err = fmt.Errorf("one of the options should be available: \n 1. local mount of the smb shared folder or; \n 2. smbclient cli installed. \n err: %s", err.Error())

	if !dstExisted {
		// Remove the destination created for smbclient files
		if rerr := os.Remove(req.Dst); rerr != nil {
			err = fmt.Errorf("%s \n failed to remove created destination folder: %s", err.Error(), rerr.Error())
		}
	}

	return err
}

func (g *SmbGetter) smbclientGetFile(req *Request) error {
	hostPath, fileDir, err := g.findHostAndFilePath(req.u)
	if err != nil {
		return err
	}
	file := ""
	// Get file and subdirectory name when existent
	if strings.Contains(fileDir, "/") {
		i := strings.LastIndex(fileDir, "/")
		file = fileDir[i+1:]
		fileDir = fileDir[:i]
	} else {
		file = fileDir
		fileDir = "."
	}

	getFileCmd := "smbclient -N"

	// Append auth user and password to cmd
	auth := req.u.User.Username()
	if auth != "" {
		if password, ok := req.u.User.Password(); ok {
			auth = auth + "%" + password
		}
		getFileCmd = getFileCmd + " -U " + auth
	}

	baseCmd := getFileCmd + " " + hostPath + " --directory " + fileDir

	// check file exists in the smb shared folder and is not a directory
	isDir, err := isDirectory(baseCmd, file)
	if err != nil {
		return err
	}
	if isDir {
		return fmt.Errorf("%s is a directory", file)
	}

	// download file
	getFileCmd = baseCmd + " --command " + fmt.Sprintf("'get %s'", file)
	cmd := exec.Command("bash", "-c", getFileCmd)

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

	_, err = runSmbClientCommand(cmd)
	return err
}

func (g *SmbGetter) findHostAndFilePath(u *url.URL) (string, string, error) {
	// Host path
	hostPath := "//" + u.Host

	// Get shared directory
	path := strings.TrimPrefix(u.Path, "/")
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

func isDirectory(baseCmd string, file string) (bool, error) {
	fileInfoCmd := baseCmd + " --command " + fmt.Sprintf("'allinfo %s'", file)
	cmd := exec.Command("bash", "-c", fileInfoCmd)
	output, err := runSmbClientCommand(cmd)
	if err != nil {
		return false, err
	}
	return strings.Contains(output, "attributes: D"), nil
}

func runSmbClientCommand(cmd *exec.Cmd) (string, error) {
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	if exiterr, ok := err.(*exec.ExitError); ok {
		// The program has exited with an exit code != 0
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			return buf.String(), fmt.Errorf(
				"%s exited with %d: %s",
				cmd.Path,
				status.ExitStatus(),
				buf.String())
		}
	}
	return buf.String(), nil
}
