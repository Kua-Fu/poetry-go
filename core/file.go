package core

import (
	"fmt"
	"os"
)

// File file overwrite os.File
type File struct {
	FilePath string
	File     *os.File
	IsDir    bool
}

// FSDirectory fs directory
type FSDirectory struct {
	DirPath string
	Dir     *File
}

// CreateFile create file
func CreateFile(filePath string, isDir bool, isCreate bool) (*File, error) {
	var fPtr *os.File
	var err error
	if isCreate { // file has create
		fPtr, err = os.Open(filePath)
	} else {
		fPtr, err = os.Create(filePath)
	}

	if err != nil {
		return nil, err
	}

	f := File{
		FilePath: filePath,
		File:     fPtr,
		IsDir:    isDir,
	}
	return &f, nil
}

// WriteInt64 write int64 to file
func (f *File) WriteInt64(n int64) error {
	b, err := Int64ToByte(n)
	if err != nil {
		return err
	}
	_, err = f.File.Write(b)
	if err != nil {
		return err
	}
	return nil
}

// WriteString write string
func (f *File) WriteString(s string) error {
	var (
		err error
	)
	_, err = f.File.WriteString(s)
	if err != nil {
		return err
	}
	return nil
}

// WriteByte write string
func (f *File) WriteByte(b byte) error {
	var (
		err error
	)
	_, err = f.File.Write([]byte{b})
	if err != nil {
		return err
	}
	return nil
}

// Init check file exist and is a directory
func (fs *FSDirectory) Init(dirPath string) error {

	info, err := os.Stat(dirPath)

	if os.IsNotExist(err) {
		return fmt.Errorf(dirPath + "does not exist.")
	}
	if !info.IsDir() {
		return fmt.Errorf(dirPath + "is not a directory.")
	}

	f, err := CreateFile(dirPath, true, true)
	if err != nil {
		return err
	}
	fs.DirPath = dirPath
	fs.Dir = f
	return nil
}
