package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/privval"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/p2p"
)

var (
	errInvalidPrivateKey = errors.New("invalid validator private key")
	errPublicKeyMismatch = errors.New("public key does not match private key derivation")
	errAddressMismatch   = errors.New("address does not match public key")

	errInvalidSignStateStep   = errors.New("invalid sign state step value")
	errInvalidSignStateHeight = errors.New("invalid sign state height value")
	errInvalidSignStateRound  = errors.New("invalid sign state round value")

	errSignatureMismatch      = errors.New("signature does not match signature bytes")
	errSignatureValuesMissing = errors.New("missing signature value")

	errInvalidNodeKey = errors.New("invalid node p2p key")
)

// saveSecretData saves the given data as Amino JSON to the path
func saveSecretData(data any, path string) error {
	// Get Amino JSON
	marshalledData, err := amino.MarshalJSONIndent(data, "", "\t")
	if err != nil {
		return fmt.Errorf("unable to marshal data into JSON, %w", err)
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
	privval.FilePVKey | privval.FilePVLastSignState | p2p.NodeKey
}

// readSecretData reads the secret data from the given path
func readSecretData[T secretData](
	path string,
) (*T, error) {
	dataRaw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read data, %w", err)
	}

	var data T
	if err := amino.UnmarshalJSON(dataRaw, &data); err != nil {
		return nil, fmt.Errorf("unable to unmarshal data, %w", err)
	}

	return &data, nil
}

// validateValidatorKey validates the validator's private key
func validateValidatorKey(key *privval.FilePVKey) error {
	// Make sure the private key is set
	if key.PrivKey == nil {
		return errInvalidPrivateKey
	}

	// Make sure the public key is derived
	// from the private one
	if !key.PrivKey.PubKey().Equals(key.PubKey) {
		return errPublicKeyMismatch
	}

	// Make sure the address is derived
	// from the public key
	if key.PubKey.Address().Compare(key.Address) != 0 {
		return errAddressMismatch
	}

	return nil
}

// validateValidatorState validates the validator's last sign state
func validateValidatorState(state *privval.FilePVLastSignState) error {
	// Make sure the sign step is valid
	if state.Step < 0 {
		return errInvalidSignStateStep
	}

	// Make sure the height is valid
	if state.Height < 0 {
		return errInvalidSignStateHeight
	}

	// Make sure the round is valid
	if state.Round < 0 {
		return errInvalidSignStateRound
	}

	return nil
}

// validateValidatorStateSignature validates the signature section
// of the last sign validator state
func validateValidatorStateSignature(
	state *privval.FilePVLastSignState,
	key crypto.PubKey,
) error {
	// Make sure the signature and signature bytes are valid
	signBytesPresent := state.SignBytes != nil
	signaturePresent := state.Signature != nil

	if signBytesPresent && !signaturePresent ||
		!signBytesPresent && signaturePresent {
		return errSignatureValuesMissing
	}

	if !signaturePresent {
		// No need to verify further
		return nil
	}

	// Make sure the signature bytes match the signature
	if !key.VerifyBytes(state.SignBytes, state.Signature) {
		return errSignatureMismatch
	}

	return nil
}

// validateNodeKey validates the node's p2p key
func validateNodeKey(key *p2p.NodeKey) error {
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

	if key != nodeKeyKey &&
		key != validatorPrivateKeyKey &&
		key != validatorStateKey {
		return fmt.Errorf(
			"invalid secrets key value [%s, %s, %s]",
			validatorPrivateKeyKey,
			validatorStateKey,
			nodeKeyKey,
		)
	}

	return nil
}

// getAvailableSecretsKeys formats and returns the available secret keys (constants)
func getAvailableSecretsKeys() string {
	return fmt.Sprintf(
		"[%s, %s, %s]",
		validatorPrivateKeyKey,
		nodeKeyKey,
		validatorStateKey,
	)
}
