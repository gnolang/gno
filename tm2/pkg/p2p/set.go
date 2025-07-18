package p2p

import (
	"sync"

	"github.com/gnolang/gno/tm2/pkg/p2p/types"
)

type set struct {
	mux sync.RWMutex

	peers    map[types.ID]PeerConn
	outbound uint64
	inbound  uint64
}

// newSet creates an empty peer set
func newSet() *set {
	return &set{
		peers:    make(map[types.ID]PeerConn),
		outbound: 0,
		inbound:  0,
	}
}

// Add adds the peer to the set
func (s *set) Add(peer PeerConn) {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.peers[peer.ID()] = peer

	if peer.IsOutbound() {
		s.outbound += 1

		return
	}

	s.inbound += 1
}

// Has returns true if the set contains the peer referred to by this
// peerKey, otherwise false.
func (s *set) Has(peerKey types.ID) bool {
	s.mux.RLock()
	defer s.mux.RUnlock()

	_, exists := s.peers[peerKey]

	return exists
}

// Get looks up a peer by the peer ID. Returns nil if peer is not
// found.
func (s *set) Get(key types.ID) PeerConn {
	s.mux.RLock()
	defer s.mux.RUnlock()

	p, found := s.peers[key]
	if !found {
		// TODO change this to an error, it doesn't make
		// sense to propagate an implementation detail like this
		return nil
	}

	return p
}

// Remove discards peer by its Key, if the peer was previously memoized.
// Returns true if the peer was removed, and false if it was not found.
// in the set.
func (s *set) Remove(key types.ID) bool {
	s.mux.Lock()
	defer s.mux.Unlock()

	p, found := s.peers[key]
	if !found {
		return false
	}

	delete(s.peers, key)

	if p.IsOutbound() {
		s.outbound -= 1

		return true
	}

	s.inbound -= 1

	return true
}

// NumInbound returns the number of inbound peers
func (s *set) NumInbound() uint64 {
	s.mux.RLock()
	defer s.mux.RUnlock()

	return s.inbound
}

// NumOutbound returns the number of outbound peers
func (s *set) NumOutbound() uint64 {
	s.mux.RLock()
	defer s.mux.RUnlock()

	return s.outbound
}

// List returns the list of peers
func (s *set) List() []PeerConn {
	s.mux.RLock()
	defer s.mux.RUnlock()

	peers := make([]PeerConn, 0)
	for _, p := range s.peers {
		peers = append(peers, p)
	}

	return peers
}
