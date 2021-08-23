package dbadapter

import (
	dbm "github.com/gnolang/gno/pkgs/db"

	"github.com/gnolang/gno/pkgs/store/cache"
	"github.com/gnolang/gno/pkgs/store/types"
)

// Wrapper type for dbm.Db with implementation of Store
type Store struct {
	dbm.DB
}

// CacheWrap cache wraps the underlying store.
func (dsa Store) CacheWrap() types.Store {
	return cache.New(dsa)
}

func (dsa Store) Write() {
	panic("unexpected .Write() on dbadapter.Store.")
}

// dbm.DB implements Store.
var _ types.Store = Store{}
