package gnoland

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
)

var (
	ErrBalanceEmptyAddress = errors.New("balance address is empty")
	ErrBalanceEmptyAmount  = errors.New("balance amount is empty")
)

const (
	// MaxSessionsPerAccount is the maximum number of sessions allowed per account
	MaxSessionsPerAccount = 64
)

const (
	// XXX rename these to flagXyz.

	// flagUnrestricted allows flagUnrestricted transfers.
	flagUnrestrictedAccount BitSet = 1 << iota

	// TODO: flagValidatorAccount marks an account as validator.
	flagValidatorAccount

	// TODO: flagRealmAccount marks an account as realm.
	flagRealmAccount
)

// validAccountFlags defines the set of all valid flags for accounts
var validAccountFlags = flagUnrestrictedAccount | flagValidatorAccount | flagRealmAccount

// Session flags - using the same BitSet type
const (
	flagSessionManagerSession BitSet = 1 << iota // Replaces CanManageOtherSessions
	flagPackageManagerSession                    // Replaces CanManagePackages
)

// validSessionFlags defines the set of all valid flags for sessions
var validSessionFlags = flagSessionManagerSession | flagPackageManagerSession

// bitSet represents a set of flags stored in a 64-bit unsigned integer.
// Each bit in the BitSet corresponds to a specific flag.
type BitSet uint64

func (bs BitSet) String() string {
	return fmt.Sprintf("0x%016X", uint64(bs)) // Show all 64 bits
}

var _ std.AccountUnrestricter = &GnoAccount{}

// Modify GnoAccount to work with GnoSessions
type GnoAccount struct {
	std.BaseAccount
	Attributes BitSet       `json:"attributes" yaml:"attributes"`
	Sessions   []GnoSession `json:"sessions" yaml:"sessions"`
}

// gc (garbage collect) removes expired sessions from the account.
// It returns the number of sessions that were removed.
func (ga *GnoAccount) gc() int {
	if len(ga.Sessions) == 0 {
		return 0
	}

	now := time.Now()
	initialCount := len(ga.Sessions)
	validSessions := make([]GnoSession, 0, initialCount)

	// Keep only non-expired sessions
	for _, session := range ga.Sessions {
		if session.ExpirationTime.IsZero() || now.Before(session.ExpirationTime) {
			validSessions = append(validSessions, session)
		}
	}

	// If we found expired sessions, update the slice
	if len(validSessions) < initialCount {
		ga.Sessions = validSessions
		return initialCount - len(validSessions)
	}

	return 0 // No sessions were removed
}

// CreateSession implements Session interface with GnoSession specifics
func (ga *GnoAccount) CreateSession(pubKey crypto.PubKey) (std.Session, error) {
	// Clean up expired sessions before adding new ones
	ga.gc()

	// Check if we're at the maximum number of sessions
	if len(ga.Sessions) >= MaxSessionsPerAccount {
		return nil, fmt.Errorf("maximum number of sessions reached (%d)", MaxSessionsPerAccount)
	}

	// Check if a session with this pubKey already exists
	for _, session := range ga.Sessions {
		if session.GetPubKey().Equals(pubKey) {
			return nil, errors.New("session with this public key already exists")
		}
	}

	// Create a new session with default settings
	session := NewGnoSession(ga.Address, pubKey)

	// Add to sessions collection
	ga.Sessions = append(ga.Sessions, *session)
	return session, nil
}

// GetSessionPubkeys returns all non-expired session pubkeys
func (ga *GnoAccount) GetSessionPubkeys() []crypto.PubKey {
	var pubkeys []crypto.PubKey
	now := time.Now()

	for _, session := range ga.Sessions {
		if session.ExpirationTime.IsZero() || now.Before(session.ExpirationTime) {
			pubkeys = append(pubkeys, session.GetPubKey())
		}
	}
	return pubkeys
}

// GetSession gets a specific session by pubkey
func (ga *GnoAccount) GetSession(pubKey crypto.PubKey) (*GnoSession, error) {
	for i := range ga.Sessions {
		if ga.Sessions[i].GetPubKey().Equals(pubKey) {
			// Check if session is expired
			if ga.Sessions[i].IsExpired() {
				return nil, errors.New("session has expired")
			}
			return &ga.Sessions[i], nil
		}
	}
	return nil, errors.New("session not found")
}

// RevokeSession implements Account interface with expiration check
func (ga *GnoAccount) RevokeSession(pubKey crypto.PubKey) error {
	for i, session := range ga.Sessions {
		if session.GetPubKey().Equals(pubKey) {
			// Remove session at index i
			ga.Sessions = append(ga.Sessions[:i], ga.Sessions[i+1:]...)
			return nil
		}
	}
	return errors.New("session not found")
}

// RevokeOtherSessions implements Account interface with permission check
func (ga *GnoAccount) RevokeOtherSessions(currentPubKey crypto.PubKey) error {
	var currentSession *GnoSession
	newSessions := make([]GnoSession, 0, 1)

	// Find current session
	for _, session := range ga.Sessions {
		if session.GetPubKey().Equals(currentPubKey) {
			currentSession = &session
			newSessions = append(newSessions, session)
			break
		}
	}

	if currentSession == nil {
		return errors.New("current session not found")
	}

	// Check if session has permission to manage other sessions
	if !currentSession.IsSessionManager() {
		return errors.New("current session does not have permission to manage other sessions")
	}

	ga.Sessions = newSessions
	return nil
}

func (ga *GnoAccount) setFlag(flag BitSet) {
	if !isValidAccountFlag(flag) {
		panic(fmt.Sprintf("setFlag: invalid account flag %d (binary: %b). Valid flags: %b",
			flag, flag, validAccountFlags))
	}
	ga.Attributes |= flag
}

func (ga *GnoAccount) clearFlag(flag BitSet) {
	if !isValidAccountFlag(flag) {
		panic(fmt.Sprintf("clearFlag: invalid account flag %d (binary: %b). Valid flags: %b",
			flag, flag, validAccountFlags))
	}
	ga.Attributes &= ^flag
}

func (ga *GnoAccount) hasFlag(flag BitSet) bool {
	if !isValidAccountFlag(flag) {
		panic(fmt.Sprintf("hasFlag: invalid account flag %d (binary: %b). Valid flags: %b",
			flag, flag, validAccountFlags))
	}
	return ga.Attributes&flag != 0
}

// isValidAccountFlag ensures valid account flags
func isValidAccountFlag(flag BitSet) bool {
	return flag&^validAccountFlags == 0 && flag != 0
}

// SetUnrestricted allows the account to bypass global transfer locking restrictions.
// By default, accounts are restricted when global transfer locking is enabled.
func (ga *GnoAccount) SetUnrestricted() {
	ga.setFlag(flagUnrestrictedAccount)
}

// IsUnrestricted checks whether the account is flagUnrestricted.
func (ga *GnoAccount) IsUnrestricted() bool {
	return ga.hasFlag(flagUnrestrictedAccount)
}

// String implements fmt.Stringer
func (ga *GnoAccount) String() string {
	return fmt.Sprintf("%s\n  Attributes:	 %s",
		ga.BaseAccount.String(),
		ga.Attributes.String(),
	)
}

func ProtoGnoAccount() std.Account {
	return &GnoAccount{}
}

// GnoSession extends BaseSession with ACL capabilities
//
// Currently, a session is linked to a specific public key, which means authentication
// and authorization are tied to a particular cryptographic identity. This design allows
// for straightforward signature-based verification.
//
// In future iterations, the session concept could be extended beyond just public keys.
// For example, sessions might be:
// - Linked to smart contracts that implement custom authorization logic
// - Associated with multi-sig requirements
// - Connected to external identity providers
// - Managed by governance protocols
//
// This would allow for more sophisticated authentication and delegation mechanisms
// while keeping the core session abstraction intact.
type GnoSession struct {
	std.BaseSession
	// Access Control Lists using BitSet
	Flags                 BitSet    `json:"flags" yaml:"flags"`
	ExpirationTime        time.Time `json:"expiration_time" yaml:"expiration_time"`
	CoinsTransferCapacity std.Coins `json:"coins_transfer_capacity" yaml:"coins_transfer_capacity"`
	RealmsWhitelist       []string  `json:"realms_whitelist" yaml:"realms_whitelist"`
}

// NewGnoSession creates a new GnoSession with default ACL settings
func NewGnoSession(
	accountAddress crypto.Address,
	pubKey crypto.PubKey,
) *GnoSession {
	return &GnoSession{
		BaseSession:           *std.NewBaseSession(accountAddress, pubKey, 0), // Default sequence is 0
		Flags:                 BitSet(0),                                      // No flags set by default
		ExpirationTime:        time.Time{},                                    // Zero time means no expiration
		CoinsTransferCapacity: std.Coins{},
		RealmsWhitelist:       []string{},
	}
}

// Add setters for all properties
func (s *GnoSession) SetSequence(sequence uint64) {
	s.BaseSession.SetSequence(sequence)
}

func (s *GnoSession) SetExpirationTime(expirationTime time.Time) {
	s.ExpirationTime = expirationTime
}

func (s *GnoSession) SetCoinsTransferCapacity(capacity std.Coins) {
	s.CoinsTransferCapacity = capacity
}

func (s *GnoSession) SetRealmsWhitelist(whitelist []string) {
	s.RealmsWhitelist = whitelist
}

// IsExpired checks if the session has expired
func (s *GnoSession) IsExpired() bool {
	return !s.ExpirationTime.IsZero() && time.Now().After(s.ExpirationTime)
}

// HasRealmAccess checks if the session has access to a specific realm
// Uses filepath.Match pattern syntax which supports wildcards:
// - "*" matches any sequence of non-separator characters
// - "?" matches any single non-separator character
// - character ranges like "[a-z]" match one character from the range
// - "\" can be used to escape special characters
//
// This provides flexible access control policies like:
// - "r/my*" (all realms starting with "r/my")
// - "r/*/public" (all "public" subrealms under any realm)
// - "r/app/v[1-3]/*" (all subrealms of app versions 1-3)
func (s *GnoSession) HasRealmAccess(realm string) bool {
	// If whitelist is empty, access is allowed to all realms
	if len(s.RealmsWhitelist) == 0 {
		return true
	}

	// Check each entry in the whitelist
	for _, pattern := range s.RealmsWhitelist {
		// Use filepath.Match which implements shell-like pattern matching
		// that's more powerful than simple prefix/suffix matching
		matched, err := filepath.Match(pattern, realm)

		// If there's a pattern error, we skip this pattern
		if err != nil {
			continue
		}

		if matched {
			return true
		}
	}

	return false
}

// CanTransferAmount checks if the session has sufficient transfer capacity
func (s *GnoSession) CanTransferAmount(amount std.Coins) bool {
	if s.CoinsTransferCapacity.IsZero() {
		return true // Zero capacity means unlimited transfers
	}
	return s.CoinsTransferCapacity.IsGTE(amount)
}

// String implements fmt.Stringer
func (s *GnoSession) String() string {
	return fmt.Sprintf(`%s
  Flags:               %s
  ExpirationTime:      %s
  CoinsTransferCapacity: %s
  RealmsWhitelist:     %v`,
		s.BaseSession.String(),
		s.Flags.String(),
		s.ExpirationTime,
		s.CoinsTransferCapacity,
		s.RealmsWhitelist,
	)
}

// Helper functions for session flags
func (s *GnoSession) setFlag(flag BitSet) {
	if !isValidSessionFlag(flag) {
		panic(fmt.Sprintf("setFlag: invalid session flag %d (binary: %b). Valid flags: %b",
			flag, flag, validSessionFlags))
	}
	s.Flags |= flag
}

func (s *GnoSession) clearFlag(flag BitSet) {
	if !isValidSessionFlag(flag) {
		panic(fmt.Sprintf("clearFlag: invalid session flag %d (binary: %b). Valid flags: %b",
			flag, flag, validSessionFlags))
	}
	s.Flags &= ^flag
}

func (s *GnoSession) hasFlag(flag BitSet) bool {
	if !isValidSessionFlag(flag) {
		panic(fmt.Sprintf("hasFlag: invalid session flag %d (binary: %b). Valid flags: %b",
			flag, flag, validSessionFlags))
	}
	return s.Flags&flag != 0
}

// isValidSessionFlag ensures valid session flags
func isValidSessionFlag(flag BitSet) bool {
	return flag&^validSessionFlags == 0 && flag != 0
}

func (s *GnoSession) SetSessionManager() {
	s.setFlag(flagSessionManagerSession)
}

func (s *GnoSession) IsSessionManager() bool {
	return s.hasFlag(flagSessionManagerSession)
}

func (s *GnoSession) SetPackageManager() {
	s.setFlag(flagPackageManagerSession)
}

func (s *GnoSession) IsPackageManager() bool {
	return s.hasFlag(flagPackageManagerSession)
}

func ProtoGnoSession() std.Session {
	return &GnoSession{}
}

type GnoGenesisState struct {
	Balances []Balance         `json:"balances"`
	Txs      []TxWithMetadata  `json:"txs"`
	Auth     auth.GenesisState `json:"auth"`
	Bank     bank.GenesisState `json:"bank"`
	VM       vm.GenesisState   `json:"vm"`
}

type TxWithMetadata struct {
	Tx       std.Tx         `json:"tx"`
	Metadata *GnoTxMetadata `json:"metadata,omitempty"`
}

type GnoTxMetadata struct {
	Timestamp int64 `json:"timestamp"`
}

// ReadGenesisTxs reads the genesis txs from the given file path
func ReadGenesisTxs(ctx context.Context, path string) ([]TxWithMetadata, error) {
	// Open the txs file
	file, loadErr := os.Open(path)
	if loadErr != nil {
		return nil, fmt.Errorf("unable to open tx file %s: %w", path, loadErr)
	}
	defer file.Close()

	var (
		txs []TxWithMetadata

		scanner = bufio.NewScanner(file)
	)

	scanner.Buffer(make([]byte, 1_000_000), 2_000_000)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			// Parse the amino JSON
			var tx TxWithMetadata
			if err := amino.UnmarshalJSON(scanner.Bytes(), &tx); err != nil {
				return nil, fmt.Errorf(
					"unable to unmarshal amino JSON, %w",
					err,
				)
			}

			txs = append(txs, tx)
		}
	}

	// Check for scanning errors
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf(
			"error encountered while reading file, %w",
			err,
		)
	}

	return txs, nil
}

// SignGenesisTxs will sign all txs passed as argument using the private key.
// This signature is only valid for genesis transactions as the account number and sequence are 0
func SignGenesisTxs(txs []TxWithMetadata, privKey crypto.PrivKey, chainID string) error {
	for index, tx := range txs {
		// Upon verifying genesis transactions, the account number and sequence are considered to be 0.
		// The reason for this is that it is not possible to know the account number (or sequence!) in advance
		// when generating the genesis transaction signature
		bytes, err := tx.Tx.GetSignBytes(chainID, 0, 0)
		if err != nil {
			return fmt.Errorf("unable to get sign bytes for transaction, %w", err)
		}

		signature, err := privKey.Sign(bytes)
		if err != nil {
			return fmt.Errorf("unable to sign genesis transaction, %w", err)
		}

		txs[index].Tx.Signatures = []std.Signature{
			{
				PubKey:    privKey.PubKey(),
				Signature: signature,
			},
		}
	}

	return nil
}
