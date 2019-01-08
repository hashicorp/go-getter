package getter

import (
	"io/ioutil"
)

func tempFile(dir, pattern string) (string, error) {
	f, err := ioutil.TempFile(dir, pattern)
	if err != nil {
		return "", err
	}
	return f.Name(), nil
}
