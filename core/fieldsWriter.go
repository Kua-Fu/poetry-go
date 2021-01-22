package core

// FieldsWriter fields writer
type FieldsWriter struct {
	FieldInfos  FieldInfos
	FieldsData  *File
	FieldsIndex *File
}

// Init init fieldsWriter
func (fw *FieldsWriter) Init(dirPath string, segment string, fn FieldInfos) error {
	fw.FieldInfos = fn

	filePath := dirPath + segment + FileSuffix["fieldData"]

	fieldsData, err := CreateFile(filePath, false, false)
	if err != nil {
		return err
	}
	fw.FieldsData = fieldsData

	filePath = dirPath + segment + FileSuffix["fieldIndex"]
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
	)
	storedCount = 0
	for _, field := range doc.Fields {
		if field.IsStored {
			storedCount = storedCount + 1
		}
	}
	err = fw.FieldsData.WriteInt64(storedCount)
	if err != nil {
		return err
	}
	for _, field := range doc.Fields {
		if field.IsStored {
			fieldName := field.Name
			fiNumber, _ := fw.FieldInfos.GetNumber(fieldName)
			err = fw.FieldsData.WriteInt64(fiNumber)
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
