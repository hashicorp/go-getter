package getter

import (
	"io/ioutil"
)

func tmpFile(dir, pattern string) (string, error) {
	f, err := ioutil.TempFile(dir, pattern)
	if err != nil {
		return "", err
	}
	tempfile := f.Name()
	f.Close()
	return tempfile, nil
}
