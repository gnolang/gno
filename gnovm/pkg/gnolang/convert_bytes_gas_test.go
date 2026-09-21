package gnolang

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/store"
)

// Keep input construction outside the measured handler, as in bench_ops_test.go.
func convertSliceInput(n, offset int, backing string, elt Type) TypedValue {
	base := &ArrayValue{}
	if backing == "Data" {
		base.Data = make([]byte, n+offset+1)
		for i := range base.Data {
			base.Data[i] = 'a'
		}
	} else {
		base.List = make([]TypedValue, n+offset+1)
		for i := range base.List {
			base.List[i] = TypedValue{T: elt, N: i2n('a')}
		}
	}
	return TypedValue{T: &SliceType{Elt: elt}, V: &SliceValue{Base: base, Offset: offset, Length: n, Maxcap: n + 1}}
}

// Data is the only backing a byte slice can have now (doOpSliceLit/
// doOpSliceLit2/make/append/Go2GnoValue all produce NewDataArray), but the
// List arm stays covered: ConvertTo still handles it, and the charge must not
// depend on the backing.
func TestConvertBytesStringCPUGas(t *testing.T) {
	namedByte := &DeclaredType{PkgPath: "test", Name: "MyByte", Base: Uint8Type}
	for _, backing := range []string{"Data", "List"} {
		for _, elt := range []Type{Uint8Type, namedByte} {
			for _, offset := range []int{0, 7} {
				t.Run(fmt.Sprintf("%s/%s/offset%d", backing, elt, offset), func(t *testing.T) {
					for _, n := range []int{-1, 0, 13, 1013} {
						m := NewMachineWithOptions(MachineOptions{PkgPath: "test"})
						input := TypedValue{T: &SliceType{Elt: elt}}
						if n >= 0 {
							input = convertSliceInput(n, offset, backing, elt)
						}
						// Only the machine meter is attached: allocator gas is excluded.
						m.GasMeter = store.NewGasMeter(1_000_000)
						m.PushValue(asValue(StringType))
						m.PushValue(input)
						before := m.GasMeter.GasConsumed()
						m.doOpConvert()
						cpu := m.GasMeter.GasConsumed() - before
						length := max(n, 0)
						if got := m.PeekValue(1).GetString(); got != strings.Repeat("a", length) {
							t.Fatalf("N=%d: incorrect result %q", n, got)
						}
						want := (OpCPUConvertStrBytes + OpCPUSlopeConvertBytesStr*int64(length)) * GasFactorCPU
						if cpu != want {
							t.Errorf("N=%d: CPU=%d, want %d", length, cpu, want)
						}
						m.Release()
					}
				})
			}
		}
	}
}

// The rune slope shares doOpConvert's switch with the byte slope; without this
// the []rune→string charge is asserted nowhere in the repo and can be dropped
// by a refactor without failing a single test.
func TestConvertRunesStringCPUGas(t *testing.T) {
	namedRune := &DeclaredType{PkgPath: "test", Name: "MyRune", Base: Int32Type}
	for _, elt := range []Type{Int32Type, namedRune} {
		t.Run(fmt.Sprintf("%s", elt), func(t *testing.T) {
			for _, n := range []int{0, 13, 1013} {
				m := NewMachineWithOptions(MachineOptions{PkgPath: "test"})
				m.GasMeter = store.NewGasMeter(10_000_000)
				m.PushValue(asValue(StringType))
				m.PushValue(convertSliceInput(n, 0, "List", elt))
				before := m.GasMeter.GasConsumed()
				m.doOpConvert()
				cpu := m.GasMeter.GasConsumed() - before
				if got := m.PeekValue(1).GetString(); got != strings.Repeat("a", n) {
					t.Fatalf("N=%d: incorrect result %q", n, got)
				}
				want := (OpCPUConvertStrBytes + OpCPUSlopeConvertRunesStr*int64(n)) * GasFactorCPU
				if cpu != want {
					t.Errorf("N=%d: CPU=%d, want %d", n, cpu, want)
				}
				m.Release()
			}
		})
	}
}

func TestConvertBytesStringOutOfGasBeforeAllocation(t *testing.T) {
	for _, backing := range []string{"Data", "List"} {
		t.Run(backing, func(t *testing.T) {
			m := NewMachineWithOptions(MachineOptions{PkgPath: "test"})
			defer m.Release()
			m.PushValue(asValue(StringType))
			m.PushValue(convertSliceInput(1000, 0, backing, Uint8Type))
			m.GasMeter = store.NewGasMeter((OpCPUConvertStrBytes+OpCPUSlopeConvertBytesStr*1000)*GasFactorCPU - 1)
			before := m.Alloc.bytes
			defer func() {
				got := recover()
				if err, ok := got.(store.OutOfGasError); !ok || err.Descriptor != "CPUCycles" {
					t.Fatalf("expected CPU out of gas, got %v", got)
				}
				if m.Alloc.bytes != before {
					t.Fatal("conversion allocated before CPU gas check")
				}
			}()
			m.doOpConvert()
		})
	}
}

// Byte slice literals must be Data-backed like every other byte-slice
// producer (make, append, []byte(string), Go2GnoValue): a List backing costs
// 40 bytes per element instead of 1 and makes []byte→string an order of
// magnitude slower per byte, which is what forced the conversion slope up.
func TestByteSliceLiteralsAreDataBacked(t *testing.T) {
	namedByte := &DeclaredType{PkgPath: "test", Name: "MyByte", Base: Uint8Type}
	for _, elt := range []Type{Uint8Type, namedByte} {
		t.Run(fmt.Sprintf("%s", elt), func(t *testing.T) {
			const n = 4

			t.Run("doOpSliceLit", func(t *testing.T) {
				m := NewMachineWithOptions(MachineOptions{PkgPath: "test"})
				defer m.Release()
				st := m.Alloc.NewType(&SliceType{Elt: elt})
				elts := make([]KeyValueExpr, n)
				for i := range n {
					elts[i] = KeyValueExpr{Value: &ConstExpr{}}
				}
				m.PushValue(asValue(st))
				for i := range n {
					m.PushValue(TypedValue{T: elt, N: i2n(int64('a' + i))})
				}
				m.PushExpr(&CompositeLitExpr{Elts: elts})
				m.doOpSliceLit()
				assertDataBacked(t, m, n, "abcd")
			})

			t.Run("doOpSliceLit2", func(t *testing.T) {
				m := NewMachineWithOptions(MachineOptions{PkgPath: "test"})
				defer m.Release()
				st := m.Alloc.NewType(&SliceType{Elt: elt})
				m.PushValue(asValue(st))
				// []T{0: 'a', 3: 'd'} -> length 4, middle zero-filled.
				for _, kv := range [][2]int64{{0, 'a'}, {3, 'd'}} {
					m.PushValue(TypedValue{T: IntType, N: i2n(kv[0])})
					m.PushValue(TypedValue{T: elt, N: i2n(kv[1])})
				}
				m.PushExpr(&CompositeLitExpr{Elts: make([]KeyValueExpr, 2)})
				m.doOpSliceLit2()
				assertDataBacked(t, m, n, "a\x00\x00d")
			})
		})
	}
}

// A duplicate index must still panic once Data backing removes the
// IsDefined() signal that used to detect it.
func TestByteSliceLiteralDuplicateIndex(t *testing.T) {
	m := NewMachineWithOptions(MachineOptions{PkgPath: "test"})
	defer m.Release()
	m.PushValue(asValue(m.Alloc.NewType(&SliceType{Elt: Uint8Type})))
	for _, kv := range [][2]int64{{2, 'a'}, {2, 'b'}} {
		m.PushValue(TypedValue{T: IntType, N: i2n(kv[0])})
		m.PushValue(TypedValue{T: Uint8Type, N: i2n(kv[1])})
	}
	m.PushExpr(&CompositeLitExpr{Elts: make([]KeyValueExpr, 2)})
	defer func() {
		got := fmt.Sprint(recover())
		if want := "duplicate index 2 in array or slice literal"; got != want {
			t.Fatalf("got panic %q, want %q", got, want)
		}
	}()
	m.doOpSliceLit2()
}

func assertDataBacked(t *testing.T, m *Machine, length int, want string) {
	t.Helper()
	sv, ok := m.PeekValue(1).V.(*SliceValue)
	if !ok {
		t.Fatalf("not a slice: %T", m.PeekValue(1).V)
	}
	if sv.Length != length {
		t.Fatalf("length %d, want %d", sv.Length, length)
	}
	base := sv.GetBase(m.Store)
	if base.Data == nil {
		t.Fatalf("List-backed (%d TypedValues), want Data-backed", len(base.List))
	}
	if got := string(base.Data[:sv.Length]); got != want {
		t.Errorf("contents %q, want %q", got, want)
	}
}

// Natives returning []byte reach Gno through Go2GnoValue; its slice arm used
// to build a List backing even for bytes, unlike its own reflect.Array arm.
func TestGo2GnoByteSliceIsDataBacked(t *testing.T) {
	type myByte byte
	for _, src := range []any{
		[]byte{1, 2, 3},
		[]byte(nil),
		// Value.Bytes keys off element kind, so a named byte type must work
		// too; reflect.Copy would panic here.
		[]myByte{1, 2, 3},
	} {
		t.Run(fmt.Sprintf("%T", src), func(t *testing.T) {
			alloc := NewAllocator(math.MaxInt64)
			tv := Go2GnoValue(alloc, nil, reflect.ValueOf(src))
			rv := reflect.ValueOf(src)
			if rv.Len() == 0 {
				if tv.V != nil {
					if sv := tv.V.(*SliceValue); sv.GetBase(nil).Data == nil {
						t.Fatal("empty slice is List-backed")
					}
				}
				return
			}
			sv, ok := tv.V.(*SliceValue)
			if !ok {
				t.Fatalf("not a slice: %T", tv.V)
			}
			base := sv.GetBase(nil)
			if base.Data == nil {
				t.Fatalf("List-backed (%d TypedValues), want Data-backed", len(base.List))
			}
			if got, want := string(base.Data[:sv.Length]), "\x01\x02\x03"; got != want {
				t.Errorf("contents %q, want %q", got, want)
			}
		})
	}
}
