//go:build cgo

package gnoland

// Register the cgo-only DB backends as opt-in alternatives to the pebbledb
// default. Pure-Go builds (CGO_ENABLED=0, e.g. goreleaser) get only the pure
// backends.
import (
	_ "github.com/gnolang/gno/tm2/pkg/db/lmdbdb"
	_ "github.com/gnolang/gno/tm2/pkg/db/mdbxdb"
)
