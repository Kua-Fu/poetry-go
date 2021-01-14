package store

import (
	"fmt"
	"os"
)

// FSDirectory return File
// check file exist and is a directory
func FSDirectory(filePath string) (*os.File, error) {

	info, err := os.Stat(filePath)

	if os.IsNotExist(err) {
		return nil, fmt.Errorf(filePath + "does not exist.")
	}
	if !info.IsDir() {
		return nil, fmt.Errorf(filePath + "is not a directory.")
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	return file, nil

}
