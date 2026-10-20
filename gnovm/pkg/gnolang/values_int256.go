package gnolang

// Native 256-bit integers: int256 and uint256.
//
// A value lives in TypedValue.V as an Int256Value or a Uint256Value, each
// holding the 256 bits (two's complement for int256). A nil V is the zero
// value, and every operation stores a zero result as nil, so each value has
// exactly one representation; the amino encoding rejects a boxed zero.
//
// The values are fixed-size arrays held by value. Copying a TypedValue
// copies the interface, which shares the boxed array, but nothing ever writes
// through it: every operation builds a new value. Assignment can therefore
// never alias. See gnovm/adr/prxxxx_native_int256.md.

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"unicode/utf8"

	"github.com/holiman/uint256"

	"github.com/gnolang/gno/gnovm/pkg/gnolang/internal/softfloat"
)

// Uint256Value is the value of a non-zero uint256.
type Uint256Value struct {
	V uint256.Int
}

// Int256Value is the value of a non-zero int256, in two's complement.
type Int256Value struct {
	V uint256.Int
}

func (Uint256Value) assertValue() {}
func (Int256Value) assertValue()  {}

var (
	_ Value = Uint256Value{}
	_ Value = Int256Value{}
)

func (v Uint256Value) String() string { return v.V.Dec() }
func (v Int256Value) String() string  { return dec256(&v.V, true) }

func (v Uint256Value) DeepFill(store Store) Value { return v }
func (v Int256Value) DeepFill(store Store) Value  { return v }

func (v Uint256Value) GetShallowSize() int64 { return allocBits256 }
func (v Int256Value) GetShallowSize() int64  { return allocBits256 }

func (v Uint256Value) VisitAssociated(vis Visitor) (stop bool) { return false }
func (v Int256Value) VisitAssociated(vis Visitor) (stop bool)  { return false }

// MarshalAmino encodes the value as a canonical decimal string, in binary and
// in JSON alike: digits with no leading zero, and a "-" for a negative int256.
// Zero is never boxed (it is a nil V), so a boxed zero is refused.
func (v Uint256Value) MarshalAmino() (string, error) {
	if v.V.IsZero() {
		return "", errBoxedZero
	}
	return v.V.Dec(), nil
}

func (v Int256Value) MarshalAmino() (string, error) {
	if v.V.IsZero() {
		return "", errBoxedZero
	}
	return dec256(&v.V, true), nil
}

// UnmarshalAmino accepts only the form MarshalAmino produces, so a value has a
// single encoding in realm state.
func (v *Uint256Value) UnmarshalAmino(s string) error {
	z, err := parseCanonical256(s, false)
	v.V = z
	return err
}

func (v *Int256Value) UnmarshalAmino(s string) error {
	z, err := parseCanonical256(s, true)
	v.V = z
	return err
}

var errBoxedZero = errors.New("int256/uint256: zero must be a nil value, not a boxed zero")

// CheckCanonical256 returns an error if v, a decoded value, holds a boxed
// zero 256-bit value in any of its inline slots, recursively through inline
// composites. The binary decoder already rejects one (UnmarshalAmino refuses
// an empty payload), but amino builds a JSON null, or an absent Any payload,
// as the zero value without calling UnmarshalAmino; this check closes that
// path at the enclosing value. Objects are checked as the store loads them.
func CheckCanonical256(v Value) error {
	var err error
	var walk func(v Value)
	slot := func(tv *TypedValue) {
		switch x := tv.V.(type) {
		case Int256Value:
			if x.V.IsZero() {
				err = errBoxedZero
			}
		case Uint256Value:
			if x.V.IsZero() {
				err = errBoxedZero
			}
		default:
			walk(x)
		}
	}
	walk = func(v Value) {
		if err != nil {
			return
		}
		switch v := v.(type) {
		case *Block:
			for i := range v.Values {
				slot(&v.Values[i])
			}
		case *ArrayValue:
			for i := range v.List {
				slot(&v.List[i])
			}
		case *StructValue:
			for i := range v.Fields {
				slot(&v.Fields[i])
			}
		case *MapValue:
			for cur := v.List.Head; cur != nil; cur = cur.Next {
				slot(&cur.Key)
				slot(&cur.Value)
			}
		case *BoundMethodValue:
			slot(&v.Receiver)
		case *HeapItemValue:
			slot(&v.Value)
		case *FuncValue:
			for i := range v.Captures {
				slot(&v.Captures[i])
			}
		case PointerValue:
			if v.TV != nil {
				slot(v.TV)
			}
		}
	}
	walk(v)
	return err
}

func mustBeCanonical256(v Value) {
	if err := CheckCanonical256(v); err != nil {
		panic(err)
	}
}

// dec256 formats z in decimal, as signed two's complement if signed.
func dec256(z *uint256.Int, signed bool) string {
	if signed && z.Sign() < 0 {
		var abs uint256.Int
		abs.Neg(z)
		return "-" + abs.Dec()
	}
	return z.Dec()
}

// parseCanonical256 parses the canonical non-zero decimal form of an int256
// (signed) or uint256: an optional "-" (int256 only), then digits with no
// leading zero, within the type's range.
func parseCanonical256(s string, signed bool) (z uint256.Int, err error) {
	digits := s
	neg := signed && len(s) > 0 && s[0] == '-'
	if neg {
		digits = s[1:]
	}
	if digits == "" || digits[0] == '0' {
		return z, fmt.Errorf("int256/uint256: non-canonical decimal %q", s)
	}
	for i := 0; i < len(digits); i++ {
		if c := digits[i]; c < '0' || c > '9' {
			return z, fmt.Errorf("int256/uint256: invalid decimal %q", s)
		}
	}
	if err := z.SetFromDecimal(digits); err != nil {
		return z, fmt.Errorf("int256/uint256: %q out of range", s)
	}
	if signed {
		var limit uint256.Int
		limit.Lsh(uint256.NewInt(1), 255)
		if z.Gt(&limit) || (!neg && z.Eq(&limit)) {
			return z, fmt.Errorf("int256: %q out of range", s)
		}
		if neg {
			z.Neg(&z)
		}
	}
	return z, nil
}

// ----------------------------------------
// Kind helpers

func is256Kind(k Kind) bool {
	return k == Int256Kind || k == Uint256Kind
}

func is256Type(t Type) bool {
	if t == nil {
		return false
	}
	switch baseOf(t) {
	case Int256Type, Uint256Type:
		return true
	}
	return false
}

// ----------------------------------------
// TypedValue accessors

// GetBits256 returns the raw bits of an int256 or uint256 value.
func (tv *TypedValue) GetBits256() (z uint256.Int) {
	if debug {
		if tv.T != nil && !is256Kind(tv.T.Kind()) {
			panic(fmt.Sprintf(
				"TypedValue.GetBits256() on type %s",
				tv.T.String(),
			))
		}
	}
	switch v := tv.V.(type) {
	case nil:
		return z
	case Uint256Value:
		return v.V
	default:
		return v.(Int256Value).V
	}
}

// SetBits256 stores z, charging alloc for the boxed value. Zero is stored as
// a nil V.
func (tv *TypedValue) SetBits256(alloc *Allocator, z *uint256.Int) {
	if debug {
		if !is256Kind(tv.T.Kind()) {
			panic(fmt.Sprintf(
				"TypedValue.SetBits256() on type %s",
				tv.T.String(),
			))
		}
	}
	tv.ClearNum()
	if z.IsZero() {
		tv.V = nil
		return
	}
	alloc.Allocate(allocBits256)
	if tv.T.Kind() == Int256Kind {
		tv.V = Int256Value{V: *z}
	} else {
		tv.V = Uint256Value{V: *z}
	}
}

// GetBig256 returns the mathematical value of an int256 or uint256 value.
func (tv *TypedValue) GetBig256() *big.Int {
	z := tv.GetBits256()
	return bits256ToBig(&z, tv.T.Kind() == Int256Kind)
}

// ----------------------------------------
// Conversions between bits and math/big

var (
	big2p255      = new(big.Int).Lsh(big.NewInt(1), 255)
	big2p256      = new(big.Int).Lsh(big.NewInt(1), 256)
	bigMaxUint256 = new(big.Int).Sub(big2p256, big.NewInt(1))
	bigMaxInt256  = new(big.Int).Sub(big2p255, big.NewInt(1))
	bigMinInt256  = new(big.Int).Neg(big2p255)
)

func bits256ToBig(z *uint256.Int, signed bool) *big.Int {
	if signed && z.Sign() < 0 {
		var abs uint256.Int
		abs.Neg(z)
		b := abs.ToBig()
		return b.Neg(b)
	}
	return z.ToBig()
}

// bigToBits256 returns b modulo 2^256, as two's complement bits.
func bigToBits256(b *big.Int) (z uint256.Int) {
	z.SetFromBig(b)
	return z
}

// bigFits256 reports whether b is in the range of int256 (signed) or uint256.
func bigFits256(b *big.Int, signed bool) bool {
	if signed {
		return b.Cmp(bigMinInt256) >= 0 && b.Cmp(bigMaxInt256) <= 0
	}
	return b.Sign() >= 0 && b.Cmp(bigMaxUint256) <= 0
}

func bits256FromInt64(v int64) (z uint256.Int) {
	z.SetUint64(uint64(v))
	if v < 0 {
		// sign-extend
		z[1], z[2], z[3] = math.MaxUint64, math.MaxUint64, math.MaxUint64
	}
	return z
}

// bits256FromFloat64 converts the float64 bits u to a 256-bit integer,
// truncating toward zero. The Go spec leaves out-of-range float-to-integer
// conversions implementation-defined; Gno defines them so that every node
// gets the same answer, in the same spirit as softfloat's 64-bit conversions:
//
//   - NaN and ±Inf convert to 0;
//   - int256: a truncated value outside [-2^255, 2^255) converts to 0;
//   - uint256: a truncated value in [-2^255, 0) wraps modulo 2^256 (as
//     uint64(-1.0) does), and one outside [-2^255, 2^256) converts to 0.
//
// ok reports whether the input was an exact integer in the target's range;
// constant conversions require it.
func bits256FromFloat64(u uint64, signed bool) (z uint256.Int, ok bool) {
	f := math.Float64frombits(u)
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return z, false
	}
	bf := new(big.Float).SetFloat64(f)
	bi, acc := bf.Int(nil)
	exact := acc == big.Exact
	if bigFits256(bi, signed) {
		return bigToBits256(bi), exact
	}
	if !signed && bi.Sign() < 0 && bi.Cmp(bigMinInt256) >= 0 {
		return bigToBits256(bi), false
	}
	return z, false
}

// bits256ToFloat64 rounds the value to the nearest float64 (ties to even),
// as Go does for integer-to-float conversions.
func bits256ToFloat64(z *uint256.Int, signed bool) uint64 {
	f, _ := new(big.Float).SetInt(bits256ToBig(z, signed)).Float64()
	return math.Float64bits(f)
}

// bits256ToFloat32 rounds the value to the nearest float32 (ties to even). A
// magnitude beyond float32's range becomes ±Inf, as in Go.
func bits256ToFloat32(z *uint256.Int, signed bool) uint32 {
	f, _ := new(big.Float).SetInt(bits256ToBig(z, signed)).Float32()
	return math.Float32bits(f)
}

func runeStrFromBig(b *big.Int) string {
	if !b.IsInt64() {
		return string(utf8.RuneError)
	}
	return runeStrFromInt64(b.Int64())
}

// exactIntegerValue64 returns the value of a 64-bit-or-narrower integer tv,
// as an int64 for signed kinds and as the uint64 bits for unsigned ones.
func exactIntegerValue64(tv *TypedValue) int64 {
	switch tv.T.Kind() {
	case IntKind:
		return tv.GetInt()
	case Int8Kind:
		return int64(tv.GetInt8())
	case Int16Kind:
		return int64(tv.GetInt16())
	case Int32Kind:
		return int64(tv.GetInt32())
	case Int64Kind:
		return tv.GetInt64()
	case UintKind:
		return int64(tv.GetUint())
	case Uint8Kind:
		return int64(tv.GetUint8())
	case Uint16Kind:
		return int64(tv.GetUint16())
	case Uint32Kind:
		return int64(tv.GetUint32())
	case Uint64Kind:
		return int64(tv.GetUint64())
	default:
		panic(fmt.Sprintf("exactIntegerValue64: unexpected kind %s", tv.T.Kind()))
	}
}

// exactIntegerValue returns the mathematical value of an integer-kinded tv.
func exactIntegerValue(tv *TypedValue) *big.Int {
	switch tv.T.Kind() {
	case IntKind:
		return big.NewInt(tv.GetInt())
	case Int8Kind:
		return big.NewInt(int64(tv.GetInt8()))
	case Int16Kind:
		return big.NewInt(int64(tv.GetInt16()))
	case Int32Kind:
		return big.NewInt(int64(tv.GetInt32()))
	case Int64Kind:
		return big.NewInt(tv.GetInt64())
	case UintKind:
		return new(big.Int).SetUint64(tv.GetUint())
	case Uint8Kind:
		return new(big.Int).SetUint64(uint64(tv.GetUint8()))
	case Uint16Kind:
		return new(big.Int).SetUint64(uint64(tv.GetUint16()))
	case Uint32Kind:
		return new(big.Int).SetUint64(uint64(tv.GetUint32()))
	case Uint64Kind:
		return new(big.Int).SetUint64(tv.GetUint64())
	case Int256Kind, Uint256Kind:
		return tv.GetBig256()
	case BigintKind:
		return tv.GetBigInt()
	default:
		panic(fmt.Sprintf("exactIntegerValue: unexpected kind %s", tv.T.Kind()))
	}
}

var (
	bigMaxUint    = new(big.Int).SetUint64(math.MaxUint64)
	intKindRanges = map[Kind][2]*big.Int{
		IntKind:     {big.NewInt(math.MinInt64), big.NewInt(math.MaxInt64)},
		Int8Kind:    {big.NewInt(math.MinInt8), big.NewInt(math.MaxInt8)},
		Int16Kind:   {big.NewInt(math.MinInt16), big.NewInt(math.MaxInt16)},
		Int32Kind:   {big.NewInt(math.MinInt32), big.NewInt(math.MaxInt32)},
		Int64Kind:   {big.NewInt(math.MinInt64), big.NewInt(math.MaxInt64)},
		UintKind:    {big.NewInt(0), bigMaxUint},
		Uint8Kind:   {big.NewInt(0), big.NewInt(math.MaxUint8)},
		Uint16Kind:  {big.NewInt(0), big.NewInt(math.MaxUint16)},
		Uint32Kind:  {big.NewInt(0), big.NewInt(math.MaxUint32)},
		Uint64Kind:  {big.NewInt(0), bigMaxUint},
		Int256Kind:  {bigMinInt256, bigMaxInt256},
		Uint256Kind: {big.NewInt(0), bigMaxUint256},
	}
)

// integerKindFits reports whether b is representable by the typed integer
// kind k.
func integerKindFits(b *big.Int, k Kind) bool {
	r, ok := intKindRanges[k]
	if !ok {
		panic(fmt.Sprintf("integerKindFits: unexpected kind %s", k))
	}
	return b.Cmp(r[0]) >= 0 && b.Cmp(r[1]) <= 0
}

// ----------------------------------------
// Conversion (ConvertTo) for the 256-bit kinds

// convert256 implements ConvertTo when the source or the target kind is
// int256 or uint256 (and they differ). See ConvertTo for the contract.
func convert256(alloc *Allocator, tv *TypedValue, t Type, isConst bool) {
	tvk := tv.T.Kind()
	k := t.Kind()

	failCannot := func() {
		panic(fmt.Sprintf("cannot convert %s to %s", tvk.String(), k.String()))
	}

	// Source: compute the 256 bits, and for constants the exact value.
	var bits uint256.Int
	var exact *big.Int // nil for a non-integral or non-finite float source
	switch tvk {
	case IntKind, Int8Kind, Int16Kind, Int32Kind, Int64Kind:
		bits = bits256FromInt64(exactIntegerValue64(tv))
	case UintKind, Uint8Kind, Uint16Kind, Uint32Kind, Uint64Kind:
		bits.SetUint64(uint64(exactIntegerValue64(tv)))
	case Int256Kind, Uint256Kind:
		bits = tv.GetBits256()
	case Float32Kind, Float64Kind:
		if !is256Kind(k) {
			failCannot()
		}
		u := tv.GetFloat64Bits()
		var ok bool
		bits, ok = bits256FromFloat64(u, k == Int256Kind)
		if isConst && !ok {
			panic(fmt.Sprintf("cannot convert constant of type %s to %s", tvk, k))
		}
		tv.T = t
		tv.SetBits256(alloc, &bits)
		return
	default:
		failCannot()
	}

	// The exact value is needed only for the constant range check and for
	// the string conversion; runtime integer conversions stay on the bits.
	if isConst || k == StringKind {
		exact = exactIntegerValue(tv)
	}
	if isConst && isIntegerKind(k) && !integerKindFits(exact, k) {
		panic(fmt.Sprintf("constant %s overflows %s", exact.String(), k))
	}

	if is256Kind(k) {
		tv.T = t
		tv.SetBits256(alloc, &bits)
		return
	}
	if k != StringKind {
		tv.T, tv.V = t, nil
	}
	switch k {
	case IntKind:
		tv.SetInt(int64(bits[0]))
	case Int8Kind:
		tv.SetInt8(int8(bits[0]))
	case Int16Kind:
		tv.SetInt16(int16(bits[0]))
	case Int32Kind:
		tv.SetInt32(int32(bits[0]))
	case Int64Kind:
		tv.SetInt64(int64(bits[0]))
	case UintKind:
		tv.SetUint(bits[0])
	case Uint8Kind:
		tv.SetUint8(uint8(bits[0]))
	case Uint16Kind:
		tv.SetUint16(uint16(bits[0]))
	case Uint32Kind:
		tv.SetUint32(uint32(bits[0]))
	case Uint64Kind:
		tv.SetUint64(bits[0])
	case Float32Kind:
		f := bits256ToFloat32(&bits, tvk == Int256Kind)
		if isConst && f&^softfloat.NegZero32 == softfloat.Inf32 {
			panic(fmt.Sprintf("cannot convert constant of type %s to %s", tvk, k))
		}
		tv.SetFloat32(f)
	case Float64Kind:
		f := bits256ToFloat64(&bits, tvk == Int256Kind)
		tv.SetFloat64(f)
	case StringKind:
		tv.V = alloc.NewString(runeStrFromBig(exact))
		tv.T = t
		tv.ClearNum()
	default:
		failCannot()
	}
}

// GetFloat64Bits returns the float64 bits of a float32 or float64 value,
// widening float32 exactly (in software, like all Gno float arithmetic).
func (tv *TypedValue) GetFloat64Bits() uint64 {
	if tv.T.Kind() == Float32Kind {
		return softfloat.F32to64(tv.GetFloat32())
	}
	return tv.GetFloat64()
}

// convertUntypedBigintTo256 sets dst to the untyped constant bi as an int256 or
// uint256 of type t, rejecting any value outside the type's range.
func convertUntypedBigintTo256(dst *TypedValue, bi *big.Int, t Type) {
	signed := t.Kind() == Int256Kind
	if !bigFits256(bi, signed) {
		if bi.Sign() < 0 {
			panic("bigint underflows target kind")
		}
		panic("bigint overflows target kind")
	}
	bits := bigToBits256(bi)
	dst.T = t
	// Constants are created by the preprocessor, outside any transaction
	// allocator, as untyped bigints are.
	dst.SetBits256(nil, &bits)
}

// ----------------------------------------
// Arithmetic

// bits256Binary applies op to the bits of lv and rv and stores the result in
// lv.
func bits256Binary(alloc *Allocator, lv, rv *TypedValue, op func(z, x, y *uint256.Int)) {
	x, y := lv.GetBits256(), rv.GetBits256()
	var z uint256.Int
	op(&z, &x, &y)
	lv.SetBits256(alloc, &z)
}

func bits256Add(z, x, y *uint256.Int) { z.Add(x, y) }
func bits256Sub(z, x, y *uint256.Int) { z.Sub(x, y) }
func bits256Mul(z, x, y *uint256.Int) { z.Mul(x, y) }
func bits256And(z, x, y *uint256.Int) { z.And(x, y) }
func bits256Or(z, x, y *uint256.Int)  { z.Or(x, y) }
func bits256Xor(z, x, y *uint256.Int) { z.Xor(x, y) }
func bits256AndNot(z, x, y *uint256.Int) {
	var ny uint256.Int
	ny.Not(y)
	z.And(x, &ny)
}

// bits256Quo and bits256Rem truncate toward zero as Go does. The divisor is
// already known to be non-zero. For int256, MinInt256 / -1 wraps to
// MinInt256 and MinInt256 % -1 is 0, as in Go.
func bits256Quo(signed bool) func(z, x, y *uint256.Int) {
	if signed {
		return func(z, x, y *uint256.Int) { z.SDiv(x, y) }
	}
	return func(z, x, y *uint256.Int) { z.Div(x, y) }
}

func bits256Rem(signed bool) func(z, x, y *uint256.Int) {
	if signed {
		return func(z, x, y *uint256.Int) { z.SMod(x, y) }
	}
	return func(z, x, y *uint256.Int) { z.Mod(x, y) }
}

// bits256Shl shifts left; counts of 256 or more give 0.
func bits256Shl(alloc *Allocator, lv *TypedValue, shift uint64) {
	x := lv.GetBits256()
	var z uint256.Int
	if shift < 256 {
		z.Lsh(&x, uint(shift))
	}
	lv.SetBits256(alloc, &z)
}

// bits256Shr shifts right: arithmetically for int256, logically for uint256.
// Counts of 256 or more give 0, or -1 for a negative int256.
func bits256Shr(alloc *Allocator, lv *TypedValue, shift uint64) {
	x := lv.GetBits256()
	var z uint256.Int
	n := uint(min(shift, 256))
	if lv.T.Kind() == Int256Kind {
		z.SRsh(&x, n)
	} else if n < 256 {
		z.Rsh(&x, n)
	}
	lv.SetBits256(alloc, &z)
}

// bits256Cmp compares two values of the same 256-bit kind.
func bits256Cmp(lv, rv *TypedValue) int {
	x, y := lv.GetBits256(), rv.GetBits256()
	if lv.T.Kind() != Int256Kind {
		return x.Cmp(&y)
	}
	switch {
	case x.Eq(&y):
		return 0
	case x.Slt(&y):
		return -1
	default:
		return 1
	}
}

// shiftCount returns the shift count held by rv, which is a uint or, for a
// 256-bit count, any non-negative 256-bit value. Counts that do not fit in a
// uint64 saturate: every such shift moves all bits out.
func shiftCount(rv *TypedValue) uint64 {
	if is256Kind(rv.T.Kind()) {
		z := rv.GetBits256()
		if !z.IsUint64() {
			return math.MaxUint64
		}
		return z.Uint64()
	}
	return rv.GetUint()
}

// convertGetInt256 converts a 256-bit index or length to an int. A value
// outside the int64 range saturates, so that bounds checks reject it instead
// of silently using its truncated low bits.
func convertGetInt256(tv *TypedValue) int64 {
	z := tv.GetBits256()
	if tv.T.Kind() == Int256Kind && z.Sign() < 0 {
		if z[1]&z[2]&z[3] == math.MaxUint64 && int64(z[0]) < 0 {
			return int64(z[0]) // in the int64 range
		}
		return -1
	}
	if !z.IsUint64() || z.Uint64() > math.MaxInt64 {
		return math.MaxInt64
	}
	return int64(z.Uint64())
}

// sprint256 formats an int256 or uint256 value in decimal.
func sprint256(tv *TypedValue) string {
	z := tv.GetBits256()
	return dec256(&z, tv.T.Kind() == Int256Kind)
}
