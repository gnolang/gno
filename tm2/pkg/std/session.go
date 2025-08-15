package std

import (
	"fmt"
	"time"

	"github.com/gnolang/gno/tm2/pkg/crypto"
)

// Session represents the interface that all session types must implement
type Session interface {
	// Core session information
	GetAddress() crypto.Address
	GetPubKey() crypto.PubKey
	GetSequence() uint64
	SetSequence(uint64)

	// Session state
	IsExpired() bool
	IsActive() bool
	IsFrozen() bool

	// Session capabilities
	CanSign(msg []byte) bool
	CanExecute(path string, method string) bool
	HasPermission(permission string) bool
}

// BaseSession provides the basic implementation of a Session
type BaseSession struct {
	Address        crypto.Address `json:"address"`
	PubKey         crypto.PubKey  `json:"pubkey"`
	Sequence       uint64         `json:"sequence"`
	CreatedAt      time.Time      `json:"created_at"`
	LastUsedAt     time.Time      `json:"last_used_at"`
	State          SessionState   `json:"state"`
	ExpirationTime time.Time      `json:"expiration_time"`
}

// SessionState represents the current state of a session
type SessionState uint8

const (
	_ SessionState = iota
	SessionStateActive
	SessionStateExpired
	SessionStateFrozen
)

// NewBaseSession creates a new BaseSession instance
func NewBaseSession(address crypto.Address, pubKey crypto.PubKey, sequence uint64) *BaseSession {
	now := time.Now()
	return &BaseSession{
		Address:    address,
		PubKey:     pubKey,
		Sequence:   sequence,
		CreatedAt:  now,
		LastUsedAt: now,
		State:      SessionStateActive,
	}
}

// GetAddress returns the account address associated with the session
func (s *BaseSession) GetAddress() crypto.Address {
	return s.Address
}

// GetPubKey returns the public key associated with the session
func (s *BaseSession) GetPubKey() crypto.PubKey {
	return s.PubKey
}

// GetSequence returns the current sequence number
func (s *BaseSession) GetSequence() uint64 {
	return s.Sequence
}

// SetSequence sets the sequence number
func (s *BaseSession) SetSequence(seq uint64) {
	s.Sequence = seq
}

// IsExpired checks if the session is in expired state or has passed its expiration time
func (s *BaseSession) IsExpired() bool {
	if s.State == SessionStateExpired {
		return true
	}
	if !s.ExpirationTime.IsZero() && time.Now().After(s.ExpirationTime) {
		s.State = SessionStateExpired
		return true
	}
	return false
}

// IsActive checks if the session is in active state
func (s *BaseSession) IsActive() bool {
	return s.State == SessionStateActive
}

// IsFrozen checks if the session is in frozen state
func (s *BaseSession) IsFrozen() bool {
	return s.State == SessionStateFrozen
}

// UpdateLastUsed updates the last used timestamp
func (s *BaseSession) UpdateLastUsed() {
	s.LastUsedAt = time.Now()
}

// Freeze puts the session in frozen state
func (s *BaseSession) Freeze() {
	s.State = SessionStateFrozen
}

// Unfreeze returns the session to active state
func (s *BaseSession) Unfreeze() {
	s.State = SessionStateActive
}

// Expire puts the session in expired state
func (s *BaseSession) Expire() {
	s.State = SessionStateExpired
}

// CanSign checks if the session can sign messages
// Base implementation always returns true if session is active
func (s *BaseSession) CanSign(msg []byte) bool {
	return s.IsActive()
}

// CanExecute checks if the session can execute a specific method
// Base implementation always returns true if session is active
func (s *BaseSession) CanExecute(path string, method string) bool {
	return s.IsActive()
}

// HasPermission checks if the session has a specific permission
// Base implementation always returns true if session is active
func (s *BaseSession) HasPermission(permission string) bool {
	return s.IsActive()
}

// String returns a string representation of the BaseSession
func (s *BaseSession) String() string {
	return fmt.Sprintf("BaseSession{Address: %v, PubKey: %v, Sequence: %d, State: %v, CreatedAt: %v, LastUsedAt: %v}",
		s.Address, s.PubKey, s.Sequence, s.State, s.CreatedAt, s.LastUsedAt)
}
 