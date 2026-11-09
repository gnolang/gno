package gnolang

import (
	"io"
	"strings"
	"testing"

	stypes "github.com/gnolang/gno/tm2/pkg/store/types"
	"github.com/stretchr/testify/require"
)

// TestGarbageCollect_VisitsOperandStack is the regression test for GC
// skipping m.Values. An object reachable only from the operand stack must
// (1) be recounted by GarbageCollect, (2) be charged in the GC visit gas,
// and (3) block an over-cap Allocate through the real collect() retry path
// instead of being silently dropped from the tally.
func TestGarbageCollect_VisitsOperandStack(t *testing.T) {
	const maxBytes = 1 << 20 // 1 MiB
	const n = 600_000        // one data array: > half the cap

	m := newOperandStackTestMachine(t, maxBytes, false)

	// Allocate a data array (charged to m.Alloc) and park it on the operand
	// stack only — no block, frame, or package references it.
	av := m.Alloc.NewDataArray(Uint8Type, n)
	m.PushValue(TypedValue{T: &ArrayType{Len: n, Elt: Uint8Type}, V: av})
	_, before := m.Alloc.Status()
	require.GreaterOrEqual(t, before, int64(n), "array must be charged before GC")

	// (1) Direct GC: the parked array must survive Reset()+Recount().
	_, ok := m.GarbageCollect()
	require.True(t, ok)
	_, after := m.Alloc.Status()
	require.GreaterOrEqual(t, after, int64(n),
		"GC dropped an object reachable only from the operand stack: before=%d after=%d", before, after)

	// (2) Gas: the operand-stack object must be charged as a GC visit. Read
	// the gas GarbageCollect itself consumes (gcVisitGas(visitCount), the
	// consensus-visible axis) on two machines that differ only in whether
	// the array is parked on the stack. Measuring the real path matters:
	// re-implementing the root walk in the test would make this assertion
	// pass with the m.Values loop deleted from GarbageCollect.
	gasParked := gcGasForOperandStack(t, maxBytes, n, true)
	gasEmpty := gcGasForOperandStack(t, maxBytes, n, false)
	require.Greater(t, gasParked, gasEmpty,
		"operand-stack object not counted in GC visit gas: parked=%d empty=%d", gasParked, gasEmpty)

	// (3) Real path: an Allocate that overflows the cap triggers collect().
	// The array is still live on the stack, so GC cannot free it and the
	// allocation must be refused rather than granted against phantom
	// headroom.
	require.PanicsWithValue(t, "allocation limit exceeded", func() {
		m.Alloc.Allocate(n)
	})
}

// gcGasForOperandStack returns the gas a single GarbageCollect consumes on a
// fresh machine, optionally with an n-byte data array parked on the operand
// stack. Only the GC charge is measured: the array is allocated before the
// meter is read, and Recount (unlike Allocate) charges no gas.
func gcGasForOperandStack(t *testing.T, maxBytes int64, n int, park bool) int64 {
	t.Helper()
	m := newOperandStackTestMachine(t, maxBytes, true)
	if park {
		av := m.Alloc.NewDataArray(Uint8Type, n)
		m.PushValue(TypedValue{T: &ArrayType{Len: n, Elt: Uint8Type}, V: av})
	}
	before := m.GasMeter.GasConsumed()
	_, ok := m.GarbageCollect()
	require.True(t, ok)
	return m.GasMeter.GasConsumed() - before
}

func newOperandStackTestMachine(t *testing.T, maxBytes int64, withGasMeter bool) *Machine {
	t.Helper()
	opts := MachineOptions{
		PkgPath: "test/gcstack",
		Alloc:   NewAllocator(maxBytes),
		Output:  io.Discard,
	}
	if withGasMeter {
		opts.GasMeter = stypes.NewInfiniteGasMeter()
	}
	m := NewMachineWithOptions(opts)
	t.Cleanup(m.Release)
	return m
}

// TestGarbageCollect_VisitsAnchoredBuffer is the regression test for the
// second root: a buffer an op has allocated and is still filling. It is
// referenced by no machine root — only by Allocator.anchors — so before the
// anchor set existed, GC dropped it from the tally and Allocate re-granted
// the cap once per entry written.
func TestGarbageCollect_VisitsAnchoredBuffer(t *testing.T) {
	const maxBytes = 1 << 20 // 1 MiB
	const n = 600_000        // one data array: > half the cap

	m := newOperandStackTestMachine(t, maxBytes, false)

	// Stand in for a half-filled composite-literal element buffer: the array
	// is charged, and the only thing holding it is the buffer.
	buf := make([]TypedValue, 1)
	av := m.Alloc.NewDataArray(Uint8Type, n)
	buf[0] = TypedValue{T: &ArrayType{Len: n, Elt: Uint8Type}, V: av}
	m.Alloc.PushAnchor(buf)
	defer m.Alloc.PopAnchor()

	_, ok := m.GarbageCollect()
	require.True(t, ok)
	_, after := m.Alloc.Status()
	require.GreaterOrEqual(t, after, int64(n),
		"GC dropped a buffer that is only reachable through the anchor set: after=%d", after)

	// The buffer is still being filled, so GC cannot free it and the next
	// allocation must be refused rather than granted against phantom headroom.
	require.PanicsWithValue(t, "allocation limit exceeded", func() {
		m.Alloc.Allocate(n)
	})
}

// TestGarbageCollect_VisitsBlankSlot pins the visit of Block.Blank: the block
// holds whatever lands there, so the collection must count it.
func TestGarbageCollect_VisitsBlankSlot(t *testing.T) {
	const maxBytes = 1 << 20 // 1 MiB
	const n = 600_000        // one data array: > half the cap

	m := newOperandStackTestMachine(t, maxBytes, false)
	av := m.Alloc.NewDataArray(Uint8Type, n)
	b := &Block{Blank: TypedValue{T: &ArrayType{Len: n, Elt: Uint8Type}, V: av}}
	m.Blocks = append(m.Blocks, b)

	_, ok := m.GarbageCollect()
	require.True(t, ok)
	_, after := m.Alloc.Status()
	require.GreaterOrEqual(t, after, int64(n),
		"GC dropped a value held only by a block's blank slot: after=%d", after)
}

// TestStructValueCopy_AnchorHoldsThroughNewStruct pins the anchor in
// StructValue.Copy across its NewStruct call: the field buffer is a Go local
// until NewStruct wraps it, so a collection NewStruct triggers must still
// count the fields already copied into it.
func TestStructValueCopy_AnchorHoldsThroughNewStruct(t *testing.T) {
	const n = 600_000

	m := newOperandStackTestMachine(t, 1<<30, false)
	at := &ArrayType{Len: n, Elt: Uint8Type}
	st := &StructType{Fields: []FieldType{{Name: "F", Type: at}}}
	// The source is built off the allocator, so only the copy is charged.
	src := &StructValue{Fields: []TypedValue{{T: at, V: &ArrayValue{Data: make([]byte, n)}}}}

	// Room for the field buffer and the copied array, one byte short of the
	// struct header: NewStruct is the allocation that collects.
	_, before := m.Alloc.Status()
	m.Alloc.maxBytes = before + allocStructField + allocArray + n + allocStruct - 1
	counted := int64(-1)
	m.Alloc.SetGCFn(func() (int64, bool) {
		left, ok := m.GarbageCollect()
		_, counted = m.Alloc.Status()
		return left, ok
	})

	func() {
		defer func() { _ = recover() }()
		src.Copy(m.Alloc, nil, st, nil)
	}()
	require.GreaterOrEqual(t, counted, int64(n),
		"collection in NewStruct dropped the copied field: counted=%d", counted)
}

// TestDefaultStructValue_AnchorHoldsThroughNewStruct pins the same anchor for
// a zero value: defaultStructValue's field buffer is a Go local until
// NewStruct wraps it, so a collection NewStruct triggers must still count the
// fields already filled.
func TestDefaultStructValue_AnchorHoldsThroughNewStruct(t *testing.T) {
	const n = 600_000

	m := newOperandStackTestMachine(t, 1<<30, false)
	at := &ArrayType{Len: n, Elt: Uint8Type}
	st := &StructType{Fields: []FieldType{{Name: "F", Type: at}}}

	// Room for the field buffer and the zero array, one byte short of the
	// struct header: NewStruct is the allocation that collects.
	_, before := m.Alloc.Status()
	m.Alloc.maxBytes = before + allocStructField + allocArray + n + allocStruct - 1
	counted := int64(-1)
	m.Alloc.SetGCFn(func() (int64, bool) {
		left, ok := m.GarbageCollect()
		_, counted = m.Alloc.Status()
		return left, ok
	})

	func() {
		defer func() { _ = recover() }()
		defaultStructValue(m.Alloc, st)
	}()
	require.GreaterOrEqual(t, counted, int64(n),
		"collection in NewStruct dropped the zero-filled field: counted=%d", counted)
}

// TestNew_AnchorHoldsThroughPointer pins the anchor in the new builtin: its
// zero value is a Go local until the pointer is pushed, so a collection the
// pointer allocation triggers must still count it.
func TestNew_AnchorHoldsThroughPointer(t *testing.T) {
	const n = 600_000

	m := newOperandStackTestMachine(t, 1<<30, false)
	at := &ArrayType{Len: n, Elt: Uint8Type}
	fv := UverseNode().GetSlot(nil, "new", true).V.(*FuncValue)
	// new reads its type argument from the last block's first slot.
	m.Blocks = append(m.Blocks, &Block{Values: []TypedValue{asValue(at)}})

	// Room for the zero array, one byte short of the pointer: AllocatePointer
	// is the allocation that collects.
	_, before := m.Alloc.Status()
	m.Alloc.maxBytes = before + allocArray + n + allocPointer - 1
	counted := int64(-1)
	m.Alloc.SetGCFn(func() (int64, bool) {
		left, ok := m.GarbageCollect()
		_, counted = m.Alloc.Status()
		return left, ok
	})

	func() {
		defer func() { _ = recover() }()
		fv.nativeBody(m)
	}()
	require.GreaterOrEqual(t, counted, int64(n),
		"collection in new dropped the zero value: counted=%d", counted)
}

// TestByteSlices_AnchorHoldsThroughNewSlice pins the anchor each byte branch
// of make and append holds on its new data array: the array is a Go local
// until the slice is pushed, so a collection NewSlice triggers must still
// count it.
func TestByteSlices_AnchorHoldsThroughNewSlice(t *testing.T) {
	const n = 600_000
	bt := &SliceType{Elt: Uint8Type}
	bytes := func(l, c int) TypedValue {
		return TypedValue{T: bt, V: &SliceValue{Base: &ArrayValue{Data: make([]byte, c)}, Length: l, Maxcap: c}}
	}
	sizes := func(ns ...int) TypedValue {
		tvs := make([]TypedValue, len(ns))
		for i, x := range ns {
			tvs[i] = typedInt(x)
		}
		return TypedValue{
			T: &SliceType{Elt: &InterfaceType{}, Vrd: true},
			V: &SliceValue{Base: &ArrayValue{List: tvs}, Length: len(ns), Maxcap: len(ns)},
		}
	}
	for _, tc := range []struct {
		name   string
		native Name
		params []TypedValue // the native's parameter block
		newLen int64        // bytes in the array it builds
	}{
		{"make([]byte, n)", "make", []TypedValue{asValue(bt), sizes(n)}, n},
		{"make([]byte, 0, n)", "make", []TypedValue{asValue(bt), sizes(0, n)}, n},
		{"append(nil, b...)", "append", []TypedValue{{T: bt}, bytes(n, n)}, n},
		{"append(b, c...) past cap(b)", "append", []TypedValue{bytes(1, 1), bytes(n, n)}, n + 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := newOperandStackTestMachine(t, 1<<30, false)
			fv := UverseNode().GetSlot(nil, tc.native, true).V.(*FuncValue)
			m.Blocks = append(m.Blocks, &Block{Values: tc.params})

			// Count what is live, operands included, then leave room for the
			// new array, one byte short of its slice: NewSlice is the
			// allocation that collects.
			m.GarbageCollect()
			_, base := m.Alloc.Status()
			m.Alloc.maxBytes = base + allocArray + tc.newLen + allocSlice - 1
			counted := int64(-1)
			m.Alloc.SetGCFn(func() (int64, bool) {
				left, ok := m.GarbageCollect()
				_, counted = m.Alloc.Status()
				return left, ok
			})

			func() {
				defer func() { _ = recover() }()
				fv.nativeBody(m)
			}()
			require.GreaterOrEqual(t, counted, base+tc.newLen,
				"collection in NewSlice dropped the data array: counted=%d base=%d", counted, base)
		})
	}
}

// TestAppendString_AnchorHoldsTempCopy pins the anchor on the byte copy
// append(b, s...) makes of s: the copy is a Go local until append returns, so
// the collection the result's allocation triggers must still count it.
func TestAppendString_AnchorHoldsTempCopy(t *testing.T) {
	const n = 600_000

	m := newOperandStackTestMachine(t, 1<<30, false)
	fv := UverseNode().GetSlot(nil, "append", true).V.(*FuncValue)
	// append reads x and the variadic args from the last block's two slots.
	m.Blocks = append(m.Blocks, &Block{Values: []TypedValue{
		{T: &SliceType{Elt: Uint8Type}},
		typedString(strings.Repeat("x", n)),
	}})

	// Count what is live before append runs, then leave room for the copy
	// and its slice but not for the result: the result's array is the
	// allocation that collects.
	m.GarbageCollect()
	_, base := m.Alloc.Status()
	m.Alloc.maxBytes = base + allocArray + n + 1000
	counted := int64(-1)
	m.Alloc.SetGCFn(func() (int64, bool) {
		left, ok := m.GarbageCollect()
		_, counted = m.Alloc.Status()
		return left, ok
	})

	func() {
		defer func() { _ = recover() }()
		fv.nativeBody(m)
	}()
	require.GreaterOrEqual(t, counted, base+n,
		"collection in append dropped the copy of s: counted=%d base=%d", counted, base)
}

// TestAllocatorAnchors_Unwind pins the contract Machine.runOnce relies on to
// release anchors left behind by an op that panicked part-way through a fill.
func TestAllocatorAnchors_Unwind(t *testing.T) {
	alloc := NewAllocator(1 << 20)
	require.Equal(t, 0, alloc.AnchorDepth())

	depth := alloc.AnchorDepth()
	alloc.PushAnchor(make([]TypedValue, 2))
	alloc.PushAnchor(make([]TypedValue, 3))
	require.Equal(t, depth+2, alloc.AnchorDepth())

	alloc.TruncateAnchors(depth)
	require.Equal(t, depth, alloc.AnchorDepth())
	// Truncation must not resurrect entries on the next push.
	alloc.PushAnchor(make([]TypedValue, 1))
	require.Len(t, alloc.anchors, 1)
	require.Len(t, alloc.anchors[0].tvs, 1)
	require.Nil(t, alloc.anchors[0].obj)

	// Truncating to a depth at or above the current one is a no-op.
	alloc.TruncateAnchors(5)
	require.Equal(t, 1, alloc.AnchorDepth())

	// Object anchors share the same LIFO stack and unwind identically.
	mv := alloc.NewMap(nil)
	alloc.PushAnchorValue(mv)
	require.Equal(t, 2, alloc.AnchorDepth())
	require.Same(t, mv, alloc.anchors[1].obj)
	require.Nil(t, alloc.anchors[1].tvs)
	alloc.PopAnchor()
	require.Equal(t, 1, alloc.AnchorDepth())

	// A nil allocator is valid and anchor-free.
	var nilAlloc *Allocator
	nilAlloc.PushAnchor(make([]TypedValue, 1))
	nilAlloc.PushAnchorValue(&MapValue{})
	require.Equal(t, 0, nilAlloc.AnchorDepth())
	nilAlloc.TruncateAnchors(0)
	nilAlloc.PopAnchor()
}
