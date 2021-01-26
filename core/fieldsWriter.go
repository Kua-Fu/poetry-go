package core

import (
	"fmt"
	"path"
)

// FieldsWriter fields writer
type FieldsWriter struct {
	FieldInfos  FieldInfos
	FieldsData  *File
	FieldsIndex *File
}

// TermsWriter term info writer
type TermsWriter struct {
	FieldInfos       FieldInfos
	LastTerm         Term
	LastTi           TermInfo
	IsIndex          bool
	Size             int64
	Other            *TermsWriter
	Output           *File
	LastIndexPointer int64
}

var (
	INDEX_INTERVAL int64 = 128
)

// Init init fieldsWriter
func (fw *FieldsWriter) Init(dirPath string, segment string, fn FieldInfos) error {
	fw.FieldInfos = fn

	filePath := path.Join(dirPath, segment+FileSuffix["fieldData"])

	fieldsData, err := CreateFile(filePath, false, false)
	if err != nil {
		return err
	}
	fw.FieldsData = fieldsData

	filePath = path.Join(dirPath, segment+FileSuffix["fieldIndex"])
	fieldsIndex, err := CreateFile(filePath, false, false)
	if err != nil {
		return err
	}
	fw.FieldsIndex = fieldsIndex
	return nil
}

// AddDocument add doc
func (fw *FieldsWriter) AddDocument(doc Document) error {
	var (
		storedCount int64
		err         error
		size        int64
	)
	size, err = fw.FieldsIndex.GetSize()
	if err != nil {
		return err
	}
	err = fw.FieldsIndex.WriteInt64(size)
	if err != nil {
		return err
	}
	storedCount = 0
	for _, field := range doc.Fields {
		if field.IsStored {
			storedCount = storedCount + 1
		}
	}
	err = fw.FieldsData.WriteVarInt64(storedCount)
	if err != nil {
		return err
	}
	for _, field := range doc.Fields {
		if field.IsStored {
			fieldName := field.Name
			fiNumber, _ := fw.FieldInfos.GetNumber(fieldName)
			err = fw.FieldsData.WriteVarInt64(fiNumber)
			if err != nil {
				return err
			}
			var bits byte
			bits = 0
			if field.IsTokenized {
				bits = bits | 1
			}
			fw.FieldsData.WriteByte(bits)
			fw.FieldsData.WriteString(field.Value)
		}
	}
	return nil
}

// Close flush file to disk
func (fw *FieldsWriter) Close() error {
	var err error
	err = fw.FieldsIndex.Flush()
	if err != nil {
		return err
	}
	err = fw.FieldsData.Flush()
	if err != nil {
		return err
	}
	return nil
}

// Init termsWriter init
func (tw *TermsWriter) Init(dirPath, segment string, fieldInfos FieldInfos) error {
	var (
		err      error
		fPtr     *File
		filePath string
	)
	tw.FieldInfos = fieldInfos

	filePath = dirPath + segment + FileSuffix["termInfos"]
	fPtr, err = CreateFile(filePath, false, false)

	if err != nil {
		return err
	}

	tw.Output = fPtr
	fPtr.WriteInt(0)
	tw.Output = fPtr

	// other
	filePath = dirPath + segment + FileSuffix["termInfoIndex"]
	fPtr, err = CreateFile(filePath, false, false)

	if err != nil {
		return err
	}
	fPtr.WriteInt(0)

	other := &TermsWriter{
		FieldInfos: fieldInfos,
		IsIndex:    true,
		Output:     fPtr,
		LastTerm:   Term{},
		LastTi:     TermInfo{},
		Other:      tw,
	}

	tw.LastTerm = Term{}
	tw.LastTi = TermInfo{}
	tw.Other = other
	return nil
}

// AddTerm add term
func (tw *TermsWriter) AddTerm(term Term, ti TermInfo) error {
	if tw.IsIndex == false && term.Compare(tw.LastTerm) <= 0 {
		return fmt.Errorf("term out of order")
	}
	if ti.FrqPtr < tw.LastTi.FrqPtr {
		return fmt.Errorf("freqPointer out of order")
	}
	if ti.PrxPtr < tw.LastTi.PrxPtr {
		return fmt.Errorf("proxPointer out of order")
	}

	if tw.IsIndex == false && tw.Size%INDEX_INTERVAL == 0 {
		tw.Other.AddTerm(tw.LastTerm, tw.LastTi)
	}

	tw.WriteTerm(term)
	tw.Output.WriteVarInt(int(ti.DocFrq))
	tw.Output.WriteVarInt64(ti.FrqPtr - tw.LastTi.FrqPtr)
	tw.Output.WriteVarInt64(ti.PrxPtr - tw.LastTi.PrxPtr)

	if tw.IsIndex {
		size, err := tw.Other.Output.GetSize()
		if err != nil {
			return err
		}
		n := size - tw.LastIndexPointer
		tw.Output.WriteVarInt64(n)
		nSize, err := tw.Other.Output.GetSize()
		if err != nil {
			return err
		}
		tw.LastIndexPointer = nSize
	}

	tw.LastTi.DocFrq = ti.DocFrq
	tw.LastTi.FrqPtr = ti.FrqPtr
	tw.LastTi.PrxPtr = ti.PrxPtr

	tw.Size = tw.Size + 1

	return nil
}

// WriteTerm write term
func (tw *TermsWriter) WriteTerm(term Term) error {

	start := StringDifference(tw.LastTerm.Text, term.Text)
	l := int64(len(term.Text)) - start

	tw.Output.WriteVarInt(int(start))       // write shared prefix length
	tw.Output.WriteVarInt(int(l))           // write delta length
	tw.Output.WriteChars(term.Text[start:]) // write delta chars

	n, err := tw.FieldInfos.GetNumber(term.Field)
	if err != nil {
		return err
	}
	tw.Output.WriteVarInt(int(n))
	return nil
}

// Close flush file to disk
func (tw *TermsWriter) Close() error {

	tw.Output.SeekFrom(0) // write size at start
	tw.Output.WriteInt(int(tw.Size))

	tw.Output.Flush()

	if !tw.IsIndex {
		tw.Other.Close()
	}
	return nil
}
