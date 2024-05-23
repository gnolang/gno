package nonce

import (
	"std"

	"gno.land/p/demo/avl"
)

type Manager struct {
	nonces *avl.Tree
}

// New creates a new instance of Manager
func New() *Manager {
	return &Manager{
		nonces: avl.NewTree(),
	}
}

// GetNonce returns the next unused nonce for an address
func (m *Manager) GetNonce(addr std.Address) uint64 {
	key := addr.String()

	nonce, found := m.nonces.Get(key)
	if !found {
		m.nonces.Set(key, 0)
		return 0
	}

	return nonce.(uint64)
}

// UseNonce consumes a nonce and returns the current value, then increments the nonce
func (m *Manager) UseNonce(addr std.Address) {
	key := addr.String()

	var currentNonce uint64 = 0

	val, found := m.nonces.Get(key)
	if found {
		currentNonce = val.(uint64)
	}

	m.nonces.Set(key, currentNonce+1)
}
