package core

import (
	"os"
	"path"
	"strconv"
)

// DocumentWriter document writer
type DocumentWriter struct {
	Analyzer Analyzer
	// Directory      Directory
	FieldInfos     FieldInfos
	MaxFieldLength int64
	DirPath        string
	PostingTable   map[Term]Posting
	FieldLengths   []int64
}

// Init init document writer
func (dw *DocumentWriter) Init(dirPath string, analyzer Analyzer, mfl int64) error {
	dw.DirPath = dirPath
	dw.Analyzer = analyzer
	dw.MaxFieldLength = mfl
	return nil
}

// AddDocument add doc
func (dw *DocumentWriter) AddDocument(segment string, doc Document) (bool, error) {

	// (1) add field names
	dw.addFieldNames(segment, doc)

	// (2) add field values
	dw.addFieldValues(segment, doc)

	// (3) add field positions, (frequency and position)
	dw.addFieldPostings(segment, doc)

	// (4) add norms
	dw.addFieldNorms(segment, doc)

	return true, nil
}

// add field names
func (dw *DocumentWriter) addFieldNames(segment string, doc Document) error {
	fieldInfos := FieldInfos{
		ByNumber: []FieldInfo{},
		ByName:   map[string]FieldInfo{},
	}
	fieldInfos.Init()
	fieldInfos.AddDoc(doc)
	dw.FieldInfos = fieldInfos
	filePath := path.Join(dw.DirPath, segment+FileSuffix["fieldName"])
	fieldInfos.Write(filePath)
	return nil
}

// add field values
func (dw *DocumentWriter) addFieldValues(segment string, doc Document) error {
	var err error
	fw := FieldsWriter{}
	fw.Init(dw.DirPath, segment, dw.FieldInfos)
	err = fw.AddDocument(doc)
	if err != nil {
		return err
	}
	err = fw.Close() // flush
	if err != nil {
		return err
	}
	return nil
}

// add field postings
func (dw *DocumentWriter) addFieldPostings(segment string, doc Document) error {
	// invert doc into postingTable
	dw.PostingTable = map[Term]Posting{}
	dw.invertDocument(doc)

	// sort postingTable into an array
	postings, _ := dw.sortPostingTable()

	// write postings
	dw.writePostings(postings, segment)
	return nil
}

func (dw *DocumentWriter) invertDocument(doc Document) error {

	lenFields := len(dw.FieldInfos.ByNumber)
	dw.FieldLengths = make([]int64, lenFields)

	for _, field := range doc.Fields {
		fieldName := field.Name
		fieldNumber, _ := dw.FieldInfos.GetNumber(fieldName)
		position := dw.FieldLengths[fieldNumber] // position in field
		fieldValue := field.Value
		if field.IsIndexed {
			if !field.IsTokenized { // un-tokenized field
				dw.addPosition(fieldName, fieldValue, position)
			} else {
				// fieldValue := field.Value
			}
		}
		position = position + 1
		dw.FieldLengths[fieldNumber] = position
	}
	return nil
}

func (dw *DocumentWriter) addPosition(fieldName string, fieldValue string, position int64) error {
	// word not seen before
	term := Term{
		Field: fieldName,
		Text:  fieldValue,
	}
	posting := Posting{
		Term:      term,
		Freq:      1,
		Positions: []int64{position},
	}
	dw.PostingTable[term] = posting
	return nil
}

func (dw *DocumentWriter) sortPostingTable() ([]Posting, error) {
	var postings = []Posting{}
	for _, v := range dw.PostingTable {
		postings = append(postings, v)
	}
	return postings, nil
}

func (dw *DocumentWriter) writePostings(postings []Posting, segment string) error {
	var (
		filePath string
		err      error
	)

	filePath = path.Join(dw.DirPath, segment+FileSuffix["termFrequencies"])
	frqPtr, err := CreateFile(filePath, false, false)
	if err != nil {
		return err
	}

	filePath = path.Join(dw.DirPath, segment+FileSuffix["termPositions"])
	prxPtr, err := CreateFile(filePath, false, false)
	if err != nil {
		return err
	}

	tw := new(TermsWriter)
	tw.Init(dw.DirPath, segment, dw.FieldInfos)
	ti := TermInfo{}

	for _, posting := range postings {
		// init terminfo
		frqSize, err := frqPtr.GetSize()
		if err != nil {
			return err
		}
		prxSize, err := prxPtr.GetSize()
		if err != nil {
			return err
		}

		// add an entry to the dictionary with pointers to prox and freq files
		ti.Init(1, frqSize, prxSize)
		err = tw.AddTerm(posting.Term, ti)
		if err != nil {
			return err
		}

		// add an entry to the freq file
		f := posting.Freq
		if f == 1 { // optimize freq=1
			frqPtr.WriteVarInt(1) // set low bit of doc num.
		} else {
			frqPtr.WriteVarInt(0)      // the document number
			frqPtr.WriteVarInt(int(f)) // frequency in doc
		}

		var lastPosition int64 = 0 // write positions
		positions := posting.Positions
		i := int64(0)
		for i < posting.Freq {
			position := positions[i]
			diff := position - lastPosition
			prxPtr.WriteVarInt64(diff)
			lastPosition = position
			i = i + 1
		}
	}

	if frqPtr != nil {
		frqPtr.Flush()
	}
	if prxPtr != nil {
		prxPtr.Flush()
	}

	if tw != nil {
		tw.Close()
	}
	return nil
}

// write frq
// func (dw *DocumentWriter) writeFrq(postings []Posting, segment string) error {
// 	filePath := dw.DirPath + segment + FileSuffix["termFrequencies"]
// 	fPtr, err := CreateFile(filePath, false, false)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// write prx
func (dw *DocumentWriter) writePrx(postings []Posting, segment string) error {

	filePath := path.Join(dw.DirPath, segment+FileSuffix["termPositions"])
	fPtr, err := os.Create(filePath)
	if err != nil {
		return err
	}

	for _, posting := range postings {
		lastPosition := int64(0)
		positions := posting.Positions
		i := int64(0)
		for i < posting.Freq {
			position := positions[i]
			diff := position - lastPosition
			b, _ := Int64ToByte(diff)
			fPtr.Write(b)
			lastPosition = position
			i = i + 1
		}
	}
	return nil
}

// write tis
func (dw *DocumentWriter) writeTis(postings []Posting, segment string) error {
	return nil
}

// add field norms
func (dw *DocumentWriter) addFieldNorms(segment string, doc Document) error {
	for _, field := range doc.Fields {
		if field.IsIndexed {
			fieldNumber, err := dw.FieldInfos.GetNumber(field.Name)
			if err != nil {
				return err
			}
			filePath := path.Join(dw.DirPath, segment+FileSuffix["norms"]+strconv.FormatInt(fieldNumber, 10))
			nPtr, err := CreateFile(filePath, false, false)
			if err != nil {
				return err
			}

			n := SimilarityNorm(dw.FieldLengths[fieldNumber])
			nPtr.WriteByte(n)
			nPtr.Flush()
		}
	}
	return nil
}
