package gnolang

import (
	"fmt"
	"go/token"
	"math"
	"math/big"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gnovm/pkg/gnolang/internal/goconst"
	stypes "github.com/gnolang/gno/tm2/pkg/store/types"
)

// Operand constructors.

// lit is the operand's source, which error messages print.
func lit(v any) *BasicLitExpr { return &BasicLitExpr{Kind: INT, Value: fmt.Sprint(v)} }

func cRune(v int64) *ConstExpr {
	tv := TypedValue{T: UntypedRuneType}
	setRuneConst(&tv, big.NewInt(v))
	return &ConstExpr{Source: lit(v), TypedValue: tv}
}

func cBigint(s string) *ConstExpr {
	bi, ok := new(big.Int).SetString(s, 10)
	if !ok {
		panic(s)
	}
	return &ConstExpr{Source: lit(s), TypedValue: TypedValue{T: UntypedBigintType, V: BigintValue{V: bi}}}
}

func cBigdec(s string) *ConstExpr {
	r, ok := new(big.Rat).SetString(s)
	if !ok {
		panic(s)
	}
	return &ConstExpr{Source: &BasicLitExpr{Kind: FLOAT, Value: s}, TypedValue: TypedValue{T: UntypedBigdecType, V: BigdecValue{V: r}}}
}

func cStr(s string) *ConstExpr {
	return &ConstExpr{TypedValue: TypedValue{T: UntypedStringType, V: StringValue(s)}}
}

func cBool(b bool) *ConstExpr {
	tv := TypedValue{T: UntypedBoolType}
	tv.SetBool(b)
	return &ConstExpr{TypedValue: tv}
}

func cTyped(t Type, v int64) *ConstExpr {
	tv := TypedValue{T: t}
	setIntKind(&tv, t.Kind(), big.NewInt(v))
	return &ConstExpr{Source: lit(v), TypedValue: tv}
}

func cUint(t Type, v uint64) *ConstExpr {
	tv := TypedValue{T: t}
	setIntKind(&tv, t.Kind(), new(big.Int).SetUint64(v))
	return &ConstExpr{Source: lit(v), TypedValue: tv}
}

func cFloat32(f float32) *ConstExpr {
	tv := TypedValue{T: Float32Type}
	tv.SetFloat32(math.Float32bits(f))
	return &ConstExpr{Source: lit(f), TypedValue: tv}
}

func cFloat64(f float64) *ConstExpr {
	tv := TypedValue{T: Float64Type}
	tv.SetFloat64(math.Float64bits(f))
	return &ConstExpr{Source: lit(f), TypedValue: tv}
}

// constText renders a folded constant as "value type", exactly.
func constText(tv TypedValue) string {
	var v string
	switch baseOf(tv.T).Kind() {
	case BoolKind:
		v = fmt.Sprint(tv.GetBool())
	case StringKind:
		v = fmt.Sprintf("%q", tv.GetString())
	case Int32Kind:
		v = runeConstInt(&tv).String()
	case Float32Kind:
		v = fmt.Sprint(math.Float32frombits(tv.GetFloat32()))
	case Float64Kind:
		v = fmt.Sprint(math.Float64frombits(tv.GetFloat64()))
	case BigintKind:
		v = tv.GetBigInt().String()
	case BigdecKind:
		bd := tv.GetBigDec()
		if bd.IsFloat() {
			v = "f:" + bd.F.Text('g', 20)
		} else {
			v = bd.V.RatString()
		}
	case UintKind, Uint8Kind, Uint16Kind, Uint64Kind:
		tv2 := tv
		ConvertTo(nil, nil, &tv2, Uint64Type, false)
		v = fmt.Sprint(tv2.GetUint64())
	default:
		tv2 := tv // ConvertGetInt converts in place
		v = fmt.Sprint(tv2.ConvertGetInt())
	}
	return v + " " + tv.T.String()
}

// fold folds x and returns its text, or the error message.
func fold(gm stypes.GasMeter, x Expr) (text string, err string) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Sprint(r)
		}
	}()
	tv, ok := foldConst(gm, nil, x)
	if !ok {
		return "", "not folded"
	}
	return constText(tv), ""
}

// named gives cx the source name n, as a constant declared `const n = ...`.
func named(n Name, cx *ConstExpr) *ConstExpr { cx.Source = Nx(n); return cx }

func bin(l Expr, op Word, r Expr) *BinaryExpr { return &BinaryExpr{Left: l, Op: op, Right: r} }
func un(op Word, x Expr) *UnaryExpr           { return &UnaryExpr{Op: op, X: x} }

const maxRune = 1114111

func TestFoldConst(t *testing.T) {
	tests := []struct {
		name string
		x    Expr
		want string // value text, or "error: msg"
	}{
		// Untyped rune: exact, past int32 when needed, rune kind kept.
		{"rune mul", bin(cRune(maxRune), MUL, cRune(maxRune)), "1241243320321 <untyped> int32"},
		{"rune add", bin(cRune(1241243320321), ADD, cRune(maxRune)), "1241244434432 <untyped> int32"},
		{"rune sub", bin(cRune('a'), SUB, cRune(1241243320321)), "-1241243320224 <untyped> int32"},
		{"rune quo back in range", bin(cRune(1241243320321), QUO, cRune(maxRune)), "1114111 <untyped> int32"},
		{"rune rem", bin(cRune(1241243320321), REM, cRune(1000)), "321 <untyped> int32"},
		{"rune band", bin(cRune(1241243320321), BAND, cRune(1241243320320)), "1241243320320 <untyped> int32"},
		{"rune bor", bin(cRune(1241243320321), BOR, cRune(maxRune)), "1241243385855 <untyped> int32"},
		{"rune xor", bin(cRune(1241243320321), XOR, cRune('a')), "1241243320416 <untyped> int32"},
		{"rune bandn", bin(cRune(1241243320321), BAND_NOT, cRune(maxRune)), "1241242271744 <untyped> int32"},
		{"rune shl", bin(cRune('a'), SHL, cTyped(UintType, 40)), "106652627894272 <untyped> int32"},
		{"rune shr", bin(cRune(1241243320321), SHR, cTyped(UintType, 20)), "1183741 <untyped> int32"},
		{"rune neg", un(SUB, cRune('a')), "-97 <untyped> int32"},
		{"rune not", un(XOR, cRune('a')), "-98 <untyped> int32"},
		{"rune pos", un(ADD, cRune('a')), "97 <untyped> int32"},
		{"rune neg big", un(SUB, cRune(1241243320321)), "-1241243320321 <untyped> int32"},
		{"rune in range stays small", bin(cRune('a'), ADD, cRune('b')), "195 <untyped> int32"},
		{"rune cmp", bin(cRune(1241243320321), GTR, cRune(0)), "true <untyped> bool"},
		{"rune eql", bin(cRune(1241243320321), EQL, cRune(-2228223)), "false <untyped> bool"},
		{"rune quo zero", bin(cRune('a'), QUO, cRune(0)), "error: invalid operation: division by zero"},
		// Untyped int.
		{"int add", bin(cBigint("1"), ADD, cBigint("2")), "3 <untyped> bigint"},
		{"int quo truncates", bin(cBigint("-7"), QUO, cBigint("2")), "-3 <untyped> bigint"},
		{"int rem", bin(cBigint("-7"), REM, cBigint("2")), "-1 <untyped> bigint"},
		{"int big mul", bin(cBigint("1267650600228229401496703205376"), MUL, cBigint("1267650600228229401496703205376")), "1606938044258990275541962092341162602522202993782792835301376 <untyped> bigint"},
		{"int shl", bin(cBigint("1"), SHL, cTyped(UintType, 100)), "1267650600228229401496703205376 <untyped> bigint"},
		{"int shr", bin(cBigint("1267650600228229401496703205376"), SHR, cTyped(UintType, 90)), "1024 <untyped> bigint"},
		{"int shl max count", bin(cBigint("1"), SHL, cTyped(UintType, maxConstShift)), new(big.Int).Lsh(big.NewInt(1), maxConstShift).String() + " <untyped> bigint"},
		{"int shl past max count", bin(cBigint("1"), SHL, cTyped(UintType, maxConstShift+1)), "error: invalid operation: invalid shift count 1075"},
		{"int neg", un(SUB, cBigint("1267650600228229401496703205376")), "-1267650600228229401496703205376 <untyped> bigint"},
		{"int not", un(XOR, cBigint("0")), "-1 <untyped> bigint"},
		{"int cmp", bin(cBigint("1267650600228229401496703205376"), LSS, cBigint("1267650600228229401496703205377")), "true <untyped> bool"},
		{"int rem zero", bin(cBigint("1"), REM, cBigint("0")), "error: invalid operation: division by zero"},
		// Untyped float: exact rationals.
		{"float add", bin(cBigdec("1/10"), ADD, cBigdec("9/10")), "1 <untyped> bigdec"},
		{"float quo", bin(cBigdec("5"), QUO, cBigdec("3")), "5/3 <untyped> bigdec"},
		{"float mul", bin(cBigdec("1e100"), MUL, cBigdec("1e100")), "1" + strings.Repeat("0", 200) + " <untyped> bigdec"},
		{"float neg", un(SUB, cBigdec("1/2")), "-1/2 <untyped> bigdec"},
		{"float neg zero is zero", un(SUB, cBigdec("0")), "0 <untyped> bigdec"},
		{"float cmp", bin(cBigdec("1/123456789"), LSS, cBigdec("1/123456788")), "true <untyped> bool"},
		{"float shl integral", bin(cBigdec("2"), SHL, cTyped(UintType, 3)), "16 <untyped> bigdec"},
		{"float shl fraction", bin(cBigdec("3/2"), SHL, cTyped(UintType, 3)), "error: invalid operation: shifted operand 3/2 (type <untyped> bigdec) must be integer"},
		// Strings and bools.
		{"string add", bin(cStr("foo"), ADD, cStr("bar")), `"foobar" <untyped> string`},
		{"string lss", bin(cStr("foo"), LSS, cStr("bar")), "false <untyped> bool"},
		{"string eql", bin(cStr("foo"), EQL, cStr("foo")), "true <untyped> bool"},
		{"bool land", bin(cBool(true), LAND, cBool(false)), "false <untyped> bool"},
		{"bool lor", bin(cBool(true), LOR, cBool(false)), "true <untyped> bool"},
		{"bool not", un(NOT, cBool(true)), "false <untyped> bool"},
		// Typed integers: representable or rejected with gc's wording.
		{"int8 ok", bin(cTyped(Int8Type, 100), ADD, cTyped(Int8Type, 27)), "127 int8"},
		{"int8 overflow", bin(cTyped(Int8Type, 100), MUL, cTyped(Int8Type, 2)), "error: 100 * 2 (constant 200 of type int8) overflows int8"},
		{"uint8 underflow", bin(cTyped(Uint8Type, 0), SUB, cTyped(Uint8Type, 1)), "error: 0 - 1 (constant -1 of type uint8) overflows uint8"},
		{"int64 overflow", bin(cTyped(Int64Type, 1<<62), MUL, cTyped(Int64Type, 4)), "error: 4611686018427387904 * 4 (constant 18446744073709551616 of type int64) overflows int64"},
		{"uint64 ok", bin(cUint(Uint64Type, math.MaxUint64), QUO, cTyped(Uint64Type, 2)), "9223372036854775807 uint64"},
		{"int32 typed rune", bin(cTyped(Int32Type, maxRune), MUL, cTyped(Int32Type, maxRune)), "error: 1114111 * 1114111 (constant 1241243320321 of type int32) overflows int32"},
		{"int8 neg min", un(SUB, named("m", cTyped(Int8Type, -128))), "error: -m (constant 128 of type int8) overflows int8"},
		{"int8 quo -1", bin(named("m", cTyped(Int8Type, -128)), QUO, cTyped(Int8Type, -1)), "error: m / -1 (constant 128 of type int8) overflows int8"},
		{"uint shl overflow", bin(cTyped(UintType, 3), SHL, cTyped(UintType, 63)), "error: 3 << 63 (constant 27670116110564327424 of type uint) overflows uint"},
		{"uint8 shl ok", bin(cTyped(Uint8Type, 1), SHL, cTyped(UintType, 7)), "128 uint8"},
		{"uint8 not", un(XOR, cTyped(Uint8Type, 0)), "255 uint8"},
		{"uint32 not", un(XOR, cTyped(Uint32Type, 0)), "4294967295 uint32"},
		{"uint64 not", un(XOR, cTyped(Uint64Type, 0)), "18446744073709551615 uint64"},
		{"int64 not", un(XOR, cTyped(Int64Type, 0)), "-1 int64"},
		{"int cmp typed", bin(cTyped(Int8Type, 1), LEQ, cTyped(Int8Type, 1)), "true <untyped> bool"},
		{"int quo typed zero", bin(cTyped(IntType, 1), QUO, cTyped(IntType, 0)), "error: invalid operation: division by zero"},
		{"nested operand parenthesized", bin(&ConstExpr{Source: bin(cTyped(Int8Type, 100), ADD, cTyped(Int8Type, 27)), TypedValue: cTyped(Int8Type, 127).TypedValue}, MUL, cTyped(Int8Type, 2)), "error: (100 + 27) * 2 (constant 254 of type int8) overflows int8"},
		// Typed floats: exact, rounded once to the type, infinity rejected.
		{"float32 mul", bin(cFloat32(0.1), MUL, cFloat32(3)), "0.3 float32"},
		{"float64 add", bin(cFloat64(0.1), ADD, cFloat64(0.2)), "0.30000000000000004 float64"},
		{"float64 overflow", bin(cFloat64(1e300), MUL, cFloat64(1e300)), "error: 1e+300 * 1e+300 (constant 1e+600 of type float64) overflows float64"},
		{"float32 overflow", bin(cFloat32(3e38), MUL, cFloat32(10)), "error: 3e+38 * 10 (constant 3e+39 of type float32) overflows float32"},
		{"float32 neg zero", un(SUB, cFloat32(0)), "0 float32"},
		{"float64 quo", bin(cFloat64(1), QUO, cFloat64(3)), "0.3333333333333333 float64"},
		{"float cmp exact", bin(cFloat32(0.1), GTR, cFloat64(0.1)), "true <untyped> bool"}, // float32(0.1) > 0.1 exactly
		// Not foldable here: a nil or non-primitive operand.
		{"nil operand", bin(&ConstExpr{}, EQL, &ConstExpr{}), "error: not folded"},
		{"non const", bin(&NameExpr{}, ADD, cBigint("1")), "error: not folded"},
		{"receive", un(ARROW, cBigint("1")), "error: not folded"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := fold(nil, tc.x)
			if err != "" {
				got = "error: " + err
			}
			require.Equal(t, tc.want, got)
		})
	}
}

func TestFoldConstGas(t *testing.T) {
	gas := func(x Expr) int64 {
		gm := stypes.NewGasMeter(1_000_000_000)
		_, err := fold(gm, x)
		require.Empty(t, err)
		return gm.GasConsumed() / GasFactorCPU
	}
	flat := int64(OpCPUSlopeConstFold)
	// One flat charge per operation on word-sized operands.
	require.Equal(t, flat, gas(bin(cRune('a'), ADD, cRune('b'))))
	require.Equal(t, flat, gas(un(SUB, cTyped(Int8Type, 1))))
	require.Equal(t, flat, gas(bin(cTyped(Int64Type, 1), SHL, cTyped(UintType, 60))))
	require.Equal(t, flat, gas(bin(cBigint("1"), SHL, cTyped(UintType, 500)))) // a word shifted
	// Untyped integers: the machine's size terms, a shift by its count and
	// the operand's width, a product quadratically.
	wide := cBigint(new(big.Int).Lsh(big.NewInt(1), 10000).String()) // 10001 bits
	require.Equal(t, flat+1023*OpCPUSlopeBigIntShl/1024+10001*OpCPUSlopeBigIntShl/1024,
		gas(bin(wide, SHL, cTyped(UintType, 1023))))
	require.Equal(t, flat+(10001/32)*(10001/32)*OpCPUSlopeBigIntMulQ/32, gas(bin(wide, MUL, wide)))
	require.Equal(t, flat+10001*OpCPUSlopeBigIntAdd/1024, gas(bin(wide, ADD, cBigint("1"))))
	require.Equal(t, flat+10001*OpCPUSlopeBigIntUneg/1024, gas(un(SUB, wide)))
	require.Equal(t, flat+10001*OpCPUSlopeBigIntUxor/1024, gas(un(XOR, wide)))
	// Rationals: a base per operation, linear and quadratic terms in the
	// digits (bits/3 of numerator plus denominator); a comparison or a
	// conversion a smaller base and the cross-multiplication terms.
	r := cBigdec("1/3") // 3 bits: 1 digit
	require.Equal(t, flat+OpCPUSlopeConstRatBase+OpCPUSlopeConstRat/100, gas(bin(r, ADD, r)))
	require.Equal(t, flat+OpCPUSlopeConstRatConv+OpCPUSlopeConstRatCmp/100, gas(bin(r, LSS, r)))
	require.Equal(t, flat+OpCPUSlopeConstRatBase+OpCPUSlopeConstRatNeg/100, gas(un(SUB, r)))
	wideRat := cBigdec("1/" + new(big.Int).Lsh(big.NewInt(1), 4000).String()) // 4002 bits: 1334 digits
	require.Equal(t, flat+OpCPUSlopeConstRatBase+1334*OpCPUSlopeConstRat/100+133*133*OpCPUSlopeConstRatQ/10,
		gas(bin(wideRat, MUL, wideRat)))
	require.Equal(t, flat+OpCPUSlopeConstRatConv+1334*OpCPUSlopeConstRatCmp/100+133*133*OpCPUSlopeConstRatCmpQ/10,
		gas(bin(wideRat, LSS, wideRat)))
	// A typed float is a small rational: the operation, then the rounding
	// to the type (a conversion).
	require.Equal(t, flat+OpCPUSlopeConstRatBase+OpCPUSlopeConstRat/100+OpCPUSlopeConstRatConv+OpCPUSlopeConstRatCmp/100,
		gas(bin(cFloat64(1), QUO, cFloat64(3))))
	// The 512-bit float form: fixed width, a division costs the same again.
	bigFloat := &ConstExpr{TypedValue: TypedValue{T: UntypedBigdecType, V: BigdecValue{F: new(big.Float).SetPrec(BigdecFloatPrec).SetMantExp(big.NewFloat(1), 100000)}}}
	require.Equal(t, flat+OpCPUSlopeConstFloat, gas(bin(bigFloat, MUL, bigFloat)))
	require.Equal(t, flat+OpCPUSlopeConstFloat+OpCPUSlopeConstFloatQuo, gas(bin(bigFloat, QUO, bigFloat)))
	// Strings: a comparison per byte of the longer; a concatenation is paid
	// by its allocation.
	require.Equal(t, flat+3*OpCPUSlopeBytesCmp, gas(bin(cStr("foo"), LSS, cStr("bar"))))
	require.Equal(t, flat, gas(bin(cStr("foo"), ADD, cStr("bar"))))
	// A nil meter charges nothing and still folds.
	_, err := fold(nil, bin(cBigint("1"), ADD, cBigint("2")))
	require.Empty(t, err)
}

// TestConstFoldGasCoversTime: at every operand size the fold allows, the
// gas charged covers the time taken, on reference hardware. Times are
// measured in-process and scaled by the reference operation, a 512-bit
// integer division (~350 ns on the Xeon reference, OpCPUSlopeConstFold),
// so the check does not depend on the speed of the machine running it.
func TestConstFoldGasCoversTime(t *testing.T) {
	if testing.Short() {
		t.Skip("timing")
	}
	r := rand.New(rand.NewSource(1))
	randInt := func(bits int) *big.Int {
		n := new(big.Int).Rand(r, new(big.Int).Lsh(big.NewInt(1), uint(bits)))
		return n.SetBit(n, bits-1, 1)
	}
	bigint := func(bits int) *ConstExpr {
		return &ConstExpr{TypedValue: TypedValue{T: UntypedBigintType, V: BigintValue{V: randInt(bits)}}}
	}
	rat := func(bits int) *ConstExpr {
		d := randInt(bits)
		d.SetBit(d, 0, 1)
		return &ConstExpr{TypedValue: TypedValue{T: UntypedBigdecType, V: BigdecValue{V: new(big.Rat).SetFrac(randInt(bits), d)}}}
	}
	decimal := func(digits int) *ConstExpr { // a literal: denominator a power of ten
		n := new(big.Int).Rand(r, new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(digits)), nil))
		d := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(digits/2)), nil)
		return &ConstExpr{TypedValue: TypedValue{T: UntypedBigdecType, V: BigdecValue{V: new(big.Rat).SetFrac(n, d)}}}
	}
	float := func() *ConstExpr {
		f := new(big.Float).SetPrec(BigdecFloatPrec).SetInt(randInt(512))
		return &ConstExpr{TypedValue: TypedValue{T: UntypedBigdecType, V: BigdecValue{F: f.SetMantExp(f, 3000)}}}
	}
	str := func(n int) *ConstExpr { return cStr(strings.Repeat("a", n)) }
	ns := func(x Expr) float64 { // the best of five runs: robust to pauses
		best := math.MaxFloat64
		for range 5 {
			const n = 200
			start := time.Now()
			for range n {
				foldConst(nil, nil, x)
			}
			best = min(best, float64(time.Since(start).Nanoseconds())/n)
		}
		return best
	}
	charged := func(x Expr) float64 {
		gm := stypes.NewGasMeter(math.MaxInt64)
		foldConst(gm, nil, x)
		return float64(gm.GasConsumed() / GasFactorCPU)
	}
	// 1 gas = 1 ns on the reference; the flat charge covers the reference
	// operation at ~350 ns there.
	const refNs = 350
	ref := bin(bigint(512), QUO, bigint(500))
	scale := refNs / ns(ref)

	type tc struct {
		name string
		x    Expr
	}
	var cases []tc
	for _, bits := range []int{64, 256, 512, 1024, 2048, 3000, 4095} {
		for _, op := range []Word{ADD, SUB, MUL, QUO, LSS, EQL} {
			cases = append(cases, tc{fmt.Sprintf("rat%d %s", bits, op.TokenString()), bin(rat(bits), op, rat(bits))})
		}
		cases = append(cases, tc{fmt.Sprintf("rat%d neg", bits), un(SUB, rat(bits))})
		cases = append(cases, tc{fmt.Sprintf("rat%d op 1/3", bits), bin(rat(bits), ADD, cBigdec("1/3"))})
	}
	for _, digits := range []int{10, 100, 1000, 2400} {
		for _, op := range []Word{ADD, MUL, QUO, LSS} {
			cases = append(cases, tc{fmt.Sprintf("decimal%d %s", digits, op.TokenString()), bin(decimal(digits), op, decimal(digits))})
		}
	}
	for _, op := range []Word{ADD, MUL, QUO, LSS} {
		cases = append(cases, tc{"float512 " + op.TokenString(), bin(float(), op, float())})
	}
	for _, bits := range []int{64, 512} {
		for _, op := range []Word{ADD, SUB, MUL, QUO, REM, BAND, XOR, LSS} {
			cases = append(cases, tc{fmt.Sprintf("int%d %s", bits, op.TokenString()), bin(bigint(bits), op, bigint(bits))})
		}
		cases = append(cases, tc{fmt.Sprintf("int%d neg", bits), un(SUB, bigint(bits))},
			tc{fmt.Sprintf("int%d not", bits), un(XOR, bigint(bits))},
			tc{fmt.Sprintf("int%d shl 1074", bits), bin(bigint(bits), SHL, cTyped(UintType, maxConstShift))},
			tc{fmt.Sprintf("int%d shr 500", bits), bin(bigint(bits), SHR, cTyped(UintType, 500))})
	}
	cases = append(cases,
		tc{"rune mul", bin(cRune(maxRune), MUL, cRune(maxRune))},
		tc{"typed float64 quo", bin(cFloat64(1), QUO, cFloat64(3))},
		tc{"typed int64 mul", bin(cTyped(Int64Type, 1<<30), MUL, cTyped(Int64Type, 1<<30))},
		tc{"string 1MB eql", bin(str(1<<20), EQL, str(1<<20))},
		tc{"string 1MB lss", bin(str(1<<20), LSS, str(1<<20))},
	)
	// Charges carry margins of 1.5x and more; 10% is allowed for noise.
	for _, c := range cases {
		got, want := charged(c.x), ns(c.x)*scale*0.9
		if got < want {
			t.Errorf("%s: charged %.0f gas, took %.0f ns (%.0f on the reference)", c.name, got, want/scale, want)
		}
	}
}

func TestFoldConstDeterminism(t *testing.T) {
	xs := []Expr{
		bin(cRune(maxRune), MUL, cRune(maxRune)),
		bin(cBigdec("1/3"), MUL, cBigdec("3/7")),
		bin(cBigdec("1e1000"), MUL, cBigdec("1e1000")),
		bin(cFloat32(0.1), MUL, cFloat32(3)),
		bin(cBigint("1"), SHL, cTyped(UintType, 500)),
		un(XOR, cTyped(Uint32Type, 5)),
	}
	var first []string
	for run := range 50 {
		var got []string
		for _, x := range xs {
			tv, ok := foldConst(nil, nil, x)
			require.True(t, ok)
			got = append(got, constText(tv)+fmt.Sprint(tv.N))
		}
		if run == 0 {
			first = got
			continue
		}
		require.Equal(t, first, got, "run %d", run)
	}
}

// TestFoldConstPreprocess: the fold's rejections surface from Preprocess
// with the node's location, as the machine path's did.
func TestFoldConstPreprocess(t *testing.T) {
	for _, tc := range []struct {
		src, want string
	}{
		{"const r = '\\U0010ffff' * '\\U0010ffff'\nvar _ = r", "constant 1241243320321 overflows rune"},
		{"var _ uint64 = 'b' - 'c'", "constant -1 overflows uint64"},
		{"var _ int8 = 'a' * 'b'", "constant 9506 overflows int8"},
		{"var _ = 'a' / 0", "invalid operation: division by zero"},
		{"const c int8 = 100\nvar _ = c * 2", "c * 2 (constant 200 of type int8) overflows int8"},
		{"const k uint = 3\nvar _ = k << 63", "k << 63 (constant 27670116110564327424 of type uint) overflows uint"},
		{"var _ = 0 << 1075", "invalid operation: invalid shift count 1075"},
		{"const _ = 1 << 512", "constant shift overflow"},
		{"const c = 1 << 511\nconst _ = c + c", "constant addition overflow"},
		{"const c = 1 << 511\nconst _ = -c - c", "constant subtraction overflow"},
		{"const c = 1 << 511\nconst _ = c * 4", "constant multiplication overflow"},
		{"const _ = 1" + strings.Repeat("0", 155), "constant overflow"},
		{"const _ = 'a' << 1000", "constant shift overflow"},
		{"var _ int64 = 'a' << 40\nconst _ = 1 << 511\nconst _ = 0 << 1074\nconst _ = 1<<511 - 1 + (1 << 511)", ""},
	} {
		m := NewMachine("test", nil)
		n := m.MustParseFile("main.gno", "package test\n"+tc.src+"\nfunc main() {}\n")
		err := func() (err string) {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Sprint(r)
				}
			}()
			m.RunFiles(n)
			return ""
		}()
		if tc.want == "" {
			require.Empty(t, err, tc.src)
			continue
		}
		require.Contains(t, err, "test/main.gno:", tc.src)
		require.Contains(t, err, tc.want, tc.src)
	}
}

// TestConvertUntypedRune: both forms of an untyped rune constant convert
// exactly, range checked for integers, rounded once for floats.
func TestConvertUntypedRune(t *testing.T) {
	conv := func(v int64, to Type) (text string, err string) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Sprint(r)
			}
		}()
		tv := cRune(v).TypedValue
		ConvertUntypedTo(nil, &tv, to)
		return constText(tv), ""
	}
	for _, tc := range []struct {
		v    int64
		to   Type
		want string
	}{
		{'a', Int32Type, "97 int32"},
		{'a', UntypedRuneType, "97 <untyped> int32"},
		{1241243320321, UntypedRuneType, "1241243320321 <untyped> int32"},
		{1241243320321, Int64Type, "1241243320321 int64"},
		{1241243320321, IntType, "1241243320321 int"},
		{1241243320321, Uint64Type, "1241243320321 uint64"},
		{1241243320321, UintType, "1241243320321 uint"},
		{1241243320321, Int32Type, "error: constant 1241243320321 overflows rune"},
		{1241243320321, Uint32Type, "error: constant 1241243320321 overflows uint32"},
		{-1, Uint64Type, "error: constant -1 overflows uint64"},
		{-1, Uint8Type, "error: constant -1 overflows uint8"},
		{256, Uint8Type, "error: constant 256 overflows uint8"},
		{-129, Int8Type, "error: constant -129 overflows int8"},
		{200, Int16Type, "200 int16"},
		{'a', Float64Type, "97 float64"},
		{1241243320321, Float64Type, "1.241243320321e+12 float64"},
		{1241243320321, Float32Type, "1.2412433e+12 float32"},
		{16777217, Float32Type, "1.6777216e+07 float32"}, // rounded once, to even
		{1241243320321, UntypedBigintType, "1241243320321 <untyped> bigint"},
		{1241243320321, UntypedBigdecType, "1241243320321 <untyped> bigdec"},
		{'a', BoolType, "error: cannot convert untyped rune type to BoolKind"},
	} {
		got, err := conv(tc.v, tc.to)
		if err != "" {
			got = "error: " + err
		}
		require.Equal(t, tc.want, got, "%d to %s", tc.v, tc.to)
	}
	// A rune past int32 cannot be read as one.
	require.PanicsWithValue(t, "constant 1241243320321 overflows rune", func() {
		cRune(1241243320321).GetInt32()
	})
}

// TestConvertUntypedIntToFloat32RoundsOnce pins #6394: 1<<54 + 1<<30 + 1
// is just above a float32 midpoint and must round up.
func TestConvertUntypedIntToFloat32RoundsOnce(t *testing.T) {
	for _, tc := range []struct {
		v    string
		want float32
	}{
		{"18014399583223809", 18014400656965632},
		{"-18014399583223809", -18014400656965632},
		{"18014399583223808", 18014398509481984}, // the midpoint: to even
		{"16777217", 16777216},
		{"16777219", 16777220},
	} {
		tv := cBigint(tc.v).TypedValue
		ConvertUntypedTo(nil, &tv, Float32Type)
		require.Equal(t, tc.want, math.Float32frombits(tv.GetFloat32()), tc.v)
	}
}

// BenchmarkConstFold measures the integer cases OpCPUSlopeConstFold is
// calibrated to (512-bit operands) and the rational cap.
func BenchmarkConstFold(b *testing.B) {
	x := cBigint(strings.Repeat("9", 77)) // 256 bits
	y := cBigint(strings.Repeat("8", 77))
	w := cBigint(strings.Repeat("7", 154)) // 512 bits
	b.Run("mul512", func(b *testing.B) {
		bx := bin(x, MUL, y)
		for b.Loop() {
			foldConst(nil, nil, bx)
		}
	})
	b.Run("quo512", func(b *testing.B) {
		bx := bin(w, QUO, y)
		for b.Loop() {
			foldConst(nil, nil, bx)
		}
	})
	b.Run("rat4096", func(b *testing.B) {
		r := cBigdec(strings.Repeat("9", 1230) + "/" + strings.Repeat("7", 1231))
		bx := bin(r, ADD, r)
		for b.Loop() {
			foldConst(nil, nil, bx)
		}
	})
}

// TestConstFoldMetered: every goconst operation the fold and the constant
// conversions call is given its Meter.
func TestConstFoldMetered(t *testing.T) {
	for _, file := range []string{"constfold.go", "values_conversions.go"} {
		src, err := os.ReadFile(file)
		require.NoError(t, err)
		for i, line := range strings.Split(string(src), "\n") {
			require.NotRegexp(t, `goconst\.\w+\(nil\b`, line, "%s:%d", file, i+1)
		}
	}
}

// TestConstFoldWorkReported: each representation reports its work with
// the sizes the meter prices on, before the operation runs.
func TestConstFoldWorkReported(t *testing.T) {
	var got []goconst.Work
	m := workRecorder{&got}
	wide := cBigint(new(big.Int).Lsh(big.NewInt(1), 300).String())
	wideRat := cBigdec("1/" + new(big.Int).Lsh(big.NewInt(1), 4000).String())
	for _, tc := range []struct {
		x    Expr
		want []goconst.Work
	}{
		{bin(wide, MUL, wide), []goconst.Work{{Op: token.MUL, Rep: goconst.RepInt, Bits: 301}}},
		{bin(wide, SHL, cTyped(UintType, 100)), []goconst.Work{{Op: token.SHL, Rep: goconst.RepInt, Bits: 301, Shift: 100}}},
		{bin(wideRat, ADD, cBigdec("1/3")), []goconst.Work{{Op: token.ADD, Rep: goconst.RepRat, Bits: 4002}}},
		{bin(wideRat, LSS, wideRat), []goconst.Work{{Op: token.LSS, Rep: goconst.RepRat, Bits: 4002}}},
		{un(SUB, wideRat), []goconst.Work{{Op: token.SUB, Rep: goconst.RepRat, Bits: 4002, Unary: true}}},
		{bin(cFloat64(1), QUO, cFloat64(3)), []goconst.Work{
			{Op: token.QUO, Rep: goconst.RepRat, Bits: 3},
			{Conv: goconst.ConvFloat64, Rep: goconst.RepRat, Bits: 3},
		}},
		{bin(cStr("foo"), LSS, cStr("bar")), []goconst.Work{{Op: token.LSS, Rep: goconst.RepString, Bits: 3}}},
		{bin(cStr("foo"), ADD, cStr("bar")), []goconst.Work{
			{Op: token.ADD, Rep: goconst.RepString, Bits: 3},
			{Conv: goconst.ConvString, Rep: goconst.RepString, Bits: 6},
		}},
		{bin(cRune('a'), ADD, cRune('b')), []goconst.Work{{Op: token.ADD, Rep: goconst.RepInt64}}},
	} {
		got = got[:0]
		_, ok := foldConstWith(func() goconst.Meter { return m }, nil, tc.x)
		require.True(t, ok)
		require.Equal(t, tc.want, got, "%v", tc.x)
	}
}

type workRecorder struct{ ws *[]goconst.Work }

func (r workRecorder) Charge(w goconst.Work) { *r.ws = append(*r.ws, w) }
