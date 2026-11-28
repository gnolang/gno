package discovery

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/p2p/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// peerStoreFile is the default file name for the persisted discovered peer set.
// Defined here because it is only used by tests.
const peerStoreFile = "addrbook.json"

// generateTestAddress builds a valid NetAddress with a random peer ID.
func generateTestAddress(t *testing.T, host string, port uint16) *types.NetAddress {
	t.Helper()

	key := types.GenerateNodeKey()

	return &types.NetAddress{
		ID:   key.ID(),
		IP:   net.ParseIP(host),
		Port: port,
	}
}

func TestStore_New_Empty(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), peerStoreFile)

	s, err := NewStore(path, types.NetAddress{})
	require.NoError(t, err)

	assert.Empty(t, s.GetPeers())
}

func TestStore_AddAndGetPeers(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), peerStoreFile)

	s, err := NewStore(path, types.NetAddress{})
	require.NoError(t, err)

	addr1 := generateTestAddress(t, "1.2.3.4", 26656)
	addr2 := generateTestAddress(t, "5.6.7.8", 26656)

	s.AddPeers(addr1, addr2)

	peers := s.GetPeers()
	assert.Len(t, peers, 2)

	// Adding the same address again should be idempotent
	s.AddPeers(addr1)
	assert.Equal(t, 2, s.Size())
}

func TestStore_AddPeers_NilSkipped(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), peerStoreFile)

	s, err := NewStore(path, types.NetAddress{})
	require.NoError(t, err)

	addr := generateTestAddress(t, "1.2.3.4", 26656)

	s.AddPeers(nil, addr, nil)

	assert.Equal(t, 1, s.Size())
}

func TestStore_AddPeers_SelfAddressFiltered(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), peerStoreFile)

	self := generateTestAddress(t, "1.2.3.4", 26656)
	other := generateTestAddress(t, "5.6.7.8", 26656)

	s, err := NewStore(path, *self)
	require.NoError(t, err)

	s.AddPeers(self, other)

	// Only the non-self address should be stored
	assert.Equal(t, 1, s.Size())
	assert.Contains(t, s.GetPeers(), other)
}

func TestStore_SaveAndReload(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), peerStoreFile)

	// Create and populate store
	s, err := NewStore(path, types.NetAddress{})
	require.NoError(t, err)

	addr1 := generateTestAddress(t, "1.2.3.4", 26656)
	addr2 := generateTestAddress(t, "5.6.7.8", 26656)

	s.AddPeers(addr1, addr2)
	require.NoError(t, s.Save())

	// The file should exist on disk
	_, err = os.Stat(path)
	require.NoError(t, err)

	// Reload the store from the same file
	s2, err := NewStore(path, types.NetAddress{})
	require.NoError(t, err)

	reloaded := s2.GetPeers()
	require.Len(t, reloaded, 2)

	// Verify both addresses survived the round-trip
	reloadedStrs := make(map[string]struct{}, len(reloaded))
	for _, a := range reloaded {
		reloadedStrs[a.String()] = struct{}{}
	}

	assert.Contains(t, reloadedStrs, addr1.String())
	assert.Contains(t, reloadedStrs, addr2.String())
}

func TestStore_FileFormat_HumanReadable(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), peerStoreFile)

	s, err := NewStore(path, types.NetAddress{})
	require.NoError(t, err)

	addr := generateTestAddress(t, "1.2.3.4", 26656)
	s.AddPeers(addr)
	require.NoError(t, s.Save())

	// The file should contain human-readable address strings
	data, err := os.ReadFile(path)
	require.NoError(t, err)

	var raw storeJSON
	require.NoError(t, json.Unmarshal(data, &raw))

	require.Len(t, raw.Peers, 1)
	assert.Equal(t, addr.String(), raw.Peers[0].Addr)
}

func TestStore_Load_SkipsInvalidAddresses(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), peerStoreFile)

	validAddr := generateTestAddress(t, "1.2.3.4", 26656)

	// Write a file with a mix of valid and invalid addresses
	raw := storeJSON{
		Peers: []storeEntry{
			{Addr: validAddr.String()},
			{Addr: "not-a-valid-address"},
		},
	}

	data, err := json.MarshalIndent(raw, "", "\t")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, data, 0o644))

	s, err := NewStore(path, types.NetAddress{})
	require.NoError(t, err)

	// Only the valid address should be loaded
	assert.Equal(t, 1, s.Size())
}

func TestStore_Eviction_MaxSize(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), peerStoreFile)

	s, err := NewStore(path, types.NetAddress{}, WithMaxPeers(3))
	require.NoError(t, err)

	// Add 5 peers
	addrs := make([]*types.NetAddress, 5)
	for i := range addrs {
		addrs[i] = generateTestAddress(t, "1.2.3.4", uint16(1000+i))
	}

	s.AddPeers(addrs...)

	// Should be trimmed to max size
	assert.Equal(t, 3, s.Size())
}

func TestStore_Eviction_KeepsConnected(t *testing.T) {
	t.Parallel()

	// held returns how many of the given addresses the store holds
	held := func(s *Store, addrs []*types.NetAddress) int {
		stored := make(map[string]struct{}, s.Size())
		for _, a := range s.GetPeers() {
			stored[a.String()] = struct{}{}
		}

		count := 0

		for _, a := range addrs {
			if _, ok := stored[a.String()]; ok {
				count++
			}
		}

		return count
	}

	// fill returns a full store, every address in it connected to
	fill := func(t *testing.T, path string) (*Store, []*types.NetAddress) {
		t.Helper()

		s, err := NewStore(path, types.NetAddress{})
		require.NoError(t, err)

		connected := make([]*types.NetAddress, 0, defaultMaxPeers)
		for i := range defaultMaxPeers {
			connected = append(connected, generateTestAddress(t, "10.0.0.1", uint16(20000+i)))
		}

		s.AddPeers(connected...)

		for _, a := range connected {
			s.MarkConnected(a)
		}

		return s, connected
	}

	// share shares n fresh addresses, maxPeersShared to a response
	share := func(t *testing.T, s *Store, n int) {
		t.Helper()

		for n > 0 {
			shared := make([]*types.NetAddress, 0, maxPeersShared)
			for i := range min(n, maxPeersShared) {
				shared = append(shared, generateTestAddress(t, "203.0.113.2", uint16(1+i)))
			}

			s.AddPeers(shared...)
			n -= len(shared)
		}
	}

	// lapse dates the given addresses' connection past connectedExpiry
	lapse := func(s *Store, addrs []*types.NetAddress) {
		s.mtx.Lock()
		defer s.mtx.Unlock()

		for _, a := range addrs {
			s.peers[a.String()].LastConnected = time.Now().Add(-connectedExpiry - time.Minute)
		}
	}

	// sybils returns 30 addresses the store holds, each connected to once
	sybils := func(t *testing.T, s *Store) []*types.NetAddress {
		t.Helper()

		addrs := make([]*types.NetAddress, 30)
		for i := range addrs {
			addrs[i] = generateTestAddress(t, "198.51.100.1", uint16(30000+i))
		}

		s.AddPeers(addrs...)

		for _, a := range addrs {
			s.MarkConnected(a)
		}

		return addrs
	}

	t.Run("lapsed connection does not keep an address", func(t *testing.T) {
		t.Parallel()

		s := newTestStore(t)
		addrs := sybils(t, s)
		lapse(s, addrs)

		share(t, s, 2_010)

		assert.Equal(t, 0, held(s, addrs))
	})

	t.Run("lapsed connection read back after a restart", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), peerStoreFile)

		s, err := NewStore(path, types.NetAddress{})
		require.NoError(t, err)

		addrs := sybils(t, s)
		lapse(s, addrs)
		require.NoError(t, s.Save())

		reloaded, err := NewStore(path, types.NetAddress{})
		require.NoError(t, err)

		share(t, reloaded, 2_010)

		assert.Equal(t, 0, held(reloaded, addrs))
	})

	t.Run("lapsed connections give way to live addresses", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), peerStoreFile)

		s, dead := fill(t, path)
		lapse(s, dead)

		live := make([]*types.NetAddress, 50)
		for i := range live {
			live[i] = generateTestAddress(t, "192.0.2.1", uint16(40000+i))

			s.AddPeers(live[i])
			s.MarkConnected(live[i])
		}

		assert.Equal(t, len(live), held(s, live))

		require.NoError(t, s.Save())

		reloaded, err := NewStore(path, types.NetAddress{})
		require.NoError(t, err)

		assert.Equal(t, len(live), held(reloaded, live))
	})

	t.Run("shared addresses do not displace connected ones", func(t *testing.T) {
		t.Parallel()

		s, connected := fill(t, filepath.Join(t.TempDir(), peerStoreFile))

		share(t, s, 1_020)

		assert.Equal(t, defaultMaxPeers, s.Size())
		assert.Equal(t, defaultMaxPeers, held(s, connected))
	})

	t.Run("connected address shared again stays connected", func(t *testing.T) {
		t.Parallel()

		s, connected := fill(t, filepath.Join(t.TempDir(), peerStoreFile))

		s.AddPeers(connected...)
		share(t, s, 1_020)

		assert.Equal(t, defaultMaxPeers, held(s, connected))
	})

	t.Run("connected state survives a restart", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), peerStoreFile)

		s, connected := fill(t, path)
		require.NoError(t, s.Save())

		reloaded, err := NewStore(path, types.NetAddress{})
		require.NoError(t, err)

		share(t, reloaded, 1_020)

		assert.Equal(t, defaultMaxPeers, held(reloaded, connected))
	})

	t.Run("never-connected addresses evicted oldest first", func(t *testing.T) {
		t.Parallel()

		s, err := NewStore(filepath.Join(t.TempDir(), peerStoreFile), types.NetAddress{}, WithMaxPeers(3))
		require.NoError(t, err)

		addrs := make([]*types.NetAddress, 5)
		for i := range addrs {
			addrs[i] = generateTestAddress(t, "1.2.3.4", uint16(1000+i))
		}

		// The first is connected to, the next two only shared
		s.AddPeers(addrs[:3]...)
		s.MarkConnected(addrs[0])

		time.Sleep(time.Millisecond)
		s.AddPeers(addrs[3:]...)

		assert.Equal(t, 1, held(s, addrs[:1]))
		assert.Equal(t, 0, held(s, addrs[1:3]))
		assert.Equal(t, 2, held(s, addrs[3:]))
	})

	t.Run("connected addresses evicted oldest connection first", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), peerStoreFile)

		// A file holding more connected addresses than the store keeps
		addrs := make([]*types.NetAddress, 4)
		raw := storeJSON{}
		now := time.Now().Unix()

		for i := range addrs {
			addrs[i] = generateTestAddress(t, "1.2.3.4", uint16(1000+i))
			raw.Peers = append(raw.Peers, storeEntry{
				Addr:          addrs[i].String(),
				LastSeen:      1000,
				LastConnected: now - 10 + int64(i),
			})
		}

		data, err := json.Marshal(raw)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(path, data, 0o644))

		s, err := NewStore(path, types.NetAddress{}, WithMaxPeers(3))
		require.NoError(t, err)

		s.AddPeers(generateTestAddress(t, "5.6.7.8", 26656))

		assert.Equal(t, 0, held(s, addrs[:1]))
		assert.Equal(t, 3, held(s, addrs[1:]))
	})

	t.Run("recent connection marked again without a save", func(t *testing.T) {
		t.Parallel()

		s := newTestStore(t)
		a := generateTestAddress(t, "1.2.3.4", 26656)

		s.AddPeers(a)
		s.MarkConnected(a)
		require.NoError(t, s.Save())

		s.MarkConnected(a)

		// The mark was last queued for saving half the expiry ago, while every
		// round since renewed it in memory
		s.mtx.Lock()
		dirty := s.dirty
		s.peers[a.String()].queued = time.Now().Add(-connectedExpiry/2 - time.Minute)
		s.mtx.Unlock()

		assert.False(t, dirty)

		s.MarkConnected(a)

		s.mtx.RLock()
		defer s.mtx.RUnlock()

		assert.True(t, s.dirty)
	})

	t.Run("address the store does not hold left out", func(t *testing.T) {
		t.Parallel()

		s, err := NewStore(filepath.Join(t.TempDir(), peerStoreFile), types.NetAddress{})
		require.NoError(t, err)

		s.MarkConnected(generateTestAddress(t, "1.2.3.4", 26656))

		assert.Zero(t, s.Size())
	})
}

func TestStore_Save_NoOpWhenNotDirty(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), peerStoreFile)

	s, err := NewStore(path, types.NetAddress{})
	require.NoError(t, err)

	addr := generateTestAddress(t, "1.2.3.4", 26656)
	s.AddPeers(addr)
	require.NoError(t, s.Save())

	// File should exist after the first save
	info1, err := os.Stat(path)
	require.NoError(t, err)

	// Saving again without changes should be a no-op
	require.NoError(t, s.Save())

	info2, err := os.Stat(path)
	require.NoError(t, err)

	// Same modification time (no write happened)
	assert.Equal(t, info1.ModTime(), info2.ModTime())
}

func TestStore_Flush_ForcesWrite(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), peerStoreFile)

	s, err := NewStore(path, types.NetAddress{})
	require.NoError(t, err)

	// Flush without any peers should still create the file
	require.NoError(t, s.Flush())

	_, err = os.Stat(path)
	require.NoError(t, err)
}

func TestStore_Load_CorruptFileTreatedAsEmpty(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), peerStoreFile)

	// Write corrupt JSON to the file
	require.NoError(t, os.WriteFile(path, []byte("{not valid json"), 0o644))

	s, err := NewStore(path, types.NetAddress{})
	require.NoError(t, err)

	// The store should start empty despite the corrupt file
	assert.Equal(t, 0, s.Size())

	// A backup of the corrupt file should have been created
	corruptPath := path + ".corrupt"
	_, err = os.Stat(corruptPath)
	assert.NoError(t, err, "corrupt backup file should exist")

	// Saving should work and overwrite the corrupt file with valid data
	addr := generateTestAddress(t, "1.2.3.4", 26656)
	s.AddPeers(addr)
	require.NoError(t, s.Flush())

	// Reload should now work correctly
	s2, err := NewStore(path, types.NetAddress{})
	require.NoError(t, err)
	assert.Equal(t, 1, s2.Size())
}

func TestStore_SaveAtomic_NoCorruptionOnConcurrentAccess(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), peerStoreFile)

	s, err := NewStore(path, types.NetAddress{})
	require.NoError(t, err)

	addr := generateTestAddress(t, "1.2.3.4", 26656)

	var wg sync.WaitGroup

	// Concurrently add peers and flush (forced save) to disk
	for range 20 {
		wg.Go(func() {
			s.AddPeers(addr)
			_ = s.Flush()
		})
	}

	wg.Wait()

	// Final state should be consistent
	assert.Equal(t, 1, s.Size())
}
