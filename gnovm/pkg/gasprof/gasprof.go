// Package gasprof is a source-level gas profiler for the GnoVM. It attributes
// gas charged during execution to the gno call stack and emits a standard
// pprof protobuf profile, so the whole go tool pprof / flame-graph ecosystem
// works on gno gas.
//
// It captures every gas dimension — CPU, allocation, store I/O, amino, and
// refunds — via a GasMeter decorator (WrapMeter), and attributes each charge to
// the current gno function through an incremental call-tree cursor driven by the
// machine's frame push/pop events (O(1) per charge). See
// gnovm/adr/prxxxx_gas_profiler.md.
//
// Surfaces that drive it: gno test -gasprofile (unit tests + filetests) and the
// on-chain .app/profiletx ABCI query (gno.land/pkg/gnoclient.ProfileTx, exposed
// by gnodev). Output is a standard multi-dimension pprof; view with go tool
// pprof and switch dimension with -sample_index=cpu_gas|store_gas|...
//
// gasprof imports only stdlib + tm2/pkg/store (the GasMeter interface); it does
// NOT import gnovm/pkg/gnolang, so there is no import cycle — the machine calls
// Enter/SyncDepth/Reset and installs the decorator via WrapMeter.
package gasprof

import (
	"compress/gzip"
	"encoding/binary"
	"io"
	"strconv"

	"github.com/gnolang/gno/tm2/pkg/store"
)

// Gas dimensions. Order matters: it defines the pprof value indices.
const (
	dimCPU = iota
	dimAlloc
	dimStore
	dimOther
	dimRefund
	numDim
)

// dimensionOf maps a GasMeter descriptor to a dimension. The table mirrors the
// repo-wide descriptor inventory (see the ADR). Unknown descriptors fall into
// dimOther so no gas ever leaks out of the profile.
func dimensionOf(descriptor string) int {
	switch descriptor {
	case "CPUCycles", "GC", "parsing", "ComputeMapKey", "stream output":
		return dimCPU
	case "memory allocation":
		return dimAlloc
	case "DepthReadFlat", "DepthSet", "DepthDelete",
		"ReadFlat", "ReadPerByte", "WriteFlat", "Delete",
		"IterNextFlat", "ValuePerByte",
		"AminoEncodePerByte", "AminoDecodePerByte":
		return dimStore
	default:
		return dimOther
	}
}

// Frame is one gno call-stack entry, supplied by the machine when a call frame
// is pushed. Func is a fully-qualified display name (already receiver- and
// package-qualified).
type Frame struct {
	Func string
	File string
	Line int
}

type nodeKey struct {
	name string
	file string
	line int
}

type node struct {
	name     string
	file     string
	line     int
	flat     [numDim]int64
	children map[nodeKey]*node
	order    []nodeKey // stable child emission order (insertion order)
}

func (n *node) child(f Frame) *node {
	k := nodeKey{f.Func, f.File, f.Line}
	if c := n.children[k]; c != nil {
		return c
	}
	c := &node{name: f.Func, file: f.File, line: f.Line}
	if n.children == nil {
		n.children = make(map[nodeKey]*node)
	}
	n.children[k] = c
	n.order = append(n.order, k)
	return c
}

// Profiler accumulates gas into a call tree. It is not safe for concurrent use;
// a gno Machine runs single-goroutine.
type Profiler struct {
	root  *node
	stack []*node // cursor path; stack[0] is root, stack[i] is the i-th call frame
}

// New returns an empty profiler with its cursor at the synthetic root. Gas
// charged before any call frame (e.g. while setting up the entry call)
// attributes to root, shown as "(root)".
func New() *Profiler {
	r := &node{name: "(root)"}
	return &Profiler{root: r, stack: []*node{r}}
}

// Enter descends the cursor into the called function. The machine calls this
// once per pushed call frame (Func != nil, closures included).
func (p *Profiler) Enter(f Frame) {
	parent := p.stack[len(p.stack)-1]
	p.stack = append(p.stack, parent.child(f))
}

// Pop ascends the cursor by one, the O(1) inverse of Enter. Called on a single
// call-frame return. No-op at the root.
func (p *Profiler) Pop() {
	if len(p.stack) > 1 {
		p.stack = p.stack[:len(p.stack)-1]
	}
}

// SyncDepth re-syncs the cursor to a machine call-frame count, ascending as
// needed. Used on the bulk-truncation path (revive unwinding removes several
// call frames at once), where a single Pop is insufficient. callFrames counts
// frames with Func != nil (matching what Enter descends on).
func (p *Profiler) SyncDepth(callFrames int) {
	target := callFrames + 1 // + root
	if target < 1 {
		target = 1
	}
	if target < len(p.stack) {
		p.stack = p.stack[:target]
	}
}

// Reset returns the cursor to the root. Called at machine teardown, where the
// frame stack is discarded wholesale without per-frame pops.
func (p *Profiler) Reset() { p.stack = p.stack[:1] }

// Book records gas charged before the profiler was installed (e.g. ante-handler
// txSize/signature gas, snapshotted at install time) as a synthetic root-child
// node in the "other" dimension. This keeps reconciliation exact:
// sum(tree gross) − refunds == meter.GasConsumed().
func (p *Profiler) Book(name string, gas int64) {
	if gas == 0 {
		return
	}
	p.root.child(Frame{Func: name}).flat[dimOther] += gas
}

// record adds gas to the current cursor node under dim.
func (p *Profiler) record(dim int, amount int64) {
	p.stack[len(p.stack)-1].flat[dim] += amount
}

// Empty reports whether nothing was recorded.
func (p *Profiler) Empty() bool {
	return len(p.root.children) == 0 && p.root.zero()
}

// Totals returns the summed gas per dimension across the whole tree.
type Totals struct{ CPU, Alloc, Store, Other, Refund int64 }

func (p *Profiler) Totals() Totals {
	var t [numDim]int64
	var walk func(n *node)
	walk = func(n *node) {
		for d := range numDim {
			t[d] += n.flat[d]
		}
		for _, k := range n.order {
			walk(n.children[k])
		}
	}
	walk(p.root)
	return Totals{t[dimCPU], t[dimAlloc], t[dimStore], t[dimOther], t[dimRefund]}
}

// ---------------------------------------------------------------------------
// GasMeter decorator: observes charges, attributes them to the cursor. Delegates
// everything to the wrapped meter with identical semantics (observation only —
// never changes gas). Install via WrapMeter as the machine's GasMeter.
// ---------------------------------------------------------------------------

type meter struct {
	inner store.GasMeter
	prof  *Profiler
}

// WrapMeter returns a GasMeter that delegates to inner and records every charge
// into p. One wrapper per meter — wrapping a wrapper double-records.
func WrapMeter(inner store.GasMeter, p *Profiler) store.GasMeter {
	return &meter{inner: inner, prof: p}
}

// Unwrap returns the wrapped meter (used to disable profiling).
func (m *meter) Unwrap() store.GasMeter { return m.inner }

func (m *meter) ConsumeGas(amount store.Gas, descriptor string) {
	// Record before delegating. ConsumeGas never clamps the amount, so `amount`
	// is exactly the delta applied — including on the out-of-gas path, where the
	// meter mutates consumed and *then* panics, so recording first captures that
	// final charge and the profile still reconciles through the panic. (The only
	// path where amount != applied delta is int64 overflow, which panics before
	// mutating; that requires ~9.2e18 consumed gas and is terminal, so the
	// resulting one-charge over-count is unreachable in practice.)
	if amount != 0 {
		m.prof.record(dimensionOf(descriptor), amount)
	}
	m.inner.ConsumeGas(amount, descriptor)
}

func (m *meter) RefundGas(amount store.Gas, descriptor string) {
	// Refunds are clamped to the consumed total, so record the applied delta,
	// not the requested amount. Booked as a separate positive dimension.
	before := m.inner.GasConsumed()
	m.inner.RefundGas(amount, descriptor)
	if applied := before - m.inner.GasConsumed(); applied != 0 {
		m.prof.record(dimRefund, applied)
	}
}

func (m *meter) GasConsumed() store.Gas        { return m.inner.GasConsumed() }
func (m *meter) GasConsumedToLimit() store.Gas { return m.inner.GasConsumedToLimit() }
func (m *meter) Limit() store.Gas              { return m.inner.Limit() }
func (m *meter) Remaining() store.Gas          { return m.inner.Remaining() }
func (m *meter) IsPastLimit() bool             { return m.inner.IsPastLimit() }
func (m *meter) IsOutOfGas() bool              { return m.inner.IsOutOfGas() }

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

// WriteFolded writes Brendan-Gregg "folded" stacks ("root;child;leaf gas"),
// one per node with non-zero gas, root-first. The value is the node's flat
// total gas (cpu+alloc+store+other, minus refunds). Input format for
// flame-graph tools.
func (p *Profiler) WriteFolded(w io.Writer) error {
	var path []string
	var walk func(n *node) error
	walk = func(n *node) error {
		path = append(path, n.name)
		if g := n.flatTotal(); g != 0 {
			var line []byte
			for i, name := range path {
				if i > 0 {
					line = append(line, ';')
				}
				line = append(line, name...)
			}
			line = append(line, ' ')
			line = strconv.AppendInt(line, g, 10)
			line = append(line, '\n')
			if _, err := w.Write(line); err != nil {
				return err
			}
		}
		for _, k := range n.order {
			if err := walk(n.children[k]); err != nil {
				return err
			}
		}
		path = path[:len(path)-1]
		return nil
	}
	return walk(p.root)
}

// grossTotal is the billable gas before refunds (cpu+alloc+store+other).
func (n *node) grossTotal() int64 {
	return n.flat[dimCPU] + n.flat[dimAlloc] + n.flat[dimStore] + n.flat[dimOther]
}

// flatTotal is the net gas (gross minus refunds).
func (n *node) flatTotal() int64 { return n.grossTotal() - n.flat[dimRefund] }

// WritePprof writes a gzip-compressed pprof profile with one value index per
// dimension: [cpu, alloc, store, other, refund, total].
func (p *Profiler) WritePprof(w io.Writer) error {
	st := newStrtab()
	// sample types, in value-index order.
	types := []struct{ typ, unit string }{
		{"cpu_gas", "gas"},
		{"alloc_gas", "gas"},
		{"store_gas", "gas"},
		{"other_gas", "gas"},
		{"refund_gas", "gas"},
		{"total_gas", "gas"},
	}

	var prof buf
	for _, ty := range types {
		var vt buf
		vt.uint(1, uint64(st.get(ty.typ)))
		vt.uint(2, uint64(st.get(ty.unit)))
		prof.msg(1, vt.b) // Profile.sample_type = 1
	}

	locIDs := make(map[nodeKey]uint64)
	var nextID uint64 = 1
	getLoc := func(n *node) uint64 {
		k := nodeKey{n.name, n.file, n.line}
		if id, ok := locIDs[k]; ok {
			return id
		}
		id := nextID
		nextID++
		locIDs[k] = id

		var fn buf
		fn.uint(1, id)
		fn.uint(2, uint64(st.get(n.name)))
		fn.uint(4, uint64(st.get(n.file)))
		fn.int(5, int64(n.line))
		prof.msg(5, fn.b) // Profile.function = 5

		var ln buf
		ln.uint(1, id)
		ln.int(2, int64(n.line))
		var loc buf
		loc.uint(1, id)
		loc.msg(4, ln.b)
		prof.msg(4, loc.b) // Profile.location = 4
		return id
	}

	// One sample per node with non-zero gas. locations are leaf-first.
	var locStack []uint64 // root..node
	var walk func(n *node)
	walk = func(n *node) {
		locStack = append(locStack, getLoc(n))
		if !n.zero() {
			var smp buf
			for i := len(locStack) - 1; i >= 0; i-- { // leaf-first
				smp.uint(1, locStack[i])
			}
			// Values in dimension order, then the gross total. Matches the
			// sample_type order emitted above.
			for d := range numDim {
				smp.int(2, n.flat[d])
			}
			smp.int(2, n.grossTotal())
			prof.msg(2, smp.b) // Profile.sample = 2
		}
		for _, k := range n.order {
			walk(n.children[k])
		}
		locStack = locStack[:len(locStack)-1]
	}
	walk(p.root)

	prof.int(14, int64(st.get("total_gas"))) // default_sample_type

	for _, s := range st.list {
		prof.str(6, s) // string_table = 6, emitted last in index order
	}

	gz := gzip.NewWriter(w)
	if _, err := gz.Write(prof.b); err != nil {
		return err
	}
	return gz.Close()
}

func (n *node) zero() bool { return n.flat == [numDim]int64{} }

// ---------------------------------------------------------------------------
// tiny protobuf writer (varint + length-delimited only) and string table
// ---------------------------------------------------------------------------

type strtab struct {
	idx  map[string]int
	list []string
}

func newStrtab() *strtab {
	t := &strtab{idx: make(map[string]int)}
	t.get("") // index 0 must be ""
	return t
}

func (t *strtab) get(s string) int {
	if i, ok := t.idx[s]; ok {
		return i
	}
	i := len(t.list)
	t.idx[s] = i
	t.list = append(t.list, s)
	return i
}

type buf struct{ b []byte }

func (p *buf) tag(field, wire int) {
	p.b = binary.AppendUvarint(p.b, uint64(field)<<3|uint64(wire))
}
func (p *buf) uint(field int, v uint64) {
	p.tag(field, 0)
	p.b = binary.AppendUvarint(p.b, v)
}
func (p *buf) int(field int, v int64) { p.uint(field, uint64(v)) }
func (p *buf) bytes(field int, data []byte) {
	p.tag(field, 2)
	p.b = binary.AppendUvarint(p.b, uint64(len(data)))
	p.b = append(p.b, data...)
}
func (p *buf) str(field int, s string) { p.bytes(field, []byte(s)) }
func (p *buf) msg(field int, m []byte) { p.bytes(field, m) }
