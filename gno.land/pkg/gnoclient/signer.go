package gnoclient

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// Signer provides an interface for signing.
type Signer interface {
	Sign(SignCfg) (*std.Tx, error) // returns a signed tx, ready to be broadcasted.
	Info() keys.Info               // returns keys info, containing the address.
	Validate() error               // checks wether the signer is well configured.
}

// SignerFromKeybase represents a signer created from a Keybase.
type SignerFromKeybase struct {
	Keybase  keys.Keybase // Holds keys in memory or on disk
	Account  string       // Could be name or bech32 format
	Password string       // Password for encryption
	ChainID  string
}

func (s SignerFromKeybase) Validate() error {
	if s.ChainID == "" {
		return errors.New("missing ChainID")
	}

	_, err := s.Keybase.GetByNameOrAddress(s.Account)
	if err != nil {
		return err
	}

	// TODO: also verify if the password unlocks the account.
	return nil
}

func (s SignerFromKeybase) Info() keys.Info {
	info, err := s.Keybase.GetByNameOrAddress(s.Account)
	if err != nil {
		panic("should not happen")
	}
	return info
}

// Sign implements the Signer interface for SignerFromKeybase.
type SignCfg struct {
	UnsignedTX     std.Tx
	SequenceNumber uint64
	AccountNumber  uint64
}

func (s SignerFromKeybase) Sign(cfg SignCfg) (*std.Tx, error) {
	tx := cfg.UnsignedTX
	chainID := s.ChainID
	accountNumber := cfg.AccountNumber
	sequenceNumber := cfg.SequenceNumber
	account := s.Account
	password := s.Password

	// fill tx signatures.
	signers := tx.GetSigners()
	if tx.Signatures == nil {
		for range signers {
			tx.Signatures = append(tx.Signatures, std.Signature{
				PubKey:    nil, // zero signature
				Signature: nil, // zero signature
			})
		}
	}

	// validate document to sign.
	err := tx.ValidateBasic()
	if err != nil {
		return nil, err
	}

	// derive sign doc bytes.
	signbz := tx.GetSignBytes(chainID, accountNumber, sequenceNumber)

	sig, pub, err := s.Keybase.Sign(account, password, signbz)
	if err != nil {
		return nil, err
	}
	addr := pub.Address()
	found := false
	for i := range tx.Signatures {
		if signers[i] == addr {
			found = true
			tx.Signatures[i] = std.Signature{
				PubKey:    pub,
				Signature: sig,
			}
		}
	}

	if !found {
		return nil, fmt.Errorf("addr %v (%s) not in signer set", addr, account)
	}

	return &tx, nil
}

// Ensure SignerFromKeybase implements Signer interface.
var _ Signer = (*SignerFromKeybase)(nil)

// SignerFromBip39 creates an in-memory keybase with a single default account.
// This can be useful in scenarios where storing private keys in the filesystem isn't feasible.
//
// Warning: Using keys.NewKeyBaseFromDir is recommended where possible, as it is more secure.
func SignerFromBip39(mnemo string, passphrase string, account uint32, index uint32) (Signer, error) {
	kb := keys.NewInMemory()
	name := "default"
	passwd := "" // Password isn't needed for in-memory storage

	_, err := kb.CreateAccount(name, mnemo, passphrase, passwd, account, index)
	if err != nil {
		return nil, err
	}

	signer := SignerFromKeybase{
		Keybase:  kb,
		Account:  name,
		Password: passwd,
	}

	return &signer, nil
}
