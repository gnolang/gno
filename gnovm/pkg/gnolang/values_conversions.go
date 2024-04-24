package gnolang

import (
	"fmt"
	"math"
	"math/big"
	"reflect"
	"strconv"

	"github.com/cockroachdb/apd/v3"
)

// t cannot be nil or untyped or DataByteType.
// the conversion is forced and overflow/underflow is ignored.
// TODO: return error, and let caller also print the file and line.
func ConvertTo(alloc *Allocator, store Store, tv *TypedValue, t Type) {
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
	// special case for go-native conversions
	ntv, tvIsNat := tv.T.(*NativeType)
	nt, tIsNat := t.(*NativeType)
	if tvIsNat {
		if tIsNat {
			// both NativeType, use reflect to assert.
			if debug {
				if !ntv.Type.ConvertibleTo(nt.Type) {
					panic(fmt.Sprintf(
						"cannot convert %s to %s",
						ntv.String(), nt.String()))
				}
			}
			tv.T = t
			return
		} else {
			// convert go-native to gno type (shallow).
			*tv = go2GnoValue2(alloc, store, tv.V.(*NativeValue).Value, false)
			ConvertTo(alloc, store, tv, t)
			return
		}
	} else {
		if tIsNat {
			// convert gno to go-native type.
			rv := reflect.New(nt.Type).Elem()
			rv = gno2GoValue(tv, rv)
			if debug {
				if !rv.Type().ConvertibleTo(nt.Type) {
					panic(fmt.Sprintf(
						"cannot convert %s to %s",
						tv.String(), nt.String()))
				}
			}
			*tv = TypedValue{
				T: t,
				V: alloc.NewNative(rv),
			}
			return
		} else {
			goto GNO_CASE
		}
	}
GNO_CASE:
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
	switch tvk {
	case IntKind:
		switch k {
		case IntKind:
			x := tv.GetInt()
			tv.T = t
			tv.SetInt(x)
		case Int8Kind:
			x := int8(tv.GetInt())
			tv.T = t
			tv.SetInt8(x)
		case Int16Kind:
			x := int16(tv.GetInt())
			tv.T = t
			tv.SetInt16(x)
		case Int32Kind:
			x := int32(tv.GetInt())
			tv.T = t
			tv.SetInt32(x)
		case Int64Kind:
			x := int64(tv.GetInt())
			tv.T = t
			tv.SetInt64(x)
		case UintKind:
			x := uint(tv.GetInt())
			tv.T = t
			tv.SetUint(x)
		case Uint8Kind:
			x := uint8(tv.GetInt())
			tv.T = t
			tv.SetUint8(x)
		case Uint16Kind:
			x := uint16(tv.GetInt())
			tv.T = t
			tv.SetUint16(x)
		case Uint32Kind:
			x := uint32(tv.GetInt())
			tv.T = t
			tv.SetUint32(x)
		case Uint64Kind:
			x := uint64(tv.GetInt())
			tv.T = t
			tv.SetUint64(x)
		case Float32Kind:
			x := float32(tv.GetInt()) // XXX determinism?
			tv.T = t
			tv.SetFloat32(x)
		case Float64Kind:
			x := float64(tv.GetInt()) // XXX determinism?
			tv.T = t
			tv.SetFloat64(x)
		case StringKind:
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
			x := int(tv.GetInt8())
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
			x := uint(tv.GetInt8())
			tv.T = t
			tv.SetUint(x)
		case Uint8Kind:
			x := uint8(tv.GetInt8())
			tv.T = t
			tv.SetUint8(x)
		case Uint16Kind:
			x := uint16(tv.GetInt8())
			tv.T = t
			tv.SetUint16(x)
		case Uint32Kind:
			x := uint32(tv.GetInt8())
			tv.T = t
			tv.SetUint32(x)
		case Uint64Kind:
			x := uint64(tv.GetInt8())
			tv.T = t
			tv.SetUint64(x)
		case Float32Kind:
			x := float32(tv.GetInt8()) // XXX determinism?
			tv.T = t
			tv.SetFloat32(x)
		case Float64Kind:
			x := float64(tv.GetInt8()) // XXX determinism?
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
			x := int(tv.GetInt16())
			tv.T = t
			tv.SetInt(x)
		case Int8Kind:
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
			x := uint(tv.GetInt16())
			tv.T = t
			tv.SetUint(x)
		case Uint8Kind:
			x := uint8(tv.GetInt16())
			tv.T = t
			tv.SetUint8(x)
		case Uint16Kind:
			x := uint16(tv.GetInt16())
			tv.T = t
			tv.SetUint16(x)
		case Uint32Kind:
			x := uint32(tv.GetInt16())
			tv.T = t
			tv.SetUint32(x)
		case Uint64Kind:
			x := uint64(tv.GetInt16())
			tv.T = t
			tv.SetUint64(x)
		case Float32Kind:
			x := float32(tv.GetInt16()) // XXX determinism?
			tv.T = t
			tv.SetFloat32(x)
		case Float64Kind:
			x := float64(tv.GetInt16()) // XXX determinism?
			tv.T = t
			tv.SetFloat64(x)
		case StringKind:
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
			x := int(tv.GetInt32())
			tv.T = t
			tv.SetInt(x)
		case Int8Kind:
			x := int8(tv.GetInt32())
			tv.T = t
			tv.SetInt8(x)
		case Int16Kind:
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
			x := uint(tv.GetInt32())
			tv.T = t
			tv.SetUint(x)
		case Uint8Kind:
			x := uint8(tv.GetInt32())
			tv.T = t
			tv.SetUint8(x)
		case Uint16Kind:
			x := uint16(tv.GetInt32())
			tv.T = t
			tv.SetUint16(x)
		case Uint32Kind:
			x := uint32(tv.GetInt32())
			tv.T = t
			tv.SetUint32(x)
		case Uint64Kind:
			x := uint64(tv.GetInt32())
			tv.T = t
			tv.SetUint64(x)
		case Float32Kind:
			x := float32(tv.GetInt32()) // XXX determinism?
			tv.T = t
			tv.SetFloat32(x)
		case Float64Kind:
			x := float64(tv.GetInt32()) // XXX determinism?
			tv.T = t
			tv.SetFloat64(x)
		case StringKind:
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
			x := int(tv.GetInt64())
			tv.T = t
			tv.SetInt(x)
		case Int8Kind:
			x := int8(tv.GetInt64())
			tv.T = t
			tv.SetInt8(x)
		case Int16Kind:
			x := int16(tv.GetInt64())
			tv.T = t
			tv.SetInt16(x)
		case Int32Kind:
			x := int32(tv.GetInt64())
			tv.T = t
			tv.SetInt32(x)
		case Int64Kind:
			x := tv.GetInt64()
			tv.T = t
			tv.SetInt64(x)
		case UintKind:
			x := uint(tv.GetInt64())
			tv.T = t
			tv.SetUint(x)
		case Uint8Kind:
			x := uint8(tv.GetInt64())
			tv.T = t
			tv.SetUint8(x)
		case Uint16Kind:
			x := uint16(tv.GetInt64())
			tv.T = t
			tv.SetUint16(x)
		case Uint32Kind:
			x := uint32(tv.GetInt64())
			tv.T = t
			tv.SetUint32(x)
		case Uint64Kind:
			x := uint64(tv.GetInt64())
			tv.T = t
			tv.SetUint64(x)
		case Float32Kind:
			x := float32(tv.GetInt64()) // XXX determinism?
			tv.T = t
			tv.SetFloat32(x)
		case Float64Kind:
			x := float64(tv.GetInt64()) // XXX determinism?
			tv.T = t
			tv.SetFloat64(x)
		case StringKind:
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
			x := int(tv.GetUint())
			tv.T = t
			tv.SetInt(x)
		case Int8Kind:
			x := int8(tv.GetUint())
			tv.T = t
			tv.SetInt8(x)
		case Int16Kind:
			x := int16(tv.GetUint())
			tv.T = t
			tv.SetInt16(x)
		case Int32Kind:
			x := int32(tv.GetUint())
			tv.T = t
			tv.SetInt32(x)
		case Int64Kind:
			x := int64(tv.GetUint())
			tv.T = t
			tv.SetInt64(x)
		case UintKind:
			x := tv.GetUint()
			tv.T = t
			tv.SetUint(x)
		case Uint8Kind:
			x := uint8(tv.GetUint())
			tv.T = t
			tv.SetUint8(x)
		case Uint16Kind:
			x := uint16(tv.GetUint())
			tv.T = t
			tv.SetUint16(x)
		case Uint32Kind:
			x := uint32(tv.GetUint())
			tv.T = t
			tv.SetUint32(x)
		case Uint64Kind:
			x := uint64(tv.GetUint())
			tv.T = t
			tv.SetUint64(x)
		case Float32Kind:
			x := float32(tv.GetUint()) // XXX determinism?
			tv.T = t
			tv.SetFloat32(x)
		case Float64Kind:
			x := float64(tv.GetUint()) // XXX determinism?
			tv.T = t
			tv.SetFloat64(x)
		case StringKind:
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
			x := int(tv.GetUint8())
			tv.T = t
			tv.SetInt(x)
		case Int8Kind:
			x := int8(tv.GetUint8())
			tv.T = t
			tv.SetInt8(x)
		case Int16Kind:
			x := int16(tv.GetUint8())
			tv.T = t
			tv.SetInt16(x)
		case Int32Kind:
			x := int32(tv.GetUint8())
			tv.T = t
			tv.SetInt32(x)
		case Int64Kind:
			x := int64(tv.GetUint8())
			tv.T = t
			tv.SetInt64(x)
		case UintKind:
			x := uint(tv.GetUint8())
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
			x := float32(tv.GetUint8()) // XXX determinism?
			tv.T = t
			tv.SetFloat32(x)
		case Float64Kind:
			x := float64(tv.GetUint8()) // XXX determinism?
			tv.T = t
			tv.SetFloat64(x)
		case StringKind:
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
			x := int(tv.GetUint16())
			tv.T = t
			tv.SetInt(x)
		case Int8Kind:
			x := int8(tv.GetUint16())
			tv.T = t
			tv.SetInt8(x)
		case Int16Kind:
			x := int16(tv.GetUint16())
			tv.T = t
			tv.SetInt16(x)
		case Int32Kind:
			x := int32(tv.GetUint16())
			tv.T = t
			tv.SetInt32(x)
		case Int64Kind:
			x := int64(tv.GetUint16())
			tv.T = t
			tv.SetInt64(x)
		case UintKind:
			x := uint(tv.GetUint16())
			tv.T = t
			tv.SetUint(x)
		case Uint8Kind:
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
			x := float32(tv.GetUint16()) // XXX determinism?
			tv.T = t
			tv.SetFloat32(x)
		case Float64Kind:
			x := float64(tv.GetUint16()) // XXX determinism?
			tv.T = t
			tv.SetFloat64(x)
		case StringKind:
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
			x := int(tv.GetUint32())
			tv.T = t
			tv.SetInt(x)
		case Int8Kind:
			x := int8(tv.GetUint32())
			tv.T = t
			tv.SetInt8(x)
		case Int16Kind:
			x := int16(tv.GetUint32())
			tv.T = t
			tv.SetInt16(x)
		case Int32Kind:
			x := int32(tv.GetUint32())
			tv.T = t
			tv.SetInt32(x)
		case Int64Kind:
			x := int64(tv.GetUint32())
			tv.T = t
			tv.SetInt64(x)
		case UintKind:
			x := uint(tv.GetUint32())
			tv.T = t
			tv.SetUint(x)
		case Uint8Kind:
			x := uint8(tv.GetUint32())
			tv.T = t
			tv.SetUint8(x)
		case Uint16Kind:
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
			x := float32(tv.GetUint32()) // XXX determinism?
			tv.T = t
			tv.SetFloat32(x)
		case Float64Kind:
			x := float64(tv.GetUint32()) // XXX determinism?
			tv.T = t
			tv.SetFloat64(x)
		case StringKind:
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
			x := int(tv.GetUint64())
			tv.T = t
			tv.SetInt(x)
		case Int8Kind:
			x := int8(tv.GetUint64())
			tv.T = t
			tv.SetInt8(x)
		case Int16Kind:
			x := int16(tv.GetUint64())
			tv.T = t
			tv.SetInt16(x)
		case Int32Kind:
			x := int32(tv.GetUint64())
			tv.T = t
			tv.SetInt32(x)
		case Int64Kind:
			x := int64(tv.GetUint64())
			tv.T = t
			tv.SetInt64(x)
		case UintKind:
			x := uint(tv.GetUint64())
			tv.T = t
			tv.SetUint(x)
		case Uint8Kind:
			x := uint8(tv.GetUint64())
			tv.T = t
			tv.SetUint8(x)
		case Uint16Kind:
			x := uint16(tv.GetUint64())
			tv.T = t
			tv.SetUint16(x)
		case Uint32Kind:
			x := uint32(tv.GetUint64())
			tv.T = t
			tv.SetUint32(x)
		case Uint64Kind:
			x := tv.GetUint64()
			tv.T = t
			tv.SetUint64(x)
		case Float32Kind:
			x := float32(tv.GetUint64()) // XXX determinism?
			tv.T = t
			tv.SetFloat32(x)
		case Float64Kind:
			x := float64(tv.GetUint64()) // XXX determinism?
			tv.T = t
			tv.SetFloat64(x)
		case StringKind:
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
			x := int(tv.GetFloat32()) // XXX determinism?
			tv.T = t
			tv.SetInt(x)
		case Int8Kind:
			x := int8(tv.GetFloat32()) // XXX determinism?
			tv.T = t
			tv.SetInt8(x)
		case Int16Kind:
			x := int16(tv.GetFloat32()) // XXX determinism?
			tv.T = t
			tv.SetInt16(x)
		case Int32Kind:
			x := int32(tv.GetFloat32()) // XXX determinism?
			tv.T = t
			tv.SetInt32(x)
		case Int64Kind:
			x := int64(tv.GetFloat32()) // XXX determinism?
			tv.T = t
			tv.SetInt64(x)
		case UintKind:
			x := uint(tv.GetFloat32()) // XXX determinism?
			tv.T = t
			tv.SetUint(x)
		case Uint8Kind:
			x := uint8(tv.GetFloat32()) // XXX determinism?
			tv.T = t
			tv.SetUint8(x)
		case Uint16Kind:
			x := uint16(tv.GetFloat32()) // XXX determinism?
			tv.T = t
			tv.SetUint16(x)
		case Uint32Kind:
			x := uint32(tv.GetFloat32()) // XXX determinism?
			tv.T = t
			tv.SetUint32(x)
		case Uint64Kind:
			x := uint64(tv.GetFloat32()) // XXX determinism?
			tv.T = t
			tv.SetUint64(x)
		case Float32Kind:
			x := tv.GetFloat32() // XXX determinism?
			tv.T = t
			tv.SetFloat32(x)
		case Float64Kind:
			x := float64(tv.GetFloat32()) // XXX determinism?
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
			x := int(tv.GetFloat64()) // XXX determinism?
			tv.T = t
			tv.SetInt(x)
		case Int8Kind:
			x := int8(tv.GetFloat64()) // XXX determinism?
			tv.T = t
			tv.SetInt8(x)
		case Int16Kind:
			x := int16(tv.GetFloat64()) // XXX determinism?
			tv.T = t
			tv.SetInt16(x)
		case Int32Kind:
			x := int32(tv.GetFloat64()) // XXX determinism?
			tv.T = t
			tv.SetInt32(x)
		case Int64Kind:
			x := int64(tv.GetFloat64()) // XXX determinism?
			tv.T = t
			tv.SetInt64(x)
		case UintKind:
			x := uint(tv.GetFloat64()) // XXX determinism?
			tv.T = t
			tv.SetUint(x)
		case Uint8Kind:
			x := uint8(tv.GetFloat64()) // XXX determinism?
			tv.T = t
			tv.SetUint8(x)
		case Uint16Kind:
			x := uint16(tv.GetFloat64()) // XXX determinism?
			tv.T = t
			tv.SetUint16(x)
		case Uint32Kind:
			x := uint32(tv.GetFloat64()) // XXX determinism?
			tv.T = t
			tv.SetUint32(x)
		case Uint64Kind:
			x := uint64(tv.GetFloat64()) // XXX determinism?
			tv.T = t
			tv.SetUint64(x)
		case Float32Kind:
			x := float32(tv.GetFloat64()) // XXX determinism?
			tv.T = t
			tv.SetFloat32(x)
		case Float64Kind:
			x := tv.GetFloat64() // XXX determinism?
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
				tv.V = alloc.NewSliceFromData([]byte(tv.GetString()))
				tv.T = t // after tv.GetString()
			case Int32Kind:
				runes := []TypedValue{}
				str := tv.GetString()
				for _, r := range str {
					runes = append(runes, typedRune(r))
				}
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
					if tk == Uint8Kind {
						data := make([]byte, svl)
						copyListToData(
							data[:svl],
							svb.List[svo:svo+svl])
						strv := alloc.NewString(string(data))
						tv.T = t
						tv.V = strv
					} else if tk == Int32Kind {
						runes := make([]rune, svl)
						copyListToRunes(
							runes[:svl],
							svb.List[svo:svo+svl])
						strv := alloc.NewString(string(runes))
						tv.T = t
						tv.V = strv
					} else {
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
				} else if tvpt.Specificity() > pt.Specificity() {
					// ok
				}
			} else {
				panic(fmt.Sprintf(
					"ConvertUntypedTo expects more specific target for %v but got %s",
					tv.String(),
					t.String()))
			}
		}
	}
	// special case: native
	if nt, ok := t.(*NativeType); ok {
		// first convert untyped to typed gno value.
		gnot := go2GnoBaseType(nt.Type)
		if debug {
			if _, ok := gnot.(*NativeType); ok {
				panic("should not happen")
			}
		}
		ConvertUntypedTo(tv, gnot)
		// then convert to native value.
		// NOTE: this should only be called during preprocessing, so no alloc needed.
		ConvertTo(nilAllocator, nil, tv, t)
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
		if debug {
			if t.Kind() != BoolKind {
				panic("untyped bool can only be converted to bool kind")
			}
		}
		tv.T = t
	case UntypedRuneType:
		ConvertUntypedRuneTo(tv, t)
	case UntypedBigintType:
		if preprocessing.Load() == 0 {
			panic("untyped Bigint conversion should not happen during interpretation")
		}
		ConvertUntypedBigintTo(tv, tv.V.(BigintValue), t)
	case UntypedBigdecType:
		if preprocessing.Load() == 0 {
			panic("untyped Bigdec conversion should not happen during interpretation")
		}
		ConvertUntypedBigdecTo(tv, tv.V.(BigdecValue), t)
	case UntypedStringType:
		if preprocessing.Load() == 0 {
			panic("untyped String conversion should not happen during interpretation")
		}
		if t.Kind() == StringKind {
			tv.T = t
			return
		} else {
			ConvertTo(nilAllocator, nil, tv, t)
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
		dst.SetInt(int(sv))
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
		dst.SetUint(uint(sv))
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

func ConvertUntypedBigintTo(dst *TypedValue, bv BigintValue, t Type) {
	k := t.Kind()
	bi := bv.V
	var sv int64 = 0  // if signed.
	var uv uint64 = 0 // if unsigned.
	switch k {
	case BigintKind:
		dst.T = t
		dst.V = bv
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
		dst.SetFloat32(f32)
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
		dst.SetFloat64(f64)
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
		if strconv.IntSize == 32 {
			if math.MaxInt32 < sv {
				panic("bigint overflows target kind")
			}
			if sv < math.MinInt32 {
				panic("bigint underflows target kind")
			}
			dst.SetInt(int(sv))
		} else if strconv.IntSize == 64 {
			dst.SetInt(int(sv))
		} else {
			panic("unexpected IntSize")
		}
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
		if strconv.IntSize == 32 {
			if math.MaxUint32 < uv {
				panic("bigint overflows target kind")
			}
			dst.SetUint(uint(uv))
		} else if strconv.IntSize == 64 {
			dst.SetUint(uint(uv))
		} else {
			panic("unexpected IntSize")
		}
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

func ConvertUntypedBigdecTo(dst *TypedValue, bv BigdecValue, t Type) {
	k := t.Kind()
	bd := bv.V
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
		dst.SetFloat64(f)
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
		dst.SetFloat32(f32)
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
		dst.SetFloat64(f64)
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
