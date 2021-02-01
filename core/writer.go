package core

import (
	"math"
	"path"
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
	dir      *File         // where this index resides
	analyzer Analyzer      // how to analyze text
	segInfos *SegmentInfos // the segments
	ramDir   *File         // for temp segs
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
)

// Init init writer
func (w *Writer) Init(Dirpath string, analyzer Analyzer, create bool) error {
	fPtr, err := CreateFile(Dirpath, true, true)
	if err != nil {
		return err
	}

	w.dir = fPtr
	w.analyzer = analyzer

	segsPtr := new(SegmentInfos)
	segsPtr.empty()

	w.segInfos = segsPtr

	tempDir := "/tmp"
	tPtr, err := CreateTempFile(tempDir, "pre", true)
	if err != nil {
		return err
	}

	w.ramDir = tPtr

	if create {
		w.segInfos.write(fPtr)
	}
	return nil

}

// AddDocument Adds a document to this index
func (w *Writer) AddDocument(doc Document) error {

	dw := new(DocumentWriter)
	dw.Init(w.ramDir.filePath, w.analyzer, MaxFieldLength)
	segment := w.newSegName()
	dw.AddDocument(segment, doc)

	seg := SegmentInfo{
		name:     segment,
		docCount: 1,
		dirPath:  w.ramDir.filePath,
	}

	w.segInfos.add(seg)

	w.maybeMergeSegs()

	return nil
}

// newSegName new segment name
func (w *Writer) newSegName() string {
	nCounter := w.segInfos.counter + 1
	return "_" + strconv.FormatInt(nCounter, 10)
}

// Close close
func (w *Writer) Close() error {
	w.flushRAMSegs()
	w.closeRAMDir()
	return nil

}

// FlushRAMSegs flush ram segments
func (w *Writer) flushRAMSegs() error {
	minSegment := int64(0)
	w.mergeSegs(minSegment)
	return nil
}

// delete dir
func (w *Writer) closeRAMDir() error {
	w.ramDir.removeAll()
	return nil
}

// MaybeMergeSegs merge segs
func (w *Writer) maybeMergeSegs() error {

	targetMergeDocs := MergeFactor // 10

	for targetMergeDocs <= MaxMergeDocs {

		minSegment := int64(len(w.segInfos.segInfos))
		mergeDocs := int64(0)

		minSegment = minSegment - 1 // minSegment use as index

		for minSegment >= 0 {
			si := w.segInfos.segInfos[minSegment]
			if si.docCount >= targetMergeDocs {
				break
			}
			mergeDocs = mergeDocs + si.docCount
			minSegment = minSegment - 1
		}

		if mergeDocs >= targetMergeDocs { // found a merge to do
			w.mergeSegs(minSegment + 1) // -1 + 1 = 0
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

// mergeSegs merge segments
func (w *Writer) mergeSegs(minSegment int64) error {

	mergedName := w.newSegName()
	mergedDocCount := int64(0)

	merger := SegmentMerger{
		dirPath: w.dir.filePath,
		name:    mergedName,
		readers: []*SegmentReader{},
	}

	segsToDelete := []*SegmentReader{} // segment to delete

	for _, si := range w.segInfos.segInfos {

		reader := new(SegmentReader)
		reader.init(si)
		merger.add(reader)

		segsToDelete = append(segsToDelete, reader)

		mergedDocCount = mergedDocCount + si.docCount

	}

	merger.merge()

	// w.SegInfos; // pop old infos & add new
	seg := SegmentInfo{
		name:     mergedName,
		docCount: mergedDocCount,
		dirPath:  w.dir.filePath,
	}
	segs := SegmentInfos{
		counter:  2,
		segInfos: []SegmentInfo{seg},
	}
	w.segInfos = &segs

	w.segInfos.write(w.dir) // commit before deleting

	w.deleteSegments(segsToDelete) // delete now-unused segments

	return nil
}

//
func (w *Writer) deleteSegments(segsToDelete []*SegmentReader) error {
	// get all files should be deleted
	deleteFiles, err := w.readDeleteableFiles()
	if err != nil {
		return err
	}

	// get all files can delete, maybe some files can not delete current
	deleteables, err := w.deleteFiles(deleteFiles)
	if err != nil {
		return err
	}

	w.writeDeleteableFiles(deleteables)
	return nil
}

// read delete files
func (w *Writer) readDeleteableFiles() ([]string, error) {
	return []string{}, nil
}

// all delete files
func (w *Writer) deleteFiles(deleteFiles []string) ([]string, error) {

	return []string{}, nil
}

// writeDeleteableFiles
func (w *Writer) writeDeleteableFiles(deleteables []string) error {
	filePath := path.Join(w.dir.filePath, "deleteable.new")

	dPtr, err := CreateFile(filePath, false, false)
	if err != nil {
		return err
	}

	dPtr.writeInt(len(deleteables))

	for _, fileName := range deleteables {
		dPtr.writeString(fileName)
	}

	dPtr.close()
	nfilepath := path.Join(w.dir.filePath, "deletable")
	dPtr.rename(nfilepath)
	return nil
}
