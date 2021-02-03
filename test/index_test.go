package test

import (
	"testing"

	"github.com/Kua-Fu/gsearch/core"
)

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
	f2, _ := core.Keyword("filename", "test.txt")

	doc.Add(f1)
	doc.Add(f2)

	writer.AddDocument(*doc)
	writer.Close()

}

func TestIndexWithMultiFields(t *testing.T) {
	var (
		indexDir string
	)

	indexDir = "/Users/yz/work/github/gsearch/test/index1/"

	analyzer := new(core.Analyzer)
	writer := new(core.Writer)
	writer.Init(indexDir, *analyzer, true)

	doc := new(core.Document)

	f1, _ := core.Keyword("path", "/etc/test.txt")
	f2, _ := core.Keyword("filename", "test.txt")

	doc.Add(f1)
	doc.Add(f2)

	writer.AddDocument(*doc)
	writer.Close()

}
