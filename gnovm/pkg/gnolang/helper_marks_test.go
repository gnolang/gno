package gnolang

import (
	"fmt"
	"io"
	"runtime"
	"sort"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	storetypes "github.com/gnolang/gno/tm2/pkg/store/types"
	"github.com/stretchr/testify/require"
)

const (
	marksLibPath   = "gno.land/p/demo/tests/marklib"
	marksRealmPath = "gno.land/r/tests/vm/markr"
)

const marksLibSrc = `package marklib

type Holder struct{ F func() }

func (h *Holder) Fire() { h.F() }

func Run(f func()) { f() }

func Store(h *Holder, f func()) { h.F = f }

func RunPanic(f func()) {
	f()
	panic("boom")
}
`

const marksRealmSrc = `package markr

import "gno.land/p/demo/tests/marklib"

var count int

func Count() int { return count }

func Helper() {
	f := func() { count++ }
	f()
}

func Lend() { marklib.Run(func() { count++ }) }

func Loop(n int) {
	for i := 0; i < n; i++ {
		marklib.Run(func() { count++ })
	}
}

func Rec(n int) {
	if n == 0 {
		return
	}
	marklib.Run(func() { count++ })
	Rec(n - 1)
}

func PanicOnce() {
	defer func() { recover() }()
	marklib.RunPanic(func() { count++ })
}

func PanicLoop(n int) {
	for i := 0; i < n; i++ {
		PanicOnce()
	}
}
`

func marksDeploy(t *testing.T, store Store) {
	t.Helper()
	for _, p := range []struct{ name, path, src string }{
		{"marklib", marksLibPath, marksLibSrc},
		{"markr", marksRealmPath, marksRealmSrc},
	} {
		m := NewMachineWithOptions(MachineOptions{Store: store, Output: io.Discard})
		m.RunMemPackage(&std.MemPackage{
			Type:  MPUserProd,
			Name:  p.name,
			Path:  p.path,
			Files: []*std.MemFile{{Name: p.name + ".gno", Body: p.src}},
		}, true)
		m.Release()
	}
}

// marksMarks lists the load-time marks on pkgPath's nodes, keyed by position.
func marksMarks(store Store, pkgPath string) []string {
	pn := store.GetPackageNode(pkgPath)
	var out []string
	for _, fn := range pn.FileSet.Files {
		Transcribe(fn, func(ns []Node, ftype TransField, index int, n Node, stage TransStage) (Node, TransCtrl) {
			if stage != TRANS_ENTER {
				return n, TRANS_CONTINUE
			}
			if v := n.GetAttribute(ATTR_LOCAL_HELPER); v != nil {
				out = append(out, fmt.Sprintf("%s %s LOCAL_HELPER=%v", fn.FileName, n.GetSpan(), v))
			}
			if v := n.GetAttribute(ATTR_NONESC_PARAMS); v != nil {
				out = append(out, fmt.Sprintf("%s %s NONESC=%v", fn.FileName, n.GetSpan(), v))
			}
			return n, TRANS_CONTINUE
		})
	}
	sort.Strings(out)
	return out
}

// After a restart the nodes come only from re-preprocessing the stored
// source: no node is read back from the database, and the marks match.
func TestMarksSurviveRestart(t *testing.T) {
	db := memdb.NewMemDB()
	backend := dbadapter.StoreConstructor(db, storetypes.StoreOptions{})
	s1 := NewStore(nil, backend, backend)
	marksDeploy(t, s1)
	before := map[string][]string{}
	for _, p := range []string{marksLibPath, marksRealmPath} {
		before[p] = marksMarks(s1, p)
		t.Logf("%s marks: %v", p, before[p])
	}
	require.Contains(t, fmt.Sprint(before[marksRealmPath]), "LOCAL_HELPER=true")
	require.Contains(t, fmt.Sprint(before[marksLibPath]), "NONESC=[true]")

	s2 := NewStore(nil, backend, backend)
	require.Nil(t, s2.GetBlockNodeSafe(PackageNodeLocation(marksRealmPath)),
		"a node was read back from the database")
	m := NewMachineWithOptions(MachineOptions{Store: s2, Output: io.Discard})
	m.PreprocessAllFilesAndSaveBlockNodes()
	m.Release()
	for _, p := range []string{marksLibPath, marksRealmPath} {
		require.Equal(t, before[p], marksMarks(s2, p), p)
	}
}

// marksCall runs markr.fn(args) in a fresh machine and returns the number of
// loans left in m.lends when the call ends, and the count it read.
func marksCall(t *testing.T, store Store, fn string, args ...any) (lends, live int) {
	t.Helper()
	m := NewMachineWithOptions(MachineOptions{PkgPath: marksRealmPath, Store: store, Output: io.Discard})
	defer m.Release()
	m.Eval(Call(X(fn), args...))
	// Go heap the loans keep alive (map, tags, and the literals they key).
	if n := len(m.lends); n >= 100000 {
		var a, b runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&a)
		m.lends = nil
		runtime.GC()
		runtime.ReadMemStats(&b)
		t.Logf("%s: dropping %d loans freed %d heap bytes", fn, n, int64(a.HeapAlloc)-int64(b.HeapAlloc))
		return n, 0
	}
	for fv := range m.lends {
		if m.lendLive(fv) {
			live++
		}
	}
	return len(m.lends), live
}

func marksCount(t *testing.T, store Store) int64 {
	t.Helper()
	m := NewMachineWithOptions(MachineOptions{PkgPath: marksRealmPath, Store: store, Output: io.Discard})
	defer m.Release()
	return m.Eval(Call(X("Count")))[0].GetInt()
}

// A loan ends when its frame returns or panics, so m.lends is
// empty once the call is over.
func TestLendsGrowth(t *testing.T) {
	db := memdb.NewMemDB()
	backend := dbadapter.StoreConstructor(db, storetypes.StoreOptions{})
	store := NewStore(nil, backend, backend)
	marksDeploy(t, store)

	c0 := marksCount(t, store)
	for _, tc := range []struct {
		fn   string
		n    int
		want int
	}{
		{"Lend", 0, 0},
		{"Loop", 1000, 0},
		{"Loop", 100000, 0},
		{"Rec", 200, 0},
		{"PanicLoop", 1000, 0},
	} {
		var lends, live int
		if tc.fn == "Lend" {
			lends, live = marksCall(t, store, tc.fn)
		} else {
			lends, live = marksCall(t, store, tc.fn, tc.n)
		}
		t.Logf("%s(%d): len(m.lends)=%d live=%d", tc.fn, tc.n, lends, live)
		require.Equal(t, tc.want, lends, tc.fn)
		require.Zero(t, live, tc.fn)
	}
	// Gas a loan costs today: one closure and one Run call per iteration.
	for _, n := range []int{1000, 2000} {
		gm := storetypes.NewInfiniteGasMeter()
		m := NewMachineWithOptions(MachineOptions{PkgPath: marksRealmPath, Store: store, Output: io.Discard, GasMeter: gm})
		m.Eval(Call(X("Loop"), n))
		t.Logf("Loop(%d): gas %d, loans %d", n, gm.GasConsumed(), len(m.lends))
		m.Release()
	}
	c0 += 3000
	// Each lent literal wrote: the loans were live while their call ran.
	require.Equal(t, c0+1+1000+100000+200+1000, marksCount(t, store))
}

// The controls hang on the marks: with them stripped after a restart, the
// same helper and lent literal are refused and count stays put.
func TestControlsNeedMarks(t *testing.T) {
	db := memdb.NewMemDB()
	backend := dbadapter.StoreConstructor(db, storetypes.StoreOptions{})
	marksDeploy(t, NewStore(nil, backend, backend))

	for _, strip := range []bool{false, true} {
		s := NewStore(nil, backend, backend)
		m := NewMachineWithOptions(MachineOptions{Store: s, Output: io.Discard})
		m.PreprocessAllFilesAndSaveBlockNodes()
		m.Release()
		if strip {
			for _, p := range []string{marksLibPath, marksRealmPath} {
				for _, fn := range s.GetPackageNode(p).FileSet.Files {
					Transcribe(fn, func(ns []Node, ftype TransField, index int, n Node, stage TransStage) (Node, TransCtrl) {
						if stage == TRANS_ENTER {
							n.SetAttribute(ATTR_LOCAL_HELPER, nil)
							n.SetAttribute(ATTR_NONESC_PARAMS, nil)
						}
						return n, TRANS_CONTINUE
					})
				}
			}
		}
		for _, fn := range []string{"Helper", "Lend"} {
			c0 := marksCount(t, s)
			var r any
			func() {
				defer func() { r = recover() }()
				marksCall(t, s, fn)
			}()
			c1 := marksCount(t, s)
			t.Logf("strip=%v %s: count %d -> %d, panic: %.120v", strip, fn, c0, c1, r)
			if strip {
				require.Contains(t, fmt.Sprint(r), "read-only", fn)
				require.Equal(t, c0, c1, fn)
			} else {
				require.Nil(t, r, fn)
				require.Equal(t, c0+1, c1, fn)
			}
		}
	}
}
