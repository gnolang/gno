package gnolang

import (
	"io"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	stypes "github.com/gnolang/gno/tm2/pkg/store/types"
)

// TestException_DeferRepanicChainSurvivesAbort drives a gno program with
// nested panic + defer-repanic through (*Machine).RunMain, recovers the
// resulting Go-level panic, and asserts the structural shape of the
// *Exception that escapes. This is the gap left by panic2b.gno: that
// filetest only checks the rendered Stacktrace/Error blobs; this one
// pins that the Previous chain, per-link Stacktrace, and GoStack all
// reach the outer recoverer intact (the pre-PR UnhandledPanicError
// dropped everything but the joined message).
func TestException_DeferRepanicChainSurvivesAbort(t *testing.T) {
	db := memdb.NewMemDB()
	tm2 := dbadapter.StoreConstructor(db, stypes.StoreOptions{})
	st := NewStore(nil, tm2, tm2)
	m := NewMachineWithOptions(MachineOptions{
		Store:  st,
		Output: io.Discard,
		Alloc:  NewAllocator(64 * 1024 * 1024),
	})
	defer m.Release()

	const source = `package chain

func main() {
	defer func() { panic("B") }()
	panic("A")
}
`
	fn := m.MustParseFile("chain.gno", source)
	pn := NewPackageNode("chain", "chain", &FileSet{})
	pv := pn.NewPackage(m.Alloc)
	m.Store.SetBlockNode(pn)
	m.Store.SetCachePackage(pv)
	m.SetActivePackage(pv)

	ex := runAndRecoverException(t, m, fn)

	// Terminal-abort state.
	if !ex.Abort {
		t.Fatal("expected Abort=true on recovered *Exception")
	}

	// Descriptor is the joined chain message (chronological order:
	// original panic first).
	if !strings.Contains(ex.Descriptor, "A") || !strings.Contains(ex.Descriptor, "B") {
		t.Fatalf("Descriptor missing chain values: %q", ex.Descriptor)
	}
	if strings.Index(ex.Descriptor, "A") > strings.Index(ex.Descriptor, "B") {
		t.Fatalf("Descriptor chain ordering wrong (expected A before B): %q", ex.Descriptor)
	}

	// Chain structure: head is the latest (defer-repanic = "B"),
	// Previous is the original ("A"). NumExceptions counts the
	// whole linked list.
	if ex.Previous == nil {
		t.Fatal("expected non-nil Previous (chain dropped on the floor?)")
	}
	if got := ex.NumExceptions(); got != 2 {
		t.Fatalf("NumExceptions = %d, want 2", got)
	}
	if got := ex.Value.String(); !strings.Contains(got, "B") {
		t.Fatalf("head Exception.Value = %q, want to contain B", got)
	}
	if got := ex.Previous.Value.String(); !strings.Contains(got, "A") {
		t.Fatalf("Previous.Value = %q, want to contain A", got)
	}

	// Per-Exception stacktraces survive — this is the bit
	// makeUnhandledPanicError discarded pre-PR. Each link captured
	// its frames at its own raise site.
	if ex.Stacktrace.IsZero() {
		t.Fatal("head Stacktrace is zero")
	}
	if ex.Previous.Stacktrace.IsZero() {
		t.Fatal("Previous Stacktrace is zero (per-link stack lost)")
	}

	// GoStack — the new field, captured by pushPanic via
	// captureExceptionStack at the raise site. Must be non-empty
	// and must NOT contain "captureExceptionStack" itself (would
	// indicate the skip count is wrong).
	if ex.GoStack == "" {
		t.Fatal("GoStack is empty on recovered *Exception")
	}
	if strings.Contains(ex.GoStack, "captureExceptionStack\n") {
		t.Fatalf("GoStack leaks captureExceptionStack frame (skip off-by-one):\n%s", ex.GoStack)
	}
}

// TestException_MixedOriginGoStack pins the per-link GoStack value of
// #5681's constructor pattern. Head and Previous come from *different*
// VM helpers — head from the nil-deref path (op_expressions.go), Previous
// from the panic-builtin native body (uverse.go). A flag-gated design
// (#5670) would set head.GoStack="" in this scenario, losing the cleanup
// blow-up site.
//
// Concrete shape for the scenario below:
//
//   HEAD (defer-repanic: nil deref)
//     Value: ("runtime error: nil pointer dereference" string)
//     GoStack:
//       (*Machine).doOpStar           op_expressions.go:163
//       (*Machine).runOnce            machine.go
//       (*Machine).Run                machine.go
//
//   Previous (user panic "A")
//     Value: ("A" string)
//     GoStack:
//       makeUverseNode.func11         uverse.go        ← panic builtin's native body
//       (*Machine).doOpCallNativeBody op_call.go
//       (*Machine).runOnce            machine.go
//       (*Machine).Run                machine.go
func TestException_MixedOriginGoStack(t *testing.T) {
	db := memdb.NewMemDB()
	tm2 := dbadapter.StoreConstructor(db, stypes.StoreOptions{})
	st := NewStore(nil, tm2, tm2)
	m := NewMachineWithOptions(MachineOptions{
		Store:  st,
		Output: io.Discard,
		Alloc:  NewAllocator(64 * 1024 * 1024),
	})
	defer m.Release()

	const source = `package mixed_origin

func main() {
	defer func() {
		var p *int
		_ = *p
	}()
	panic("A")
}
`
	fn := m.MustParseFile("mixed.gno", source)
	pn := NewPackageNode("mixed_origin", "mixed_origin", &FileSet{})
	pv := pn.NewPackage(m.Alloc)
	m.Store.SetBlockNode(pn)
	m.Store.SetCachePackage(pv)
	m.SetActivePackage(pv)

	ex := runAndRecoverException(t, m, fn)

	if ex.Previous == nil {
		t.Fatal("expected 2-link chain")
	}

	// Head = defer-repanic via *p (nil-deref) — raised from doOpStar
	// in op_expressions.go.
	if !strings.Contains(ex.GoStack, "op_expressions.go") {
		t.Fatalf("head GoStack missing op_expressions.go (defer-repanic site):\n%s", ex.GoStack)
	}

	// Previous = user panic("A") — raised from the panic builtin's
	// native body in uverse.go.
	if !strings.Contains(ex.Previous.GoStack, "uverse.go") {
		t.Fatalf("Previous GoStack missing uverse.go (user-panic site):\n%s", ex.Previous.GoStack)
	}

	// And the two must differ — the whole point is mixed origins.
	if ex.GoStack == ex.Previous.GoStack {
		t.Fatalf("head and Previous GoStack identical, expected mixed origins:\n%s", ex.GoStack)
	}
}

func runAndRecoverException(t *testing.T, m *Machine, fn *FileNode) (ex *Exception) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic from unhandled gno panic, got none")
		}
		var ok bool
		ex, ok = r.(*Exception)
		if !ok {
			t.Fatalf("recovered value is %T, want *Exception: %v", r, r)
		}
	}()
	m.RunFiles(fn)
	m.RunMain()
	return nil
}
