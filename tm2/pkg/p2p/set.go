package p2p

import (
	"net"
	"sync"
)

type Set struct {
	mux sync.RWMutex

	peers    map[ID]Peer
	outbound uint64
	inbound  uint64
}

// NewSet creates an empty peer set
func NewSet() *Set {
	return &Set{
		peers:    make(map[ID]Peer),
		outbound: 0,
		inbound:  0,
	}
}

// Add adds the peer to the set
func (s *Set) Add(peer Peer) {
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
func (s *Set) Has(peerKey ID) bool {
	s.mux.RLock()
	defer s.mux.RUnlock()

	_, exists := s.peers[peerKey]

	return exists
}

// HasIP returns true if the set contains the peer referred to by this IP
// address, otherwise false.
func (s *Set) HasIP(peerIP net.IP) bool {
	s.mux.RLock()
	defer s.mux.RUnlock()

	for _, p := range s.peers {
		if p.(Peer).RemoteIP().Equal(peerIP) {
			return true
		}
	}

	return false
}

// Get looks up a peer by the provided peerKey. Returns nil if peer is not
// found.
func (s *Set) Get(key ID) Peer {
	s.mux.RLock()
	defer s.mux.RUnlock()

	p, found := s.peers[key]
	if !found {
		// TODO change this to an error, it doesn't make
		// sense to propagate an implementation detail like this
		return nil
	}

	return p.(Peer)
}

// Remove discards peer by its Key, if the peer was previously memoized.
// Returns true if the peer was removed, and false if it was not found.
// in the set.
func (s *Set) Remove(key ID) bool {
	s.mux.Lock()
	defer s.mux.Unlock()

	p, found := s.peers[key]
	if !found {
		return false
	}

	delete(s.peers, key)

	if p.(Peer).IsOutbound() {
		s.outbound -= 1

		return true
	}

	s.inbound -= 1

	return true
}

// Size returns the number of unique peers in the peer table
func (s *Set) Size() int {
	s.mux.RLock()
	defer s.mux.RUnlock()

	return len(s.peers)
}

// NumInbound returns the number of inbound peers
func (s *Set) NumInbound() uint64 {
	s.mux.RLock()
	defer s.mux.RUnlock()

	return s.inbound
}

// NumOutbound returns the number of outbound peers
func (s *Set) NumOutbound() uint64 {
	s.mux.RLock()
	defer s.mux.RUnlock()

	return s.outbound
}

// List returns the list of peers
func (s *Set) List() []Peer {
	s.mux.RLock()
	defer s.mux.RUnlock()

	peers := make([]Peer, 0)
	for _, p := range s.peers {
		peers = append(peers, p.(Peer))
	}

	return peers
}
