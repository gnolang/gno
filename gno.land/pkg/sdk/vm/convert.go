package vm

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/cockroachdb/apd/v3"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
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
			switch arg {
			case "true":
				tv.SetBool(true)
				return
			case "false":
				tv.SetBool(false)
				return
			default:
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

type jsonResults struct {
	Results json.RawMessage `json:"results"`
	Error   *string         `json:"@error,omitempty"`
}

// stringifyJSONResults converts TypedValues to JSON format using Amino marshaling.
// lastReturnType is the function signature's last return type (if available).
// When lastReturnType is provided, error detection is signature-based (per spec):
// only extract @error if the function signature declares an error return type.
// When lastReturnType is nil, fallback to value-based detection (for QueryEval).
func stringifyJSONResults(m *gno.Machine, tvs []gno.TypedValue, lastReturnType gno.Type) string {
	jres := jsonResults{Results: []byte("[]")}
	if len(tvs) > 0 {
		var err error

		// Use Amino-based JSON export with @type tags for consistency with qobject.
		// This ensures all values (including raw objects) are properly serialized.
		if jres.Results, err = gnolang.JSONExportTypedValues(tvs, nil); err != nil {
			panic("unable to marshal results")
		}

		// Check for error based on function signature (spec Step 2):
		// "If the func return type's last element is exactly a named or unnamed
		// interface type which implements error, then .Error() is called."
		last := tvs[len(tvs)-1]
		shouldExtractError := false
		if lastReturnType != nil {
			// Signature-based: check if declared return type implements error
			shouldExtractError = gno.IsErrorType(lastReturnType)
		} else {
			// Fallback for QueryEval: value-based detection
			shouldExtractError = last.ImplError()
		}

		if shouldExtractError {
			if errStr, ok := tryGetError(m, last); ok {
				jres.Error = &errStr
			}
		}
	}

	s, err := json.Marshal(jres)
	if err != nil {
		panic("unable to marshal result")
	}

	return string(s)
}

func tryGetError(m *gno.Machine, tv gno.TypedValue) (string, bool) {
	bt := gno.BaseOf(tv.T)

	// Check if type implement Error
	if _, ok := bt.(*gno.PointerType); !ok {
		// Try wrapping the value
		tv = gno.TypedValue{
			T: &gno.PointerType{Elt: tv.T},
			V: gno.PointerValue{TV: &tv, Base: tv.V},
		}
	}

	// If implements .Error(), return this
	if tv.ImplError() {
		res := m.Eval(gno.Call(gno.Sel(&gno.ConstExpr{TypedValue: tv}, "Error")))
		return res[0].GetString(), true
	}

	return "", false
}

// func stringifyJSONPrimitiveValues(m *gno.Machine, tvs []gno.TypedValue) string {
// 	var str strings.Builder

// 	str.WriteRune('[')
// 	for i, tv := range tvs {
// 		if i > 0 {
// 			str.WriteRune(',')
// 		}
// 		str.WriteString(stringifyJSONPrimitiveValue(m, tv, i == len(tvs)-1))
// 	}
// 	str.WriteRune(']')

// 	return str.String()
// }

// func stringifyJSONPrimitiveValues(m *gno.Machine, tvs []gno.TypedValue) string {
// 	var str strings.Builder

// 	str.WriteRune('[')
// 	for i, tv := range tvs {
// 		if i > 0 {
// 			str.WriteRune(',')
// 		}
// 		str.WriteString(stringifyJSONPrimitiveValue(m, tv, i == len(tvs)-1))
// 	}
// 	str.WriteRune(']')

// 	return str.String()
// }

// func stringifyJSONPrimitiveValue(m *gno.Machine, tv gno.TypedValue, isLast bool) string {
// 	if tv.T == nil {
// 		return `{"T":null,"V":null}`
// 	}

// 	bt := gno.BaseOf(tv.T)
// 	switch bt := bt.(type) {
// 	case gno.PrimitiveType:
// 		if bt == gno.UntypedStringType || bt == gno.StringType {
// 			// Special case for string, as we want it under "V" and quoted.
// 			return fmt.Sprintf(`{"T":%q,"V":%q}`, tv.T.String(), tv.GetString())
// 		}

// 		v := getPrimitiveValue(bt, tv)
// 		return fmt.Sprintf(`{"T":%q,"N":%s}`, tv.T.String(), v)
// 	case *gno.ArrayType:
// 		if bt.Elt == gno.Uint8Type {
// 			arr := tv.V.(*gno.ArrayValue)
// 			if data := arr.Data; data != nil {
// 				i := arr.GetLength()
// 				v := base64.StdEncoding.EncodeToString(data[:i])
// 				return fmt.Sprintf(`{"T":%q,"V":%q}`, tv.T.String(), v)
// 			}
// 		}
// 	case *gno.SliceType:
// 		if bt.Elt == gno.Uint8Type {
// 			slice := tv.V.(*gno.SliceValue)
// 			if data := slice.GetBase(nil).Data; data != nil {
// 				i := slice.GetLength()
// 				v := base64.StdEncoding.EncodeToString(data[:i])
// 				return fmt.Sprintf(`{"T":%q,"V":%q}`, tv.T.String(), v)
// 			}
// 		}
// 	}

// 	if tv.V == nil {
// 		return fmt.Sprintf(`{"T":%q,"V":null}`, tv.T.String())
// 	}

// 	if isLast {
// 		// Only last element error can be unwrapped
// 		if err, ok := tryGetError(m, tv); ok {
// 			return fmt.Sprintf(`{"@error":%q}`, err)
// 		}
// 	}

// 	var oid gno.ObjectID
// 	switch v := tv.V.(type) {
// 	case gno.PointerValue:
// 		oid = v.GetBase(m.Store).GetObjectID()
// 	case gno.Object:
// 		oid = v.GetObjectID()
// 	}

// 	return fmt.Sprintf(`{"T":%q,"V":%q}`, tv.T.String(), oid.String())
// }

// func getPrimitiveValue(bt gno.PrimitiveType, tv gno.TypedValue) string {
// 	switch bt {
// 	case gno.UntypedBoolType, gno.BoolType:
// 		return fmt.Sprintf("%t", tv.GetBool())
// 	case gno.UntypedStringType, gno.StringType:
// 		return strconv.Quote(tv.GetString())
// 	case gno.Float32Type:
// 		f32 := math.Float32frombits(tv.GetFloat32())
// 		return fmt.Sprintf("%f", f32)
// 	case gno.Float64Type:
// 		f64 := math.Float64frombits(tv.GetFloat64())
// 		return fmt.Sprintf("%f", f64)
// 	case gno.UntypedBigintType:
// 		return tv.V.(gno.BigintValue).V.String()
// 	case gno.UntypedBigdecType:
// 		return tv.V.(gno.BigdecValue).V.String()
// 	case gno.IntType, gno.Int8Type, gno.Int16Type, gno.Int32Type, gno.UntypedRuneType, gno.Int64Type:
// 		return fmt.Sprintf("%d", getSignedIntValue(bt, tv))
// 	case gno.UintType, gno.Uint8Type, gno.Uint16Type, gno.Uint32Type, gno.Uint64Type, gno.DataByteType:
// 		return fmt.Sprintf("%d", getUnsignedIntValue(bt, tv))
// 	default:
// 		panic("invalid primitive type - should not happen")
// 	}
// }

// func getSignedIntValue(bt gno.PrimitiveType, tv gno.TypedValue) int64 {
// 	switch bt {
// 	case gno.IntType:
// 		return tv.GetInt()
// 	case gno.Int8Type:
// 		return int64(tv.GetInt8())
// 	case gno.Int16Type:
// 		return int64(tv.GetInt16())
// 	case gno.Int32Type, gno.UntypedRuneType:
// 		return int64(tv.GetInt32())
// 	case gno.Int64Type:
// 		return tv.GetInt64()
// 	default:
// 		panic("unexpected signed integer type")
// 	}
// }

// func getUnsignedIntValue(bt gno.PrimitiveType, tv gno.TypedValue) uint64 {
// 	switch bt {
// 	case gno.UintType:
// 		return tv.GetUint()
// 	case gno.Uint8Type, gno.DataByteType:
// 		return uint64(tv.GetUint8())
// 	case gno.Uint16Type:
// 		return uint64(tv.GetUint16())
// 	case gno.Uint32Type:
// 		return uint64(tv.GetUint32())
// 	case gno.Uint64Type:
// 		return tv.GetUint64()
// 	default:
// 		panic("unexpected unsigned integer type")
// 	}
// }
