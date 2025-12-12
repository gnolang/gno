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
		// Use signature-based detection: pass error type as lastReturnType
		rep := stringifyJSONResults(m, []gnolang.TypedValue{tv}, tv.T)
		// In Amino format, error shows as PointerValue with RefValue base
		// The @error field at top level is extracted
		require.Contains(t, rep, `"@type":"/gno.PointerValue"`)
		require.Contains(t, rep, `"@type":"/gno.RefValue"`)
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
		// Use signature-based detection: pass error type as lastReturnType
		rep := stringifyJSONResults(m, []gnolang.TypedValue{tv}, tv.T)
		// In Amino format, error shows as PointerValue with RefValue base
		require.Contains(t, rep, `"@type":"/gno.PointerValue"`)
		require.Contains(t, rep, `"@type":"/gno.RefValue"`)
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
			expected: `{"results":[{"T":"int","V":42}]}`,
		},
		{
			name:     "string_value",
			code:     `package testdata; var Value = "hello"`,
			expected: `{"results":[{"T":"string","V":"hello"}]}`,
		},
		{
			name:     "bool_value",
			code:     `package testdata; var Value = true`,
			expected: `{"results":[{"T":"bool","V":true}]}`,
		},
		{
			name:     "float_value",
			code:     `package testdata; var Value = 3.14`,
			expected: `{"results":[{"T":"float64","V":3.14}]}`,
		},
		{
			name:     "zero_int",
			code:     `package testdata; var Value = 0`,
			expected: `{"results":[{"T":"int","V":0}]}`,
		},
		{
			name:     "negative_int",
			code:     `package testdata; var Value = -42`,
			expected: `{"results":[{"T":"int","V":-42}]}`,
		},
		{
			name:     "empty_string",
			code:     `package testdata; var Value = ""`,
			expected: `{"results":[{"T":"string","V":""}]}`,
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
		// In Amino format, struct shows as RefValue with ObjectID
		require.Contains(t, rep, `"T":"testdata.Item"`)
		require.Contains(t, rep, `"@type":"/gno.RefValue"`)
		require.Contains(t, rep, `"ObjectID"`)
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
		// In Amino format, empty struct shows as RefValue with ObjectID
		require.Contains(t, rep, `"T":"testdata.Empty"`)
		require.Contains(t, rep, `"@type":"/gno.RefValue"`)
		require.Contains(t, rep, `"ObjectID"`)
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
		// In Amino format, nested struct shows as RefValue
		require.Contains(t, rep, `"T":"testdata.Outer"`)
		require.Contains(t, rep, `"@type":"/gno.RefValue"`)
		require.Contains(t, rep, `"ObjectID"`)
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
		// In Amino format, struct slice shows as SliceValue with RefValue base
		require.Contains(t, rep, `"@type":"/gno.SliceValue"`)
		require.Contains(t, rep, `"@type":"/gno.RefValue"`)
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
		// In Amino format, pointer shows as PointerValue with RefValue base
		require.Contains(t, rep, `"@type":"/gno.PointerValue"`)
		require.Contains(t, rep, `"@type":"/gno.RefValue"`)
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
		// In Amino format, declared string shows as StringValue
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

		// In Amino format, self-referential pointer shows as PointerValue with RefValue
		require.Contains(t, rep, `"@type":"/gno.PointerValue"`)
		require.Contains(t, rep, `"@type":"/gno.RefValue"`)
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

		// In Amino format, pointer shows as PointerValue with RefValue base
		require.Contains(t, rep, `"@type":"/gno.PointerValue"`)
		require.Contains(t, rep, `"@type":"/gno.RefValue"`)
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

		expected := `{"results":[{"T":"int","V":42},{"T":"string","V":"hello"}]}`
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

		// In Amino format, primitive int is fine, struct and slice show as RefValue/SliceValue
		require.Contains(t, rep, `{"T":"int","V":42}`)
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
	cases := []struct {
		name     string
		code     string
		expected string
	}{
		{
			name: "custom_tag",
			code: "package testdata\ntype Tagged struct {\n\tFirstName string `json:\"first_name\"`\n}\nvar Value = Tagged{FirstName: \"John\"}",
			// In Amino format, struct shows as RefValue
			expected: `{"results":[{"T":"testdata.Tagged","V":{"@type":"/gno.RefValue","ObjectID":":1","Escaped":true}}]}`,
		},
		{
			name: "tag_with_omitempty",
			code: "package testdata\ntype WithOmit struct {\n\tName string `json:\"name,omitempty\"`\n}\nvar Value = WithOmit{Name: \"test\"}",
			// In Amino format, struct shows as RefValue
			expected: `{"results":[{"T":"testdata.WithOmit","V":{"@type":"/gno.RefValue","ObjectID":":1","Escaped":true}}]}`,
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

		// In Amino format, struct shows as RefValue (since it's non-real but still exported with ref)
		require.Contains(t, rep, `"T":"testdata.L1"`)
		require.Contains(t, rep, `"@type":"/gno.RefValue"`)
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
		// Simulate last return being error type
		rep := stringifyJSONResults(m, tvs, tv2[0].T)

		// Should have @error at top level
		require.Contains(t, rep, `"@error":"test error"`)
		require.Contains(t, rep, `{"T":"int","V":42}`)
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

		// nil error should not produce @error field
		rep := stringifyJSONResults(m, tvs, tvs[0].T)
		require.NotContains(t, rep, `"@error"`)
	})
}
