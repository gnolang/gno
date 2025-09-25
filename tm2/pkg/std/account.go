package std

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/errors"

	_ "github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	_ "github.com/gnolang/gno/tm2/pkg/crypto/mock"
	_ "github.com/gnolang/gno/tm2/pkg/crypto/multisig"
	_ "github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
)

// Account is an interface used to store coins at a given address within state.
// It presumes a notion of sequence numbers for replay protection, a notion of
// account numbers for replay protection for previously pruned accounts, and a
// pubkey for authentication purposes.
//
// Many complex conditions can be used in the concrete struct which implements Account.
type Account interface {
	GetAddress() crypto.Address
	SetAddress(crypto.Address) error // errors if already set.

	GetPubKey() crypto.PubKey // can return nil.
	SetPubKey(crypto.PubKey) error

	GetAccountNumber() uint64
	SetAccountNumber(uint64) error

	GetSequence() uint64
	SetSequence(uint64) error

	GetCoins() Coins
	SetCoins(Coins) error

	// Ensure that account implements stringer
	String() string
}

type AccountUnrestricter interface {
	IsTokenLockWhitelisted() bool
	SetTokenLockWhitelisted(bool)
}

//----------------------------------------
// BaseAccount

// BaseAccount - a base account structure.
// This can be extended by embedding within in your *Account structure.
type BaseAccount struct {
	Address       crypto.Address `json:"address" yaml:"address"`
	Coins         Coins          `json:"coins" yaml:"coins"`
	PubKey        crypto.PubKey  `json:"public_key" yaml:"public_key"`
	AccountNumber uint64         `json:"account_number" yaml:"account_number"`
	Sequence      uint64         `json:"sequence" yaml:"sequence"`
}

// NewBaseAccount creates a new BaseAccount object
func NewBaseAccount(address crypto.Address, coins Coins,
	pubKey crypto.PubKey, accountNumber uint64, sequence uint64,
) *BaseAccount {
	return &BaseAccount{
		Address:       address,
		Coins:         coins,
		PubKey:        pubKey,
		AccountNumber: accountNumber,
		Sequence:      sequence,
	}
}

// String implements fmt.Stringer
func (acc BaseAccount) String() string {
	var pubkey string

	if acc.PubKey != nil {
		pubkey = crypto.PubKeyToBech32(acc.PubKey)
	}

	return fmt.Sprintf(`Account:
  Address:       %s
  Pubkey:        %s
  Coins:         %s
  AccountNumber: %d
  Sequence:      %d`,
		acc.Address, pubkey, acc.Coins, acc.AccountNumber, acc.Sequence,
	)
}

// ProtoBaseAccount - a prototype function for BaseAccount
func ProtoBaseAccount() Account {
	return &BaseAccount{}
}

// NewBaseAccountWithAddress - returns a new base account with a given address
func NewBaseAccountWithAddress(addr crypto.Address) BaseAccount {
	return BaseAccount{
		Address: addr,
	}
}

// GetAddress - Implements Account.
func (acc BaseAccount) GetAddress() crypto.Address {
	return acc.Address
}

// SetAddress - Implements Account.
func (acc *BaseAccount) SetAddress(addr crypto.Address) error {
	if !acc.Address.IsZero() {
		return errors.New("cannot override BaseAccount address")
	}
	acc.Address = addr
	return nil
}

// GetPubKey - Implements Account.
func (acc BaseAccount) GetPubKey() crypto.PubKey {
	return acc.PubKey
}

// SetPubKey - Implements Account.
func (acc *BaseAccount) SetPubKey(pubKey crypto.PubKey) error {
	acc.PubKey = pubKey
	return nil
}

// GetCoins - Implements Account.
func (acc *BaseAccount) GetCoins() Coins {
	return acc.Coins
}

// SetCoins - Implements Account.
func (acc *BaseAccount) SetCoins(coins Coins) error {
	acc.Coins = coins
	return nil
}

// GetAccountNumber - Implements Account
func (acc *BaseAccount) GetAccountNumber() uint64 {
	return acc.AccountNumber
}

// SetAccountNumber - Implements Account
func (acc *BaseAccount) SetAccountNumber(accNumber uint64) error {
	acc.AccountNumber = accNumber
	return nil
}

// GetSequence - Implements Account.
func (acc *BaseAccount) GetSequence() uint64 {
	return acc.Sequence
}

// SetSequence - Implements Account.
func (acc *BaseAccount) SetSequence(seq uint64) error {
	acc.Sequence = seq
	return nil
}
