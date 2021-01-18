package index

import "os"

// SegmentReader segment reader
type SegmentReader struct {
	directory *os.File
	segment   string
}

func (sr *SegmentReader) init(si SegmentInfo) error {
	sr.directory = si.dir
	sr.segment = si.name
	// fieldInfos := new(FieldInfos)
	// fieldInfos.init()
	return nil
}
