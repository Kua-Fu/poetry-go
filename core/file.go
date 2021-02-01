package core

import (
	"fmt"
	"io/ioutil"
	"os"
)

// File file overwrite os.File
type File struct {
	filePath string
	file     *os.File
	isDir    bool
}

// FSDirectory fs directory
type FSDirectory struct {
	dirPath string
	dir     *File
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
			filePath: fPath,
			file:     fPtr,
			isDir:    true,
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
		filePath: filePath,
		file:     fPtr,
		isDir:    isDir,
	}
	return &f, nil
}

// remove dir
func (f *File) removeAll() error {
	err := os.RemoveAll(f.filePath)
	if err != nil {
		fmt.Println("--delete err--", err)
		return err
	}
	fmt.Println("--delete dir success--", f.filePath)
	return nil
}

// Rename rename file
func (f *File) rename(filePath string) error {

	err := os.Rename(f.filePath, filePath)
	if err != nil {
		return err
	}
	f.filePath = filePath
	return nil
}

// Flush flush file to disk
func (f *File) flush() error {
	err := f.file.Sync()
	if err != nil {
		return err
	}
	return err
}

// close close file
func (f *File) close() error {
	err := f.file.Close()
	if err != nil {
		return err
	}
	return nil
}

// GetSize get size
func (f *File) getSize() (int64, error) {
	var (
		size int64
		err  error
	)
	fileInfo, err := os.Stat(f.filePath)
	if err != nil {
		return size, err
	}
	size = fileInfo.Size()
	return size, nil
}

// WriteInt64 write int64 to file
func (f *File) writeInt64(n int64) error {
	err := f.writeInt(int(n >> 32))
	if err != nil {
		return err
	}
	err = f.writeInt(int(n))
	if err != nil {
		return err
	}
	return nil
}

// readInt64 read int64
func (f *File) readInt64() (int64, error) {
	var (
		ih, il int
		ii     int64
		err    error
	)
	ih, err = f.readInt()
	if err != nil {
		return ii, err
	}

	ih = int(int64(ih) << 32)

	il, err = f.readInt()
	if err != nil {
		return ii, err
	}

	il = il & 0xFFFFFFFF

	ii = int64(ih | il)
	return ii, nil
}

// WriteInt write int
func (f *File) writeInt(n int) error {
	f.writeByte(byte(n >> 24))
	f.writeByte(byte(n >> 16))
	f.writeByte(byte(n >> 8))
	f.writeByte(byte(n))
	return nil
}

// readint read int
func (f *File) readInt() (int, error) {
	var (
		i   int
		b   byte
		err error
	)
	b, err = f.readByte()
	if err != nil {
		return i, err
	}
	b1 := b & 0xFF << 24

	b, err = f.readByte()
	if err != nil {
		return i, err
	}
	b2 := b & 0xFF << 16

	b, err = f.readByte()
	if err != nil {
		return i, err
	}
	b3 := b & 0xFF << 8

	b, err = f.readByte()
	if err != nil {
		return i, err
	}
	b4 := b & 0xFF

	i = int(b1 | b2 | b3 | b4)
	return i, nil
}

// WriteVarInt64 write var int64
func (f *File) writeVarInt64(n int64) error {
	for (n & ^0x7F) != 0 {
		s1 := byte((n & 0x7f) | 0x80)
		f.writeByte(s1)
		n >>= 7
	}
	s2 := byte(n)
	f.writeByte(s2)
	return nil
}

// WriteVarInt write var int
func (f *File) writeVarInt(n int) error {
	for (n & ^0x7F) != 0 {
		s1 := byte((n & 0x7f) | 0x80)
		f.writeByte(s1)
		n >>= 7
	}
	s2 := byte(n)
	f.writeByte(s2)
	return nil
}

// ReadVarInt64 read
func (f *File) readVarInt64() (int64, error) {
	var (
		i, shift int64
		b        byte
		err      error
	)
	b, err = f.readByte()

	if err != nil {
		return i, err
	}

	i = int64(b) & 0x7F

	shift = 7

	for int64(b)&0x80 != 0 {
		b, err = f.readByte()
		i = i | (int64(b)&0x7F)<<shift
		shift = shift + 7
	}

	return i, nil
}

// readVarInt read
func (f *File) readVarInt() (int, error) {
	var (
		b     byte
		i     int
		shift int
	)

	b, _ = f.readByte()

	shift = 7

	i = int(b & 0x7F)
	for (b & 0x80) != 0 {
		b, _ = f.readByte()
		i = i | int((b&0x7F)<<shift)
		shift += 7
	}
	return i, nil
}

// WriteChars write chars
func (f *File) writeChars(s string) error {
	_, err := f.file.WriteString(s)
	if err != nil {
		return err
	}
	return nil
}

// readChars read chars
func (f *File) readChars(b []byte, withOff bool, off int64) error {
	var err error
	if withOff {
		_, err = f.file.ReadAt(b, off)
	} else {
		_, err = f.file.Read(b)
	}

	if err != nil {
		return err
	}
	return nil
}

// WriteString write string
func (f *File) writeString(s string) error {
	var (
		err error
	)
	l := len(s)
	err = f.writeVarInt(l) // (1) 写入字符串长度
	if err != nil {
		return err
	}
	_, err = f.file.WriteString(s) // (2) 写入字符串
	if err != nil {
		return err
	}
	return nil
}

// ReadString read string
func (f *File) readString() (string, error) {
	var (
		n   int
		err error
	)
	n, err = f.readVarInt() // read string length
	if err != nil {
		return "", err
	}

	b := make([]byte, n)
	_, err = f.file.Read(b)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// WriteByte write string
func (f *File) writeByte(b byte) error {
	var (
		err error
	)
	_, err = f.file.Write([]byte{b})
	if err != nil {
		return err
	}
	return nil
}

// ReadByte read byte
func (f *File) readByte() (byte, error) {
	var (
		err error
	)
	b := make([]byte, 1)
	_, err = f.file.Read(b)
	if err != nil {
		return 0, err
	}
	return b[0], nil
}

// SeekFrom seek file
func (f *File) seekFrom(n int64) error {
	_, err := f.file.Seek(n, 0)
	if err != nil {
		return err
	}
	return nil
}

// Init check file exist and is a directory
func (fs *FSDirectory) init(dirPath string) error {

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
	fs.dirPath = dirPath
	fs.dir = f
	return nil
}
