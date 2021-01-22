package core

import (
	"strconv"
)

// Directory type
// type Directory *os.File

// File file type
// type File *os.File

// Int64ToByte int64 to []byte
func Int64ToByte(i int64) ([]byte, error) {
	return []byte(strconv.FormatInt(i, 10)), nil
}

var (
	// FileSuffix file suffix
	FileSuffix = map[string]string{
		"fieldData":       ".fdt", // field data, The stored fields for documents
		"fieldIndex":      ".fdx", // field index, Contains pointers to field data
		"fieldName":       ".fnm", // field name, Stores information about the fields
		"termFrequencies": ".frq", // term frequencies, Contains the list of docs which contain each term along with frequency
		"termPositions":   ".prx", // term positions, Stores position information about where a term occurs in the index
	}
)
