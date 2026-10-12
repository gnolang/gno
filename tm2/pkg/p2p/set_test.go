package p2p

import (
	"sort"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/p2p/mock"
	"github.com/gnolang/gno/tm2/pkg/p2p/types"
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
		require.NoError(t, s.Add(peer))

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
		require.NoError(t, s.Add(peer))

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

func TestSet_Add_DuplicateInbound(t *testing.T) {
	t.Parallel()

	var (
		key = types.GenerateNodeKey()
		s   = newSet()
	)

	peer := &mock.Peer{
		IDFn: func() types.ID {
			return key.ID()
		},
		IsOutboundFn: func() bool {
			return false
		},
	}

	// Add the same inbound peer twice
	require.NoError(t, s.Add(peer))
	require.Error(t, s.Add(peer))

	// Counter should reflect 1 peer, not 2
	assert.EqualValues(t, 1, s.NumInbound())
	assert.EqualValues(t, 0, s.NumOutbound())
	assert.Len(t, s.List(), 1)
}

func TestSet_Add_DuplicateOutbound(t *testing.T) {
	t.Parallel()

	var (
		key = types.GenerateNodeKey()
		s   = newSet()
	)

	peer := &mock.Peer{
		IDFn: func() types.ID {
			return key.ID()
		},
		IsOutboundFn: func() bool {
			return true
		},
	}

	// Add the same outbound peer twice
	require.NoError(t, s.Add(peer))
	require.Error(t, s.Add(peer))

	// Counter should reflect 1 peer, not 2
	assert.EqualValues(t, 0, s.NumInbound())
	assert.EqualValues(t, 1, s.NumOutbound())
	assert.Len(t, s.List(), 1)
}

func TestSet_Remove_NonExistent(t *testing.T) {
	t.Parallel()

	s := newSet()

	// Removing a non-existent peer should return false
	assert.False(t, s.Remove("nonexistent"))

	// Counters should remain at zero
	assert.EqualValues(t, 0, s.NumInbound())
	assert.EqualValues(t, 0, s.NumOutbound())
}

func TestSet_Add_Remove_DuplicateCycle(t *testing.T) {
	t.Parallel()

	var (
		key = types.GenerateNodeKey()
		s   = newSet()
	)

	peer := &mock.Peer{
		IDFn: func() types.ID {
			return key.ID()
		},
		IsOutboundFn: func() bool {
			return false
		},
	}

	// Add the same peer 10 times (simulates the reported attack)
	require.NoError(t, s.Add(peer))
	for range 9 {
		require.Error(t, s.Add(peer))
	}

	// Should only count as 1 peer
	assert.EqualValues(t, 1, s.NumInbound())
	assert.EqualValues(t, 0, s.NumOutbound())
	assert.Len(t, s.List(), 1)

	// Single remove should clean up completely
	assert.True(t, s.Remove(peer.ID()))
	assert.EqualValues(t, 0, s.NumInbound())
	assert.EqualValues(t, 0, s.NumOutbound())
	assert.Len(t, s.List(), 0)
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
			require.NoError(t, s.Add(peer))

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
			require.NoError(t, s.Add(peer))
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
			require.NoError(t, s.Add(peer))
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
