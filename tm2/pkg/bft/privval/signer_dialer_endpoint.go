package privval

import (
	"log/slog"
	"time"

	"github.com/gnolang/gno/tm2/pkg/service"
)

const (
	DefaultMaxDialRetries      = 10
	DefaultDialRetryIntervalMS = 100
)

// SignerServiceEndpointOption sets an optional parameter on the SignerDialerEndpoint.
type SignerServiceEndpointOption func(*SignerDialerEndpoint)

// SignerDialerEndpointReadWriteTimeout sets the read and write timeout for
// connections from client processes.
func SignerDialerEndpointReadWriteTimeout(timeout time.Duration) SignerServiceEndpointOption {
	return func(ss *SignerDialerEndpoint) { ss.readWriteTimeout = timeout }
}

// SignerDialerEndpointMaxDialRetries sets the amount of attempted retries to
// acceptNewConnection.
func SignerDialerEndpointMaxDialRetries(retries uint) SignerServiceEndpointOption {
	return func(ss *SignerDialerEndpoint) { ss.maxDialRetries = retries }
}

// SignerDialerEndpointDialRetryInterval sets the retry wait interval to a
// custom value.
func SignerDialerEndpointDialRetryInterval(interval time.Duration) SignerServiceEndpointOption {
	return func(ss *SignerDialerEndpoint) { ss.dialRetryInterval = interval }
}

// SignerDialerEndpoint dials using its dialer and responds to any signature
// requests using its privVal.
type SignerDialerEndpoint struct {
	signerEndpoint

	dialer SocketDialer

	maxDialRetries    uint
	dialRetryInterval time.Duration
}

// NewSignerDialerEndpoint returns a SignerDialerEndpoint that will dial using the given
// dialer and respond to any signature requests over the connection
// using the given privVal.
func NewSignerDialerEndpoint(
	logger *slog.Logger,
	dialer SocketDialer,
	options ...SignerServiceEndpointOption,
) *SignerDialerEndpoint {
	sd := &SignerDialerEndpoint{
		dialer:            dialer,
		dialRetryInterval: DefaultDialRetryIntervalMS * time.Millisecond,
		maxDialRetries:    DefaultMaxDialRetries,
	}

	sd.BaseService = *service.NewBaseService(logger, "SignerDialerEndpoint", sd)
	sd.signerEndpoint.readWriteTimeout = DefaultReadWriteTimeoutSeconds * time.Second

	for _, option := range options {
		option(sd)
	}

	return sd
}

func (sd *SignerDialerEndpoint) ensureConnection() error {
	if sd.IsConnected() {
		return nil
	}

	retries := uint(0)
	for retries < sd.maxDialRetries {
		conn, err := sd.dialer()

		if err != nil {
			retries++
			sd.Logger.Debug(
				"SignerDialer: Reconnection failed",
				"retries",
				retries,
				"max",
				sd.maxDialRetries,
				"err",
				err,
			)
			// Wait between retries
			time.Sleep(sd.dialRetryInterval)
		} else {
			sd.SetConnection(conn)
			sd.Logger.Debug("SignerDialer: Connection Ready")
			return nil
		}
	}

	sd.Logger.Debug(
		"SignerDialer: Max retries exceeded",
		"retries",
		retries,
		"max",
		sd.maxDialRetries,
	)

	return ErrNoConnection
}
