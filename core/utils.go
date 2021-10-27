package core

import (
	"container/heap"
	"fmt"
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
		"termInfoIndex":   ".tii", // term info index, The index into the Term Infos
		"termInfos":       ".tis", // term infos, part of the term dictionary, stores term info
		"norms":           ".f",   // norms
	}

	// IndexInterval index interval
	IndexInterval int64 = 128
)

// Int64ToByte int64 to []byte
func Int64ToByte(i int64) ([]byte, error) {
	fmt.Println("int64 to byte")
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

// ================================priorityQueue=======================================

//
// type Item struct {
// 	value    string // value of item
// 	priority int    // the priority of the item
// 	index    int    // the index of the item in heap
// }

type PriorityQueue []*SegmentMergeInfo

func (pq PriorityQueue) Len() int {
	return len(pq)
}

// Less compare item in heap
func (pq PriorityQueue) Less(i, j int) bool {
	ri := pq[i]
	rj := pq[j]

	res := (ri.term).compare(*rj.term)
	if res == 0 {
		return ri.base < rj.base
	}
	return res < 0
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = int64(i)
	pq[j].index = int64(j)
}

// Push add new item
func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item, _ := x.(*SegmentMergeInfo)
	item.index = int64(n)
	*pq = append(*pq, item)
}

// Pop pop item
func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[0 : n-1]

	return item
}

// Top get least item
func (pq *PriorityQueue) Top() interface{} {
	if pq.Len() > 0 {
		return (*pq)[0]
	}
	return nil
}

func (pq *PriorityQueue) update(item *SegmentMergeInfo, term *Term) {
	item.term = term
	// item.priority = priority
	heap.Fix(pq, int(item.index))
}
