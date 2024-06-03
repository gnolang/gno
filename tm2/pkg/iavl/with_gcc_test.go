//go:build gcc

// This file exists because some of the DBs e.g CLevelDB
// require gcc as the compiler before they can ran otherwise
// we'll encounter crashes such as in https://github.com/tendermint/merkleeyes/issues/39

package iavl

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db"
	_ "github.com/gnolang/gno/tm2/pkg/db/cleveldb"
)

func BenchmarkImmutableAvlTreeCLevelDB(b *testing.B) {
	db := db.NewDB("test", db.CLevelDBBackendStr, "./")
	benchmarkImmutableAvlTreeWithDB(b, db)
}
