package db

import (
	"fmt"
	"strings"
)

type BackendType string

// These are valid backend types.
const (
	// GoLevelDBBackend represents goleveldb (github.com/syndtr/goleveldb - most
	// popular implementation)
	//   - pure go
	//   - stable
	//   - not well maintained
	//   - we should move away from this
	GoLevelDBBackend BackendType = "goleveldb"
	// CLevelDBBackend represents cleveldb (uses levigo wrapper)
	//   - fast
	//   - requires gcc / cgo
	//   - use cleveldb build tag (go build -tags cleveldb)
	//   - wrapper is not well maintained
	//   - we should move away from this
	CLevelDBBackend BackendType = "cleveldb"
	// MemDBBackend represents in-memoty key value store, which is mostly used
	// for testing.
	MemDBBackend BackendType = "memdb"
	// FSDBBackend represents filesystem database
	//	 - EXPERIMENTAL
	//   - slow
	FSDBBackend BackendType = "fsdb"
	// BoltDBBackend represents bolt (uses etcd's fork of bolt -
	// go.etcd.io/bbolt)
	//   - EXPERIMENTAL
	//   - may be faster is some use-cases (random reads - indexer)
	//   - use boltdb build tag (go build -tags boltdb)
	//   - maintained
	//   - pure go
	//  We should keep this
	BoltDBBackend BackendType = "boltdb"
	// RocksDBBackend represents rocksdb (uses github.com/tecbot/gorocksdb)
	//   - EXPERIMENTAL
	//   - requires gcc / cgo
	//   - use gorocksdb build tag (go build -tags gorocksdb)
	//   - supports rocksdb v6 series but not further
	//   - zero maintenance upstream
	// If we support rocks, we should drop this
	RocksDBBackend BackendType = "gorocksdb"
	// GRocksDBBackend represents grocksdb (uses github.com/linxGnu/grocksdb)
	//   - EXPERIMENTAL
	//   - reguires gcc / cgo
	//   - use grocksdb build tag 
	//   - supports rocksdb v7 series and optimistic transaction dbs
	//   - well maintained
	// If we use rocks at all, this is the best choice for a wrapper
	GRocksDBBackend BackendType = "grocksdb"
	// BadgerDBBackend represents badgerdb (github.com/dgraph-io/badger)
	//  - EXPERIMENTAL
	//  - does not require gcc/cgo
	//  - good maitenance
	//  - use badgerdb build tag
	// Jacob's current favorite for default
	BadgerDBBackend BackendType = "badgerdb"
)

type dbCreator func(name string, dir string) (DB, error)

var backends = map[BackendType]dbCreator{}

func registerDBCreator(backend BackendType, creator dbCreator, force bool) {
	_, ok := backends[backend]
	if !force && ok {
		return
	}
	backends[backend] = creator
}

// NewDB creates a new database of type backend with the given name.
// NOTE: function panics if:
//   - backend is unknown (not registered)
//   - creator function, provided during registration, returns error
func NewDB(name string, backend BackendType, dir string) DB {
	dbCreator, ok := backends[backend]
	if !ok {
		keys := make([]string, len(backends))
		i := 0
		for k := range backends {
			keys[i] = string(k)
			i++
		}
		panic(fmt.Sprintf("Unknown db_backend %s, expected either %s", backend, strings.Join(keys, " or ")))
	}

	db, err := dbCreator(name, dir)
	if err != nil {
		panic(fmt.Sprintf("Error initializing DB: %v", err))
	}
	return db
}
