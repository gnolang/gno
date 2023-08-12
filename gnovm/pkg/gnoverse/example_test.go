package gnoverse_test

import (
	"fmt"

	"github.com/gnolang/gno/gnovm/pkg/gnoverse"
	"github.com/gnolang/gno/tm2/pkg/db"
)

func Example() {
	// configure a new Sandbox.
	localDB, _ := db.NewDB("gnolang", db.MemDBBackend, "datadir") // make persistent with db.GoLevelDBBackend
	s := gnoverse.Sandbox{
		DB: localDB,
	}

	// initialize sandbox' components.
	_ = s.Init()

	// interact with the sandbox.
	// TODO: transactions, etc.

	// print state
	fmt.Println(s)
	fmt.Println("Done.")

	// Output:
	// Done.
}

func ExempleNewTestingSandbox() {
	// create and initialize a full memory-based sandbox.
	s := gnoverse.NewTestingSandbox()

	// interact...
	_ = s
}
