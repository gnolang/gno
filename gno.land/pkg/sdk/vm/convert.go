package vm

import (
	"encoding/base64"
	"fmt"
	"strconv"

	"github.com/cockroachdb/apd/v3"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

func assertCharNotPlus(b byte) {
	if b == '+' {
		panic("numbers cannot start with +")
	}
}

// These convert string representations of public-facing arguments to GNO types.
// The limited set of input types available should map 1:1 to types supported
// in FunctionSignature{}.
// String representation of arg must be deterministic.
// NOTE: very important that there is no malleability.
func convertArgToGno(store gno.Store, arg string, argT gno.Type) (gno.TypedValue, error) {
	tv := gno.TypedValue{T: argT}

	switch bt := gno.BaseOf(argT).(type) {
	case gno.PrimitiveType:
		switch bt {
		case gno.BoolType:
			// XXX: should we use `strconv.ParseBool` here ?
			if arg == "true" {
				tv.SetBool(true)
				return tv, nil
			}

			if arg == "false" {
				tv.SetBool(false)
				return tv, nil
			}

			return gno.TypedValue{}, fmt.Errorf("unexpected bool value %q", arg)
		case gno.StringType:
			tv.SetString(gno.StringValue(arg))
			return tv, nil
		case gno.IntType:
			assertCharNotPlus(arg[0])
			i64, err := strconv.ParseInt(arg, 10, 64)
			if err != nil {
				return gno.TypedValue{}, fmt.Errorf("error parsing int %q: %w", arg, err)
			}

			tv.SetInt(int(i64))
			return tv, nil
		case gno.Int8Type:
			assertCharNotPlus(arg[0])
			i8, err := strconv.ParseInt(arg, 10, 8)
			if err != nil {
				return gno.TypedValue{}, fmt.Errorf("error parsing int8 %q: %w", arg, err)
			}
			tv.SetInt8(int8(i8))
			return tv, nil
		case gno.Int16Type:
			assertCharNotPlus(arg[0])
			i16, err := strconv.ParseInt(arg, 10, 16)
			if err != nil {
				return gno.TypedValue{}, fmt.Errorf("error parsing int16 %q: %w", arg, err)
			}
			tv.SetInt16(int16(i16))
			return tv, nil
		case gno.Int32Type:
			assertCharNotPlus(arg[0])
			i32, err := strconv.ParseInt(arg, 10, 32)
			if err != nil {
				return gno.TypedValue{}, fmt.Errorf("error parsing int32 %q: %w", arg, err)
			}
			tv.SetInt32(int32(i32))
			return tv, nil
		case gno.Int64Type:
			assertCharNotPlus(arg[0])
			i64, err := strconv.ParseInt(arg, 10, 64)
			if err != nil {
				return gno.TypedValue{}, fmt.Errorf("error parsing int64 %q: %w", arg, err)
			}
			tv.SetInt64(i64)
			return tv, nil
		case gno.UintType:
			assertCharNotPlus(arg[0])
			u64, err := strconv.ParseUint(arg, 10, 64)
			if err != nil {
				return gno.TypedValue{}, fmt.Errorf("error parsing uint %q: %w", arg, err)
			}
			tv.SetUint(uint(u64))
			return tv, nil
		case gno.Uint8Type:
			assertCharNotPlus(arg[0])
			u8, err := strconv.ParseUint(arg, 10, 8)
			if err != nil {
				return gno.TypedValue{}, fmt.Errorf("error parsing uint8 %q: %w", arg, err)
			}
			tv.SetUint8(uint8(u8))
			return tv, nil
		case gno.Uint16Type:
			assertCharNotPlus(arg[0])
			u16, err := strconv.ParseUint(arg, 10, 16)
			if err != nil {
				return gno.TypedValue{}, fmt.Errorf("error parsing uint16 %q: %w", arg, err)
			}
			tv.SetUint16(uint16(u16))
			return tv, nil
		case gno.Uint32Type:
			assertCharNotPlus(arg[0])
			u32, err := strconv.ParseUint(arg, 10, 32)
			if err != nil {
				return gno.TypedValue{}, fmt.Errorf("error parsing uint32 %q: %w", arg, err)
			}
			tv.SetUint32(uint32(u32))
			return tv, nil
		case gno.Uint64Type:
			assertCharNotPlus(arg[0])
			u64, err := strconv.ParseUint(arg, 10, 64)
			if err != nil {
				return gno.TypedValue{}, fmt.Errorf("error parsing uint64 %q: %w", arg, err)
			}
			tv.SetUint64(u64)
			return tv, nil
		case gno.Float32Type:
			value := convertFloat(arg, 32)
			tv.SetFloat32(float32(value))
			return tv, nil
		case gno.Float64Type:
			value := convertFloat(arg, 64)
			tv.SetFloat64(value)
			return tv, nil
		default:
			return gno.TypedValue{}, fmt.Errorf("unexpected primitive type %s", bt.String())
		}

	case *gno.ArrayType:
		if bt.Elt == gno.Uint8Type {
			bz, err := base64.StdEncoding.DecodeString(arg)
			if err != nil {
				return gno.TypedValue{}, fmt.Errorf("error parsing byte array %.50q: %w", arg, err)
			}
			tv.V = &gno.ArrayValue{
				Data: bz,
			}
			return tv, nil
		}

		utv, err := UnmarshalTypedValueJSON(store.GetAllocator(), store, []byte(arg), argT)
		if err != nil {
			return gno.TypedValue{}, fmt.Errorf("error parsing array %.50q: %w", arg, err)
		}

		return utv, nil

	case *gno.SliceType:
		if bt.Elt == gno.Uint8Type {
			bz, err := base64.StdEncoding.DecodeString(arg)
			if err != nil {
				return gno.TypedValue{}, fmt.Errorf("error parsing byte slice %.50q: %w", arg, err)
			}
			tv.V = &gno.SliceValue{
				Base: &gno.ArrayValue{
					Data: bz,
				},
				Offset: 0,
				Length: len(bz),
				Maxcap: len(bz),
			}
			return tv, nil
		}

		utv, err := UnmarshalTypedValueJSON(store.GetAllocator(), store, []byte(arg), argT)
		if err != nil {
			return gno.TypedValue{}, fmt.Errorf("error unmarshal slice %.50q: %w", arg, err)
		}

		return utv, nil
	default:
	}

	var kind gno.Kind
	if kind = argT.Kind(); kind == gno.PointerKind {
		kind = argT.Elem().Kind()
	}

	// If there is no struct or struct pointer, return an error. We don't
	// want to deal with nested pointers as arguments.
	if kind != gno.StructKind {
		return gno.TypedValue{}, fmt.Errorf("unexpected type in contract arg: %s(%q)", gno.KindOf(argT).String(), argT)
	}

	// Handle struct kind
	utv, err := UnmarshalTypedValueJSON(store.GetAllocator(), store, []byte(arg), argT)
	if err != nil {
		return gno.TypedValue{}, fmt.Errorf("error unmarshal struct %.50q: %w", arg, err)
	}

	return utv, nil
}

func convertFloat(value string, precision int) float64 {
	assertCharNotPlus(value[0])
	dec, _, err := apd.NewFromString(value)
	if err != nil {
		panic(fmt.Sprintf("error parsing float%d %s: %v", precision, value, err))
	}

	f64, err := strconv.ParseFloat(dec.String(), precision)
	if err != nil {
		panic(fmt.Sprintf("error value exceeds float%d precision %s: %v", precision, value, err))
	}

	return f64
}
