//go:build cgo

package benchmarks

// cgo-only db backends, registered (via init) so -disk-backend can select them.
// Behind the cgo build tag so the benchmarks package still compiles with
// CGO_ENABLED=0 — there, -disk-backend=lmdbdb just reports "unknown db_backend".
import (
	_ "github.com/gnolang/gno/tm2/pkg/db/lmdbdb" // -disk-backend=lmdbdb
	_ "github.com/gnolang/gno/tm2/pkg/db/mdbxdb" // -disk-backend=mdbxdb
)
