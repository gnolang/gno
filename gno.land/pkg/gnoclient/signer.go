package gnoclient

import (
	"fmt"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// Signer provides an interface for signing transactions.
type Signer interface {
	Sign(SignCfg) (*std.Tx, error) // Signs a transaction and returns a signed tx ready for broadcasting.
	Info() keys.Info               // Returns key information, including the address.
	Validate() error               // Checks whether the signer is properly configured.
}

// SignerFromKeybase represents a signer created from a Keybase.
type SignerFromKeybase struct {
	Keybase  keys.Keybase // Stores keys in memory or on disk
	Account  string       // Account name or bech32 format
	Password string       // Password for encryption
	ChainID  string       // Chain ID for transaction signing
}

// Validate checks if the signer is properly configured.
func (s SignerFromKeybase) Validate() error {
	if s.ChainID == "" {
		return errors.New("missing ChainID")
	}

	_, err := s.Keybase.GetByNameOrAddress(s.Account)
	if err != nil {
		return err
	}

	// To verify if the password unlocks the account, sign a blank transaction.
	msg := vm.MsgCall{
		Caller: s.Info().GetAddress(),
	}
	signCfg := SignCfg{
		UnsignedTX: std.Tx{
			Msgs: []std.Msg{msg},
			Fee:  std.NewFee(0, std.NewCoin("ugnot", 1000000)),
		},
	}
	if _, err = s.Sign(signCfg); err != nil {
		return err
	}

	return nil
}

// Info gets keypair information.
func (s SignerFromKeybase) Info() keys.Info {
	info, err := s.Keybase.GetByNameOrAddress(s.Account)
	if err != nil {
		panic("should not happen")
	}
	return info
}

// SignCfg provides the signing configuration, containing:
// unsigned transaction data, account number, and account sequence.
type SignCfg struct {
	UnsignedTX     std.Tx
	SequenceNumber uint64
	AccountNumber  uint64
}

// Sign implements the Signer interface for SignerFromKeybase.
func (s SignerFromKeybase) Sign(cfg SignCfg) (*std.Tx, error) {
	tx := cfg.UnsignedTX
	chainID := s.ChainID
	accountNumber := cfg.AccountNumber
	sequenceNumber := cfg.SequenceNumber
	account := s.Account
	password := s.Password

	// Initialize tx signatures.
	signers := tx.GetSigners()
	if tx.Signatures == nil {
		for range signers {
			tx.Signatures = append(tx.Signatures, std.Signature{
				PubKey:    nil, // Zero signature
				Signature: nil, // Zero signature
			})
		}
	}

	// Validate the transaction to sign.
	err := tx.ValidateBasic()
	if err != nil {
		return nil, err
	}

	// Derive sign doc bytes.
	signbz, err := tx.GetSignBytes(chainID, accountNumber, sequenceNumber)
	if err != nil {
		return nil, fmt.Errorf("unable to get tx signature payload, %w", err)
	}

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
		return nil, fmt.Errorf("address %v (%s) not in signer set", addr, account)
	}

	return &tx, nil
}

// Ensure SignerFromKeybase implements the Signer interface.
var _ Signer = (*SignerFromKeybase)(nil)

// SignerFromBip39 creates a signer from an in-memory keybase with a single default account, derived from the given mnemonic.
// This can be useful in scenarios where storing private keys in the filesystem isn't feasible.
//
// Warning: Using keys.NewKeyBaseFromDir to get a keypair from local storage is recommended where possible, as it is more secure.
func SignerFromBip39(mnemonic string, chainID string, passphrase string, account uint32, index uint32) (Signer, error) {
	kb := keys.NewInMemory()
	name := "default"
	password := "" // Password isn't needed for in-memory storage

	_, err := kb.CreateAccount(name, mnemonic, passphrase, password, account, index)
	if err != nil {
		return nil, err
	}

	signer := SignerFromKeybase{
		Keybase:  kb,
		Account:  name,
		Password: password,
		ChainID:  chainID,
	}

	return &signer, nil
}
