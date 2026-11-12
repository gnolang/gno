package discovery

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"os"
	"sort"
	"sync"
	"time"

	osm "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/gnolang/gno/tm2/pkg/p2p/types"
)

// defaultMaxPeers is the default maximum number of peers kept in the address book.
const defaultMaxPeers = 1000

// connectedExpiry is how long a connection keeps an address ahead of shared
// ones after the node last held it.
const connectedExpiry = 24 * time.Hour

// knownAddress wraps a peer address with metadata used for eviction.
type knownAddress struct {
	Addr     *types.NetAddress
	LastSeen time.Time // when a peer last shared the address

	// LastConnected is when the node last held a connection it dialed to the
	// address, zero if it never has
	LastConnected time.Time

	// queued is when MarkConnected last asked for this mark to be saved
	queued time.Time
}

// connected reports whether the node held a connection it dialed to the
// address within connectedExpiry of now.
func (ka *knownAddress) connected(now time.Time) bool {
	return now.Sub(ka.LastConnected) < connectedExpiry
}

// storeJSON is the on-disk representation of the peer store.
type storeJSON struct {
	Peers []storeEntry `json:"peers"`
}

type storeEntry struct {
	Addr          string `json:"addr"`
	LastSeen      int64  `json:"last_seen"`
	LastConnected int64  `json:"last_connected,omitempty"`
}

// StoreOption configures the peer Store.
type StoreOption func(*Store)

// WithLogger sets the store logger.
func WithLogger(logger *slog.Logger) StoreOption {
	return func(s *Store) {
		if logger != nil {
			s.logger = logger
		}
	}
}

// WithMaxPeers sets the maximum number of peers the store keeps.
// When the limit is reached, entries are evicted, see evict.
func WithMaxPeers(maxPeers int) StoreOption {
	return func(s *Store) {
		if maxPeers > 0 {
			s.maxPeers = maxPeers
		}
	}
}

// Store persists discovered peer addresses to disk so they survive node restarts.
// It is safe for concurrent use.
type Store struct {
	mtx sync.RWMutex

	logger   *slog.Logger
	filePath string
	self     types.NetAddress // own address, never stored
	maxPeers int

	dirty      bool
	generation uint64

	peers map[string]*knownAddress // keyed by address string (ID@IP:Port)
}

// NewStore creates a new peer store backed by the given file path.
// self is the node's own address and is never stored.
// Existing peers are loaded from disk on construction.
func NewStore(filePath string, self types.NetAddress, opts ...StoreOption) (*Store, error) {
	s := &Store{
		filePath: filePath,
		self:     self,
		maxPeers: defaultMaxPeers,
		logger:   slog.Default(),
		peers:    make(map[string]*knownAddress),
	}

	for _, opt := range opts {
		opt(s)
	}

	if err := s.load(); err != nil {
		return nil, err
	}

	return s, nil
}

// AddPeers adds the given peer addresses to the store.
// The node's own address is ignored, as are nil entries.
// When the store exceeds its capacity, entries are evicted, see evict.
func (s *Store) AddPeers(addrs ...*types.NetAddress) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	for _, addr := range addrs {
		if addr == nil || addr.Same(s.self) {
			continue
		}

		key := addr.String()
		if ka, exists := s.peers[key]; exists {
			ka.LastSeen = time.Now()

			continue
		}

		s.dirty = true
		s.generation++

		s.peers[key] = &knownAddress{Addr: addr, LastSeen: time.Now()}
	}

	s.evict()
}

// MarkConnected records that the node holds a connection it dialed to the given
// address. An address the store does not hold is ignored.
func (s *Store) MarkConnected(addr *types.NetAddress) {
	if addr == nil {
		return
	}

	s.mtx.Lock()
	defer s.mtx.Unlock()

	ka, exists := s.peers[addr.String()]
	if !exists {
		return
	}

	ka.LastConnected = time.Now()

	// Save a new mark, or one half the expiry newer than the last one queued
	// for saving, rather than every renewal
	if ka.LastConnected.Sub(ka.queued) >= connectedExpiry/2 {
		ka.queued = ka.LastConnected
		s.dirty = true
		s.generation++
	}
}

// GetPeers returns all stored peer addresses.
func (s *Store) GetPeers() []*types.NetAddress {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	peers := make([]*types.NetAddress, 0, len(s.peers))
	for _, ka := range s.peers {
		copied := *ka.Addr
		copied.IP = append(net.IP(nil), ka.Addr.IP...)
		peers = append(peers, &copied)
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
// It is a no-op when the store has not changed since the last successful save.
func (s *Store) Save() error {
	s.mtx.Lock()
	if !s.dirty {
		s.mtx.Unlock()

		return nil
	}

	savedGeneration := s.generation

	entries := make([]storeEntry, 0, len(s.peers))
	for _, ka := range s.peers {
		entry := storeEntry{
			Addr:     ka.Addr.String(),
			LastSeen: ka.LastSeen.Unix(),
		}

		if !ka.LastConnected.IsZero() {
			entry.LastConnected = ka.LastConnected.Unix()
		}

		entries = append(entries, entry)
	}
	s.mtx.Unlock()

	data, err := json.MarshalIndent(storeJSON{Peers: entries}, "", "\t")
	if err != nil {
		return fmt.Errorf("unable to marshal peer store, %w", err)
	}

	if err := osm.WriteFileAtomic(s.filePath, data, 0o644); err != nil {
		return fmt.Errorf("unable to write peer store, %w", err)
	}

	// Only clear dirty if no modifications happened since the snapshot.
	// Otherwise the next Save will persist the newer peers.
	s.mtx.Lock()
	if s.generation == savedGeneration {
		s.dirty = false
	}
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

// evict removes entries until the store fits within maxPeers. Addresses the
// node has not held a connection to within connectedExpiry go first, those
// shared longest ago leading, then the ones it held longest ago. So an address
// a peer only shares never displaces one the node held within connectedExpiry.
// Caller must hold the mutex.
func (s *Store) evict() {
	if len(s.peers) <= s.maxPeers {
		return
	}

	entries := make([]*knownAddress, 0, len(s.peers))
	for _, ka := range s.peers {
		entries = append(entries, ka)
	}

	// Sort in eviction order
	now := time.Now()
	sort.Slice(entries, func(i, j int) bool {
		a, b := entries[i], entries[j]
		aConnected, bConnected := a.connected(now), b.connected(now)

		if aConnected != bConnected {
			return !aConnected
		}

		if !aConnected {
			return a.LastSeen.Before(b.LastSeen)
		}

		return a.LastConnected.Before(b.LastConnected)
	})

	toRemove := len(s.peers) - s.maxPeers
	for i := range toRemove {
		delete(s.peers, entries[i].Addr.String())
	}

	s.dirty = true
}

// load reads the peer store from disk. A missing file is not an error.
// A corrupt file is logged and treated as empty.
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
		corruptPath := s.filePath + ".corrupt"

		if copyErr := os.WriteFile(corruptPath, data, 0o644); copyErr != nil {
			s.logger.Warn("corrupt peer store file, failed to create backup", "file", s.filePath, "backup", corruptPath, "err", err, "copy_err", copyErr)
		} else {
			s.logger.Warn("corrupt peer store file, moved to backup", "file", s.filePath, "backup", corruptPath, "err", err)
		}

		return nil
	}

	s.mtx.Lock()
	defer s.mtx.Unlock()

	now := time.Now()

	for _, entry := range raw.Peers {
		addr, err := types.NewNetAddressFromString(entry.Addr)
		if err != nil {
			// Skip invalid addresses rather than failing the whole load
			continue
		}

		if addr.Same(s.self) {
			continue
		}

		lastSeen := time.Unix(entry.LastSeen, 0)
		if entry.LastSeen == 0 {
			lastSeen = now
		}

		ka := &knownAddress{
			Addr:     addr,
			LastSeen: lastSeen,
		}

		if entry.LastConnected != 0 {
			ka.LastConnected = time.Unix(entry.LastConnected, 0)
		}

		s.peers[addr.String()] = ka
	}

	return nil
}
