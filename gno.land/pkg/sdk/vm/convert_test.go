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
	cases := []struct {
		name     string
		errorMsg string
		expected string
	}{
		{
			name:     "non-empty error",
			errorMsg: "my error",
			// Simplified format: type is *testdata.myError, value is dereferenced struct, error string included
			expected: `{"results":[{"T":"*testdata.myError","V":{},"error":"my error"}],"@error":"my error"}`,
		},
		{
			name:     "empty error",
			errorMsg: "",
			expected: `{"results":[{"T":"*testdata.myError","V":{},"error":""}],"@error":""}`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := gnolang.NewMachine("testdata", nil)
			defer m.Release()

			code := fmt.Sprintf(`
package testdata
type myError struct { }
func (err *myError) Error() string { return %q }
var Value error = &myError{}`, tc.errorMsg)

			nn := m.MustParseFile("testdata.gno", code)
			m.RunFiles(nn)
			m.RunDeclaration(gnolang.ImportD("testdata", "testdata"))

			tps := m.Eval(gnolang.Sel(gnolang.Nx("testdata"), "Value"))
			require.Len(t, tps, 1)

			tv := tps[0]
			// Use signature-based detection: pass error type as lastReturnType
			// to simulate MsgCall behavior where function signature is known.
			// We use gErrorType (via IsErrorType) to check if signature declares error.
			rep := stringifyJSONResults(m, []gnolang.TypedValue{tv}, tv.T)
			require.Equal(t, tc.expected, rep)
		})
	}
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
	cases := []struct {
		name     string
		code     string
		expected string
	}{
		{
			name: "simple_struct",
			code: `package testdata
type Item struct { ID int; Name string }
var Value = Item{ID: 1, Name: "test"}`,
			expected: `{"results":[{"T":"testdata.Item","V":{"ID":{"T":"int","V":1},"Name":{"T":"string","V":"test"}}}]}`,
		},
		{
			name: "empty_struct",
			code: `package testdata
type Empty struct {}
var Value = Empty{}`,
			expected: `{"results":[{"T":"testdata.Empty","V":{}}]}`,
		},
		{
			name: "nested_struct",
			code: `package testdata
type Inner struct { Value int }
type Outer struct { Inner Inner }
var Value = Outer{Inner: Inner{Value: 42}}`,
			expected: `{"results":[{"T":"testdata.Outer","V":{"Inner":{"T":"testdata.Inner","V":{"Value":{"T":"int","V":42}}}}}]}`,
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
// Slice Tests
// ============================================================================

func TestConvertJSONSlices(t *testing.T) {
	cases := []struct {
		name     string
		code     string
		expected string
	}{
		{
			name:     "int_slice",
			code:     `package testdata; var Value = []int{1, 2, 3}`,
			expected: `{"results":[{"T":"[]int","V":[1,2,3]}]}`,
		},
		{
			name:     "string_slice",
			code:     `package testdata; var Value = []string{"a", "b"}`,
			expected: `{"results":[{"T":"[]string","V":["a","b"]}]}`,
		},
		{
			name:     "empty_slice",
			code:     `package testdata; var Value = []int{}`,
			expected: `{"results":[{"T":"[]int","V":[]}]}`,
		},
		{
			name: "struct_slice",
			code: `package testdata
type Item struct { ID int }
var Value = []Item{{ID: 1}, {ID: 2}}`,
			expected: `{"results":[{"T":"[]testdata.Item","V":[{"T":"testdata.Item","V":{"ID":{"T":"int","V":1}}},{"T":"testdata.Item","V":{"ID":{"T":"int","V":2}}}]}]}`,
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
		require.Equal(t, `{"results":[{"T":"*testdata.Item","V":null}]}`, rep)
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
		require.Contains(t, rep, `"T":"*testdata.Item"`)
		require.Contains(t, rep, `"ID":{"T":"int","V":42}`)
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
	cases := []struct {
		name     string
		code     string
		expected string
	}{
		{
			name:     "declared_int",
			code:     `package testdata; type MyInt int; var Value MyInt = 42`,
			expected: `{"results":[{"T":"testdata.MyInt","V":42,"base":"int"}]}`,
		},
		{
			name:     "declared_string",
			code:     `package testdata; type MyString string; var Value MyString = "hello"`,
			expected: `{"results":[{"T":"testdata.MyString","V":"hello","base":"string"}]}`,
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

		// Should detect cycle with @ref
		require.Contains(t, rep, `"@ref"`)
		require.Contains(t, rep, `"Value":{"T":"int","V":1}`)
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

		// Should have all three values
		require.Contains(t, rep, `"Value":{"T":"int","V":1}`)
		require.Contains(t, rep, `"Value":{"T":"int","V":2}`)
		require.Contains(t, rep, `"Value":{"T":"int","V":3}`)
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

		require.Contains(t, rep, `{"T":"int","V":42}`)
		require.Contains(t, rep, `"T":"testdata.Item"`)
		require.Contains(t, rep, `{"T":"[]int","V":[1,2,3]}`)
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
			expected: `{"results":[{"T":"testdata.Tagged","V":{"first_name":{"T":"string","V":"John"}}}]}`,
		},
		{
			name: "tag_with_omitempty",
			code: "package testdata\ntype WithOmit struct {\n\tName string `json:\"name,omitempty\"`\n}\nvar Value = WithOmit{Name: \"test\"}",
			expected: `{"results":[{"T":"testdata.WithOmit","V":{"name":{"T":"string","V":"test"}}}]}`,
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

		// Should contain first and last elements
		require.Contains(t, rep, `0`)
		require.Contains(t, rep, `49`)
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

		// Should contain the deeply nested value
		require.Contains(t, rep, `"V":{"T":"string","V":"deep"}`)
		require.Contains(t, rep, `"T":"testdata.L1"`)
		require.Contains(t, rep, `"T":"testdata.L5"`)
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
