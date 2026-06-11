//go:build cgo

package gnoland

// Register the cgo-only DB backends. Pure-Go builds (CGO_ENABLED=0, e.g.
// goreleaser) get only the pure backends, and the config default falls back
// to pebbledb accordingly.
import (
	_ "github.com/gnolang/gno/tm2/pkg/db/lmdbdb"
	_ "github.com/gnolang/gno/tm2/pkg/db/mdbxdb"
)
