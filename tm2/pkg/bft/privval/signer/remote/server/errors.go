package server

import "errors"

// Errors returned by the remote signer server.
var (
	ErrNoListenAddressProvided = errors.New("no listen address provided")
	ErrServerAlreadyStarted    = errors.New("server already started")
	ErrServerAlreadyStopped    = errors.New("server already stopped")

	// Connection.
	ErrInvalidAddressProtocol = errors.New("invalid server address protocol")
	ErrListenFailed           = errors.New("failed to listen")
)
