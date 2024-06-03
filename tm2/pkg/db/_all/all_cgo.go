//go:build cgo

package all

import (
	// Keep in sync with list of cgo backends.
	// Add non-cgo backends in all.go.
	_ "github.com/gnolang/gno/tm2/pkg/db/cleveldb"
	_ "github.com/gnolang/gno/tm2/pkg/db/rocksdb"
)
