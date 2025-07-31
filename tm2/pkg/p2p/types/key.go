package types

import (
	"bytes"
	"fmt"
	"os"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bech32"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/errors"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	stded25519 "golang.org/x/crypto/ed25519"
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
	PrivKey ed25519.PrivKeyEd25519 `json:"priv_key"` // our priv key
}

// ID returns the bech32 representation
// of the node's public p2p key, with
// the bech32 prefix
func (nk *NodeKey) ID() ID {
	return nk.PrivKey.PubKey().Address().ID()
}

// NodeKey validation errors.
var (
	errInvalidNodeKey = errors.New("invalid node p2p key")
)

// validate validates the NodeKey.
func (nk *NodeKey) validate() error {
	if nk == nil {
		return fmt.Errorf("%w: node key is nil", errInvalidNodeKey)
	}

	// Check if the private key has the expected size for ed25519
	if len(nk.PrivKey) != stded25519.PrivateKeySize {
		return fmt.Errorf("%w: private key has invalid length: got %d, want %d",
			errInvalidNodeKey, len(nk.PrivKey), stded25519.PrivateKeySize)
	}

	// Check if the public key portion is properly initialized
	// Ed25519 private keys store the public key in the second half
	pubKeyPortion := nk.PrivKey[32:]
	if bytes.Equal(pubKeyPortion, make([]byte, 32)) {
		return fmt.Errorf("%w: public key portion is all uninitialized", errInvalidNodeKey)
	}

	// Verify that the embedded public key matches what we get from PubKey()
	derivedPubKey := nk.PrivKey.PubKey()
	if !derivedPubKey.Equals(ed25519.PubKeyEd25519(pubKeyPortion)) {
		return fmt.Errorf("%w: embedded public key doesn't match derived public key", errInvalidNodeKey)
	}

	// Check if the address can be encoded using bech32.
	addr := derivedPubKey.Address()
	if _, err := bech32.Encode(crypto.Bech32AddrPrefix, addr[:]); err != nil {
		return fmt.Errorf("%w: unable to encode address: %w", errInvalidNodeKey, err)
	}

	return nil
}

// save persists the NodeKey to its file path.
func (nk *NodeKey) save(filePath string) error {
	// Check if the NodeKey is valid.
	if err := nk.validate(); err != nil {
		return err
	}

	// Marshal the NodeKey to JSON bytes using amino.
	jsonBytes, err := amino.MarshalJSONIndent(nk, "", "  ")
	if err != nil {
		return fmt.Errorf("unable to marshal NodeKey to JSON: %w", err)
	}

	// Write the JSON bytes to the file.
	if err := osm.WriteFileAtomic(filePath, jsonBytes, 0o600); err != nil {
		return err
	}

	return nil
}

// LoadNodeKey loads the node key from the given path.
func LoadNodeKey(filePath string) (*NodeKey, error) {
	// Read the JSON bytes from the file.
	rawJSONBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON bytes into a NodeKey using amino.
	nk := &NodeKey{}
	err = amino.UnmarshalJSON(rawJSONBytes, nk)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal NodeKey from %v: %w", filePath, err)
	}

	// Validate the NodeKey.
	if err := nk.validate(); err != nil {
		return nil, err
	}

	return nk, nil
}

// GenerateNodeKey generates a random node P2P key.
func GenerateNodeKey() *NodeKey {
	return &NodeKey{PrivKey: ed25519.GenPrivKey()}
}

// GeneratePersistedNodeKey generates a new random NodeKey persisted to disk.
func GeneratePersistedNodeKey(filePath string) (*NodeKey, error) {
	// Generate a new random NodeKey.
	fk := GenerateNodeKey()

	// Persist the NodeKey to disk.
	if err := fk.save(filePath); err != nil {
		return nil, err
	}

	return fk, nil
}

// LoadOrMakeNodeKey returns a new NodeKey instance from the given file path.
// If the file does not exist, a new NodeKey is generated and persisted to disk.
func LoadOrMakeNodeKey(filePath string) (*NodeKey, error) {
	// If the file exists, load the NodeKey from the file.
	if osm.FileExists(filePath) {
		return LoadNodeKey(filePath)
	}

	// If the file does not exist, generate a new NodeKey and persist it to disk.
	return GeneratePersistedNodeKey(filePath)
}
