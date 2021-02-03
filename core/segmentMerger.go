package core

import (
	"path"
	"strconv"
)

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

	fieldsPtr.write(filePath, false)

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

// ================================SegmentMergeInfo=======================================
func (s *SegmentMergeInfo) init(base int64, term *Term, termInfo *TermInfo, reader *SegmentReader) error {
	s.term = term
	s.termInfo = termInfo
	s.base = base
	s.reader = reader
	return nil
}
