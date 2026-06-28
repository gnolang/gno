package discovery

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"sync"
	"testing"

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
		wg.Add(1)

		go func() {
			defer wg.Done()
			s.AddPeers(addr)
			_ = s.Flush()
		}()
	}

	wg.Wait()

	// Final state should be consistent
	assert.Equal(t, 1, s.Size())
}
