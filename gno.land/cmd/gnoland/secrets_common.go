package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/gnolang/gno/tm2/pkg/amino"
	signer "github.com/gnolang/gno/tm2/pkg/bft/privval/signer/local"
	"github.com/gnolang/gno/tm2/pkg/bft/privval/state"
	"github.com/gnolang/gno/tm2/pkg/p2p/types"
)

var errInvalidNodeKey = errors.New("invalid node p2p key")

// saveNodeKey saves the given NodeKey as Amino JSON to the path
func saveNodeKey(nodeKey *types.NodeKey, path string) error {
	// Get Amino JSON
	marshalledData, err := amino.MarshalJSONIndent(nodeKey, "", "\t")
	if err != nil {
		return fmt.Errorf("unable to marshal NodeKey into JSON, %w", err)
	}

	// Save the data to disk
	if err := os.WriteFile(path, marshalledData, 0o644); err != nil {
		return fmt.Errorf("unable to save data to path, %w", err)
	}

	return nil
}

// isValidDirectory verifies the directory at the given path exists
func isValidDirectory(dirPath string) bool {
	fileInfo, err := os.Stat(dirPath)
	if err != nil {
		return false
	}

	// Check if the path is indeed a directory
	return fileInfo.IsDir()
}

type secretData interface {
	signer.FileKey | state.FileState | types.NodeKey
}

// readNodeKey reads the NodeKey from the given path
func readNodeKey(path string) (*types.NodeKey, error) {
	dataRaw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read data, %w", err)
	}

	var nodeKey types.NodeKey
	if err := amino.UnmarshalJSON(dataRaw, &nodeKey); err != nil {
		return nil, fmt.Errorf("unable to unmarshal NodeKey, %w", err)
	}

	return &nodeKey, nil
}

// validateNodeKey validates the node's p2p key
func validateNodeKey(key *types.NodeKey) error {
	if key.PrivKey == nil {
		return errInvalidNodeKey
	}

	return nil
}

// verifySecretsKey verifies the secrets key value from the passed in arguments
func verifySecretsKey(args []string) error {
	// Check if any key is set
	if len(args) == 0 {
		return nil
	}

	// Check if more than 1 key is set
	if len(args) > 1 {
		return errInvalidSecretsKey
	}

	// Verify the set key
	key := args[0]

	if key != nodeIDKey &&
		key != validatorPrivateKeyKey &&
		key != validatorStateKey {
		return fmt.Errorf(
			"invalid secrets key value [%s, %s, %s]",
			validatorPrivateKeyKey,
			validatorStateKey,
			nodeIDKey,
		)
	}

	return nil
}

// getAvailableSecretsKeys formats and returns the available secret keys (constants)
func getAvailableSecretsKeys() string {
	return fmt.Sprintf(
		"[%s, %s, %s]",
		validatorPrivateKeyKey,
		nodeIDKey,
		validatorStateKey,
	)
}
