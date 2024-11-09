package types

import (
	"fmt"
	"os"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	osm "github.com/gnolang/gno/tm2/pkg/os"
)

// ID represents the cryptographically unique Peer ID
type ID = crypto.ID

// NewIDFromStrings returns an array of ID's build using
// the provided strings
func NewIDFromStrings(idStrs []string) ([]ID, []error) {
	var (
		ids  = make([]ID, 0, len(idStrs))
		errs = make([]error, 0, len(idStrs))
	)

	for _, idStr := range idStrs {
		id := ID(idStr)
		if err := id.Validate(); err != nil {
			errs = append(errs, err)

			continue
		}

		ids = append(ids, id)
	}

	return ids, errs
}

// NodeKey is the persistent peer key.
// It contains the nodes private key for authentication.
// NOTE: keep in sync with gno.land/cmd/gnoland/secrets.go
type NodeKey struct {
	crypto.PrivKey `json:"priv_key"` // our priv key
}

// ID returns the bech32 representation
// of the node's public p2p key, with
// the bech32 prefix
func (k NodeKey) ID() ID {
	return k.PubKey().Address().ID()
}

// LoadOrGenNodeKey attempts to load the NodeKey from the given filePath.
// If the file does not exist, it generates and saves a new NodeKey.
func LoadOrGenNodeKey(path string) (*NodeKey, error) {
	// Check if the key exists
	if osm.FileExists(path) {
		// Load the node key
		return LoadNodeKey(path)
	}

	// Key is not present on path,
	// generate a fresh one
	nodeKey := GenerateNodeKey()
	if err := saveNodeKey(path, nodeKey); err != nil {
		return nil, fmt.Errorf("unable to save node key, %w", err)
	}

	return nodeKey, nil
}

// LoadNodeKey loads the node key from the given path
func LoadNodeKey(path string) (*NodeKey, error) {
	// Load the key
	jsonBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read key, %w", err)
	}

	var nodeKey NodeKey

	// Parse the key
	if err = amino.UnmarshalJSON(jsonBytes, &nodeKey); err != nil {
		return nil, fmt.Errorf("unable to JSON unmarshal node key, %w", err)
	}

	return &nodeKey, nil
}

// GenerateNodeKey generates a random
// node P2P key, based on ed25519
func GenerateNodeKey() *NodeKey {
	privKey := ed25519.GenPrivKey()

	return &NodeKey{
		PrivKey: privKey,
	}
}

// saveNodeKey saves the node key
func saveNodeKey(path string, nodeKey *NodeKey) error {
	// Get Amino JSON
	marshalledData, err := amino.MarshalJSONIndent(nodeKey, "", "\t")
	if err != nil {
		return fmt.Errorf("unable to marshal node key into JSON, %w", err)
	}

	// Save the data to disk
	if err := os.WriteFile(path, marshalledData, 0o644); err != nil {
		return fmt.Errorf("unable to save node key to path, %w", err)
	}

	return nil
}
