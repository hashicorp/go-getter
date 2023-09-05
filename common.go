package getter

import (
	"os"
)

func tmpFile(dir, pattern string) (string, error) {
	f, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return "", err
	}
	f.Close()
	return f.Name(), nil
}
