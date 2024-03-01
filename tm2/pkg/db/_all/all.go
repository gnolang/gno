// Package all imports all available databases. It is useful mostly in tests.
//
// cgo databases (rocksdb, cleveldb) will be excluded if CGO_ENABLED=0.
package all

import (
	// Keep in sync with list of non-cgo backends.
	// Add cgo backends in all_cgo.go.
	_ "github.com/gnolang/gno/tm2/pkg/db/boltdb"
	_ "github.com/gnolang/gno/tm2/pkg/db/fsdb"
	_ "github.com/gnolang/gno/tm2/pkg/db/goleveldb"
	_ "github.com/gnolang/gno/tm2/pkg/db/memdb"
)
