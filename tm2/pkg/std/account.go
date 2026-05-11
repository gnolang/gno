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

//----------------------------------------
// BaseSessionAccount

// BaseSessionAccount is a delegated signing account linked to a master.
// It is keyed under the master account in the store.
//
// Session accounts do not hold coins — fees are always deducted from the
// master account. GetCoins always returns nil and SetCoins rejects non-empty
// coins to prevent accidental trapping of funds.
//
// SpendLimit controls how many coins the session can transfer per period
// (via MsgCall.Send, MsgSend, gas fees, etc.). If SpendLimit is empty,
// the session cannot spend any coins at all — this is useful for sessions
// where another signer pays gas, or for calling functions that don't
// require coin transfers. SpendLimit must include the gas fee denom
// (e.g., ugnot) or the session won't be able to pay gas fees — spending
// is checked per-denom, and a missing denom means zero allowance.
//
// Zero-value semantics (important — each field has a special meaning when zero):
//   - ExpiresAt == 0:  no expiry; session is valid until revoked.
//   - SpendPeriod == 0: SpendLimit is a lifetime cap (no periodic reset).
//   - SpendLimit nil/empty: no spending allowed (including gas); useful only
//     when another signer pays gas. NOT "unrestricted" — fails closed.
//   - SpendUsed nil/empty: zero spent (Coins treats nil the same as empty).
//   - SpendReset == 0: initial state (no period has elapsed yet). Set to
//     BlockTime at session creation by the handler.
type BaseSessionAccount struct {
	BaseAccount
	MasterAddress crypto.Address `json:"master_address" yaml:"master_address"`
	ExpiresAt     int64          `json:"expires_at" yaml:"expires_at"`                         // unix ts; 0 = no expiry
	SpendLimit    Coins          `json:"spend_limit,omitempty" yaml:"spend_limit,omitempty"`   // nil/empty = no spending allowed (fail-closed, NOT unrestricted)
	SpendPeriod   int64          `json:"spend_period,omitempty" yaml:"spend_period,omitempty"` // seconds; 0 = lifetime cap (no reset)
	SpendUsed     Coins          `json:"spend_used,omitempty" yaml:"spend_used,omitempty"`     // nil/empty = 0 spent
	SpendReset    int64          `json:"spend_reset,omitempty" yaml:"spend_reset,omitempty"`   // unix ts; start of current period
}

// GetCoins always returns nil — session accounts do not hold coins.
func (acc BaseSessionAccount) GetCoins() Coins { return nil }

// SetCoins rejects non-empty coins. Session accounts should never hold coins;
// fees are deducted from the master account.
func (acc *BaseSessionAccount) SetCoins(coins Coins) error {
	if !coins.IsZero() {
		return fmt.Errorf("session accounts cannot hold coins")
	}
	return nil
}

// DelegatedAccount is implemented by session accounts that delegate
// fee payment and identity to a master account.
type DelegatedAccount interface {
	Account
	GetMasterAddress() crypto.Address
	SetMasterAddress(crypto.Address) error
	GetExpiresAt() int64
	SetExpiresAt(int64) error
	GetSpendLimit() Coins
	SetSpendLimit(Coins) error
	GetSpendPeriod() int64
	SetSpendPeriod(int64) error
	GetSpendUsed() Coins
	SetSpendUsed(Coins) error
	GetSpendReset() int64
	SetSpendReset(int64) error
}

// SessionAccountsContextKey is the context key used by the auth ante handler
// to propagate the set of session accounts it resolved during Phase 1 into
// downstream handler execution (bank keeper hooks, VM keeper, gno runtime).
//
// The value stored under this key has the exact type:
//
//	map[crypto.Address]DelegatedAccount
//
// Each entry maps a signer address (the master account address returned by
// msg.GetSigners(), NOT the session pubkey address) to the DelegatedAccount
// that was loaded for it via the Signature.SessionAddr field
// (when non-zero — zero SessionAddr means master-signed).
//
// Contract for readers:
//   - A nil value or absent key means "not a session tx" — treat as master.
//   - A zero-length map means "no session signers in this tx" — equivalent
//     to nil for all practical purposes.
//   - Presence of (signerAddr, da) in the map means the ante has already
//     verified the session exists, is unexpired, and the tx signer is
//     authorized as that session; downstream code can rely on these
//     invariants without re-checking.
//   - The DelegatedAccount values are SHARED POINTERS — mutations to
//     da.SpendUsed propagate across all readers within the same tx.
//     Callers of auth.DeductSessionSpend / CheckAndDeductSessionSpend
//     rely on this for cumulative spend tracking.
//
// Contract for writers: only the auth ante should populate this key. It is
// set in Phase 4 of the ante (after signature verification succeeds) and
// read by any keeper hook that needs to attribute coin movement or other
// authority to a session rather than a master account.
type SessionAccountsContextKey struct{}

const (
	MaxSessionsPerAccount   = 16
	MaxAllowPathsPerSession = 8
	MaxSessionDuration      = 30 * 24 * 60 * 60 // 30 days in seconds
)

// ProtoBaseSessionAccount - a prototype function for BaseSessionAccount
func ProtoBaseSessionAccount() Account {
	return &BaseSessionAccount{}
}

// String implements fmt.Stringer
func (acc BaseSessionAccount) String() string {
	var pubkey string

	if acc.PubKey != nil {
		pubkey = crypto.PubKeyToBech32(acc.PubKey)
	}

	return fmt.Sprintf(`SessionAccount:
  Address:       %s
  Pubkey:        %s
  Coins:         %s
  AccountNumber: %d
  Sequence:      %d
  MasterAddress: %s
  ExpiresAt:     %d
  SpendLimit:    %s
  SpendPeriod:   %d
  SpendUsed:     %s
  SpendReset:    %d`,
		acc.Address, pubkey, acc.GetCoins(), acc.AccountNumber, acc.Sequence,
		acc.MasterAddress, acc.ExpiresAt, acc.SpendLimit, acc.SpendPeriod,
		acc.SpendUsed, acc.SpendReset,
	)
}

// GetMasterAddress - Implements DelegatedAccount.
func (acc BaseSessionAccount) GetMasterAddress() crypto.Address {
	return acc.MasterAddress
}

// SetMasterAddress - Implements DelegatedAccount.
func (acc *BaseSessionAccount) SetMasterAddress(addr crypto.Address) error {
	if !acc.MasterAddress.IsZero() {
		return errors.New("cannot override BaseSessionAccount master address")
	}
	acc.MasterAddress = addr
	return nil
}

// GetExpiresAt - Implements DelegatedAccount.
func (acc BaseSessionAccount) GetExpiresAt() int64 {
	return acc.ExpiresAt
}

// SetExpiresAt - Implements DelegatedAccount.
func (acc *BaseSessionAccount) SetExpiresAt(t int64) error {
	acc.ExpiresAt = t
	return nil
}

// GetSpendLimit - Implements DelegatedAccount.
func (acc BaseSessionAccount) GetSpendLimit() Coins {
	return acc.SpendLimit
}

// SetSpendLimit - Implements DelegatedAccount.
func (acc *BaseSessionAccount) SetSpendLimit(coins Coins) error {
	acc.SpendLimit = coins
	return nil
}

// GetSpendPeriod - Implements DelegatedAccount.
func (acc BaseSessionAccount) GetSpendPeriod() int64 {
	return acc.SpendPeriod
}

// SetSpendPeriod - Implements DelegatedAccount.
func (acc *BaseSessionAccount) SetSpendPeriod(period int64) error {
	acc.SpendPeriod = period
	return nil
}

// GetSpendUsed - Implements DelegatedAccount.
func (acc BaseSessionAccount) GetSpendUsed() Coins {
	return acc.SpendUsed
}

// SetSpendUsed - Implements DelegatedAccount.
func (acc *BaseSessionAccount) SetSpendUsed(coins Coins) error {
	acc.SpendUsed = coins
	return nil
}

// GetSpendReset - Implements DelegatedAccount.
func (acc BaseSessionAccount) GetSpendReset() int64 {
	return acc.SpendReset
}

// SetSpendReset - Implements DelegatedAccount.
func (acc *BaseSessionAccount) SetSpendReset(t int64) error {
	acc.SpendReset = t
	return nil
}
