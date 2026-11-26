package gnolang

import (
	"io"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	storetypes "github.com/gnolang/gno/tm2/pkg/store/types"
	"github.com/stretchr/testify/require"
)

const loansRealmPath = "gno.land/r/tests/vm/loanr"

const loansRealmSrc = `package loanr

import "gno.land/p/demo/tests/marklib"

var count int

func Count() int { return count }

func Loop(n int) {
	for i := 0; i < n; i++ {
		marklib.Run(func() { count++ })
	}
}

// Nest lends n literals at once: each one's call is still running.
func Nest(n int) {
	if n == 0 {
		return
	}
	marklib.Run(func() {
		count++
		Nest(n - 1)
	})
}

// NestPanic panics at the deepest lending call; the top recovers.
func NestPanic(n int) {
	defer func() { recover() }()
	nestPanic(n)
}

func nestPanic(n int) {
	if n == 0 {
		panic("deep")
	}
	marklib.Run(func() {
		count++
		nestPanic(n - 1)
	})
}

func PanicLoop(n int) {
	for i := 0; i < n; i++ {
		NestPanic(3)
	}
}
`

// liveLendBound is the live lending frames times their arguments.
func liveLendBound(m *Machine) int {
	n := 0
	for i := range m.Frames {
		if m.Frames[i].lendTok != nil {
			n += m.Frames[i].NumArgs
		}
	}
	return n
}

// A loan dies with the frame it was lent to (return, recovered panic), so
// m.lends never holds more than the live lending frames' arguments.
func TestLoansBounded(t *testing.T) {
	db := memdb.NewMemDB()
	backend := dbadapter.StoreConstructor(db, storetypes.StoreOptions{})
	store := NewStore(nil, backend, backend)
	marksDeploy(t, store)
	m0 := NewMachineWithOptions(MachineOptions{Store: store, Output: io.Discard})
	m0.RunMemPackage(&std.MemPackage{
		Type: MPUserProd, Name: "loanr", Path: loansRealmPath,
		Files: []*std.MemFile{{Name: "loanr.gno", Body: loansRealmSrc}},
	}, true)
	m0.Release()
	count := func() int64 {
		m := NewMachineWithOptions(MachineOptions{PkgPath: loansRealmPath, Store: store, Output: io.Discard})
		defer m.Release()
		return m.Eval(Call(X("Count")))[0].GetInt()
	}

	var maxLen, checks int
	lendArgsHook = func(m *Machine) {
		checks++
		if b := liveLendBound(m); len(m.lends) > b {
			t.Fatalf("len(m.lends)=%d > live bound %d", len(m.lends), b)
		}
		maxLen = max(maxLen, len(m.lends))
	}
	defer func() { lendArgsHook = nil }()

	c0 := count()
	for _, tc := range []struct {
		fn      string
		n       int
		wantMax int
	}{
		{"Loop", 10000, 1},
		{"Nest", 100, 100},      // every level's loan is live at the deepest call
		{"NestPanic", 100, 100}, // the panic unwinds all 100 lending frames
		{"PanicLoop", 1000, 3},
	} {
		maxLen, checks = 0, 0
		m := NewMachineWithOptions(MachineOptions{PkgPath: loansRealmPath, Store: store, Output: io.Discard})
		m.Eval(Call(X(tc.fn), tc.n))
		t.Logf("%s(%d): checks %d, max len(m.lends) %d, after %d", tc.fn, tc.n, checks, maxLen, len(m.lends))
		require.GreaterOrEqual(t, checks, tc.wantMax, tc.fn)
		require.Equal(t, tc.wantMax, maxLen, tc.fn)
		require.Zero(t, len(m.lends), tc.fn)
		m.Release()
	}
	// Every loan wrote while its call ran.
	require.Equal(t, c0+10000+100+100+3000, count())
}

const loansWalkRealmPath = "gno.land/r/tests/vm/loanwalk"

const loansWalkRealmSrc = `package loanwalk

var count int

func sink(x func()) {}
func call(g func()) { g() }

// Walk builds an n-deep chain of closures, then hands either the chain or a
// one-level closure to sink n times while a lent literal's loan is live.
func Walk(n, deep int) {
	shallow := func() {}
	prev := func() {}
	for i := 0; i < n; i++ {
		p := prev
		prev = func() { p() }
	}
	call(func() {
		for i := 0; i < n; i++ {
			if deep == 1 {
				sink(prev)
			} else {
				sink(shallow)
			}
		}
		count++
	})
}
`

// Following captures below the first level costs CPU per value visited, so
// a deep chain handed out repeatedly while a loan is live is paid for.
func TestLoansDeepWalkCharged(t *testing.T) {
	db := memdb.NewMemDB()
	backend := dbadapter.StoreConstructor(db, storetypes.StoreOptions{})
	store := NewStore(nil, backend, backend)
	m0 := NewMachineWithOptions(MachineOptions{Store: store, Output: io.Discard})
	m0.RunMemPackage(&std.MemPackage{
		Type: MPUserProd, Name: "loanwalk", Path: loansWalkRealmPath,
		Files: []*std.MemFile{{Name: "loanwalk.gno", Body: loansWalkRealmSrc}},
	}, true)
	m0.Release()
	cycles := func(n, deep int) int64 {
		m := NewMachineWithOptions(MachineOptions{PkgPath: loansWalkRealmPath, Store: store, Output: io.Discard})
		defer m.Release()
		m.Eval(Call(X("Walk"), n, deep))
		return m.Cycles
	}
	const n = 200
	shallow, deep := cycles(n, 0), cycles(n, 1)
	t.Logf("cycles: shallow %d, deep %d", shallow, deep)
	require.GreaterOrEqual(t, deep-shallow, int64(n*n/2))
}

const loansSliceRealmPath = "gno.land/r/tests/vm/loanslice"

const loansSliceRealmSrc = `package loanslice

var count int

func sinkv(xs ...func()) {}
func call(g func()) { g() }

// Spread hands an n-element slice of closures, or a one-element one, to a
// variadic parameter n times while a lent literal's loan is live.
func Spread(n, wide int) {
	xs := make([]func(), n)
	for i := range xs {
		xs[i] = func() {}
	}
	one := xs[:1]
	call(func() {
		for i := 0; i < n; i++ {
			if wide == 1 {
				sinkv(xs...)
			} else {
				sinkv(one...)
			}
		}
		count++
	})
}
`

// Walking a variadic slice's elements while a loan is live costs CPU per
// element, so a wide slice handed out repeatedly is paid for.
func TestLoansSliceWalkCharged(t *testing.T) {
	db := memdb.NewMemDB()
	backend := dbadapter.StoreConstructor(db, storetypes.StoreOptions{})
	store := NewStore(nil, backend, backend)
	m0 := NewMachineWithOptions(MachineOptions{Store: store, Output: io.Discard})
	m0.RunMemPackage(&std.MemPackage{
		Type: MPUserProd, Name: "loanslice", Path: loansSliceRealmPath,
		Files: []*std.MemFile{{Name: "loanslice.gno", Body: loansSliceRealmSrc}},
	}, true)
	m0.Release()
	cycles := func(n, wide int) int64 {
		m := NewMachineWithOptions(MachineOptions{PkgPath: loansSliceRealmPath, Store: store, Output: io.Discard})
		defer m.Release()
		m.Eval(Call(X("Spread"), n, wide))
		return m.Cycles
	}
	const n = 200
	narrow, wide := cycles(n, 0), cycles(n, 1)
	t.Logf("cycles: one element %d, %d elements %d", narrow, n, wide)
	require.GreaterOrEqual(t, wide-narrow, int64(n*(n-1)))
}
