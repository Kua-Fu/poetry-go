package core

import (
	"fmt"
	"path"
	"strconv"
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

// SegmentReader segment reader
type SegmentReader struct {
	seg          *SegmentInfo      // segmentInfo Ptr
	fieldInfos   *FieldInfos       // fieldInfos
	fieldsReader *FieldsReader     // fields reader
	termsReader  *TermsReader      // terms reader
	norms        *map[string]*Norm // norms
}

// SegmentMerger segment merger
type SegmentMerger struct {
	dirPath    string           // segment dir
	name       string           // segment name
	readers    []*SegmentReader // segment reader
	fieldInfos *FieldInfos
	tw         *TermsWriter
}

// SegmentMergeInfo segment merge info
type SegmentMergeInfo struct {
	index    int64
	term     *Term
	termInfo *TermInfo
	base     int64
	reader   *SegmentReader
	// postings
}

// SegmentMergeQueue segment merge queue
// type SegmentMergeQueue struct {
// }

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

// ================================SegmentMergeInfo=======================================
func (s *SegmentMergeInfo) init(base int64, term *Term, termInfo *TermInfo, reader *SegmentReader) error {
	s.term = term
	s.termInfo = termInfo
	s.base = base
	s.reader = reader
	return nil
}

// ================================SegmentReader=======================================

// max doc
func (sr *SegmentReader) maxDoc() int64 {
	return sr.fieldsReader.size
}

// numDocs
func (sr *SegmentReader) numDocs() int64 {
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

	sr.termsReader = tr
	sr.openNorms()

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

// openNorms open norms
func (sr *SegmentReader) openNorms() error {
	sr.norms = &(map[string]*Norm{})
	for _, fi := range sr.fieldInfos.byNumber {
		if fi.isIndexed {
			filepath := path.Join(sr.seg.dirPath, sr.seg.name+FileSuffix["norms"]+strconv.FormatInt(fi.number, 10))

			fPtr, err := CreateFile(filepath, false, true)
			if err != nil {
				return err
			}

			norm := Norm{
				fPtr:  fPtr,
				bytes: []byte{},
			}
			(*sr.norms)[fi.name] = &norm
		}
	}
	return nil
}

// normStream norm stream
func (sr *SegmentReader) normStream(field string) (*File, error) {
	if norm, ok := (*sr.norms)[field]; ok {
		fPtr := norm.fPtr // why in lucene is clone inputStream?
		fPtr.seekFrom(0)
		return fPtr, nil
	}
	return nil, fmt.Errorf("no such norm")

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

	sm.mergeFieldNorms() // (4) merge field norms

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
	var (
		filePath string
	)

	filePath = path.Join(sm.dirPath, sm.name+FileSuffix["termFrequencies"])
	frqPtr, err := CreateFile(filePath, false, false)
	if err != nil {
		return err
	}

	filePath = path.Join(sm.dirPath, sm.name+FileSuffix["termPositions"])
	prxPtr, err := CreateFile(filePath, false, false)
	if err != nil {
		return err
	}

	tw := new(TermsWriter)
	tw.init(sm.dirPath, sm.name, sm.fieldInfos)
	sm.tw = tw

	sm.mergeTermInfos(frqPtr, prxPtr)

	// file close
	tw.close()
	frqPtr.close()
	prxPtr.close()

	return nil
}

// mergeTermInfos merge term infos
func (sm *SegmentMerger) mergeTermInfos(frqPtr *File, prxPtr *File) error {

	// queue := make(PriorityQueue, len(sm.readers))
	queue := make(PriorityQueue, 0)
	base := int64(0)
	for _, r := range sm.readers {
		termsPtr, _ := r.termsReader.terms()
		// add every term
		for i, term := range termsPtr.terms {
			smi := new(SegmentMergeInfo)
			smi.init(base, term, termsPtr.termInfos[i], r)
			queue.Push(smi)
		}

		base = base + r.numDocs()

	}

	// reduce
	match := []*SegmentMergeInfo{}

	for queue.Len() > 0 {
		matchSize := 0
		smiPtr, _ := queue.Pop().(*SegmentMergeInfo)
		match = append(match, smiPtr)
		matchSize = matchSize + 1

		termPtr := match[0].term
		top, _ := queue.Top().(*SegmentMergeInfo)

		for top != nil && termPtr.compare(*top.term) == 0 {
			smiPtr, _ := queue.Pop().(*SegmentMergeInfo)
			match = append(match, smiPtr)
			matchSize = matchSize + 1
			top, _ = queue.Top().(*SegmentMergeInfo)
		}

		sm.mergeTermInfo(match, matchSize)

		// for _, smiPtr := range match {
		// 	queue.Push(smiPtr) // restore queue
		// }

	}
	return nil
}

// mergeTermInfo merge every term
func (sm *SegmentMerger) mergeTermInfo(match []*SegmentMergeInfo, matchSize int) error {
	for _, smi := range match {
		term := smi.term
		termInfo := smi.termInfo
		sm.tw.addTerm(*term, *termInfo)
	}
	return nil
}

// mergeFieldNorms merge field norms
func (sm *SegmentMerger) mergeFieldNorms() error {

	for i, fi := range sm.fieldInfos.byNumber {
		if fi.isIndexed {
			filePath := path.Join(sm.dirPath, sm.name+FileSuffix["norms"]+strconv.FormatInt(int64(i), 10))
			nfPtr, err := CreateFile(filePath, false, false)
			if err != nil {
				return err
			}
			// reader
			for _, reader := range sm.readers {
				fPtr, _ := reader.normStream(fi.name)
				maxDoc := reader.maxDoc()
				k := 0
				// write norm
				for k < int(maxDoc) {
					b, _ := fPtr.readByte()
					nfPtr.writeByte(b)
					k = k + 1
				}
				fPtr.close()
			}
			// close nfPtr
			nfPtr.close()
		}
	}
	return nil
}
