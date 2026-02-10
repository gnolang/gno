package dbadapter

import (
	dbm "github.com/gnolang/gno/tm2/pkg/db"

	"github.com/gnolang/gno/tm2/pkg/store/cache"
	"github.com/gnolang/gno/tm2/pkg/store/types"
)

// Implements CommitStoreConstructor.
func StoreConstructor(db dbm.DB, opts types.StoreOptions) types.CommitStore {
	return Store{
		DB: db,
	}
}

// Wrapper type for dbm.Db with implementation of Store
type Store struct {
	DB dbm.DB
}

// Get returns nil iff key doesn't exist. Panics on nil key.
func (dsa Store) Get(key []byte) []byte {
	v, err := dsa.DB.Get(key)
	if err != nil {
		panic(err)
	}
	return v
}

// Has checks if a key exists. Panics on nil key.
func (dsa Store) Has(key []byte) bool {
	v, err := dsa.DB.Has(key)
	if err != nil {
		panic(err)
	}
	return v
}

// Set sets the key. Panics on nil key or value.
func (dsa Store) Set(key, value []byte) {
	err := dsa.DB.Set(key, value)
	if err != nil {
		panic(err)
	}
}

// Delete deletes the key. Panics on nil key.
func (dsa Store) Delete(key []byte) {
	err := dsa.DB.Delete(key)
	if err != nil {
		panic(err)
	}
}

// Iterator over a domain of keys in ascending order.
func (dsa Store) Iterator(start, end []byte) types.Iterator {
	it, err := dsa.DB.Iterator(start, end)
	if err != nil {
		panic(err)
	}
	return it
}

// Iterator over a domain of keys in descending order.
func (dsa Store) ReverseIterator(start, end []byte) types.Iterator {
	it, err := dsa.DB.ReverseIterator(start, end)
	if err != nil {
		panic(err)
	}
	return it
}

// CacheWrap cache wraps the underlying store.
func (dsa Store) CacheWrap() types.Store {
	return cache.New(dsa)
}

// Implements Store.
func (dsa Store) Write() {
	// CacheWrap().Write() gets called, but not dsa.Write().
	panic("unexpected .Write() on dbadapter.Store.")
}

// Implements Committer/CommitStore.
func (dsa Store) Commit() types.CommitID {
	// Always returns a zero commitID, as dbadapter store doesn't merkleize.
	return types.CommitID{
		Version: 0,
		Hash:    nil,
	}
}

// Implements Committer/CommitStore.
func (dsa Store) LastCommitID() types.CommitID {
	// Always returns a zero commitID, as dbadapter store doesn't merkleize.
	return types.CommitID{
		Version: 0,
		Hash:    nil,
	}
}

// Implements Committer/CommitStore.
func (dsa Store) GetStoreOptions() types.StoreOptions {
	return types.StoreOptions{}
}

// Implements Committer/CommitStore.
func (dsa Store) SetStoreOptions(types.StoreOptions) {
}

// Implements Committer/CommitStore.
func (dsa Store) LoadLatestVersion() error {
	return nil
}

// Implements Committer/CommitStore.
func (dsa Store) LoadVersion(ver int64) error {
	return nil
}

// dbm.DB implements Store.
var _ types.Store = Store{}
