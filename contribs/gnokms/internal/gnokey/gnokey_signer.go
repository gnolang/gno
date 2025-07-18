package gnokey

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
)

// gnokeySigner is a gno-kms signer based on gnokey.
type gnokeySigner struct {
	keyBase  keys.Keybase
	keyInfo  keys.Info
	password string
}

// gnokeySigner type implements types.Signer.
var _ types.Signer = (*gnokeySigner)(nil)

// PubKey implements types.Signer.
func (gk *gnokeySigner) PubKey() crypto.PubKey {
	return gk.keyInfo.GetPubKey()
}

// Sign implements types.Signer.
func (gk *gnokeySigner) Sign(signBytes []byte) ([]byte, error) {
	signature, _, err := gk.keyBase.Sign(gk.keyInfo.GetName(), gk.password, signBytes)
	return signature, err
}

// Close implements types.Signer.
func (gk *gnokeySigner) Close() error {
	gk.keyBase.CloseDB()
	return nil
}

// newGnokeySigner initializes a new gnokey signer with the provided key name and asks
// the user for a password if necessary.
func newGnokeySigner(
	gnFlags *gnokeyFlags,
	keyName string,
	io commands.IO,
) (*gnokeySigner, error) {
	// Load the keybase located at the home directory.
	keyBase, _ := keys.NewKeyBaseFromDir(gnFlags.home)

	// Get the key info from the keybase.
	info, err := keyBase.GetByNameOrAddress(keyName)
	if err != nil {
		return nil, fmt.Errorf("unable to get key from keybase: %w", err)
	}

	var password string

	// Check if a password is required according to the key type.
	switch info.GetType() {
	case keys.TypeLocal:
		for {
			// Get the password from the user.
			password, err = io.GetPassword(
				"Enter password to decrypt the key",
				gnFlags.insecurePasswordStdin,
			)
			if err != nil {
				return nil, fmt.Errorf("unable to get decryption key: %w", err)
			}

			// Check if the password is correct.
			if _, _, err = keyBase.Sign(keyName, password, []byte{}); err != nil {
				io.ErrPrintln("Invalid password, try again\n")
				continue
			}

			break
		}
		// TODO: suggest to use `gnokms ledger` when it will be implemented.
	// case keys.TypeLedger: // Ledger is not supported
	// 	return nil, fmt.Errorf("unsupported key type: use 'gnokms ledger' for ledger keys")
	default: // Offline and Multi types are not supported.
		return nil, fmt.Errorf("unsupported key type: %s", info.GetType())
	}

	return &gnokeySigner{
		keyBase:  keyBase,
		keyInfo:  info,
		password: password,
	}, nil
}
