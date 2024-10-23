package p2p

import (
	"net"
	"sort"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/p2p/mock"
	"github.com/gnolang/gno/tm2/pkg/p2p/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generatePeers generates random node peers
func generatePeers(t *testing.T, count int) []*mock.Peer {
	t.Helper()

	peers := make([]*mock.Peer, count)

	for i := 0; i < count; i++ {
		id := types.GenerateNodeKey().ID()
		peers[i] = &mock.Peer{
			IDFn: func() types.ID {
				return id
			},
		}
	}

	return peers
}

func TestSet_Add(t *testing.T) {
	t.Parallel()

	var (
		numPeers = 100
		peers    = generatePeers(t, numPeers)

		s = newSet()
	)

	for _, peer := range peers {
		// Add the peer
		s.Add(peer)

		// Make sure the peer is present
		assert.True(t, s.Has(peer.ID()))
	}

	assert.EqualValues(t, numPeers, s.Size())
}

func TestSet_Remove(t *testing.T) {
	t.Parallel()

	var (
		numPeers = 100
		peers    = generatePeers(t, numPeers)

		s = newSet()
	)

	// Add the initial peers
	for _, peer := range peers {
		// Add the peer
		s.Add(peer)

		// Make sure the peer is present
		require.True(t, s.Has(peer.ID()))
	}

	require.EqualValues(t, numPeers, s.Size())

	// Remove the peers
	// Add the initial peers
	for _, peer := range peers {
		// Add the peer
		s.Remove(peer.ID())

		// Make sure the peer is present
		assert.False(t, s.Has(peer.ID()))
	}
}

func TestSet_HasIP(t *testing.T) {
	t.Parallel()

	t.Run("present peer with IP", func(t *testing.T) {
		t.Parallel()

		var (
			peers = generatePeers(t, 100)
			ip    = net.ParseIP("0.0.0.0")

			s = newSet()
		)

		// Make sure at least one peer has the set IP
		peers[len(peers)/2].RemoteIPFn = func() net.IP {
			return ip
		}

		// Add the peers
		for _, peer := range peers {
			s.Add(peer)
		}

		// Make sure the peer is present
		assert.True(t, s.HasIP(ip))
	})

	t.Run("missing peer with IP", func(t *testing.T) {
		t.Parallel()

		var (
			peers = generatePeers(t, 100)
			ip    = net.ParseIP("0.0.0.0")

			s = newSet()
		)

		// Add the peers
		for _, peer := range peers {
			s.Add(peer)
		}

		// Make sure the peer is not present
		assert.False(t, s.HasIP(ip))
	})
}

func TestSet_Get(t *testing.T) {
	t.Parallel()

	t.Run("existing peer", func(t *testing.T) {
		t.Parallel()

		var (
			peers = generatePeers(t, 100)
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
			peers = generatePeers(t, 100)
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
			peers = generatePeers(t, 100)
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
