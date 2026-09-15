package common

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/gnolang/gno/tm2/pkg/amino"
	r "github.com/gnolang/gno/tm2/pkg/bft/privval/signer/remote"
	rss "github.com/gnolang/gno/tm2/pkg/bft/privval/signer/remote/server"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"go.uber.org/multierr"
)

// ErrEmptyAuthorizedKeys is returned when a TCP signer has no authorized clients.
var ErrEmptyAuthorizedKeys = errors.New("TCP listener requires authorized keys")

// NewSignerServer creates a new remote signer server with the given gnokms signer.
func NewSignerServer(
	commonFlags *ServerFlags,
	signer types.Signer,
	logger *slog.Logger,
) (*rss.RemoteSignerServer, error) {
	protocol, _ := osm.ProtocolAndAddress(commonFlags.Listener)

	// Create server options.
	options := []rss.Option{
		rss.WithKeepAlivePeriod(commonFlags.KeepAlivePeriod),
		rss.WithResponseTimeout(commonFlags.ResponseTimeout),
	}

	// Load the auth keys file if it exists for mutual authentication.
	if osm.FileExists(commonFlags.AuthKeysFile) {
		authKeysFile, err := LoadAuthKeysFile(commonFlags.AuthKeysFile)
		if err != nil {
			return nil, fmt.Errorf("invalid auth keys file: %w", err)
		}
		if protocol == r.TCPProtocol && len(authKeysFile.AuthorizedKeys()) == 0 {
			return nil, fmt.Errorf("%w: auth keys file %q contains no authorized keys; run 'gnokms auth authorized add <public-key>'", ErrEmptyAuthorizedKeys, commonFlags.AuthKeysFile)
		}

		// Add the authorized keys and server private key to the server options.
		options = append(options,
			rss.WithAuthorizedKeys(authKeysFile.AuthorizedKeys()),
			rss.WithServerPrivKey(authKeysFile.ServerIdentity.PrivKey),
		)
	} else if protocol == r.TCPProtocol {
		return nil, fmt.Errorf("%w: auth keys file not found at %q; run 'gnokms auth generate', then 'gnokms auth authorized add <public-key>'", ErrEmptyAuthorizedKeys, commonFlags.AuthKeysFile)
	}

	// Initialize the remote signer server with its options.
	server, err := rss.NewRemoteSignerServer(
		signer,
		commonFlags.Listener,
		logger.With("module", "remote_signer_server"),
		options...,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize remote signer server: %w", err)
	}

	return server, err
}

// printValidatorInfo prints the validator info in genesis and bech32 formats.
func printValidatorInfo(signer types.Signer, logger *slog.Logger) error {
	// Check if the signer is nil.
	if signer == nil {
		return errors.New("signer is nil")
	}

	// Get the public key of the signer.
	pubKey := signer.PubKey()

	// Create a genesis validator with the signer's public key.
	genesisValidator := types.GenesisValidator{
		PubKey:  pubKey,
		Address: pubKey.Address(),
		Power:   10,
		Name:    "gnokms_remote_signer",
	}

	// Marshal the genesis validator info to JSON using amino.
	const indent = "  "
	genesisValidatorInfo, err := amino.MarshalJSONIndent(genesisValidator, "", indent)
	if err != nil {
		return fmt.Errorf("unable to marshal genesis validator info to JSON: %w", err)
	}

	// Print the validator info in genesis and bech32 formats.
	logger.Info(fmt.Sprintf("Validator info:\n%s\n%s\n%s\n%s%s%s\n%s%s%s",
		"Genesis format:",
		genesisValidatorInfo,
		"Bech32 format:",
		indent, "pub_key: ", pubKey.String(),
		indent, "address: ", pubKey.Address().String(),
	))

	return nil
}

// RunSignerServer initializes and start a remote signer server with the given gnokms signer.
// It then waits for the server to finish.
func RunSignerServer(ctx context.Context, commonFlags *ServerFlags, signer types.Signer, io commands.IO) error {
	// Initialize the logger.
	logger, flusher, err := LoggerFromServerFlags(commonFlags, io)
	if err != nil {
		return fmt.Errorf("logger initialization failed: %w", err)
	}

	// Flush any remaining server logs on exit.
	defer flusher()

	// Print the validator info of the signer.
	if err := printValidatorInfo(signer, logger); err != nil {
		return fmt.Errorf("unable to print genesis validator info: %w", err)
	}

	// Initialize the remote signer server with the gnokms signer.
	server, err := NewSignerServer(commonFlags, signer, logger)
	if err != nil {
		return fmt.Errorf("signer server initialization failed: %w", err)
	}

	// Start the remote signer server.
	if err := server.Start(); err != nil {
		return fmt.Errorf("signer server start failed: %w", err)
	}

	// Catch SIGINT/SIGTERM/SIGQUIT signals to stop the server gracefully.
	serverCtx, _ := signal.NotifyContext(
		ctx,
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	// Wait for the server context to be done.
	<-serverCtx.Done()

	// Close the server and the signer gracefully.
	return multierr.Combine(
		server.Stop(),
		signer.Close(),
	)
}
