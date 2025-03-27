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

	GetCoins() Coins
	SetCoins(Coins) error

	// Session management
	CreateSession(pubKey crypto.PubKey, sequence uint64) (Session, error)
	RevokeSession(pubKey crypto.PubKey) error
	GetSessions() []Session

	String() string
}

// Session represents authentication and replay protection details for an Account.
// Multiple sessions can be associated with a single Account.
//
// Each account has exactly one master session, created at account initialization.
// The master session:
// - Cannot be revoked
// - Has all permissions and unlimited transfer capacity
// - Never expires
// Master sessions should be kept secure and used only for account recovery or
// critical operations. Regular sessions with limited permissions should be used
// for daily operations.
//
// Currently, a session is linked to a specific public key, which means authentication
// and authorization are tied to a particular cryptographic identity.
//
// In future iterations, the session concept could be extended beyond just public keys.
// For example, sessions might be:
// - Linked to smart contracts that implement custom authorization logic
// - Associated with multi-sig requirements
// - Connected to external identity providers
// - Managed by governance protocols
type Session interface {
	GetAccountAddress() crypto.Address // Reference to parent account
	SetAccountAddress(crypto.Address) error

	GetPubKey() crypto.PubKey
	SetPubKey(crypto.PubKey) error

	GetSequence() uint64
	SetSequence(uint64) error

	// Account reference
	GetAccount() Account

	// Master session status
	IsMaster() bool

	String() string
}

type AccountUnrestricter interface {
	IsUnrestricted() bool
}

//----------------------------------------
// BaseAccount

// BaseAccount - a base account structure.
// This can be extended by embedding within in your *Account structure.
type BaseAccount struct {
	Address       crypto.Address  `json:"address" yaml:"address"`
	Coins         Coins           `json:"coins" yaml:"coins"`
	AccountNumber uint64          `json:"account_number" yaml:"account_number"`
	SessionKeys   []crypto.PubKey `json:"session_keys" yaml:"session_keys"` // First key is master
}

// NewBaseAccount creates a new BaseAccount object
func NewBaseAccount(address crypto.Address, coins Coins, accountNumber uint64, pubkey crypto.PubKey) *BaseAccount {
	return &BaseAccount{
		Address:       address,
		Coins:         coins,
		AccountNumber: accountNumber,
		SessionKeys:   []crypto.PubKey{pubkey}, // First key is master
	}
}

// String implements fmt.Stringer
func (acc BaseAccount) String() string {
	return fmt.Sprintf(`Account:
  Address:       %s
  Coins:         %s
  AccountNumber: %d
  SessionKeys:   %v`,
		acc.Address, acc.Coins, acc.AccountNumber, acc.SessionKeys,
	)
}

// ProtoBaseAccount - a prototype function for BaseAccount
func ProtoBaseAccount() Account {
	return &BaseAccount{}
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

// GetCoins - Implements Account.
func (acc *BaseAccount) GetCoins() Coins {
	return acc.Coins
}

// SetCoins - Implements Account.
func (acc *BaseAccount) SetCoins(coins Coins) error {
	acc.Coins = coins
	return nil
}

// GetAccountNumber - Implements Account.
func (acc *BaseAccount) GetAccountNumber() uint64 {
	return acc.AccountNumber
}

// SetAccountNumber - Implements Account.
func (acc *BaseAccount) SetAccountNumber(accNumber uint64) error {
	acc.AccountNumber = accNumber
	return nil
}

// CreateSession creates a new session and stores its key
func (acc *BaseAccount) CreateSession(pubKey crypto.PubKey, sequence uint64) (Session, error) {
	// Check if a session with this pubKey already exists
	for _, existingKey := range acc.SessionKeys {
		if existingKey.Equals(pubKey) {
			return nil, errors.New("session with this public key already exists")
		}
	}

	// Store the session key
	acc.SessionKeys = append(acc.SessionKeys, pubKey)

	// Create and return the session object
	session := NewBaseSession(acc.Address, pubKey, sequence, len(acc.SessionKeys) == 1) // master if first key
	session.setAccount(acc)
	return session, nil
}

// RevokeSession removes a session key
func (acc *BaseAccount) RevokeSession(pubKey crypto.PubKey) error {
	for i, key := range acc.SessionKeys {
		if key.Equals(pubKey) {
			// Prevent revoking master session (first key)
			if i == 0 {
				return errors.New("cannot revoke master session")
			}
			// Remove key at index i
			acc.SessionKeys = append(acc.SessionKeys[:i], acc.SessionKeys[i+1:]...)
			return nil
		}
	}
	return errors.New("session not found")
}

// GetSessions creates Session objects for all stored keys
func (acc *BaseAccount) GetSessions() []Session {
	sessions := make([]Session, len(acc.SessionKeys))
	for i, key := range acc.SessionKeys {
		session := NewBaseSession(acc.Address, key, 0, i == 0) // master if first key
		session.setAccount(acc)
		sessions[i] = session
	}
	return sessions
}

// BaseSession - a base session structure for authentication.
// This can be extended by embedding within in your *Session structure.
type BaseSession struct {
	AccountAddress crypto.Address `json:"account_address" yaml:"account_address"`
	PubKey         crypto.PubKey  `json:"public_key" yaml:"public_key"`
	Sequence       uint64         `json:"sequence" yaml:"sequence"`
	account        Account        `json:"-" yaml:"-"`           // Reference to the parent account
	master         bool           `json:"master" yaml:"master"` // Whether this is a master session
}

// Add GetAccount method to BaseSession
func (s *BaseSession) GetAccount() Account {
	return s.account
}

func NewBaseSession(accountAddress crypto.Address, pubKey crypto.PubKey, sequence uint64, master bool) *BaseSession {
	return &BaseSession{
		AccountAddress: accountAddress,
		PubKey:         pubKey,
		Sequence:       sequence,
		master:         master,
	}
}

// Helper method to set the account reference (should be called when creating a session)
func (s *BaseSession) setAccount(acc Account) {
	s.account = acc
}

func (s BaseSession) String() string {
	var pubkey string
	if s.PubKey != nil {
		pubkey = crypto.PubKeyToBech32(s.PubKey)
	}
	return fmt.Sprintf(`Session:
  AccountAddress: %s
  Pubkey:        %s
  Sequence:      %d`,
		s.AccountAddress, pubkey, s.Sequence,
	)
}

func (s BaseSession) GetAccountAddress() crypto.Address {
	return s.AccountAddress
}

func (s *BaseSession) SetAccountAddress(addr crypto.Address) error {
	if !s.AccountAddress.IsZero() {
		return errors.New("cannot override BaseSession account address")
	}
	s.AccountAddress = addr
	return nil
}

func (s BaseSession) GetPubKey() crypto.PubKey {
	return s.PubKey
}

func (s *BaseSession) SetPubKey(pubKey crypto.PubKey) error {
	s.PubKey = pubKey
	return nil
}

func (s BaseSession) GetSequence() uint64 {
	return s.Sequence
}

func (s *BaseSession) SetSequence(seq uint64) error {
	s.Sequence = seq
	return nil
}

// Add IsMaster implementation
func (s *BaseSession) IsMaster() bool {
	return s.master
}
