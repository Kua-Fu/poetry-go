package core

import "path"

// FieldsReader fields reader
type FieldsReader struct {
	fieldInfos  *FieldInfos
	fieldsData  *File
	fieldsIndex *File
	size        int64
}

// TermsReader terms reader
type TermsReader struct {
	fieldInfos *FieldInfos
	termsData  *File
	termsIndex *File
	segTerms   *SegmentTerms
}

// SegmentTerms segment term enum
type SegmentTerms struct {
	input      *File
	fieldInfos *FieldInfos
	size       int64
	isIndex    bool
	terms      []*Term
	termInfos  []*TermInfo
	ptrs       []int64
}

// ================================FieldsReader=======================================

// init init reader
func (fr *FieldsReader) init(dirPath string, segment string, fn *FieldInfos) error {
	fr.fieldInfos = fn

	filePath := path.Join(dirPath, segment+FileSuffix["fieldData"])

	fieldsData, err := CreateFile(filePath, false, true)
	if err != nil {
		return err
	}

	fr.fieldsData = fieldsData

	filePath = path.Join(dirPath, segment+FileSuffix["fieldIndex"])
	fieldsIndex, err := CreateFile(filePath, false, true)
	if err != nil {
		return err
	}

	size, err := fieldsIndex.getSize()
	if err != nil {
		return err
	}

	fr.size = size / 8 // get field size
	fr.fieldsIndex = fieldsIndex

	return nil

}

// get doc
func (fr *FieldsReader) doc(n int64) (Document, error) {

	var (
		i   int
		doc Document
		err error
	)

	fr.fieldsIndex.seekFrom(n * 8) // get index offset

	position, err := fr.fieldsIndex.readInt64()
	if err != nil {
		return doc, err
	}

	fr.fieldsData.seekFrom(position)

	numFields, err := fr.fieldsData.readVarInt()
	if err != nil {
		return doc, err
	}

	for i < numFields {

		fieldNumber, _ := fr.fieldsData.readVarInt()
		fi, _ := fr.fieldInfos.getFieldInfo(fieldNumber)
		b, _ := fr.fieldsData.readByte() // istoken info
		v, _ := fr.fieldsData.readString()

		field := Field{
			name:        fi.name,
			value:       v,
			isStored:    true,
			isIndexed:   fi.isIndexed,
			isTokenized: (b & 1) != 0,
		}
		doc.Add(field)

		i = i + 1
	}

	return doc, nil
}

// ================================TermsReader=======================================

func (tr *TermsReader) init(dirPath string, segment string, fn *FieldInfos) error {

	var (
		filePath string
		dataPtr  *File
		indexPtr *File
		err      error
	)

	// tis
	filePath = path.Join(dirPath, segment+FileSuffix["termInfos"])
	dataPtr, err = CreateFile(filePath, false, true)
	if err != nil {
		return err
	}
	tr.termsData = dataPtr

	// tii
	filePath = path.Join(dirPath, segment+FileSuffix["termInfoIndex"])
	indexPtr, err = CreateFile(filePath, false, true)
	if err != nil {
		return err
	}
	tr.termsIndex = indexPtr

	tr.fieldInfos = fn

	tr.readIndex()
	return nil
}

// terms get terms
func (tr *TermsReader) terms() (*SegmentTerms, error) {

	return tr.segTerms, nil
}

// readIndex terms read index
func (tr *TermsReader) readIndex() error {
	segTerms := new(SegmentTerms)

	// get all term, termInfo
	segTerms.init(tr.termsIndex, tr.termsData, tr.fieldInfos, true)

	tr.segTerms = segTerms

	// indexSize := segTerms.size
	// indexTerms := []Term{}
	// for i < indexSize{

	// 	indexTerms = append(indexTerms, )
	// }

	return nil
}

// ================================SegmentTerms=======================================

func (st *SegmentTerms) init(tIndexPtr, tDataPtr *File, fieldInfos *FieldInfos, isIndex bool) error {

	st.fieldInfos = fieldInfos
	st.isIndex = isIndex

	// n, err := tIndexPtr.readInt() // (1) read size
	n, err := tDataPtr.readInt() // (1) read size
	if err != nil {
		return err
	}
	st.size = int64(n)

	// read term
	i := 0
	for i < n {
		st.readTerm(tDataPtr)
		st.readTermInfo(tDataPtr)
		st.readIndexPtr(tDataPtr)
		i = i + 1
	}

	return nil
}

// readTerm read and add term
func (st *SegmentTerms) readTerm(fPtr *File) error {
	start, _ := fPtr.readVarInt()
	length, _ := fPtr.readVarInt()
	// totalLength := start + length ? why need total length

	b := make([]byte, length)
	fPtr.readChars(b, false, int64(start))

	i, _ := fPtr.readVarInt()
	name, _ := st.fieldInfos.getFieldName(i)

	term := Term{
		field: name,
		text:  string(b),
	}
	st.terms = append(st.terms, &term)
	return nil

}

// readTermInfo read and add termInfo
func (st *SegmentTerms) readTermInfo(fPtr *File) error {

	docFrq, _ := fPtr.readVarInt()
	frqPtr, _ := fPtr.readVarInt64()
	prxPtr, _ := fPtr.readVarInt64()

	ti := TermInfo{
		docFrq: int64(docFrq),
		frqPtr: frqPtr,
		prxPtr: prxPtr,
	}

	st.termInfos = append(st.termInfos, &ti)

	return nil
}

// readIndexPtr read and add index pointer
func (st *SegmentTerms) readIndexPtr(fPtr *File) error {
	var (
		i int64 // current pointer
		n int
	)
	n = len(st.ptrs)
	if n == 0 {
		i = 0
	} else {
		i = st.ptrs[n-1]
	}
	if st.isIndex {
		index, _ := fPtr.readVarInt64()
		st.ptrs = append(st.ptrs, index+i)
	}
	return nil
}
