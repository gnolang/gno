package gnolang

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Test Helper Functions
// ============================================================================

// runJSONExportSimpleTest is a helper function that evaluates Gno code and
// compares the JSON output against expected value.
func runJSONExportSimpleTest(t *testing.T, code, expected string) {
	t.Helper()

	m := NewMachine("testdata", nil)
	defer m.Release()

	nn := m.MustParseFile("testdata.gno", code)
	m.RunFiles(nn)
	m.RunDeclaration(ImportD("testdata", "testdata"))

	tvs := m.Eval(Sel(Nx("testdata"), "Value"))
	require.Len(t, tvs, 1)

	result, err := JSONExportTypedValuesSimple(m, tvs)
	require.NoError(t, err)

	// Unmarshal both for comparison (handles key ordering differences)
	var got, want interface{}
	err = json.Unmarshal(result, &got)
	require.NoError(t, err, "failed to unmarshal result: %s", string(result))
	err = json.Unmarshal([]byte(expected), &want)
	require.NoError(t, err, "failed to unmarshal expected: %s", expected)

	require.Equal(t, want, got, "JSON mismatch.\nGot: %s\nWant: %s", string(result), expected)
}

// ============================================================================
// Primitive Tests
// ============================================================================

func TestJSONExportSimplePrimitives(t *testing.T) {
	cases := []struct {
		name     string
		valueRep string
		expected string
	}{
		// Boolean
		{"bool_true", "true", `[{"T":"bool","V":true}]`},
		{"bool_false", "false", `[{"T":"bool","V":false}]`},

		// int types - basic
		{"int_positive", "int(42)", `[{"T":"int","V":42}]`},
		{"int_zero", "int(0)", `[{"T":"int","V":0}]`},
		{"int_negative", "int(-42)", `[{"T":"int","V":-42}]`},
		{"int8_basic", "int8(42)", `[{"T":"int8","V":42}]`},
		{"int16_basic", "int16(42)", `[{"T":"int16","V":42}]`},
		{"int32_basic", "int32(42)", `[{"T":"int32","V":42}]`},
		{"int64_basic", "int64(42)", `[{"T":"int64","V":42}]`},

		// int edge cases
		{"int8_min", "int8(-128)", `[{"T":"int8","V":-128}]`},
		{"int8_max", "int8(127)", `[{"T":"int8","V":127}]`},
		{"int16_min", "int16(-32768)", `[{"T":"int16","V":-32768}]`},
		{"int16_max", "int16(32767)", `[{"T":"int16","V":32767}]`},
		{"int32_min", "int32(-2147483648)", `[{"T":"int32","V":-2147483648}]`},
		{"int32_max", "int32(2147483647)", `[{"T":"int32","V":2147483647}]`},
		{"int64_min", "int64(-9223372036854775808)", `[{"T":"int64","V":-9223372036854775808}]`},
		{"int64_max", "int64(9223372036854775807)", `[{"T":"int64","V":9223372036854775807}]`},

		// uint types - basic
		{"uint_positive", "uint(42)", `[{"T":"uint","V":42}]`},
		{"uint_zero", "uint(0)", `[{"T":"uint","V":0}]`},
		{"uint8_basic", "uint8(42)", `[{"T":"uint8","V":42}]`},
		{"uint16_basic", "uint16(42)", `[{"T":"uint16","V":42}]`},
		{"uint32_basic", "uint32(42)", `[{"T":"uint32","V":42}]`},
		{"uint64_basic", "uint64(42)", `[{"T":"uint64","V":42}]`},

		// uint edge cases
		{"uint8_max", "uint8(255)", `[{"T":"uint8","V":255}]`},
		{"uint16_max", "uint16(65535)", `[{"T":"uint16","V":65535}]`},
		{"uint32_max", "uint32(4294967295)", `[{"T":"uint32","V":4294967295}]`},
		{"uint64_max", "uint64(18446744073709551615)", `[{"T":"uint64","V":18446744073709551615}]`},

		// Float types
		{"float32_basic", "float32(3.14)", `[{"T":"float32","V":3.14}]`},
		{"float64_basic", "float64(3.14)", `[{"T":"float64","V":3.14}]`},
		{"float64_zero", "float64(0)", `[{"T":"float64","V":0}]`},
		{"float64_negative", "float64(-3.14)", `[{"T":"float64","V":-3.14}]`},
		{"float64_small", "float64(0.000001)", `[{"T":"float64","V":0.000001}]`},
		{"float64_large", "float64(1e10)", `[{"T":"float64","V":10000000000}]`},

		// String type
		{"string_basic", `"hello world"`, `[{"T":"string","V":"hello world"}]`},
		{"string_empty", `""`, `[{"T":"string","V":""}]`},
		{"string_unicode", `"日本語"`, `[{"T":"string","V":"日本語"}]`},
		{"string_emoji", `"emoji 🎉"`, `[{"T":"string","V":"emoji 🎉"}]`},
		{"string_escape", `"line1\nline2\ttab"`, `[{"T":"string","V":"line1\nline2\ttab"}]`},
		{"string_quotes", `"say \"hello\""`, `[{"T":"string","V":"say \"hello\""}]`},
		{"string_backslash", `"path\\to\\file"`, `[{"T":"string","V":"path\\to\\file"}]`},

		// Rune type (int32)
		{"rune_A", `'A'`, `[{"T":"int32","V":65}]`},
		{"rune_unicode", `'日'`, `[{"T":"int32","V":26085}]`},
		{"rune_newline", `'\n'`, `[{"T":"int32","V":10}]`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code := fmt.Sprintf(`package testdata; var Value = %s`, tc.valueRep)
			runJSONExportSimpleTest(t, code, tc.expected)
		})
	}
}

// ============================================================================
// Declared Type Tests (with base field)
// ============================================================================

func TestJSONExportSimpleDeclaredTypes(t *testing.T) {
	cases := []struct {
		name     string
		code     string
		expected string
	}{
		{
			name:     "declared_int",
			code:     `package testdata; type MyInt int; var Value MyInt = 42`,
			expected: `[{"T":"testdata.MyInt","V":42,"base":"int"}]`,
		},
		{
			name:     "declared_string",
			code:     `package testdata; type MyString string; var Value MyString = "hello"`,
			expected: `[{"T":"testdata.MyString","V":"hello","base":"string"}]`,
		},
		{
			name:     "declared_bool",
			code:     `package testdata; type MyBool bool; var Value MyBool = true`,
			expected: `[{"T":"testdata.MyBool","V":true,"base":"bool"}]`,
		},
		{
			name:     "declared_float64",
			code:     `package testdata; type MyFloat float64; var Value MyFloat = 3.14`,
			expected: `[{"T":"testdata.MyFloat","V":3.14,"base":"float64"}]`,
		},
		{
			name:     "declared_uint8",
			code:     `package testdata; type Byte uint8; var Value Byte = 255`,
			expected: `[{"T":"testdata.Byte","V":255,"base":"uint8"}]`,
		},
		{
			name:     "declared_nested",
			code:     `package testdata; type MyInt int; type MyInt2 MyInt; var Value MyInt2 = 42`,
			expected: `[{"T":"testdata.MyInt2","V":42,"base":"int"}]`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runJSONExportSimpleTest(t, tc.code, tc.expected)
		})
	}
}

// ============================================================================
// Struct Tests
// ============================================================================

func TestJSONExportSimpleStructs(t *testing.T) {
	cases := []struct {
		name     string
		code     string
		expected string
	}{
		{
			name: "empty_struct",
			code: `package testdata
type Empty struct {}
var Value = Empty{}`,
			expected: `[{"T":"testdata.Empty","V":{"objectid":":1"}}]`,
		},
		{
			name: "single_field_struct",
			code: `package testdata
type Single struct { Value int }
var Value = Single{42}`,
			expected: `[{"T":"testdata.Single","V":{"Value":{"T":"int","V":42},"objectid":":1"}}]`,
		},
		{
			name: "multi_field_struct",
			code: `package testdata
type Multi struct {
	Name   string
	Age    int
	Active bool
	Score  float64
}
var Value = Multi{"Alice", 30, true, 95.5}`,
			expected: `[{"T":"testdata.Multi","V":{"Name":{"T":"string","V":"Alice"},"Age":{"T":"int","V":30},"Active":{"T":"bool","V":true},"Score":{"T":"float64","V":95.5},"objectid":":1"}}]`,
		},
		{
			name: "nested_struct_2_level",
			code: `package testdata
type Inner struct { Value int }
type Outer struct { Name string; Inner Inner }
var Value = Outer{"test", Inner{42}}`,
			expected: `[{"T":"testdata.Outer","V":{"Name":{"T":"string","V":"test"},"Inner":{"T":"testdata.Inner","V":{"Value":{"T":"int","V":42},"objectid":":2"}},"objectid":":1"}}]`,
		},
		{
			name: "nested_struct_3_level",
			code: `package testdata
type Level3 struct { Value string }
type Level2 struct { L3 Level3 }
type Level1 struct { L2 Level2 }
var Value = Level1{Level2{Level3{"deep"}}}`,
			expected: `[{"T":"testdata.Level1","V":{"L2":{"T":"testdata.Level2","V":{"L3":{"T":"testdata.Level3","V":{"Value":{"T":"string","V":"deep"},"objectid":":3"}},"objectid":":2"}},"objectid":":1"}}]`,
		},
		{
			name: "struct_with_zero_values",
			code: `package testdata
type ZeroValues struct {
	Int    int
	Str    string
	Bool   bool
	Float  float64
}
var Value = ZeroValues{}`,
			expected: `[{"T":"testdata.ZeroValues","V":{"Int":{"T":"int","V":0},"Str":{"T":"string","V":""},"Bool":{"T":"bool","V":false},"Float":{"T":"float64","V":0},"objectid":":1"}}]`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runJSONExportSimpleTest(t, tc.code, tc.expected)
		})
	}
}

// ============================================================================
// Recursive/Cyclic Structure Tests
// ============================================================================

func TestJSONExportSimpleRecursive(t *testing.T) {
	t.Run("self_referential_cycle", func(t *testing.T) {
		code := `package testdata
type Recursive struct {
	Name string
	Self *Recursive
}
var Value = &Recursive{Name: "root"}
func init() { Value.Self = Value }`

		m := NewMachine("testdata", nil)
		defer m.Release()

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(ImportD("testdata", "testdata"))

		tvs := m.Eval(Sel(Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		result, err := JSONExportTypedValuesSimple(m, tvs)
		require.NoError(t, err)

		// Verify the cycle is detected with @ref and objectid
		resultStr := string(result)
		t.Logf("Self-referential cycle output: %s", resultStr)
		require.Contains(t, resultStr, `"Name":{"T":"string","V":"root"}`)
		require.Contains(t, resultStr, `"@ref"`)
		require.Contains(t, resultStr, `"objectid"`)
	})

	t.Run("mutually_recursive_cycle", func(t *testing.T) {
		code := `package testdata
type A struct { B *B }
type B struct { A *A }
var Value = &A{}
func init() {
	b := &B{}
	Value.B = b
	b.A = Value
}`

		m := NewMachine("testdata", nil)
		defer m.Release()

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(ImportD("testdata", "testdata"))

		tvs := m.Eval(Sel(Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		result, err := JSONExportTypedValuesSimple(m, tvs)
		require.NoError(t, err)

		// Should contain @ref and objectid due to cycle
		resultStr := string(result)
		t.Logf("Mutually recursive cycle output: %s", resultStr)
		require.Contains(t, resultStr, `"@ref"`)
		require.Contains(t, resultStr, `"objectid"`)
	})

	t.Run("linked_list_no_cycle", func(t *testing.T) {
		code := `package testdata
type Node struct {
	Value int
	Next  *Node
}
var Value = &Node{Value: 1, Next: &Node{Value: 2, Next: &Node{Value: 3}}}`

		m := NewMachine("testdata", nil)
		defer m.Release()

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(ImportD("testdata", "testdata"))

		tvs := m.Eval(Sel(Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		result, err := JSONExportTypedValuesSimple(m, tvs)
		require.NoError(t, err)

		// Verify all three nodes are present
		resultStr := string(result)
		require.Contains(t, resultStr, `"Value":{"T":"int","V":1}`)
		require.Contains(t, resultStr, `"Value":{"T":"int","V":2}`)
		require.Contains(t, resultStr, `"Value":{"T":"int","V":3}`)
	})

	t.Run("deep_nesting", func(t *testing.T) {
		// Create a deeply nested structure (10 levels)
		code := `package testdata
type DeepNode struct {
	Depth int
	Next  *DeepNode
}
var Value *DeepNode
func init() {
	Value = &DeepNode{Depth: 1}
	current := Value
	for i := 2; i <= 10; i++ {
		current.Next = &DeepNode{Depth: i}
		current = current.Next
	}
}`

		m := NewMachine("testdata", nil)
		defer m.Release()

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(ImportD("testdata", "testdata"))

		tvs := m.Eval(Sel(Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		result, err := JSONExportTypedValuesSimple(m, tvs)
		require.NoError(t, err)

		// Verify depth 10 is present
		require.Contains(t, string(result), `"Depth":{"T":"int","V":10}`)
	})
}

// ============================================================================
// Slice Tests
// ============================================================================

func TestJSONExportSimpleSlices(t *testing.T) {
	cases := []struct {
		name     string
		code     string
		expected string
	}{
		{
			name:     "empty_int_slice",
			code:     `package testdata; var Value = []int{}`,
			expected: `[{"T":"[]int","V":[]}]`,
		},
		{
			name:     "int_slice",
			code:     `package testdata; var Value = []int{1, 2, 3}`,
			expected: `[{"T":"[]int","V":[1,2,3]}]`,
		},
		{
			name:     "string_slice",
			code:     `package testdata; var Value = []string{"a", "b", "c"}`,
			expected: `[{"T":"[]string","V":["a","b","c"]}]`,
		},
		{
			name:     "bool_slice",
			code:     `package testdata; var Value = []bool{true, false, true}`,
			expected: `[{"T":"[]bool","V":[true,false,true]}]`,
		},
		{
			name:     "float_slice",
			code:     `package testdata; var Value = []float64{1.1, 2.2, 3.3}`,
			expected: `[{"T":"[]float64","V":[1.1,2.2,3.3]}]`,
		},
		{
			name: "struct_slice",
			code: `package testdata
type Item struct { ID int }
var Value = []Item{{ID: 1}, {ID: 2}}`,
			expected: `[{"T":"[]testdata.Item","V":[{"T":"testdata.Item","V":{"ID":{"T":"int","V":1},"objectid":":1"}},{"T":"testdata.Item","V":{"ID":{"T":"int","V":2},"objectid":":2"}}]}]`,
		},
		{
			name:     "nested_int_slice",
			code:     `package testdata; var Value = [][]int{{1, 2}, {3, 4}}`,
			expected: `[{"T":"[][]int","V":[{"T":"[]int","V":[1,2]},{"T":"[]int","V":[3,4]}]}]`,
		},
		{
			name: "pointer_slice",
			code: `package testdata
type Item struct { ID int }
var Value = []*Item{&Item{ID: 1}, &Item{ID: 2}}`,
			expected: `[{"T":"[]*testdata.Item","V":[{"T":"*testdata.Item","V":{"ID":{"T":"int","V":1},"objectid":":1"}},{"T":"*testdata.Item","V":{"ID":{"T":"int","V":2},"objectid":":2"}}]}]`,
		},
		{
			name:     "single_element_slice",
			code:     `package testdata; var Value = []int{42}`,
			expected: `[{"T":"[]int","V":[42]}]`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runJSONExportSimpleTest(t, tc.code, tc.expected)
		})
	}
}

// ============================================================================
// Array Tests
// ============================================================================

func TestJSONExportSimpleArrays(t *testing.T) {
	cases := []struct {
		name     string
		code     string
		expected string
	}{
		{
			name:     "int_array",
			code:     `package testdata; var Value = [3]int{1, 2, 3}`,
			expected: `[{"T":"[3]int","V":[1,2,3]}]`,
		},
		{
			name:     "string_array",
			code:     `package testdata; var Value = [2]string{"hello", "world"}`,
			expected: `[{"T":"[2]string","V":["hello","world"]}]`,
		},
		{
			name:     "single_element_array",
			code:     `package testdata; var Value = [1]int{42}`,
			expected: `[{"T":"[1]int","V":[42]}]`,
		},
		{
			name:     "byte_array",
			code:     `package testdata; var Value = [4]byte{0x41, 0x42, 0x43, 0x44}`,
			expected: `[{"T":"[4]uint8","V":[65,66,67,68]}]`,
		},
		{
			name: "struct_array",
			code: `package testdata
type Item struct { ID int }
var Value = [2]Item{{ID: 1}, {ID: 2}}`,
			expected: `[{"T":"[2]testdata.Item","V":[{"T":"testdata.Item","V":{"ID":{"T":"int","V":1},"objectid":":1"}},{"T":"testdata.Item","V":{"ID":{"T":"int","V":2},"objectid":":2"}}]}]`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runJSONExportSimpleTest(t, tc.code, tc.expected)
		})
	}
}

// ============================================================================
// Map Tests
// ============================================================================

func TestJSONExportSimpleMaps(t *testing.T) {
	t.Run("string_int_map", func(t *testing.T) {
		code := `package testdata; var Value = map[string]int{"a": 1, "b": 2}`

		m := NewMachine("testdata", nil)
		defer m.Release()

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(ImportD("testdata", "testdata"))

		tvs := m.Eval(Sel(Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		result, err := JSONExportTypedValuesSimple(m, tvs)
		require.NoError(t, err)

		// Map key ordering is non-deterministic, so check structure
		var parsed []map[string]interface{}
		err = json.Unmarshal(result, &parsed)
		require.NoError(t, err)
		require.Len(t, parsed, 1)
		require.Equal(t, "map[string]int", parsed[0]["T"])

		v := parsed[0]["V"].(map[string]interface{})
		require.NotNil(t, v["a"])
		require.NotNil(t, v["b"])
	})

	t.Run("empty_map", func(t *testing.T) {
		code := `package testdata; var Value = map[string]int{}`

		m := NewMachine("testdata", nil)
		defer m.Release()

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(ImportD("testdata", "testdata"))

		tvs := m.Eval(Sel(Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		result, err := JSONExportTypedValuesSimple(m, tvs)
		require.NoError(t, err)

		// Empty map should be empty object or null
		require.Contains(t, string(result), `"map[string]int"`)
	})

	t.Run("int_key_map", func(t *testing.T) {
		code := `package testdata; var Value = map[int]string{1: "one", 2: "two"}`

		m := NewMachine("testdata", nil)
		defer m.Release()

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(ImportD("testdata", "testdata"))

		tvs := m.Eval(Sel(Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		result, err := JSONExportTypedValuesSimple(m, tvs)
		require.NoError(t, err)

		// Non-string key maps should be array of pairs
		require.Contains(t, string(result), `"map[int]string"`)
	})

	t.Run("struct_value_map", func(t *testing.T) {
		code := `package testdata
type Item struct { ID int }
var Value = map[string]Item{"item1": {ID: 1}}`

		m := NewMachine("testdata", nil)
		defer m.Release()

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(ImportD("testdata", "testdata"))

		tvs := m.Eval(Sel(Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		result, err := JSONExportTypedValuesSimple(m, tvs)
		require.NoError(t, err)

		require.Contains(t, string(result), `"map[string]testdata.Item"`)
		require.Contains(t, string(result), `"item1"`)
	})
}

// ============================================================================
// Pointer Tests
// ============================================================================

func TestJSONExportSimplePointers(t *testing.T) {
	t.Run("pointer_to_struct", func(t *testing.T) {
		code := `package testdata
type Data struct { Value int }
var Value = &Data{Value: 42}`

		m := NewMachine("testdata", nil)
		defer m.Release()

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(ImportD("testdata", "testdata"))

		tvs := m.Eval(Sel(Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		result, err := JSONExportTypedValuesSimple(m, tvs)
		require.NoError(t, err)

		resultStr := string(result)
		require.Contains(t, resultStr, `"T":"*testdata.Data"`)
		require.Contains(t, resultStr, `"Value":{"T":"int","V":42}`)
		// objectid may or may not be present depending on state
	})

	t.Run("nil_pointer", func(t *testing.T) {
		code := `package testdata
type Data struct { Value int }
var Value *Data = nil`

		runJSONExportSimpleTest(t, code, `[{"T":"*testdata.Data","V":null}]`)
	})

	t.Run("pointer_in_struct", func(t *testing.T) {
		code := `package testdata
type Item struct { ID int }
type Container struct { Data *Item }
var Value = Container{Data: &Item{ID: 42}}`

		m := NewMachine("testdata", nil)
		defer m.Release()

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(ImportD("testdata", "testdata"))

		tvs := m.Eval(Sel(Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		result, err := JSONExportTypedValuesSimple(m, tvs)
		require.NoError(t, err)

		resultStr := string(result)
		require.Contains(t, resultStr, `"T":"testdata.Container"`)
		require.Contains(t, resultStr, `"T":"*testdata.Item"`)
		require.Contains(t, resultStr, `"ID":{"T":"int","V":42}`)
	})

	t.Run("nil_pointer_in_struct", func(t *testing.T) {
		code := `package testdata
type Item struct { ID int }
type Container struct { Data *Item }
var Value = Container{Data: nil}`

		m := NewMachine("testdata", nil)
		defer m.Release()

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(ImportD("testdata", "testdata"))

		tvs := m.Eval(Sel(Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		result, err := JSONExportTypedValuesSimple(m, tvs)
		require.NoError(t, err)

		resultStr := string(result)
		require.Contains(t, resultStr, `"Data":{"T":"*testdata.Item","V":null}`)
	})
}

// ============================================================================
// Error Type Tests
// ============================================================================

func TestJSONExportSimpleErrors(t *testing.T) {
	t.Run("pointer_receiver_error", func(t *testing.T) {
		code := `package testdata
type PtrError struct { msg string }
func (e *PtrError) Error() string { return e.msg }
var Value error = &PtrError{msg: "pointer error"}`

		m := NewMachine("testdata", nil)
		defer m.Release()

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(ImportD("testdata", "testdata"))

		tvs := m.Eval(Sel(Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		result, err := JSONExportTypedValuesSimple(m, tvs)
		require.NoError(t, err)

		resultStr := string(result)
		require.Contains(t, resultStr, `"error":"pointer error"`)
	})

	t.Run("value_receiver_error", func(t *testing.T) {
		code := `package testdata
type ValError struct { msg string }
func (e ValError) Error() string { return e.msg }
var Value error = ValError{msg: "value error"}`

		m := NewMachine("testdata", nil)
		defer m.Release()

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(ImportD("testdata", "testdata"))

		tvs := m.Eval(Sel(Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		result, err := JSONExportTypedValuesSimple(m, tvs)
		require.NoError(t, err)

		resultStr := string(result)
		require.Contains(t, resultStr, `"error":"value error"`)
	})

	t.Run("empty_error_message", func(t *testing.T) {
		code := `package testdata
type EmptyError struct {}
func (e *EmptyError) Error() string { return "" }
var Value error = &EmptyError{}`

		m := NewMachine("testdata", nil)
		defer m.Release()

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(ImportD("testdata", "testdata"))

		tvs := m.Eval(Sel(Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		result, err := JSONExportTypedValuesSimple(m, tvs)
		require.NoError(t, err)

		resultStr := string(result)
		require.Contains(t, resultStr, `"error":""`)
	})
}

// ============================================================================
// JSON Tag Tests
// ============================================================================

func TestJSONExportSimpleJSONTags(t *testing.T) {
	cases := []struct {
		name     string
		code     string
		expected string
	}{
		{
			name: "custom_json_tag",
			code: "package testdata\ntype Tagged struct {\n\tFirstName string `json:\"first_name\"`\n}\nvar Value = Tagged{FirstName: \"John\"}",
			expected: `[{"T":"testdata.Tagged","V":{"first_name":{"T":"string","V":"John"},"objectid":":1"}}]`,
		},
		{
			name: "mixed_tags",
			code: "package testdata\ntype Mixed struct {\n\tCustom string `json:\"custom\"`\n\tNoTag  string\n}\nvar Value = Mixed{Custom: \"a\", NoTag: \"b\"}",
			expected: `[{"T":"testdata.Mixed","V":{"custom":{"T":"string","V":"a"},"NoTag":{"T":"string","V":"b"},"objectid":":1"}}]`,
		},
		{
			name: "tag_with_omitempty",
			code: "package testdata\ntype WithOmit struct {\n\tName string `json:\"name,omitempty\"`\n}\nvar Value = WithOmit{Name: \"test\"}",
			expected: `[{"T":"testdata.WithOmit","V":{"name":{"T":"string","V":"test"},"objectid":":1"}}]`,
		},
		{
			name: "skip_tag",
			code: "package testdata\ntype WithSkip struct {\n\tPublic  string\n\tPrivate string `json:\"-\"`\n}\nvar Value = WithSkip{Public: \"visible\", Private: \"hidden\"}",
			// json:"-" should use the Go field name "Private" as fallback
			expected: `[{"T":"testdata.WithSkip","V":{"Public":{"T":"string","V":"visible"},"Private":{"T":"string","V":"hidden"},"objectid":":1"}}]`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runJSONExportSimpleTest(t, tc.code, tc.expected)
		})
	}
}

// ============================================================================
// Function Tests
// ============================================================================

func TestJSONExportSimpleFunctions(t *testing.T) {
	t.Run("named_function", func(t *testing.T) {
		code := `package testdata
func MyFunc() {}
var Value = MyFunc`

		m := NewMachine("testdata", nil)
		defer m.Release()

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(ImportD("testdata", "testdata"))

		tvs := m.Eval(Sel(Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		result, err := JSONExportTypedValuesSimple(m, tvs)
		require.NoError(t, err)

		resultStr := string(result)
		require.Contains(t, resultStr, `"T":"func"`)
		require.Contains(t, resultStr, `MyFunc`)
	})
}

// ============================================================================
// Interface Tests
// ============================================================================

func TestJSONExportSimpleInterfaces(t *testing.T) {
	t.Run("interface_with_int", func(t *testing.T) {
		code := `package testdata; var Value interface{} = 42`

		m := NewMachine("testdata", nil)
		defer m.Release()

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(ImportD("testdata", "testdata"))

		tvs := m.Eval(Sel(Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		result, err := JSONExportTypedValuesSimple(m, tvs)
		require.NoError(t, err)

		// Interface with int should show the concrete type
		resultStr := string(result)
		require.Contains(t, resultStr, `"V":42`)
	})

	t.Run("interface_with_struct", func(t *testing.T) {
		code := `package testdata
type Item struct { ID int }
var Value interface{} = Item{ID: 42}`

		m := NewMachine("testdata", nil)
		defer m.Release()

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(ImportD("testdata", "testdata"))

		tvs := m.Eval(Sel(Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		result, err := JSONExportTypedValuesSimple(m, tvs)
		require.NoError(t, err)

		resultStr := string(result)
		require.Contains(t, resultStr, `"ID"`)
	})
}

// ============================================================================
// Stress Tests
// ============================================================================

func TestJSONExportSimpleStress(t *testing.T) {
	t.Run("large_slice", func(t *testing.T) {
		// Create a slice with 100 elements
		var elements []string
		for i := 0; i < 100; i++ {
			elements = append(elements, fmt.Sprintf("%d", i))
		}
		code := fmt.Sprintf(`package testdata; var Value = []int{%s}`, strings.Join(elements, ","))

		m := NewMachine("testdata", nil)
		defer m.Release()

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(ImportD("testdata", "testdata"))

		tvs := m.Eval(Sel(Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		result, err := JSONExportTypedValuesSimple(m, tvs)
		require.NoError(t, err)

		// Verify it contains first and last elements
		resultStr := string(result)
		require.Contains(t, resultStr, `0`)
		require.Contains(t, resultStr, `99`)
	})

	t.Run("large_string", func(t *testing.T) {
		largeStr := strings.Repeat("x", 1000)
		code := fmt.Sprintf(`package testdata; var Value = %q`, largeStr)

		m := NewMachine("testdata", nil)
		defer m.Release()

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(ImportD("testdata", "testdata"))

		tvs := m.Eval(Sel(Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		result, err := JSONExportTypedValuesSimple(m, tvs)
		require.NoError(t, err)

		// Verify the large string is present
		resultStr := string(result)
		require.Contains(t, resultStr, largeStr)
	})

	t.Run("many_fields_struct", func(t *testing.T) {
		// Create a struct with 20 fields
		var fields []string
		var values []string
		for i := 0; i < 20; i++ {
			fields = append(fields, fmt.Sprintf("F%d int", i))
			values = append(values, fmt.Sprintf("F%d: %d", i, i))
		}
		code := fmt.Sprintf(`package testdata
type BigStruct struct {
	%s
}
var Value = BigStruct{%s}`, strings.Join(fields, "\n\t"), strings.Join(values, ", "))

		m := NewMachine("testdata", nil)
		defer m.Release()

		nn := m.MustParseFile("testdata.gno", code)
		m.RunFiles(nn)
		m.RunDeclaration(ImportD("testdata", "testdata"))

		tvs := m.Eval(Sel(Nx("testdata"), "Value"))
		require.Len(t, tvs, 1)

		result, err := JSONExportTypedValuesSimple(m, tvs)
		require.NoError(t, err)

		// Verify first and last fields
		resultStr := string(result)
		require.Contains(t, resultStr, `"F0":{"T":"int","V":0}`)
		require.Contains(t, resultStr, `"F19":{"T":"int","V":19}`)
	})
}

// ============================================================================
// Original Legacy Tests (kept for backward compatibility)
// ============================================================================

func TestConvertJSONValuePrimtive(t *testing.T) {
	cases := []struct {
		ValueRep string // Go representation
		Expected string // string representation
	}{
		// Boolean
		{"true", `{"T":"bool","V":true}`},
		{"false", `{"T":"bool","V":false}`},

		// int types
		{"int(42)", `{"T":"int","V":42}`}, // Needs to be quoted for amino
		{"int8(42)", `{"T":"int8","V":42}`},
		{"int16(42)", `{"T":"int16","V":42}`},
		{"int32(42)", `{"T":"int32","V":42}`},
		{"int64(42)", `{"T":"int64","V":42}`},

		// uint types
		{"uint(42)", `{"T":"uint","V":42}`},
		{"uint8(42)", `{"T":"uint8","V":42}`},
		{"uint16(42)", `{"T":"uint16","V":42}`},
		{"uint32(42)", `{"T":"uint32","V":42}`},
		{"uint64(42)", `{"T":"uint64","V":42}`},

		// Float types
		{"float32(3.14)", `{"T":"float32","V":3.14}`},
		{"float64(3.14)", `{"T":"float64","V":3.14}`},

		// String type
		{`"hello world"`, `{"T":"string","V":"hello world"}`},

		// UntypedRuneType
		{`'A'`, `{"T":"int32","V":65}`},

		// DataByteType (assuming DataByte is an alias for uint8)
		{"uint8(42)", `{"T":"uint8","V":42}`},

		// Byte slice - Base is a RefValue reference with ObjectID
		{`[]byte("AB")`, `{"T":"[]uint8","V":{"@type":"/gno.SliceValue","Base":{"@type":"/gno.RefValue","ObjectID":":1","Escaped":true},"Offset":"0","Length":"2","Maxcap":"8"}}`},

		// Byte array - exported as RefValue reference
		{`[2]byte{0x41, 0x42}`, `{"T":"[2]uint8","V":{"@type":"/gno.RefValue","ObjectID":":1","Escaped":true}}`},

		// XXX: BigInt
		// XXX: BigDec
	}

	for _, tc := range cases {
		t.Run(tc.ValueRep, func(t *testing.T) {
			m := NewMachine("testdata", nil)
			defer m.Release()

			nn := m.MustParseFile("testdata.gno",
				fmt.Sprintf(`package testdata; var Value = %s`, tc.ValueRep))
			m.RunFiles(nn)
			m.RunDeclaration(ImportD("testdata", "testdata"))

			tps := m.Eval(Sel(Nx("testdata"), "Value"))
			require.Len(t, tps, 1)

			tv := tps[0]

			rep, err := JSONExportTypedValue(tv, nil)
			require.NoError(t, err)

			require.Equal(t, string(tc.Expected), string(rep))
		})
	}
}

func TestConvertJSONValueStruct(t *testing.T) {
	const StructsFile = `
package testdata

// E struct
type E struct { S string }

func (e *E) String() string { return e.S }
`
	t.Run("null", func(t *testing.T) {
		m := NewMachine("testdata", nil)
		defer m.Release()

		const expected = `{"T":"*RefType{testdata.E}","V":null}`

		nn := m.MustParseFile("struct.gno", StructsFile)
		m.RunFiles(nn)
		nn = m.MustParseFile("testdata.gno", `package testdata; var Value *E = nil`)
		m.RunFiles(nn)
		m.RunDeclaration(ImportD("testdata", "testdata"))

		tps := m.Eval(Sel(Nx("testdata"), "Value"))
		require.Len(t, tps, 1)

		tv := tps[0]
		rep, err := JSONExportTypedValue(tv, nil)
		require.NoError(t, err)

		require.Equal(t, string(expected), string(rep))
	})

	t.Run("struct value", func(t *testing.T) {
		m := NewMachine("testdata", nil)
		defer m.Release()

		const value = "Hello World"
		// Struct values are exported as RefValue references to break cycles
		const expected = `{"T":"testdata.E","V":{"@type":"/gno.RefValue","ObjectID":":1","Escaped":true}}`

		nn := m.MustParseFile("struct.gno", StructsFile)
		m.RunFiles(nn)
		nn = m.MustParseFile("testdata.gno",
			fmt.Sprintf(`package testdata; var Value = E{%q}`, value))
		m.RunFiles(nn)
		m.RunDeclaration(ImportD("testdata", "testdata"))

		tps := m.Eval(Sel(Nx("testdata"), "Value"))
		require.Len(t, tps, 1)

		tv := tps[0]

		rep, err := JSONExportTypedValue(tv, nil)
		require.NoError(t, err)

		require.Equal(t, string(expected), string(rep))
	})

	t.Run("struct pointer", func(t *testing.T) {
		m := NewMachine("testdata", nil)
		defer m.Release()

		const value = "Hello World"
		// Pointer values have their Base as RefValue reference
		const expected = `{"T":"*RefType{testdata.E}","V":{"@type":"/gno.PointerValue","TV":null,"Base":{"@type":"/gno.RefValue","ObjectID":":1","Escaped":true},"Index":"0"}}`

		nn := m.MustParseFile("struct.gno", StructsFile)
		m.RunFiles(nn)
		nn = m.MustParseFile("testdata.gno",
			fmt.Sprintf(`package testdata; var Value = &E{%q}`, value))
		m.RunFiles(nn)
		m.RunDeclaration(ImportD("testdata", "testdata"))

		tps := m.Eval(Sel(Nx("testdata"), "Value"))
		require.Len(t, tps, 1)

		tv := tps[0]
		rep, err := JSONExportTypedValue(tv, nil)
		require.NoError(t, err)

		require.Equal(t, string(expected), string(rep))
	})
}

func TestConvertJSONValueRecusiveStruct(t *testing.T) {
	const RecursiveValueFile = `
package testdata
type Recursive struct {
        MyString string
	Nested *Recursive
}
var RecursiveStruct = &Recursive{ MyString: "Hello World" }
func init() {
	RecursiveStruct.Nested = RecursiveStruct
}
`
	m := NewMachine("testdata", nil)
	defer m.Release()

	nn := m.MustParseFile("testdata.gno", RecursiveValueFile)
	m.RunFiles(nn)
	m.RunDeclaration(ImportD("testdata", "testdata"))

	tps := m.Eval(Sel(Nx("testdata"), "RecursiveStruct"))
	require.Len(t, tps, 1)
	tv := tps[0]

	data, err := JSONExportTypedValue(tv, nil)
	require.NoError(t, err)
	fmt.Println(string(data))
}

func TestConvertJSONValueRecusiveStructWithSeen(t *testing.T) {
	const RecursiveValueFile = `
package testdata
type Recursive struct {
        MyString string
	Nested *Recursive
}
var RecursiveStruct = &Recursive{ MyString: "Hello World" }
func init() {
	RecursiveStruct.Nested = RecursiveStruct
}
`
	m := NewMachine("testdata", nil)
	defer m.Release()

	nn := m.MustParseFile("testdata.gno", RecursiveValueFile)
	m.RunFiles(nn)
	m.RunDeclaration(ImportD("testdata", "testdata"))

	tps := m.Eval(Sel(Nx("testdata"), "RecursiveStruct"))
	require.Len(t, tps, 1)
	tv := tps[0]

	var s struct {
		Refs map[string]json.RawMessage `json:"refs"`
		Val  json.RawMessage            `json:"Value"`
	}

	seen := map[Object]int{}
	data, err := JSONExportTypedValue(tv, seen)
	s.Val = data

	s.Refs = map[string]json.RawMessage{}
	for o := range seen {
		oid := o.GetObjectID()
		if oid.NewTime == 0 {
			continue
		}

		s.Refs[oid.String()] = amino.MustMarshalJSONAny(o)
	}

	data2, err := json.Marshal(s)
	require.NoError(t, err)

	fmt.Println(string(data2))
}
