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
// It represents the core identity and assets, without authentication details.
//
// It presumes a notion of account numbers for replay protection for previously
// pruned accounts, a notion of coins, which are an amount of a specific asset.
// It does not presume a notion of pubkeys or authentication.
//
// Many complex conditions can be used in the concrete struct which implements Account.
type Account interface {
	GetAddress() crypto.Address
	SetAddress(crypto.Address) error // errors if already set.

	GetAccountNumber() uint64
	SetAccountNumber(uint64) error

	GetSequenceSum() uint64
	SetSequenceSum(uint64) error

	GetCoins() Coins
	SetCoins(Coins) error

	// Master key access
	SetMasterKey(crypto.PubKey) (AccountKey, error)
	GetMasterKey() AccountKey

	// Session management
	AddSession(pubKey crypto.PubKey) (AccountKey, error)
	DelSession(pubKey crypto.PubKey) error
	DelAllSessions() error

	// Key getters
	GetKey(pubKey crypto.PubKey) (AccountKey, error)
	GetAllKeys() []AccountKey

	String() string
}

// AccountKey represents authentication methods for an account.
// This can be either a MasterKey or a Session.
// XXX: support realm "keys" in the future
type AccountKey interface {
	GetPubKey() crypto.PubKey
	SetPubKey(crypto.PubKey) error
	GetSequence() uint64
	SetSequence(uint64) error
	String() string
}

type AccountUnrestricter interface {
	IsUnrestricted() bool
}

type AccountSession interface {
	GetKey() AccountKey
	String() string
}

//----------------------------------------
// BaseAccount

// BaseAccount - a base account structure.
// This can be extended by embedding within in your *Account structure.
type BaseAccount struct {
	Address       crypto.Address `json:"address" yaml:"address"`
	MasterKey     AccountKey     `json:"master_key" yaml:"master_key"`
	RootSequence  uint64         `json:"root_sequence" yaml:"root_sequence"` // sequence for master key
	Sessions      []AccountKey   `json:"sessions" yaml:"sessions"`           // sessions
	Coins         Coins          `json:"coins" yaml:"coins"`
	AccountNumber uint64         `json:"account_number" yaml:"account_number"`
	SequenceSum   uint64         `json:"sequence_sum" yaml:"sequence_sum"` // sum of all key sequences, total amount of calls made by this account by any key
}

// NewBaseAccount creates a new BaseAccount object
func NewBaseAccount(address crypto.Address, coins Coins, pubKey crypto.PubKey, accountNumber uint64,
) *BaseAccount {
	key := NewBaseAccountKey(pubKey, 0)
	return &BaseAccount{
		Address:       address,
		MasterKey:     key,
		RootSequence:  0,
		Coins:         coins,
		AccountNumber: accountNumber,
		Sessions:      []AccountKey{},
		SequenceSum:   0,
	}
}

// String implements fmt.Stringer
func (acc BaseAccount) String() string {
	return fmt.Sprintf(`Account:
  Address:       %s
  Coins:         %s
  AccountNumber: %d
  RootSequence:  %d
  SequenceSum:   %d
  Sessions:      %d`,
		acc.Address, acc.Coins, acc.AccountNumber, acc.RootSequence, acc.SequenceSum, len(acc.Sessions),
	)
}

// ProtoBaseAccount - a prototype function for BaseAccount
func ProtoBaseAccount() Account {
	return &BaseAccount{}
}

// GetAddress implements Account.
func (acc BaseAccount) GetAddress() crypto.Address {
	return acc.Address
}

// SetAddress implements Account.
func (acc *BaseAccount) SetAddress(addr crypto.Address) error {
	if !acc.Address.IsZero() {
		return errors.New("cannot override BaseAccount address")
	}
	acc.Address = addr
	return nil
}

// GetCoins implements Account.
func (acc *BaseAccount) GetCoins() Coins {
	return acc.Coins
}

// SetCoins implements Account.
func (acc *BaseAccount) SetCoins(coins Coins) error {
	acc.Coins = coins
	return nil
}

// GetAccountNumber implements Account.
func (acc *BaseAccount) GetAccountNumber() uint64 {
	return acc.AccountNumber
}

// SetAccountNumber implements Account.
func (acc *BaseAccount) SetAccountNumber(accNumber uint64) error {
	acc.AccountNumber = accNumber
	return nil
}

// AddSession implements Account.
func (acc *BaseAccount) AddSession(pubKey crypto.PubKey) (AccountKey, error) {
	// Check if the pubkey is the master key.
	if acc.MasterKey.GetPubKey().Equals(pubKey) {
		return nil, ErrAccountKeyAlreadyExists(acc.MasterKey.String())
	}

	// Check if a session with this pubKey already exists for this account.
	// Note: A public key can currently be used to manage multiple accounts
	// by signing with the appropriate account address and the appropriate
	// per-account sequence number.
	// This is intentional, as it allows a single key to control multiple
	// accounts while maintaining proper replay protection.
	for _, existingSess := range acc.Sessions {
		if existingSess.GetPubKey().Equals(pubKey) {
			return nil, ErrAccountKeyAlreadyExists(existingSess.String())
		}
	}

	// When adding a session, we initialize its sequence number to the account's current
	// sequence sum. This prevents replay attacks from previously pruned sessions,
	// since any old transactions would have sequence numbers lower than the current sequence
	// sum and thus be rejected. Multiple active sessions may use the same sequence
	// number concurrently, which is cryptographically safe since each signature is still
	// unique.
	sequenceNumber := acc.SequenceSum

	// Create and store the session key.
	sess := NewBaseAccountKey(pubKey, sequenceNumber)
	acc.Sessions = append(acc.Sessions, sess)

	return sess, nil
}

// GetAllKeys - Implements Account.
func (acc *BaseAccount) GetAllKeys() []AccountKey {
	return append([]AccountKey{acc.MasterKey}, acc.Sessions...)
}

// GetMasterKey implements Account.
func (acc *BaseAccount) GetMasterKey() AccountKey {
	return acc.MasterKey
}

// SetSequenceSum implements Account.
func (acc *BaseAccount) SetSequenceSum(sequenceSum uint64) error {
	acc.SequenceSum = sequenceSum
	return nil
}

// GetSequenceSum implements Account.
func (acc *BaseAccount) GetSequenceSum() uint64 {
	return acc.SequenceSum
}

// DelSession implements Account.
func (acc *BaseAccount) DelSession(pubKey crypto.PubKey) error {
	for i, sess := range acc.Sessions {
		if sess.GetPubKey().Equals(pubKey) {
			// Remove key at index i
			acc.Sessions = append(acc.Sessions[:i], acc.Sessions[i+1:]...)
			return nil
		}
	}
	return errors.New("session not found")
}

// DelAllSessions implements Account.
func (acc *BaseAccount) DelAllSessions() error {
	acc.Sessions = []AccountKey{}
	return nil
}

// GetKey implements Account.
func (acc BaseAccount) GetKey(pubKey crypto.PubKey) (AccountKey, error) {
	if acc.MasterKey.GetPubKey().Equals(pubKey) {
		return acc.MasterKey, nil
	}
	for _, sess := range acc.Sessions {
		if sess.GetPubKey().Equals(pubKey) {
			return sess, nil
		}
	}
	return nil, errors.New("key not found")
}

func (acc *BaseAccount) SetMasterKey(pubKey crypto.PubKey) (AccountKey, error) {
	acc.MasterKey = NewBaseAccountKey(pubKey, 0)
	return acc.MasterKey, nil
}

// SequenceByPubKey returns the sequence number for a given public key.
// If the public key is the master key, it returns the root sequence.
// If the public key is a session key, it returns the session sequence.
// If the public key is not found, it returns an error.
func (acc BaseAccount) SequenceByPubKey(pubKey crypto.PubKey) (uint64, error) {
	if acc.MasterKey.GetPubKey().Equals(pubKey) {
		return acc.RootSequence, nil
	}
	for _, sess := range acc.Sessions {
		if sess.GetPubKey().Equals(pubKey) {
			return sess.GetSequence(), nil
		}
	}
	return 0, errors.New("key not found")
}

// BaseAccountKey - a base structure for authentication.
type BaseAccountKey struct {
	PubKey   crypto.PubKey `json:"public_key" yaml:"public_key"`
	Sequence uint64        `json:"sequence" yaml:"sequence"`
}

// ProtoBaseAccountKey - a prototype function for BaseAccountKey
func ProtoBaseAccountKey() AccountKey {
	return &BaseAccountKey{}
}

func NewBaseAccountKey(pubKey crypto.PubKey, sequence uint64) *BaseAccountKey {
	return &BaseAccountKey{
		PubKey:   pubKey,
		Sequence: sequence,
	}
}

// String implements AccountKey and fmt.Stringer	.
func (k BaseAccountKey) String() string {
	var pubkey string
	if k.PubKey != nil {
		pubkey = crypto.PubKeyToBech32(k.PubKey)
	}
	return fmt.Sprintf(`AccountKey:
  Pubkey:    %s
  Sequence:  %d`,
		pubkey, k.Sequence,
	)
}

// GetPubKey implements AccountKey.
func (k BaseAccountKey) GetPubKey() crypto.PubKey {
	return k.PubKey
}

// SetPubKey implements AccountKey.
func (k *BaseAccountKey) SetPubKey(pubKey crypto.PubKey) error {
	k.PubKey = pubKey
	return nil
}

// GetSequence implements AccountKey.
func (k BaseAccountKey) GetSequence() uint64 {
	return k.Sequence
}

// SetSequence implements AccountKey.
func (k *BaseAccountKey) SetSequence(seq uint64) error {
	k.Sequence = seq
	return nil
}
