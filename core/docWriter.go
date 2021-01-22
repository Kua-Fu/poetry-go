package core

import "os"

// DocumentWriter document writer
type DocumentWriter struct {
	Analyzer Analyzer
	// Directory      Directory
	FieldInfos     FieldInfos
	MaxFieldLength int64
	DirPath        string
	PostingTable   map[Term]Posting
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

	// write field values
	fieldsWriter := FieldsWriter{}
	fieldsWriter.Init(dw.DirPath, segment, dw.FieldInfos)
	fieldsWriter.AddDocument(doc)

	// invert doc into postingTable
	dw.PostingTable = map[Term]Posting{}
	dw.invertDocument(doc)

	// sort postingTable into an array
	postings, _ := dw.sortPostingTable()

	// write postings
	dw.writePostings(postings, segment)

	// write norms of indexed fields
	dw.writeNorms(segment, doc)

	return true, nil
}

// addFieldName add field names
func (dw *DocumentWriter) addFieldNames(segment string, doc Document) error {
	fieldInfos := FieldInfos{
		ByNumber: []FieldInfo{},
		ByName:   map[string]FieldInfo{},
	}
	fieldInfos.Add(doc)
	dw.FieldInfos = fieldInfos
	filePath := dw.DirPath + segment + FileSuffix["fieldName"]
	fieldInfos.Write(filePath)
	return nil
}

func (dw *DocumentWriter) invertDocument(doc Document) error {
	lenFields := len(dw.FieldInfos.ByNumber)
	fieldLengths := make([]int64, lenFields)

	for _, field := range doc.Fields {
		fieldName := field.Name
		fieldNumber, _ := dw.FieldInfos.GetNumber(fieldName)
		position := fieldLengths[fieldNumber] // position in field
		fieldValue := field.Value
		if field.IsIndexed {
			if !field.IsTokenized { // un-tokenized field
				dw.addPosition(fieldName, fieldValue, position)
			} else {
				// fieldValue := field.Value
			}
		}
		position = position + 1
		fieldLengths[fieldNumber] = position
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
	return nil, nil
}

func (dw *DocumentWriter) writePostings(postings []Posting, segment string) error {

	// write frq
	dw.writeFrq(postings, segment)

	// write prx
	dw.writePrx(postings, segment)

	// write tis
	dw.writeTis(postings, segment)

	return nil
}

// write frq
func (dw *DocumentWriter) writeFrq(postings []Posting, segment string) error {
	filePath := dw.DirPath + segment + FileSuffix["termFrequencies"]
	fPtr, err := os.Create(filePath)
	if err != nil {
		return err
	}

	for _, posting := range postings {
		f := posting.Freq
		if f == 1 {
			fPtr.Write([]byte{1})
		} else {
			fPtr.Write([]byte{0})
			b, _ := Int64ToByte(f)
			fPtr.Write(b)
		}
	}
	return nil
}

// write prx
func (dw *DocumentWriter) writePrx(postings []Posting, segment string) error {

	filePath := dw.DirPath + segment + FileSuffix["termPositions"]
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

// write Norms
func (dw *DocumentWriter) writeNorms(segment string, doc Document) error {
	return nil
}
