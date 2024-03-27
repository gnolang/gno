package db

import (
	"fmt"
	"slices"

	"golang.org/x/exp/maps"
)

type BackendType string

func (b BackendType) String() string {
	return string(b)
}

// These are valid backend types.
//
// The backends themselves must be imported to be used (ie. using the blank
// import, `import _ "github.com/gnolang/gno/tm2/pkg/db/goleveldb"`). To allow
// for end-user customization at build time, the package
// "github.com/gnolang/gno/tm2/pkg/db/_tags" can be imported -- this package
// will import each database depending on whether its build tag is provided.
//
// This can be used in conjunction with specific to provide defaults, for instance:
//
//	package main
//
//	import (
//		"github.com/gnolang/gno/tm2/pkg/db"
//		_ "github.com/gnolang/gno/tm2/pkg/db/_tags" // allow user to customize with build tags
//		_ "github.com/gnolang/gno/tm2/pkg/db/memdb" // always support memdb
//	)
//
//	func main() {
//		db.NewDB("mydb", db.BackendType(userProvidedBackend), "./data")
//	}
const (
	// GoLevelDBBackend represents goleveldb (github.com/gnolang/goleveldb - most
	// popular implementation)
	//   - stable
	GoLevelDBBackend BackendType = "goleveldb"
	// CLevelDBBackend represents cleveldb (uses levigo wrapper)
	//   - fast
	//   - requires gcc
	CLevelDBBackend BackendType = "cleveldb"
	// MemDBBackend represents in-memory key value store, which is mostly used
	// for testing.
	MemDBBackend BackendType = "memdb"
	// FSDBBackend represents filesystem database
	//   - EXPERIMENTAL
	//   - slow
	FSDBBackend BackendType = "fsdb"
	// BoltDBBackend represents bolt (uses etcd's fork of bolt -
	// go.etcd.io/bbolt)
	//   - EXPERIMENTAL
	//   - may be faster is some use-cases (random reads - indexer)
	BoltDBBackend BackendType = "boltdb"
	// RocksDBBackend represents rocksdb (uses github.com/linuxGnu/grocksdb)
	//   - EXPERIMENTAL
	//   - requires gcc
	RocksDBBackend BackendType = "rocksdb"
)

type dbCreator func(name string, dir string) (DB, error)

var backends = map[BackendType]dbCreator{}

// InternalRegisterDBCreator is used by the init functions of imported databases
// to register their own dbCreators.
//
// This function is not meant for usage outside of db/.
func InternalRegisterDBCreator(backend BackendType, creator dbCreator, force bool) {
	_, ok := backends[backend]
	if !force && ok {
		return
	}
	backends[backend] = creator
}

// BackendList returns a list of available db backends. The list is sorted.
func BackendList() []BackendType {
	keys := maps.Keys(backends)
	slices.Sort(keys)
	return keys
}

// NewDB creates a new database of type backend with the given name.
// NOTE: function panics if:
//   - backend is unknown (not registered)
//   - creator function, provided during registration, returns error
func NewDB(name string, backend BackendType, dir string) (DB, error) {
	dbCreator, ok := backends[backend]
	if !ok {
		keys := BackendList()
		return nil, fmt.Errorf("unknown db_backend %s. Expected one of %v", backend, keys)
	}

	db, err := dbCreator(name, dir)
	if err != nil {
		return nil, fmt.Errorf("error initializing DB: %w", err)
	}
	return db, nil
}
