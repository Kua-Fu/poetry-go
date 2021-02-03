package core

import (
	"fmt"
	"path"
	"strconv"
)

// TermsReader terms reader
type TermsReader struct {
	fieldInfos *FieldInfos
	termsData  *File
	termsIndex *File
	segTerms   *SegmentTerms
}

// SegmentReader segment reader
type SegmentReader struct {
	seg          *SegmentInfo      // segmentInfo Ptr
	fieldInfos   *FieldInfos       // fieldInfos
	fieldsReader *FieldsReader     // fields reader
	termsReader  *TermsReader      // terms reader
	norms        *map[string]*Norm // norms
}

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
		} else {
			isIndexed = false
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
