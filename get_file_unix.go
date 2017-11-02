// +build !windows

package getter

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
)

func (g *FileGetter) Get(dst string, u *url.URL) error {
	path := u.Path
	if u.RawPath != "" {
		path = u.RawPath
	}

	// The source path must exist and be a directory to be usable.
	if fi, err := os.Stat(path); err != nil {
		return fmt.Errorf("source path error: %s", err)
	} else if !fi.IsDir() {
		return fmt.Errorf("source path must be a directory")
	}
	g.totalSize = fi.Size()

	fi, err := os.Lstat(dst)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// If the destination already exists, it must be a symlink
	if err == nil {
		mode := fi.Mode()
		if mode&os.ModeSymlink == 0 {
			return fmt.Errorf("destination exists and is not a symlink")
		}

		// Remove the destination
		if err := os.Remove(dst); err != nil {
			return err
		}
	}

	// Create all the parent directories
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	g.PercentComplete = 100
	return os.Symlink(path, dst)
}

func (g *FileGetter) GetFile(dst string, u *url.URL) error {
	path := u.Path
	if u.RawPath != "" {
		path = u.RawPath
	}

	// The source path must exist and be a file to be usable.
	if fi, err := os.Stat(path); err != nil {
		return fmt.Errorf("source path error: %s", err)
	} else if fi.IsDir() {
		return fmt.Errorf("source path must be a file")
	}
	g.totalSize = fi.Size()

	_, err := os.Lstat(dst)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// If the destination already exists, it must be a symlink
	if err == nil {
		// Remove the destination
		if err := os.Remove(dst); err != nil {
			return err
		}
	}

	// Create all the parent directories
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	// If we're not copying, just symlink and we're done
	if !g.Copy {
		g.PercentComplete = 100
		return os.Symlink(path, dst)
	}

	// Copy
	srcF, err := os.Open(path)
	if err != nil {
		return err
	}
	defer srcF.Close()

	dstF, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstF.Close()

	go g.CalcPercentComplete(dst)

	nwritten, err = io.Copy(dstF, srcF)
	g.Done <- nwritten
	return err
}

func (g *FileGetter) GetProgress() int {
	return g.PercentComplete
}

func (g *HttpGetter) CalcPercentComplete(dst) {
	// stat file every n seconds to figure out the download progress
	var stop bool = false
	dstfile, err := os.Open(dst)
	defer dstfile.Close()

	if err != nil {
		log.Printf("couldn't open file for reading: %s", err)
		return
	}
	for {
		select {
		case <-g.Done:
			stop = true
		default:
			fi, err := dstfile.Stat()
			if err != nil {
				fmt.Printf("Error stating file: %s", err)
				return
			}
			size := fi.Size()

			// catch edge case that would break our percentage calc
			if size == 0 {
				size = 1
			}

			g.PercentComplete = int(float64(size) / float64(g.totalSize) * 100)
		}

		if stop {
			break
		}
		// repeat check once per second
		time.Sleep(time.Second)
	}
}
