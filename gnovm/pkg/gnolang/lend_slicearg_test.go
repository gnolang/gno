package gnolang

import (
	"io"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	storetypes "github.com/gnolang/gno/tm2/pkg/store/types"
	"github.com/stretchr/testify/require"
)

// A slice argument whose elements cannot be functions costs no loan work
// per element while a loan is live.
func TestLendSliceArgGas(t *testing.T) {
	store := lendBenchStore(t)
	const n = 100
	perCall := func(l int) int64 {
		gm := storetypes.NewInfiniteGasMeter()
		m := NewMachineWithOptions(MachineOptions{PkgPath: lendBenchRealmPath, Store: store, Output: io.Discard, GasMeter: gm})
		defer m.Release()
		m.Eval(Call(X("SliceArg"), n, l)) // warm the store cache
		g0 := gm.GasConsumed()
		m.Eval(Call(X("SliceArg"), n, l))
		return (gm.GasConsumed() - g0) / n
	}
	g10, g1000 := perCall(10), perCall(1000)
	require.Less(t, g1000-g10, int64(500), "gas per call: len=10 %d, len=1000 %d", g10, g1000)
}

const lendSliceLibPath = "gno.land/p/demo/tests/lendsl"

const lendSliceLibSrc = `package lendsl

type F func()

type I interface{ Call() }

func (f F) Call() { f() }

func keepFn(fs ...func()) { _ = fs }
func keepF(fs ...F)       { _ = fs }
func keepAny(xs ...any)   { _ = xs }
func keepI(xs ...I)       { _ = xs }

func keepSliceFn(fs []func()) { _ = fs }
func keepSliceF(fs []F)       { _ = fs }
func keepSliceAny(xs []any)   { _ = xs }
func keepSliceI(xs []I)       { _ = xs }

func Call(f func()) { f() }

// V*: the lent f goes into a variadic slot, then is called.
func VFn(f func())   { keepFn(f); f() }
func VF(f F)         { keepF(f); f() }
func VAny(f func())  { keepAny(f); f() }
func VI(f F)         { keepI(f); f() }
func VWrap(f func()) { keepFn(func() { f() }); f() }

// B*: f goes into a slice built before the call, then is called.
func BFn(f func())  { fs := []func(){f}; keepSliceFn(fs); f() }
func BF(f F)        { fs := []F{f}; keepSliceF(fs); f() }
func BAny(f func()) { xs := []any{f}; keepSliceAny(xs); f() }
func BI(f F)        { xs := []I{f}; keepSliceI(xs); f() }
`

const lendSliceRealmPath = "gno.land/r/tests/vm/lendsl"

const lendSliceRealmSrc = `package lendsl

import "gno.land/p/demo/tests/lendsl"

var count int

func Try(how string) (r string) {
	defer func() {
		if recover() != nil {
			r = "refused"
		}
	}()
	switch how {
	case "control":
		lendsl.Call(func() { count++ })
	case "VFn":
		lendsl.VFn(func() { count++ })
	case "VF":
		lendsl.VF(func() { count++ })
	case "VAny":
		lendsl.VAny(func() { count++ })
	case "VI":
		lendsl.VI(func() { count++ })
	case "VWrap":
		lendsl.VWrap(func() { count++ })
	case "BFn":
		lendsl.BFn(func() { count++ })
	case "BF":
		lendsl.BF(func() { count++ })
	case "BAny":
		lendsl.BAny(func() { count++ })
	case "BI":
		lendsl.BI(func() { count++ })
	default:
		panic("unknown case")
	}
	return "ok"
}

func Count() int { return count }
`

// A lent callback put in a slice argument of any element type that can hold
// it loses its loan in the callee, and its later write is refused.
func TestLendSliceArgStillDrops(t *testing.T) {
	db := memdb.NewMemDB()
	backend := dbadapter.StoreConstructor(db, storetypes.StoreOptions{})
	store := NewStore(nil, backend, backend)
	for _, p := range []struct{ name, path, src string }{
		{"lendsl", lendSliceLibPath, lendSliceLibSrc},
		{"lendsl", lendSliceRealmPath, lendSliceRealmSrc},
	} {
		m := NewMachineWithOptions(MachineOptions{Store: store, Output: io.Discard})
		m.RunMemPackage(&std.MemPackage{Type: MPUserProd, Name: p.name, Path: p.path, Files: []*std.MemFile{{Name: p.name + ".gno", Body: p.src}}}, true)
		m.Release()
	}
	eval := func(x Expr) *TypedValue {
		m := NewMachineWithOptions(MachineOptions{PkgPath: lendSliceRealmPath, Store: store, Output: io.Discard})
		defer m.Release()
		res := m.Eval(x)
		require.Empty(t, m.lends)
		return &res[0]
	}
	for _, tc := range []struct {
		how  string
		lent bool // the callback is lent when the callee starts
		want string
	}{
		{"control", true, "ok"},
		{"VFn", true, "refused"},
		{"VF", true, "refused"},
		{"VAny", true, "refused"},
		{"VI", true, "refused"},
		{"VWrap", true, "refused"},
		{"BFn", false, "refused"},
		{"BF", false, "refused"},
		{"BAny", false, "refused"},
		{"BI", false, "refused"},
	} {
		var peak int
		var inKeep []int
		defer func() { lendArgsHook = nil }()
		lendArgsHook = func(m *Machine) {
			peak = max(peak, len(m.lends))
			if fv := m.Frames[len(m.Frames)-1].Func; fv != nil && strings.HasPrefix(string(fv.Name), "keep") {
				inKeep = append(inKeep, len(m.lends))
			}
		}
		before := eval(Call(X("Count"))).GetInt()
		got := eval(Call(X("Try"), Str(tc.how))).GetString()
		lendArgsHook = nil
		after := eval(Call(X("Count"))).GetInt()
		require.Equal(t, tc.want, got, tc.how)
		if tc.want == "ok" {
			require.Equal(t, before+1, after, tc.how)
		} else {
			require.Equal(t, before, after, "%s: a refused write landed", tc.how)
			require.Len(t, inKeep, 1, tc.how)
			require.Zero(t, inKeep[0], "%s: a loan outlived its escape", tc.how)
		}
		require.Equal(t, tc.lent, peak > 0, "%s: lent", tc.how)
	}
}

// sliceMayHoldFunc answers true for any type it cannot read as a slice, so
// the loan walk runs; it answers false only for element bases that hold no func.
func TestLendSliceArgMayHoldFunc(t *testing.T) {
	fn := &FuncType{}
	declF := &DeclaredType{PkgPath: "gno.land/p/demo/x", Name: "F", Base: fn}
	declS := &DeclaredType{PkgPath: "gno.land/p/demo/x", Name: "S", Base: &SliceType{Elt: IntType}}
	declFS := &DeclaredType{PkgPath: "gno.land/p/demo/x", Name: "FS", Base: &SliceType{Elt: fn}}
	for _, tc := range []struct {
		name string
		t    Type
		want bool
	}{
		{"nil", nil, true},
		{"non-slice", IntType, true},
		{"nil elt", &SliceType{}, true},
		{"[]int", &SliceType{Elt: IntType}, false},
		{"[]*int", &SliceType{Elt: &PointerType{Elt: IntType}}, false},
		{"[]struct", &SliceType{Elt: &StructType{}}, false},
		{"[]map", &SliceType{Elt: &MapType{Key: StringType, Value: IntType}}, false},
		{"[][]int", &SliceType{Elt: &SliceType{Elt: IntType}}, false},
		{"[][3]int", &SliceType{Elt: &ArrayType{Len: 3, Elt: IntType}}, false},
		{"[]func()", &SliceType{Elt: fn}, true},
		{"[]interface{}", &SliceType{Elt: &InterfaceType{}}, true},
		{"[]F", &SliceType{Elt: declF}, true},
		{"S []int", declS, false},
		{"FS []func()", declFS, true},
	} {
		require.Equal(t, tc.want, sliceMayHoldFunc(tc.t), tc.name)
	}
}
