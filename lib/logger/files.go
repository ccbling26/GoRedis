package logger

import (
	"fmt"
	"os"
)

func CheckNotExist(name string) bool {
	_, err := os.Stat(name)
	return os.IsNotExist(err)
}

func CheckPermission(name string) bool {
	_, err := os.Stat(name)
	return os.IsPermission(err)
}

func MkDir(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}

func IsNotExistMkDir(dir string) error {
	if CheckNotExist(dir) {
		return MkDir(dir)
	}
	return nil
}

func MustOpen(filename, dir string) (*os.File, error) {
	if CheckPermission(dir) {
		return nil, fmt.Errorf("Permission denied dir: %s", dir)
	}
	if err := IsNotExistMkDir(dir); err != nil {
		return nil, fmt.Errorf("Error during make dir %s, err: %s", dir, err)
	}
	f, err := os.OpenFile(dir+string(os.PathSeparator)+filename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("Fail to open file, err: %s", err)
	}
	return f, nil
}
