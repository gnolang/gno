package client

import "errors"

// Errors returned by the remote signer client.
var (
	// Request.
	ErrSendingRequestFailed  = errors.New("failed to send request")
	ErrInvalidResponseType   = errors.New("invalid response type")
	ErrResponseContainsError = errors.New("response contains error")

	// Connection.
	ErrInvalidAddressProtocol = errors.New("invalid server address protocol")
	ErrMaxRetriesExceeded     = errors.New("maximum retries exceeded")

	// Misc.
	ErrClientAlreadyClosed = errors.New("client already closed")
)
