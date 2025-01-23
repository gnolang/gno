package gnokey

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
)

// gnokeyPrivVal is a private validator that uses gnokey to sign proposals and votes.
type gnokeyPrivVal struct {
	keyBase  keys.Keybase
	keyInfo  keys.Info
	password string
}

var _ types.PrivValidator = (*gnokeyPrivVal)(nil)

// GetPubKey implements types.PrivValidator.
func (gk *gnokeyPrivVal) GetPubKey() (crypto.PubKey, error) {
	return gk.keyInfo.GetPubKey(), nil
}

// SignProposal implements types.PrivValidator.
func (gk *gnokeyPrivVal) SignProposal(chainID string, proposal *types.Proposal) error {
	// Sign the proposal.
	sig, _, err := gk.keyBase.Sign(gk.keyInfo.GetName(), gk.password, proposal.SignBytes(chainID))
	if err != nil {
		return fmt.Errorf("unable to sign proposal bytes: %w", err)
	}

	// Save the signature (the proposal will be returned to the client).
	proposal.Signature = sig

	return nil
}

// SignVote implements types.PrivValidator.
func (gk *gnokeyPrivVal) SignVote(chainID string, vote *types.Vote) error {
	// Sign the vote.
	sig, _, err := gk.keyBase.Sign(gk.keyInfo.GetName(), gk.password, vote.SignBytes(chainID))
	if err != nil {
		return fmt.Errorf("unable to sign vote bytes: %w", err)
	}

	// Save the vote (the vote will be returned to the client).
	vote.Signature = sig

	return nil
}

// newGnokeyPrivVal initializes a new gnokey private validator with the provided key name and asks
// the user for a password if necessary.
func newGnokeyPrivVal(
	io commands.IO,
	gnFlags *gnokeyFlags,
	keyName string,
) (*gnokeyPrivVal, error) {
	// Load the keybase located at the home directory.
	keyBase, err := keys.NewKeyBaseFromDir(gnFlags.home)
	if err != nil {
		return nil, fmt.Errorf("unable to load keybase: %w", err)
	}

	// Get the key info from the keybase.
	info, err := keyBase.GetByNameOrAddress(keyName)
	if err != nil {
		return nil, fmt.Errorf("unable to get key from keybase: %w", err)
	}

	var password string

	// Check if a password is required according to the key type.
	switch info.GetType() {
	case keys.TypeLedger: // No password required.
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
	default: // Offline and Multi types are not supported.
		return nil, fmt.Errorf("unsupported key type: %s", info.GetType())
	}

	return &gnokeyPrivVal{
		keyBase:  keyBase,
		keyInfo:  info,
		password: password,
	}, nil
}
