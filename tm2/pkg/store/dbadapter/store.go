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
	dbm.DB
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
