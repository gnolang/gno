package vm

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/cockroachdb/apd/v3"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

func assertNoPlusPrefix(s string) {
	if strings.HasPrefix(s, "+") {
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
			assertNoPlusPrefix(arg)
			i64, err := strconv.ParseInt(arg, 10, 64)
			if err != nil {
				panic(fmt.Sprintf(
					"error parsing int %q: %v",
					arg, err))
			}
			tv.SetInt(int(i64))
			return
		case gno.Int8Type:
			assertNoPlusPrefix(arg)
			i8, err := strconv.ParseInt(arg, 10, 8)
			if err != nil {
				panic(fmt.Sprintf(
					"error parsing int8 %q: %v",
					arg, err))
			}
			tv.SetInt8(int8(i8))
			return
		case gno.Int16Type:
			assertNoPlusPrefix(arg)
			i16, err := strconv.ParseInt(arg, 10, 16)
			if err != nil {
				panic(fmt.Sprintf(
					"error parsing int16 %q: %v",
					arg, err))
			}
			tv.SetInt16(int16(i16))
			return
		case gno.Int32Type:
			assertNoPlusPrefix(arg)
			i32, err := strconv.ParseInt(arg, 10, 32)
			if err != nil {
				panic(fmt.Sprintf(
					"error parsing int32 %q: %v",
					arg, err))
			}
			tv.SetInt32(int32(i32))
			return
		case gno.Int64Type:
			assertNoPlusPrefix(arg)
			i64, err := strconv.ParseInt(arg, 10, 64)
			if err != nil {
				panic(fmt.Sprintf(
					"error parsing int64 %q: %v",
					arg, err))
			}
			tv.SetInt64(i64)
			return
		case gno.UintType:
			assertNoPlusPrefix(arg)
			u64, err := strconv.ParseUint(arg, 10, 64)
			if err != nil {
				panic(fmt.Sprintf(
					"error parsing uint %q: %v",
					arg, err))
			}
			tv.SetUint(uint(u64))
			return
		case gno.Uint8Type:
			assertNoPlusPrefix(arg)
			u8, err := strconv.ParseUint(arg, 10, 8)
			if err != nil {
				panic(fmt.Sprintf(
					"error parsing uint8 %q: %v",
					arg, err))
			}
			tv.SetUint8(uint8(u8))
			return
		case gno.Uint16Type:
			assertNoPlusPrefix(arg)
			u16, err := strconv.ParseUint(arg, 10, 16)
			if err != nil {
				panic(fmt.Sprintf(
					"error parsing uint16 %q: %v",
					arg, err))
			}
			tv.SetUint16(uint16(u16))
			return
		case gno.Uint32Type:
			assertNoPlusPrefix(arg)
			u32, err := strconv.ParseUint(arg, 10, 32)
			if err != nil {
				panic(fmt.Sprintf(
					"error parsing uint32 %q: %v",
					arg, err))
			}
			tv.SetUint32(uint32(u32))
			return
		case gno.Uint64Type:
			assertNoPlusPrefix(arg)
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

func JSONValues(m *gno.Machine, tvs []gno.TypedValue) string {
	var str strings.Builder

	str.WriteRune('[')
	for i, tv := range tvs {
		if i > 0 {
			str.WriteRune(',')
		}
		str.WriteString(JSONValue(m, tv))
	}
	str.WriteRune(']')

	return str.String()
}

func JSONValue(m *gno.Machine, tv gno.TypedValue) string {
	if tv.T == nil {
		return "null"
	}

	switch bt := gno.BaseOf(tv.T).(type) {
	case gno.PrimitiveType:
		switch bt {
		case gno.IntType:
			return fmt.Sprintf("%d", tv.GetInt())
		case gno.Int8Type:
			return fmt.Sprintf("%d", tv.GetInt8())
		case gno.Int16Type:
			return fmt.Sprintf("%d", tv.GetInt16())
		case gno.UntypedRuneType, gno.Int32Type:
			return fmt.Sprintf("%d", tv.GetInt32())
		case gno.Int64Type:
			return fmt.Sprintf("%d", tv.GetInt64())
		case gno.UintType:
			return fmt.Sprintf("%d", tv.GetUint())
		case gno.Uint8Type:
			return fmt.Sprintf("%d", tv.GetUint8())
		case gno.DataByteType:
			return fmt.Sprintf("%d", tv.GetDataByte())
		case gno.Uint16Type:
			return fmt.Sprintf("%d", tv.GetUint16())
		case gno.Uint32Type:
			return fmt.Sprintf("%d", tv.GetUint32())
		case gno.Uint64Type:
			return fmt.Sprintf("%d", tv.GetUint64())
		case gno.Float32Type:
			return fmt.Sprintf("%f", tv.GetFloat32())
		case gno.Float64Type:
			return fmt.Sprintf("%f", tv.GetFloat64())
		case gno.UntypedBigintType, gno.BigintType:
			return tv.V.(gno.BigintValue).V.String()
		case gno.UntypedBigdecType, gno.BigdecType:
			return tv.V.(gno.BigdecValue).V.String()
		case gno.UntypedBoolType, gno.BoolType:
			return fmt.Sprintf("%t", tv.GetBool())
		case gno.UntypedStringType, gno.StringType:
			return strconv.Quote(tv.GetString())
		default:
			panic("invalid primitive type - should not happen")
		}
	case *gno.PointerType:
		// Check if Pointer we type implement Stringer / Error

		// If implements .Error(), return it.
		if tv.IsError() {
			res := m.Eval(gno.Call(gno.Sel(&gno.ConstExpr{TypedValue: tv}, "Error")))
			return strconv.Quote(res[0].GetString())
		}
		// If implements .String(), return it.
		if tv.IsStringer() {
			res := m.Eval(gno.Call(gno.Sel(&gno.ConstExpr{TypedValue: tv}, "String")))
			return strconv.Quote(res[0].GetString())
		}
	default:
		// Check if pointer wraped value can implement Stringer / Error
		ptv := gno.TypedValue{
			T: &gno.PointerType{Elt: tv.T},
			V: gno.PointerValue{TV: &tv, Base: tv.V},
		}

		// If implements .Error(), return it.
		if ptv.IsError() {
			res := m.Eval(gno.Call(gno.Sel(&gno.ConstExpr{TypedValue: ptv}, "Error")))
			return strconv.Quote(res[0].GetString())
		}
		// If implements .String(), return it.
		if ptv.IsStringer() {
			res := m.Eval(gno.Call(gno.Sel(&gno.ConstExpr{TypedValue: ptv}, "String")))
			return strconv.Quote(res[0].GetString())
		}
	}

	if tv.V == nil {
		return "null"
	}

	var id string
	if pv, ok := tv.V.(gno.PointerValue); ok {
		id = pv.GetBase(m.Store).GetObjectID().String()
	}

	return strconv.Quote(fmt.Sprintf(`<%s:%s>`, tv.T.String(), id))
}

func convertFloat(value string, precision int) float64 {
	assertNoPlusPrefix(value)
	dec, _, err := apd.NewFromString(value)
	if err != nil {
		panic(fmt.Sprintf("error parsing float%d %q: %v", precision, value, err))
	}

	f64, err := strconv.ParseFloat(dec.String(), precision)
	if err != nil {
		panic(fmt.Sprintf("error value exceeds float%d precision %q: %v", precision, value, err))
	}

	return f64
}
