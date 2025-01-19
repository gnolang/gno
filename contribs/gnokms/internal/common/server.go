package common

import (
	"fmt"

	"github.com/gnolang/gno/gno.land/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/bft/privval"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

// Used to flush the logger.
type logFlusher func()

// NewSignerServer creates a new remote signer server with the given private validator.
func NewSignerServer(
	io commands.IO,
	commonFlags *Flags,
	privVal types.PrivValidator,
) (*privval.SignerServer, logFlusher, error) {
	// Initialize the zap logger.
	zapLogger, err := log.InitializeLogger(
		io.Out(),
		commonFlags.LogLevel,
		commonFlags.LogFormat,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to initialize zap logger, %w", err)
	}

	// Keep a reference to the zap logger flush function.
	flush := func() { _ = zapLogger.Sync() }

	// Wrap the zap logger with a slog logger.
	logger := log.ZapLoggerToSlog(zapLogger)

	// Initialize the signer dialer with the connection parameters.
	dialer, err := privval.NewSignerDialer(commonFlags.NodeAddr, commonFlags.DialTimeout, logger)
	if err != nil {
		return nil, nil, err
	}

	// Initialize the remote signer server with the dialer and the gnokey private validator.
	server := privval.NewSignerServer(dialer, commonFlags.ChainID, privVal)

	return server, flush, nil
}

// RunSignerServer initializes and start a remote signer server with the given private validator.
// It then waits for the server to finish.
func RunSignerServer(io commands.IO, commonFlags *Flags, privVal types.PrivValidator) error {
	// Initialize the remote signer server with the gnokey private validator.
	server, flush, err := NewSignerServer(io, commonFlags, privVal)
	if err != nil {
		return err
	}

	// Flush any remaining server logs on exit.
	defer flush()

	// Start the remote signer server, then wait for it to finish.
	if err := server.Start(); err != nil {
		return err
	}
	server.Wait()

	return nil
}
