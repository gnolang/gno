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
			mk, isNaN := vals[0].ComputeMapKey(nil, store, false)
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
			mk1, nan1 := v1[0].ComputeMapKey(nil, store, false)
			mk2, nan2 := v2[0].ComputeMapKey(nil, store, false)
			require.False(t, nan1)
			require.False(t, nan2)
			assert.NotEqual(t, mk1, mk2)
		})
	}
}

// TestMapKeyGasContract pins the two-sided ComputeMapKey gas contract
// around map keys, against an explicit reference cost:
//
//	reference = Σ ComputeMapKey(key_i)  — the pure key-hash cost of the
//	            keys, by definition (measured once, directly).
//
//	1. Restore: rebuilding a decoded map's vmap (loadObjectSafe →
//	   fillTypesOfValue) must charge exactly the reference — one
//	   ComputeMapKey per entry, and nothing else. A stray charge
//	   creeping into the fill walk fails here.
//	2. Write: inserting the same keys via GetPointerForKey (nil alloc,
//	   so allocation gas is deliberately out of scope) must also charge
//	   exactly the reference. An extra charge on the write path fails
//	   here.
//
// If a future (deliberate, symmetric) charge is added to these paths,
// update the reference computation; the final write==restore assertion
// is the durable invariant and must never need touching.
//
// Scope notes: keys are in-memory primitives, so neither side touches
// the store (no amino gas can leak in). The VM-level write path above
// GetPointerForKey (GetPointerAtIndex, i.e. `m[k] = v` in gno code) is
// pinned end-to-end by the compute_map_key_restore_gas txtar; the
// nil-meter case (genesis, tools) must not panic and must still
// rebuild.
func TestMapKeyGasContract(t *testing.T) {
	const n = 5
	ds := NewStore(NewAllocator(1<<30), nil, nil)

	// A MapValue as it looks right after amino decode: entries present
	// in List, vmap not yet rebuilt.
	newDecodedMap := func() *MapValue {
		mv := &MapValue{List: &MapList{}}
		for i := range n {
			item := mv.List.Append(nil, typedInt(i))
			item.Value = typedString("v")
		}
		return mv
	}

	// Reference: the pure key-hash cost of the n keys.
	//
	// If you deliberately added a new charge to the restore or write
	// path and the reference assertions below went red: extend THIS
	// block so the reference includes the new charge. Leave the final
	// write==restore symmetry assertion untouched — if that one is red,
	// the two paths diverged, which is a bug.
	gmRef := storetypes.NewGasMeter(1 << 30)
	for i := range n {
		k := typedInt(i)
		_, isNaN := k.ComputeMapKey(gmRef, ds, false)
		require.False(t, isNaN)
	}
	reference := gmRef.GasConsumed()
	require.GreaterOrEqual(t, reference, int64(n*OpCPUComputeMapKey),
		"sanity: the per-call constant must fire for each key")

	// Contract 1: the restore rebuild charges exactly the reference.
	gmRestore := storetypes.NewGasMeter(1 << 30)
	mv := newDecodedMap()
	fillTypesOfValue(gmRestore, ds, mv)
	require.Len(t, mv.vmap, n, "vmap must be rebuilt with one slot per entry")
	require.Equal(t, reference, gmRestore.GasConsumed(),
		"restore must charge one ComputeMapKey per entry and nothing else")

	// Contract 2: the write path charges exactly the reference.
	gmWrite := storetypes.NewGasMeter(1 << 30)
	mvWrite := &MapValue{}
	mvWrite.MakeMap()
	for i := range n {
		ptr := mvWrite.GetPointerForKey(nil, gmWrite, ds, typedInt(i))
		*ptr.TV = typedString("v")
	}
	require.Equal(t, reference, gmWrite.GasConsumed(),
		"write must charge one ComputeMapKey per key and nothing else")

	// The durable invariant: write and restore charge the same, whatever
	// the composition. Redundant while both equal the reference above;
	// load-bearing if a future (deliberate, symmetric) charge is added
	// and the reference pins are updated.
	require.Equal(t, gmRestore.GasConsumed(), gmWrite.GasConsumed(),
		"write/restore symmetry must hold regardless of charge composition")

	// nil meter: must not panic and must still rebuild.
	mv2 := newDecodedMap()
	fillTypesOfValue(nil, ds, mv2)
	require.Len(t, mv2.vmap, n)
}

// cap(Block.Values) is charged by (*Block).GetShallowSize, so the growth
// policy is consensus-visible: it must stay a pure function of the requested
// size, never Go's growslice. Pin the exact capacities.
func TestGrowBlockValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		oldCap   int
		numNames int
		wantCap  int
	}{
		{"empty to zero", 0, 0, 0},
		{"empty to one", 0, 1, 1},
		{"doubling from nil", 0, 3, 4},
		{"doubling to pool cap", 0, 14, 16},
		{"grow within capacity", 32, 20, 32},
		{"exact fit is not grown", 8, 8, 8},
		{"doubling stops at threshold", 0, 512, 512},
		// Past the threshold: fixed increments, not another doubling.
		{"first step past threshold", 0, 513, 768},
		{"one step covers 600", 0, 600, 768},
		{"two steps cover 800", 0, 800, 1024},
		{"stepping from a large cap", 1024, 1100, 1280},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			old := make([]TypedValue, tt.oldCap)
			got := growBlockValues(old, tt.numNames)
			assert.Len(t, got, tt.numNames, "length")
			assert.Equal(t, tt.wantCap, cap(got), "capacity")
		})
	}
}

// Growing must preserve existing slots and zero the new ones.
func TestGrowBlockValuesPreservesContents(t *testing.T) {
	t.Parallel()

	old := make([]TypedValue, 3)
	for i := range old {
		old[i] = TypedValue{T: IntType}
		old[i].SetInt(int64(i + 1))
	}
	got := growBlockValues(old, 10)
	require.Len(t, got, 10)
	for i := range 3 {
		assert.Equal(t, int64(i+1), got[i].GetInt(), "slot %d preserved", i)
	}
	for i := 3; i < 10; i++ {
		assert.Zero(t, got[i], "slot %d zeroed", i)
	}
}
