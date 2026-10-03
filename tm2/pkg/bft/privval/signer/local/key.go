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

// KeyType identifies a signing scheme for the validator file key.
type KeyType string

const (
	KeyTypeEd25519   KeyType = "ed25519"
	KeyTypeSecp256k1 KeyType = "secp256k1"
)

// DefaultKeyType is the scheme used by GenerateFileKey when no type is
// specified. Kept at ed25519 for backwards compatibility with existing
// tooling and CI.
const DefaultKeyType = KeyTypeEd25519

// ParseKeyType returns a KeyType for the given string, or an error if the
// value is not a recognised scheme.
func ParseKeyType(s string) (KeyType, error) {
	switch KeyType(s) {
	case KeyTypeEd25519:
		return KeyTypeEd25519, nil
	case KeyTypeSecp256k1:
		return KeyTypeSecp256k1, nil
	default:
		return "", fmt.Errorf("unsupported validator key type %q (want %q or %q)",
			s, KeyTypeEd25519, KeyTypeSecp256k1)
	}
}

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

// GenerateFileKey generates a new random FileKey using DefaultKeyType.
// Kept for backwards compatibility; new callers should use
// GenerateFileKeyOfType to be explicit about the scheme.
func GenerateFileKey() *FileKey {
	fk, err := GenerateFileKeyOfType(DefaultKeyType)
	if err != nil {
		// DefaultKeyType is a compile-time constant we control, so
		// any error here would be a programmer mistake. Panic instead
		// of swallowing.
		panic(fmt.Sprintf("GenerateFileKey: %v", err))
	}
	return fk
}

// GenerateFileKeyOfType generates a new random FileKey using the given scheme.
func GenerateFileKeyOfType(keyType KeyType) (*FileKey, error) {
	var privKey crypto.PrivKey
	switch keyType {
	case KeyTypeEd25519:
		privKey = ed25519.GenPrivKey()
	case KeyTypeSecp256k1:
		privKey = secp256k1.GenPrivKey()
	default:
		return nil, fmt.Errorf("unsupported validator key type %q", keyType)
	}

	return &FileKey{
		PrivKey: privKey,
		PubKey:  privKey.PubKey(),
		Address: privKey.PubKey().Address(),
	}, nil
}

// GeneratePersistedFileKey generates a new random FileKey of DefaultKeyType
// persisted to disk.
func GeneratePersistedFileKey(filePath string) (*FileKey, error) {
	return GeneratePersistedFileKeyOfType(filePath, DefaultKeyType)
}

// GeneratePersistedFileKeyOfType generates a new random FileKey of the given
// scheme persisted to disk.
func GeneratePersistedFileKeyOfType(filePath string, keyType KeyType) (*FileKey, error) {
	fk, err := GenerateFileKeyOfType(keyType)
	if err != nil {
		return nil, err
	}

	if err := fk.save(filePath); err != nil {
		return nil, err
	}

	return fk, nil
}

// LoadOrMakeFileKey returns a new FileKey instance from the given file path.
// If the file does not exist, a new FileKey is generated and persisted to disk
// using DefaultKeyType.
func LoadOrMakeFileKey(filePath string) (*FileKey, error) {
	// If the file exists, load the FileKey from the file.
	if osm.FileExists(filePath) {
		return LoadFileKey(filePath)
	}

	// If the file does not exist, generate a new FileKey and persist it to disk.
	return GeneratePersistedFileKey(filePath)
}
