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

	GetGlobalSequence() uint64
	SetGlobalSequence(uint64) error

	GetCoins() Coins
	SetCoins(Coins) error

	// Session management
	GetSession(pubKey crypto.PubKey) (Session, error)
	AddSession(pubKey crypto.PubKey) (Session, error)
	DelSession(pubKey crypto.PubKey) error
	GetAllSessions() []Session

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
	// Account reference
	GetAccount() Account
	GetAccountAddress() crypto.Address

	GetPubKey() crypto.PubKey
	SetPubKey(crypto.PubKey) error

	GetSequence() uint64
	SetSequence(uint64) error

	// Master session status
	IsMaster() bool

	// XXX: IsValid checks if the session is valid (not expired, etc)
	//      sdk.Context is not available (import cycle)
	// IsValid(ctx sdk.Context) bool

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
	Address        crypto.Address `json:"address" yaml:"address"`
	Sessions       []Session      `json:"sessions" yaml:"sessions"` // First session is master
	Coins          Coins          `json:"coins" yaml:"coins"`
	AccountNumber  uint64         `json:"account_number" yaml:"account_number"`
	GlobalSequence uint64         `json:"global_sequence" yaml:"global_sequence"` // sum of all session sequences
}

// NewBaseAccount creates a new BaseAccount object
func NewBaseAccount(address crypto.Address, coins Coins, pubKey crypto.PubKey, accountNumber uint64,
) *BaseAccount {
	return &BaseAccount{
		Address:        address,
		Coins:          coins,
		AccountNumber:  accountNumber,
		Sessions:       []Session{NewMasterSession(pubKey)},
		GlobalSequence: 0,
	}
}

// String implements fmt.Stringer
func (acc BaseAccount) String() string {
	return fmt.Sprintf(`Account:
  Address:        %s
  Coins:          %s
  AccountNumber:  %d
  GlobalSequence: %d
  Sessions:       %d`,
		acc.Address, acc.Coins, acc.AccountNumber, acc.GlobalSequence, len(acc.Sessions),
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

// AddSession - Implements Account.
func (acc *BaseAccount) AddSession(sess Session) error {
	newPubKey := sess.GetPubKey()

	// Check if a session with this pubKey already exists for this account.
	// Note: A public key can currently be used to manage multiple accounts
	// by signing with the appropriate account address and the appropriate
	// per-account sequence number.
	// This is intentional, as it allows a single key to control multiple
	// accounts while maintaining proper replay protection.
	for _, existingSess := range acc.Sessions {
		if existingSess.GetPubKey().Equals(newPubKey) {
			// When re-adding a previously pruned session, we need to prevent replay attacks.
			// We do this by setting the session's sequence number to the account's current
			// global sequence number, rather than starting from 0. This ensures any
			// previously signed transactions cannot be replayed.
			sess.SetSequence(account.GetGlobalSequence())
			return ErrSessionAlreadyExists
		}
	}

	// Store the session key
	acc.Sessions = append(acc.Sessions, sess)
	return nil
}

func (acc *BaseAccount) SetGlobalSequence(globalSequence uint64) error {
	acc.GlobalSequence = globalSequence
	return nil
}

func (acc *BaseAccount) GetGlobalSequence() uint64 {
	return acc.GlobalSequence
}

// RevokeSession removes a session key
func (acc *BaseAccount) RevokeSession(pubKey crypto.PubKey) error {
	for i, sess := range acc.Sessions {
		if sess.GetPubKey().Equals(pubKey) {
			if sess.IsMaster() {
				return errors.New("cannot revoke master session")
			}
			// Remove key at index i
			acc.Sessions = append(acc.Sessions[:i], acc.Sessions[i+1:]...)
			return nil
		}
	}
	return errors.New("session not found")
}

// GetSessions creates Session objects for all stored keys
func (acc *BaseAccount) GetSessions() []Session {
	sessions := make([]Session, len(acc.Sessions))
	for i, sess := range acc.Sessions {
		sess.setAccount(acc)
		sessions[i] = sess
	}
	return sessions
}

// BaseSession - a base session structure for authentication.
// This can be extended by embedding within in your *Session structure.
type BaseSession struct {
	PubKey   crypto.PubKey `json:"public_key" yaml:"public_key"`
	Sequence uint64        `json:"sequence" yaml:"sequence"`
	Master   bool          `json:"master" yaml:"master"`
	account  Account       `json:"-" yaml:"-"` // Reference to the parent account
}

// Add GetAccount method to BaseSession
func (s *BaseSession) GetAccount() Account {
	return s.account
}

func NewMasterSession(pubKey crypto.PubKey) *BaseSession {
	return &BaseSession{
		PubKey:   pubKey,
		Sequence: 0,
		Master:   true,
	}
}

func NewBaseSession(pubKey crypto.PubKey, sequence uint64) *BaseSession {
	return &BaseSession{
		PubKey:   pubKey,
		Sequence: sequence,
		Master:   false,
	}
}

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
  Pubkey:         %s
  Sequence:       %d
  Master:         %t`,
		s.GetAccountAddress(), pubkey, s.Sequence, s.Master,
	)
}

func (s BaseSession) GetAccountAddress() crypto.Address {
	if s.account == nil {
		// should only happen when forget to call sess.setAccount
		return crypto.Address{}
	}
	return s.account.GetAddress()
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
func (s BaseSession) IsMaster() bool {
	return s.Master
}
