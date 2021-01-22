package core

import (
	"math"
	"strconv"
)

/*
An IndexWriter creates and maintains an index.
The third argument to the constructor determines whether a new index is created,
or whether an existing index is opened for the addition of new documents.

In either case, documents are added with the addDocument method,
When finished adding documents, close should be called.

If an index will not have more documents added for a while and optimal search performance is desired,
then the optimize method should be called before the index is closed.
*/

// Writer index writer
type Writer struct {
	directory    File         // where this index resides
	analyzer     Analyzer     // how to analyze text
	segInfos     SegmentInfos // the segments
	ramDirectory File         // for temp segs
}

var (
	maxFieldLength int64 = 10000
	// Determines how often segment indexes are merged by addDocument().
	// With smaller values, less RAM is used while indexing,
	// and searches on unoptimized indexes are faster, but indexing speed is slower.
	// With larger values more RAM is used while indexing and searches on unoptimized indexes are slower,
	// but indexing is faster.
	// Thus larger values (> 10) are best for batched index creation,
	// and smaller values (< 10) for indexes that are interactively maintained.
	// This must never be less than 2.  The default value is 10.
	mergeFactor int64 = 10

	// Determines the largest number of documents ever merged by addDocument()
	// Small values (e.g., less than 10,000) are best for interactive indexing,
	// as this limits the length of pauses while indexing to a few seconds.
	// Larger values are best for batched indexing and speedier searches.
	// The default value is math.MaxInt64
	maxMergeDocs int64 = math.MaxInt64

	infoStream string
)

// Init init writer
func (w *Writer) Init(dir File, analyzer Analyzer) error {
	w.directory = dir
	w.analyzer = analyzer
	return nil

}

// AddDocument Adds a document to this index
func (w *Writer) AddDocument(doc Document) error {
	// dw := DocumentWriter{
	// 	DirPath:        w.ramDirectory,
	// 	Analyzer:       w.analyzer,
	// 	MaxFieldLength: maxFieldLength,
	// }
	// segmentName := w.newSegmentName()
	// dw.AddDocument(segmentName, doc)

	// segmentInfo := SegmentInfo{
	// 	name:     segmentName,
	// 	docCount: 1,
	// 	dir:      w.ramDirectory,
	// }
	// w.segInfos.addElement(segmentInfo)
	// w.maybeMergeSegments()
	return nil
}

func (w *Writer) newSegmentName() string {
	nCounter := w.segInfos.counter + 1
	return "_" + strconv.FormatInt(nCounter, 10)

}

func (w *Writer) maybeMergeSegments() error {
	targetMergeDocs := mergeFactor
	for targetMergeDocs <= maxMergeDocs {
		minSegment := int64(len(w.segInfos.segInfos))
		mergeDocs := int64(0)
		for minSegment >= 0 {
			si, _ := w.segInfos.info(minSegment)
			if si.docCount >= targetMergeDocs {
				break
			}
			mergeDocs = mergeDocs + si.docCount
			minSegment = minSegment - 1
		}

		if mergeDocs >= targetMergeDocs { // found a merge to do
			w.mergeSegments(minSegment + 1)
		} else {
			break
		}
		targetMergeDocs = targetMergeDocs * mergeFactor
	}
	// while (targetMergeDocs <= maxMergeDocs) {

	// }
	return nil

}

// Pops segments off of segmentInfos stack down to minSegment, merges them,
// and pushes the merged index onto the top of the segmentInfos stack.
func (w *Writer) mergeSegments(minSegment int64) error {
	// mergedName := w.newSegmentName()
	// mergedDocCount := int64(0)
	// if infoStream != "" {
	// 	fmt.Println("merging segments")
	// }
	// merger := SegmentMerger{
	// 	directory: w.directory,
	// 	segment:   mergedName,
	// }
	// segmentsToDelete := []interface{}{}
	// for _, si := range w.segInfos.segInfos {
	// 	if infoStream != "" {
	// 		fmt.Println(" " + si.name + " (" + strconv.FormatInt(si.docCount, 10) + " docs)")
	// 	}
	// 	SegmentReader reader = new SegmentReader(si)
	// }
	return nil
}

// SegmentInfos segment infos
type SegmentInfos struct {
	counter  int64
	segInfos []SegmentInfo
}

func (s *SegmentInfos) info(i int64) (SegmentInfo, error) {
	return s.segInfos[i], nil
}

func (s *SegmentInfos) addElement(segInfo SegmentInfo) error {
	s.segInfos = append(s.segInfos, segInfo)
	return nil
}

// SegmentInfo  segment info
type SegmentInfo struct {
	name     string
	docCount int64
	dir      File
}
