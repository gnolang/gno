package common

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bech32"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	stded25519 "golang.org/x/crypto/ed25519"
)

// ServerIdentity defines the server identity keypair.
type ServerIdentity struct {
	PrivKey ed25519.PrivKeyEd25519 `json:"priv_key" comment:"gnokms server private key used to authenticate with clients"`
	PubKey  string                 `json:"pub_key" comment:"gnokms server public key that should be authorized by clients"`
}

// AuthKeysFile defines the content of the auth keys file.
type AuthKeysFile struct {
	ServerIdentity       ServerIdentity `json:"server_identity" comment:"the server identity ed25519 keypair"`
	ClientAuthorizedKeys []string       `json:"authorized_keys" comment:"list of client authorized public keys"`

	authorizedKeys []ed25519.PubKeyEd25519
}

// AuthKeysFile validation errors.
var (
	errInvalidPrivateKey    = errors.New("invalid private key")
	errPublicKeyMismatch    = errors.New("public key does not match private key derivation")
	errInvalidPublicKey     = errors.New("invalid public key")
	errInvalidPublicKeyType = errors.New("not an ed25519 public key")
)

// SortAndDeduplicate sorts and deduplicates the given string slice.
func SortAndDeduplicate(keys []string) []string {
	slices.Sort(keys)
	return slices.Compact(keys)
}

// validate validates the AuthKeysFile.
func (akf *AuthKeysFile) validate() error {
	if akf == nil {
		return fmt.Errorf("%w: auth key is nil", errInvalidPrivateKey)
	}

	// Check if the private key has the expected size for ed25519
	if len(akf.ServerIdentity.PrivKey) != stded25519.PrivateKeySize {
		return fmt.Errorf("%w: private key has invalid length: got %d, want %d",
			errInvalidPrivateKey, len(akf.ServerIdentity.PrivKey), stded25519.PrivateKeySize)
	}

	// Check if the public key portion is properly initialized
	// Ed25519 private keys store the public key in the second half
	pubKeyPortion := akf.ServerIdentity.PrivKey[32:]
	if bytes.Equal(pubKeyPortion, make([]byte, 32)) {
		return fmt.Errorf("%w: public key portion is all uninitialized", errInvalidPublicKey)
	}

	// Verify that the embedded public key matches what we get from PubKey()
	derivedPubKey := akf.ServerIdentity.PrivKey.PubKey()
	if !derivedPubKey.Equals(ed25519.PubKeyEd25519(pubKeyPortion)) {
		return fmt.Errorf("%w: embedded public key doesn't match derived public key", errInvalidPublicKey)
	}

	// Make sure the ServerIdentity public key is derived from the private one.
	if akf.ServerIdentity.PrivKey.PubKey().String() != akf.ServerIdentity.PubKey {
		return errPublicKeyMismatch
	}

	// Check if the address can be encoded using bech32.
	addr := derivedPubKey.Address()
	if _, err := bech32.Encode(crypto.Bech32AddrPrefix, addr[:]); err != nil {
		return fmt.Errorf("%w: unable to encode address: %w", errInvalidPublicKeyType, err)
	}

	// Sort and deduplicate the list of authorized keys.
	akf.ClientAuthorizedKeys = SortAndDeduplicate(akf.ClientAuthorizedKeys)

	// Validate the list of authorized keys.
	for _, authorizedKey := range akf.ClientAuthorizedKeys {
		if _, err := Bech32ToEd25519PubKey(authorizedKey); err != nil {
			return fmt.Errorf("%w: %w", errInvalidPublicKey, err)
		}
	}

	return nil
}

// Save persists the AuthKeysFile to its file path.
func (akf *AuthKeysFile) Save(filePath string) error {
	// Check if the AuthKeysFile is valid.
	if err := akf.validate(); err != nil {
		return err
	}

	// Marshal the AuthKeysFile to JSON bytes using amino.
	jsonBytes, err := amino.MarshalJSONIndent(akf, "", "  ")
	if err != nil {
		return fmt.Errorf("unable to marshal AuthKeysFile to JSON: %w", err)
	}

	// Ensure the parent directory exists.
	parentDir := filepath.Dir(filePath)
	if err := osm.EnsureDir(parentDir, 0o700); err != nil {
		return err
	}

	// Write the JSON bytes to the file.
	if err := osm.WriteFileAtomic(filePath, jsonBytes, 0o600); err != nil {
		return err
	}

	return nil
}

// Bech32ToEd25519PubKey converts a bech32 encoded public key to an ed25519 public key.
func Bech32ToEd25519PubKey(bech32PubKey string) (ed25519.PubKeyEd25519, error) {
	// Decode the bech32 encoded public key.
	pubKey, err := crypto.PubKeyFromBech32(bech32PubKey)
	if err != nil {
		return ed25519.PubKeyEd25519{}, err
	}

	// Check if the public key is an ed25519 public key.
	ed25519PubKey, ok := pubKey.(ed25519.PubKeyEd25519)
	if !ok {
		return ed25519.PubKeyEd25519{}, errInvalidPublicKeyType
	}

	return ed25519PubKey, nil
}

// LoadAuthKeysFile reads an AuthKeysFile from the given file path.
func LoadAuthKeysFile(filePath string) (*AuthKeysFile, error) {
	// Read the JSON bytes from the file.
	rawJSONBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON bytes into a AuthKeysFile using amino.
	akf := &AuthKeysFile{}
	err = amino.UnmarshalJSON(rawJSONBytes, &akf)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal AuthKeysFile from %v: %w", filePath, err)
	}

	// Decode the authorized keys.
	akf.authorizedKeys = make([]ed25519.PubKeyEd25519, len(akf.ClientAuthorizedKeys))

	for i, authorizedKey := range akf.ClientAuthorizedKeys {
		ed25519PubKey, err := Bech32ToEd25519PubKey(authorizedKey)
		if err != nil {
			return nil, err
		}
		akf.authorizedKeys[i] = ed25519PubKey
	}

	// Validate the AuthKeysFile.
	if err := akf.validate(); err != nil {
		return nil, err
	}

	return akf, nil
}

// AuthorizedKeys decodes the bech32 authorized keys from the AuthKeysFile.
func (akf *AuthKeysFile) AuthorizedKeys() []ed25519.PubKeyEd25519 {
	return akf.authorizedKeys
}

// GeneratePersistedAuthKeysFile generates a new AuthKeysFile with a random
// server keypair and empty authorized keys list then persists it to disk.
func GeneratePersistedAuthKeysFile(filePath string) (*AuthKeysFile, error) {
	// Generate a new random private key.
	privKey := ed25519.GenPrivKey()

	// Create a new AuthKeysFile instance.
	afk := &AuthKeysFile{
		ServerIdentity: ServerIdentity{
			PrivKey: privKey,
			PubKey:  privKey.PubKey().String(),
		},
		ClientAuthorizedKeys: []string{},
	}

	// Persist the AuthKeysFile to disk.
	if err := afk.Save(filePath); err != nil {
		return nil, err
	}

	return afk, nil
}
