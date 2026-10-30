package gnolang

import (
	"io"
	"math"
	"math/big"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
)

// TestGasInt256Ops pins the gas of the 256-bit operators against the same
// handler on int: the CPU difference is the OpCPU*Int256 constant minus the
// 64-bit cost, and a non-zero result also pays for one boxed Bits256Value.
func TestGasInt256Ops(t *testing.T) {
	gasOf := func(lv, rv TypedValue, op func(m *Machine), unary bool) int64 {
		gm := store.NewGasMeter(math.MaxInt64)
		m := NewMachineWithOptions(MachineOptions{
			PkgPath:  "test",
			GasMeter: gm,
			Alloc:    NewAllocator(math.MaxInt64),
		})
		defer m.Release()
		m.Alloc.SetGasMeter(gm)
		m.PushValue(lv)
		if unary {
			m.PushExpr(&UnaryExpr{})
		} else {
			m.PushValue(rv)
			m.PushExpr(&BinaryExpr{})
		}
		before := gm.GasConsumed()
		op(m)
		return gm.GasConsumed() - before
	}
	boxGas := allocGas(allocBits256)
	a := tv256(Uint256Type, big.NewInt(1000))
	b := tv256(Uint256Type, big.NewInt(7))
	ia, ib := intTV(1000), intTV(7)
	cnt := uintTV(3)

	cases := []struct {
		name          string
		op            func(m *Machine)
		unary         bool
		cpu256, cpu64 int64
		boxes         bool // the result is a new non-zero 256-bit value
	}{
		{"add", (*Machine).doOpAdd, false, OpCPUAddInt256, OpCPUAddInt, true},
		{"sub", (*Machine).doOpSub, false, OpCPUSubInt256, OpCPUSubInt, true},
		{"mul", (*Machine).doOpMul, false, OpCPUMulInt256, OpCPUMulInt, true},
		{"quo", (*Machine).doOpQuo, false, OpCPUQuoInt256, OpCPUQuoInt, true},
		{"rem", (*Machine).doOpRem, false, OpCPURemInt256 - OpCPURem, 0, true},
		{"bor", (*Machine).doOpBor, false, OpCPUBorInt256 - OpCPUBor, 0, true},
		{"eql", (*Machine).doOpEql, false, OpCPUEqlInt256 - OpCPUEql, 0, false},
		{"lss", (*Machine).doOpLss, false, OpCPULssInt256 - OpCPULss, 0, false},
		{"uneg", (*Machine).doOpUneg, true, OpCPUUnegInt256 - OpCPUUneg, 0, true},
	}
	for _, c := range cases {
		g256 := gasOf(a, b, c.op, c.unary)
		g64 := gasOf(ia, ib, c.op, c.unary)
		want := c.cpu256 - c.cpu64
		if c.boxes {
			want += boxGas
		}
		assert.Equal(t, want, g256-g64, c.name)
	}
	// Shifts take a uint count.
	g256 := gasOf(a, cnt, (*Machine).doOpShl, false)
	g64 := gasOf(ia, cnt, (*Machine).doOpShl, false)
	assert.Equal(t, OpCPUShlInt256-OpCPUShl+boxGas, g256-g64, "shl")

	// A zero result is stored as nil and allocates nothing.
	zero := gasOf(a, a, (*Machine).doOpSub, false)
	nonzero := gasOf(a, b, (*Machine).doOpSub, false)
	assert.Equal(t, boxGas, nonzero-zero, "zero result")
}

// TestBits256AccountingMatchesStrings checks that 256-bit values are
// accounted for exactly as strings are: an immutable value in V, charged
// once when it is built and shared by its copies, then counted once per slot
// by the GC recount. In each scenario, the number of slots the recount (and
// the allocator) sees for a non-zero 256-bit value is the same as for a
// non-empty string.
func TestBits256AccountingMatchesStrings(t *testing.T) {
	scenarios := map[string]string{
		// One value copied into many slots.
		"copies": `
var keep [100]T
func main() {
	x := V
	for i := range keep {
		keep[i] = x
	}
}`,
		// A method value formed through an interface, retained.
		"bound": `
type N T
func (N) M() {}
var keep []func()
func main() {
	var i interface{ M() } = N(V)
	for j := 0; j < 100; j++ {
		keep = append(keep, i.M)
	}
}`,
		// A saved method called repeatedly, each call retaining its
		// receiver copy.
		"saved": `
type N T
var keep []*N
func (n N) M() { keep = append(keep, &n) }
func main() {
	var i interface{ M() } = N(V)
	f := i.M
	for j := 0; j < 100; j++ {
		f()
	}
}`,
		// A deferred closure retaining a named result.
		"named": `
var keep []*T
func g(v T) (x T) {
	defer func() { keep = append(keep, &x) }()
	return v
}
func main() {
	v := V
	for j := 0; j < 100; j++ {
		g(v)
	}
}`,
	}
	run := func(src, typ, val string) (tracked, recount int64) {
		src = strings.ReplaceAll(strings.ReplaceAll(src, "T", typ), "V", val)
		m := NewMachineWithOptions(MachineOptions{
			PkgPath: "main",
			Alloc:   NewAllocator(math.MaxInt64),
		})
		defer m.Release()
		m.RunFiles(m.MustParseFile("main.gno", "package main\n"+src))
		m.RunMain()
		_, tracked = m.Alloc.Status()
		m.GarbageCollect()
		_, recount = m.Alloc.Status()
		return tracked, recount
	}
	const strLen = 10
	for name, src := range scenarios {
		for _, typ := range []string{"uint256", "int256"} {
			t0, r0 := run(src, typ, typ+"(0)")
			t1, r1 := run(src, typ, typ+"(7)")
			s0t, s0r := run(src, "string", `string("")`)
			s1t, s1r := run(src, "string", `string("abcdefghij")`)
			t.Logf("%s/%s: slots %d, allocated %d", name, typ, (r1-r0)/allocBits256, (t1-t0)/allocBits256)
			require.Zero(t, (r1-r0)%allocBits256, "%s/%s", name, typ)
			require.Zero(t, (s1r-s0r)%strLen, "%s/%s", name, typ)
			require.Equal(t, (s1r-s0r)/strLen, (r1-r0)/allocBits256,
				"%s/%s: recounted slots differ from strings", name, typ)
			require.Equal(t, (s1t-s0t)/strLen, (t1-t0)/allocBits256,
				"%s/%s: allocated values differ from strings", name, typ)
		}
	}
}

// TestGasInt256Equality checks that the 256-bit equality surcharge is paid by
// every comparison, not only by == and != expressions: element by element in
// arrays, field by field in structs, through interface fields, and in switch
// cases, and only as far as the comparison goes.
func TestGasInt256Equality(t *testing.T) {
	gasOf := func(src string) int64 {
		gm := store.NewGasMeter(math.MaxInt64)
		m := NewMachineWithOptions(MachineOptions{
			PkgPath:  "main",
			GasMeter: gm,
			Alloc:    NewAllocator(math.MaxInt64),
		})
		defer m.Release()
		m.Alloc.SetGasMeter(gm)
		m.RunFiles(m.MustParseFile("main.gno", "package main\n"+src))
		before := gm.GasConsumed()
		m.RunMain()
		return gm.GasConsumed() - before
	}
	const extra = OpCPUEqlInt256 - OpCPUEql
	cases := []struct {
		name  string
		src   string // T is replaced by the element type
		comps int64  // 256-bit comparisons made
	}{
		{"array", `func main() { var a, b [1000]T; println(a == b) }`, 1000},
		{"array_early_failure", `func main() { var a, b [1000]T; a[0] = 1; println(a == b) }`, 1},
		{"struct", `type S struct{ A, B, C T }
func main() { var a, b S; println(a == b) }`, 3},
		{"interface_field", `type S struct{ A any }
func main() { a, b := S{T(0)}, S{T(0)}; println(a == b) }`, 1},
		{"switch", `func main() { var x T; switch x { case 3: case 2: case 0: } }`, 3},
		{"neq", `func main() { var x, y T; println(x != y) }`, 1},
	}
	for _, c := range cases {
		g64 := gasOf(strings.ReplaceAll(c.src, "T", "uint64"))
		for _, typ := range []string{"uint256", "int256"} {
			g256 := gasOf(strings.ReplaceAll(c.src, "T", typ))
			// The constant 1 is boxed at preprocessing, so assigning it
			// allocates nothing at run time.
			assert.Equal(t, c.comps*extra, g256-g64, "%s/%s", c.name, typ)
		}
	}
}

// TestBits256ConstPreprocessCharge checks that a 256-bit constant's box,
// built by a constant conversion during preprocessing, is charged once on
// the transaction's preprocessing allocator, as a string literal is.
func TestBits256ConstPreprocessCharge(t *testing.T) {
	preprocess := func(src string) (bytes, gas int64) {
		gm := store.NewGasMeter(math.MaxInt64)
		st, preAlloc := newPreprocessAllocTestStore(t, math.MaxInt64, gm)
		defer st.SetPreprocessAllocator(nil)
		m := NewMachineWithOptions(MachineOptions{
			PkgPath: "gno.land/r/test/consts",
			Store:   st,
			Output:  io.Discard,
			Alloc:   NewAllocator(math.MaxInt64),
		})
		defer m.Release()
		mpkg := &std.MemPackage{
			Type:  MPUserProd,
			Name:  "consts",
			Path:  "gno.land/r/test/consts",
			Files: []*std.MemFile{{Name: "a.gno", Body: "package consts\n" + src}},
		}
		m.RunMemPackage(mpkg, false)
		_, bytes = preAlloc.Status()
		return bytes, gm.GasConsumed()
	}
	cases := []struct {
		name, zero, nonzero string
	}{
		{"untyped int", "const A uint256 = 0", "const A uint256 = 1"},
		{"untyped int, int256", "const B int256 = 0", "const B int256 = 2"},
		{"rune", "const C uint256 = '\\x00'", "const C uint256 = 'x'"},
		{"integral decimal", "const D uint256 = 0.0", "const D uint256 = 1e3"},
		{"typed conversion", "const I int8 = 0\nconst E = uint256(I)", "const I int8 = 5\nconst E = uint256(I)"},
		{"between 256-bit kinds", "const U uint256 = 0\nconst F = int256(U + 0)", "const U uint256 = 0\nconst F = int256(U + 7)"},
	}
	for _, c := range cases {
		b0, g0 := preprocess(c.zero)
		b1, g1 := preprocess(c.nonzero)
		// The programs differ only in the values, so they run the same
		// operations. "between 256-bit kinds": 7 converted to uint256, the
		// folded sum, and the conversion to int256 each build a box.
		boxes := int64(1)
		if c.name == "between 256-bit kinds" {
			boxes = 3
		}
		assert.Equal(t, boxes*allocBits256, b1-b0, "%s: bytes", c.name)
		assert.Equal(t, boxes*allocGas(allocBits256), g1-g0, "%s: gas", c.name)
	}
}
