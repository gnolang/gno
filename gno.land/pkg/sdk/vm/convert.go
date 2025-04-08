package vm

import (
	"encoding/base64"
	"fmt"
	"math"
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
			tv.SetInt(i64)
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
			tv.SetUint(u64)
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
			value := convertStringToFloat(arg, 32)
			tv.SetFloat32(math.Float32bits(float32(value)))
			return
		case gno.Float64Type:
			value := convertStringToFloat(arg, 64)
			tv.SetFloat64(math.Float64bits(value))
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

func convertStringToFloat(value string, precision int) float64 {
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

func JSONPrimitiveValues(m *gno.Machine, tvs []gno.TypedValue) string {
	var str strings.Builder

	str.WriteRune('[')
	for i, tv := range tvs {
		if i > 0 {
			str.WriteRune(',')
		}
		str.WriteString(JSONPrimitiveValue(m, tv))
	}
	str.WriteRune(']')

	return str.String()
}

func JSONPrimitiveValue(m *gno.Machine, tv gno.TypedValue) string {
	if tv.T == nil {
		return "null"
	}

	bt := gno.BaseOf(tv.T)
	switch bt := bt.(type) {
	case gno.PrimitiveType:
		switch bt {
		case gno.UntypedBoolType, gno.BoolType:
			return fmt.Sprintf("%t", tv.GetBool())
		case gno.UntypedStringType, gno.StringType:
			return strconv.Quote(tv.GetString())
		case gno.Float32Type:
			f32 := math.Float32frombits(tv.GetFloat32())
			return fmt.Sprintf("%f", f32)
		case gno.Float64Type:
			f64 := math.Float64frombits(tv.GetFloat64())
			return fmt.Sprintf("%f", f64)
		case gno.UntypedBigintType:
			return tv.V.(gno.BigintValue).V.String()
		case gno.UntypedBigdecType:
			return tv.V.(gno.BigdecValue).V.String()
		case gno.IntType, gno.Int8Type, gno.Int16Type, gno.Int32Type, gno.UntypedRuneType, gno.Int64Type:
			return fmt.Sprintf("%d", getSignedIntValue(bt, tv))
		case gno.UintType, gno.Uint8Type, gno.Uint16Type, gno.Uint32Type, gno.Uint64Type, gno.DataByteType:
			return fmt.Sprintf("%d", getUnsignedIntValue(bt, tv))
		default:
			panic("invalid primitive type - should not happen")
		}
	case *gno.ArrayType:
		if bt.Elt == gno.Uint8Type {
			arr := tv.V.(*gno.ArrayValue)
			if data := arr.Data; data != nil {
				i := arr.GetLength()
				return `"` + base64.StdEncoding.EncodeToString(data[:i]) + `"`
			}
		}
	case *gno.SliceType:
		if bt.Elt == gno.Uint8Type {
			slice := tv.V.(*gno.SliceValue)
			if data := slice.GetBase(nil).Data; data != nil {
				i := slice.GetLength()
				return `"` + base64.StdEncoding.EncodeToString(data[:i]) + `"`
			}
		}
	}

	if tv.V == nil {
		return "null"
	}

	if res, ok := tryGetError(m, tv); ok {
		return res
	}

	var oid gno.ObjectID
	switch v := tv.V.(type) {
	case gno.PointerValue:
		oid = v.GetBase(m.Store).GetObjectID()
	case gno.Object:
		oid = v.GetObjectID()
	}

	if !oid.IsZero() {
		return strconv.Quote(fmt.Sprintf(`<obj:%s:%s>`, tv.T.String(), oid))
	}

	return strconv.Quote(fmt.Sprintf(`<obj:%s:0>`, tv.T.String()))
}

func getSignedIntValue(bt gno.PrimitiveType, tv gno.TypedValue) int64 {
	switch bt {
	case gno.IntType:
		return tv.GetInt()
	case gno.Int8Type:
		return int64(tv.GetInt8())
	case gno.Int16Type:
		return int64(tv.GetInt16())
	case gno.Int32Type, gno.UntypedRuneType:
		return int64(tv.GetInt32())
	case gno.Int64Type:
		return tv.GetInt64()
	default:
		panic("unexpected signed integer type")
	}
}

func getUnsignedIntValue(bt gno.PrimitiveType, tv gno.TypedValue) uint64 {
	switch bt {
	case gno.UintType:
		return tv.GetUint()
	case gno.Uint8Type, gno.DataByteType:
		return uint64(tv.GetUint8())
	case gno.Uint16Type:
		return uint64(tv.GetUint16())
	case gno.Uint32Type:
		return uint64(tv.GetUint32())
	case gno.Uint64Type:
		return tv.GetUint64()
	default:
		panic("unexpected unsigned integer type")
	}
}

func tryGetError(m *gno.Machine, tv gno.TypedValue) (string, bool) {
	bt := gno.BaseOf(tv.T)

	// Check if type implement Error
	ptv := tv
	if _, ok := bt.(*gno.PointerType); !ok {
		// Try wrapping the valye
		ptv = gno.TypedValue{
			T: &gno.PointerType{Elt: tv.T},
			V: gno.PointerValue{TV: &tv, Base: tv.V},
		}
	}

	// If implements .Error(), return this
	if ptv.ImplError() {
		res := m.Eval(gno.Call(gno.Sel(&gno.ConstExpr{TypedValue: ptv}, "Error")))
		return strconv.Quote(res[0].GetString()), true
	}

	return "", false
}
