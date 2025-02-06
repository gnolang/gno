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
type FileKey struct {
	Address types.Address  `json:"address" comment:"the validator address"`
	PubKey  crypto.PubKey  `json:"pub_key" comment:"the validator public key"`
	PrivKey crypto.PrivKey `json:"priv_key" comment:"the validator private key"`

	filePath string
}

// save persists the FileKey to its file path.
func (fk *FileKey) save() error {
	// Check if the file path is set.
	if fk.filePath == "" {
		return errors.New("filePath not set")
	}

	// Marshal the FileKey to JSON bytes using amino.
	jsonBytes, err := amino.MarshalJSONIndent(fk, "", "  ")
	if err != nil {
		return err
	}

	// Write the JSON bytes to the file.
	if err = osm.WriteFileAtomic(fk.filePath, jsonBytes, 0o600); err != nil {
		return err
	}

	return nil
}

// loadFileKey reads a FileKey from the given file path.
func loadFileKey(filePath string) (*FileKey, error) {
	// Read the JSON bytes from the file.
	rawJSONBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON bytes into a FileKey using amino.
	fk := &FileKey{}
	err = amino.UnmarshalJSON(rawJSONBytes, &fk)
	if err != nil {
		return nil, fmt.Errorf("unable to read FileKey from %v: %v\n", filePath, err)
	}

	// Overwrite pubkey and address for convenience.
	fk.PubKey = fk.PrivKey.PubKey()
	fk.Address = fk.PubKey.Address()
	fk.filePath = filePath

	return fk, nil
}

// generateFileKey generate a new random FileKey and persists it to disk.
func generateFileKey(filePath string) (*FileKey, error) {
	// Generate a new random private key.
	privKey := ed25519.GenPrivKey()

	// Create a new FileKey instance.
	fk := &FileKey{
		Address:  privKey.PubKey().Address(),
		PubKey:   privKey.PubKey(),
		PrivKey:  privKey,
		filePath: filePath,
	}

	// Persist the FileKey to disk.
	if err := fk.save(); err != nil {
		return nil, err
	}

	return fk, nil
}

// NewFileKey returns a new FileKey instance from the given file path.
// If the file does not exist, a new FileKey is generated and persisted to disk.
func NewFileKey(filePath string) (*FileKey, error) {
	// If the file exists, load the FileKey from the file.
	if osm.FileExists(filePath) {
		return loadFileKey(filePath)
	}

	// If the file does not exist, generate a new FileKey and persist it to disk.
	return generateFileKey(filePath)
}
