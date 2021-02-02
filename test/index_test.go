package test

import (
	"testing"

	"github.com/Kua-Fu/gsearch/core"
)

func TestDoc(t *testing.T) {
	var (
		err      error
		indexDir string
	)

	indexDir = "/Users/yz/work/github/gsearch/test/index/"

	analyzer := new(core.Analyzer)
	writer := new(core.DocumentWriter)
	writer.Init(indexDir, *analyzer, int64(1000))

	doc := new(core.Document)
	f1, err := core.Keyword("path", "/etc/test.txt")
	if err != nil {
		t.Errorf("doc index err")
	}
	doc.Add(f1)

	segment := "s1"
	writer.AddDocument(segment, *doc)
}

func TestDocWithMulitFields(t *testing.T) {
	var (
		indexDir string
	)

	indexDir = "/Users/yz/work/github/gsearch/test/index/"

	analyzer := new(core.Analyzer)
	writer := new(core.DocumentWriter)
	writer.Init(indexDir, *analyzer, int64(1000))

	doc := new(core.Document)
	f1, _ := core.Keyword("path", "/etc/test.txt")
	f2, _ := core.Keyword("filename", "test.txt")
	doc.Add(f1)
	doc.Add(f2)

	segment := "s1"
	writer.AddDocument(segment, *doc)
}

func TestIndex(t *testing.T) {
	var (
		indexDir string
	)

	indexDir = "/Users/yz/work/github/gsearch/test/index1/"

	analyzer := new(core.Analyzer)
	writer := new(core.Writer)
	writer.Init(indexDir, *analyzer, true)

	doc := new(core.Document)

	f1, _ := core.Keyword("path", "/etc/test.txt")
	f2, _ := core.Keyword("path2", "/etc/test1.txt")
	f3, _ := core.Keyword("path3", "/etc/test2.txt")
	f4, _ := core.Keyword("path4", "/etc/test3.txt")

	doc.Add(f1)
	doc.Add(f2)
	doc.Add(f3)
	doc.Add(f4)

	writer.AddDocument(*doc)
	writer.Close()

}
