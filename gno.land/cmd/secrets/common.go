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

// generateLastSignValidatorState generates the empty last sign state
func generateLastSignValidatorState() privval.FilePVLastSignState {
	return privval.FilePVLastSignState{} // Empty last sign state
}

// generateNodeKey generates the p2p node key
func generateNodeKey() *p2p.NodeKey {
	privKey := ed25519.GenPrivKey()

	return &p2p.NodeKey{
		PrivKey: privKey,
	}
}

// saveDataToPath saves the given data as Amino JSON to the path
func saveDataToPath(data any, path string) error {
	// Get Amino JSON
	marshalledState, err := amino.MarshalJSONIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("unable to marshal data into JSON, %w", err)
	}

	// Save the data to disk
	if err := os.WriteFile(path, marshalledState, 0o644); err != nil {
		return fmt.Errorf("unable to save data to disk, %w", err)
	}

	return nil
}
