// Package all imports all available databases. It is useful mostly in tests.
// The cgo-only backends (lmdbdb, mdbxdb) are registered via all_cgo.go.
package all

import (
	_ "github.com/gnolang/gno/tm2/pkg/db/boltdb"
	_ "github.com/gnolang/gno/tm2/pkg/db/goleveldb"
	_ "github.com/gnolang/gno/tm2/pkg/db/memdb"
	_ "github.com/gnolang/gno/tm2/pkg/db/pebbledb"
)
