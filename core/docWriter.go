package core

import (
	"os"
	"path"
	"strconv"
)

// DocumentWriter document writer
type DocumentWriter struct {
	analyzer Analyzer
	// Directory      Directory
	fieldInfos     *FieldInfos
	maxFieldLength int64
	dirPath        string
	postingTable   map[Term]Posting
	fieldLengths   []int64
}

// Init init document writer
func (dw *DocumentWriter) Init(dirPath string, analyzer Analyzer, mfl int64) error {
	dw.dirPath = dirPath
	dw.analyzer = analyzer
	dw.maxFieldLength = mfl
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

	fieldsPtr := new(FieldInfos)
	fieldsPtr.empty()
	fieldsPtr.init()

	fieldsPtr.addDoc(doc)

	dw.fieldInfos = fieldsPtr
	filePath := path.Join(dw.dirPath, segment+FileSuffix["fieldName"])
	fieldsPtr.write(filePath, false)
	return nil
}

// add field values
func (dw *DocumentWriter) addFieldValues(segment string, doc Document) error {
	var err error
	fw := FieldsWriter{}
	fw.init(dw.dirPath, segment, dw.fieldInfos)
	err = fw.addDocument(doc)
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
	dw.postingTable = map[Term]Posting{}
	dw.invertDocument(doc)

	// sort postingTable into an array
	postings, _ := dw.sortPostingTable()

	// write postings
	dw.writePostings(postings, segment)
	return nil
}

func (dw *DocumentWriter) invertDocument(doc Document) error {

	lenFields := len(dw.fieldInfos.byNumber)
	dw.fieldLengths = make([]int64, lenFields)

	for _, field := range doc.Fields {
		fieldName := field.name
		fieldNumber, _ := dw.fieldInfos.getNumber(fieldName)
		position := dw.fieldLengths[fieldNumber] // position in field
		fieldValue := field.value
		if field.isIndexed {
			if !field.isTokenized { // un-tokenized field
				dw.addPosition(fieldName, fieldValue, position)
			} else {
				// fieldValue := field.Value
			}
		}
		position = position + 1
		dw.fieldLengths[fieldNumber] = position
	}
	return nil
}

func (dw *DocumentWriter) addPosition(fieldName string, fieldValue string, position int64) error {
	// word not seen before
	term := Term{
		field: fieldName,
		text:  fieldValue,
	}
	posting := Posting{
		term:      term,
		freq:      1,
		positions: []int64{position},
	}
	dw.postingTable[term] = posting
	return nil
}

func (dw *DocumentWriter) sortPostingTable() ([]Posting, error) {
	var postings = []Posting{}
	for _, v := range dw.postingTable {
		postings = append(postings, v)
	}
	return postings, nil
}

func (dw *DocumentWriter) writePostings(postings []Posting, segment string) error {
	var (
		filePath string
		err      error
	)

	filePath = path.Join(dw.dirPath, segment+FileSuffix["termFrequencies"])
	frqPtr, err := CreateFile(filePath, false, false)
	if err != nil {
		return err
	}

	filePath = path.Join(dw.dirPath, segment+FileSuffix["termPositions"])
	prxPtr, err := CreateFile(filePath, false, false)
	if err != nil {
		return err
	}

	tw := new(TermsWriter)
	tw.init(dw.dirPath, segment, dw.fieldInfos)
	ti := TermInfo{}

	for _, posting := range postings {
		// init terminfo
		frqSize, err := frqPtr.getSize()
		if err != nil {
			return err
		}
		prxSize, err := prxPtr.getSize()
		if err != nil {
			return err
		}

		// add an entry to the dictionary with pointers to prox and freq files
		ti.Init(1, frqSize, prxSize)
		err = tw.addTerm(posting.term, ti)
		if err != nil {
			return err
		}

		// add an entry to the freq file
		f := posting.freq
		if f == 1 { // optimize freq=1
			frqPtr.writeVarInt(1) // set low bit of doc num.
		} else {
			frqPtr.writeVarInt(0)      // the document number
			frqPtr.writeVarInt(int(f)) // frequency in doc
		}

		var lastPosition int64 = 0 // write positions
		positions := posting.positions
		i := int64(0)
		for i < posting.freq {
			position := positions[i]
			diff := position - lastPosition
			prxPtr.writeVarInt64(diff)
			lastPosition = position
			i = i + 1
		}
	}

	if frqPtr != nil {
		frqPtr.flush()
	}
	if prxPtr != nil {
		prxPtr.flush()
	}

	if tw != nil {
		tw.close()
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

	filePath := path.Join(dw.dirPath, segment+FileSuffix["termPositions"])
	fPtr, err := os.Create(filePath)
	if err != nil {
		return err
	}

	for _, posting := range postings {
		lastPosition := int64(0)
		positions := posting.positions
		i := int64(0)
		for i < posting.freq {
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
		if field.isIndexed {
			fieldNumber, err := dw.fieldInfos.getNumber(field.name)
			if err != nil {
				return err
			}
			filePath := path.Join(dw.dirPath, segment+FileSuffix["norms"]+strconv.FormatInt(fieldNumber, 10))
			nPtr, err := CreateFile(filePath, false, false)
			if err != nil {
				return err
			}

			n := SimilarityNorm(dw.fieldLengths[fieldNumber])
			nPtr.writeByte(n)
			nPtr.flush()
		}
	}
	return nil
}
