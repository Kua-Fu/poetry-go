package core

import (
	"fmt"
	"io/ioutil"
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

// CreateTempFile temp file
func CreateTempFile(filePath string, prefix string, isDir bool) (*File, error) {
	var (
		fPtr  *os.File
		err   error
		fPath string
		f     File
	)
	if isDir {
		fPath, err = ioutil.TempDir(filePath, prefix)
		if err != nil {
			return nil, err
		}
		fPtr, err = os.Open(filePath)
		if err != nil {
			return nil, err
		}
		f = File{
			FilePath: fPath,
			File:     fPtr,
			IsDir:    true,
		}

	}

	return &f, nil

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

// Rename rename file
func (f *File) Rename(filePath string) error {

	err := os.Rename(f.FilePath, filePath)
	if err != nil {
		return err
	}
	f.FilePath = filePath
	return nil
}

// Flush flush file to disk
func (f *File) Flush() error {
	err := f.File.Sync()
	if err != nil {
		return err
	}
	return err
}

// GetSize get size
func (f *File) GetSize() (int64, error) {
	var (
		size int64
		err  error
	)
	fileInfo, err := os.Stat(f.FilePath)
	if err != nil {
		return size, err
	}
	size = fileInfo.Size()
	return size, nil
}

// WriteInt64 write int64 to file
// func (f *File) WriteInt64(n int64) error {
// 	b, err := Int64ToByte(n)
// 	if err != nil {
// 		return err
// 	}
// 	_, err = f.File.Write(b)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }
func (f *File) WriteInt64(n int64) error {
	err := f.WriteInt(int(n >> 32))
	if err != nil {
		return err
	}
	err = f.WriteInt(int(n))
	if err != nil {
		return err
	}
	return nil
}

// WriteInt write int
func (f *File) WriteInt(n int) error {
	f.WriteByte(byte(n >> 24))
	f.WriteByte(byte(n >> 16))
	f.WriteByte(byte(n >> 8))
	f.WriteByte(byte(n))
	return nil
}

// WriteVarInt64 write var int64
func (f *File) WriteVarInt64(n int64) error {
	for (n & ^0x7F) != 0 {
		s1 := byte((n & 0x7f) | 0x80)
		f.WriteByte(s1)
		n >>= 7
	}
	s2 := byte(n)
	f.WriteByte(s2)
	return nil
}

// WriteVarInt write var int
func (f *File) WriteVarInt(n int) error {
	for (n & ^0x7F) != 0 {
		s1 := byte((n & 0x7f) | 0x80)
		f.WriteByte(s1)
		n >>= 7
	}
	s2 := byte(n)
	f.WriteByte(s2)
	return nil
}

// ReadVarInt64 read
func (f *File) ReadVarInt64() (int64, error) {
	var (
		i, shift int64
		b        byte
		err      error
	)
	b, err = f.ReadByte()

	if err != nil {
		return i, err
	}

	i = int64(b) & 0x7F

	shift = 7

	for int64(b)&0x80 != 0 {
		b, err = f.ReadByte()
		i = i | (int64(b)&0x7F)<<shift
		shift = shift + 7
	}

	return i, nil
}

// WriteChars write chars
func (f *File) WriteChars(s string) error {
	_, err := f.File.WriteString(s)
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
	l := int64(len(s))
	err = f.WriteVarInt64(l) // (1) 写入字符串长度
	if err != nil {
		return err
	}
	_, err = f.File.WriteString(s) // (2) 写入字符串
	if err != nil {
		return err
	}
	return nil
}

// ReadString read string
func (f *File) ReadString() (string, error) {
	var (
		n   int64
		err error
	)
	n, err = f.ReadVarInt64() // 先读取字符串的长度
	if err != nil {
		return "", err
	}
	if n < 1 {
		n = 1 // 如果字符串长度为0，表示字符串为"", 需要读取1个字节
	}
	b := make([]byte, n)
	_, err = f.File.Read(b)
	if err != nil {
		return "", err
	}
	return string(b), nil
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

// ReadByte read byte
func (f *File) ReadByte() (byte, error) {
	var (
		err error
	)
	b := make([]byte, 1)
	_, err = f.File.Read(b)
	if err != nil {
		return 0, err
	}
	return b[0], nil
}

// SeekFrom seek file
func (f *File) SeekFrom(n int64) error {
	_, err := f.File.Seek(n, 0)
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
