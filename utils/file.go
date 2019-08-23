package utils

import (
	"errors"
	"os"
	"path"
)

// FileExists returns whether the file exists.
func FileExist(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// IsDir returns whether the name is a directory.
func IsDir(name string) (bool, error) {
	fi, err := os.Stat(name)
	if err != nil {
		return false, err
	} else if fi.IsDir() {
		return true, nil
	}

	return false, nil
}

// IsFile Returns whether the name is a file.
func IsFile(name string) (bool, error) {
	r, err := IsDir(name)
	return !r, err
}

// MkDir new folder.
// If existName is true the path with a file name, delete file name when creating.
func MkDir(dir string, perm os.FileMode, existName bool) error {
	if existName {
		dir = path.Dir(dir)
	}
	if len(dir) == 0 {
		return errors.New("dir is empty")
	}

	return os.MkdirAll(dir, perm)
}

// WriteFile write data to file.
// flag value: os.O_RDWR、os.O_CREATE、os.O_APPEND ...
func WriteFile(name string, data []byte, flag int, perm os.FileMode) (int, error) {
	fd, err := os.OpenFile(name, flag, perm)
	if err != nil {
		return 0, err
	}
	defer fd.Close()

	return fd.Write(data)
}
