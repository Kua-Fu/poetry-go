package index

import (
	"os"

	"github.com/Kua-Fu/gsearch/core/analysis"
	"github.com/Kua-Fu/gsearch/core/document"
)

/*
This class accepts multiple added documents and directly writes a single segment file.
It does this more efficiently than creating a single segment per document (with DocumentWriter)
and doing standard merges on those segments.

Each added document is passed to the DocConsumer, which in turn processes the document and interacts with
other consumers in the indexing chain.
Certain consumers, like StoredFieldsWriter and TermVectorsTermsWriter, , digest a document and
immediately write bytes to the "doc store" files
(ie, they do not consume RAM per document, except while they are processing the document).

Other consumers, eg FreqProxTermsWriter and NormsWriter, buffer bytes in RAM and flush only
when a new segment is produced.

Once we have used our allowed RAM buffer, or the number of added docs is large enough
 (in the case we are flushing by doc count instead of RAM usage), we create a real segment and flush it to the Directory.

Threads:
Multiple threads are allowed into addDocument at once.
There is an initial synchronized call to getThreadState which allocates a ThreadState for this thread.
The same thread will get the same ThreadState over time (thread affinity) so that if there are consistent patterns
(for example each thread is indexing a different content source) then we make better use of RAM.
Then processDocument is called on that ThreadState without synchronization
(most of the "heavy lifting" is in this call).
Finally the synchronized "finishDocument" is called to flush changes to the directory.

When flush is called by IndexWriter we forcefully idle all threads and flush only once they are all idle.
This means you can call flush with a given thread even while other threads are actively adding/deleting documents.

Exceptions:
Because this class directly updates in-memory posting lists,
and flushes stored fields and term vectors directly to files in the directory,
there are certain limited times when an exception can corrupt this state.
For example,
a disk full while flushing stored fields leaves this file in a corrupt state.
Or, an OOM exception while appending to the in-memory posting lists can corrupt that posting list.
We call such exceptions "aborting exceptions".
In these cases we must call abort() to discard all docs added since the last flush.

All other exceptions ("non-aborting exceptions") can still partially update the index structures.
These updates are consistent, but, they represent only a part of the document seen up until the exception was hit.
When this happens, we immediately mark the document as deleted so that the document is always atomically
("all or none") added to the index.
*/

// DocumentWriter document writer
type DocumentWriter struct {
	analyzer       analysis.Analyzer
	directory      *os.File
	fieldInfos     FieldInfos
	maxFieldLength int64
	// Writer          Writer
	// Directory       *os.File
	// Segment         string // Current segment we are working on
	// docStoreSegment string // Current doc-store segment we are writing
	// docStoreOffset  int64  // Current starting doc-store offset of current segment
	// nextDocID       int64  // Next docID to be added
	// numDocsInRAM    int64  // docs buffered in RAM
	// numDocsInStore  int64  // docs written to doc stores
}

// AddDocument add doc
func (dw *DocumentWriter) AddDocument(segment string, doc document.Document) (bool, error) {
	return true, nil
}

// FieldInfos field infos
type FieldInfos struct {
	byNumber []interface{}
	byName   map[string]interface{}
}

func (fi *FieldInfos) init(d *os.File, name string) error {
	// InputStream input = d.open(name)
	return nil
}
