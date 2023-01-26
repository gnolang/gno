package errors

// ----------------------------------------
// Error types

type (
	EncodingError struct{}
	BadNonce      struct{}
	Unauthorized  struct{}
	UnknownError  struct{}
)

// ----------------------------------------
// All errors must implement abci.Error

func (EncodingError) AssertABCIError() {}
func (BadNonce) AssertABCIError()      {}
func (Unauthorized) AssertABCIError()  {}
func (UnknownError) AssertABCIError()  {}

func (EncodingError) Error() string { return "EncodingError" }
func (BadNonce) Error() string      { return "BadNonce" }
func (Unauthorized) Error() string  { return "Unauthorized" }
func (UnknownError) Error() string  { return "UnknownError" }
