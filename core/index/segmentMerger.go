package index

import (
	"os"
)

type SegmentMerger struct {
	directory  *os.File
	segment    string
	fieldInfos FieldInfos
	readers    []interface{}
}
