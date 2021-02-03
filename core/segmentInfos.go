package core

import (
	"path"
)

// SegmentInfo  segment info
type SegmentInfo struct {
	name     string
	docCount int64
	dirPath  string
}

// SegmentInfos segment infos
type SegmentInfos struct {
	counter  int64
	segInfos []SegmentInfo
}

// Norm norm
type Norm struct {
	fPtr  *File
	bytes []byte
}

// Info segments info
func (s *SegmentInfos) empty() error {
	s.counter = 0
	s.segInfos = []SegmentInfo{}
	return nil
}

// Info segments info
func (s *SegmentInfos) Info(i int64) (SegmentInfo, error) {
	return s.segInfos[i], nil
}

// AddElement add element
func (s *SegmentInfos) add(segInfo SegmentInfo) error {
	s.segInfos = append(s.segInfos, segInfo)
	return nil
}

// Write create file
func (s *SegmentInfos) write(fPtr *File) error {
	filePath := path.Join(fPtr.filePath, "segments.new")

	sPtr, err := CreateFile(filePath, false, false)
	if err != nil {
		return err
	}

	sPtr.writeInt(int(s.counter))
	sPtr.writeInt(len(s.segInfos))

	for _, seg := range s.segInfos {
		sPtr.writeString(seg.name)
		sPtr.writeInt(int(seg.docCount))
	}

	sPtr.close()
	nfilepath := path.Join(fPtr.filePath, "segments")
	sPtr.rename(nfilepath)
	return nil
}
