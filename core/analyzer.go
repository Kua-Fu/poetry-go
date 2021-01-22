package core

// Analyzer text to token
type Analyzer struct {
}

// TokenSlice get token slice
func (ay *Analyzer) TokenSlice() ([]Token, error) {

	return nil, nil
}

/*
A Token is an occurence of a term from the text of a field.
It consists of a term's text,
the start and end offset of the term in the text of the field,
and a type string.

The start and end offsets permit applications to re-associate a token with its source text,
e.g.,
to display highlighted query terms in a document browser,
or to show matching text fragments in a KWIC (KeyWord In Context) display,
etc.

The type is an interned string,
assigned by a lexical analyzer (a.k.a. tokenizer),
naming the lexical or syntactic class that the token belongs to.
For example an end of sentence marker token might be implemented with type "eos".
The default token type is "word".
*/

// Token token
type Token struct {
	TermText    string // the text of the term
	StartOffset int64  // start in source text
	EndOffset   int64  // end in source text
	Type        string // lexical type
}
