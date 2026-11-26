package gnolang

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	storetypes "github.com/gnolang/gno/tm2/pkg/store/types"
)

// Benchmarks of what the loan list (Machine.lends) costs per call: each runs
// one Gno program in the VM and reports ns, allocs and gas per call. The
// program text uses no loan internals, so it also runs on a build without them.

const (
	lendBenchAvlPath   = "gno.land/p/nt/avl/v0"
	lendBenchLibPath   = "gno.land/p/demo/tests/lendlib"
	lendBenchRealmPath = "gno.land/r/tests/vm/lendr"
	lendBenchCallPath  = "gno.land/r/tests/vm/lendc"
)

const lendBenchLibSrc = `package lendlib

type holder struct{ f func() }

func Noop(x int) int { return x + 1 }

func Run(f func()) { f() }

func First(xs []int) int { return xs[0] }

func Run4(a, b, c, d func()) {
	a()
	b()
	c()
	d()
}

func RunN(n int, f func()) {
	for i := 0; i < n; i++ {
		f()
	}
}

func keep(f func()) {
	var h holder
	h.f = f
	_ = h
}

func keepAll(fs ...func()) { _ = fs }

// HandOn, InLit and InSlice each receive a lent f and let it escape.
func HandOn(f func()) { keep(f) }

func InLit(f func()) { keep(func() { f() }) }

func InSlice(f func()) { keepAll(f) }
`

const lendBenchRealmSrc = `package lendr

import (
	"gno.land/p/demo/tests/lendlib"
	"gno.land/p/nt/avl/v0"
)

var (
	count int
	sum   int
	tree  avl.Tree
)

func init() {
	for i := 0; i < 1000; i++ {
		k := string([]byte{byte('a' + i/100), byte('a' + i/10%10), byte('a' + i%10)})
		tree.Set(k, i)
	}
}

func named() { _ = count }

func Plain(n int) {
	for i := 0; i < n; i++ {
		_ = lendlib.Noop(i)
	}
}

func Lit1(n int) {
	for i := 0; i < n; i++ {
		lendlib.Run(func() { _ = count })
	}
}

func Lit4(n int) {
	for i := 0; i < n; i++ {
		lendlib.Run4(func() { _ = count }, func() { _ = count }, func() { _ = count }, func() { _ = count })
	}
}

func Named(n int) {
	for i := 0; i < n; i++ {
		lendlib.Run(named)
	}
}

func CallN(n, k int) {
	for i := 0; i < n; i++ {
		lendlib.RunN(k, func() { _ = count })
	}
}

// nest holds one live loan per level; the bottom calls a lent callback 10 times.
func nest(d int) {
	if d == 0 {
		lendlib.RunN(10, func() { _ = count })
		return
	}
	lendlib.Run(func() { nest(d - 1) })
}

func Nest(n, d int) {
	for i := 0; i < n; i++ {
		nest(d)
	}
}

func Drop(n int, how string) {
	for i := 0; i < n; i++ {
		switch how {
		case "handon":
			lendlib.HandOn(func() { _ = count })
		case "lit":
			lendlib.InLit(func() { _ = count })
		case "slice":
			lendlib.InSlice(func() { _ = count })
		}
	}
}

// SliceArg passes an l-element slice n times while a loan is live.
func SliceArg(n, l int) {
	xs := make([]int, l)
	lendlib.Run(func() {
		for i := 0; i < n; i++ {
			_ = lendlib.First(xs)
		}
	})
}

func IterRead(cur realm) int {
	tree.Iterate("", "", func(k string, v any) bool { return v.(int) < 0 })
	return count
}

func IterWrite(cur realm) int {
	tree.Iterate("", "", func(k string, v any) bool {
		sum += v.(int)
		return false
	})
	return sum
}
`

const lendBenchCallSrc = `package lendc

import "gno.land/r/tests/vm/lendr"

func Iter(cur realm, write bool) int {
	if write {
		return lendr.IterWrite(cross(cur))
	}
	return lendr.IterRead(cross(cur))
}
`

// lendBenchStore deploys avl (read from examples/), lendlib, lendr and lendc.
func lendBenchStore(tb testing.TB) Store {
	tb.Helper()
	db := memdb.NewMemDB()
	backend := dbadapter.StoreConstructor(db, storetypes.StoreOptions{})
	store := NewStore(nil, backend, backend)
	avlDir := filepath.Join("..", "..", "..", "examples", "gno.land", "p", "nt", "avl", "v0")
	var avlFiles []*std.MemFile
	for _, name := range []string{"node.gno", "tree.gno"} {
		body, err := os.ReadFile(filepath.Join(avlDir, name))
		if err != nil {
			tb.Fatal(err)
		}
		avlFiles = append(avlFiles, &std.MemFile{Name: name, Body: string(body)})
	}
	for _, p := range []struct {
		name, path string
		files      []*std.MemFile
	}{
		{"avl", lendBenchAvlPath, avlFiles},
		{"lendlib", lendBenchLibPath, []*std.MemFile{{Name: "lendlib.gno", Body: lendBenchLibSrc}}},
		{"lendr", lendBenchRealmPath, []*std.MemFile{{Name: "lendr.gno", Body: lendBenchRealmSrc}}},
		{"lendc", lendBenchCallPath, []*std.MemFile{{Name: "lendc.gno", Body: lendBenchCallSrc}}},
	} {
		m := NewMachineWithOptions(MachineOptions{Store: store, Output: io.Discard})
		m.RunMemPackage(&std.MemPackage{Type: MPUserProd, Name: p.name, Path: p.path, Files: p.files}, true)
		m.Release()
	}
	return store
}

// lendBenchEval runs mk() in pkgPath b.N times and reports ns and gas per call,
// calls being the number of measured calls one mk() makes.
func lendBenchEval(b *testing.B, pkgPath string, calls int, mk func() Expr) {
	b.Helper()
	store := lendBenchStore(b)
	gm := storetypes.NewInfiniteGasMeter()
	m := NewMachineWithOptions(MachineOptions{PkgPath: pkgPath, Store: store, Output: io.Discard, GasMeter: gm})
	defer m.Release()
	m.Eval(mk()) // warm the store cache; also checks the program runs
	var gas int64 = -1
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		g0 := gm.GasConsumed()
		m.Eval(mk())
		if g := int64(gm.GasConsumed() - g0); gas < 0 {
			gas = g
		} else if g != gas {
			b.Fatalf("gas differs between runs: %d then %d", gas, g)
		}
	}
	b.StopTimer()
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*calls), "ns/call")
	b.ReportMetric(float64(gas)/float64(calls), "gas/call")
}

const lendBenchN = 1000

// a. A call that lends nothing.
func BenchmarkLendPlain(b *testing.B) {
	lendBenchEval(b, lendBenchRealmPath, lendBenchN, func() Expr { return Call(X("Plain"), lendBenchN) })
}

// b. One and four literals written in place.
func BenchmarkLendLit1(b *testing.B) {
	lendBenchEval(b, lendBenchRealmPath, lendBenchN, func() Expr { return Call(X("Lit1"), lendBenchN) })
}

func BenchmarkLendLit4(b *testing.B) {
	lendBenchEval(b, lendBenchRealmPath, lendBenchN, func() Expr { return Call(X("Lit4"), lendBenchN) })
}

// c. A named function as argument (never lent).
func BenchmarkLendNamed(b *testing.B) {
	lendBenchEval(b, lendBenchRealmPath, lendBenchN, func() Expr { return Call(X("Named"), lendBenchN) })
}

// d. The lent callback called k times by the callee; a call is one RunN.
func BenchmarkLendCallback(b *testing.B) {
	for _, k := range []int{1, 10, 100} {
		b.Run(fmt.Sprintf("k=%d", k), func(b *testing.B) {
			lendBenchEval(b, lendBenchRealmPath, 100, func() Expr { return Call(X("CallN"), 100, k) })
		})
	}
}

// e. Calls nested d deep, one live loan per level; a call is one nest(d).
func BenchmarkLendNest(b *testing.B) {
	for _, d := range []int{1, 8, 64} {
		b.Run(fmt.Sprintf("d=%d", d), func(b *testing.B) {
			lendBenchEval(b, lendBenchRealmPath, 100, func() Expr { return Call(X("Nest"), 100, d) })
		})
	}
}

// f. The callee lets its lent callback escape: handed on, captured, in a slice.
func BenchmarkLendDrop(b *testing.B) {
	for _, how := range []string{"handon", "lit", "slice"} {
		b.Run(how, func(b *testing.B) {
			lendBenchEval(b, lendBenchRealmPath, lendBenchN, func() Expr { return Call(X("Drop"), lendBenchN, Str(how)) })
		})
	}
}

// A slice argument passed while a loan is live, by slice length.
func BenchmarkLendSliceArg(b *testing.B) {
	for _, l := range []int{10, 1000, 100000} {
		b.Run(fmt.Sprintf("len=%d", l), func(b *testing.B) {
			lendBenchEval(b, lendBenchRealmPath, lendBenchN, func() Expr { return Call(X("SliceArg"), lendBenchN, l) })
		})
	}
}

// g. avl Iterate over 1000 items as a crossing call from another realm; a
// call is one callback. "write" writes the realm and needs a live loan.
func BenchmarkLendAvl(b *testing.B) {
	for _, write := range []bool{false, true} {
		name := "read"
		if write {
			name = "write"
		}
		b.Run(name, func(b *testing.B) {
			lendBenchEval(b, lendBenchCallPath, 1000, func() Expr { return Call(X("Iter"), Nx(".origin"), X(fmt.Sprint(write))) })
		})
	}
}
