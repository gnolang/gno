package vm

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
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
		// (ephemeral objects are expanded inline, not shown as RefValue)
		// The @error field at top level is extracted
		require.Contains(t, rep, `"@type":"/gno.PointerValue"`)
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
		// (ephemeral objects are expanded inline, not shown as RefValue)
		require.Contains(t, rep, `"@type":"/gno.PointerValue"`)
		require.Contains(t, rep, `"ObjectInfo"`)
		require.Contains(t, rep, `"@error":""`)
	})
}

// ============================================================================
// Primitive Value Tests
// ============================================================================

func TestConvertJSONPrimitives(t *testing.T) {
	cases := []struct {
		name     string
		code     string
		expected string
	}{
		{
			name:     "int_value",
			code:     `package testdata; var Value = 42`,
			expected: `{"results":[{"T":"int","V":{"@type":"/google.protobuf.Int64Value","value":"42"}}]}`,
		},
		{
			name:     "string_value",
			code:     `package testdata; var Value = "hello"`,
			expected: `{"results":[{"T":"string","V":{"@type":"/google.protobuf.StringValue","value":"hello"}}]}`,
		},
		{
			name:     "bool_value",
			code:     `package testdata; var Value = true`,
			expected: `{"results":[{"T":"bool","V":{"@type":"/google.protobuf.BoolValue","value":true}}]}`,
		},
		{
			name:     "float_value",
			code:     `package testdata; var Value = 3.14`,
			expected: `{"results":[{"T":"float64","V":{"@type":"/google.protobuf.StringValue","value":"3.14"}}]}`,
		},
		{
			name:     "zero_int",
			code:     `package testdata; var Value = 0`,
			expected: `{"results":[{"T":"int","V":{"@type":"/google.protobuf.Int64Value","value":"0"}}]}`,
		},
		{
			name:     "negative_int",
			code:     `package testdata; var Value = -42`,
			expected: `{"results":[{"T":"int","V":{"@type":"/google.protobuf.Int64Value","value":"-42"}}]}`,
		},
		{
			name:     "empty_string",
			code:     `package testdata; var Value = ""`,
			expected: `{"results":[{"T":"string","V":{"@type":"/google.protobuf.StringValue","value":""}}]}`,
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
			require.Equal(t, tc.expected, rep)
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
		// Ephemeral (unreal) structs are expanded inline showing their content
		require.Contains(t, rep, `"T":"testdata.Item"`)
		require.Contains(t, rep, `"ObjectInfo"`)
		require.Contains(t, rep, `"Fields"`)
		// Field names should be included
		require.Contains(t, rep, `"N":"ID"`)
		require.Contains(t, rep, `"N":"Name"`)
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
		// Ephemeral empty struct is expanded inline
		require.Contains(t, rep, `"T":"testdata.Empty"`)
		require.Contains(t, rep, `"ObjectInfo"`)
		// Empty struct has empty Fields array (no fields to export)
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
		// Ephemeral nested struct is expanded inline
		require.Contains(t, rep, `"T":"testdata.Outer"`)
		require.Contains(t, rep, `"ObjectInfo"`)
		require.Contains(t, rep, `"Fields"`)
		// Field name should be included
		require.Contains(t, rep, `"N":"Inner"`)
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
		// In Amino format, slice shows as SliceValue with RefValue base
		require.Contains(t, rep, `"T":"[]int"`)
		require.Contains(t, rep, `"@type":"/gno.SliceValue"`)
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
		// In Amino format, slice shows as SliceValue with RefValue base
		require.Contains(t, rep, `"T":"[]string"`)
		require.Contains(t, rep, `"@type":"/gno.SliceValue"`)
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
		// Empty slice still shows as SliceValue in Amino format
		require.Contains(t, rep, `"T":"[]int"`)
		require.Contains(t, rep, `"@type":"/gno.SliceValue"`)
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
		// Ephemeral struct slice shows with JSONArrayValue base containing inline structs
		require.Contains(t, rep, `"@type":"/gno.SliceValue"`)
		require.Contains(t, rep, `"@type":"/gno.JSONArrayValue"`)
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
		// In Amino format, nil pointer shows with RefType and null value
		require.Contains(t, rep, `"T":"*RefType{testdata.Item}"`)
		require.Contains(t, rep, `"V":null`)
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
		// Ephemeral pointer shows as PointerValue with StructValue base (expanded inline)
		require.Contains(t, rep, `"@type":"/gno.PointerValue"`)
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

		// Parse to check structure (key order is non-deterministic)
		var result map[string]interface{}
		err := json.Unmarshal([]byte(rep), &result)
		require.NoError(t, err)

		results := result["results"].([]interface{})
		require.Len(t, results, 1)

		firstResult := results[0].(map[string]interface{})
		require.Equal(t, "map[string]int", firstResult["T"])
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
		require.Contains(t, rep, `"map[string]int"`)
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
		// In Amino format, declared int shows as type name with null value (primitive)
		require.Contains(t, rep, `"T":"testdata.MyInt"`)
		require.Contains(t, rep, `"V":null`)
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
		// Declared string shows type name with StringValue in Amino format
		require.Contains(t, rep, `"T":"testdata.MyString"`)
		require.Contains(t, rep, `"@type":"/gno.StringValue"`)
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

		// Self-referential pointer shows as PointerValue with JSONStructValue base
		// The cycle reference (RefValue) is inside a base64-encoded nested value
		require.Contains(t, rep, `"@type":"/gno.PointerValue"`)
		require.Contains(t, rep, `"@type":"/gno.JSONStructValue"`)
		require.Contains(t, rep, `"ID":":1"`) // Ephemeral object ID
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

		// Linked list pointer shows as PointerValue with JSONStructValue base
		// Nested pointers are inside base64-encoded values
		require.Contains(t, rep, `"@type":"/gno.PointerValue"`)
		require.Contains(t, rep, `"@type":"/gno.JSONStructValue"`)
		require.Contains(t, rep, `"ID":":1"`) // First ephemeral object ID
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

		// In Amino format, primitives are wrapped in protobuf type wrappers
		expected := `{"results":[{"T":"int","V":{"@type":"/google.protobuf.Int64Value","value":"42"}},{"T":"string","V":{"@type":"/google.protobuf.StringValue","value":"hello"}}]}`
		require.Equal(t, expected, rep)
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

		// In Amino format, primitive int is wrapped, struct and slice show as RefValue/SliceValue
		require.Contains(t, rep, `"T":"int"`)
		require.Contains(t, rep, `"@type":"/google.protobuf.Int64Value"`)
		require.Contains(t, rep, `"T":"testdata.Item"`)
		require.Contains(t, rep, `"T":"[]int"`)
		require.Contains(t, rep, `"@type":"/gno.SliceValue"`)
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
	// JSON tags in Gno structs - note that Amino serialization doesn't use JSON tags,
	// it always uses field names. The ephemeral structs are expanded inline.
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
		// Ephemeral struct is expanded inline showing its fields
		require.Contains(t, rep, `"T":"testdata.Tagged"`)
		require.Contains(t, rep, `"ObjectInfo"`)
		require.Contains(t, rep, `"Fields"`)
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
		// Ephemeral struct is expanded inline showing its fields
		require.Contains(t, rep, `"T":"testdata.WithOmit"`)
		require.Contains(t, rep, `"ObjectInfo"`)
		require.Contains(t, rep, `"Fields"`)
	})

	t.Run("json_skip_tag", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		// Struct with a json:"-" tagged field that should be skipped
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
		// Struct should only have 1 field (Skipped field is filtered out)
		require.Contains(t, rep, `"T":"testdata.WithSkip"`)
		require.Contains(t, rep, `"ObjectInfo"`)
		require.Contains(t, rep, `"visible"`)
		require.NotContains(t, rep, `"hidden"`)
	})
}

// ============================================================================
// Field Visibility Tests
// ============================================================================

func TestConvertJSONFieldVisibility(t *testing.T) {
	t.Run("unexported_field_included", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		// Struct with an unexported field - now included since ExportUnexported=true
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
		// Both fields should be present (ExportUnexported=true)
		require.Contains(t, rep, `"T":"testdata.MixedVisibility"`)
		require.Contains(t, rep, `"ObjectInfo"`)
		require.Contains(t, rep, `"public"`)
		require.Contains(t, rep, `"private"`)
	})

	t.Run("all_unexported_struct", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		// Struct with only unexported fields - now included since ExportUnexported=true
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
		// Struct should have all fields (ExportUnexported=true)
		require.Contains(t, rep, `"T":"testdata.AllPrivate"`)
		require.Contains(t, rep, `"ObjectInfo"`)
		require.Contains(t, rep, `"a"`)              // privateA value
		require.Contains(t, rep, `"Fields":[{`)      // has fields, not null
		require.NotContains(t, rep, `"Fields":null`) // fields are included
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

		// In Amino format, slice shows as SliceValue with RefValue base
		require.Contains(t, rep, `"@type":"/gno.SliceValue"`)
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

		// Nested structs show as JSONStructValue with proper field names
		require.Contains(t, rep, `"T":"testdata.L1"`)
		require.Contains(t, rep, `"@type":"/gno.JSONStructValue"`)
		// Verify the deep nesting shows the value in Amino format (string wrapped)
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
		// Int is wrapped in Amino format
		require.Contains(t, rep, `"T":"int"`)
		require.Contains(t, rep, `"@type":"/google.protobuf.Int64Value"`)
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
	t.Run("unexported_fields_included_with_option", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		// Struct with unexported fields (like avl.Tree with its 'node' field)
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

		// Export with ExportUnexported=true (qobject behavior)
		opts := gnolang.JSONExporterOptions{ExportUnexported: true}
		jsonBytes, err := opts.ExportObject(m, sv)
		require.NoError(t, err)

		// The 'node' field should be included
		require.Contains(t, string(jsonBytes), `"node"`, "unexported 'node' field should be included")
		require.Contains(t, string(jsonBytes), `"Fields":[{`, "should have non-empty Fields")
	})

	t.Run("unexported_fields_excluded_by_default", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		// Struct with only unexported fields
		code := `package testdata
type AllPrivate struct {
	hidden string
}
var Value = &AllPrivate{hidden: "secret"}`

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

		// Export with default options (ExportUnexported=false)
		jsonBytes, err := gnolang.JSONExportObject(m, sv)
		require.NoError(t, err)

		// The 'hidden' field should NOT be included with default options
		require.NotContains(t, string(jsonBytes), `"hidden"`, "unexported 'hidden' field should be excluded by default")
		require.Contains(t, string(jsonBytes), `"Fields":[]`, "should have empty Fields")
	})
}

// ============================================================================
// Ephemeral Object Incremental ID Tests
// ============================================================================

func TestConvertJSONEphemeralIncrementalIDs(t *testing.T) {
	t.Run("linked_list_incremental_ids", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		// Create a 3-node linked list (non-persisted/ephemeral)
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

		// Verify that ephemeral objects get unique incremental IDs (:1, :2, :3, etc.)
		// NOT all :0 which would indicate the bug
		require.Contains(t, rep, `":1"`, "first ephemeral object should have ID :1")
		require.Contains(t, rep, `":2"`, "second ephemeral object should have ID :2")
		require.Contains(t, rep, `":3"`, "third ephemeral object should have ID :3")

		// Should NOT contain :0 for any ephemeral object (that's the bug)
		// Note: :0 might appear in other contexts, so we check the ObjectInfo.ID specifically
		require.NotContains(t, rep, `"ID":":0"`, "ephemeral objects should not have ID :0")
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

		// Should have at least one incremental ID
		require.Contains(t, rep, `":1"`, "ephemeral object should have ID :1")
		// The cycle reference should also use :1 (same object)
	})

	t.Run("array_incremental_ids", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		// Array containing pointers to ephemeral structs
		code := `package testdata
type Item struct { Value int }
var Value = [3]*Item{&Item{1}, &Item{2}, &Item{3}}`

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(gnolang.ImportD("testdata", "testdata"))

		tvs := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		rep := stringifyJSONResults(m, tvs, nil)

		// Array and its elements should have unique IDs
		require.Contains(t, rep, `":1"`, "array should have incremental ID")
		require.Contains(t, rep, `":2"`, "first element should have incremental ID")
		require.NotContains(t, rep, `"ID":":0"`, "ephemeral objects should not have ID :0")
	})

	t.Run("map_incremental_ids", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		// Map containing pointers to ephemeral structs
		code := `package testdata
type Item struct { Value int }
var Value = map[string]*Item{"a": &Item{1}, "b": &Item{2}}`

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(gnolang.ImportD("testdata", "testdata"))

		tvs := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		rep := stringifyJSONResults(m, tvs, nil)

		// Map and its values should have unique IDs
		require.Contains(t, rep, `":1"`, "map should have incremental ID")
		require.Contains(t, rep, `":2"`, "first value should have incremental ID")
		require.NotContains(t, rep, `"ID":":0"`, "ephemeral objects should not have ID :0")
	})
}

// ============================================================================
// Cycle Detection Tests - Comprehensive
// ============================================================================

func TestConvertJSONCycleDetection(t *testing.T) {
	t.Run("ephemeral_self_cycle_same_id", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		// Self-referential cycle: node points to itself
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

		// The struct should have ID :1
		require.Contains(t, rep, `"ID":":1"`, "ephemeral struct should have ID :1")
		// The self-reference (cycle) should be a RefValue pointing to the same ID :1
		require.Contains(t, rep, `"ObjectID":":1"`, "cycle reference should point to same ID :1")
		// Should NOT have :0 IDs
		require.NotContains(t, rep, `"ID":":0"`, "should not have zero ID")
	})

	t.Run("ephemeral_mutual_recursion_cycle", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		// Mutual recursion: A -> B -> A
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

		// Should have two different objects with IDs :1 and :2
		require.Contains(t, rep, `":1"`, "should have ID :1")
		require.Contains(t, rep, `":2"`, "should have ID :2")
		// The back-reference should use RefValue
		require.Contains(t, rep, `"@type":"/gno.RefValue"`, "cycle should use RefValue")
		// Should NOT have :0 IDs
		require.NotContains(t, rep, `"ID":":0"`, "should not have zero ID")
	})

	t.Run("ephemeral_linked_list_cycle", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		// Linked list cycle: 1 -> 2 -> 3 -> 1
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
	n3.Next = n1  // cycle back to n1
	Value = n1
}`

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(gnolang.ImportD("testdata", "testdata"))

		tvs := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		rep := stringifyJSONResults(m, tvs, nil)

		// Should have IDs for the nodes (either in ID or ObjectID field)
		require.Contains(t, rep, `":1"`, "should have ID :1 for first node")
		require.Contains(t, rep, `":2"`, "should have ID :2 for second node")
		// Cycle is detected via RefValue - n3.Next points back to n1
		require.Contains(t, rep, `"@type":"/gno.RefValue"`, "cycle should use RefValue")
		// Should NOT have :0 IDs for any ephemeral object
		require.NotContains(t, rep, `":0"`, "should not have zero ID")
	})

	t.Run("ephemeral_deep_tree_with_shared_nodes", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		// Tree where multiple parents point to the same child (diamond pattern)
		// This tests that shared references get the same ID
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

		// Should have multiple objects with unique IDs
		require.Contains(t, rep, `":1"`, "should have ID :1")
		require.Contains(t, rep, `":2"`, "should have ID :2")
		require.Contains(t, rep, `":3"`, "should have ID :3")
		// The shared node should be referenced via RefValue the second time
		require.Contains(t, rep, `"@type":"/gno.RefValue"`, "shared reference should use RefValue")
		// Should NOT have :0 IDs
		require.NotContains(t, rep, `":0"`, "should not have zero ID")
	})

	t.Run("ephemeral_no_cycle_distinct_ids", func(t *testing.T) {
		m := gnolang.NewMachine("testdata", nil)
		defer m.Release()

		// Linear chain without cycle - each node should have unique ID
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

		// Each node should have a distinct incremental ID
		// Some nodes may be RefValue (ObjectID) vs StructValue (ID) depending on depth
		require.Contains(t, rep, `":1"`, "first node should have ID :1")
		require.Contains(t, rep, `":2"`, "second node should have ID :2")
		require.Contains(t, rep, `":3"`, "third node should have ID :3")
		// Should NOT have :0 IDs
		require.NotContains(t, rep, `":0"`, "should not have zero ID")
	})
}
