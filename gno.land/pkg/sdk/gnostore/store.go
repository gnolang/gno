// Package gnostore implements a tm2 store which can interoperate with the GnoVM's
// own store.
package gnostore

import (
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/store"
	"github.com/gnolang/gno/tm2/pkg/store/cache"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/gnolang/gno/tm2/pkg/store/iavl"
	"github.com/gnolang/gno/tm2/pkg/store/types"
)

// StoreConstructor implements store.CommitStoreConstructor.
// It can be used in conjunction with CommitMultiStore.MountStoreWithDB.
// initialize should only contain basic setter for the immutable config
// (like SetNativeStore); it should not initialize packages.
func StoreConstructor(alloc *gnolang.Allocator, initialize func(gs gnolang.Store)) store.CommitStoreConstructor {
	return func(db dbm.DB, opts types.StoreOptions) types.CommitStore {
		iavlStore := iavl.StoreConstructor(db, opts)
		base := dbadapter.StoreConstructor(db, opts)
		gno := gnolang.NewStore(alloc, base, iavlStore)
		initialize(gno)
		return &Store{
			Store: iavlStore.(*iavl.Store),
			opts:  opts,
			base:  base.(dbadapter.Store),
			gno:   gno,
		}
	}
}

func GetGnoStore(s types.Store) gnolang.Store {
	gs, ok := s.(interface {
		GnoStore() gnolang.Store
	})
	if ok {
		return gs.GnoStore()
	}
	return nil
}

type Store struct {
	*iavl.Store // iavl

	opts types.StoreOptions
	base dbadapter.Store
	gno  gnolang.Store
}

func (s *Store) GetStoreOptions() types.StoreOptions { return s.opts }

func (s *Store) SetStoreOptions(opts2 types.StoreOptions) {
	s.opts = opts2
	s.Store.SetStoreOptions(opts2)
}

func (s *Store) GnoStore() gnolang.Store { return s.gno }

type cacheStore struct {
	types.Store

	base    types.Store
	gno     gnolang.TransactionStore
	rootGno gnolang.Store
}

func (s *Store) CacheWrap() types.Store {
	s2 := &cacheStore{
		Store:   cache.New(s.Store),
		base:    cache.New(s.base),
		rootGno: s.gno,
	}
	s2.gno = s.gno.BeginTransaction(s2.base, s2.Store)
	return s2
}

func (store *cacheStore) Write() {
	store.Store.Write()
	store.base.Write()
	store.gno.Write()
}

func (store *cacheStore) Flush() {
	store.Store.(types.Flusher).Flush()
	store.base.(types.Flusher).Flush()
	store.gno.Write()
}

func (s *cacheStore) CacheWrap() types.Store {
	s2 := &cacheStore{
		Store:   cache.New(s.Store),
		base:    cache.New(s.base),
		rootGno: s.rootGno,
	}
	s2.gno = s.rootGno.BeginTransaction(s2.base, s2.Store)
	return s2
}

func (s *cacheStore) GnoStore() gnolang.Store { return s.gno }
