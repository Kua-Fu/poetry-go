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
	Directory    *File        // where this index resides
	Analyzer     Analyzer     // how to analyze text
	SegInfos     SegmentInfos // the segments
	RAMDirectory *File        // for temp segs
}

var (
	// MaxFieldLength max field length
	MaxFieldLength int64 = 10000

	// Determines how often segment indexes are merged by addDocument().
	// With smaller values, less RAM is used while indexing,
	// and searches on unoptimized indexes are faster, but indexing speed is slower.
	// With larger values more RAM is used while indexing and searches on unoptimized indexes are slower,
	// but indexing is faster.
	// Thus larger values (> 10) are best for batched index creation,
	// and smaller values (< 10) for indexes that are interactively maintained.
	// This must never be less than 2.  The default value is 10.

	// MergeFactor merge factor
	MergeFactor int64 = 10

	// Determines the largest number of documents ever merged by addDocument()
	// Small values (e.g., less than 10,000) are best for interactive indexing,
	// as this limits the length of pauses while indexing to a few seconds.
	// Larger values are best for batched indexing and speedier searches.
	// The default value is math.MaxInt64

	// MaxMergeDocs max merge docs
	MaxMergeDocs int64 = math.MaxInt64

	infoStream string
)

// Init init writer
func (w *Writer) Init(Dirpath string, analyzer Analyzer, create bool) error {
	fPtr, err := CreateFile(Dirpath, true, true)
	if err != nil {
		return err
	}

	w.Directory = fPtr
	w.Analyzer = analyzer

	segs := SegmentInfos{
		Counter:  0,
		SegInfos: []SegmentInfo{},
	}
	w.SegInfos = segs

	tempDir := "/Users/yz/work/github/gsearch/test/"
	tPtr, err := CreateTempFile(tempDir, "pre", true)
	if err != nil {
		return err
	}

	w.RAMDirectory = tPtr

	if create {
		w.SegInfos.Write(fPtr)
	}
	return nil

}

// AddDocument Adds a document to this index
func (w *Writer) AddDocument(doc Document) error {

	dw := new(DocumentWriter)
	dw.Init(w.RAMDirectory.FilePath, w.Analyzer, int64(1000))
	segment := w.NewSegName()
	dw.AddDocument(segment, doc)

	seg := SegmentInfo{
		Name:     segment,
		DocCount: 1,
		Dirpath:  w.RAMDirectory.FilePath,
	}

	w.SegInfos.AddElement(seg)

	w.MaybeMergeSegs()
	return nil
}

// NewSegName new segment name
func (w *Writer) NewSegName() string {
	nCounter := w.SegInfos.Counter + 1
	return "_" + strconv.FormatInt(nCounter, 10)
}

// Close close
func (w *Writer) Close() error {
	w.FlushRAMSegs()
	return nil

}

// FlushRAMSegs flush ram segments
func (w *Writer) FlushRAMSegs() error {
	minSegment := int64(0)
	w.MergeSegs(minSegment)
	return nil
}

// MaybeMergeSegs merge segs
func (w *Writer) MaybeMergeSegs() error {

	targetMergeDocs := MergeFactor

	for targetMergeDocs <= MaxMergeDocs {

		minSegment := int64(len(w.SegInfos.SegInfos))
		mergeDocs := int64(0)

		for minSegment >= 0 {
			si, _ := w.SegInfos.Info(minSegment)
			if si.DocCount >= targetMergeDocs {
				break
			}
			mergeDocs = mergeDocs + si.DocCount
			minSegment = minSegment - 1
		}

		if mergeDocs >= targetMergeDocs { // found a merge to do
			w.MergeSegs(minSegment + 1)
		} else {
			break
		}
		targetMergeDocs = targetMergeDocs * MergeFactor
	}
	// while (targetMergeDocs <= maxMergeDocs) {

	// }
	return nil

}

// Pops segments off of segmentInfos stack down to minSegment, merges them,
// and pushes the merged index onto the top of the segmentInfos stack.

// MergeSegs merge segments
func (w *Writer) MergeSegs(minSegment int64) error {
	mergedName := w.NewSegName()
	mergedDocCount := int64(0)
	// if infoStream != "" {
	// 	fmt.Println("merging segments")
	// }
	merger := SegmentMerger{
		DirPath: w.Directory.FilePath,
		Name:    mergedName,
		Readers: []*SegmentReader{},
	}
	segsToDelete := []*SegmentReader{}
	for _, si := range w.SegInfos.SegInfos {

		reader := new(SegmentReader)
		reader.Init(si)
		merger.Add(reader)

		segsToDelete = append(segsToDelete, reader)

		mergedDocCount = mergedDocCount + si.DocCount

	}

	// merger.Merge()

	// w.SegInfos; // pop old infos & add new
	seg := SegmentInfo{
		Name:     mergedName,
		DocCount: mergedDocCount,
		Dirpath:  w.Directory.FilePath,
	}
	segs := SegmentInfos{
		Counter:  1,
		SegInfos: []SegmentInfo{seg},
	}
	w.SegInfos = segs

	w.SegInfos.Write(w.Directory)

	return nil
}
