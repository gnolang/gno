package client

import "errors"

// Errors returned by the remote signer client.
var (
	// Init.
	ErrInvalidAddressProtocol = errors.New("invalid client address protocol")
	ErrNilLogger              = errors.New("nil logger")
	ErrFetchingPubKeyFailed   = errors.New("failed to fetch public key")

	// Request.
	ErrSendingRequestFailed  = errors.New("failed to send request")
	ErrInvalidResponseType   = errors.New("invalid response type")
	ErrResponseContainsError = errors.New("response contains error")

	// Connection.
	ErrMaxRetriesExceeded = errors.New("maximum retries exceeded")

	// State.
	ErrClientAlreadyClosed = errors.New("client already closed")
)
