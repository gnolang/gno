package gnolang

import (
	"fmt"
	"math"
	"testing"

	storetypes "github.com/gnolang/gno/tm2/pkg/store/types"
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
			name: "uint64(1) then uint8(0): Sign should be 0",
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
				t.Errorf("Sign() = %d, want %d", got, tt.wantSign)
			}
		})
	}
}

func TestSignFloat(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(tv *TypedValue)
		wantSign       int
		expectPanicMsg string
	}{
		{
			name: "float32 positive",
			setup: func(tv *TypedValue) {
				tv.T = Float32Type
				tv.SetFloat32(math.Float32bits(1.25))
			},
			wantSign: 1,
		},
		{
			name: "float32 negative",
			setup: func(tv *TypedValue) {
				tv.T = Float32Type
				tv.SetFloat32(math.Float32bits(-1.25))
			},
			wantSign: -1,
		},
		{
			name: "float32 zero",
			setup: func(tv *TypedValue) {
				tv.T = Float32Type
				tv.SetFloat32(math.Float32bits(0))
			},
			wantSign: 0,
		},
		{
			name: "float64 positive",
			setup: func(tv *TypedValue) {
				tv.T = Float64Type
				tv.SetFloat64(math.Float64bits(1.25))
			},
			wantSign: 1,
		},
		{
			name: "float64 negative",
			setup: func(tv *TypedValue) {
				tv.T = Float64Type
				tv.SetFloat64(math.Float64bits(-1.25))
			},
			wantSign: -1,
		},
		{
			name: "float64 zero",
			setup: func(tv *TypedValue) {
				tv.T = Float64Type
				tv.SetFloat64(math.Float64bits(0))
			},
			wantSign: 0,
		},
		{
			name: "float32 +Inf",
			setup: func(tv *TypedValue) {
				tv.T = Float32Type
				tv.SetFloat32(math.Float32bits(float32(math.Inf(1))))
			},
			wantSign: 1,
		},
		{
			name: "float32 -Inf",
			setup: func(tv *TypedValue) {
				tv.T = Float32Type
				tv.SetFloat32(math.Float32bits(float32(math.Inf(-1))))
			},
			wantSign: -1,
		},
		{
			name: "float64 +Inf",
			setup: func(tv *TypedValue) {
				tv.T = Float64Type
				tv.SetFloat64(math.Float64bits(math.Inf(1)))
			},
			wantSign: 1,
		},
		{
			name: "float64 -Inf",
			setup: func(tv *TypedValue) {
				tv.T = Float64Type
				tv.SetFloat64(math.Float64bits(math.Inf(-1)))
			},
			wantSign: -1,
		},
		{
			name: "float32 NaN",
			setup: func(tv *TypedValue) {
				tv.T = Float32Type
				tv.SetFloat32(math.Float32bits(float32(math.NaN())))
			},
			expectPanicMsg: "sign of NaN is undefined",
		},
		{
			name: "float64 NaN",
			setup: func(tv *TypedValue) {
				tv.T = Float64Type
				tv.SetFloat64(math.Float64bits(math.NaN()))
			},
			expectPanicMsg: "sign of NaN is undefined",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tv TypedValue
			tt.setup(&tv)

			if tt.expectPanicMsg != "" {
				assert.PanicsWithValue(t, tt.expectPanicMsg, func() { tv.Sign() })
				return
			}

			got := tv.Sign()
			if got != tt.wantSign {
				t.Errorf("Sign() = %d, want %d", got, tt.wantSign)
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
			store := NewStore(nil, nil, nil)
			m := NewMachine("main", store)
			x := m.MustParseExpr(tc.valX)
			vals := m.Eval(x)
			require.Len(t, vals, 1)
			mk, isNaN := vals[0].ComputeMapKey(store, false)
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
			store := NewStore(nil, nil, nil)
			m := NewMachine("main", store)
			v1 := m.Eval(m.MustParseExpr(pair[0]))
			v2 := m.Eval(m.MustParseExpr(pair[1]))
			require.Len(t, v1, 1)
			require.Len(t, v2, 1)
			mk1, nan1 := v1[0].ComputeMapKey(store, false)
			mk2, nan2 := v2[0].ComputeMapKey(store, false)
			require.False(t, nan1)
			require.False(t, nan2)
			assert.NotEqual(t, mk1, mk2)
		})
	}
}

// TestComputeMapKey_GasVsDepthLimit verifies Maxwell's claim that gas metering
// alone is the primary defense against the deep-nesting attack.
//
// A [1]any{...} chain N levels deep costs N*OpCPUComputeMapKey gas units when
// used as a map key. The Go stack overflows at roughly 1_900_000 levels. With
// MaxGas=3_000_000_000 and OpCPUComputeMapKey=10, the attacker can reach
// 3_000_000_000/10 = 300_000_000 levels before running out of gas — far beyond
// the ~1_900_000 crash point. Gas alone is therefore NOT sufficient to prevent
// the attack; the depth guard (maxComputeMapKeyDepth) is required.
func TestComputeMapKey_GasVsDepthLimit(t *testing.T) {
	const (
		maxGas          int64 = 3_000_000_000
		stackCrashDepth int   = 1_900_000
	)

	// Depth at which gas is exhausted at current rate.
	gasLimitDepth := int(maxGas / OpCPUComputeMapKey)

	// Gas alone does NOT protect: attacker can go deeper than stack crash
	// before the gas meter fires.
	require.Greater(t, gasLimitDepth, stackCrashDepth,
		"gas metering (at OpCPUComputeMapKey=%d, MaxGas=%d) would only stop "+
			"the attack at depth %d, but stack overflow occurs at ~%d — "+
			"the depth guard is still required",
		OpCPUComputeMapKey, maxGas, gasLimitDepth, stackCrashDepth)

	// But with the depth guard, the attack is stopped at maxComputeMapKeyDepth,
	// which is well below the stack crash point.
	require.Less(t, maxComputeMapKeyDepth, stackCrashDepth,
		"maxComputeMapKeyDepth=%d should be below stack crash depth=%d",
		maxComputeMapKeyDepth, stackCrashDepth)

	// Confirm empirically: a depth > maxComputeMapKeyDepth panics before
	// running out of gas.
	alloc := NewAllocator(1 << 62)
	gm := storetypes.NewGasMeter(maxGas)
	store := NewStore(alloc, nil, nil)
	txStore := store.BeginTransaction(nil, nil, gm)

	// Build a chain one level past the depth guard.
	var a TypedValue
	a.T = IntType
	a.SetInt(0)
	for i := 0; i < maxComputeMapKeyDepth+1; i++ {
		av := &ArrayValue{List: []TypedValue{a}}
		a = TypedValue{
			T: &ArrayType{Len: 1, Elt: &InterfaceType{}},
			V: av,
		}
	}

	assert.Panics(t, func() {
		a.ComputeMapKey(txStore, false)
	}, "expected panic from depth guard, not stack overflow or OOG")

	// Confirm the panic was the depth guard, not OOG.
	assert.False(t, gm.IsOutOfGas(),
		"gas should not be exhausted when depth guard fires: consumed=%d limit=%d",
		gm.GasConsumed(), maxGas)
}
