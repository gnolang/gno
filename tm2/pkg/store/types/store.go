package types

import (
	"bytes"
	"fmt"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	std "github.com/gnolang/gno/tm2/pkg/std"
)

type Store interface {
	// Get returns nil iff key doesn't exist. Panics on nil key.
	Get(key []byte) []byte

	// Has checks if a key exists. Panics on nil key.
	Has(key []byte) bool

	// Set sets the key. Panics on nil key or value.
	Set(key, value []byte)

	// Delete deletes the key. Panics on nil key.
	Delete(key []byte)

	// Iterator over a domain of keys in ascending order. End is exclusive.
	// Start must be less than end, or the Iterator is invalid.
	// Iterator must be closed by caller.
	// To iterate over entire domain, use store.Iterator(nil, nil)
	// CONTRACT: No writes may happen within a domain while an iterator exists over it.
	// Exceptionally allowed for cachekv.Store, safe to write in the modules.
	Iterator(start, end []byte) Iterator

	// Iterator over a domain of keys in descending order. End is exclusive.
	// Start must be less than end, or the Iterator is invalid.
	// Iterator must be closed by caller.
	// CONTRACT: No writes may happen within a domain while an iterator exists over it.
	// Exceptionally allowed for cachekv.Store, safe to write in the modules.
	ReverseIterator(start, end []byte) Iterator

	// Returns a cache-wrapped store.
	CacheWrap() Store

	// If cache-wrapped store, writes to underlying store.
	// Does not writes through layers of cache.
	Write()
}

// Alias iterator to db's Iterator for convenience.
type Iterator = dbm.Iterator

// Queryable allows a Store to expose internal state to the abci.Query
// interface. Multistore can route requests to the proper Store.
//
// This is an optional, but useful extension to any CommitStore
type Queryable interface {
	Query(abci.RequestQuery) abci.ResponseQuery
}

type Printer interface {
	Print()
}

type WriteThrougher interface {
	WriteThrough(int)
}

// Can be called to clear empty reads all caches.
// NOTE: currently only works for *CacheStore
type ClearThrougher interface {
	ClearThrough()
}

// Can be called to write through all caches.
// NOTE: currently only works for *CacheStore
type Flusher interface {
	Flush()
}

type Writer interface {
	Write()
}

// ----------------------------------------
// MultiStore

type MultiStore interface {
	// Convenience for fetching substores.
	// If the store does not exist, panics.
	GetStore(StoreKey) Store

	// Returns a cache-wrapped multi-store.
	MultiCacheWrap() MultiStore

	// If cache-wrapped multi-store, flushes to underlying store.
	MultiWrite()
}

// ----------------------------------------
// Committer, CommitID

// Something that can persist to disk
type Committer interface {
	Commit() CommitID
	LastCommitID() CommitID
	GetStoreOptions() StoreOptions
	SetStoreOptions(StoreOptions)
	LoadLatestVersion() error

	// Load a specific persisted version. When you load an old version, or when
	// the last commit attempt didn't complete, the next commit after loading
	// must be idempotent (return the same commit id). Otherwise the behavior is
	// undefined.
	LoadVersion(ver int64) error
}

// Stores of MultiStore must implement CommitStore.
type CommitStore interface {
	Committer
	Store
}

// Used by MultiStores to mount a new store.
type CommitStoreConstructor func(db dbm.DB, opts StoreOptions) CommitStore

// A non-cache MultiStore.
type CommitMultiStore interface {
	Committer
	MultiStore

	// Mount a store of type using the given db.
	// If db == nil, the new store will use the CommitMultiStore db.
	MountStoreWithDB(key StoreKey, cons CommitStoreConstructor, db dbm.DB)

	// Panics on a nil key.
	GetCommitStore(key StoreKey) CommitStore

	// MultiImmutableCacheWrapWithVersion is analogous to MultiCacheWrap
	// except that it attempts to load immutable stores at a given version
	// (height). An error is returned if any store cannot be loaded. This
	// should only be used for querying and iterating at past heights.
	MultiImmutableCacheWrapWithVersion(version int64) (MultiStore, error)
}

// CommitID contains the tree version number and its merkle root.
type CommitID struct {
	Version int64
	Hash    []byte
}

func (cid CommitID) Equals(oid CommitID) bool {
	return cid.Version == oid.Version && bytes.Equal(cid.Hash, oid.Hash)
}

func (cid CommitID) IsZero() bool {
	return cid.Version == 0 && len(cid.Hash) == 0
}

func (cid CommitID) String() string {
	return fmt.Sprintf("CommitID{%v:%X}", cid.Hash, cid.Version)
}

// ----------------------------------------
// Keys for accessing substores

// StoreKey is a key used to index stores in a MultiStore.
type StoreKey interface {
	Name() string
	String() string
}

type storeKey struct {
	name string
}

// NewStoreKey returns a new pointer to a StoreKey.
// Use a pointer so keys don't collide.
func NewStoreKey(name string) *storeKey {
	return &storeKey{
		name: name,
	}
}

func (key *storeKey) Name() string {
	return key.name
}

func (key *storeKey) String() string {
	return fmt.Sprintf("storeKey{%p, %s}", key, key.name)
}

// ----------------------------------------
// KVPair

type KVPair = std.KVPair
