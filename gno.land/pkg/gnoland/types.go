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
	ErrBalanceEmptyAddress  = errors.New("balance address is empty")
	ErrBalanceEmptyAmount   = errors.New("balance amount is empty")
	ErrSessionAlreadyExists = errors.New("session already exists")
	ErrSessionNotFound      = errors.New("session not found")
	ErrSessionInvalid       = errors.New("session is invalid")
	ErrSessionUnauthorized  = errors.New("session is unauthorized")
)

// Account flags
const (
	// flagUnrestrictedAccount allows unrestricted transfers.
	flagUnrestrictedAccount BitSet = 1 << iota

	// flagAccountRealm marks an account as realm.
	// XXX: flagAccountRealm

	// flagAccountTOSAccepted marks an account as having accepted the terms of service.
	// XXX: flagAccountTOSAccepted

	// flagAccountFreeze marks an account as frozen.
	// XXX: flagAccountFreeze
)

// validAccountFlags defines the set of all valid flags for accounts
var validAccountFlags = flagUnrestrictedAccount /* XXX: | flagAccountRealm | ... */

// Session flags
const (
	// flagSessionUnlimitedTransferCapacity allows unlimited coin transfers (ignores capacity field).
	flagSessionUnlimitedTransferCapacity BitSet = 1 << iota

	// flagSessionCanManageSessions is a flag that allows the session to manage other sessions.
	flagSessionCanManageSessions

	// flagSessionCanManagePackages is a flag that allows the session to manage packages.
	flagSessionCanManagePackages

	// flagSessionValidationOnly is a flag limiting the session to validator permissions only.
	flagSessionValidationOnly

	// flagSessionCanMakeIBCCalls is a flag that allows the session to make IBC calls.
	// XXX: flagSessionCanMakeIBCCalls
)

// validSessionFlags defines the set of all valid flags for sessions
var validSessionFlags = flagSessionUnlimitedTransferCapacity | flagSessionCanManageSessions | flagSessionCanManagePackages | flagSessionValidationOnly

const (
	// MaxSessionsPerAccount is the maximum number of sessions allowed per account
	MaxSessionsPerAccount = 64
)

// bitSet represents a set of flags stored in a 64-bit unsigned integer.
// Each bit in the BitSet corresponds to a specific flag.
type BitSet uint64

func (bs BitSet) String() string {
	return fmt.Sprintf("0x%016X", uint64(bs)) // Show all 64 bits
}

var _ std.AccountUnrestricter = &GnoAccount{}

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
		hasExpiration := !session.ExpirationTime.IsZero()
		isExpired := hasExpiration && now.After(session.ExpirationTime)
		
		if !isExpired {
			validSessions = append(validSessions, session)
		}
	}

	// Update sessions if any were removed
	removedCount := initialCount - len(validSessions)
	if removedCount > 0 {
		ga.Sessions = validSessions
	}

	return removedCount
}

// CreateSession implements Session interface with GnoSession specifics
func (ga *GnoAccount) CreateSession(pubKey crypto.PubKey) (*GnoSession, error) {
	// Clean up expired sessions before adding new ones
	ga.gc()

	// Check if we're at the maximum number of sessions
	currentSessionCount := len(ga.Sessions)
	if currentSessionCount >= MaxSessionsPerAccount {
		return nil, fmt.Errorf("maximum number of sessions reached (%d)", MaxSessionsPerAccount)
	}

	// Check if a session with this pubKey already exists
	for _, existingSession := range ga.Sessions {
		if existingSession.MatchesPubKey(pubKey) {
			return nil, errors.New("session with this public key already exists")
		}
	}

	// Create a new session
	accountAddr := ga.Address
	newSession := NewGnoSession(accountAddr, pubKey)

	// Add to sessions collection
	ga.Sessions = append(ga.Sessions, *newSession)
	return newSession, nil
}

// GetSessions returns all non-expired sessions
// Implements the Account interface
func (ga *GnoAccount) GetSessions() []GnoSession {
	// Clean up expired sessions first
	ga.gc()

	// Return copy of sessions
	sessions := make([]GnoSession, len(ga.Sessions))
	copy(sessions, ga.Sessions)
	return sessions
}

// GetSession gets a specific session by pubkey
func (ga *GnoAccount) GetSession(pubKey crypto.PubKey) (*GnoSession, error) {
	for i := range ga.Sessions {
		session := &ga.Sessions[i]
		
		if session.MatchesPubKey(pubKey) {
			// Check if session is expired
			if session.IsExpired() {
				return nil, errors.New("session has expired")
			}
			return session, nil
		}
	}
	return nil, errors.New("session not found")
}

// RevokeSession implements Account interface with expiration check
func (ga *GnoAccount) RevokeSession(pubKey crypto.PubKey) error {
	for i, session := range ga.Sessions {
		if session.MatchesPubKey(pubKey) {
			// Remove session at index i
			ga.Sessions = append(ga.Sessions[:i], ga.Sessions[i+1:]...)
			return nil
		}
	}
	return errors.New("session not found")
}

// RevokeOtherSessions implements Account interface with permission check
func (ga *GnoAccount) RevokeOtherSessions(currentPubKey crypto.PubKey) error {
	panic("not implemented")
}

func (ga *GnoAccount) setFlag(flag BitSet) {
	isValid := isValidAccountFlag(flag)
	if !isValid {
		validFlags := validAccountFlags
		panic(fmt.Sprintf("setFlag: invalid account flag %d (binary: %b). Valid flags: %b",
			flag, flag, validFlags))
	}
	ga.Attributes |= flag
}

func (ga *GnoAccount) clearFlag(flag BitSet) {
	isValid := isValidAccountFlag(flag)
	if !isValid {
		validFlags := validAccountFlags
		panic(fmt.Sprintf("clearFlag: invalid account flag %d (binary: %b). Valid flags: %b",
			flag, flag, validFlags))
	}
	ga.Attributes &= ^flag
}

func (ga *GnoAccount) hasFlag(flag BitSet) bool {
	isValid := isValidAccountFlag(flag)
	if !isValid {
		validFlags := validAccountFlags
		panic(fmt.Sprintf("hasFlag: invalid account flag %d (binary: %b). Valid flags: %b",
			flag, flag, validFlags))
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
	std.BaseAccountKey
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
	return NewGnoSessionWithSequence(accountAddress, pubKey, 0)
}

// NewGnoSessionWithSequence creates a new GnoSession with a specific initial sequence
func NewGnoSessionWithSequence(
	accountAddress crypto.Address,
	pubKey crypto.PubKey,
	initialSequence uint64,
) *GnoSession {
	baseAccountKey := std.NewBaseAccountKey(pubKey, initialSequence)
	
	session := &GnoSession{
		BaseAccountKey:        *baseAccountKey,
		Flags:                 BitSet(0),
		ExpirationTime:        time.Time{}, // Zero time means never expire
		CoinsTransferCapacity: std.Coins{}, // Zero capacity means no transfers (unless unlimited flag set)
		RealmsWhitelist:       []string{},  // Empty whitelist means access to all realms
	}

	return session
}

// GetPubKey returns the session's public key for better readability
func (s *GnoSession) GetPubKey() crypto.PubKey {
	return s.BaseAccountKey.GetPubKey()
}

// MatchesPubKey checks if the session matches the given public key
func (s *GnoSession) MatchesPubKey(pubKey crypto.PubKey) bool {
	sessionPubKey := s.GetPubKey()
	return sessionPubKey.Equals(pubKey)
}

// Add setters for all properties
func (s *GnoSession) SetSequence(sequence uint64) error {
	return s.BaseAccountKey.SetSequence(sequence)
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
	// Sessions with zero expiration time never expire
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
	whitelist := s.RealmsWhitelist
	
	// Empty whitelist means access to all realms
	if len(whitelist) == 0 {
		return true
	}

	// Check each pattern in whitelist
	for _, pattern := range whitelist {
		isMatch, err := filepath.Match(pattern, realm)
		if err != nil {
			continue // Skip malformed patterns
		}
		if isMatch {
			return true
		}
	}
	return false
}

// CanTransferAmount checks if the session has sufficient transfer capacity
func (s *GnoSession) CanTransferAmount(amount std.Coins) bool {
	// Check if session has unlimited transfer capacity
	if s.HasUnlimitedTransferCapacity() {
		return true
	}
	
	// Check against capacity limit
	capacity := s.CoinsTransferCapacity
	if capacity.IsZero() {
		// No capacity means no transfers allowed (unless unlimited flag is set)
		return false
	}
	
	return capacity.IsAllGTE(amount)
}

// ConsumeTransferCapacity decreases the session's transfer capacity by the given amount
// This should be called when a transfer is actually executed
func (s *GnoSession) ConsumeTransferCapacity(amount std.Coins) error {
	// Unlimited capacity sessions don't need to consume capacity
	if s.HasUnlimitedTransferCapacity() {
		return nil
	}
	
	capacity := s.CoinsTransferCapacity
	if capacity.IsZero() {
		return errors.New("no transfer capacity available")
	}
	
	// Check if we have enough capacity
	if !capacity.IsAllGTE(amount) {
		return errors.New("insufficient transfer capacity")
	}
	
	// Subtract the amount from capacity
	newCapacity := capacity.Sub(amount)
	s.CoinsTransferCapacity = newCapacity
	
	return nil
}

// String implements fmt.Stringer
func (s *GnoSession) String() string {
	baseInfo := s.BaseAccountKey.String()
	flags := s.Flags.String()
	expiration := s.ExpirationTime
	capacity := s.CoinsTransferCapacity
	whitelist := s.RealmsWhitelist
	
	return fmt.Sprintf(`%s
  Flags:               %s
  ExpirationTime:      %s
  CoinsTransferCapacity: %s
  RealmsWhitelist:     %v`,
		baseInfo, flags, expiration, capacity, whitelist)
}

// Helper functions for session flags
func (s *GnoSession) setFlag(flag BitSet) {
	isValid := isValidSessionFlag(flag)
	if !isValid {
		validFlags := validSessionFlags
		panic(fmt.Sprintf("setFlag: invalid session flag %d (binary: %b). Valid flags: %b",
			flag, flag, validFlags))
	}
	s.Flags |= flag
}

func (s *GnoSession) clearFlag(flag BitSet) {
	isValid := isValidSessionFlag(flag)
	if !isValid {
		validFlags := validSessionFlags
		panic(fmt.Sprintf("clearFlag: invalid session flag %d (binary: %b). Valid flags: %b",
			flag, flag, validFlags))
	}
	s.Flags &= ^flag
}

func (s *GnoSession) hasFlag(flag BitSet) bool {
	isValid := isValidSessionFlag(flag)
	if !isValid {
		validFlags := validSessionFlags
		panic(fmt.Sprintf("hasFlag: invalid session flag %d (binary: %b). Valid flags: %b",
			flag, flag, validFlags))
	}
	return s.Flags&flag != 0
}

// isValidSessionFlag ensures valid session flags
func isValidSessionFlag(flag BitSet) bool {
	return flag&^validSessionFlags == 0 && flag != 0
}

func (s *GnoSession) SetUnlimitedTransferCapacity() {
	s.setFlag(flagSessionUnlimitedTransferCapacity)
}

func (s *GnoSession) HasUnlimitedTransferCapacity() bool {
	return s.hasFlag(flagSessionUnlimitedTransferCapacity)
}

func (s *GnoSession) SetCanManageSessions() {
	s.setFlag(flagSessionCanManageSessions)
}

func (s *GnoSession) CanManageSessions() bool {
	return s.hasFlag(flagSessionCanManageSessions)
}

func (s *GnoSession) SetCanManagePackages() {
	s.setFlag(flagSessionCanManagePackages)
}

func (s *GnoSession) CanManagePackages() bool {
	return s.hasFlag(flagSessionCanManagePackages)
}

func (s *GnoSession) SetValidationOnly() {
	s.setFlag(flagSessionValidationOnly)
}

func (s *GnoSession) IsValidationOnly() bool {
	return s.hasFlag(flagSessionValidationOnly)
}

func ProtoGnoSession() std.AccountKey {
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

// NewGnoAccountWithMasterKey initializes an account with a root key
func NewGnoAccountWithMasterKey(address crypto.Address, pubKey crypto.PubKey) *GnoAccount {
	baseAccount := std.BaseAccount{
		Address: address,
	}
	
	account := &GnoAccount{
		BaseAccount: baseAccount,
	}

	// Set the root key - this will be the master key for the account
	account.SetRootKey(pubKey)

	return account
}

