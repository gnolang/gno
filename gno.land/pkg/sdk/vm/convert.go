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
	"github.com/gnolang/gno/tm2/pkg/amino"
)

func assertNoPlusPrefix(s string) {
	if strings.HasPrefix(s, "+") {
		panic("numbers cannot start with +")
	}
}

// maxFloatArgLen bounds a float argument before apd sees it. apd converts the
// whole mantissa to a big.Int before it reads the exponent, work that grows
// with the square of the digit count while nothing charges gas for it.
const maxFloatArgLen = 1024

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
			value := convertFloat(arg, 32)
			tv.SetFloat32(math.Float32bits(float32(value)))
			return
		case gno.Float64Type:
			value := convertFloat(arg, 64)
			tv.SetFloat64(math.Float64bits(value))
			return
		default:
			panic(fmt.Sprintf("unexpected primitive type %s", bt.String()))
		}
	case *gno.ArrayType:
		if bt.Elt == gno.Uint8Type {
			// Refuse before decoding. DecodeString allocates
			// DecodedLen(len(arg)) up front, and the result is discarded
			// below unless it is exactly bt.Len bytes, so an oversized
			// argument must not get to size that allocation.
			//
			// This is an equality, not an upper bound: the decoder ignores
			// \r and \n, so len(arg) bounds nothing on its own — a payload
			// may carry arbitrarily many of them. Requiring the canonical
			// encoded length also removes that padding as a source of
			// malleability, which this file's header forbids: it otherwise
			// spells one array value in unboundedly many ways.
			//
			// It does not subsume the decoded-length check below. Padding
			// makes EncodedLen constant across three inputs (1, 2 and 3
			// bytes all encode to 4 chars), so an argument of the right
			// encoded length can still decode to bt.Len±2 bytes.
			if want := base64.StdEncoding.EncodedLen(bt.Len); len(arg) != want {
				panic(fmt.Sprintf(
					"array length mismatch: declared [%d]byte, got a %d byte argument, want %d",
					bt.Len, len(arg), want))
			}
			bz, err := base64.StdEncoding.DecodeString(arg)
			if err != nil {
				panic(fmt.Sprintf(
					"error parsing byte array %q: %v",
					arg, err))
			}
			if len(bz) != bt.Len {
				panic(fmt.Sprintf(
					"array length mismatch: declared [%d]byte, got %d bytes",
					bt.Len, len(bz)))
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
	assertNoPlusPrefix(value)
	if len(value) > maxFloatArgLen {
		panic(fmt.Sprintf(
			"error parsing float%d: argument is %d bytes, over the %d byte limit",
			precision, len(value), maxFloatArgLen))
	}
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

// stringifyJSONResults converts TypedValues to JSON format using Amino
// encoding. It first exports values (replacing persisted objects with RefValues
// and breaking ephemeral cycles), then serializes with amino.MarshalJSON.
// ft is the function type (if available) used for signature-based error
// detection. maxExportBytes caps the estimated serialized size of the export
// walk; an oversized result returns gno.ErrExportSizeExceeded before anything
// is marshaled. Only pass <= 0 (unbounded) for trusted input — every query
// path must pass maxQueryExportBytes.
func stringifyJSONResults(m *gno.Machine, tvs []gno.TypedValue, ft *gno.FuncType, maxExportBytes int64) (string, error) {
	jres := jsonResults{Results: []byte("[]")}
	if len(tvs) > 0 {
		// Export values: replace persisted objects with RefValues,
		// break ephemeral cycles with synthetic ":N" RefValues.
		exported, err := gno.ExportValues(tvs, maxExportBytes)
		if err != nil {
			return "", err
		}

		bz, err := amino.MarshalJSON(exported)
		if err != nil {
			return "", fmt.Errorf("unable to marshal results: %w", err)
		}
		jres.Results = bz

		// Error extraction needs a live machine: the error-interface checks are
		// metered against m.GasMeter and .Error() is evaluated on m. Both
		// checks live inside tryGetError so its recover covers them.
		if m != nil {
			if errStr, ok := tryGetError(m, tvs[len(tvs)-1], ft); ok {
				jres.Error = &errStr
			}
		}
	}

	s, err := json.Marshal(jres)
	if err != nil {
		return "", fmt.Errorf("unable to marshal result: %w", err)
	}

	return string(s), nil
}

// tryGetError extracts the @error field from the last result. ft is the called
// function's signature when one is known (nil on the QueryEval path).
func tryGetError(m *gno.Machine, tv gno.TypedValue, ft *gno.FuncType) (errStr string, ok bool) {
	// Everything below runs in a panic-safe context. ANY panic — out-of-gas,
	// buggy .Error() method, typed-nil receiver nil-deref — gracefully
	// degrades: no @error field, but the already-computed Results JSON is
	// preserved.
	//
	// Rationale for catching OOG here (rather than re-panicking): the main
	// expression evaluation has already succeeded by the time tryGetError
	// runs, and its results have already been Amino-marshaled into
	// jsonResults.Results. Re-panicking OOG here would discard that
	// successful payload and return an empty body with an OOG error, which
	// is strictly worse for the caller. Instead we treat the @error field
	// as best-effort and surface whatever results we already have.
	//
	// The recover must be installed before the satisfaction checks below, not
	// just around m.Eval: both run the metered embedding walk, so both can
	// panic OutOfGasError, and a panic escaping from either would discard the
	// payload exactly as described above.
	defer func() {
		if r := recover(); r != nil {
			errStr = ""
			ok = false
		}
	}()

	// With a signature, only a last result whose declared type implements
	// error is extracted; without one (QueryEval) the dynamic-type check
	// below decides alone.
	if ft != nil && len(ft.Results) > 0 {
		if !gno.IsErrorType(m.GasMeter, ft.Results[len(ft.Results)-1].Type) {
			return "", false
		}
	}

	// Check if the dynamic type implements error (metered walk).
	if !tv.ImplError(m.GasMeter) {
		return "", false
	}

	res := m.Eval(gno.Call(gno.Sel(&gno.ConstExpr{TypedValue: tv}, "Error")))
	// The realm's Error() output is caller-controlled and is assembled into
	// the @error field outside the export size guard (which only bounds
	// jres.Results). Bound it like every other diagnostic string in this module
	// (see boundedString): otherwise a realm whose Error() builds a huge string
	// re-introduces the unmetered response amplification the guard removes for
	// the results. See TestQueryEvalJSON_AtErrorBounded.
	return truncate(res[0].GetString()), true
}
