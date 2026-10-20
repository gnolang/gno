package gnolang

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"math"
	"math/big"
	"math/rand"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
)

// int256TestValues returns edge values plus seeded random ones, as exact
// integers in the range of the given kind.
func int256TestValues(signed bool) []*big.Int {
	lo, hi := big.NewInt(0), bigMaxUint256
	if signed {
		lo, hi = bigMinInt256, bigMaxInt256
	}
	vals := []*big.Int{
		big.NewInt(0), big.NewInt(1), big.NewInt(2), big.NewInt(3), big.NewInt(7), big.NewInt(10),
		new(big.Int).SetUint64(math.MaxUint64),
		new(big.Int).Add(new(big.Int).SetUint64(math.MaxUint64), big.NewInt(1)),
		new(big.Int).Lsh(big.NewInt(1), 128),
		new(big.Int).Lsh(big.NewInt(1), 254),
		new(big.Int).Rsh(hi, 1),
		hi, new(big.Int).Sub(hi, big.NewInt(1)),
		lo, new(big.Int).Add(lo, big.NewInt(1)),
	}
	if signed {
		vals = append(vals, big.NewInt(-1), big.NewInt(-2), big.NewInt(-7),
			big.NewInt(math.MinInt64), new(big.Int).Neg(new(big.Int).Lsh(big.NewInt(1), 128)))
	}
	r := rand.New(rand.NewSource(256))
	for range 40 {
		bits := r.Intn(257)
		v := new(big.Int).Rand(r, new(big.Int).Lsh(big.NewInt(1), uint(bits)))
		if signed {
			v.Sub(v, new(big.Int).Lsh(big.NewInt(1), uint(max(bits-1, 0))))
		}
		if v.Cmp(lo) >= 0 && v.Cmp(hi) <= 0 {
			vals = append(vals, v)
		}
	}
	return vals
}

// wrap256 reduces b modulo 2^256 into the range of the kind.
func wrap256(b *big.Int, signed bool) *big.Int {
	z := bigToBits256(b)
	return bits256ToBig(&z, signed)
}

func tv256(t Type, b *big.Int) TypedValue {
	tv := TypedValue{T: t}
	z := bigToBits256(b)
	tv.SetBits256(nil, &z)
	return tv
}

func TestBits256Arithmetic(t *testing.T) {
	for _, typ := range []Type{Int256Type, Uint256Type} {
		signed := typ == Int256Type
		vals := int256TestValues(signed)
		type binop struct {
			name string
			do   func(lv, rv *TypedValue) *Exception
			ref  func(a, b *big.Int) *big.Int // nil result: division by zero
		}
		ops := []binop{
			{"+", func(lv, rv *TypedValue) *Exception { addAssign(nil, lv, rv); return nil }, func(a, b *big.Int) *big.Int { return new(big.Int).Add(a, b) }},
			{"-", func(lv, rv *TypedValue) *Exception { subAssign(nil, lv, rv); return nil }, func(a, b *big.Int) *big.Int { return new(big.Int).Sub(a, b) }},
			{"*", func(lv, rv *TypedValue) *Exception { mulAssign(nil, lv, rv); return nil }, func(a, b *big.Int) *big.Int { return new(big.Int).Mul(a, b) }},
			{"/", func(lv, rv *TypedValue) *Exception { return quoAssign(nil, lv, rv) }, func(a, b *big.Int) *big.Int {
				if b.Sign() == 0 {
					return nil
				}
				return new(big.Int).Quo(a, b)
			}},
			{"%", func(lv, rv *TypedValue) *Exception { return remAssign(nil, lv, rv) }, func(a, b *big.Int) *big.Int {
				if b.Sign() == 0 {
					return nil
				}
				return new(big.Int).Rem(a, b)
			}},
			{"&", func(lv, rv *TypedValue) *Exception { bandAssign(nil, lv, rv); return nil }, func(a, b *big.Int) *big.Int { return new(big.Int).And(a, b) }},
			{"|", func(lv, rv *TypedValue) *Exception { borAssign(nil, lv, rv); return nil }, func(a, b *big.Int) *big.Int { return new(big.Int).Or(a, b) }},
			{"^", func(lv, rv *TypedValue) *Exception { xorAssign(nil, lv, rv); return nil }, func(a, b *big.Int) *big.Int { return new(big.Int).Xor(a, b) }},
			{"&^", func(lv, rv *TypedValue) *Exception { bandnAssign(nil, lv, rv); return nil }, func(a, b *big.Int) *big.Int { return new(big.Int).AndNot(a, b) }},
		}
		for _, op := range ops {
			for _, a := range vals {
				for _, b := range vals {
					lv, rv := tv256(typ, a), tv256(typ, b)
					exc := op.do(&lv, &rv)
					want := op.ref(a, b)
					if want == nil {
						require.NotNil(t, exc, "%s: %s %s %s should panic", typ, a, op.name, b)
						continue
					}
					require.Nil(t, exc)
					want = wrap256(want, signed)
					require.Equal(t, want.String(), lv.GetBig256().String(), "%s: %s %s %s", typ, a, op.name, b)
					if want.Sign() == 0 {
						require.Nil(t, lv.V, "zero must be stored as a nil V")
					}
					// comparisons
					c := a.Cmp(b)
					require.Equal(t, c == 0, isEql(nil, ptr(tv256(typ, a)), ptr(tv256(typ, b)), false))
					require.Equal(t, c < 0, isLss(nil, ptr(tv256(typ, a)), ptr(tv256(typ, b))))
					require.Equal(t, c <= 0, isLeq(nil, ptr(tv256(typ, a)), ptr(tv256(typ, b))))
					require.Equal(t, c > 0, isGtr(nil, ptr(tv256(typ, a)), ptr(tv256(typ, b))))
					require.Equal(t, c >= 0, isGeq(nil, ptr(tv256(typ, a)), ptr(tv256(typ, b))))
				}
			}
		}
		// shifts, including counts of 0, 255, 256 and beyond
		m := NewMachine("test", nil)
		defer m.Release()
		for _, a := range vals {
			for _, n := range []uint64{0, 1, 63, 64, 65, 127, 128, 200, 255, 256, 257, 1000, math.MaxUint64} {
				count := TypedValue{T: UintType}
				count.SetUint(n)

				lv := tv256(typ, a)
				shlAssign(m, &lv, &count)
				want := new(big.Int)
				if n < 512 {
					want = wrap256(new(big.Int).Lsh(a, uint(n)), signed)
				}
				require.Equal(t, want.String(), lv.GetBig256().String(), "%s: %s << %d", typ, a, n)

				lv = tv256(typ, a)
				shrAssign(m, &lv, &count)
				// big.Int.Rsh is arithmetic (floor) for negative values,
				// as Go's >> is for signed integers.
				want = new(big.Int).Rsh(a, uint(min(n, 1000)))
				require.Equal(t, want.String(), lv.GetBig256().String(), "%s: %s >> %d", typ, a, n)
			}
		}
	}
}

func ptr(tv TypedValue) *TypedValue { return &tv }

func TestBits256ShiftCount256(t *testing.T) {
	m := NewMachine("test", nil)
	defer m.Release()
	one := tv256(Uint256Type, big.NewInt(1))
	// A 256-bit count is not truncated: 2^64 + 1 shifts everything out.
	count := tv256(Uint256Type, new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), 64), big.NewInt(1)))
	shlAssign(m, &one, &count)
	assert.Nil(t, one.V)
	assert.Equal(t, uint64(math.MaxUint64), shiftCount(&count))
	small := tv256(Int256Type, big.NewInt(3))
	assert.Equal(t, uint64(3), shiftCount(&small))
}

func TestBits256Unary(t *testing.T) {
	m := NewMachine("test", nil)
	defer m.Release()
	for _, typ := range []Type{Int256Type, Uint256Type} {
		signed := typ == Int256Type
		for _, a := range int256TestValues(signed) {
			xv := tv256(typ, a)
			m.PushValue(xv)
			m.PushExpr(&UnaryExpr{Op: SUB})
			m.doOpUneg()
			got := m.PopValue()
			assert.Equal(t, wrap256(new(big.Int).Neg(a), signed).String(), got.GetBig256().String(), "-%s", a)

			m.PushValue(xv)
			m.PushExpr(&UnaryExpr{Op: XOR})
			m.doOpUxor()
			got = m.PopValue()
			assert.Equal(t, wrap256(new(big.Int).Not(a), signed).String(), got.GetBig256().String(), "^%s", a)
		}
	}
}

func TestBits256Conversions(t *testing.T) {
	type kindRange struct {
		t      Type
		signed bool
		bits   uint
	}
	ints := []kindRange{
		{IntType, true, 64}, {Int8Type, true, 8}, {Int16Type, true, 16}, {Int32Type, true, 32}, {Int64Type, true, 64},
		{UintType, false, 64}, {Uint8Type, false, 8}, {Uint16Type, false, 16}, {Uint32Type, false, 32}, {Uint64Type, false, 64},
		{Int256Type, true, 256}, {Uint256Type, false, 256},
	}
	truncate := func(b *big.Int, bits uint, signed bool) *big.Int {
		mod := new(big.Int).Lsh(big.NewInt(1), bits)
		r := new(big.Int).Mod(b, mod)
		if signed && r.Bit(int(bits-1)) == 1 {
			r.Sub(r, mod)
		}
		return r
	}
	for _, src := range []Type{Int256Type, Uint256Type} {
		for _, a := range int256TestValues(src == Int256Type) {
			for _, dst := range ints {
				tv := tv256(src, a)
				ConvertTo(nil, nil, &tv, dst.t, false)
				assert.Equal(t, truncate(a, dst.bits, dst.signed).String(), exactIntegerValue(&tv).String(),
					"%s(%s %s)", dst.t, src, a)
			}
		}
	}
	// From every integer kind into the 256-bit kinds: sign extension for
	// signed sources, zero extension for unsigned ones.
	for _, src := range ints[:10] {
		lo := new(big.Int).Neg(new(big.Int).Lsh(big.NewInt(1), src.bits-1))
		hi := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), src.bits-1), big.NewInt(1))
		if !src.signed {
			lo, hi = big.NewInt(0), new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), src.bits), big.NewInt(1))
		}
		for _, a := range []*big.Int{lo, hi, big.NewInt(0), big.NewInt(1)} {
			for _, dst := range []Type{Int256Type, Uint256Type} {
				tv := TypedValue{T: src.t}
				setExactInteger(&tv, a)
				ConvertTo(nil, nil, &tv, dst, false)
				assert.Equal(t, wrap256(a, dst == Int256Type).String(), tv.GetBig256().String(), "%s(%s %s)", dst, src.t, a)
			}
		}
	}
}

// setExactInteger sets a 64-bit-or-less integer tv from b.
func setExactInteger(tv *TypedValue, b *big.Int) {
	switch tv.T.Kind() {
	case IntKind:
		tv.SetInt(b.Int64())
	case Int8Kind:
		tv.SetInt8(int8(b.Int64()))
	case Int16Kind:
		tv.SetInt16(int16(b.Int64()))
	case Int32Kind:
		tv.SetInt32(int32(b.Int64()))
	case Int64Kind:
		tv.SetInt64(b.Int64())
	case UintKind:
		tv.SetUint(b.Uint64())
	case Uint8Kind:
		tv.SetUint8(uint8(b.Uint64()))
	case Uint16Kind:
		tv.SetUint16(uint16(b.Uint64()))
	case Uint32Kind:
		tv.SetUint32(uint32(b.Uint64()))
	case Uint64Kind:
		tv.SetUint64(b.Uint64())
	}
}

func TestBits256FloatConversions(t *testing.T) {
	f64 := func(f float64) TypedValue {
		tv := TypedValue{T: Float64Type}
		tv.SetFloat64(math.Float64bits(f))
		return tv
	}
	cases := []struct {
		f      float64
		dst    Type
		want   string
		exactC bool // valid as a constant conversion
	}{
		{0, Uint256Type, "0", true},
		{1.5, Uint256Type, "1", false},
		{-1.5, Int256Type, "-1", false},
		{1e30, Uint256Type, "1000000000000000019884624838656", true},
		{-1e30, Int256Type, "-1000000000000000019884624838656", true},
		{math.Ldexp(1, 255), Int256Type, "0", false}, // out of range
		{math.Ldexp(1, 255), Uint256Type, big2p255.String(), true},
		{-math.Ldexp(1, 255), Int256Type, bigMinInt256.String(), true},
		{math.Ldexp(1, 256), Uint256Type, "0", false},    // out of range
		{-1, Uint256Type, bigMaxUint256.String(), false}, // wraps like uint64(-1.0)
		{math.Inf(1), Uint256Type, "0", false},
		{math.NaN(), Int256Type, "0", false},
	}
	for _, c := range cases {
		tv := f64(c.f)
		ConvertTo(nil, nil, &tv, c.dst, false)
		assert.Equal(t, c.want, tv.GetBig256().String(), "%s(%v)", c.dst, c.f)

		tv = f64(c.f)
		if c.exactC {
			assert.NotPanics(t, func() { ConvertTo(nil, nil, &tv, c.dst, true) }, "const %s(%v)", c.dst, c.f)
		} else {
			assert.Panics(t, func() { ConvertTo(nil, nil, &tv, c.dst, true) }, "const %s(%v)", c.dst, c.f)
		}
	}

	// To floats: round to nearest, ties to even.
	toF := []struct {
		v    *big.Int
		src  Type
		want float64
	}{
		{bigMaxUint256, Uint256Type, math.Ldexp(1, 256)},
		{bigMinInt256, Int256Type, -math.Ldexp(1, 255)},
		{new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), 200), big.NewInt(1)), Uint256Type, math.Ldexp(1, 200)},
		// 2^53+1 is a tie between 2^53 and 2^53+2: ties to even gives 2^53.
		{new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), 53), big.NewInt(1)), Int256Type, math.Ldexp(1, 53)},
		{big.NewInt(-3), Int256Type, -3},
	}
	for _, c := range toF {
		tv := tv256(c.src, c.v)
		ConvertTo(nil, nil, &tv, Float64Type, false)
		assert.Equal(t, c.want, math.Float64frombits(tv.GetFloat64()), "float64(%s)", c.v)
	}
	// float32 overflow: +Inf at runtime, an error for constants.
	tv := tv256(Uint256Type, bigMaxUint256)
	ConvertTo(nil, nil, &tv, Float32Type, false)
	assert.True(t, math.IsInf(float64(math.Float32frombits(tv.GetFloat32())), 1))
	tv = tv256(Uint256Type, bigMaxUint256)
	assert.Panics(t, func() { ConvertTo(nil, nil, &tv, Float32Type, true) })
}

func TestBits256ConstConversionRange(t *testing.T) {
	// Typed constant conversions check the exact range.
	tv := tv256(Uint256Type, new(big.Int).Lsh(big.NewInt(1), 255))
	assert.PanicsWithValue(t, "constant 57896044618658097711785492504343953926634992332820282019728792003956564819968 overflows Int256Kind",
		func() { ConvertTo(nil, nil, &tv, Int256Type, true) })
	tv = tv256(Int256Type, big.NewInt(-1))
	assert.PanicsWithValue(t, "constant -1 overflows Uint256Kind", func() { ConvertTo(nil, nil, &tv, Uint256Type, true) })
	tv = tv256(Uint256Type, big.NewInt(256))
	assert.PanicsWithValue(t, "constant 256 overflows Uint8Kind", func() { ConvertTo(nil, nil, &tv, Uint8Type, true) })
	tv = TypedValue{T: Int8Type}
	tv.SetInt8(-1)
	assert.PanicsWithValue(t, "constant -1 overflows Uint256Kind", func() { ConvertTo(nil, nil, &tv, Uint256Type, true) })

	// Untyped constants: [0, 2^256-1] and [-2^255, 2^255-1].
	conv := func(b *big.Int, t Type) func() {
		return func() {
			tv := TypedValue{T: UntypedBigintType, V: BigintValue{V: b}}
			ConvertUntypedTo(&tv, t)
		}
	}
	assert.NotPanics(t, conv(bigMaxUint256, Uint256Type))
	assert.PanicsWithValue(t, "bigint overflows target kind", conv(big2p256, Uint256Type))
	assert.PanicsWithValue(t, "bigint underflows target kind", conv(big.NewInt(-1), Uint256Type))
	assert.NotPanics(t, conv(bigMinInt256, Int256Type))
	assert.NotPanics(t, conv(bigMaxInt256, Int256Type))
	assert.PanicsWithValue(t, "bigint overflows target kind", conv(big2p255, Int256Type))
	assert.PanicsWithValue(t, "bigint underflows target kind", conv(new(big.Int).Sub(bigMinInt256, big.NewInt(1)), Int256Type))
}

func TestBits256MapKey(t *testing.T) {
	keys := map[MapKey]string{}
	for _, typ := range []Type{Int256Type, Uint256Type} {
		for _, a := range int256TestValues(typ == Int256Type) {
			tv := tv256(typ, a)
			k, isNaN := tv.ComputeMapKey(nil, nil, false)
			require.False(t, isNaN)
			id := typ.String() + ":" + a.String()
			if prev, ok := keys[k]; ok {
				require.Equal(t, prev, id, "map key collision")
			}
			keys[k] = id
			// The same value computed differently has the same key.
			tv2 := tv256(typ, new(big.Int).Set(a))
			k2, _ := tv2.ComputeMapKey(nil, nil, false)
			require.Equal(t, k, k2)
		}
	}
}

func TestBits256Amino(t *testing.T) {
	for _, typ := range []Type{Int256Type, Uint256Type} {
		for _, a := range int256TestValues(typ == Int256Type) {
			// Inside a TypedValue, as realm state stores it.
			tv := tv256(typ, a)
			bz := amino.MustMarshal(tv)
			var tv2 TypedValue
			amino.MustUnmarshal(bz, &tv2)
			require.Equal(t, a.String(), tv2.GetBig256().String())
			require.Equal(t, tv, tv2)
			// Inside an interface and a composite.
			av := &ArrayValue{List: []TypedValue{tv, {T: typ}}}
			bz = amino.MustMarshalAny(av)
			var back Value
			amino.MustUnmarshalAny(bz, &back)
			require.Equal(t, av.List, back.(*ArrayValue).List)
		}
	}
}

// TestBits256AminoJSON pins the external contract: the value is a decimal
// string, signed for int256, in a type of its own; zero has no V.
func TestBits256AminoJSON(t *testing.T) {
	cases := []struct {
		typ  Type
		v    *big.Int
		json string
	}{
		{Int256Type, big.NewInt(128), `{"T":{"@type":"/gno.PrimitiveType","value":"2097152"},"V":{"@type":"/gno.Int256Value","value":"128"}}`},
		{Int256Type, big.NewInt(-128), `{"T":{"@type":"/gno.PrimitiveType","value":"2097152"},"V":{"@type":"/gno.Int256Value","value":"-128"}}`},
		{Int256Type, bigMinInt256, `{"T":{"@type":"/gno.PrimitiveType","value":"2097152"},"V":{"@type":"/gno.Int256Value","value":"` + bigMinInt256.String() + `"}}`},
		{Uint256Type, bigMaxUint256, `{"T":{"@type":"/gno.PrimitiveType","value":"4194304"},"V":{"@type":"/gno.Uint256Value","value":"` + bigMaxUint256.String() + `"}}`},
		{Uint256Type, big.NewInt(0), `{"T":{"@type":"/gno.PrimitiveType","value":"4194304"}}`},
	}
	for _, c := range cases {
		tv := tv256(c.typ, c.v)
		bz := amino.MustMarshalJSON(tv)
		require.Equal(t, c.json, string(bz))
		var tv2 TypedValue
		amino.MustUnmarshalJSON(bz, &tv2)
		require.Equal(t, c.v.String(), tv2.GetBig256().String())
	}
}

// TestBits256CanonicalZero checks that zero has one encoding: a boxed zero is
// neither produced nor accepted, nor is any non-canonical decimal.
func TestBits256CanonicalZero(t *testing.T) {
	_, err := amino.Marshal(TypedValue{T: Uint256Type, V: Uint256Value{}})
	require.Error(t, err)
	_, err = Int256Value{}.MarshalAmino()
	require.Error(t, err)

	// An empty payload, which is how a boxed zero would encode, is refused
	// on decode, so it cannot appear inside a TypedValue, an interface or a
	// composite.
	require.Error(t, (&Uint256Value{}).UnmarshalBinary2(amino.NewCodec(), nil, 0))
	require.Error(t, (&Int256Value{}).UnmarshalBinary2(amino.NewCodec(), nil, 0))
	for _, s := range []string{"", "0", "00", "05", "-5", "+5", "1x"} {
		var v Uint256Value
		require.Error(t, v.UnmarshalAmino(s), "uint256 %q", s)
	}
	for _, s := range []string{"", "0", "-0", "05", "+5", "57896044618658097711785492504343953926634992332820282019728792003956564819968"} {
		var v Int256Value
		require.Error(t, v.UnmarshalAmino(s), "int256 %q", s)
	}
	var v Int256Value
	require.NoError(t, v.UnmarshalAmino("-57896044618658097711785492504343953926634992332820282019728792003956564819968"))

	// A zero default and the result of arithmetic share one key and encoding.
	zero := TypedValue{T: Uint256Type}
	a, b := tv256(Uint256Type, big.NewInt(9)), tv256(Uint256Type, big.NewInt(9))
	subAssign(nil, &a, &b)
	require.Nil(t, a.V)
	require.Equal(t, amino.MustMarshal(zero), amino.MustMarshal(a))
}

func TestBits256ConvertGetInt(t *testing.T) {
	for _, c := range []struct {
		v    *big.Int
		typ  Type
		want int64
	}{
		{big.NewInt(5), Uint256Type, 5},
		{big.NewInt(-5), Int256Type, -5},
		{new(big.Int).Lsh(big.NewInt(1), 64), Uint256Type, math.MaxInt64}, // saturates, not truncates to 0
		{new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), 64), big.NewInt(1)), Uint256Type, math.MaxInt64},
		{bigMinInt256, Int256Type, -1},
	} {
		tv := tv256(c.typ, c.v)
		assert.Equal(t, c.want, tv.ConvertGetInt(), "%s", c.v)
	}
}

// TestGoTypesSizes pins what go/types accepts for the uint256 and int256
// shims: uint256 constants in [0, 2^256-1], and int256 constants in the int64
// range only, which is a language restriction.
func TestGoTypesSizes(t *testing.T) {
	cases := []struct {
		src string
		ok  bool
	}{
		{`var x uint256 = 1<<256 - 1`, true},
		{`var x uint256 = 1 << 256`, false},
		{`var x uint256 = -1`, false},
		{`var x = ^uint256(0)`, true},
		{`var x int256 = -1`, true},
		{`var x int256 = -9223372036854775808`, true},
		{`var x int256 = 9223372036854775808`, false},
		{`var x int = 9223372036854775807`, true},
		{`var x int = 9223372036854775808`, false},
		{`var x int8 = -128`, true},
		{`var x = uint256(1) + int256(1)`, false},             // distinct types
		{`const c int256 = 1 << 100`, false},                  // the int256 restriction
		{`const f = 1e30; var x = int256(f)`, false},          // also for float constants
		{`var u uint256 = 1 << 200; var x = int256(u)`, true}, // the workaround
	}
	for _, c := range cases {
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, "f.go", "package p\ntype uint256 uintptr\ntype int256 int\n"+c.src, 0)
		require.NoError(t, err)
		var errs []error
		conf := types.Config{
			Sizes: GoTypesSizes,
			Error: func(err error) { errs = append(errs, err) },
		}
		conf.Check("p", fset, []*ast.File{f}, nil)
		assert.Equal(t, c.ok, len(errs) == 0, "%s: %v", c.src, errs)
	}
}

func TestBits256Gonative(t *testing.T) {
	for _, typ := range []Type{Int256Type, Uint256Type} {
		for _, a := range int256TestValues(typ == Int256Type) {
			tv := tv256(typ, a)
			rv := Gno2GoValue(&tv, reflect.Value{})
			if typ == Int256Type {
				require.Equal(t, goInt256Type, rv.Type())
			} else {
				require.Equal(t, goUint256Type, rv.Type())
			}
			back := Go2GnoValue(nil, nil, rv)
			require.Equal(t, typ, back.T)
			require.Equal(t, a.String(), back.GetBig256().String())
			if a.Sign() == 0 {
				require.Nil(t, back.V)
			}
		}
	}
	// Distinct Go types, although both are [4]uint64 underneath.
	require.Equal(t, Uint256Type, go2GnoType(reflect.TypeFor[Uint256]()))
	require.Equal(t, Int256Type, go2GnoType(reflect.TypeFor[Int256]()))
	require.Equal(t, goUint256Type, gno2GoType(Uint256Type))
	require.Equal(t, goInt256Type, gno2GoType(Int256Type))
}

// TestBits256CopyAccounting checks the logical, per-occurrence accounting:
// every slot holding a non-zero 256-bit value is charged one box, whether
// the box was built for it or shared by a copy, so allocation, the GC
// recount and the charge on reload agree.
func TestBits256CopyAccounting(t *testing.T) {
	const n = 1000
	alloc := NewAllocator(math.MaxInt64)
	x := tv256(Uint256Type, bigMaxUint256)
	av := &ArrayValue{List: make([]TypedValue, n)}
	_, before := alloc.Status()
	for i := range av.List {
		av.List[i] = x.Copy(alloc)
	}
	_, after := alloc.Status()
	require.Equal(t, int64(n*allocBits256), after-before, "allocation")

	var recount int64
	for i := range av.List {
		recount += av.List[i].V.GetShallowSize()
	}
	require.Equal(t, int64(n*allocBits256), recount, "GC recount")
	require.Equal(t, int64(n*allocBits256), internalRefSize(av), "charge on reload")

	// Persisted and reloaded, each slot holds its own value.
	bz := amino.MustMarshalAny(av)
	var back Value
	amino.MustUnmarshalAny(bz, &back)
	for _, tv := range back.(*ArrayValue).List {
		require.Equal(t, bigMaxUint256.String(), tv.GetBig256().String())
	}
}

// TestBits256CanonicalZeroJSONNull checks that a typed JSON null payload,
// which amino decodes as a boxed zero without calling UnmarshalAmino, is
// rejected at the enclosing value, also nested in interface slots and inline
// composites.
func TestBits256CanonicalZeroJSONNull(t *testing.T) {
	for _, c := range []struct {
		name, at string
	}{{"Int256Value", "2097152"}, {"Uint256Value", "4194304"}} {
		slot := `{"T":{"@type":"/gno.PrimitiveType","value":"` + c.at + `"},"V":{"@type":"/gno.` + c.name + `","value":null}}`
		var tv TypedValue
		require.NoError(t, amino.UnmarshalJSON([]byte(slot), &tv))
		require.NotNil(t, tv.V, "amino builds a boxed zero from null")

		// In an array slot, and in an interface-typed slot of a struct
		// held inline by an array.
		for _, js := range []string{
			`{"@type":"/gno.ArrayValue","List":[` + slot + `]}`,
			`{"@type":"/gno.ArrayValue","List":[{"T":{"@type":"/gno.InterfaceType"},"V":{"@type":"/gno.StructValue","Fields":[` + slot + `]}}]}`,
		} {
			var v Value
			require.NoError(t, amino.UnmarshalJSON([]byte(js), &v), js)
			require.Error(t, CheckCanonical256(v), js)
			require.Panics(t, func() { mustBeCanonical256(v) })
		}
	}
	// A canonical value passes.
	av := &ArrayValue{List: []TypedValue{tv256(Int256Type, big.NewInt(-1)), {T: Uint256Type}}}
	require.NoError(t, CheckCanonical256(av))
}
