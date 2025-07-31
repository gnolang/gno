package local

import (
	"fmt"
	"os"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/errors"
	osm "github.com/gnolang/gno/tm2/pkg/os"
)

// FileKey is a struct that contains the private key, public key, and address of a
// FileSigner. It is persisted to disk in JSON format using amino.
// NOTE: keep in sync with gno.land/cmd/gnoland/secrets.go
type FileKey struct {
	PrivKey crypto.PrivKey `json:"priv_key" comment:"the validator private key"`
	PubKey  crypto.PubKey  `json:"pub_key" comment:"the validator public key"`
	Address types.Address  `json:"address" comment:"the validator address"`
}

// FileKey validation errors.
var (
	errInvalidPrivateKey = errors.New("invalid private key")
	errPublicKeyMismatch = errors.New("public key does not match private key derivation")
	errAddressMismatch   = errors.New("address does not match public key")
)

// validate validates the FileKey.
func (fk *FileKey) validate() error {
	// Make sure the private key is set.
	if fk.PrivKey == nil {
		return errInvalidPrivateKey
	}

	// Make sure the public key is derived from the private one.
	if !fk.PrivKey.PubKey().Equals(fk.PubKey) {
		return errPublicKeyMismatch
	}

	// Make sure the address is derived from the public key.
	if fk.PubKey.Address().Compare(fk.Address) != 0 {
		return errAddressMismatch
	}

	return nil
}

// save persists the FileKey to its file path.
func (fk *FileKey) save(filePath string) error {
	// Check if the FileKey is valid.
	if err := fk.validate(); err != nil {
		return err
	}

	// Marshal the FileKey to JSON bytes using amino.
	jsonBytes, err := amino.MarshalJSONIndent(fk, "", "  ")
	if err != nil {
		return fmt.Errorf("unable to marshal FileKey to JSON: %w", err)
	}

	// Write the JSON bytes to the file.
	if err := osm.WriteFileAtomic(filePath, jsonBytes, 0o600); err != nil {
		return err
	}

	return nil
}

// LoadFileKey reads a FileKey from the given file path.
func LoadFileKey(filePath string) (*FileKey, error) {
	// Read the JSON bytes from the file.
	rawJSONBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON bytes into a FileKey using amino.
	fk := &FileKey{}
	err = amino.UnmarshalJSON(rawJSONBytes, fk)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal FileKey from %v: %w", filePath, err)
	}

	// Validate the FileKey.
	if err := fk.validate(); err != nil {
		return nil, err
	}

	return fk, nil
}

// GenerateFileKey generates a new random FileKey.
func GenerateFileKey() *FileKey {
	// Generate a new random private key.
	privKey := ed25519.GenPrivKey()

	// Create a new FileKey instance.
	return &FileKey{
		PrivKey: privKey,
		PubKey:  privKey.PubKey(),
		Address: privKey.PubKey().Address(),
	}
}

// GeneratePersistedFileKey generates a new random FileKey persisted to disk.
func GeneratePersistedFileKey(filePath string) (*FileKey, error) {
	// Generate a new random FileKey.
	fk := GenerateFileKey()

	// Persist the FileKey to disk.
	if err := fk.save(filePath); err != nil {
		return nil, err
	}

	return fk, nil
}

// LoadOrMakeFileKey returns a new FileKey instance from the given file path.
// If the file does not exist, a new FileKey is generated and persisted to disk.
func LoadOrMakeFileKey(filePath string) (*FileKey, error) {
	// If the file exists, load the FileKey from the file.
	if osm.FileExists(filePath) {
		return LoadFileKey(filePath)
	}

	// If the file does not exist, generate a new FileKey and persist it to disk.
	return GeneratePersistedFileKey(filePath)
}
