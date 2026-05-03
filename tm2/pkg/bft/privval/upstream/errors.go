package upstream

// errors.go: privval socket-protocol error types.
//
// Mirror of cometbft/privval/errors.go. Kept structurally identical so a
// reviewer comparing the two can match types one-for-one.

import (
	"errors"
	"fmt"
)

// EndpointTimeoutError is returned when a privval endpoint times out
// waiting for a signer connection. Implements the net.Error interface.
type EndpointTimeoutError struct{}

func (e EndpointTimeoutError) Error() string   { return "endpoint connection timed out" }
func (e EndpointTimeoutError) Timeout() bool   { return true }
func (e EndpointTimeoutError) Temporary() bool { return true }

// Socket errors.
var (
	ErrConnectionTimeout = EndpointTimeoutError{}
	ErrNoConnection      = errors.New("endpoint is not connected")
	ErrReadTimeout       = errors.New("endpoint read timed out")
	ErrWriteTimeout      = errors.New("endpoint write timed out")
)

// RemoteSignerErrorWrapper allows the local side to surface a privval
// RemoteSignerError (returned over the wire by tmkms) as a Go error.
//
// Mirrors cometbft/privval/errors.go::RemoteSignerError. Distinct from
// upstreampb.RemoteSignerError (the wire-level message) — this is the
// Go-level error type the validator code sees after unwrapping.
type RemoteSignerErrorWrapper struct {
	Code        int32
	Description string
}

func (e *RemoteSignerErrorWrapper) Error() string {
	return fmt.Sprintf("remote signer error #%d: %s", e.Code, e.Description)
}

// timeoutError matches the net package's timeout-error interface.
// Used to detect whether a returned error from net.Conn read/write was
// due to deadline expiry.
type timeoutError interface {
	Timeout() bool
}
