package gnolang

import (
	"math/big"
	"strings"
	"testing"
)

// These helpers are unreachable from .gno source today (operands are
// constant-folded at preprocess) but the gas charge is still wired
// defensively. The tests below pin the metering so a refactor of is*()
// can't silently drop it.

func newMeteredMachine() (*Machine, *recordingMeter) {
	rm := &recordingMeter{}
	return &Machine{GasMeter: rm}, rm
}

func bigintTV(n int64) TypedValue {
	return TypedValue{T: UntypedBigintType, V: BigintValue{V: big.NewInt(n)}}
}

func bigdecTV(t *testing.T, s string) TypedValue {
	t.Helper()
	r, ok := new(big.Rat).SetString(s)
	if !ok {
		t.Fatalf("big.Rat.SetString(%q) failed", s)
	}
	return TypedValue{T: UntypedBigdecType, V: BigdecValue{V: r}}
}

// wideBigint / wideBigdec build operands large enough that the per-N
// charge rounds to non-zero gas after integer division.
func wideBigint() TypedValue {
	return TypedValue{T: UntypedBigintType, V: BigintValue{V: new(big.Int).Lsh(big.NewInt(1), 2048)}}
}

func wideBigdec(t *testing.T) TypedValue {
	t.Helper()
	coeff, ok := new(big.Int).SetString("1"+strings.Repeat("2", 199), 10)
	if !ok {
		t.Fatal("wideBigdec: SetString failed")
	}
	return TypedValue{T: UntypedBigdecType, V: BigdecValue{V: new(big.Rat).SetInt(coeff)}}
}

func TestIsCompareBig_ChargesGas(t *testing.T) {
	// Distinct operand pairs so a future pointer-equality fast-path can't
	// short-circuit the charge.
	bigintLv := wideBigint()
	bigintRv := wideBigint()
	b := bigintRv.V.(BigintValue).V
	b.SetBit(b, 0, 1)

	bigdecLv := wideBigdec(t)
	bigdecRv := wideBigdec(t)
	bigdecRv.V.(BigdecValue).V.Neg(bigdecRv.V.(BigdecValue).V)

	kinds := []struct {
		name   string
		lv, rv TypedValue
	}{
		{"BigInt", bigintLv, bigintRv},
		{"BigDec", bigdecLv, bigdecRv},
	}
	comparators := []struct {
		name string
		call func(m *Machine, lv, rv *TypedValue) bool
	}{
		{"isEql", func(m *Machine, lv, rv *TypedValue) bool { return isEql(m, lv, rv, false) }},
		{"isLss", isLss},
		{"isLeq", isLeq},
		{"isGtr", isGtr},
		{"isGeq", isGeq},
	}
	for _, k := range kinds {
		for _, c := range comparators {
			t.Run(k.name+"/"+c.name, func(t *testing.T) {
				m, rm := newMeteredMachine()
				c.call(m, &k.lv, &k.rv)
				if rm.GasConsumed() == 0 {
					t.Fatalf("expected gas consumption, got 0")
				}
			})
		}
	}
}

// TestIncrCPUBig_DeclaredTypeWrapper verifies the type gate inside the
// per-N helpers uses baseOf, not strict equality — otherwise a hypothetical
// `type Foo bigint` / `type Bar bigdec` would silently bypass metering.
// Constructs the *DeclaredType directly because Gno has no user-facing
// path to do so today.
func TestIncrCPUBig_DeclaredTypeWrapper(t *testing.T) {
	bigintWrap := &DeclaredType{PkgPath: "test", Name: "FooBigint", Base: UntypedBigintType}
	bigdecWrap := &DeclaredType{PkgPath: "test", Name: "BarBigdec", Base: UntypedBigdecType}

	intTV := wideBigint()
	intTV.T = bigintWrap
	decTV := wideBigdec(t)
	decTV.T = bigdecWrap

	cases := []struct {
		name string
		call func(m *Machine)
	}{
		{"incrCPUBigInt", func(m *Machine) { m.incrCPUBigInt(&intTV, &intTV, OpCPUSlopeBigIntCmp) }},
		{"incrCPUBigIntQuad", func(m *Machine) { m.incrCPUBigIntQuad(&intTV, &intTV, OpCPUSlopeBigIntMulQ) }},
		{"incrCPUBigUnary", func(m *Machine) { m.incrCPUBigUnary(&intTV, OpCPUSlopeBigIntUneg) }},
		{"incrCPUBigDec", func(m *Machine) { m.incrCPUBigDec(&decTV, &decTV, OpCPUSlopeBigDecCmp) }},
		{"incrCPUBigDecQuad", func(m *Machine) { m.incrCPUBigDecQuad(&decTV, &decTV, OpCPUSlopeBigDecMulQ) }},
		{"incrCPUBigDecUnary", func(m *Machine) { m.incrCPUBigDecUnary(&decTV, OpCPUSlopeBigDecUneg) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m, rm := newMeteredMachine()
			tc.call(m)
			if rm.GasConsumed() == 0 {
				t.Fatalf("declared-type wrapper bypassed metering")
			}
		})
	}
}

// TestIsCompareBig_CorrectResults sanity-checks comparator semantics in
// case a future refactor of is*() swaps a comparator while keeping the
// gas charge intact.
func TestIsCompareBig_CorrectResults(t *testing.T) {
	t.Run("BigInt", func(t *testing.T) {
		m, _ := newMeteredMachine()
		a := bigintTV(7)
		b := bigintTV(11)
		if isEql(m, &a, &b, false) {
			t.Error("isEql(7, 11) = true, want false")
		}
		if !isLss(m, &a, &b) {
			t.Error("isLss(7, 11) = false, want true")
		}
		if !isLeq(m, &a, &b) {
			t.Error("isLeq(7, 11) = false, want true")
		}
		if isGtr(m, &a, &b) {
			t.Error("isGtr(7, 11) = true, want false")
		}
		if isGeq(m, &a, &b) {
			t.Error("isGeq(7, 11) = true, want false")
		}
	})

	t.Run("BigDec", func(t *testing.T) {
		m, _ := newMeteredMachine()
		a := bigdecTV(t, "7.5")
		b := bigdecTV(t, "11.25")
		if isEql(m, &a, &b, false) {
			t.Error("isEql(7.5, 11.25) = true, want false")
		}
		if !isLss(m, &a, &b) {
			t.Error("isLss(7.5, 11.25) = false, want true")
		}
		if !isLeq(m, &a, &b) {
			t.Error("isLeq(7.5, 11.25) = false, want true")
		}
		if isGtr(m, &a, &b) {
			t.Error("isGtr(7.5, 11.25) = true, want false")
		}
		if isGeq(m, &a, &b) {
			t.Error("isGeq(7.5, 11.25) = true, want false")
		}
	})
}
