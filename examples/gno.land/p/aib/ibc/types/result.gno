package types

// ResponseResultType defines the possible outcomes of the execution of a message
type ResponseResultType int32

const (
	// Default zero value enumeration
	RESPONSE_UNSPECIFIED ResponseResultType = 0
	// The message did not call the IBC application callbacks (because, for example, the packet had already been relayed)
	RESPONSE_NOOP ResponseResultType = 1
	// The message was executed successfully
	RESPONSE_SUCCESS ResponseResultType = 2
	// The message was executed unsuccessfully
	RESPONSE_FAILURE ResponseResultType = 3
)
