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
	CreateSession(pubKey crypto.PubKey) (Session, error)
	RevokeSession(pubKey crypto.PubKey) error
	RevokeOtherSessions(currentPubKey crypto.PubKey) error
	GetSessions() []Session

	String() string
}

// Session represents authentication and replay protection details for an Account.
// Multiple sessions can be associated with a single Account.
//
// It presumes a notion of sequence numbers for replay protection, and a pubkey
// for authentication purposes.
//
// Many complex conditions can be used in the concrete struct which implements Session.
type Session interface {
	GetAccountAddress() crypto.Address // Reference to parent account
	SetAccountAddress(crypto.Address) error

	GetPubKey() crypto.PubKey
	SetPubKey(crypto.PubKey) error

	GetSequence() uint64
	SetSequence(uint64) error

	// Account reference
	GetAccount() Account

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
	Address       crypto.Address `json:"address" yaml:"address"`
	Coins         Coins          `json:"coins" yaml:"coins"`
	AccountNumber uint64         `json:"account_number" yaml:"account_number"`
	Sessions      []Session      `json:"sessions" yaml:"sessions"`
}

// NewBaseAccount creates a new BaseAccount object
func NewBaseAccount(address crypto.Address, coins Coins, accountNumber uint64) *BaseAccount {
	return &BaseAccount{
		Address:       address,
		Coins:         coins,
		AccountNumber: accountNumber,
	}
}

// String implements fmt.Stringer
func (acc BaseAccount) String() string {
	return fmt.Sprintf(`Account:
  Address:       %s
  Coins:         %s
  AccountNumber: %d`,
		acc.Address, acc.Coins, acc.AccountNumber,
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

func (acc *BaseAccount) CreateSession(pubKey crypto.PubKey) (Session, error) {
	// Check if a session with this pubKey already exists
	for _, session := range acc.Sessions {
		if session.GetPubKey().Equals(pubKey) {
			return nil, errors.New("session with this public key already exists")
		}
	}

	// Create new session
	session := NewBaseSession(acc.Address, pubKey, 0)
	session.setAccount(acc)
	acc.Sessions = append(acc.Sessions, session)
	return session, nil
}

func (acc *BaseAccount) RevokeSession(pubKey crypto.PubKey) error {
	for i, session := range acc.Sessions {
		if session.GetPubKey().Equals(pubKey) {
			// Remove session at index i
			acc.Sessions = append(acc.Sessions[:i], acc.Sessions[i+1:]...)
			return nil
		}
	}
	return errors.New("session not found")
}

func (acc *BaseAccount) RevokeOtherSessions(currentPubKey crypto.PubKey) error {
	var currentSession Session
	newSessions := make([]Session, 0, 1)

	// Find and keep only the current session
	for _, session := range acc.Sessions {
		if session.GetPubKey().Equals(currentPubKey) {
			currentSession = session
			newSessions = append(newSessions, session)
		}
	}

	if currentSession == nil {
		return errors.New("current session not found")
	}

	acc.Sessions = newSessions
	return nil
}

func (acc *BaseAccount) GetSessions() []Session {
	return acc.Sessions
}

// BaseSession - a base session structure for authentication.
// This can be extended by embedding within in your *Session structure.
type BaseSession struct {
	AccountAddress crypto.Address `json:"account_address" yaml:"account_address"`
	PubKey         crypto.PubKey  `json:"public_key" yaml:"public_key"`
	Sequence       uint64         `json:"sequence" yaml:"sequence"`
	account        Account        `json:"-" yaml:"-"` // Reference to the parent account
}

// Add GetAccount method to BaseSession
func (s *BaseSession) GetAccount() Account {
	return s.account
}

// Modify NewBaseSession to accept Account parameter
func NewBaseSession(accountAddress crypto.Address, pubKey crypto.PubKey, sequence uint64) *BaseSession {
	return &BaseSession{
		AccountAddress: accountAddress,
		PubKey:         pubKey,
		Sequence:       sequence,
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
