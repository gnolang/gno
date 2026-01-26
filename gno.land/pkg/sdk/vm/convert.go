package vm

import (
	"encoding/base64"
	"encoding/json"
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

// stringifyJSONResults converts TypedValues to JSON format.
// ft is the function type (if available) used for signature-based error detection.
// When ft is provided, @error is extracted only if the function signature declares
// an error return type. When ft is nil, fallback to value-based detection.
func stringifyJSONResults(m *gno.Machine, tvs []gno.TypedValue, ft *gno.FuncType) string {
	jres := jsonResults{Results: []byte("[]")}
	if len(tvs) > 0 {
		var err error

		opts := gno.JSONExporterOptions{MaxDepth: 10, ExportUnexported: true}
		if jres.Results, err = opts.ExportTypedValues(tvs); err != nil {
			panic("unable to marshal results")
		}

		// Check for error based on function signature. If the func return type's last
		// element is exactly a named or unnamed interface type which implements error,
		// then .Error() is called.
		last := tvs[len(tvs)-1]
		shouldExtractError := false
		if ft != nil && len(ft.Results) > 0 {
			// Signature-based: check if declared return type implements error
			lastReturnType := ft.Results[len(ft.Results)-1].Type
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
	// Check if type implements error interface
	if !tv.ImplError() {
		return "", false
	}

	// Call .Error() method using the same approach as TypedValue.Sprint()
	res := m.Eval(gno.Call(gno.Sel(&gno.ConstExpr{TypedValue: tv}, "Error")))
	return res[0].GetString(), true
}
