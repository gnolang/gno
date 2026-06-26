package discovery

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	osm "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/gnolang/gno/tm2/pkg/p2p/types"
)

// peerStoreFile is the default file name for the persisted discovered peer set
const peerStoreFile = "addrbook.json"

// storeJSON is the on-disk representation of the peer store.
// Addresses are serialized as "ID@IP:Port" strings to avoid
// encoding/json's default []byte handling of net.IP.
type storeJSON struct {
	Addrs []string `json:"addrs"`
}

// Store persists discovered peer addresses to disk so they survive node restarts.
// It is safe for concurrent use.
type Store struct {
	mtx   sync.RWMutex
	dirty bool

	filePath string
	peers    map[string]*types.NetAddress // keyed by address string (ID@IP:Port)
}

// NewStore creates a new peer store backed by the given file path.
// Existing peers are loaded from disk on construction.
func NewStore(filePath string) (*Store, error) {
	s := &Store{
		filePath: filePath,
		peers:    make(map[string]*types.NetAddress),
	}

	if err := s.load(); err != nil {
		return nil, fmt.Errorf("unable to load peer store, %w", err)
	}

	return s, nil
}

// AddPeers adds the given peer addresses to the store, ignoring nil entries.
func (s *Store) AddPeers(addrs ...*types.NetAddress) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	for _, addr := range addrs {
		if addr == nil {
			continue
		}

		key := addr.String()
		if _, exists := s.peers[key]; !exists {
			s.dirty = true
		}

		s.peers[key] = addr
	}
}

// GetPeers returns a copy of all stored peer addresses.
func (s *Store) GetPeers() []*types.NetAddress {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	peers := make([]*types.NetAddress, 0, len(s.peers))
	for _, addr := range s.peers {
		peers = append(peers, addr)
	}

	return peers
}

// Size returns the number of stored peer addresses.
func (s *Store) Size() int {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	return len(s.peers)
}

// Save persists the peer store to disk atomically.
// It is a no-op when the store has not changed since the last save.
func (s *Store) Save() error {
	s.mtx.Lock()
	if !s.dirty {
		s.mtx.Unlock()

		return nil
	}

	addrs := make([]string, 0, len(s.peers))
	for _, addr := range s.peers {
		addrs = append(addrs, addr.String())
	}
	s.mtx.Unlock()

	data, err := json.MarshalIndent(storeJSON{Addrs: addrs}, "", "\t")
	if err != nil {
		return fmt.Errorf("unable to marshal peer store, %w", err)
	}

	if err := osm.WriteFileAtomic(s.filePath, data, 0o644); err != nil {
		return fmt.Errorf("unable to write peer store, %w", err)
	}

	s.mtx.Lock()
	s.dirty = false
	s.mtx.Unlock()

	return nil
}

// Flush forces a save of the peer store to disk regardless of dirty state.
func (s *Store) Flush() error {
	s.mtx.Lock()
	s.dirty = true
	s.mtx.Unlock()

	return s.Save()
}

// load reads the peer store from disk. A missing or corrupt file is not an error;
// the store simply starts empty.
func (s *Store) load() error {
	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return fmt.Errorf("unable to read peer store, %w", err)
	}

	var raw storeJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		// A corrupt file is treated as empty rather than failing startup.
		// The next save will overwrite it with valid data.
		return nil
	}

	s.mtx.Lock()
	defer s.mtx.Unlock()

	for _, addrStr := range raw.Addrs {
		addr, err := types.NewNetAddressFromString(addrStr)
		if err != nil {
			// Skip invalid addresses rather than failing the whole load
			continue
		}

		s.peers[addr.String()] = addr
	}

	return nil
}
