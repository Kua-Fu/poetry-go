package core

import (
	"path"
)

// FieldsWriter fields writer
type FieldsWriter struct {
	fieldInfos  *FieldInfos
	fieldsData  *File
	fieldsIndex *File
}

// NormsWriter norms writer
type NormsWriter struct {
}

// ================================FieldsWriter=======================================

// Init init fieldsWriter
func (fw *FieldsWriter) init(dirPath string, segment string, fn *FieldInfos) error {
	fw.fieldInfos = fn

	filePath := path.Join(dirPath, segment+FileSuffix["fieldData"])

	fieldsData, err := CreateFile(filePath, false, false)
	if err != nil {
		return err
	}
	fw.fieldsData = fieldsData

	filePath = path.Join(dirPath, segment+FileSuffix["fieldIndex"])
	fieldsIndex, err := CreateFile(filePath, false, false)
	if err != nil {
		return err
	}
	fw.fieldsIndex = fieldsIndex
	return nil
}

// AddDocument add doc
func (fw *FieldsWriter) addDocument(doc Document) error {
	var (
		storedCount int64
		err         error
		size        int64
	)
	size, err = fw.fieldsIndex.getSize()
	if err != nil {
		return err
	}
	err = fw.fieldsIndex.writeInt64(size)
	if err != nil {
		return err
	}
	storedCount = 0
	for _, field := range doc.Fields {
		if field.isStored {
			storedCount = storedCount + 1
		}
	}
	err = fw.fieldsData.writeVarInt64(storedCount)
	if err != nil {
		return err
	}
	for _, field := range doc.Fields {
		if field.isStored {
			fieldName := field.name
			fiNumber, _ := fw.fieldInfos.getNumber(fieldName)
			err = fw.fieldsData.writeVarInt64(fiNumber)
			if err != nil {
				return err
			}
			var bits byte
			bits = 0
			if field.isTokenized {
				bits = bits | 1
			}
			fw.fieldsData.writeByte(bits)
			fw.fieldsData.writeString(field.value)
		}
	}
	return nil
}

// Close flush file to disk
func (fw *FieldsWriter) Close() error {
	var err error
	err = fw.fieldsIndex.flush()
	if err != nil {
		return err
	}
	err = fw.fieldsData.flush()
	if err != nil {
		return err
	}
	return nil
}
