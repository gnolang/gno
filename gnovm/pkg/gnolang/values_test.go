package gnolang

import (
	"fmt"
	"testing"
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

func TestSignStaleUpperBytes(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(tv *TypedValue) // first assignment
		apply    func(tv *TypedValue) // second assignment
		wantSign int
	}{
		{
			name: "int64(-1) then int8(1): Sign should be +1",
			setup: func(tv *TypedValue) {
				tv.T = Int64Type
				tv.SetInt64(-1) // fills all 8 bytes with 0xFF
			},
			apply: func(tv *TypedValue) {
				tv.T = Int8Type
				tv.SetInt8(1) // only writes N[0]
			},
			wantSign: 1,
		},
		{
			name: "int64(-1) then int32(1): Sign should be +1",
			setup: func(tv *TypedValue) {
				tv.T = Int64Type
				tv.SetInt64(-1)
			},
			apply: func(tv *TypedValue) {
				tv.T = Int32Type
				tv.SetInt32(1)
			},
			wantSign: 1,
		},
		{
			name: "uint64(0xFFFFFFFFFFFFFFFF) then uint8(0): Sign should be 0",
			setup: func(tv *TypedValue) {
				tv.T = Uint64Type
				tv.SetUint64(1)
			},
			apply: func(tv *TypedValue) {
				tv.T = Uint8Type
				tv.SetUint8(0)
			},
			wantSign: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tv TypedValue
			tt.setup(&tv)
			tt.apply(&tv)

			got := tv.Sign()
			if got != tt.wantSign {
				t.Errorf("Sign() = %d, want %d (stale upper bytes in N buffer affected result)", got, tt.wantSign)
			}
		})
	}
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
