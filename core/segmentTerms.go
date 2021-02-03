package core

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
