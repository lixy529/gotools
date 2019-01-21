package utils

import (
	"errors"
	"os"
	"path"
	"fmt"
	"syscall"
	"reflect"
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

// FileCtime returns the creation time of the file.
func FileCtime(path string) (int64, int64, error) {
	fi, err := os.Stat(path)
	if err != nil {
		fmt.Println("Stat:", err)
		return -1, -1, err
	}

	sysInfo := fi.Sys()
	if stat, ok := sysInfo.(*syscall.Stat_t); ok {
		// linux use Ctim, mac use Ctimespec
		//return stat.Ctimespec.Sec, stat.Ctimespec.Nsec, nil
		//return stat.Ctim.Sec, stat.Ctim.Nsec, nil
		// 为了兼容使用下面反射处理
		elem := reflect.ValueOf(stat).Elem()
		type_ := elem.Type()
		for i := 0; i < type_.NumField(); i++ {
			fieldName := type_.Field(i).Name
			if fieldName == "Ctimespec" || fieldName == "Ctim" {
				ctim := elem.Field(i).Interface().(syscall.Timespec)
				return ctim.Sec, ctim.Nsec, nil
			}
		}
		return -1, -1, errors.New("Not found create time field")
	}

	return -1, -1, errors.New("Assertion error")
}
