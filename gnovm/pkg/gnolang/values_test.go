package gnolang

import (
	"fmt"
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
		{
			// With omitType == false
			`[2]interface{}{"hi,wor", int64(1)}`,
			"[2]interface{}:[\rstring:hi,wor,\x0eint64:\x01\x00\x00\x00\x00\x00\x00\x00]",
			false,
		},
		{
			// With omitType == false
			`struct{a interface{}; b interface{}}{"hi,wor", int64(1)}`,
			"struct{main.a interface{};main.b interface{}}:{\rstring:hi,wor,\x0eint64:\x01\x00\x00\x00\x00\x00\x00\x00}",
			false,
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
