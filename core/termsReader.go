package core

import "path"

func (tr *TermsReader) init(dirPath string, segment string, fn *FieldInfos) error {

	var (
		filePath string
		dataPtr  *File
		indexPtr *File
		err      error
	)

	// tis
	filePath = path.Join(dirPath, segment+FileSuffix["termInfos"])
	dataPtr, err = CreateFile(filePath, false, true)
	if err != nil {
		return err
	}
	tr.termsData = dataPtr

	// tii
	filePath = path.Join(dirPath, segment+FileSuffix["termInfoIndex"])
	indexPtr, err = CreateFile(filePath, false, true)
	if err != nil {
		return err
	}
	tr.termsIndex = indexPtr

	tr.fieldInfos = fn

	tr.readIndex()
	return nil
}

// terms get terms
func (tr *TermsReader) terms() (*SegmentTerms, error) {

	return tr.segTerms, nil
}

// readIndex terms read index
func (tr *TermsReader) readIndex() error {
	segTerms := new(SegmentTerms)

	// get all term, termInfo
	segTerms.init(tr.termsIndex, tr.termsData, tr.fieldInfos, true)

	tr.segTerms = segTerms

	// indexSize := segTerms.size
	// indexTerms := []Term{}
	// for i < indexSize{

	// 	indexTerms = append(indexTerms, )
	// }

	return nil
}
