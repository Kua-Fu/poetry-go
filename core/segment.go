package core

import (
	"path"
)

// SegmentInfos segment infos
type SegmentInfos struct {
	Counter  int64
	SegInfos []SegmentInfo
}

// SegmentInfo  segment info
type SegmentInfo struct {
	Name     string
	DocCount int64
	Dirpath  string
}

// SegmentMerger segment merger
type SegmentMerger struct {
	DirPath string           // segment dir
	Name    string           // segment name
	Readers []*SegmentReader // segment reader
}

// SegmentReader segment reader
type SegmentReader struct {
	SegPtr     *SegmentInfo // segmentInfo Ptr
	FieldInfos FieldInfos   // fieldInfos
}

// Info segments info
func (s *SegmentInfos) Info(i int64) (SegmentInfo, error) {
	return s.SegInfos[i], nil
}

// AddElement add element
func (s *SegmentInfos) AddElement(segInfo SegmentInfo) error {
	s.SegInfos = append(s.SegInfos, segInfo)
	return nil
}

// Write create file
func (s *SegmentInfos) Write(fPtr *File) error {
	filePath := fPtr.FilePath + "segments.new"

	sPtr, err := CreateFile(filePath, false, false)
	if err != nil {
		return err
	}

	sPtr.WriteInt(int(s.Counter))
	sPtr.WriteInt(len(s.SegInfos))

	for _, seg := range s.SegInfos {
		sPtr.WriteString(seg.Name)
		sPtr.WriteInt(int(seg.DocCount))
	}

	sPtr.Flush()
	nfilepath := fPtr.FilePath + "segments"
	sPtr.Rename(nfilepath)
	return nil
}

// Init segment reader init
func (sr *SegmentReader) Init(si SegmentInfo) error {

	sr.SegPtr = &si

	sr.FieldInfos = FieldInfos{ // init fieldInfo
		ByNumber: []FieldInfo{},
		ByName:   map[string]FieldInfo{},
	}

	// deserialize fnm info
	sr.InitFieldNames()

	return nil
}

// InitFieldNames deserialize fnm info
func (sr *SegmentReader) InitFieldNames() error {

	var (
		err      error
		filepath string
	)

	filepath = path.Join(sr.SegPtr.Dirpath, sr.SegPtr.Name+FileSuffix["fieldName"])

	f, err := CreateFile(filepath, false, true)
	if err != nil {
		return err
	}

	n, err := f.ReadVarInt64() // (1) get field count

	for n > 0 {

		s, err := f.ReadString() // (2) get field values

		if err != nil {
			return err
		}

		fi := FieldInfo{
			Name:      s,
			IsIndexed: true,
			Number:    int64(len(sr.FieldInfos.ByNumber)),
		}

		// init fieldInfos
		sr.FieldInfos.ByNumber = append(sr.FieldInfos.ByNumber, fi)
		sr.FieldInfos.ByName[s] = fi

		n = n - 1
	}
	return nil
}

// Add add reader
func (sm *SegmentMerger) Add(r *SegmentReader) error {
	sm.Readers = append(sm.Readers, r)
	return nil
}

// Merge merge segment
func (sm *SegmentMerger) Merge() error {

	sm.MergeFieldNames()

	sm.MergeFieldValues()

	// sm.MergeNorms()

	return nil
}

// MergeFieldNames merge field names
func (sm *SegmentMerger) MergeFieldNames() error {

	fieldInfos := FieldInfos{
		ByName:   map[string]FieldInfo{},
		ByNumber: []FieldInfo{},
	}
	for _, r := range sm.Readers {
		fieldInfos.AddFields(r.FieldInfos)
	}
	filePath := path.Join(sm.DirPath, sm.Name+FileSuffix["fieldName"])
	fieldInfos.Write(filePath)

	return nil
}

// MergeFieldValues merge field values
func (sm *SegmentMerger) MergeFieldValues() error {

	return nil
}
