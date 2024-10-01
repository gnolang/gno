package p2p

import (
	"net"
	"sync"
)

type Set struct {
	peers sync.Map // p2p.ID -> p2p.Peer
}

// NewSet creates an empty peer set
func NewSet() *Set {
	return &Set{}
}

// Add adds the peer to the set
func (s *Set) Add(peer Peer) {
	s.peers.Store(peer.ID(), peer)
}

// Has returns true if the set contains the peer referred to by this
// peerKey, otherwise false.
func (s *Set) Has(peerKey ID) bool {
	_, ok := s.peers.Load(peerKey)

	return ok
}

// HasIP returns true if the set contains the peer referred to by this IP
// address, otherwise false.
func (s *Set) HasIP(peerIP net.IP) bool {
	hasIP := false

	s.peers.Range(func(_, value interface{}) bool {
		peer := value.(Peer)

		if peer.RemoteIP().Equal(peerIP) {
			hasIP = true

			return false
		}

		return true
	})

	return hasIP
}

// Get looks up a peer by the provided peerKey. Returns nil if peer is not
// found.
func (s *Set) Get(key ID) Peer {
	peerRaw, found := s.peers.Load(key)
	if !found {
		// TODO change this to an error, it doesn't make
		// sense to propagate an implementation detail like this
		return nil
	}

	return peerRaw.(Peer)
}

// Remove discards peer by its Key, if the peer was previously memoized.
// Returns true if the peer was removed, and false if it was not found.
// in the set.
func (s *Set) Remove(key ID) bool {
	_, existed := s.peers.LoadAndDelete(key)

	return existed
}

// Size returns the number of unique peers in the peer table
func (s *Set) Size() int {
	size := 0

	s.peers.Range(func(_, _ any) bool {
		size++

		return true
	})

	return size
}

// List returns the list of peers
func (s *Set) List() []Peer {
	peers := make([]Peer, 0)

	s.peers.Range(func(_, value any) bool {
		peers = append(peers, value.(Peer))

		return true
	})

	return peers
}
