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
	output     *File
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
	filePath := path.Join(dirPath, segment+FileSuffix["termInfos"])
	fPtr, err := CreateFile(filePath, false, true)
	if err != nil {
		return err
	}
	tr.output = fPtr
	return nil
}
