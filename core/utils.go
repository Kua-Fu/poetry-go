package core

import (
	"math"
	"strconv"
)

var (
	// FileSuffix file suffix
	FileSuffix = map[string]string{
		"fieldData":       ".fdt", // field data, The stored fields for documents
		"fieldIndex":      ".fdx", // field index, Contains pointers to field data
		"fieldName":       ".fnm", // field name, Stores information about the fields
		"termFrequencies": ".frq", // term frequencies, Contains the list of docs which contain each term along with frequency
		"termPositions":   ".prx", // term positions, Stores position information about where a term occurs in the index
		"termInfoIndex":   ".tii", // term info index, The index into the Term Infos file
		"termInfos":       ".tis", // term infos, part of the term dictionary, stores term info
		"norms":           ".f",   // norms
	}
)

// Int64ToByte int64 to []byte
func Int64ToByte(i int64) ([]byte, error) {
	return []byte(strconv.FormatInt(i, 2)), nil
}

// StringDifference get string difference
func StringDifference(s, d string) int64 {
	var l, i int = 0, 0
	lenS := len(s) // byte length
	lenD := len(d) // byte length
	if len(s) < len(d) {
		l = lenS
	} else {
		l = lenD
	}
	for i < l {
		if s[i] != d[i] {
			return int64(i)
		}
		i = i + 1
	}
	return int64(l)
}

//SimilarityNorm similarity norm
func SimilarityNorm(n int64) byte {
	f := float64(n)
	d := 255.0 / math.Sqrt(f)
	return byte(math.Ceil(d))
}
