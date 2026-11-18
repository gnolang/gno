package local

import (
	"fmt"
	"os"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
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
	errUnknownKeyType    = errors.New("unknown validator key type")
)

// KeyType names the signature scheme of a validator consensus key.
type KeyType string

const (
	KeyTypeSecp256k1 KeyType = "secp256k1"
	KeyTypeEd25519   KeyType = "ed25519"

	// DefaultKeyType is secp256k1 so that commits can be verified cheaply
	// with ecrecover by EVM light clients.
	DefaultKeyType = KeyTypeSecp256k1
)

// keyGenerators maps each supported key type to its random key generator.
var keyGenerators = map[KeyType]func() crypto.PrivKey{
	KeyTypeSecp256k1: func() crypto.PrivKey { return secp256k1.GenPrivKey() },
	KeyTypeEd25519:   func() crypto.PrivKey { return ed25519.GenPrivKey() },
}

// Validate returns an error if k is not a supported key type. Callers that
// write files should call it before touching the disk, so a mistyped type
// leaves nothing behind.
func (k KeyType) Validate() error {
	if _, ok := keyGenerators[k]; !ok {
		return fmt.Errorf("%w: %q", errUnknownKeyType, k)
	}

	return nil
}

// genPrivKey generates a random private key of the given type.
func genPrivKey(keyType KeyType) (crypto.PrivKey, error) {
	if err := keyType.Validate(); err != nil {
		return nil, err
	}

	return keyGenerators[keyType](), nil
}

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

// GenerateFileKey generates a new random FileKey of the DefaultKeyType.
func GenerateFileKey() *FileKey {
	fk, err := GenerateFileKeyOfType(DefaultKeyType)
	if err != nil {
		panic(err) // DefaultKeyType is always known
	}

	return fk
}

// GenerateFileKeyOfType generates a new random FileKey of the given type.
func GenerateFileKeyOfType(keyType KeyType) (*FileKey, error) {
	// Generate a new random private key.
	privKey, err := genPrivKey(keyType)
	if err != nil {
		return nil, err
	}

	// Create a new FileKey instance.
	return &FileKey{
		PrivKey: privKey,
		PubKey:  privKey.PubKey(),
		Address: privKey.PubKey().Address(),
	}, nil
}

// GeneratePersistedFileKey generates a new random FileKey of the
// DefaultKeyType persisted to disk.
func GeneratePersistedFileKey(filePath string) (*FileKey, error) {
	return GeneratePersistedFileKeyOfType(filePath, DefaultKeyType)
}

// GeneratePersistedFileKeyOfType generates a new random FileKey of the given
// type persisted to disk.
func GeneratePersistedFileKeyOfType(filePath string, keyType KeyType) (*FileKey, error) {
	// Generate a new random FileKey.
	fk, err := GenerateFileKeyOfType(keyType)
	if err != nil {
		return nil, err
	}

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
