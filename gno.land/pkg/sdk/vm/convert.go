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
func convertArgToGno(arg string, argT gno.Type) (tv gno.TypedValue) {
	tv.T = argT
	switch bt := gno.BaseOf(argT).(type) {
	case gno.PrimitiveType:
		switch bt {
		case gno.BoolType:
			if arg == "true" {
				tv.SetBool(true)
				return
			} else if arg == "false" {
				tv.SetBool(false)
				return
			} else {
				panic(fmt.Sprintf(
					"unexpected bool value %q",
					arg))
			}
		case gno.StringType:
			tv.SetString(gno.StringValue(arg))
			return
		case gno.IntType:
			assertCharNotPlus(arg[0])
			i64, err := strconv.ParseInt(arg, 10, 64)
			if err != nil {
				panic(fmt.Sprintf(
					"error parsing int %q: %v",
					arg, err))
			}
			tv.SetInt(int(i64))
			return
		case gno.Int8Type:
			assertCharNotPlus(arg[0])
			i8, err := strconv.ParseInt(arg, 10, 8)
			if err != nil {
				panic(fmt.Sprintf(
					"error parsing int8 %q: %v",
					arg, err))
			}
			tv.SetInt8(int8(i8))
			return
		case gno.Int16Type:
			assertCharNotPlus(arg[0])
			i16, err := strconv.ParseInt(arg, 10, 16)
			if err != nil {
				panic(fmt.Sprintf(
					"error parsing int16 %q: %v",
					arg, err))
			}
			tv.SetInt16(int16(i16))
			return
		case gno.Int32Type:
			assertCharNotPlus(arg[0])
			i32, err := strconv.ParseInt(arg, 10, 32)
			if err != nil {
				panic(fmt.Sprintf(
					"error parsing int32 %q: %v",
					arg, err))
			}
			tv.SetInt32(int32(i32))
			return
		case gno.Int64Type:
			assertCharNotPlus(arg[0])
			i64, err := strconv.ParseInt(arg, 10, 64)
			if err != nil {
				panic(fmt.Sprintf(
					"error parsing int64 %q: %v",
					arg, err))
			}
			tv.SetInt64(i64)
			return
		case gno.UintType:
			assertCharNotPlus(arg[0])
			u64, err := strconv.ParseUint(arg, 10, 64)
			if err != nil {
				panic(fmt.Sprintf(
					"error parsing uint %q: %v",
					arg, err))
			}
			tv.SetUint(uint(u64))
			return
		case gno.Uint8Type:
			assertCharNotPlus(arg[0])
			u8, err := strconv.ParseUint(arg, 10, 8)
			if err != nil {
				panic(fmt.Sprintf(
					"error parsing uint8 %q: %v",
					arg, err))
			}
			tv.SetUint8(uint8(u8))
			return
		case gno.Uint16Type:
			assertCharNotPlus(arg[0])
			u16, err := strconv.ParseUint(arg, 10, 16)
			if err != nil {
				panic(fmt.Sprintf(
					"error parsing uint16 %q: %v",
					arg, err))
			}
			tv.SetUint16(uint16(u16))
			return
		case gno.Uint32Type:
			assertCharNotPlus(arg[0])
			u32, err := strconv.ParseUint(arg, 10, 32)
			if err != nil {
				panic(fmt.Sprintf(
					"error parsing uint32 %q: %v",
					arg, err))
			}
			tv.SetUint32(uint32(u32))
			return
		case gno.Uint64Type:
			assertCharNotPlus(arg[0])
			u64, err := strconv.ParseUint(arg, 10, 64)
			if err != nil {
				panic(fmt.Sprintf(
					"error parsing uint64 %q: %v",
					arg, err))
			}
			tv.SetUint64(u64)
			return
		case gno.Float32Type:
			value := convertFloat(arg, 32)
			tv.SetFloat32(float32(value))
			return
		case gno.Float64Type:
			value := convertFloat(arg, 64)
			tv.SetFloat64(value)
			return
		default:
			panic(fmt.Sprintf("unexpected primitive type %s", bt.String()))
		}
	case *gno.ArrayType:
		if bt.Elt == gno.Uint8Type {
			bz, err := base64.StdEncoding.DecodeString(arg)
			if err != nil {
				panic(fmt.Sprintf(
					"error parsing byte array %q: %v",
					arg, err))
			}
			tv.V = &gno.ArrayValue{
				Data: bz,
			}
			return
		} else {
			panic("unexpected array type in contract arg")
		}
	case *gno.SliceType:
		if bt.Elt == gno.Uint8Type {
			bz, err := base64.StdEncoding.DecodeString(arg)
			if err != nil {
				panic(fmt.Sprintf(
					"error parsing byte array %q: %v",
					arg, err))
			}
			tv.V = &gno.SliceValue{
				Base: &gno.ArrayValue{
					Data: bz,
				},
				Offset: 0,
				Length: len(bz),
				Maxcap: len(bz),
			}
			return
		} else {
			panic("unexpected slice type in contract arg")
		}
	default:
		panic(fmt.Sprintf("unexpected type in contract arg: %v", argT))
	}
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
