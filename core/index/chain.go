package index

// Chain indexing chain
type Chain struct {
	termsVectorsTermsWriter *TermsVectorsTermsWriter
	freqProxTermWriter      *FreqProxTermsWriter
	normsWriter             *NormsWriter
	termsHash               *TermsHash
	docInverter             *DocInverter
}

// DefaultChain default chain
var DefaultChain = Chain{
	// termHashConsumer
}

// DefaultChain

func (dc *Chain) getChain() error {

	return nil
}
