package core

// NError custom error
type NError struct {
	Step string // the logic step raise error
	Err  error  // source error
}

// Error string error
func (n *NError) Error() string {
	return "step: " + n.Step + "; " + n.Err.Error()
}
