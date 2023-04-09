package gnolang

import (
	"encoding/base64"
	"fmt"
	"strconv"
)

// These convert string representations of public-facing arguments to GNO types.
// The limited set of input types available should map 1:1 to types supported
// in FunctionSignature{}.
// String representation of arg must be deterministic.
// NOTE: very important that there is no malleability.
func ConvertArgToGno(arg string, argT Type) (tv TypedValue) {
	tv.T = argT
	switch bt := BaseOf(argT).(type) {
	case PrimitiveType:
		switch bt {
		case BoolType:
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
		case StringType:
			tv.SetString(StringValue(arg))
			return
		case IntType:
			if arg[0] == '+' {
				panic("numbers cannot start with +")
			}
			i64, err := strconv.ParseInt(arg, 10, 64)
			if err != nil {
				panic(fmt.Sprintf(
					"error parsing int %q: %v",
					arg, err))
			}
			tv.SetInt(int(i64))
			return
		case Int8Type:
			if arg[0] == '+' {
				panic("numbers cannot start with +")
			}
			i8, err := strconv.ParseInt(arg, 10, 8)
			if err != nil {
				panic(fmt.Sprintf(
					"error parsing int8 %q: %v",
					arg, err))
			}
			tv.SetInt8(int8(i8))
			return
		case Int16Type:
			if arg[0] == '+' {
				panic("numbers cannot start with +")
			}
			i16, err := strconv.ParseInt(arg, 10, 16)
			if err != nil {
				panic(fmt.Sprintf(
					"error parsing int16 %q: %v",
					arg, err))
			}
			tv.SetInt16(int16(i16))
			return
		case Int32Type:
			if arg[0] == '+' {
				panic("numbers cannot start with +")
			}
			i32, err := strconv.ParseInt(arg, 10, 32)
			if err != nil {
				panic(fmt.Sprintf(
					"error parsing int32 %q: %v",
					arg, err))
			}
			tv.SetInt32(int32(i32))
			return
		case Int64Type:
			if arg[0] == '+' {
				panic("numbers cannot start with +")
			}
			i64, err := strconv.ParseInt(arg, 10, 64)
			if err != nil {
				panic(fmt.Sprintf(
					"error parsing int64 %q: %v",
					arg, err))
			}
			tv.SetInt64(i64)
			return
		case UintType:
			if arg[0] == '+' {
				panic("numbers cannot start with +")
			}
			u64, err := strconv.ParseUint(arg, 10, 64)
			if err != nil {
				panic(fmt.Sprintf(
					"error parsing uint %q: %v",
					arg, err))
			}
			tv.SetUint(uint(u64))
			return
		case Uint8Type:
			if arg[0] == '+' {
				panic("numbers cannot start with +")
			}
			u8, err := strconv.ParseUint(arg, 10, 8)
			if err != nil {
				panic(fmt.Sprintf(
					"error parsing uint8 %q: %v",
					arg, err))
			}
			tv.SetUint8(uint8(u8))
			return
		case Uint16Type:
			if arg[0] == '+' {
				panic("numbers cannot start with +")
			}
			u16, err := strconv.ParseUint(arg, 10, 16)
			if err != nil {
				panic(fmt.Sprintf(
					"error parsing uint16 %q: %v",
					arg, err))
			}
			tv.SetUint16(uint16(u16))
			return
		case Uint32Type:
			if arg[0] == '+' {
				panic("numbers cannot start with +")
			}
			u32, err := strconv.ParseUint(arg, 10, 32)
			if err != nil {
				panic(fmt.Sprintf(
					"error parsing uint32 %q: %v",
					arg, err))
			}
			tv.SetUint32(uint32(u32))
			return
		case Uint64Type:
			if arg[0] == '+' {
				panic("numbers cannot start with +")
			}
			u64, err := strconv.ParseUint(arg, 10, 64)
			if err != nil {
				panic(fmt.Sprintf(
					"error parsing uint64 %q: %v",
					arg, err))
			}
			tv.SetUint64(uint64(u64))
			return
		default:
			panic(fmt.Sprintf("unexpected primitive type %s", bt.String()))
		}
	case *ArrayType:
		if bt.Elt == Uint8Type {
			bz, err := base64.StdEncoding.DecodeString(arg)
			if err != nil {
				panic(fmt.Sprintf(
					"error parsing byte array %q: %v",
					arg, err))
			}
			tv.V = &ArrayValue{
				Data: bz,
			}
			return
		} else {
			panic("unexpected array type in contract arg")
		}
	case *SliceType:
		if bt.Elt == Uint8Type {
			bz, err := base64.StdEncoding.DecodeString(arg)
			if err != nil {
				panic(fmt.Sprintf(
					"error parsing byte array %q: %v",
					arg, err))
			}
			tv.V = &SliceValue{
				Base: &ArrayValue{
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
