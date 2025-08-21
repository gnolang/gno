package server

import "errors"

// Errors returned by the remote signer server.
var (
	// Init.
	ErrNilSigner              = errors.New("nil signer")
	ErrInvalidAddressProtocol = errors.New("invalid server address protocol")
	ErrNilLogger              = errors.New("nil logger")

	// Connection.
	ErrListenFailed = errors.New("failed to listen")

	// State.
	ErrServerAlreadyStarted = errors.New("server already started")
	ErrServerAlreadyStopped = errors.New("server already stopped")
)
