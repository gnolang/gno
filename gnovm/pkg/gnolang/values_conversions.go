package gnolang

import (
	"fmt"
	"math"
	"math/big"
	"unicode/utf8"

	"github.com/cockroachdb/apd/v3"
	"github.com/gnolang/gno/gnovm/pkg/gnolang/internal/softfloat"
)

// t cannot be nil or untyped or DataByteType.
// the conversion is forced and overflow/underflow is ignored.
// TODO: return error, and let caller also print the file and line.
func ConvertTo(alloc *Allocator, store Store, tv *TypedValue, t Type, isConst bool) {
	if debug {
		if t == nil {
			panic("ConvertTo() requires non-nil type")
		}
		if isUntyped(t) {
			panic("cannot convert to untyped type")
		}
		if isDataByte(t) {
			panic("cannot convert to databyte type")
		}
		if isDataByte(tv.T) {
			panic("should not happen")
		}
	}
	// special case for interface target
	if t.Kind() == InterfaceKind {
		return
	}
	// special case for undefined/nil source
	if tv.IsUndefined() {
		switch t.Kind() {
		case BoolKind, StringKind, IntKind, Int8Kind, Int16Kind, Int32Kind, Int64Kind, UintKind, Uint8Kind, Uint16Kind, Uint32Kind, Uint64Kind, Float32Kind, Float64Kind, BigintKind, BigdecKind:
			panic(fmt.Sprintf("cannot convert %v to %v", tv, t))
		}
		tv.T = t
		return
	}
	// general case
	tvk := tv.T.Kind()
	k := t.Kind()
	if tvk == k {
		tv.T = t // simple conversion.
		return
	}

	validate := func(from Kind, to Kind, cmp func() bool) {
		if isConst {
			msg := fmt.Sprintf("cannot convert constant of type %s to %s", from, to)
			if cmp != nil && cmp() {
				return
			}
			panic(msg)
		}
	}

	switch tvk {
	case IntKind:
		switch k {
		case IntKind:
			x := tv.GetInt()
			tv.T = t
			tv.SetInt(x)
		case Int8Kind:
			validate(IntKind, Int8Kind, func() bool { return tv.GetInt() >= math.MinInt8 && tv.GetInt() <= math.MaxInt8 })

			x := int8(tv.GetInt())
			tv.T = t
			tv.SetInt8(x)
		case Int16Kind:
			validate(IntKind, Int16Kind, func() bool { return tv.GetInt() >= math.MinInt16 && tv.GetInt() <= math.MaxInt16 })

			x := int16(tv.GetInt())
			tv.T = t
			tv.SetInt16(x)
		case Int32Kind:
			validate(IntKind, Int32Kind, func() bool { return tv.GetInt() >= math.MinInt32 && tv.GetInt() <= math.MaxInt32 })

			x := int32(tv.GetInt())
			tv.T = t
			tv.SetInt32(x)
		case Int64Kind:
			x := tv.GetInt()
			tv.T = t
			tv.SetInt64(x)
		case UintKind:
			validate(IntKind, UintKind, func() bool { return tv.GetInt() >= 0 })

			x := uint64(tv.GetInt())
			tv.T = t
			tv.SetUint(x)
		case Uint8Kind:
			validate(IntKind, Uint8Kind, func() bool { return tv.GetInt() >= 0 && tv.GetInt() <= math.MaxUint8 })

			x := uint8(tv.GetInt())
			tv.T = t
			tv.SetUint8(x)
		case Uint16Kind:
			validate(IntKind, Uint16Kind, func() bool { return tv.GetInt() >= 0 && tv.GetInt() <= math.MaxUint16 })

			x := uint16(tv.GetInt())
			tv.T = t
			tv.SetUint16(x)
		case Uint32Kind:
			validate(IntKind, Uint32Kind, func() bool { return tv.GetInt() >= 0 && uint64(tv.GetInt()) <= math.MaxUint32 })

			x := uint32(tv.GetInt())
			tv.T = t
			tv.SetUint32(x)
		case Uint64Kind:
			validate(IntKind, Uint64Kind, func() bool { return tv.GetInt() >= 0 })

			x := uint64(tv.GetInt())
			tv.T = t
			tv.SetUint64(x)
		case Float32Kind:
			x := softfloat.Fintto32(tv.GetInt())
			tv.T = t
			tv.SetFloat32(x)
		case Float64Kind:
			x := softfloat.Fintto64(tv.GetInt())
			tv.T = t
			tv.SetFloat64(x)
		case StringKind:
			validate(IntKind, StringKind, nil)
			tv.V = alloc.NewString(string(rune(tv.GetInt())))
			tv.T = t
			tv.ClearNum()
		default:
			panic(fmt.Sprintf(
				"cannot convert %s to %s",
				tvk.String(), k.String()))
		}
	case Int8Kind:
		switch k {
		case IntKind:
			x := int64(tv.GetInt8())
			tv.T = t
			tv.SetInt(x)
		case Int8Kind:
			x := tv.GetInt8()
			tv.T = t
			tv.SetInt8(x)
		case Int16Kind:
			x := int16(tv.GetInt8())
			tv.T = t
			tv.SetInt16(x)
		case Int32Kind:
			x := int32(tv.GetInt8())
			tv.T = t
			tv.SetInt32(x)
		case Int64Kind:
			x := int64(tv.GetInt8())
			tv.T = t
			tv.SetInt64(x)
		case UintKind:
			validate(Int8Kind, UintKind, func() bool { return tv.GetInt8() >= 0 })

			x := uint64(tv.GetInt8())
			tv.T = t
			tv.SetUint(x)
		case Uint8Kind:
			validate(Int8Kind, Uint8Kind, func() bool { return tv.GetInt8() >= 0 })

			x := uint8(tv.GetInt8())
			tv.T = t
			tv.SetUint8(x)
		case Uint16Kind:
			validate(Int8Kind, Uint16Kind, func() bool { return tv.GetInt8() >= 0 })

			x := uint16(tv.GetInt8())
			tv.T = t
			tv.SetUint16(x)
		case Uint32Kind:
			validate(Int8Kind, Uint32Kind, func() bool { return tv.GetInt8() >= 0 })

			x := uint32(tv.GetInt8())
			tv.T = t
			tv.SetUint32(x)
		case Uint64Kind:
			validate(Int8Kind, Uint64Kind, func() bool { return tv.GetInt8() >= 0 })

			x := uint64(tv.GetInt8())
			tv.T = t
			tv.SetUint64(x)
		case Float32Kind:
			x := softfloat.Fint32to32(int32(tv.GetInt8()))
			tv.T = t
			tv.SetFloat32(x)
		case Float64Kind:
			x := softfloat.Fint32to64(int32(tv.GetInt8()))
			tv.T = t
			tv.SetFloat64(x)
		case StringKind:
			tv.V = alloc.NewString(string(rune(tv.GetInt8())))
			tv.T = t
			tv.ClearNum()
		default:
			panic(fmt.Sprintf(
				"cannot convert %s to %s",
				tvk.String(), k.String()))
		}
	case Int16Kind:
		switch k {
		case IntKind:
			x := int64(tv.GetInt16())
			tv.T = t
			tv.SetInt(x)
		case Int8Kind:
			validate(Int16Kind, Int8Kind, func() bool { return tv.GetInt16() >= math.MinInt8 && tv.GetInt16() <= math.MaxInt8 })

			x := int8(tv.GetInt16())
			tv.T = t
			tv.SetInt8(x)
		case Int16Kind:
			x := tv.GetInt16()
			tv.T = t
			tv.SetInt16(x)
		case Int32Kind:
			x := int32(tv.GetInt16())
			tv.T = t
			tv.SetInt32(x)
		case Int64Kind:
			x := int64(tv.GetInt16())
			tv.T = t
			tv.SetInt64(x)
		case UintKind:
			validate(Int16Kind, UintKind, func() bool { return tv.GetInt16() >= 0 })

			x := uint64(tv.GetInt16())
			tv.T = t
			tv.SetUint(x)
		case Uint8Kind:
			validate(Int16Kind, Uint8Kind, func() bool { return tv.GetInt16() >= 0 && tv.GetInt16() <= math.MaxUint8 })

			x := uint8(tv.GetInt16())
			tv.T = t
			tv.SetUint8(x)
		case Uint16Kind:
			validate(Int16Kind, Uint16Kind, func() bool { return tv.GetInt16() >= 0 })

			x := uint16(tv.GetInt16())
			tv.T = t
			tv.SetUint16(x)
		case Uint32Kind:
			validate(Int16Kind, Uint32Kind, func() bool { return tv.GetInt16() >= 0 })

			x := uint32(tv.GetInt16())
			tv.T = t
			tv.SetUint32(x)
		case Uint64Kind:
			validate(Int16Kind, Uint64Kind, func() bool { return tv.GetInt16() >= 0 })

			x := uint64(tv.GetInt16())
			tv.T = t
			tv.SetUint64(x)
		case Float32Kind:
			x := softfloat.Fint32to32(int32(tv.GetInt16()))
			tv.T = t
			tv.SetFloat32(x)
		case Float64Kind:
			x := softfloat.Fint32to64(int32(tv.GetInt16()))
			tv.T = t
			tv.SetFloat64(x)
		case StringKind:
			validate(Int16Kind, StringKind, nil)

			tv.V = alloc.NewString(string(rune(tv.GetInt16())))
			tv.T = t
			tv.ClearNum()
		default:
			panic(fmt.Sprintf(
				"cannot convert %s to %s",
				tvk.String(), k.String()))
		}
	case Int32Kind:
		switch k {
		case IntKind:
			x := int64(tv.GetInt32())
			tv.T = t
			tv.SetInt(x)
		case Int8Kind:
			validate(Int32Kind, Int8Kind, func() bool { return tv.GetInt32() >= math.MinInt8 && tv.GetInt32() <= math.MaxInt8 })

			x := int8(tv.GetInt32())
			tv.T = t
			tv.SetInt8(x)
		case Int16Kind:
			validate(Int32Kind, Int16Kind, func() bool { return tv.GetInt32() >= math.MinInt16 && tv.GetInt32() <= math.MaxInt16 })

			x := int16(tv.GetInt32())
			tv.T = t
			tv.SetInt16(x)
		case Int32Kind:
			x := tv.GetInt32()
			tv.T = t
			tv.SetInt32(x)
		case Int64Kind:
			x := int64(tv.GetInt32())
			tv.T = t
			tv.SetInt64(x)
		case UintKind:
			validate(Int32Kind, UintKind, func() bool { return tv.GetInt32() >= 0 })

			x := uint64(tv.GetInt32())
			tv.T = t
			tv.SetUint(x)
		case Uint8Kind:
			validate(Int32Kind, Uint8Kind, func() bool { return tv.GetInt32() >= 0 && tv.GetInt32() <= math.MaxUint8 })

			x := uint8(tv.GetInt32())
			tv.T = t
			tv.SetUint8(x)
		case Uint16Kind:
			validate(Int32Kind, Uint16Kind, func() bool { return tv.GetInt32() >= 0 && tv.GetInt32() <= math.MaxUint16 })

			x := uint16(tv.GetInt32())
			tv.T = t
			tv.SetUint16(x)
		case Uint32Kind:
			validate(Int32Kind, Uint32Kind, func() bool { return tv.GetInt32() >= 0 })

			x := uint32(tv.GetInt32())
			tv.T = t
			tv.SetUint32(x)
		case Uint64Kind:
			validate(Int32Kind, Uint64Kind, func() bool { return tv.GetInt32() >= 0 })

			x := uint64(tv.GetInt32())
			tv.T = t
			tv.SetUint64(x)
		case Float32Kind:
			x := softfloat.Fint32to32(tv.GetInt32())
			tv.T = t
			tv.SetFloat32(x)
		case Float64Kind:
			x := softfloat.Fint32to64(tv.GetInt32())
			tv.T = t
			tv.SetFloat64(x)
		case StringKind:
			validate(Int32Kind, StringKind, nil)

			tv.V = alloc.NewString(string(tv.GetInt32()))
			tv.T = t
			tv.ClearNum()
		default:
			panic(fmt.Sprintf(
				"cannot convert %s to %s",
				tvk.String(), k.String()))
		}
	case Int64Kind:
		switch k {
		case IntKind:
			x := tv.GetInt64()
			tv.T = t
			tv.SetInt(x)
		case Int8Kind:
			validate(Int64Kind, Int8Kind, func() bool { return tv.GetInt64() >= math.MinInt8 && tv.GetInt64() <= math.MaxInt8 })

			x := int8(tv.GetInt64())
			tv.T = t
			tv.SetInt8(x)
		case Int16Kind:
			validate(Int64Kind, Int16Kind, func() bool { return tv.GetInt64() >= math.MinInt16 && tv.GetInt64() <= math.MaxInt16 })

			x := int16(tv.GetInt64())
			tv.T = t
			tv.SetInt16(x)
		case Int32Kind:
			validate(Int64Kind, Int32Kind, func() bool { return tv.GetInt64() >= math.MinInt32 && tv.GetInt64() <= math.MaxInt32 })

			x := int32(tv.GetInt64())
			tv.T = t
			tv.SetInt32(x)
		case Int64Kind:
			x := tv.GetInt64()
			tv.T = t
			tv.SetInt64(x)
		case UintKind:
			validate(Int64Kind, UintKind, func() bool { return tv.GetInt64() >= 0 && uint(tv.GetInt64()) <= math.MaxUint })

			x := uint64(tv.GetInt64())
			tv.T = t
			tv.SetUint(x)
		case Uint8Kind:
			validate(Int64Kind, Uint8Kind, func() bool { return tv.GetInt64() >= 0 && tv.GetInt64() <= math.MaxUint8 })

			x := uint8(tv.GetInt64())
			tv.T = t
			tv.SetUint8(x)
		case Uint16Kind:
			validate(Int64Kind, Uint16Kind, func() bool { return tv.GetInt64() >= 0 && tv.GetInt64() <= math.MaxUint16 })

			x := uint16(tv.GetInt64())
			tv.T = t
			tv.SetUint16(x)
		case Uint32Kind:
			validate(Int64Kind, Uint32Kind, func() bool { return tv.GetInt64() >= 0 && tv.GetInt64() <= math.MaxUint32 })

			x := uint32(tv.GetInt64())
			tv.T = t
			tv.SetUint32(x)
		case Uint64Kind:
			validate(Int64Kind, Uint64Kind, func() bool { return tv.GetInt64() >= 0 })

			x := uint64(tv.GetInt64())
			tv.T = t
			tv.SetUint64(x)
		case Float32Kind:
			x := softfloat.Fint64to32(tv.GetInt64())
			tv.T = t
			tv.SetFloat32(x)
		case Float64Kind:
			x := softfloat.Fint64to64(tv.GetInt64())
			tv.T = t
			tv.SetFloat64(x)
		case StringKind:
			validate(Int64Kind, Uint64Kind, nil)

			tv.V = alloc.NewString(string(rune(tv.GetInt64())))
			tv.T = t
			tv.ClearNum()
		default:
			panic(fmt.Sprintf(
				"cannot convert %s to %s",
				tvk.String(), k.String()))
		}
	case UintKind:
		switch k {
		case IntKind:
			validate(UintKind, IntKind, func() bool { return tv.GetUint() <= math.MaxInt })

			x := int64(tv.GetUint())
			tv.T = t
			tv.SetInt(x)
		case Int8Kind:
			validate(UintKind, Int8Kind, func() bool { return tv.GetUint() <= math.MaxInt8 })

			x := int8(tv.GetUint())
			tv.T = t
			tv.SetInt8(x)
		case Int16Kind:
			validate(UintKind, Int16Kind, func() bool { return tv.GetUint() <= math.MaxInt16 })

			x := int16(tv.GetUint())
			tv.T = t
			tv.SetInt16(x)
		case Int32Kind:
			validate(UintKind, Int32Kind, func() bool { return tv.GetUint() <= math.MaxInt32 })

			x := int32(tv.GetUint())
			tv.T = t
			tv.SetInt32(x)
		case Int64Kind:
			validate(UintKind, Int64Kind, func() bool { return tv.GetUint() <= math.MaxInt64 })

			x := int64(tv.GetUint())
			tv.T = t
			tv.SetInt64(x)
		case UintKind:
			x := tv.GetUint()
			tv.T = t
			tv.SetUint(x)
		case Uint8Kind:
			validate(UintKind, Uint8Kind, func() bool { return tv.GetUint() <= math.MaxUint8 })

			x := uint8(tv.GetUint())
			tv.T = t
			tv.SetUint8(x)
		case Uint16Kind:
			validate(UintKind, Uint16Kind, func() bool { return tv.GetUint() <= math.MaxUint16 })

			x := uint16(tv.GetUint())
			tv.T = t
			tv.SetUint16(x)
		case Uint32Kind:
			validate(UintKind, Uint32Kind, func() bool { return tv.GetUint() <= math.MaxUint32 })

			x := uint32(tv.GetUint())
			tv.T = t
			tv.SetUint32(x)
		case Uint64Kind:
			x := tv.GetUint()
			tv.T = t
			tv.SetUint64(x)
		case Float32Kind:
			x := softfloat.Fuint64to32(tv.GetUint())
			tv.T = t
			tv.SetFloat32(x)
		case Float64Kind:
			x := softfloat.Fuint64to64(tv.GetUint())
			tv.T = t
			tv.SetFloat64(x)
		case StringKind:
			validate(UintKind, StringKind, nil)

			tv.V = alloc.NewString(string(rune(tv.GetUint())))
			tv.T = t
			tv.ClearNum()
		default:
			panic(fmt.Sprintf(
				"cannot convert %s to %s",
				tvk.String(), k.String()))
		}
	case Uint8Kind:
		switch k {
		case IntKind:
			x := int64(tv.GetUint8())
			tv.T = t
			tv.SetInt(x)
		case Int8Kind:
			validate(Uint8Kind, Int8Kind, func() bool { return tv.GetUint8() <= math.MaxInt8 })

			x := int8(tv.GetUint8())
			tv.T = t
			tv.SetInt8(x)
		case Int16Kind:
			validate(Uint8Kind, Int16Kind, func() bool { return int64(tv.GetUint8()) <= math.MaxInt16 })

			x := int16(tv.GetUint8())
			tv.T = t
			tv.SetInt16(x)
		case Int32Kind:
			validate(Uint8Kind, Int32Kind, func() bool { return int64(tv.GetUint8()) <= math.MaxInt32 })

			x := int32(tv.GetUint8())
			tv.T = t
			tv.SetInt32(x)
		case Int64Kind:
			validate(Uint8Kind, Int64Kind, func() bool { return true })

			x := int64(tv.GetUint8())
			tv.T = t
			tv.SetInt64(x)
		case UintKind:
			x := uint64(tv.GetUint8())
			tv.T = t
			tv.SetUint(x)
		case Uint8Kind:
			x := tv.GetUint8()
			tv.T = t
			tv.SetUint8(x)
		case Uint16Kind:
			x := uint16(tv.GetUint8())
			tv.T = t
			tv.SetUint16(x)
		case Uint32Kind:
			x := uint32(tv.GetUint8())
			tv.T = t
			tv.SetUint32(x)
		case Uint64Kind:
			x := uint64(tv.GetUint8())
			tv.T = t
			tv.SetUint64(x)
		case Float32Kind:
			x := softfloat.Fuint64to32(uint64(tv.GetUint8()))
			tv.T = t
			tv.SetFloat32(x)
		case Float64Kind:
			x := softfloat.Fuint64to64(uint64(tv.GetUint8()))
			tv.T = t
			tv.SetFloat64(x)
		case StringKind:
			validate(Uint8Kind, StringKind, nil)

			tv.V = alloc.NewString(string(rune(tv.GetUint8())))
			tv.T = t
			tv.ClearNum()
		default:
			panic(fmt.Sprintf(
				"cannot convert %s to %s",
				tvk.String(), k.String()))
		}
	case Uint16Kind:
		switch k {
		case IntKind:
			x := int64(tv.GetUint16())
			tv.T = t
			tv.SetInt(x)
		case Int8Kind:
			validate(Uint16Kind, Int8Kind, func() bool { return tv.GetUint16() <= math.MaxInt8 })

			x := int8(tv.GetUint16())
			tv.T = t
			tv.SetInt8(x)
		case Int16Kind:
			validate(Uint16Kind, Int16Kind, func() bool { return tv.GetUint16() <= math.MaxInt16 })

			x := int16(tv.GetUint16())
			tv.T = t
			tv.SetInt16(x)
		case Int32Kind:
			validate(Uint16Kind, Int32Kind, func() bool { return int64(tv.GetUint16()) <= math.MaxInt32 })

			x := int32(tv.GetUint16())
			tv.T = t
			tv.SetInt32(x)
		case Int64Kind:
			validate(Uint16Kind, Int64Kind, func() bool { return true })

			x := int64(tv.GetUint16())
			tv.T = t
			tv.SetInt64(x)
		case UintKind:
			x := uint64(tv.GetUint16())
			tv.T = t
			tv.SetUint(x)
		case Uint8Kind:
			validate(Uint16Kind, Uint8Kind, func() bool { return int64(tv.GetUint16()) <= math.MaxUint8 })

			x := uint8(tv.GetUint16())
			tv.T = t
			tv.SetUint8(x)
		case Uint16Kind:
			x := tv.GetUint16()
			tv.T = t
			tv.SetUint16(x)
		case Uint32Kind:
			x := uint32(tv.GetUint16())
			tv.T = t
			tv.SetUint32(x)
		case Uint64Kind:
			x := uint64(tv.GetUint16())
			tv.T = t
			tv.SetUint64(x)
		case Float32Kind:
			x := softfloat.Fuint64to32(uint64(tv.GetUint16()))
			tv.T = t
			tv.SetFloat32(x)
		case Float64Kind:
			x := softfloat.Fuint64to64(uint64(tv.GetUint16()))
			tv.T = t
			tv.SetFloat64(x)
		case StringKind:
			validate(Uint16Kind, StringKind, nil)

			tv.V = alloc.NewString(string(rune(tv.GetUint16())))
			tv.T = t
			tv.ClearNum()
		default:
			panic(fmt.Sprintf(
				"cannot convert %s to %s",
				tvk.String(), k.String()))
		}
	case Uint32Kind:
		switch k {
		case IntKind:
			validate(Uint32Kind, IntKind, func() bool { return int64(tv.GetUint32()) <= math.MaxInt })

			x := int64(tv.GetUint32())
			tv.T = t
			tv.SetInt(x)
		case Int8Kind:
			validate(Uint32Kind, Int8Kind, func() bool { return int64(tv.GetUint32()) <= math.MaxInt8 })

			x := int8(tv.GetUint32())
			tv.T = t
			tv.SetInt8(x)
		case Int16Kind:
			validate(Uint32Kind, Int16Kind, func() bool { return int64(tv.GetUint32()) <= math.MaxInt16 })

			x := int16(tv.GetUint32())
			tv.T = t
			tv.SetInt16(x)
		case Int32Kind:
			validate(Uint32Kind, Int32Kind, func() bool { return int64(tv.GetUint32()) <= math.MaxInt32 })

			x := int32(tv.GetUint32())
			tv.T = t
			tv.SetInt32(x)
		case Int64Kind:
			x := int64(tv.GetUint32())
			tv.T = t
			tv.SetInt64(x)
		case UintKind:
			x := uint64(tv.GetUint32())
			tv.T = t
			tv.SetUint64(x)
		case Uint8Kind:
			validate(Uint32Kind, Uint8Kind, func() bool { return int(tv.GetUint32()) <= math.MaxUint8 })

			x := uint8(tv.GetUint32())
			tv.T = t
			tv.SetUint8(x)
		case Uint16Kind:
			validate(Uint32Kind, Uint16Kind, func() bool { return int(tv.GetUint32()) <= math.MaxUint16 })

			x := uint16(tv.GetUint32())
			tv.T = t
			tv.SetUint16(x)
		case Uint32Kind:
			x := tv.GetUint32()
			tv.T = t
			tv.SetUint32(x)
		case Uint64Kind:
			x := uint64(tv.GetUint32())
			tv.T = t
			tv.SetUint64(x)
		case Float32Kind:
			x := softfloat.Fuint64to32(uint64(tv.GetUint32()))
			tv.T = t
			tv.SetFloat32(x)
		case Float64Kind:
			x := softfloat.Fuint64to64(uint64(tv.GetUint32()))
			tv.T = t
			tv.SetFloat64(x)
		case StringKind:
			validate(Uint32Kind, StringKind, nil)

			tv.V = alloc.NewString(string(rune(tv.GetUint32())))
			tv.T = t
			tv.ClearNum()
		default:
			panic(fmt.Sprintf(
				"cannot convert %s to %s",
				tvk.String(), k.String()))
		}
	case Uint64Kind:
		switch k {
		case IntKind:
			validate(Uint64Kind, IntKind, func() bool { return int64(tv.GetUint64()) <= math.MaxInt })

			x := int64(tv.GetUint64())
			tv.T = t
			tv.SetInt(x)
		case Int8Kind:
			validate(Uint64Kind, Int8Kind, func() bool { return int64(tv.GetUint64()) <= math.MaxInt8 })

			x := int8(tv.GetUint64())
			tv.T = t
			tv.SetInt8(x)
		case Int16Kind:
			validate(Uint64Kind, Int16Kind, func() bool { return int64(tv.GetUint64()) <= math.MaxInt16 })

			x := int16(tv.GetUint64())
			tv.T = t
			tv.SetInt16(x)
		case Int32Kind:
			validate(Uint64Kind, Int32Kind, func() bool { return int64(tv.GetUint64()) <= math.MaxInt32 })

			x := int32(tv.GetUint64())
			tv.T = t
			tv.SetInt32(x)
		case Int64Kind:
			validate(Uint64Kind, Int64Kind, func() bool { return tv.GetUint64() <= math.MaxInt64 })

			x := int64(tv.GetUint64())
			tv.T = t
			tv.SetInt64(x)
		case UintKind:
			validate(Uint64Kind, UintKind, func() bool { return tv.GetUint64() <= math.MaxUint })

			x := tv.GetUint64()
			tv.T = t
			tv.SetUint64(x)
		case Uint8Kind:
			validate(Uint64Kind, Uint8Kind, func() bool { return int64(tv.GetUint64()) <= math.MaxUint8 })

			x := uint8(tv.GetUint64())
			tv.T = t
			tv.SetUint8(x)
		case Uint16Kind:
			validate(Uint64Kind, Uint16Kind, func() bool { return int64(tv.GetUint64()) <= math.MaxUint16 })

			x := uint16(tv.GetUint64())
			tv.T = t
			tv.SetUint16(x)
		case Uint32Kind:
			validate(Uint64Kind, Uint32Kind, func() bool { return tv.GetUint64() <= math.MaxUint32 })

			x := uint32(tv.GetUint64())
			tv.T = t
			tv.SetUint32(x)
		case Uint64Kind:
			x := tv.GetUint64()
			tv.T = t
			tv.SetUint64(x)
		case Float32Kind:
			x := softfloat.Fuint64to32(tv.GetUint64())
			tv.T = t
			tv.SetFloat32(x)
		case Float64Kind:
			x := softfloat.Fuint64to64(tv.GetUint64())
			tv.T = t
			tv.SetFloat64(x)
		case StringKind:
			validate(Uint64Kind, StringKind, nil)

			tv.V = alloc.NewString(string(rune(tv.GetUint64())))
			tv.T = t
			tv.ClearNum()
		default:
			panic(fmt.Sprintf(
				"cannot convert %s to %s",
				tvk.String(), k.String()))
		}
	case Float32Kind:
		switch k {
		case IntKind:
			validate(Float32Kind, IntKind, func() bool {
				f32 := tv.GetFloat32()
				return softfloat.Fint64to32(softfloat.F32toint64(f32)) == f32
			})

			x := softfloat.F32toint64(tv.GetFloat32())
			tv.T = t
			tv.SetInt(x)
		case Int8Kind:
			validate(Float32Kind, Int8Kind, func() bool {
				f32 := tv.GetFloat32()
				trunc := int8(softfloat.F32toint64(f32))
				return softfloat.Fint64to32(int64(trunc)) == f32
			})

			x := int8(softfloat.F32toint32(tv.GetFloat32()))
			tv.T = t
			tv.SetInt8(x)
		case Int16Kind:
			validate(Float32Kind, Int16Kind, func() bool {
				f32 := tv.GetFloat32()
				trunc := int16(softfloat.F32toint64(f32))
				return softfloat.Fint64to32(int64(trunc)) == f32
			})

			x := int16(softfloat.F32toint32(tv.GetFloat32()))
			tv.T = t
			tv.SetInt16(x)
		case Int32Kind:
			validate(Float32Kind, Int32Kind, func() bool {
				f32 := tv.GetFloat32()
				trunc := int32(softfloat.F32toint64(f32))
				return softfloat.Fint64to32(int64(trunc)) == f32
			})

			x := softfloat.F32toint32(tv.GetFloat32())
			tv.T = t
			tv.SetInt32(x)
		case Int64Kind:
			validate(Float32Kind, Int64Kind, func() bool {
				f32 := tv.GetFloat32()
				return softfloat.Fint64to32(softfloat.F32toint64(f32)) == f32
			})

			x := softfloat.F32toint64(tv.GetFloat32())
			tv.T = t
			tv.SetInt64(x)
		case UintKind:
			validate(Float32Kind, UintKind, func() bool {
				f32 := tv.GetFloat32()
				return softfloat.Fuint64to32(softfloat.F32touint64(f32)) == f32
			})

			x := softfloat.F32touint64(tv.GetFloat32())
			tv.T = t
			tv.SetUint(x)
		case Uint8Kind:
			validate(Float32Kind, Uint8Kind, func() bool {
				f32 := tv.GetFloat32()
				trunc := uint8(softfloat.F32touint64(f32))
				return softfloat.Fuint64to32(uint64(trunc)) == f32
			})

			x := uint8(softfloat.F32touint64(tv.GetFloat32()))
			tv.T = t
			tv.SetUint8(x)
		case Uint16Kind:
			validate(Float32Kind, Uint16Kind, func() bool {
				f32 := tv.GetFloat32()
				trunc := uint16(softfloat.F32touint64(f32))
				return softfloat.Fuint64to32(uint64(trunc)) == f32
			})

			x := uint16(softfloat.F32touint64(tv.GetFloat32()))
			tv.T = t
			tv.SetUint16(x)
		case Uint32Kind:
			validate(Float32Kind, Uint32Kind, func() bool {
				f32 := tv.GetFloat32()
				trunc := uint32(softfloat.F32touint64(f32))
				return softfloat.Fuint64to32(uint64(trunc)) == f32
			})

			x := uint32(softfloat.F32touint64(tv.GetFloat32()))
			tv.T = t
			tv.SetUint32(x)
		case Uint64Kind:
			validate(Float32Kind, Uint64Kind, func() bool {
				f32 := tv.GetFloat32()
				return softfloat.Fuint64to32(softfloat.F32touint64(f32)) == f32
			})

			x := softfloat.F32touint64(tv.GetFloat32())
			tv.T = t
			tv.SetUint64(x)
		case Float32Kind:
			x := tv.GetFloat32() // ???
			tv.T = t
			tv.SetFloat32(x)
		case Float64Kind:
			x := softfloat.F32to64(tv.GetFloat32())
			tv.T = t
			tv.SetFloat64(x)
		default:
			panic(fmt.Sprintf(
				"cannot convert %s to %s",
				tvk.String(), k.String()))
		}
	case Float64Kind:
		switch k {
		case IntKind:
			validate(Float64Kind, IntKind, func() bool {
				f64 := tv.GetFloat64()
				return softfloat.Fint64to64(softfloat.F64toint64(f64)) == f64
			})

			xp, _ := softfloat.F64toint(tv.GetFloat64())
			tv.T = t
			tv.SetInt(xp)
		case Int8Kind:
			validate(Float64Kind, Int8Kind, func() bool {
				f64 := tv.GetFloat64()
				trunc := int8(softfloat.F64toint64(f64))
				return softfloat.Fint64to64(int64(trunc)) == f64
			})

			x := int8(softfloat.F64toint32(tv.GetFloat64()))
			tv.T = t
			tv.SetInt8(x)
		case Int16Kind:
			validate(Float64Kind, Int16Kind, func() bool {
				f64 := tv.GetFloat64()
				trunc := int16(softfloat.F64toint64(f64))
				return softfloat.Fint64to64(int64(trunc)) == f64
			})

			x := int16(softfloat.F64toint32(tv.GetFloat64()))
			tv.T = t
			tv.SetInt16(x)
		case Int32Kind:
			validate(Float64Kind, Int32Kind, func() bool {
				f64 := tv.GetFloat64()
				trunc := int32(softfloat.F64toint64(f64))
				return softfloat.Fint64to64(int64(trunc)) == f64
			})

			x := softfloat.F64toint32(tv.GetFloat64())
			tv.T = t
			tv.SetInt32(x)
		case Int64Kind:
			validate(Float64Kind, Int64Kind, func() bool {
				f64 := tv.GetFloat64()
				return softfloat.Fint64to64(softfloat.F64toint64(f64)) == f64
			})

			x := softfloat.F64toint64(tv.GetFloat64())
			tv.T = t
			tv.SetInt64(x)
		case UintKind:
			validate(Float64Kind, UintKind, func() bool {
				f64 := tv.GetFloat64()
				return softfloat.Fuint64to64(softfloat.F64touint64(f64)) == f64
			})

			x := softfloat.F64touint64(tv.GetFloat64())
			tv.T = t
			tv.SetUint(x)
		case Uint8Kind:
			validate(Float64Kind, Uint8Kind, func() bool {
				f64 := tv.GetFloat64()
				return softfloat.Fuint64to64(softfloat.F64touint64(f64)&(1<<8-1)) == f64
			})

			x := uint8(softfloat.F64touint64(tv.GetFloat64()))
			tv.T = t
			tv.SetUint8(x)
		case Uint16Kind:
			validate(Float64Kind, Uint16Kind, func() bool {
				f64 := tv.GetFloat64()
				return softfloat.Fuint64to64(softfloat.F64touint64(f64)&(1<<16-1)) == f64
			})

			x := uint16(softfloat.F64touint64(tv.GetFloat64()))
			tv.T = t
			tv.SetUint16(x)
		case Uint32Kind:
			validate(Float64Kind, Uint32Kind, func() bool {
				f64 := tv.GetFloat64()
				return softfloat.Fuint64to64(softfloat.F64touint64(f64)&(1<<32-1)) == f64
			})

			x := uint32(softfloat.F64touint64(tv.GetFloat64()))
			tv.T = t
			tv.SetUint32(x)
		case Uint64Kind:
			validate(Float64Kind, Uint64Kind, func() bool {
				f64 := tv.GetFloat64()
				return softfloat.Fuint64to64(softfloat.F64touint64(f64)) == f64
			})

			x := softfloat.F64touint64(tv.GetFloat64())
			tv.T = t
			tv.SetUint64(x)
		case Float32Kind:
			validate(Float64Kind, Float32Kind, func() bool {
				// TODO(morgan): Doesn't account for loss of precision in
				// subnromal value, doesn't account for negative values.
				return softfloat.Fle64(tv.GetFloat64(), math.Float64bits(float64(math.MaxFloat32)))
			})

			x := softfloat.F64to32(tv.GetFloat64())
			tv.T = t
			tv.SetFloat32(x)
		case Float64Kind:
			x := tv.GetFloat64() // ???
			tv.T = t
			tv.SetFloat64(x)
		default:
			panic(fmt.Sprintf(
				"cannot convert %s to %s",
				tvk.String(), k.String()))
		}
	case StringKind:
		bt := baseOf(t)
		switch cbt := bt.(type) {
		case *SliceType:
			switch cbt.Elt.Kind() {
			case Uint8Kind:
				data := []byte(tv.GetString())
				data = data[:len(data):len(data)] // defensive: force cap = len (not guaranteed by Go spec)
				tv.V = alloc.NewSliceFromData(data)
				tv.T = t // after tv.GetString()
			case Int32Kind:
				str := tv.GetString()
				runes := make([]TypedValue, 0, utf8.RuneCountInString(str))
				for _, r := range str {
					runes = append(runes, typedRune(r))
				}
				runes = runes[:len(runes):len(runes)] // force cap = len
				tv.V = alloc.NewSliceFromList(runes)
				tv.T = t // after tv.GetString()
			default:
				panic(fmt.Sprintf(
					"cannot convert %s to %s",
					tvk.String(), t.String()))
			}
		default:
			panic(fmt.Sprintf(
				"cannot convert %s to %s",
				tvk.String(), k.String()))
		}
	case SliceKind:
		if t.Kind() == StringKind {
			tk := tv.T.Elem().Kind()
			if tk != Uint8Kind && tk != Int32Kind {
				panic(fmt.Sprintf(
					"cannot convert %s to %s",
					tv.T.String(), t.String()))
			}
			switch sv := tv.V.(type) {
			case nil:
				tv.T = t
				tv.V = alloc.NewString(string(""))
			case *SliceValue:
				svo := sv.Offset
				svl := sv.Length
				svb := sv.GetBase(store)
				if svb.Data == nil {
					switch tk {
					case Uint8Kind:
						data := make([]byte, svl)
						copyListToData(
							data[:svl],
							svb.List[svo:svo+svl])
						strv := alloc.NewString(string(data))
						tv.T = t
						tv.V = strv
					case Int32Kind:
						runes := make([]rune, svl)
						copyListToRunes(
							runes[:svl],
							svb.List[svo:svo+svl])
						strv := alloc.NewString(string(runes))
						tv.T = t
						tv.V = strv
					default:
						panic("should not happen")
					}
				} else {
					data := svb.Data[svo : svo+svl]
					strv := alloc.NewString(string(data))
					tv.T = t
					tv.V = strv
				}
			default:
				panic("should not happen")
			}
		} else {
			panic(fmt.Sprintf(
				"cannot convert %s to %s",
				tv.T.String(), k.String()))
		}
	default:
		panic(fmt.Sprintf(
			"cannot convert %s to %s",
			tvk.String(), k.String()))
	}
}

// TODO: move untyped const stuff to preprocess.go
// Convert the untyped const tv to the t type.
// If t is nil, the default type is used.
// Panics if conversion is illegal.
// TODO: method on TypedValue?
func ConvertUntypedTo(tv *TypedValue, t Type) {
	if debug {
		defer func() {
			debug.Printf("ConvertUntypedTo done, tv: %v \n", tv)
		}()
	}
	if debug {
		if !isUntyped(tv.T) {
			panic(fmt.Sprintf(
				"ConvertUntypedTo expects untyped const source but got %s",
				tv.T.String()))
		}
		if isUntyped(t) {
			tvpt, ok1 := baseOf(tv.T).(PrimitiveType)
			pt, ok2 := baseOf(t).(PrimitiveType)
			if ok1 && ok2 {
				if tvpt == pt {
					// do nothing
					return
				}
			} else {
				panic(fmt.Sprintf(
					"ConvertUntypedTo expects more specific target for %v but got %s",
					tv.String(),
					t.String()))
			}
		}
	}
	// special case: simple conversion
	if t != nil && tv.T.Kind() == t.Kind() {
		tv.T = t
		return
	}
	// general case
	if t == nil {
		t = defaultTypeOf(tv.T)
	}
	switch tv.T {
	case UntypedBoolType:
		tv.T = t
	case UntypedRuneType:
		ConvertUntypedRuneTo(tv, t)
	case UntypedBigintType:
		ConvertUntypedBigintTo(tv, tv.V.(BigintValue), t)
	case UntypedBigdecType:
		ConvertUntypedBigdecTo(tv, tv.V.(BigdecValue), t)
	case UntypedStringType:
		if t.Kind() == StringKind {
			tv.T = t
			return
		} else {
			ConvertTo(nilAllocator, nil, tv, t, false)
		}
	default:
		panic(fmt.Sprintf(
			"unexpected untyped const type %s",
			tv.T.String()))
	}
}

// All fields may be modified to complete the conversion.
func ConvertUntypedRuneTo(dst *TypedValue, t Type) {
	sv := dst.GetInt32()
	k := t.Kind()
	switch k {
	case IntKind, Int8Kind, Int16Kind, Int32Kind, Int64Kind:
	case UintKind, Uint8Kind, Uint16Kind, Uint32Kind, Uint64Kind:
	case StringKind, BigintKind, BigdecKind:
	default:
		panic(fmt.Sprintf(
			"cannot convert untyped rune type to %s",
			k.String()))
	}
	// Set .T exactly to t, and set .V to nil
	// since the result is only primitive (unless string)
	dst.T = t
	if debug {
		if dst.V != nil {
			panic("should not happen")
		}
	}
	// Check the bounds of sv or uv, given that
	// the value is within int64 or uint64.
	switch k {
	case IntKind:
		dst.SetInt(int64(sv))
	case Int8Kind:
		if math.MaxInt8 < sv {
			panic("rune overflows target kind")
		}
		if sv < math.MinInt8 {
			panic("rune underflows target kind")
		}
		dst.SetInt8(int8(sv))
	case Int16Kind:
		if math.MaxInt16 < sv {
			panic("rune overflows target kind")
		}
		if sv < math.MinInt16 {
			panic("rune underflows target kind")
		}
		dst.SetInt16(int16(sv))
	case Int32Kind:
		dst.SetInt32(sv)
	case Int64Kind:
		dst.SetInt64(int64(sv))
	case UintKind:
		if sv < 0 {
			panic("rune underflows target kind")
		}
		dst.SetUint(uint64(sv))
	case Uint8Kind:
		if sv < 0 {
			panic("rune underflows target kind")
		}
		if math.MaxUint8 < sv {
			panic("rune overflows target kind")
		}
		dst.SetUint8(uint8(sv))
	case Uint16Kind:
		if sv < 0 {
			panic("rune underflows target kind")
		}
		if math.MaxUint16 < sv {
			panic("rune overflows target kind")
		}
		dst.SetUint16(uint16(sv))
	case Uint32Kind:
		if sv < 0 {
			panic("rune underflows target kind")
		}
		dst.SetUint32(uint32(sv))
	case Uint64Kind:
		dst.SetUint64(uint64(sv))
	case StringKind:
		panic("not yet implemented")
	case BigintKind:
		dst.ClearNum()
		dst.V = BigintValue{V: big.NewInt(int64(sv))}
	case BigdecKind:
		dst.ClearNum()
		dst.V = BigdecValue{V: apd.New(int64(sv), 0)}
	default:
		panic(fmt.Sprintf("unexpected target %v", k))
	}
}

func ConvertUntypedBigintTo(dst *TypedValue, biv BigintValue, t Type) {
	k := t.Kind()
	bi := biv.V
	var sv int64 = 0  // if signed.
	var uv uint64 = 0 // if unsigned.
	switch k {
	case BigintKind:
		dst.T = t
		dst.V = biv
		return // done
	case BoolKind:
		panic("not yet implemented")
	case InterfaceKind:
		t = IntType // default
		fallthrough
	case IntKind, Int8Kind, Int16Kind, Int32Kind, Int64Kind:
		// preliminary bounds check... more comes later.
		if !bi.IsInt64() {
			if bi.Sign() == 1 {
				panic("bigint overflows target kind")
			} else {
				panic("bigint underflows target kind")
			}
		}
		sv = bi.Int64()
	case UintKind, Uint8Kind, Uint16Kind, Uint32Kind, Uint64Kind:
		// preliminary bounds check... more comes later.
		if !bi.IsUint64() {
			if bi.Sign() == 1 {
				panic("bigint overflows target kind")
			} else {
				panic("bigint underflows target kind")
			}
		}
		uv = bi.Uint64()
	case Float32Kind:
		dst.T = t
		dst.V = nil
		// 24 for float32
		bf := big.NewFloat(0.0).SetInt(bi).SetPrec(24)
		if bf.IsInf() {
			panic("bigint overflows float32")
		}
		f32, acc := bf.Float32()
		if f32 == 0 && (acc == big.Below || acc == big.Above) {
			panic("bigint underflows float32 (too close to zero)")
		}
		dst.SetFloat32(math.Float32bits(f32))
		return // done
	case Float64Kind:
		dst.T = t
		dst.V = nil
		// 53 for float64
		bf := big.NewFloat(0.0).SetInt(bi).SetPrec(53)
		if bf.IsInf() {
			panic("bigint overflows float64")
		}
		f64, acc := bf.Float64()
		if f64 == 0 && (acc == big.Below || acc == big.Above) {
			panic("bigint underflows float64 (too close to zero)")
		}
		dst.SetFloat64(math.Float64bits(f64))
		return // done
	case BigdecKind:
		dst.T = t
		dst.V = BigdecValue{V: apd.NewWithBigInt(new(apd.BigInt).SetMathBigInt(bi), 0)}
		return // done
	default:
		panic(fmt.Sprintf(
			"cannot convert untyped bigint type to %s",
			k.String()))
	}
	// Set .T exactly to t, and set .V to nil
	// since the result is only primitive (unless string)
	dst.T = t
	dst.V = nil
	// Check the bounds of sv or uv, given that
	// the value is within int64 or uint64.
	switch k {
	case BoolKind:
		panic("not yet implemented")
	case IntKind, InterfaceKind:
		dst.SetInt(sv)
	case Int8Kind:
		if math.MaxInt8 < sv {
			panic("bigint overflows target kind")
		}
		if sv < math.MinInt8 {
			panic("bigint underflows target kind")
		}
		dst.SetInt8(int8(sv))
	case Int16Kind:
		if math.MaxInt16 < sv {
			panic("bigint overflows target kind")
		}
		if sv < math.MinInt16 {
			panic("bigint underflows target kind")
		}
		dst.SetInt16(int16(sv))
	case Int32Kind:
		if math.MaxInt32 < sv {
			panic("bigint overflows target kind")
		}
		if sv < math.MinInt32 {
			panic("bigint underflows target kind")
		}
		dst.SetInt32(int32(sv))
	case Int64Kind:
		dst.SetInt64(sv)
	case UintKind:
		dst.SetUint(uv)
	case Uint8Kind:
		if math.MaxUint8 < uv {
			panic("bigint overflows target kind")
		}
		dst.SetUint8(uint8(uv))
	case Uint16Kind:
		if math.MaxUint16 < uv {
			panic("bigint overflows target kind")
		}
		dst.SetUint16(uint16(uv))
	case Uint32Kind:
		if math.MaxUint32 < uv {
			panic("bigint overflows target kind")
		}
		dst.SetUint32(uint32(uv))
	case Uint64Kind:
		dst.SetUint64(uv)
	case StringKind:
		panic("not yet implemented")
	default:
		panic(fmt.Sprintf("cannot convert untyped bigint to %v", k))
	}
}

func ConvertUntypedBigdecTo(dst *TypedValue, bdv BigdecValue, t Type) {
	k := t.Kind()
	bd := bdv.V
	switch k {
	case BigintKind:
		if !isInteger(bd) {
			panic(fmt.Sprintf(
				"cannot convert untyped bigdec to integer -- %s not an exact integer",
				bd.String(),
			))
		}
		dst.T = t
		dst.V = BigintValue{V: toBigInt(bd)}
		return // done
	case BoolKind:
		panic("cannot convert untyped bigdec to bool")
	case InterfaceKind:
		dst.T = Float64Type
		dst.V = nil
		f, _ := bd.Float64()
		dst.SetFloat64(math.Float64bits(f))
		return
	case IntKind, Int8Kind, Int16Kind, Int32Kind, Int64Kind:
		fallthrough
	case UintKind, Uint8Kind, Uint16Kind, Uint32Kind, Uint64Kind:
		if !isInteger(bd) {
			panic(fmt.Sprintf(
				"cannot convert untyped bigdec to integer -- %s not an exact integer",
				bd.String(),
			))
		}
		ConvertUntypedBigintTo(dst, BigintValue{V: toBigInt(bd)}, t)
		return
	case Float32Kind:
		dst.T = t
		dst.V = nil
		f64, err := bd.Float64()
		if err != nil {
			panic(fmt.Errorf("cannot convert untyped bigdec to float64: %w", err))
		}

		bf := big.NewFloat(f64)
		f32, _ := bf.Float32()
		if math.IsInf(float64(f32), 0) {
			panic("cannot convert untyped bigdec to float32 -- too close to +-Inf")
		}
		dst.SetFloat32(math.Float32bits(f32))
		return
	case Float64Kind:
		dst.T = t
		dst.V = nil
		f64, err := bd.Float64()
		if err != nil {
			panic(fmt.Errorf("cannot convert untyped bigdec to float64: %w", err))
		}
		if math.IsInf(f64, 0) {
			panic("cannot convert untyped bigdec to float64 -- too close to +-Inf")
		}
		dst.SetFloat64(math.Float64bits(f64))
		return
	default:
		panic(fmt.Sprintf(
			"cannot convert untyped bigdec type to %s",
			k.String()))
	}
}

// ----------------------------------------
// apd.Decimal utility

func isInteger(d *apd.Decimal) bool {
	d2 := apd.New(0, 0)
	res, err := apd.BaseContext.RoundToIntegralExact(d2, d)
	if err != nil {
		panic("should not happen")
	}
	integer := !res.Inexact()
	return integer
}

func toBigInt(d *apd.Decimal) *big.Int {
	d2 := apd.New(0, 0)
	_, err := apd.BaseContext.RoundToIntegralExact(d2, d)
	if err != nil {
		panic("should not happen")
	}
	d2s := d2.String()
	bi := big.NewInt(0)
	_, ok := bi.SetString(d2s, 10)
	if !ok {
		panic(fmt.Sprintf(
			"invalid integer constant: %s",
			d2s))
	}
	return bi
}

func IsExactBigDec(v Value) bool {
	if bd, ok := v.(BigdecValue); ok {
		return isInteger(bd.V)
	}
	return false
}
