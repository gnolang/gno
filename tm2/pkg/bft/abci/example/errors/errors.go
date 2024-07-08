package errors

// ----------------------------------------
// Error types

type (
	EncodingError     struct{}
	BadNonceError     struct{}
	UnauthorizedError struct{}
	UnknownError      struct{}
)

// ----------------------------------------
// All errors must implement abci.Error

func (EncodingError) AssertABCIError()     {}
func (BadNonceError) AssertABCIError()     {}
func (UnauthorizedError) AssertABCIError() {}
func (UnknownError) AssertABCIError()      {}

func (EncodingError) Error() string     { return "EncodingError" }
func (BadNonceError) Error() string     { return "BadNonceError" }
func (UnauthorizedError) Error() string { return "UnauthorizedError" }
func (UnknownError) Error() string      { return "UnknownError" }
