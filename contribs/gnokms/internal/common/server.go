package common

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/amino"
	sserver "github.com/gnolang/gno/tm2/pkg/bft/privval/signer/remote/server"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	osm "github.com/gnolang/gno/tm2/pkg/os"
)

// NewSignerServer creates a new remote signer server with the given gnokms signer.
func NewSignerServer(
	commonFlags *ServerFlags,
	signer types.Signer,
	logger *slog.Logger,
) (*sserver.RemoteSignerServer, error) {
	// Split the listen addresses into a slice.
	listenAddresses := strings.Split(commonFlags.ListenAddresses, ",")

	// Create server options.
	options := []sserver.Option{
		sserver.WithKeepAlivePeriod(commonFlags.KeepAlivePeriod),
		sserver.WithResponseTimeout(commonFlags.ResponseTimeout),
	}

	// Load the auth keys file if it exists for mutual authentication.
	if osm.FileExists(commonFlags.AuthKeysFile) {
		authKeysFile, err := LoadAuthKeysFile(commonFlags.AuthKeysFile)
		if err != nil {
			return nil, fmt.Errorf("invalid auth keys file: %w", err)
		}

		// Get the authorized keys from the auth keys file.
		authorizedKeys, err := authKeysFile.AuthorizedKeys()
		if err != nil { // Will be caught by only if the authorized keys are invalid.
			return nil, fmt.Errorf("unable to get authorized keys from auth keys file: %w", err)
		}

		// Add the authorized keys and server private key to the server options.
		options = append(options,
			sserver.WithAuthorizedKeys(authorizedKeys),
			sserver.WithServerPrivKey(*authKeysFile.ServerIdentity.PrivKey),
		)
	} else {
		// Log a warning if the auth keys file does not exist.
		logger.Warn("Mutual auth keys not found, gnokms and its clients will not be able to authenticate")
		logger.Warn("For more security, generate mutual auth keys using 'gnokms auth generate'")
	}

	// Initialize the remote signer server with its options.
	server, err := sserver.NewRemoteSignerServer(
		signer,
		listenAddresses,
		logger.With("module", "remote_signer_server"),
		options...,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize remote signer server: %w", err)
	}

	return server, err
}

// genesisValidatorInfoFromSigner gets a genesis validator info from the given signer.
func genesisValidatorInfoFromSigner(signer types.Signer) (string, error) {
	// Get the public key of the signer.
	pubKey, err := signer.PubKey()
	if err != nil {
		return "", fmt.Errorf("unable to get signer public key: %w", err)
	}

	// Create a genesis validator with the signer's public key.
	genesisValidator := types.GenesisValidator{
		PubKey:  pubKey,
		Address: pubKey.Address(),
		Power:   10,
		Name:    "gnokms_remote_signer",
	}

	// Marshal the genesis validator info to JSON using amino.
	genesisValidatorInfo, err := amino.MarshalJSONIndent(genesisValidator, "", "  ")
	if err != nil {
		return "", fmt.Errorf("unable to marshal genesis validator info: %w", err)
	}

	return string(genesisValidatorInfo), nil
}

// RunSignerServer initializes and start a remote signer server with the given gnokms signer.
// It then waits for the server to finish.
func RunSignerServer(commonFlags *ServerFlags, signer types.Signer, io commands.IO) error {
	// Initialize the logger.
	logger, flusher, err := LoggerFromServerFlags(commonFlags, io)
	if err != nil {
		return fmt.Errorf("logger initialization failed: %w", err)
	}

	// Flush any remaining server logs on exit.
	defer flusher()

	// Print the public key of the signer as a genesis validator.
	info, err := genesisValidatorInfoFromSigner(signer)
	if err != nil {
		return fmt.Errorf("unable to print genesis validator info: %w", err)
	}
	logger.Info(fmt.Sprintf("Genesis validator info:\n%s", info))

	// Initialize the remote signer server with the gnokms signer.
	server, err := NewSignerServer(commonFlags, signer, logger)
	if err != nil {
		return fmt.Errorf("signer server initialization failed: %w", err)
	}

	// Catch SIGINT signal to stop the server gracefully.
	catch := make(chan os.Signal, 1)
	signal.Notify(catch, os.Interrupt)
	go func() {
		<-catch
		logger.Info("Caught interrupt signal, stopping signer server...")
		server.Stop()
	}()

	// Start the remote signer server, then wait for it to finish.
	if err := server.Start(); err != nil {
		return fmt.Errorf("signer server start failed: %w", err)
	}
	server.Wait()

	return nil
}
