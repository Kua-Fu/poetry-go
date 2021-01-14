package document

import "github.com/Kua-Fu/gsearch/core/analysis"

// AbstractField abstract Field
type AbstractField struct {
	name                        string
	storeTermVector             bool
	storeOffsetWithTermVector   bool
	storePositionWithTermVector bool
	omitNorms                   bool
	isStored                    bool
	isIndexed                   bool
	isTokenized                 bool
	isBinary                    bool
	lazy                        bool
	omitTermFreqAndPositions    bool
	boost                       float64
	fieldsData                  interface{}
	tokenStream                 analysis.TokenStream
	binaryLength                int
	binaryOffset                int
}

func (af *AbstractField) init() {
	af.name = "body"
	af.isIndexed = true
	af.isTokenized = true
	af.boost = 1.0
}

func (af *AbstractField) setBoost(boost float64) error {
	af.boost = boost
	return nil
}

func (af *AbstractField) getBoost() (float64, error) {
	return af.boost, nil
}

func (af *AbstractField) getName() (string, error) {
	return af.name, nil
}

func (af *AbstractField) setStoreTermVector() error {
	return nil
}

func (af *AbstractField) getIsStored() (bool, error) {
	return af.isStored, nil
}

func (af *AbstractField) getIsIndexed() (bool, error) {
	return af.isIndexed, nil
}

func (af *AbstractField) getIsTokenized() (bool, error) {
	return af.isTokenized, nil
}

func (af *AbstractField) getStoreTermVector() (bool, error) {
	return af.storeTermVector, nil
}

func (af *AbstractField) getStoreOffsetWithTermVector() (bool, error) {
	return af.storeOffsetWithTermVector, nil
}

func (af *AbstractField) getStorePositionWithTermVector() (bool, error) {
	return af.storePositionWithTermVector, nil
}

func (af *AbstractField) getIsBinary() (bool, error) {
	return af.isBinary, nil
}
