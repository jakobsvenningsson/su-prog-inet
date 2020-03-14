package client_connection

import (
	"os"
	"path/filepath"
)

type ftpDirPath struct {
	root    string
	current string
}

func (fd *ftpDirPath) path() string {
	return fd.root + fd.current
}

func (fd *ftpDirPath) exist(path string) bool {
	return fileExist(filepath.Join(fd.root, path))
}

func fileExist(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}
