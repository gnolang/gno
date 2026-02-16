package gnolang

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockTypedValueStruct struct {
	field int
}

func (m *mockTypedValueStruct) assertValue()          {}
func (m *mockTypedValueStruct) GetShallowSize() int64 { return 0 }
func (m *mockTypedValueStruct) VisitAssociated(vis Visitor) (stop bool) {
	return true
}

func (m *mockTypedValueStruct) String() string {
	return fmt.Sprintf("MockTypedValueStruct(%d)", m.field)
}

func (m *mockTypedValueStruct) DeepFill(store Store) Value {
	return m
}

func TestGetLengthPanic(t *testing.T) {
	tests := []struct {
		name     string
		tv       TypedValue
		expected string
	}{
		{
			name: "NonArrayPointer",
			tv: TypedValue{
				T: &PointerType{Elt: &StructType{}},
				V: PointerValue{
					TV: &TypedValue{
						T: &StructType{},
						V: &mockTypedValueStruct{field: 42},
					},
				},
			},
			expected: "unexpected type for len(): *struct{}",
		},
		{
			name: "UnexpectedType",
			tv: TypedValue{
				T: &StructType{},
				V: &mockTypedValueStruct{field: 42},
			},
			expected: "unexpected type for len(): struct{}",
		},
		{
			name: "UnexpectedPointerType",
			tv: TypedValue{
				T: &PointerType{Elt: &StructType{}},
				V: nil,
			},
			expected: "unexpected type for len(): *struct{}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("the code did not panic")
				} else {
					if r != tt.expected {
						t.Errorf("expected panic message to be %q, got %q", tt.expected, r)
					}
				}
			}()

			tt.tv.GetLength()
		})
	}
}

func TestComputeMapKey(t *testing.T) {
	tt := []struct {
		valX  string
		want  MapKey
		isNaN bool
	}{
		{`int64(1)`, "int64:\x01\x00\x00\x00\x00\x00\x00\x00", false},
		{`int32(255)`, "int32:\xff\x00\x00\x00", false},
		// basic string
		{`"hello"`, "string:hello", false},
		// string that contains bytes which look similar to an encoded int64 key.
		{`"int64:\x01\x00\x00\x00\x00\x00\x00\x00"`, "string:int64:\x01\x00\x00\x00\x00\x00\x00\x00", false},
		// NaN should be reported via isNaN == true and empty key.
		{`func() float64 { p := float64(0); return 0/p }()`, MapKey(""), true},
		{`func() float32 { p := float32(0); return 0/p }()`, MapKey(""), true},
		// float negative zero normalization
		{`float32(-0.0)`, "float32:\x00\x00\x00\x00", false},
		{`float64(-0.0)`, "float64:\x00\x00\x00\x00\x00\x00\x00\x00", false},
		// more examples
		{`uint8(255)`, "uint8:\xff", false},
		{`true`, "bool:\x01", false},
		{`false`, "bool:\x00", false},
		{`nil`, "nil", false},
		{
			`struct{a int; b bool}{1, true}`,
			"struct{main.a int;main.b bool}:{\x08\x01\x00\x00\x00\x00\x00\x00\x00,\x01\x01}",
			false,
		},
		{`[8]byte{'a', 'b'}`, "[8]uint8:[ab\x00\x00\x00\x00\x00\x00]", false},
		{`[1]string{}`, "[1]string:[\x00]", false},
		{`""`, "string:", false},
		{`"\x00"`, "string:\x00", false},
		{
			`struct{a int; b string; c bool}{}`,
			"struct{main.a int;main.b string;main.c bool}:{\x08\x00\x00\x00\x00\x00\x00\x00\x00,\x00,\x01\x00}",
			false,
		},
		{
			`[1][1]int{{42}}`,
			"[1][1]int:[\x0b[\x08*\x00\x00\x00\x00\x00\x00\x00]]",
			false,
		},

		// Regressions from https://github.com/gnolang/gno/issues/4567
		{
			`[2]string{"hi,wor", "ld"}`,
			"[2]string:[\x06hi,wor,\x02ld]",
			false,
		},
		{
			`[2]string{"hi", "wor,ld"}`,
			"[2]string:[\x02hi,\x06wor,ld]",
			false,
		},
		{
			`[2]string{"hi,\x07wor", "ld"}`,
			"[2]string:[\x07hi,\x07wor,\x02ld]",
			false,
		},
		{
			`[2]string{"hi", "wor,\x02ld"}`,
			"[2]string:[\x02hi,\x07wor,\x02ld]",
			false,
		},
		{
			`struct{a string; b string}{"x", "y,z"}`,
			"struct{main.a string;main.b string}:{\x01x,\x03y,z}",
			false,
		},
		{
			`struct{a string; b string}{"x,y", "z"}`,
			"struct{main.a string;main.b string}:{\x03x,y,\x01z}",
			false,
		},

		// Check child types which use omitTypes. (because of interface)
		{
			`[2]interface{}{"hi,wor", int64(1)}`,
			"[2]interface{}:[\rstring:hi,wor,\x0eint64:\x01\x00\x00\x00\x00\x00\x00\x00]",
			false,
		},
		{
			`struct{a interface{}; b interface{}}{"hi,wor", int64(1)}`,
			"struct{main.a interface{};main.b interface{}}:{\rstring:hi,wor,\x0eint64:\x01\x00\x00\x00\x00\x00\x00\x00}",
			false,
		},

		// NaN propagation
		{
			`func() struct{f float64} { p := float64(0); return struct{f float64}{0/p} }()`,
			MapKey(""), true,
		},
		{
			`func() [1]float64 { p := float64(0); return [1]float64{0/p} }()`,
			MapKey(""), true,
		},
	}
	for _, tc := range tt {
		t.Run(tc.valX, func(t *testing.T) {
			m := NewMachine("main", nil)
			x := m.MustParseExpr(tc.valX)
			vals := m.Eval(x)
			require.Len(t, vals, 1)
			mk, isNaN := vals[0].ComputeMapKey(nil, false)
			assert.Equal(t, tc.want, mk)
			assert.Equal(t, tc.isNaN, isNaN)
		})
	}
}

func TestComputeMapKey_collisions(t *testing.T) {
	pairs := [][2]string{
		{`[2]string{"", "abcd"}`, `[2]string{"abcd", ""}`},
		{`[1]interface{}{int8(1)}`, `[1]interface{}{uint8(1)}`},
		{`[1]interface{}{int8(1)}`, `[1]interface{}{true}`},
		{`[2][1]int{{1}, {2}}`, `[2][1]int{{2}, {1}}`},
	}
	for _, pair := range pairs {
		t.Run(pair[0]+" vs "+pair[1], func(t *testing.T) {
			m := NewMachine("main", nil)
			v1 := m.Eval(m.MustParseExpr(pair[0]))
			v2 := m.Eval(m.MustParseExpr(pair[1]))
			require.Len(t, v1, 1)
			require.Len(t, v2, 1)
			mk1, nan1 := v1[0].ComputeMapKey(nil, false)
			mk2, nan2 := v2[0].ComputeMapKey(nil, false)
			require.False(t, nan1)
			require.False(t, nan2)
			assert.NotEqual(t, mk1, mk2)
		})
	}
}

// makeTypedInt creates a TypedValue holding an int for use in test lists.
func makeTypedInt(n int) TypedValue {
	tv := TypedValue{T: IntType}
	tv.SetInt(int64(n))
	return tv
}

// makeTestMap creates a MapValue with n entries for testing ProtectedString.
func makeTestMap(n int) *MapValue {
	mv := &MapValue{}
	mv.MakeMap(n)
	for i := 0; i < n; i++ {
		key := makeTypedInt(i)
		item := mv.List.Append(nilAllocator, key)
		item.Value = makeTypedInt(i * 10)
		mk, _ := key.ComputeMapKey(nil, false)
		mv.vmap[mk] = item
	}
	return mv
}

func makeIntSliceValue(n int) *SliceValue {
	list := make([]TypedValue, n)
	for i := range list {
		list[i] = makeTypedInt(i)
	}
	return &SliceValue{
		Base:   &ArrayValue{List: list},
		Offset: 0,
		Length: n,
		Maxcap: n,
	}
}

func makeIntArrayValue(n int) *ArrayValue {
	list := make([]TypedValue, n)
	for i := range list {
		list[i] = makeTypedInt(i)
	}
	return &ArrayValue{List: list}
}

func TestProtectedStringTruncation(t *testing.T) {
	// --- Slice: non-byte ---
	t.Run("slice 256 not truncated", func(t *testing.T) {
		sv := makeIntSliceValue(256)
		result := sv.String()
		assert.True(t, strings.HasPrefix(result, "slice[(0 int),"))
		assert.False(t, strings.Contains(result, "..."))
	})

	t.Run("slice 257 truncated", func(t *testing.T) {
		sv := makeIntSliceValue(257)
		assert.Equal(t, "slice[...(257 elements)]", sv.String())
	})

	t.Run("slice 1000 truncated", func(t *testing.T) {
		sv := makeIntSliceValue(1000)
		assert.Equal(t, "slice[...(1000 elements)]", sv.String())
	})

	t.Run("slice small renders fully", func(t *testing.T) {
		sv := makeIntSliceValue(3)
		assert.Equal(t, "slice[(0 int),(1 int),(2 int)]", sv.String())
	})

	// --- Slice: byte ---
	t.Run("byte slice 256 not truncated", func(t *testing.T) {
		data := make([]byte, 256)
		sv := &SliceValue{
			Base:   &ArrayValue{Data: data},
			Offset: 0,
			Length: 256,
			Maxcap: 256,
		}
		result := sv.String()
		assert.True(t, strings.HasPrefix(result, "slice[0x"))
		assert.False(t, strings.Contains(result, "..."))
	})

	t.Run("byte slice 257 truncated", func(t *testing.T) {
		data := make([]byte, 257)
		sv := &SliceValue{
			Base:   &ArrayValue{Data: data},
			Offset: 0,
			Length: 257,
			Maxcap: 257,
		}
		result := sv.String()
		assert.Contains(t, result, "...(257)")
	})

	// --- Array: non-byte ---
	t.Run("array 256 not truncated", func(t *testing.T) {
		av := makeIntArrayValue(256)
		result := av.String()
		assert.True(t, strings.HasPrefix(result, "array[(0 int),"))
		assert.False(t, strings.Contains(result, "..."))
	})

	t.Run("array 257 truncated", func(t *testing.T) {
		av := makeIntArrayValue(257)
		assert.Equal(t, "array[...(257 elements)]", av.String())
	})

	t.Run("array 1000 truncated", func(t *testing.T) {
		av := makeIntArrayValue(1000)
		assert.Equal(t, "array[...(1000 elements)]", av.String())
	})

	// --- Array: byte ---
	t.Run("byte array 256 not truncated", func(t *testing.T) {
		av := &ArrayValue{Data: make([]byte, 256)}
		result := av.String()
		assert.True(t, strings.HasPrefix(result, "array[0x"))
		assert.False(t, strings.Contains(result, "..."))
	})

	t.Run("byte array 257 truncated", func(t *testing.T) {
		av := &ArrayValue{Data: make([]byte, 257)}
		result := av.String()
		assert.Contains(t, result, "...")
	})

	// --- Map ---
	t.Run("map 256 not truncated", func(t *testing.T) {
		mv := makeTestMap(256)
		result := mv.String()
		assert.True(t, strings.HasPrefix(result, "map{("))
		assert.False(t, strings.Contains(result, "..."))
	})

	t.Run("map 257 truncated", func(t *testing.T) {
		mv := makeTestMap(257)
		assert.Equal(t, "map{...(257 entries)}", mv.String())
	})

	t.Run("map 1000 truncated", func(t *testing.T) {
		mv := makeTestMap(1000)
		assert.Equal(t, "map{...(1000 entries)}", mv.String())
	})

	t.Run("map small renders fully", func(t *testing.T) {
		mv := makeTestMap(1)
		result := mv.String()
		assert.Equal(t, "map{(0 int):(0 int)}", result)
	})
}
