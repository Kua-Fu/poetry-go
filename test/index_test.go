package test

import (
	"testing"

	"github.com/Kua-Fu/gsearch/core/analysis"
	"github.com/Kua-Fu/gsearch/core/document"
	"github.com/Kua-Fu/gsearch/core/index"
	"github.com/Kua-Fu/gsearch/core/store"
)

func TestIndex(t *testing.T) {

	//
	field1 := document.Field{
		Name:  "f1",
		Value: "11",
		Index: true,
		Store: true,
	}

	// // 1) gen document
	fields := []document.Field{field1}
	document := document.Document{
		Boost:  1.0,
		Fields: fields,
	}

	analyzer1 := analysis.Analyzer{}

	// // 1) gen writer
	indexDir := "/Users/yz/work/github/gsearch/indexDir"
	dir, err := store.FSDirectory(indexDir)
	if err != nil {
		t.Errorf(err.Error())
	}
	writer := index.Writer{
		Directory: dir,
		Analyzer:  analyzer1,
	}

	// // 3) add doc
	writer.AddDocument(document)

}
