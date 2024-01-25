package main

import (
	"fmt"
	"os"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/privval"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/p2p"
)

// generateValidatorPrivateKey generates the validator's private key
func generateValidatorPrivateKey() privval.FilePVKey {
	privKey := ed25519.GenPrivKey()

	return privval.FilePVKey{
		Address: privKey.PubKey().Address(),
		PubKey:  privKey.PubKey(),
		PrivKey: privKey,
	}
}

// saveValidatorPrivateKey saves the validator's private key to the given path
func saveValidatorPrivateKey(key privval.FilePVKey, path string) error {
	// Get Amino JSON
	marshalledKey, err := amino.MarshalJSONIndent(key, "", "  ")
	if err != nil {
		return fmt.Errorf("unable to marshal validator private key into JSON, %w", err)
	}

	// Save the key to disk
	if err := os.WriteFile(path, marshalledKey, 0o644); err != nil {
		return fmt.Errorf("unable to save validator private key, %w", err)
	}

	return nil
}

// generateLastSignValidatorState generates the empty last sign state
func generateLastSignValidatorState() privval.FilePVLastSignState {
	return privval.FilePVLastSignState{} // Empty last sign state
}

// saveLastSignValidatorState saves the last sign validator state to the given path
func saveLastSignValidatorState(state privval.FilePVLastSignState, path string) error {
	// Get Amino JSON
	marshalledState, err := amino.MarshalJSONIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("unable to marshal last validator sign state into JSON, %w", err)
	}

	// Save the sign state to disk
	if err := os.WriteFile(path, marshalledState, 0o644); err != nil {
		return fmt.Errorf("unable to save last validator sign state, %w", err)
	}

	return nil
}

// generateNodeKey generates the p2p node key
func generateNodeKey() *p2p.NodeKey {
	privKey := ed25519.GenPrivKey()

	return &p2p.NodeKey{
		PrivKey: privKey,
	}
}

// saveNodeKey saves the node key to the given path
func saveNodeKey(key *p2p.NodeKey, path string) error {
	// Get Amino JSON
	marshalledKey, err := amino.MarshalJSON(key)
	if err != nil {
		return fmt.Errorf("unable to marshal node key into JSON, %w", err)
	}

	// Save the sign state to disk
	if err := os.WriteFile(path, marshalledKey, 0o644); err != nil {
		return fmt.Errorf("unable to save node key, %w", err)
	}

	return nil
}
