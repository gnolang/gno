package common

import (
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/log"
	sserver "github.com/gnolang/gno/tm2/pkg/bft/privval/signer/remote/server"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
)

// Used to flush the logger.
type logFlusher func()

// NewSignerServer creates a new remote signer server with the given private validator.
func NewSignerServer(
	io commands.IO,
	commonFlags *Flags,
	signer types.Signer,
) (*sserver.RemoteSignerServer, logFlusher, error) {
	// Initialize the zap logger.
	zapLogger, err := log.InitializeLogger(
		io.Out(),
		commonFlags.LogLevel,
		commonFlags.LogFormat,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to initialize zap logger: %w", err)
	}

	// Keep a reference to the zap logger flush function.
	flush := func() { _ = zapLogger.Sync() }

	// Wrap the zap logger with a slog logger.
	logger := log.ZapLoggerToSlog(zapLogger)

	// Split the listen addresses into a slice.
	listenAddresses := strings.Split(commonFlags.ListenAddresses, ",")

	// Initialize the remote signer server with its options.
	server, err := sserver.NewRemoteSignerServer(
		signer,
		listenAddresses,
		logger.With("module", "remote_signer_server"),
		sserver.WithKeepAlivePeriod(commonFlags.KeepAlivePeriod),
		sserver.WithResponseTimeout(commonFlags.ResponseTimeout),
		sserver.WithServerPrivKey(ed25519.GenPrivKey()), // TODO: Add server private key.
		sserver.WithAuthorizedKeys(nil),                 // TODO: Add authorized keys.
	)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to initialize remote signer server: %w", err)
	}

	return server, flush, err
}

// RunSignerServer initializes and start a remote signer server with the given private validator.
// It then waits for the server to finish.
func RunSignerServer(io commands.IO, commonFlags *Flags, signer types.Signer) error {
	// Initialize the remote signer server with the private validator.
	server, flush, err := NewSignerServer(io, commonFlags, signer)
	if err != nil {
		return err
	}

	// Flush any remaining server logs on exit.
	defer flush()

	// Catch SIGINT signal to stop the server gracefully.
	catch := make(chan os.Signal, 1)
	signal.Notify(catch, os.Interrupt)
	go func() {
		<-catch
		io.Println("Caught interrupt signal, stopping server...")
		server.Stop()
	}()

	// Start the remote signer server, then wait for it to finish.
	if err := server.Start(); err != nil {
		return err
	}
	server.Wait()

	return nil
}
