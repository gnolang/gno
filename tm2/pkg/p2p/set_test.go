package p2p

import (
	"sort"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/p2p/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSet_Add(t *testing.T) {
	t.Parallel()

	var (
		numPeers = 100
		peers    = mock.GeneratePeers(t, numPeers)

		s = newSet()
	)

	for _, peer := range peers {
		// Add the peer
		s.Add(peer)

		// Make sure the peer is present
		assert.True(t, s.Has(peer.ID()))
	}

	assert.EqualValues(t, numPeers, s.NumInbound()+s.NumOutbound())
}

func TestSet_Remove(t *testing.T) {
	t.Parallel()

	var (
		numPeers = 100
		peers    = mock.GeneratePeers(t, numPeers)

		s = newSet()
	)

	// Add the initial peers
	for _, peer := range peers {
		// Add the peer
		s.Add(peer)

		// Make sure the peer is present
		require.True(t, s.Has(peer.ID()))
	}

	require.EqualValues(t, numPeers, s.NumInbound()+s.NumOutbound())

	// Remove the peers
	// Add the initial peers
	for _, peer := range peers {
		// Add the peer
		s.Remove(peer.ID())

		// Make sure the peer is present
		assert.False(t, s.Has(peer.ID()))
	}
}

func TestSet_Get(t *testing.T) {
	t.Parallel()

	t.Run("existing peer", func(t *testing.T) {
		t.Parallel()

		var (
			peers = mock.GeneratePeers(t, 100)
			s     = newSet()
		)

		for _, peer := range peers {
			id := peer.ID()
			s.Add(peer)

			assert.True(t, s.Get(id).ID() == id)
		}
	})

	t.Run("missing peer", func(t *testing.T) {
		t.Parallel()

		var (
			peers = mock.GeneratePeers(t, 100)
			s     = newSet()
		)

		for _, peer := range peers {
			s.Add(peer)
		}

		p := s.Get("random ID")
		assert.Nil(t, p)
	})
}

func TestSet_List(t *testing.T) {
	t.Parallel()

	t.Run("empty peer set", func(t *testing.T) {
		t.Parallel()

		// Empty set
		s := newSet()

		// Linearize the set
		assert.Len(t, s.List(), 0)
	})

	t.Run("existing peer set", func(t *testing.T) {
		t.Parallel()

		var (
			peers = mock.GeneratePeers(t, 100)
			s     = newSet()
		)

		for _, peer := range peers {
			s.Add(peer)
		}

		// Linearize the set
		listedPeers := s.List()

		require.Len(t, listedPeers, len(peers))

		// Make sure the lists are sorted
		// for easier comparison
		sort.Slice(listedPeers, func(i, j int) bool {
			return listedPeers[i].ID() < listedPeers[j].ID()
		})

		sort.Slice(peers, func(i, j int) bool {
			return peers[i].ID() < peers[j].ID()
		})

		// Compare the lists
		for index, listedPeer := range listedPeers {
			assert.Equal(t, listedPeer.ID(), peers[index].ID())
		}
	})
}
