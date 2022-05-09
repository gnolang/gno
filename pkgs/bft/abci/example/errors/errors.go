package errors

//----------------------------------------
// Error types

type (
	EncodingError struct{}
	BadNonce      struct{}
	Unauthorized  struct{}
	UnknownError  struct{}
)

//----------------------------------------
// All errors must implement abci.Error

func (_ EncodingError) AssertABCIError() {}
func (_ BadNonce) AssertABCIError()      {}
func (_ Unauthorized) AssertABCIError()  {}
func (_ UnknownError) AssertABCIError()  {}

func (_ EncodingError) Error() string { return "EncodingError" }
func (_ BadNonce) Error() string      { return "BadNonce" }
func (_ Unauthorized) Error() string  { return "Unauthorized" }
func (_ UnknownError) Error() string  { return "UnknownError" }
