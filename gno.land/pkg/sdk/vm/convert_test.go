package vm

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/amino"
	stypes "github.com/gnolang/gno/tm2/pkg/store/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertEmptyNumbers(t *testing.T) {
	tests := []struct {
		argT        gnolang.Type
		expectedErr string
	}{
		{gnolang.UintType, `error parsing uint "": strconv.ParseUint: parsing "": invalid syntax`},
		{gnolang.Uint64Type, `error parsing uint64 "": strconv.ParseUint: parsing "": invalid syntax`},
		{gnolang.Uint32Type, `error parsing uint32 "": strconv.ParseUint: parsing "": invalid syntax`},
		{gnolang.Uint16Type, `error parsing uint16 "": strconv.ParseUint: parsing "": invalid syntax`},
		{gnolang.Uint8Type, `error parsing uint8 "": strconv.ParseUint: parsing "": invalid syntax`},
		{gnolang.IntType, `error parsing int "": strconv.ParseInt: parsing "": invalid syntax`},
		{gnolang.Int64Type, `error parsing int64 "": strconv.ParseInt: parsing "": invalid syntax`},
		{gnolang.Int32Type, `error parsing int32 "": strconv.ParseInt: parsing "": invalid syntax`},
		{gnolang.Int16Type, `error parsing int16 "": strconv.ParseInt: parsing "": invalid syntax`},
		{gnolang.Int8Type, `error parsing int8 "": strconv.ParseInt: parsing "": invalid syntax`},
		{gnolang.Float64Type, `error parsing float64 "": parse mantissa: `},
		{gnolang.Float32Type, `error parsing float32 "": parse mantissa: `},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%v", tt.argT)
		t.Run(testname, func(t *testing.T) {
			run := func() {
				_ = convertArgToGno("", tt.argT)
			}
			assert.PanicsWithValue(t, tt.expectedErr, run)
		})
	}
}

func TestConvertByteArrayLengthValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		declaredLen int
		inputLen    int
		shouldPanic bool
	}{
		{"exact match", 32, 32, false},
		{"oversized input", 32, 100, true},
		{"undersized input", 32, 16, true},
		{"empty input", 32, 0, true},
		{"one byte array exact", 1, 1, false},
		{"one byte array oversized", 1, 2, true},
		{"zero length array exact", 0, 0, false},
		{"zero length array oversized", 0, 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			arrType := &gnolang.ArrayType{Len: tt.declaredLen, Elt: gnolang.Uint8Type}
			input := make([]byte, tt.inputLen)
			for i := range input {
				input[i] = byte(i)
			}
			b64 := base64.StdEncoding.EncodeToString(input)

			if tt.shouldPanic {
				require.PanicsWithValue(t, fmt.Sprintf("array length mismatch: declared [%d]byte, got %d bytes", tt.declaredLen, tt.inputLen), func() {
					convertArgToGno(b64, arrType)
				})
			} else {
				tv := convertArgToGno(b64, arrType)
				av, ok := tv.V.(*gnolang.ArrayValue)
				require.True(t, ok)
				assert.Equal(t, tt.declaredLen, av.GetLength())
			}
		})
	}
}
// ============================================================================
// Error Type Tests
// ============================================================================

func TestConvertError(t *testing.T) {
	t.Run("non-empty error", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		code := `
package testdata
type myError struct { }
func (err *myError) Error() string { return "my error" }
var Value error = &myError{}`

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(gnolang.ImportD("testdata", "testdata"))

		tps := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value"))
		require.Len(t, tps, 1)

		tv := tps[0]
		// Create a FuncType with error return type for signature-based detection
		ft := &gnolang.FuncType{Results: []gnolang.FieldType{{Type: tv.T}}}
		rep := stringifyJSONResults(m, []gnolang.TypedValue{tv}, ft)
		// In Amino format, error shows as PointerValue with expanded StructValue base
		require.Contains(t, rep, `"/gno.PointerValue"`)
		require.Contains(t, rep, `"ObjectInfo"`)
		require.Contains(t, rep, `"@error":"my error"`)
	})

	t.Run("empty error", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		code := `
package testdata
type myError struct { }
func (err *myError) Error() string { return "" }
var Value error = &myError{}`

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(gnolang.ImportD("testdata", "testdata"))

		tps := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value"))
		require.Len(t, tps, 1)

		tv := tps[0]
		// Create a FuncType with error return type for signature-based detection
		ft := &gnolang.FuncType{Results: []gnolang.FieldType{{Type: tv.T}}}
		rep := stringifyJSONResults(m, []gnolang.TypedValue{tv}, ft)
		// In Amino format, error shows as PointerValue with expanded StructValue base
		require.Contains(t, rep, `"/gno.PointerValue"`)
		require.Contains(t, rep, `"ObjectInfo"`)
		require.Contains(t, rep, `"@error":""`)
	})
}

func TestConvertErrorPanicking(t *testing.T) {
	t.Run("panicking_error_method", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		code := `
package testdata
type panicError struct {}
func (e *panicError) Error() string { panic("boom") }
var Value error = &panicError{}`

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(gnolang.ImportD("testdata", "testdata"))

		tps := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value"))
		require.Len(t, tps, 1)

		tv := tps[0]
		ft := &gnolang.FuncType{Results: []gnolang.FieldType{{Type: tv.T}}}
		// Should not panic; gracefully omits @error
		rep := stringifyJSONResults(m, []gnolang.TypedValue{tv}, ft)
		require.NotContains(t, rep, `"@error"`)
		// Results should still be present
		require.Contains(t, rep, `"results"`)
	})

	t.Run("out_of_gas_error_method", func(t *testing.T) {
		// Use a machine with a very tight gas limit so .Error() exhausts it.
		m := gnolang.NewMachineWithOptions(gnolang.MachineOptions{
			PkgPath:  "testdata",
			GasMeter: stypes.NewGasMeter(100_000),
		})
		defer m.Release()

		code := `
package testdata
type gasError struct {}
func (e *gasError) Error() string {
	s := ""
	for i := 0; i < 10000; i++ {
		s += "x"
	}
	return s
}
var Value error = &gasError{}`

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(gnolang.ImportD("testdata", "testdata"))

		tps := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value"))
		require.Len(t, tps, 1)

		tv := tps[0]
		ft := &gnolang.FuncType{Results: []gnolang.FieldType{{Type: tv.T}}}

		// stringifyJSONResults should re-panic with OutOfGasError
		require.Panics(t, func() {
			stringifyJSONResults(m, []gnolang.TypedValue{tv}, ft)
		})

		// Verify it's specifically an OOG panic
		func() {
			defer func() {
				r := recover()
				require.NotNil(t, r)
				err, ok := r.(error)
				require.True(t, ok, "panic should be an error")
				var oog stypes.OutOfGasError
				require.True(t, errors.As(err, &oog), "panic should be OutOfGasError, got: %v", err)
			}()
			stringifyJSONResults(m, []gnolang.TypedValue{tv}, ft)
		}()
	})
}

// ============================================================================
// Primitive Value Tests
// ============================================================================

func TestConvertJSONPrimitives(t *testing.T) {
	cases := []struct {
		name      string
		code      string
		checkType string // substring to find in T
		checkVal  string // substring to find in result (value or N)
	}{
		{
			name:      "int_value",
			code:      `package testdata; var Value = 42`,
			checkType: `/gno.PrimitiveType`,
			checkVal:  `"N":`, // int stored in N (base64)
		},
		{
			name:      "string_value",
			code:      `package testdata; var Value = "hello"`,
			checkType: `/gno.PrimitiveType`,
			checkVal:  `"value":"hello"`,
		},
		{
			name:      "bool_value",
			code:      `package testdata; var Value = true`,
			checkType: `/gno.PrimitiveType`,
			checkVal:  `"N":`, // bool stored in N (base64)
		},
		{
			name:      "float_value",
			code:      `package testdata; var Value = 3.14`,
			checkType: `/gno.PrimitiveType`,
			checkVal:  `"N":`, // float stored in N (base64)
		},
		{
			name:      "zero_int",
			code:      `package testdata; var Value = 0`,
			checkType: `/gno.PrimitiveType`,
			checkVal:  `"results"`, // zero int omits N (amino omitempty)
		},
		{
			name:      "negative_int",
			code:      `package testdata; var Value = -42`,
			checkType: `/gno.PrimitiveType`,
			checkVal:  `"N":`, // negative int stored in N (base64)
		},
		{
			name:      "empty_string",
			code:      `package testdata; var Value = ""`,
			checkType: `/gno.PrimitiveType`,
			checkVal:  `"value":""`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := gnolang.NewMachine("testdata", nil)
			defer m.Release()

			nn := m.MustParseFile("testdata.gno", tc.code)
			m.RunFiles(nn)
			m.RunDeclaration(gnolang.ImportD("testdata", "testdata"))

			tvs := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value"))
			require.Len(t, tvs, 1)

			rep := stringifyJSONResults(m, tvs, nil)

			// Should be valid JSON
			var result map[string]json.RawMessage
			require.NoError(t, json.Unmarshal([]byte(rep), &result))
			require.Contains(t, string(result["results"]), tc.checkType)
			require.Contains(t, rep, tc.checkVal)
		})
	}
}

// ============================================================================
// Struct Tests
// ============================================================================

func TestConvertJSONStructs(t *testing.T) {
	t.Run("simple_struct", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		code := `package testdata
type Item struct { ID int; Name string }
var Value = Item{ID: 1, Name: "test"}`

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(gnolang.ImportD("testdata", "testdata"))

		tvs := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		rep := stringifyJSONResults(m, tvs, nil)
		// Type is Amino-encoded RefType with type ID
		require.Contains(t, rep, `"ID":"testdata.Item"`)
		require.Contains(t, rep, `/gno.StructValue`)
		require.Contains(t, rep, `"ObjectInfo"`)
		require.Contains(t, rep, `"Fields"`)
		// Values should be present
		require.Contains(t, rep, `"test"`)
	})

	t.Run("empty_struct", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		code := `package testdata
type Empty struct {}
var Value = Empty{}`

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(gnolang.ImportD("testdata", "testdata"))

		tvs := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		rep := stringifyJSONResults(m, tvs, nil)
		// Type is Amino-encoded RefType
		require.Contains(t, rep, `"ID":"testdata.Empty"`)
		require.Contains(t, rep, `/gno.StructValue`)
		require.Contains(t, rep, `"ObjectInfo"`)
		require.Contains(t, rep, `"Fields":[]`)
	})

	t.Run("nested_struct", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		code := `package testdata
type Inner struct { Value int }
type Outer struct { Inner Inner }
var Value = Outer{Inner: Inner{Value: 42}}`

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(gnolang.ImportD("testdata", "testdata"))

		tvs := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		rep := stringifyJSONResults(m, tvs, nil)
		// Outer type
		require.Contains(t, rep, `"ID":"testdata.Outer"`)
		// Inner type
		require.Contains(t, rep, `"ID":"testdata.Inner"`)
		require.Contains(t, rep, `/gno.StructValue`)
		require.Contains(t, rep, `"Fields"`)
	})
}

// ============================================================================
// Slice Tests
// ============================================================================

func TestConvertJSONSlices(t *testing.T) {
	t.Run("int_slice", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		code := `package testdata; var Value = []int{1, 2, 3}`

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(gnolang.ImportD("testdata", "testdata"))

		tvs := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		rep := stringifyJSONResults(m, tvs, nil)
		// Amino encodes the type as a SliceType object, not a string
		require.Contains(t, rep, `/gno.SliceType`)
		require.Contains(t, rep, `/gno.SliceValue`)
	})

	t.Run("string_slice", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		code := `package testdata; var Value = []string{"a", "b"}`

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(gnolang.ImportD("testdata", "testdata"))

		tvs := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		rep := stringifyJSONResults(m, tvs, nil)
		require.Contains(t, rep, `/gno.SliceType`)
		require.Contains(t, rep, `/gno.SliceValue`)
		require.Contains(t, rep, `"value":"a"`)
		require.Contains(t, rep, `"value":"b"`)
	})

	t.Run("empty_slice", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		code := `package testdata; var Value = []int{}`

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(gnolang.ImportD("testdata", "testdata"))

		tvs := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		rep := stringifyJSONResults(m, tvs, nil)
		require.Contains(t, rep, `/gno.SliceType`)
		require.Contains(t, rep, `/gno.SliceValue`)
		require.Contains(t, rep, `"Length":"0"`)
	})

	t.Run("struct_slice", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		code := `package testdata
type Item struct { ID int }
var Value = []Item{{ID: 1}, {ID: 2}}`

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(gnolang.ImportD("testdata", "testdata"))

		tvs := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		rep := stringifyJSONResults(m, tvs, nil)
		// Amino: SliceValue with ArrayValue base containing StructValues
		require.Contains(t, rep, `/gno.SliceValue`)
		require.Contains(t, rep, `/gno.ArrayValue`)
		require.Contains(t, rep, `"ObjectInfo"`)
		require.Contains(t, rep, `"Length":"2"`)
	})
}

// ============================================================================
// Pointer Tests
// ============================================================================

func TestConvertJSONPointers(t *testing.T) {
	t.Run("nil_pointer", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		code := `package testdata
type Item struct { ID int }
var Value *Item = nil`

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(gnolang.ImportD("testdata", "testdata"))

		tvs := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		rep := stringifyJSONResults(m, tvs, nil)
		// Nil pointer: T is PointerType, no V field (amino omits nil)
		require.Contains(t, rep, `/gno.PointerType`)
		require.Contains(t, rep, `"ID":"testdata.Item"`)
		// V should not be present for nil pointer
		require.NotContains(t, rep, `"V":`)
	})

	t.Run("pointer_to_struct", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		code := `package testdata
type Item struct { ID int }
var Value = &Item{ID: 42}`

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(gnolang.ImportD("testdata", "testdata"))

		tvs := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		rep := stringifyJSONResults(m, tvs, nil)
		// Ephemeral pointer shows as PointerValue with HeapItemValue base
		require.Contains(t, rep, `/gno.PointerValue`)
		require.Contains(t, rep, `"ObjectInfo"`)
	})
}

// ============================================================================
// Map Tests
// ============================================================================

func TestConvertJSONMaps(t *testing.T) {
	t.Run("string_int_map", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		code := `package testdata; var Value = map[string]int{"a": 1, "b": 2}`

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(gnolang.ImportD("testdata", "testdata"))

		tvs := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		rep := stringifyJSONResults(m, tvs, nil)

		// Parse to check structure
		var result map[string]interface{}
		err := json.Unmarshal([]byte(rep), &result)
		require.NoError(t, err)

		results := result["results"].([]interface{})
		require.Len(t, results, 1)

		// Type is an Amino-encoded MapType object
		firstResult := results[0].(map[string]interface{})
		tObj := firstResult["T"].(map[string]interface{})
		require.Equal(t, "/gno.MapType", tObj["@type"])
	})

	t.Run("empty_map", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		code := `package testdata; var Value = map[string]int{}`

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(gnolang.ImportD("testdata", "testdata"))

		tvs := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		rep := stringifyJSONResults(m, tvs, nil)
		require.Contains(t, rep, `/gno.MapType`)
	})
}

// ============================================================================
// Declared Type Tests
// ============================================================================

func TestConvertJSONDeclaredTypes(t *testing.T) {
	t.Run("declared_int", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		code := `package testdata; type MyInt int; var Value MyInt = 42`

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(gnolang.ImportD("testdata", "testdata"))

		tvs := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		rep := stringifyJSONResults(m, tvs, nil)
		// Declared type shows as RefType with type name
		require.Contains(t, rep, `"ID":"testdata.MyInt"`)
		require.Contains(t, rep, `/gno.RefType`)
		// Primitive value stored in N (base64)
		require.Contains(t, rep, `"N":`)
	})

	t.Run("declared_string", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		code := `package testdata; type MyString string; var Value MyString = "hello"`

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(gnolang.ImportD("testdata", "testdata"))

		tvs := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		rep := stringifyJSONResults(m, tvs, nil)
		// Declared string type
		require.Contains(t, rep, `"ID":"testdata.MyString"`)
		require.Contains(t, rep, `/gno.StringValue`)
		require.Contains(t, rep, `"value":"hello"`)
	})
}

// ============================================================================
// Recursive Structure Tests
// ============================================================================

func TestConvertJSONRecursive(t *testing.T) {
	t.Run("self_referential", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		code := `package testdata
type Node struct {
	Value int
	Self  *Node
}
var Value = &Node{Value: 1}
func init() { Value.Self = Value }`

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(gnolang.ImportD("testdata", "testdata"))

		tvs := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		rep := stringifyJSONResults(m, tvs, nil)

		// Self-referential cycle: PointerValue with StructValue, cycle broken by RefValue
		require.Contains(t, rep, `/gno.PointerValue`)
		require.Contains(t, rep, `/gno.StructValue`)
		require.Contains(t, rep, `/gno.ExportRefValue`)
		require.Contains(t, rep, `"ObjectID":":1"`) // Synthetic cycle-breaking ID
	})

	t.Run("linked_list", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		code := `package testdata
type Node struct {
	Value int
	Next  *Node
}
var Value = &Node{Value: 1, Next: &Node{Value: 2, Next: &Node{Value: 3}}}`

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(gnolang.ImportD("testdata", "testdata"))

		tvs := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		rep := stringifyJSONResults(m, tvs, nil)

		// Linear linked list (no cycle): all nodes expanded inline
		require.Contains(t, rep, `/gno.PointerValue`)
		require.Contains(t, rep, `/gno.StructValue`)
		// No cycle, so no RefValue for cycle breaking
		require.NotContains(t, rep, `/gno.ExportRefValue`)
	})
}

// ============================================================================
// Multiple Return Values Tests
// ============================================================================

func TestConvertJSONMultipleValues(t *testing.T) {
	t.Run("two_primitives", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		code := `package testdata
var Value1 = 42
var Value2 = "hello"`

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(gnolang.ImportD("testdata", "testdata"))

		tv1 := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value1"))
		tv2 := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value2"))
		require.Len(t, tv1, 1)
		require.Len(t, tv2, 1)

		tvs := []gnolang.TypedValue{tv1[0], tv2[0]}
		rep := stringifyJSONResults(m, tvs, nil)

		// Should be valid JSON with two results
		var result map[string]json.RawMessage
		require.NoError(t, json.Unmarshal([]byte(rep), &result))

		var results []json.RawMessage
		require.NoError(t, json.Unmarshal(result["results"], &results))
		require.Len(t, results, 2)

		// Both should contain PrimitiveType
		require.Contains(t, rep, `/gno.PrimitiveType`)
		// String value should be present
		require.Contains(t, rep, `"value":"hello"`)
	})

	t.Run("mixed_types", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		code := `package testdata
type Item struct { ID int }
var Value1 = 42
var Value2 = Item{ID: 1}
var Value3 = []int{1, 2, 3}`

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(gnolang.ImportD("testdata", "testdata"))

		tv1 := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value1"))
		tv2 := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value2"))
		tv3 := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value3"))

		tvs := []gnolang.TypedValue{tv1[0], tv2[0], tv3[0]}
		rep := stringifyJSONResults(m, tvs, nil)

		// Should contain all three types
		require.Contains(t, rep, `/gno.PrimitiveType`)
		require.Contains(t, rep, `"ID":"testdata.Item"`)
		require.Contains(t, rep, `/gno.SliceType`)
		require.Contains(t, rep, `/gno.SliceValue`)
	})

	t.Run("empty_results", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		tvs := []gnolang.TypedValue{}
		rep := stringifyJSONResults(m, tvs, nil)

		require.Equal(t, `{"results":[]}`, rep)
	})
}

// ============================================================================
// JSON Tag Tests
// ============================================================================

func TestConvertJSONTags(t *testing.T) {
	// Note: Amino serialization does NOT use Gno-level JSON tags.
	// All fields are serialized regardless of json tags.
	t.Run("custom_tag", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		code := `package testdata
type Tagged struct {
	FirstName string ` + "`json:\"first_name\"`" + `
}
var Value = Tagged{FirstName: "John"}`
		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(gnolang.ImportD("testdata", "testdata"))

		tvs := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		rep := stringifyJSONResults(m, tvs, nil)
		require.Contains(t, rep, `"ID":"testdata.Tagged"`)
		require.Contains(t, rep, `"ObjectInfo"`)
		require.Contains(t, rep, `"Fields"`)
		require.Contains(t, rep, `"John"`)
	})

	t.Run("tag_with_omitempty", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		code := `package testdata
type WithOmit struct {
	Name string ` + "`json:\"name,omitempty\"`" + `
}
var Value = WithOmit{Name: "test"}`
		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(gnolang.ImportD("testdata", "testdata"))

		tvs := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		rep := stringifyJSONResults(m, tvs, nil)
		require.Contains(t, rep, `"ID":"testdata.WithOmit"`)
		require.Contains(t, rep, `"ObjectInfo"`)
		require.Contains(t, rep, `"Fields"`)
		require.Contains(t, rep, `"test"`)
	})

	t.Run("json_skip_tag", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		// Struct with a json:"-" tagged field
		// Note: Amino does NOT respect json:"-" — all fields are included
		code := `package testdata
type WithSkip struct {
	Public string
	Skipped string ` + "`json:\"-\"`" + `
}
var Value = WithSkip{Public: "visible", Skipped: "hidden"}`
		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(gnolang.ImportD("testdata", "testdata"))

		tvs := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		rep := stringifyJSONResults(m, tvs, nil)
		require.Contains(t, rep, `"ID":"testdata.WithSkip"`)
		require.Contains(t, rep, `"ObjectInfo"`)
		require.Contains(t, rep, `"visible"`)
		// Amino includes ALL fields regardless of json tags
		require.Contains(t, rep, `"hidden"`)
	})
}

// ============================================================================
// Field Visibility Tests
// ============================================================================

func TestConvertJSONFieldVisibility(t *testing.T) {
	t.Run("unexported_field_included", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		code := `package testdata
type MixedVisibility struct {
	PublicField string
	privateField string
}
var Value = MixedVisibility{PublicField: "public", privateField: "private"}`
		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(gnolang.ImportD("testdata", "testdata"))

		tvs := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		rep := stringifyJSONResults(m, tvs, nil)
		// Both fields should be present (Amino includes all fields)
		require.Contains(t, rep, `"ID":"testdata.MixedVisibility"`)
		require.Contains(t, rep, `"ObjectInfo"`)
		require.Contains(t, rep, `"public"`)
		require.Contains(t, rep, `"private"`)
	})

	t.Run("all_unexported_struct", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		code := `package testdata
type AllPrivate struct {
	privateA string
	privateB int
}
var Value = AllPrivate{privateA: "a", privateB: 42}`
		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(gnolang.ImportD("testdata", "testdata"))

		tvs := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		rep := stringifyJSONResults(m, tvs, nil)
		// Struct should have all fields (Amino includes unexported fields)
		require.Contains(t, rep, `"ID":"testdata.AllPrivate"`)
		require.Contains(t, rep, `"ObjectInfo"`)
		require.Contains(t, rep, `"a"`)          // privateA value
		require.Contains(t, rep, `"Fields":[{`) // has fields
	})
}

// ============================================================================
// Stress Tests
// ============================================================================

func TestConvertJSONStress(t *testing.T) {
	t.Run("large_slice", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		// Create a slice with 50 elements
		var elements []string
		for i := 0; i < 50; i++ {
			elements = append(elements, fmt.Sprintf("%d", i))
		}
		code := fmt.Sprintf(`package testdata; var Value = []int{%s}`, strings.Join(elements, ","))

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(gnolang.ImportD("testdata", "testdata"))

		tvs := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		rep := stringifyJSONResults(m, tvs, nil)

		require.Contains(t, rep, `/gno.SliceValue`)
		require.Contains(t, rep, `"Length":"50"`)
	})

	t.Run("deeply_nested", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		code := `package testdata
type L5 struct { V string }
type L4 struct { L5 L5 }
type L3 struct { L4 L4 }
type L2 struct { L3 L3 }
type L1 struct { L2 L2 }
var Value = L1{L2{L3{L4{L5{"deep"}}}}}`

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(gnolang.ImportD("testdata", "testdata"))

		tvs := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		rep := stringifyJSONResults(m, tvs, nil)

		// Nested structs show as StructValue
		require.Contains(t, rep, `"ID":"testdata.L1"`)
		require.Contains(t, rep, `/gno.StructValue`)
		// Deep nested string value
		require.Contains(t, rep, `"value":"deep"`)
	})
}

// ============================================================================
// Error at Top Level Tests
// ============================================================================

func TestConvertJSONErrorAtTopLevel(t *testing.T) {
	t.Run("error_with_other_returns", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		code := `package testdata
type MyError struct {}
func (e *MyError) Error() string { return "test error" }
var Value1 = 42
var Value2 error = &MyError{}`

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(gnolang.ImportD("testdata", "testdata"))

		tv1 := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value1"))
		tv2 := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value2"))

		tvs := []gnolang.TypedValue{tv1[0], tv2[0]}
		// Simulate function returning (int, error)
		ft := &gnolang.FuncType{Results: []gnolang.FieldType{{Type: tv1[0].T}, {Type: tv2[0].T}}}
		rep := stringifyJSONResults(m, tvs, ft)

		// Should have @error at top level
		require.Contains(t, rep, `"@error":"test error"`)
		// Int should be present as PrimitiveType
		require.Contains(t, rep, `/gno.PrimitiveType`)
	})

	t.Run("nil_error", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		code := `package testdata
var Value error = nil`

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(gnolang.ImportD("testdata", "testdata"))

		tvs := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		// nil error should not produce @error field (func returns error type)
		ft := &gnolang.FuncType{Results: []gnolang.FieldType{{Type: tvs[0].T}}}
		rep := stringifyJSONResults(m, tvs, ft)
		require.NotContains(t, rep, `"@error"`)
	})
}

// ============================================================================
// ExportObject Tests (qobject path)
// ============================================================================

func TestExportObjectUnexportedFields(t *testing.T) {
	t.Run("export_object_with_amino", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		code := `package testdata
type Tree struct {
	node *Node
}
type Node struct {
	key   string
	value int
}
var Value = &Tree{node: &Node{key: "test", value: 42}}`

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(gnolang.ImportD("testdata", "testdata"))

		tvs := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		// Get the struct value as Object
		pv, ok := tvs[0].V.(gnolang.PointerValue)
		require.True(t, ok, "expected pointer value")
		sv, ok := pv.Base.(*gnolang.HeapItemValue)
		require.True(t, ok, "expected heap item value")

		// Export object and serialize with Amino
		exported := gnolang.ExportObject(sv)
		jsonBytes, err := amino.MarshalJSONAny(exported)
		require.NoError(t, err)

		t.Logf("Export object output: %s", string(jsonBytes))
		// All fields are included with standard Amino encoding
		require.Contains(t, string(jsonBytes), "StructValue")
		require.Contains(t, string(jsonBytes), "Fields")
	})
}

// ============================================================================
// Ephemeral Object Incremental ID Tests
// ============================================================================

func TestConvertJSONEphemeralIncrementalIDs(t *testing.T) {
	t.Run("linked_list_incremental_ids", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		// Create a 3-node linked list (non-persisted/ephemeral, no cycles)
		code := `package testdata
type Node struct {
	Value int
	Next  *Node
}
var Value = &Node{Value: 1, Next: &Node{Value: 2, Next: &Node{Value: 3}}}`

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(gnolang.ImportD("testdata", "testdata"))

		tvs := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		rep := stringifyJSONResults(m, tvs, nil)

		// Linear list (no cycle): all objects expanded inline, no cycle-breaking RefValues.
		// Ephemeral objects without cycles get ":0" ObjectInfo IDs.
		require.Contains(t, rep, `/gno.StructValue`)
		require.Contains(t, rep, `/gno.PointerValue`)
		// No RefValue needed (no cycles)
		require.NotContains(t, rep, `/gno.ExportRefValue`)
	})

	t.Run("self_cycle_incremental_id", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		// Self-referential cycle - only one unique object
		code := `package testdata
type Node struct {
	Value int
	Self  *Node
}
var Value = &Node{Value: 1}
func init() { Value.Self = Value }`

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(gnolang.ImportD("testdata", "testdata"))

		tvs := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		rep := stringifyJSONResults(m, tvs, nil)

		// The cycle reference should use a RefValue with synthetic ID
		require.Contains(t, rep, `/gno.ExportRefValue`, "cycle should use ExportRefValue")
		// Synthetic ObjectID uses zero-address prefix with counter
		require.Contains(t, rep, `"ObjectID":":1"`, "cycle ref should point to :1")
	})

	t.Run("array_no_cycles", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		// Array containing pointers to ephemeral structs (no cycles)
		code := `package testdata
type Item struct { Value int }
var Value = [3]*Item{&Item{1}, &Item{2}, &Item{3}}`

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(gnolang.ImportD("testdata", "testdata"))

		tvs := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		rep := stringifyJSONResults(m, tvs, nil)

		// No cycles, so all objects expanded inline. Should contain StructValue.
		require.Contains(t, rep, "StructValue")
	})

	t.Run("map_no_cycles", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		// Map containing pointers to ephemeral structs (no cycles)
		code := `package testdata
type Item struct { Value int }
var Value = map[string]*Item{"a": &Item{1}, "b": &Item{2}}`

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(gnolang.ImportD("testdata", "testdata"))

		tvs := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		rep := stringifyJSONResults(m, tvs, nil)

		// No cycles, so all objects expanded inline. Should contain MapValue.
		require.Contains(t, rep, "MapValue")
	})
}

// ============================================================================
// Cycle Detection Tests - Comprehensive
// ============================================================================

func TestConvertJSONCycleDetection(t *testing.T) {
	t.Run("ephemeral_self_cycle", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		code := `package testdata
type Node struct {
	Value int
	Self  *Node
}
var Value = &Node{Value: 42}
func init() { Value.Self = Value }`

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(gnolang.ImportD("testdata", "testdata"))

		tvs := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		rep := stringifyJSONResults(m, tvs, nil)

		// Cycle is broken via RefValue with synthetic ObjectID
		require.Contains(t, rep, `/gno.ExportRefValue`, "cycle should use ExportRefValue")
		// Synthetic ObjectID uses zero-address prefix with counter
		require.Contains(t, rep, `"ObjectID":":1"`, "cycle ref should point to :1")
	})

	t.Run("ephemeral_mutual_recursion_cycle", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		code := `package testdata
type NodeA struct {
	Value int
	B     *NodeB
}
type NodeB struct {
	Name string
	A    *NodeA
}
var Value *NodeA
func init() {
	a := &NodeA{Value: 1}
	b := &NodeB{Name: "mutual"}
	a.B = b
	b.A = a
	Value = a
}`

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(gnolang.ImportD("testdata", "testdata"))

		tvs := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		rep := stringifyJSONResults(m, tvs, nil)

		// The back-reference (B.A -> A) should use RefValue
		require.Contains(t, rep, `/gno.ExportRefValue`, "cycle should use ExportRefValue")
		// Both objects should be expanded inline, cycle broken at back-ref
		require.Contains(t, rep, `/gno.StructValue`)
	})

	t.Run("ephemeral_linked_list_cycle", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		code := `package testdata
type Node struct {
	Value int
	Next  *Node
}
var Value *Node
func init() {
	n1 := &Node{Value: 1}
	n2 := &Node{Value: 2}
	n3 := &Node{Value: 3}
	n1.Next = n2
	n2.Next = n3
	n3.Next = n1
	Value = n1
}`

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(gnolang.ImportD("testdata", "testdata"))

		tvs := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		rep := stringifyJSONResults(m, tvs, nil)

		// Cycle: n3.Next -> n1 should be a RefValue
		require.Contains(t, rep, `/gno.ExportRefValue`, "cycle should use ExportRefValue")
		// All three nodes should be expanded as StructValues (except the back-ref)
		require.Contains(t, rep, `/gno.StructValue`)
	})

	t.Run("ephemeral_diamond_shared_node", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		code := `package testdata
type Node struct {
	Value int
	Left  *Node
	Right *Node
}
var Value *Node
func init() {
	shared := &Node{Value: 3}
	left := &Node{Value: 1, Right: shared}
	right := &Node{Value: 2, Left: shared}
	Value = &Node{Value: 0, Left: left, Right: right}
}`

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(gnolang.ImportD("testdata", "testdata"))

		tvs := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		rep := stringifyJSONResults(m, tvs, nil)

		// The shared node is seen twice — second time should be a RefValue
		require.Contains(t, rep, `/gno.ExportRefValue`, "shared reference should use ExportRefValue")
	})

	t.Run("ephemeral_no_cycle_linear", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		code := `package testdata
type Node struct {
	Value int
	Next  *Node
}
var Value = &Node{Value: 1, Next: &Node{Value: 2, Next: &Node{Value: 3, Next: nil}}}`

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(gnolang.ImportD("testdata", "testdata"))

		tvs := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		rep := stringifyJSONResults(m, tvs, nil)

		// No cycles, so no RefValue for cycle breaking — all expanded inline
		require.Contains(t, rep, `/gno.StructValue`)
		require.NotContains(t, rep, `/gno.ExportRefValue`)
	})
}
