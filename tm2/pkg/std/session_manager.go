package std

import (
	"fmt"
	"sync"
	"time"

	"github.com/gnolang/gno/tm2/pkg/crypto"
)

// Session related errors
var (
	errSessionNotFound   = fmt.Errorf("session not found")
	errSessionExpired    = fmt.Errorf("session has expired")
	errTooManySessions   = fmt.Errorf("maximum number of sessions reached")
	errSessionInvalid    = fmt.Errorf("invalid session")
	errSessionFrozen     = fmt.Errorf("session is frozen")
)


// SessionManager handles session lifecycle and validation
type SessionManager struct {
	mu             sync.RWMutex
	activeSessions map[string]*BaseSession
	config         SessionConfig
}

// SessionConfig contains configuration for session management
type SessionConfig struct {
	// max sessions per account
	MaxSessionsPerAccount int
	// session expiration time (0 means no expiration)
	ExpirationDuration time.Duration
	// session cleanup interval
	CleanupInterval time.Duration
}

// NewSessionManager creates a new session manager with given configuration
func NewSessionManager(config SessionConfig) *SessionManager {
	sm := &SessionManager{
		activeSessions: make(map[string]*BaseSession),
		config:         config,
	}

	// start session cleanup routine
	if config.CleanupInterval > 0 {
		go sm.cleanupLoop()
	}

	return sm
}

// CreateSession creates a new session for the given account
func (sm *SessionManager) CreateSession(address crypto.Address, pubKey crypto.PubKey) (*BaseSession, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// check max sessions per account
	if sm.countSessionsByAddress(address) >= sm.config.MaxSessionsPerAccount {
		return nil, errTooManySessions
	}

	// create new session
	session := NewBaseSession(address, pubKey, 0)

	// set expiration time
	if sm.config.ExpirationDuration > 0 {
		expirationTime := time.Now().Add(sm.config.ExpirationDuration)
		session.ExpirationTime = expirationTime
	}

	// save session
	key := pubKey.String()
	sm.activeSessions[key] = session

	return session, nil
}

// GetSession retrieves a session by public key
func (sm *SessionManager) GetSession(pubKey crypto.PubKey) (*BaseSession, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	key := pubKey.String()
	session, exists := sm.activeSessions[key]
	if !exists {
		return nil, errSessionNotFound
	}

	// check expired session
	if session.IsExpired() {
		return nil, errSessionExpired
	}

	return session, nil
}

// RemoveSession removes a session
func (sm *SessionManager) RemoveSession(pubKey crypto.PubKey) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	key := pubKey.String()
	if _, exists := sm.activeSessions[key]; !exists {
		return errSessionNotFound
	}

	delete(sm.activeSessions, key)
	return nil
}

// cleanupLoop periodically removes expired sessions
func (sm *SessionManager) cleanupLoop() {
	ticker := time.NewTicker(sm.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		sm.cleanup()
	}
}

// cleanup removes expired sessions
func (sm *SessionManager) cleanup() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for key, session := range sm.activeSessions {
		if session.IsExpired() {
			delete(sm.activeSessions, key)
		}
	}
}

// countSessionsByAddress counts active sessions for an address
func (sm *SessionManager) countSessionsByAddress(address crypto.Address) int {
	count := 0
	for _, session := range sm.activeSessions {
		if session.Address == address {
			count++
		}
	}
	return count
}

// GetSessionsByAddress returns all active sessions for an address
func (sm *SessionManager) GetSessionsByAddress(address crypto.Address) []*BaseSession {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var sessions []*BaseSession
	for _, session := range sm.activeSessions {
		if session.Address == address && !session.IsExpired() {
			sessions = append(sessions, session)
		}
	}
	return sessions
}

// UpdateSession updates session state
func (sm *SessionManager) UpdateSession(pubKey crypto.PubKey, updateFn func(*BaseSession) error) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	key := pubKey.String()
	session, exists := sm.activeSessions[key]
	if !exists {
		return errSessionNotFound
	}

	return updateFn(session)
}
