package gnolang

import (
	"math"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"

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
