package gnolang

import (
	"math"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

// TestBits256BoundReceiverAccounting checks that a method value formed
// through an interface, which saves its 256-bit operand without a Copy,
// charges the operand's box like any other slot: the difference between
// non-zero and zero receivers is the same in what was allocated and in what
// the GC recounts.
func TestBits256BoundReceiverAccounting(t *testing.T) {
	run := func(typ, val string) (tracked, recount int64) {
		src := `package main
type N ` + typ + `
func (N) M() {}
var keep []func()
func main() {
	var i interface{ M() } = N(` + val + `)
	for j := 0; j < 100; j++ {
		keep = append(keep, i.M)
	}
}`
		m := NewMachineWithOptions(MachineOptions{
			PkgPath: "main",
			Alloc:   NewAllocator(math.MaxInt64),
		})
		defer m.Release()
		m.RunFiles(m.MustParseFile("main.gno", src))
		m.RunMain()
		_, tracked = m.Alloc.Status()
		m.GarbageCollect()
		_, recount = m.Alloc.Status()
		return tracked, recount
	}
	for _, typ := range []string{"uint256", "int256"} {
		t0, r0 := run(typ, "0")
		t1, r1 := run(typ, "7")
		require.GreaterOrEqual(t, t1-t0, r1-r0, typ)
		require.GreaterOrEqual(t, r1-r0, int64(100*allocBits256), typ)
	}
	// On reload, a bound receiver is charged its box, and a zero one nothing.
	for _, typ := range []Type{Int256Type, Uint256Type} {
		bmv := &BoundMethodValue{Receiver: tv256(typ, big.NewInt(7))}
		require.Equal(t, int64(allocBits256), internalRefSize(bmv))
		require.Equal(t, int64(0), internalRefSize(&BoundMethodValue{Receiver: TypedValue{T: typ}}))
	}
}
