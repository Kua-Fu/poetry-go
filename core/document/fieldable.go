package document

import (
	"io"

	"github.com/Kua-Fu/gsearch/core/analysis"
)

// Fieldable fieldable
type Fieldable interface {
	setBoost(float64)
	getBoost() float64
	getName() string
	stringValue() string
	readerValue() io.Reader
	tokenStreamValue() analysis.TokenStream
}
