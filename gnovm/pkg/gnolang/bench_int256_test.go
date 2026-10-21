package gnolang

import (
	"math/big"
	"testing"

	bm "github.com/gnolang/gno/gnovm/pkg/benchops"
)

// Benchmarks for the int256/uint256 op handlers, each paired with the same
// handler on int so that the two can be compared in one run. The OpCPU*Int256
// constants are ratio-scaled from these pairs: see the comment above them in
// machine.go and gnovm/adr/pr6387_native_int256.md.
//
//	go test -run '^$' -bench 'Op256' -benchtime 2000000x ./gnovm/pkg/gnolang

var (
	// Full-width operands: every limb is set, the worst case for the
	// add/sub carry chains and for multiplication.
	bench256A = mustBig("0xfedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210")
	bench256B = mustBig("0x0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	// Divisors of 64, 128 and 192 bits: holiman/uint256's division cost
	// depends on the divisor's length.
	bench256D64  = mustBig("0xfedcba9876543217")
	bench256D128 = mustBig("0xfedcba9876543210fedcba9876543217")
	bench256D192 = mustBig("0xfedcba9876543210fedcba9876543210fedcba9876543217")
)

func mustBig(s string) *big.Int {
	b, ok := new(big.Int).SetString(s, 0)
	if !ok {
		panic(s)
	}
	return b
}

func benchBinaryOp(b *testing.B, lv, rv TypedValue, op func(m *Machine)) {
	b.Helper()
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(lv)
		m.PushValue(rv)
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		op(m)
		bm.SwitchOpCode(bmSetup)
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func benchUnaryOp(b *testing.B, xv TypedValue, op func(m *Machine)) {
	b.Helper()
	m := benchMachine()
	defer m.Release()
	expr := &UnaryExpr{}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(xv)
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		op(m)
		bm.SwitchOpCode(bmSetup)
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func intTV(v int64) TypedValue {
	tv := TypedValue{T: IntType}
	tv.SetInt(v)
	return tv
}

func uintTV(v uint64) TypedValue {
	tv := TypedValue{T: UintType}
	tv.SetUint(v)
	return tv
}

// Baselines on int, through the same harness.
func BenchmarkOp256Base_Add(b *testing.B) {
	benchBinaryOp(b, intTV(-42), intTV(8), (*Machine).doOpAdd)
}

func BenchmarkOp256Base_Sub(b *testing.B) {
	benchBinaryOp(b, intTV(-42), intTV(8), (*Machine).doOpSub)
}

func BenchmarkOp256Base_Mul(b *testing.B) {
	benchBinaryOp(b, intTV(-42), intTV(8), (*Machine).doOpMul)
}

func BenchmarkOp256Base_Quo(b *testing.B) {
	benchBinaryOp(b, intTV(-4200), intTV(8), (*Machine).doOpQuo)
}

func BenchmarkOp256Base_Rem(b *testing.B) {
	benchBinaryOp(b, intTV(-4200), intTV(8), (*Machine).doOpRem)
}

func BenchmarkOp256Base_Band(b *testing.B) {
	benchBinaryOp(b, intTV(-42), intTV(8), (*Machine).doOpBand)
}

func BenchmarkOp256Base_Bor(b *testing.B) {
	benchBinaryOp(b, intTV(-42), intTV(8), (*Machine).doOpBor)
}

func BenchmarkOp256Base_Xor(b *testing.B) {
	benchBinaryOp(b, intTV(-42), intTV(8), (*Machine).doOpXor)
}

func BenchmarkOp256Base_Bandn(b *testing.B) {
	benchBinaryOp(b, intTV(-42), intTV(8), (*Machine).doOpBandn)
}

func BenchmarkOp256Base_Shl(b *testing.B) {
	benchBinaryOp(b, intTV(-42), uintTV(7), (*Machine).doOpShl)
}

func BenchmarkOp256Base_Shr(b *testing.B) {
	benchBinaryOp(b, intTV(-42), uintTV(7), (*Machine).doOpShr)
}

func BenchmarkOp256Base_Eql(b *testing.B) {
	benchBinaryOp(b, intTV(-42), intTV(-42), (*Machine).doOpEql)
}

func BenchmarkOp256Base_Lss(b *testing.B) {
	benchBinaryOp(b, intTV(-42), intTV(-41), (*Machine).doOpLss)
}

func BenchmarkOp256Base_Uneg(b *testing.B) {
	benchUnaryOp(b, intTV(-42), (*Machine).doOpUneg)
}

func BenchmarkOp256Base_Uxor(b *testing.B) {
	benchUnaryOp(b, intTV(-42), (*Machine).doOpUxor)
}

func BenchmarkOp256Base_Convert(b *testing.B) {
	benchConvert(b, intTV(-42), Int64Type)
}

// The 256-bit ops.
func BenchmarkOp256_Add(b *testing.B) {
	benchBinaryOp(b, tv256(Uint256Type, bench256A), tv256(Uint256Type, bench256B), (*Machine).doOpAdd)
}

func BenchmarkOp256_Sub(b *testing.B) {
	benchBinaryOp(b, tv256(Uint256Type, bench256B), tv256(Uint256Type, bench256A), (*Machine).doOpSub)
}

func BenchmarkOp256_Mul(b *testing.B) {
	benchBinaryOp(b, tv256(Uint256Type, bench256A), tv256(Uint256Type, bench256B), (*Machine).doOpMul)
}

func BenchmarkOp256_Quo_Unsigned64(b *testing.B) {
	benchBinaryOp(b, tv256(Uint256Type, bench256A), tv256(Uint256Type, bench256D64), (*Machine).doOpQuo)
}

func BenchmarkOp256_Quo_Unsigned128(b *testing.B) {
	benchBinaryOp(b, tv256(Uint256Type, bench256A), tv256(Uint256Type, bench256D128), (*Machine).doOpQuo)
}

func BenchmarkOp256_Quo_Unsigned192(b *testing.B) {
	benchBinaryOp(b, tv256(Uint256Type, bench256A), tv256(Uint256Type, bench256D192), (*Machine).doOpQuo)
}

func BenchmarkOp256_Quo_Signed128(b *testing.B) {
	// Negative dividend and divisor: two negations around the division.
	benchBinaryOp(b, tv256(Int256Type, new(big.Int).Neg(new(big.Int).Rsh(bench256A, 1))),
		tv256(Int256Type, new(big.Int).Neg(bench256D128)), (*Machine).doOpQuo)
}

func BenchmarkOp256_Rem_Unsigned128(b *testing.B) {
	benchBinaryOp(b, tv256(Uint256Type, bench256A), tv256(Uint256Type, bench256D128), (*Machine).doOpRem)
}

func BenchmarkOp256_Rem_Signed128(b *testing.B) {
	benchBinaryOp(b, tv256(Int256Type, new(big.Int).Neg(new(big.Int).Rsh(bench256A, 1))),
		tv256(Int256Type, new(big.Int).Neg(bench256D128)), (*Machine).doOpRem)
}

func BenchmarkOp256_Band(b *testing.B) {
	benchBinaryOp(b, tv256(Uint256Type, bench256A), tv256(Uint256Type, bench256B), (*Machine).doOpBand)
}

func BenchmarkOp256_Bor(b *testing.B) {
	benchBinaryOp(b, tv256(Uint256Type, bench256A), tv256(Uint256Type, bench256B), (*Machine).doOpBor)
}

func BenchmarkOp256_Xor(b *testing.B) {
	benchBinaryOp(b, tv256(Uint256Type, bench256A), tv256(Uint256Type, bench256B), (*Machine).doOpXor)
}

func BenchmarkOp256_Bandn(b *testing.B) {
	benchBinaryOp(b, tv256(Uint256Type, bench256A), tv256(Uint256Type, bench256B), (*Machine).doOpBandn)
}

func BenchmarkOp256_Shl(b *testing.B) {
	benchBinaryOp(b, tv256(Uint256Type, bench256A), uintTV(77), (*Machine).doOpShl)
}

func BenchmarkOp256_Shr_Signed(b *testing.B) {
	benchBinaryOp(b, tv256(Int256Type, new(big.Int).Neg(bench256B)), uintTV(77), (*Machine).doOpShr)
}

func BenchmarkOp256_Eql(b *testing.B) {
	benchBinaryOp(b, tv256(Uint256Type, bench256A), tv256(Uint256Type, bench256A), (*Machine).doOpEql)
}

func BenchmarkOp256_Lss_Signed(b *testing.B) {
	benchBinaryOp(b, tv256(Int256Type, new(big.Int).Neg(bench256B)), tv256(Int256Type, bench256B), (*Machine).doOpLss)
}

func BenchmarkOp256_Uneg(b *testing.B) {
	benchUnaryOp(b, tv256(Int256Type, bench256B), (*Machine).doOpUneg)
}

func BenchmarkOp256_Uxor(b *testing.B) {
	benchUnaryOp(b, tv256(Uint256Type, bench256A), (*Machine).doOpUxor)
}

func benchConvert(b *testing.B, xv TypedValue, t Type) {
	b.Helper()
	m := benchMachine()
	defer m.Release()

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		tv := xv
		bm.SwitchOpCode(bmTarget)
		ConvertTo(m.Alloc, m.Store, &tv, t, false)
		bm.SwitchOpCode(bmSetup)
	}
	reportBenchops(b)
}

func BenchmarkOp256_Convert_FromInt(b *testing.B) {
	benchConvert(b, intTV(-42), Int256Type)
}

func BenchmarkOp256_Convert_ToInt(b *testing.B) {
	benchConvert(b, tv256(Int256Type, new(big.Int).Neg(bench256B)), IntType)
}

func BenchmarkOp256_Convert_ToFloat64(b *testing.B) {
	benchConvert(b, tv256(Int256Type, new(big.Int).Neg(bench256B)), Float64Type)
}

func BenchmarkOp256_Convert_FromFloat64(b *testing.B) {
	tv := TypedValue{T: Float64Type}
	tv.SetFloat64(0xc6f0000000000001) // about -2.1e33
	benchConvert(b, tv, Int256Type)
}
