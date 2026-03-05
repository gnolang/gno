package gnolang_test

import (
	"bytes"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/gnolang/gno/tm2/pkg/store/iavl"
	stypes "github.com/gnolang/gno/tm2/pkg/store/types"
)

func TestInitOrderDeterminism(t *testing.T) {
	// This test verifies that package-level variable initialization order
	// is deterministic even when variables have multiple dependencies.
	// Non-deterministic initialization would cause consensus failure in
	// a blockchain context because different validators would compute
	// different on-chain state.
	//
	// It also verifies that each initializer runs exactly once (matching
	// Go specification behavior), not double-initialized when a variable
	// is first resolved as a dependency and then encountered again in
	// declaration order.
	code := `
package main

var events []string

func emit(s string) string {
    events = append(events, s)
    return s
}

var (
    Z = A + "-" + B + "-" + C + "-" + D + "-" + E
    A = emit("A_INITIALIZED")
    B = emit("B_INITIALIZED")
    C = emit("C_INITIALIZED")
    D = emit("D_INITIALIZED")
    E = emit("E_INITIALIZED")
)

func main() {
    for _, e := range events {
        println(e)
    }
}
`
	// Expected output matches Go behavior: each emit called exactly once,
	// in dependency-resolution order (alphabetical by sorted dep names).
	expected := "A_INITIALIZED\nB_INITIALIZED\nC_INITIALIZED\nD_INITIALIZED\nE_INITIALIZED\n"

	eventLogs := make(map[string]int)

	for i := 0; i < 100; i++ {
		db := memdb.NewMemDB()
		baseStore := dbadapter.StoreConstructor(db, stypes.StoreOptions{})
		iavlStore := iavl.StoreConstructor(db, stypes.StoreOptions{})
		store := gnolang.NewStore(nil, baseStore, iavlStore)

		alloc := gnolang.NewAllocator(10 * 1024 * 1024)
		output := new(bytes.Buffer)

		m := gnolang.NewMachineWithOptions(gnolang.MachineOptions{
			PkgPath: "main",
			Store:   store,
			Alloc:   alloc,
			Output:  output,
		})

		n := m.MustParseFile("main.gno", code)
		m.RunFiles(n)
		m.RunMain()

		eventLog := output.String()
		eventLogs[eventLog]++
	}

	t.Logf("Observed %d distinct event orderings:", len(eventLogs))
	for log, count := range eventLogs {
		t.Logf("  %d times:\n%s", count, log)
	}

	if len(eventLogs) > 1 {
		t.Fatalf("Non-deterministic init order: found %d distinct states. "+
			"This would cause consensus failure (different AppHash on different validators).",
			len(eventLogs))
	}

	// Verify each initializer runs exactly once, matching Go behavior.
	for log := range eventLogs {
		if log != expected {
			t.Fatalf("Unexpected init order.\nExpected:\n%s\nGot:\n%s", expected, log)
		}
	}
}
