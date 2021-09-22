package gno

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
)

// t cannot be nil or untyped or DataByteType.
// the conversion is forced and overflow/underflow is ignored.
func ConvertTo(store Store, tv *TypedValue, t Type) {
	if debug {
		if t == nil {
			panic("ConvertTo() requires non-nil type")
		}
		if isUntyped(t) {
			panic("cannot convert to untyped type")
		}
		if tv.T == DataByteType || t == DataByteType {
			panic("should not happen")
		}
	}
	// special case for go-native conversions
	ntv, tvIsNat := tv.T.(*nativeType)
	nt, tIsNat := t.(*nativeType)
	if tvIsNat {
		if tIsNat {
			// both nativeType, use reflect to assert.
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
			// convert go-native to gno type.
			*tv = go2GnoValue2(tv.V.(*nativeValue).Value)
			ConvertTo(store, tv, t)
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
				V: &nativeValue{Value: rv},
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
			x := int(tv.GetInt())
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
		case StringKind:
			tv.V = StringValue(string(rune(tv.GetInt())))
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
			x := int8(tv.GetInt8())
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
		case StringKind:
			tv.V = StringValue(string(rune(tv.GetInt8())))
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
			x := int16(tv.GetInt16())
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
		case StringKind:
			tv.V = StringValue(string(rune(tv.GetInt16())))
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
			x := int32(tv.GetInt32())
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
		case StringKind:
			tv.V = StringValue(string(rune(tv.GetInt32())))
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
			x := int64(tv.GetInt64())
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
		case StringKind:
			tv.V = StringValue(string(rune(tv.GetInt64())))
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
			x := uint(tv.GetUint())
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
		case StringKind:
			tv.V = StringValue(string(rune(tv.GetUint())))
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
			x := uint8(tv.GetUint8())
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
		case StringKind:
			tv.V = StringValue(string(rune(tv.GetUint8())))
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
			x := uint16(tv.GetUint16())
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
		case StringKind:
			tv.V = StringValue(string(rune(tv.GetUint16())))
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
			x := uint32(tv.GetUint32())
			tv.T = t
			tv.SetUint32(x)
		case Uint64Kind:
			x := uint64(tv.GetUint32())
			tv.T = t
			tv.SetUint64(x)
		case StringKind:
			tv.V = StringValue(string(rune(tv.GetUint32())))
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
			x := uint64(tv.GetUint64())
			tv.T = t
			tv.SetUint64(x)
		case StringKind:
			tv.V = StringValue(string(rune(tv.GetUint64())))
			tv.T = t
			tv.ClearNum()
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
				tv.V = newSliceFromData([]byte(tv.GetString()))
				tv.T = t // after tv.GetString()
			case Int32Kind:
				runes := []TypedValue{}
				str := tv.GetString()
				for _, r := range str {
					runes = append(runes, typedRune(r))
				}
				tv.V = newSliceFromList(runes)
				tv.T = t // after tv.GetString()
			default:
				panic(fmt.Sprintf(
					"cannot convert %s to %s",
					tvk.String(), t.String()))
			}
			/* TODO deleteme, native types handled above.
			case *nativeType:
				switch cbt.Kind() {
				case StringKind:
					tv.V = &nativeValue{
						Value: reflect.ValueOf(
							string(tv.GetString()),
						),
					}
					tv.T = t // after tv.GetString()
				case SliceKind:
					tv.V = &nativeValue{
						Value: reflect.ValueOf(
							[]byte(tv.GetString()),
						),
					}
					tv.T = t // after tv.GetString()
				case InterfaceKind:
					tv.T = StringType
				default:
					panic(fmt.Sprintf(
						"cannot convert %s to %s",
						tvk.String(), t.String()))
				}
			*/
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
						strv := StringValue(string(data))
						tv.T = t
						tv.V = strv
					} else if tk == Int32Kind {
						runes := make([]rune, svl)
						copyListToRunes(
							runes[:svl],
							svb.List[svo:svo+svl])
						strv := StringValue(string(runes))
						tv.T = t
						tv.V = strv
					} else {
						panic("should not happen")
					}
				} else {
					data := svb.Data[svo : svo+svl]
					strv := StringValue(string(data))
					tv.T = t
					tv.V = strv
				}
				/* TODO deleteme, native types handled above
				case *nativeValue:
					data := sv.Value.Bytes()
					strv := StringValue(string(data))
					tv.T = t
					tv.V = strv
				*/
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
			"cannot convert %s to any other kind",
			tvk.String()))
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
			panic(fmt.Sprintf(
				"ConvertUntypedTo expects typed target but got %s",
				t.String()))
		}
	}
	// special case: native
	if nt, ok := t.(*nativeType); ok {
		// first convert untyped to typed gno value.
		gnot := go2GnoBaseType(nt.Type)
		if debug {
			if _, ok := gnot.(*nativeType); ok {
				panic("should not happen")
			}
		}
		ConvertUntypedTo(tv, gnot)
		// then convert to native value.
		ConvertTo(nil, tv, t)
	}
	// special case: simple conversion
	if t != nil && tv.T.Kind() == t.Kind() {
		tv.T = t
		return
	}
	// general case
	switch tv.T {
	case UntypedBoolType:
		if t == nil {
			t = BoolType
		}
		if debug {
			if t.Kind() != BoolKind {
				panic("untyped bool can only be converted to bool kind")
			}
		}
		tv.T = t
	case UntypedRuneType:
		if t == nil {
			t = Int32Type
		}
		ConvertUntypedRuneTo(tv, t)
	case UntypedBigintType:
		if t == nil {
			t = IntType
		}
		ConvertUntypedBigintTo(tv, tv.V.(BigintValue), t)
	case UntypedStringType:
		if t == nil {
			t = StringType
		}
		if t.Kind() == StringKind {
			tv.T = t
			return
		} else {
			ConvertTo(nil, tv, t)
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
	case StringKind:
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
		dst.SetInt32(int32(sv))
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
	default:
		panic(fmt.Sprintf("unexpected target %v", k))

	}
}

// All fields may be modified to complete the conversion.
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
