package core

import (
	"fmt"
	"path"
)

// TermsWriter term info writer
type TermsWriter struct {
	fieldInfos       *FieldInfos
	lastTerm         Term
	lastTi           TermInfo
	isIndex          bool
	size             int64
	other            *TermsWriter
	output           *File
	lastIndexPointer int64
}

// Init termsWriter init
func (tw *TermsWriter) init(dirPath, segment string, fieldInfos *FieldInfos) error {
	var (
		err      error
		fPtr     *File
		filePath string
	)
	tw.fieldInfos = fieldInfos

	filePath = path.Join(dirPath, segment+FileSuffix["termInfos"])
	fPtr, err = CreateFile(filePath, false, false)

	if err != nil {
		return err
	}

	tw.output = fPtr
	fPtr.writeInt(0)
	tw.output = fPtr

	// other
	filePath = path.Join(dirPath, segment+FileSuffix["termInfoIndex"])
	fPtr, err = CreateFile(filePath, false, false)

	if err != nil {
		return err
	}
	fPtr.writeInt(0) // (1) write int

	other := &TermsWriter{
		fieldInfos: fieldInfos,
		isIndex:    true,
		output:     fPtr,
		lastTerm:   Term{},
		lastTi:     TermInfo{},
		other:      tw,
	}

	tw.lastTerm = Term{}
	tw.lastTi = TermInfo{}
	tw.other = other
	return nil
}

// AddTerm add term
func (tw *TermsWriter) addTerm(term Term, ti TermInfo) error {
	if tw.isIndex == false && term.compare(tw.lastTerm) <= 0 {
		return fmt.Errorf("term out of order")
	}
	if ti.frqPtr < tw.lastTi.frqPtr {
		return fmt.Errorf("freqPointer out of order")
	}
	if ti.prxPtr < tw.lastTi.prxPtr {
		return fmt.Errorf("proxPointer out of order")
	}

	if tw.isIndex == false && tw.size%IndexInterval == 0 {
		tw.other.addTerm(tw.lastTerm, tw.lastTi)
	}

	tw.writeTerm(term)
	tw.output.writeVarInt(int(ti.docFrq))
	tw.output.writeVarInt64(ti.frqPtr - tw.lastTi.frqPtr)
	tw.output.writeVarInt64(ti.prxPtr - tw.lastTi.prxPtr)

	if tw.isIndex {
		size, err := tw.other.output.getSize()
		if err != nil {
			return err
		}
		n := size - tw.lastIndexPointer
		tw.output.writeVarInt64(n)
		nSize, err := tw.other.output.getSize()
		if err != nil {
			return err
		}
		tw.lastIndexPointer = nSize
	}

	tw.lastTi.docFrq = ti.docFrq
	tw.lastTi.frqPtr = ti.frqPtr
	tw.lastTi.prxPtr = ti.prxPtr

	tw.size = tw.size + 1

	return nil
}

// WriteTerm write term
func (tw *TermsWriter) writeTerm(term Term) error {

	start := StringDifference(tw.lastTerm.text, term.text)
	l := int64(len(term.text)) - start

	tw.output.writeVarInt(int(start))       // write shared prefix length
	tw.output.writeVarInt(int(l))           // write delta length
	tw.output.writeChars(term.text[start:]) // write delta chars

	n, err := tw.fieldInfos.getNumber(term.field)
	if err != nil {
		return err
	}
	tw.output.writeVarInt(int(n))
	return nil
}

// Close flush file to disk
func (tw *TermsWriter) close() error {

	tw.output.seekFrom(0) // write size at start
	tw.output.writeInt(int(tw.size))

	tw.output.flush()

	if !tw.isIndex {
		tw.other.close()
	}
	return nil
}
