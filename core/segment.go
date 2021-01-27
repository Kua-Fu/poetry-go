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

// SegmentReader segment reader
type SegmentReader struct {
	seg          *SegmentInfo  // segmentInfo Ptr
	fieldInfos   *FieldInfos   // fieldInfos
	fieldsReader *FieldsReader // fields reader
	termsReader  *TermsReader  // terms reader
}

// SegmentMerger segment merger
type SegmentMerger struct {
	dirPath    string           // segment dir
	name       string           // segment name
	readers    []*SegmentReader // segment reader
	fieldInfos *FieldInfos
}

// SegmentMergeInfo segment merge info
type SegmentMergeInfo struct {
	term   *Term
	base   int64
	reader *SegmentReader
	// postings
}

// SegmentMergeQueue segment merge queue
type SegmentMergeQueue struct {
}

// ================================SegmentInfos=======================================

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
	filePath := fPtr.filePath + "segments.new"

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

	sPtr.flush()
	nfilepath := fPtr.filePath + "segments"
	sPtr.rename(nfilepath)
	return nil
}

// ================================SegmentReader=======================================

// max doc
func (sr *SegmentReader) maxDoc() int64 {
	return sr.fieldsReader.size
}

// Init segment reader init
func (sr *SegmentReader) init(si SegmentInfo) error {

	sr.seg = &si

	fieldsPtr := new(FieldInfos)
	fieldsPtr.empty()

	sr.fieldInfos = fieldsPtr

	// deserialize fnm info
	sr.initFieldNames()

	// fields reader
	fr := new(FieldsReader)
	fr.init(si.dirPath, si.name, sr.fieldInfos)
	sr.fieldsReader = fr

	// terms info
	tr := new(TermsReader)
	tr.init(si.dirPath, si.name, sr.fieldInfos)

	return nil
}

// InitFieldNames deserialize fnm info
func (sr *SegmentReader) initFieldNames() error {

	var (
		err       error
		filepath  string
		isIndexed bool
	)

	filepath = path.Join(sr.seg.dirPath, sr.seg.name+FileSuffix["fieldName"])

	f, err := CreateFile(filepath, false, true)
	if err != nil {
		return err
	}

	n, err := f.readVarInt() // (1) get field count

	for n > 0 {

		s, err := f.readString() // (2) get field values

		if err != nil {
			return err
		}

		b, err := f.readByte()
		if b == 1 {
			isIndexed = true
		}

		fi := FieldInfo{
			name:      s,
			isIndexed: isIndexed,
			number:    int64(len(sr.fieldInfos.byNumber)),
		}

		// init fieldInfos
		sr.fieldInfos.byNumber = append(sr.fieldInfos.byNumber, fi)
		sr.fieldInfos.byName[s] = fi

		n = n - 1
	}
	return nil
}

// ================================SegmentMerger=======================================

// Add add reader
func (sm *SegmentMerger) add(r *SegmentReader) error {
	sm.readers = append(sm.readers, r)
	return nil
}

// merge merge segment
func (sm *SegmentMerger) merge() error {

	sm.mergeFieldNames() // (1) merge field names

	sm.mergeFieldValues() // (2) merge field values

	sm.mergeFieldPostings() // (3) merge field postings

	return nil
}

// mergeFieldNames merge field names
func (sm *SegmentMerger) mergeFieldNames() error {

	fieldsPtr := new(FieldInfos)
	fieldsPtr.empty()

	for _, r := range sm.readers { // add field info
		fieldsPtr.addFields(r.fieldInfos)
	}

	sm.fieldInfos = fieldsPtr // init fieldinfos

	filePath := path.Join(
		sm.dirPath,
		sm.name+FileSuffix["fieldName"],
	)

	fieldsPtr.write(filePath)

	return nil
}

// MergeFieldValues merge field values
func (sm *SegmentMerger) mergeFieldValues() error {
	fw := FieldsWriter{}
	fw.init(sm.dirPath, sm.name, sm.fieldInfos)
	for _, r := range sm.readers {
		maxDoc := r.maxDoc()
		i := int64(0)
		for i < maxDoc {
			doc, _ := r.fieldsReader.doc(i)
			fw.addDocument(doc)
			i = i + 1
		}
	}
	fw.Close()
	return nil
}

// mergeFieldPostings merge field postings
func (sm *SegmentMerger) mergeFieldPostings() error {

	// queue := SegmentMergeQueue{}

	// for _, r := range sm.readers {

	// }

	return nil
}
