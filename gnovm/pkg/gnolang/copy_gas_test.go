package gnolang

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/store"
)

// cpuGasMeter tallies the CPU gas charged to it apart from the allocation
// gas charged to the same meter.
type cpuGasMeter struct {
	store.GasMeter
	cpu store.Gas
}

func (gm *cpuGasMeter) ConsumeGas(amount store.Gas, descriptor string) {
	if descriptor == "CPUCycles" {
		gm.cpu += amount
	}
	gm.GasMeter.ConsumeGas(amount, descriptor)
}

func intArrayValue(n int) TypedValue {
	return TypedValue{
		T: &ArrayType{Len: n, Elt: IntType},
		V: &ArrayValue{List: make([]TypedValue, n)},
	}
}

func byteArrayValue(n int) TypedValue {
	return TypedValue{
		T: &ArrayType{Len: n, Elt: Uint8Type},
		V: &ArrayValue{Data: make([]byte, n)},
	}
}

// structOfType returns a struct type of n fields of type ft.
func structOfType(n int, ft Type) *StructType {
	fields := make([]FieldType, n)
	for i := range fields {
		fields[i] = FieldType{Name: Name(fmt.Sprintf("F%d", i)), Type: ft}
	}
	return &StructType{Fields: fields}
}

func TestCopyCPUGas(t *testing.T) {
	nested := TypedValue{
		T: &ArrayType{Len: 4, Elt: &ArrayType{Len: 5, Elt: IntType}},
		V: &ArrayValue{List: make([]TypedValue, 4)},
	}
	for i := range 4 {
		nested.V.(*ArrayValue).List[i] = intArrayValue(5)
	}
	inner := intArrayValue(10)
	strct := TypedValue{
		T: &StructType{Fields: []FieldType{{Name: "A", Type: IntType}, {Name: "B", Type: inner.T}}},
		V: &StructValue{Fields: []TypedValue{{T: IntType}, inner}},
	}
	for _, tc := range []struct {
		name string
		tv   TypedValue
		want int64
	}{
		{"[0]int", intArrayValue(0), 0},
		{"[1000]int", intArrayValue(1000), 1000 * OpCPUSlopeCopyElement},
		{"[1000]byte", byteArrayValue(1000), 1000 * OpCPUSlopeCopyPrimitive},
		{"[4][5]int", nested, (4 + 4*5) * OpCPUSlopeCopyElement},
		{"struct{A int; B [10]int}", strct, (2 + 10) * OpCPUSlopeCopyElement},
	} {
		t.Run(tc.name, func(t *testing.T) {
			alloc := NewAllocator(1 << 30)
			gm := &cpuGasMeter{GasMeter: store.NewGasMeter(1 << 40)}
			alloc.SetGasMeter(gm)
			tc.tv.Copy(alloc, nil)
			if want := tc.want * GasFactorCPU; gm.cpu != want {
				t.Errorf("CPU gas = %d, want %d", gm.cpu, want)
			}
		})
	}
}

// preprocess.go's tv.Copy(nil, nil) copies with a nil allocator, which charges
// nothing.
func TestCopyNilAllocator(t *testing.T) {
	cp := intArrayValue(3).Copy(nil, nil)
	if n := len(cp.V.(*ArrayValue).List); n != 3 {
		t.Fatalf("copied %d elements, want 3", n)
	}
}

func TestCopyOutOfGasBeforeAllocation(t *testing.T) {
	st := structOfType(1000, IntType)
	for _, tc := range []struct {
		name string
		tv   TypedValue
		cpu  int64
	}{
		{"[1000]int", intArrayValue(1000), 1000 * OpCPUSlopeCopyElement},
		{"[1000]byte", byteArrayValue(1000), 1000 * OpCPUSlopeCopyPrimitive},
		{"struct of 1000 ints", TypedValue{
			T: st,
			V: &StructValue{Fields: make([]TypedValue, 1000)},
		}, 1000 * OpCPUSlopeCopyElement},
	} {
		t.Run(tc.name, func(t *testing.T) {
			alloc := NewAllocator(1 << 30)
			alloc.SetGasMeter(store.NewGasMeter(tc.cpu*GasFactorCPU - 1))
			defer func() {
				got := recover()
				if err, ok := got.(store.OutOfGasError); !ok || err.Descriptor != "CPUCycles" {
					t.Fatalf("expected CPU out of gas, got %v", got)
				}
				if _, bytes := alloc.Status(); bytes != 0 {
					t.Fatalf("copy allocated %d bytes before the CPU gas check", bytes)
				}
			}()
			tc.tv.Copy(alloc, nil)
		})
	}
}

func TestZeroFillCPUGas(t *testing.T) {
	inner := &ArrayType{Len: 10, Elt: IntType}
	for _, tc := range []struct {
		name string
		t    Type
		want int64
	}{
		{"[1000]int", &ArrayType{Len: 1000, Elt: IntType}, 1000 * OpCPUSlopeCopyElement},
		{"[1000]byte", &ArrayType{Len: 1000, Elt: Uint8Type}, 0},
		{"[1000]any", &ArrayType{Len: 1000, Elt: &InterfaceType{}}, 0},
		{"[4][5]int", &ArrayType{Len: 4, Elt: &ArrayType{Len: 5, Elt: IntType}}, (4 + 4*5) * OpCPUSlopeCopyElement},
		{"struct{A int; B [10]int}", &StructType{Fields: []FieldType{{Name: "A", Type: IntType}, {Name: "B", Type: inner}}}, (2 + 10) * OpCPUSlopeCopyElement},
	} {
		t.Run(tc.name, func(t *testing.T) {
			alloc := NewAllocator(1 << 30)
			gm := &cpuGasMeter{GasMeter: store.NewGasMeter(1 << 40)}
			alloc.SetGasMeter(gm)
			defaultTypedValue(alloc, tc.t)
			if want := tc.want * GasFactorCPU; gm.cpu != want {
				t.Errorf("CPU gas = %d, want %d", gm.cpu, want)
			}
		})
	}
}

// A zero-fill out of gas stops after allocating its own buffer and before
// writing any element, so none of the nested defaults is allocated.
func TestZeroFillOutOfGasBeforeFill(t *testing.T) {
	inner := &ArrayType{Len: 10, Elt: IntType}
	ref := NewAllocator(1 << 30)
	ref.NewListArray(nil, 1000)
	_, listBytes := ref.Status()
	ref = NewAllocator(1 << 30)
	ref.NewStructFields(1000)
	_, fieldBytes := ref.Status()
	for _, tc := range []struct {
		name  string
		t     Type
		bytes int64
	}{
		{"[1000][10]int", &ArrayType{Len: 1000, Elt: inner}, listBytes},
		{"struct of 1000 [10]int", structOfType(1000, inner), fieldBytes},
	} {
		t.Run(tc.name, func(t *testing.T) {
			alloc := NewAllocator(1 << 30)
			alloc.SetGasMeter(store.NewGasMeter(1000*OpCPUSlopeCopyElement*GasFactorCPU - 1))
			defer func() {
				got := recover()
				if err, ok := got.(store.OutOfGasError); !ok || err.Descriptor != "CPUCycles" {
					t.Fatalf("expected CPU out of gas, got %v", got)
				}
				if _, bytes := alloc.Status(); bytes != tc.bytes {
					t.Fatalf("zero fill allocated %d bytes, want only its own buffer's %d", bytes, tc.bytes)
				}
			}()
			defaultTypedValue(alloc, tc.t)
		})
	}
}

func TestZeroFillTooLargeRefusedBeforeCharge(t *testing.T) {
	st := structOfType(1000, IntType)
	for _, tc := range []struct {
		name string
		t    Type
		gc   bool
	}{
		{"[1000]int", &ArrayType{Len: 1000, Elt: IntType}, false},
		{"[1000]int with GC", &ArrayType{Len: 1000, Elt: IntType}, true},
		{"struct of 1000 ints", st, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			alloc := NewAllocator(1000)
			if tc.gc {
				alloc.SetGCFn(func() (int64, bool) { return 0, false })
			}
			gm := &cpuGasMeter{GasMeter: store.NewGasMeter(1 << 40)}
			alloc.SetGasMeter(gm)
			defer func() {
				got := recover()
				if msg, ok := got.(string); !ok || !strings.HasPrefix(msg, "allocation limit exceeded") {
					t.Fatalf("expected allocation limit exceeded, got %v", got)
				}
				if gm.cpu != 0 {
					t.Fatalf("zero fill charged %d CPU gas before refusing a buffer larger than the cap", gm.cpu)
				}
			}()
			defaultTypedValue(alloc, tc.t)
		})
	}
}
