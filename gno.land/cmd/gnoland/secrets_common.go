package main

import (
	"errors"
	"fmt"
	"os"

	fstate "github.com/gnolang/gno/tm2/pkg/bft/privval/state"
	"github.com/gnolang/gno/tm2/pkg/crypto"
)

var errSignatureMismatch = errors.New("signature does not match signature bytes")

// isValidDirectory verifies the directory at the given path exists
func isValidDirectory(dirPath string) bool {
	fileInfo, err := os.Stat(dirPath)
	if err != nil {
		return false
	}

	// Check if the path is indeed a directory
	return fileInfo.IsDir()
}

// validateValidatorStateSignature validates the signature section
// of the last sign validator state
func validateValidatorStateSignature(
	state *fstate.FileState,
	key crypto.PubKey,
) error {
	if state.Signature == nil {
		// No need to verify further
		return nil
	}

	// Make sure the signature bytes match the signature
	if !key.VerifyBytes(state.SignBytes, state.Signature) {
		return errSignatureMismatch
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
