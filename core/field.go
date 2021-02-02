package core

import (
	"fmt"
	"strings"
)

/*
A field is a section of a Document.
Each field has two parts, a name and a value.
Values may be free text, provided as a String or as a Reader,
or they may be atomic keywords, which are not further processed.
Such keywords may be used to represent dates, urls, etc.
Fields are optionally stored in the index,
so that they may be returned with hits on the document.
*/

// Field field
type Field struct {
	name        string
	value       string
	isStored    bool
	isIndexed   bool
	isTokenized bool
}

// FieldInfo field info
type FieldInfo struct {
	name      string
	isIndexed bool
	number    int64
}

// FieldInfos field infos
type FieldInfos struct {
	byNumber []FieldInfo // has order
	byName   map[string]FieldInfo
}

/*
A Term represents a word from text.
This is the unit of search.
It is composed of two elements,
the text of the word, as a string,
and the name of the field that the text occured in, an interned string.

Note that terms may represent more than words from text fields,
but also things like dates, email addresses, urls, etc.
*/

// Term term
type Term struct {
	field string
	text  string
}

// TermInfo term info
type TermInfo struct {
	docFrq int64
	frqPtr int64
	prxPtr int64
}

// Posting posting
// info about a Term in a doc
type Posting struct {
	term      Term    // the Term
	freq      int64   // its frequency in doc
	positions []int64 // positions it occurs at
}

// Keyword keyword type field
func Keyword(name string, value string) (Field, error) {
	f := Field{
		name:        name,
		value:       value,
		isStored:    true,
		isIndexed:   true,
		isTokenized: false,
	}
	return f, nil
}

// ================================FieldInfo=======================================

// isIndexByte get field info index info
func (f *FieldInfo) isIndexByte() byte {
	var b byte
	b = 0
	if f.isIndexed {
		b = 1
	}
	return b
}

// ================================FieldInfos=======================================

// Empty empty byname, bynumber
func (f *FieldInfos) empty() error {
	f.byName = map[string]FieldInfo{}
	f.byNumber = []FieldInfo{}
	return nil
}

// Init init field infos
func (f *FieldInfos) init() error {
	f.addField("", false)
	return nil
}

// AddField add field
func (f *FieldInfos) addField(name string, isIndex bool) error {
	_, found := f.byName[name]
	if !found {
		fieldInfo := FieldInfo{
			name:      name,
			isIndexed: isIndex,
			number:    int64(len(f.byNumber)),
		}
		f.byNumber = append(f.byNumber, fieldInfo)
		f.byName[name] = fieldInfo
	}
	return nil
}

// AddFields add fields
func (f *FieldInfos) addFields(fs *FieldInfos) error {
	for _, fi := range fs.byNumber {
		f.addField(fi.name, fi.isIndexed)
	}
	return nil
}

// AddDoc add doc
func (f *FieldInfos) addDoc(doc Document) error {
	fields := doc.Fields
	for _, field := range fields {
		fieldName := field.name
		if _, found := f.byName[fieldName]; !found { // not in byName
			fi := FieldInfo{
				name:      fieldName,
				isIndexed: field.isIndexed,
				number:    int64(len(f.byNumber)),
			}
			f.byNumber = append(f.byNumber, fi)
			f.byName[fieldName] = fi
		}
	}
	return nil
}

// Write write
func (f *FieldInfos) write(filePath string) error {
	var (
		fPtr *File
		err  error
	)
	fPtr, err = CreateFile(filePath, false, false)
	if err != nil {
		return err
	}

	// (1) write fields size
	fieldSize := len(f.byNumber)
	err = fPtr.writeVarInt(fieldSize)
	if err != nil {
		return err
	}

	if fieldSize > 0 { // first input, last output
		i := 1
		for i <= fieldSize {
			fi := f.byNumber[fieldSize-i]
			// (2) write field name
			err = fPtr.writeString(fi.name)
			if err != nil {
				return err
			}

			// (3) write isIndex info
			fPtr.writeByte(fi.isIndexByte())
			i = i + 1
		}
	}
	return nil
}

// getFieldInfo get field info
func (f *FieldInfos) getFieldInfo(i int) (*FieldInfo, error) {
	if len(f.byNumber) <= i {
		return nil, fmt.Errorf("no such field info")
	}
	fi := f.byNumber[i]
	return &fi, nil
}

// GetNumber get number
func (f *FieldInfos) getNumber(fieldName string) (int64, error) {
	fi, found := f.byName[fieldName]
	if found {
		return fi.number, nil
	}
	return int64(-1), fmt.Errorf("not found field")
}

// getFieldName get field name
func (f *FieldInfos) getFieldName(i int) (string, error) {
	if len(f.byNumber) <= i {
		return "", fmt.Errorf("no such field name")
	}
	str := f.byNumber[i].name
	return str, nil
}

// Init termInfo init
func (ti *TermInfo) Init(docFrq, fp, pp int64) error {
	ti.docFrq = docFrq
	ti.frqPtr = fp
	ti.prxPtr = pp
	return nil
}

// Compare term compare
func (t *Term) compare(d Term) int {
	fc := strings.Compare(t.field, d.field)
	if fc == 0 {
		return strings.Compare(t.text, d.text)
	}
	return fc
}
