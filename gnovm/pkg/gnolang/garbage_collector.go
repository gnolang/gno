package gnolang

import (
	"math/bits"
	"reflect"

	"github.com/gnolang/gno/tm2/pkg/overflow"
)

// gcVisitGasTable[k] = gas per GC visit when log2(visitCount) == k.
// 1 gas = 1 nanosecond on reference hardware.
// Per-visit cost increases with heap size due to CPU cache effects:
// small heaps fit in L2/L3 (~29ns/visit), large heaps hit DRAM (~700ns/visit).
//
// Calibrated from BenchmarkGCVisit on DigitalOcean Dedicated (2-core),
// Intel Xeon Platinum 8168 @ 2.70GHz.
//
// See gnovm/pkg/gnolang/bench_gc_test.go for benchmarks.
var gcVisitGasTable = [25]int64{
	29, 29, 29, 29, 29, 29, 29, // 2^0 - 2^6:   1-64 visits       (~29ns, L1/L2)
	40, 40, 40, // 2^7 - 2^9:   128-512 visits     (~40ns, L2/L3)
	91, 91, 91, // 2^10 - 2^12: 1K-4K visits       (~91ns, L3)
	160, 160, 160, 197, // 2^13 - 2^16: 8K-64K visits      (~160-197ns, L3/DRAM)
	290, 290, 380, // 2^17 - 2^19: 128K-512K visits   (~290-380ns, DRAM)
	380, 380, 520, // 2^20 - 2^22: 1M-4M visits       (~380-520ns, DRAM)
	700, 700, // 2^23 - 2^24: 8M-16M visits      (~700ns, DRAM+TLB)
}

// gcStackScanSlopeGas is the cost of inspecting one TypedValue slot during
// GarbageCollect's linear walks over m.Values and Allocator.anchors.
//
// Priced separately from gcVisitGasTable: a slot scan is a sequential pass
// over a contiguous []TypedValue, not a pointer-chasing object visit, and a
// slot whose V is nil (every primitive) produces no visit at all. Without
// this the walks are O(len(m.Values) + sum(len(anchors))) work charged zero
// gas, which a deep operand stack can inflate without bound.
//
// Rough calibration from BenchmarkGCStackScan (developer machine, NOT the
// Xeon Platinum 8168 that gcVisitGasTable is anchored on): 1.3 ns/slot
// in-cache at 1K slots rising to 3.3 ns/slot DRAM-bound at 10M. 4 sits just
// above the measured maximum and matches OpCPUSlopeCopyPrimitive, the
// existing per-element slope for a comparable linear pass. Re-measure on
// reference hardware before relying on the exact figure.
const gcStackScanSlopeGas = 4

// gcStackScanGas returns total gas for a linear walk over slots TypedValues.
func gcStackScanGas(slots int64) int64 {
	if slots <= 0 {
		return 0
	}
	return overflow.Mulp(slots, gcStackScanSlopeGas)
}

// gcScannedSlots reports how many TypedValue slots GarbageCollect's linear
// walks inspect: the operand stack plus every in-flight anchor buffer.
func (m *Machine) gcScannedSlots() int64 {
	slots := int64(len(m.Values))
	if m.Alloc != nil {
		for _, a := range m.Alloc.anchors {
			slots = overflow.Addp(slots, int64(len(a.tvs)))
		}
	}
	return slots
}

// gcVisitGas returns total gas for a GC traversal of visitCount objects.
// Uses a per-visit cost that scales with heap size (cache effects).
func gcVisitGas(visitCount int64) int64 {
	if visitCount <= 0 {
		return 0
	}
	k := bits.Len64(uint64(visitCount)) - 1
	if k >= len(gcVisitGasTable) {
		k = len(gcVisitGasTable) - 1
	}
	return overflow.Mulp(visitCount, gcVisitGasTable[k])
}

// Visit visits all reachable associated values.
// It is used primarily for GC.
// The caller must provide a callback visitor
// which knows how to break cycles, otherwise
// the Visit function may recurse infinitely.
// (the GC does this with GCCycle)
// It does not call the visitor on itself.
type Visitor func(v Value) (stop bool)

// Returns the amount of memory left over. If the allocator limit is exceeded
// it returns false.  It doesn't actually garbage collect, but it recalculates
// allocated memory from what is already reachable.
// NOTE:
//
//	the tv.T types must not be measured.  this is because the types are
//	supposed to pre-exist, and memory allocation for tv.T depends on the
//	impl, whether it re-uses the same Type or not.
//
// XXX: make sure tv.T isn't bumped from allocation either.
// XXX: record original value and verify after GC
func (m *Machine) GarbageCollect() (left int64, ok bool) {
	// times objects are visited for gc
	var visitCount int64

	// Slots scanned by the linear walks below. Captured before the walks so
	// the charge is unaffected by an early return; neither m.Values nor
	// Allocator.anchors is mutated by GC.
	scannedSlots := m.gcScannedSlots()

	defer func() {
		gasCPU := overflow.Addp(gcVisitGas(visitCount), gcStackScanGas(scannedSlots))
		if debug {
			debug.Printf("GasConsumed for GC: %v\n", gasCPU)
		}
		if m.GasMeter != nil {
			m.GasMeter.ConsumeGas(gasCPU, "GC")
		}
	}()

	defer func() {
		m.Store.GarbageCollectObjectCache(m.GCCycle)
	}()

	// We don't need the old value anymore.
	m.Alloc.Reset()

	// This is the only place where it's bumped.
	m.GCCycle += 1

	// Construct visitor callback.
	vis := GCVisitorFn(m.GCCycle, m.Alloc, &visitCount)

	// Visit the operand stack (m.Values): intermediate results, call args
	// not yet popped into a block, composite-literal elements waiting for
	// doOp*Lit. Objects reachable only from here were previously invisible
	// to GC, so Reset()+Recount() dropped them from alloc.bytes and the Go
	// heap could grow past maxBytes.
	//
	// Known over-counts, both fail-safe (cap trips earlier, never later):
	//  - Recount(allocTypedValue) counts the 40B slot wrapper that runtime
	//    never charges (as Frame.Visit and Exception.Visit already do).
	//    Bounded: each slot costs metered work (a source byte or a charged
	//    Block). Kept on purpose: for primitive slots (V == nil) it is the
	//    only thing the cap sees, so the stack's own footprint stays counted.
	//  - vis dedups Objects only, so a string aliased into N slots is
	//    charged N times its length; same at every other root (blocks,
	//    arrays, frames). A follow-up will dedup string backings.
	for i := range m.Values {
		m.Alloc.Recount(allocTypedValue)
		if v := m.Values[i].V; v != nil {
			if stop := vis(v); stop {
				return -1, false
			}
		}
	}

	// Visit buffers still under construction (Allocator.anchors): the
	// element, field, default-fill and argument buffers that ops allocate
	// as Go locals and fill one entry at a time. The operand-stack walk
	// above does not cover them — the sources it sees may all alias one
	// value ([]T{a, a, ...}), while each copy written into the buffer is a
	// fresh allocation — so without this a fill that trips the cap frees
	// its own headroom and continues. Recount the slots the buffer holds
	// (allocArrayItem, allocStructField and allocBlockItem are all
	// allocTypedValue); only the container header is missed, and just
	// until the op roots it.
	for _, a := range m.Alloc.anchors {
		if a.obj != nil {
			// Whole-object anchor (doOpMapLit): the visitor recounts it
			// and everything already written into it.
			if stop := vis(a.obj); stop {
				return -1, false
			}
			continue
		}
		m.Alloc.Recount(allocTypedValue * int64(len(a.tvs)))
		for i := range a.tvs {
			if v := a.tvs[i].V; v != nil {
				if stop := vis(v); stop {
					return -1, false
				}
			}
		}
	}

	// Visit blocks
	for _, block := range m.Blocks {
		if block == nil {
			continue
		}
		stop := vis(block)
		if stop {
			return -1, false
		}
	}

	// Account for blocks parked in the per-machine pool. They are dead from
	// the program's perspective, but the machine pins each one for reuse — a
	// Block plus its capacity-retained Values backing array — so the alloc
	// tally must include them; otherwise recycling a block would appear to
	// free memory that is in fact still held. Pooled blocks are zeroed and
	// reference nothing else, so count them directly (by capacity, the real
	// retained footprint) rather than walking them.
	for _, b := range m.blockPool {
		m.Alloc.Recount(allocBlock + allocBlockItem*int64(cap(b.Values)))
	}

	// Visit frames
	for _, frame := range m.Frames {
		stop := frame.Visit(m.Alloc, vis)
		if stop {
			return -1, false
		}
	}

	// Visit package
	stop := vis(m.Package)
	if stop {
		return -1, false
	}

	// Visit staging package.
	// Stating package is partially loaded package.
	// it's more efficient to vist it than to
	// iterate over the whole cache.
	if tpv := m.Store.GetStagingPackage(); tpv != nil {
		stop = vis(tpv)
		if stop {
			return -1, false
		}
	}

	// Visit exceptions
	if m.Exception != nil {
		e := m.Exception
		// Visit m.Exception and its previous Exceptions
		for e != nil {
			stop = e.Visit(m.Alloc, vis)
			if stop {
				return -1, false
			}
			e = e.Previous
		}

		// Visit next Exceptions
		e = m.Exception.Next
		for e != nil {
			stop = e.Visit(m.Alloc, vis)
			if stop {
				return -1, false
			}
			e = e.Next
		}
	}

	if debugAssert {
		// Recycle-safety invariant: a pooled (released) block must be
		// unreachable, so this recount must never visit one. Released blocks
		// are zeroed (LastGCCycle == 0); the visitor stamps the current cycle
		// on every object it reaches. So a pooled block carrying the current
		// cycle was reached — i.e. some live reference outlived its pop (the
		// hazard that removing Defer.Parent eliminated). Reference-path
		// agnostic: fires for any future regression that re-pins a dead block.
		for _, b := range m.blockPool {
			if b.GetLastGCCycle() == m.GCCycle {
				panic("GarbageCollect: recount reached a pooled block — a popped block is still GC-reachable (recycle-safety invariant violated)")
			}
		}
	}

	// Return bytes remaining.
	maxBytes, bytes := m.Alloc.Status()
	return maxBytes - bytes, true
}

// isUverseValue reports whether v is the global .uverse package value or its
// block — the only GC-reachable objects shared across machines.
func isUverseValue(v Value) bool {
	switch v := v.(type) {
	case *PackageValue:
		return v.PkgPath == uversePkgPath
	case *Block:
		if pn, ok := v.Source.(*PackageNode); ok {
			return pn.PkgPath == uversePkgPath
		}
	}
	return false
}

// Returns a visitor that bumps the GCCycle counter
// and stops if alloc is out of memory.
func GCVisitorFn(gcCycle int64, alloc *Allocator, visitCount *int64) Visitor {
	var vis func(value Value) bool

	vis = func(v Value) bool {
		if debug {
			debug.Printf("Visit, v: %v (type: %v)\n", v, reflect.TypeOf(v))
		}

		// The .uverse package is a process-global singleton shared by every
		// machine (SetCachePackage(Uverse())). Counting it here would write
		// per-machine GC state (LastGCCycle) into that shared object, so a
		// concurrent machine's GC could skip it — making GC gas
		// non-deterministic across parallel in-memory nodes. Its contents are
		// already excluded from traversal by the .uverse checks in the
		// VisitAssociated methods; exclude the package value and its block
		// from the visit count too, so GC gas never depends on shared state.
		if isUverseValue(v) {
			return false
		}

		if oo, isObject := v.(Object); isObject {
			// Return if already measured.
			if debug {
				debug.Printf("oo.GetLastGCCycle: %d, gcCycle: %d\n", oo.GetLastGCCycle(), gcCycle)
			}

			if oo.GetLastGCCycle() == gcCycle {
				return false // but don't stop
			}
		}

		*visitCount++ // Count operations for gas calculation

		// Add object size to alloc.
		size := v.GetShallowSize()

		// Stop if alloc max exceeded during GC.
		// NOTE: Unlikely to occur, but keep it here for
		// now to handle potential edge cases.
		// Consider removing it later if no issues arise.
		maxBytes, curBytes := alloc.Status()
		if maxBytes < curBytes+size {
			return true
		}

		alloc.Recount(size)

		// bump before visiting associated,
		// this avoids infinite recursion.
		if oo, isObject := v.(Object); isObject {
			oo.SetLastGCCycle(gcCycle)
		}

		// Invoke the traverser on v.
		stop := v.VisitAssociated(vis)

		return stop
	}
	return vis
}

// ---------------------------------------------------------------
// Visit associated

func (sv *SliceValue) VisitAssociated(vis Visitor) (stop bool) {
	// Visit base.
	if sv.Base != nil {
		stop = vis(sv.Base)
	}
	return
}

func (av *ArrayValue) VisitAssociated(vis Visitor) (stop bool) {
	// Visit each value.
	for i := 0; i < len(av.List); i++ {
		v := av.List[i].V
		if v == nil {
			continue
		}
		stop = vis(v)
		if stop {
			return
		}
	}
	return
}

func (fv *FuncValue) VisitAssociated(vis Visitor) (stop bool) {
	if fv.PkgPath == ".uverse" {
		return
	}
	// visit captures
	for _, tv := range fv.Captures {
		v := tv.V
		if v == nil {
			continue
		}
		stop = vis(v)
		if stop {
			return
		}
	}

	// Visit parent.
	switch v := fv.Parent.(type) {
	case nil:
		return
	case *Block:
		if v != nil {
			stop = vis(v)
		}
	case RefValue:
		stop = vis(v)
	}

	return
}

func (sv *StructValue) VisitAssociated(vis Visitor) (stop bool) {
	// Visit each value.
	for i := 0; i < len(sv.Fields); i++ {
		v := sv.Fields[i].V
		if v == nil {
			continue
		}
		stop = vis(v)
		if stop {
			return
		}
	}
	return
}

func (bmv *BoundMethodValue) VisitAssociated(vis Visitor) (stop bool) {
	// Visit receiver.
	v := bmv.Receiver.V
	if v != nil {
		stop = vis(v)
	}

	// Visit func
	fv := bmv.Func
	if fv != nil {
		stop = vis(fv)
	}

	return
}

func (mv *MapValue) VisitAssociated(vis Visitor) (stop bool) {
	// visit mv.List.
	for cur := mv.List.Head; cur != nil; cur = cur.Next {
		// vis key
		k := cur.Key.V
		if k != nil {
			stop = vis(k)
		}

		if stop {
			return
		}

		// vis value
		v := cur.Value.V
		if v != nil {
			stop = vis(v)
		}

		if stop {
			return
		}
	}
	return
}

func (pv *PackageValue) VisitAssociated(vis Visitor) (stop bool) {
	if pv.PkgPath == ".uverse" {
		return false
	}

	// visit pv.Block
	v := pv.Block
	if v != nil {
		stop = vis(pv.Block)
	}

	if stop {
		return
	}

	// visit pv.FBlocks
	for _, fb := range pv.FBlocks {
		if fb == nil {
			continue
		}

		stop = vis(fb)
		if stop {
			return
		}
	}

	// do NOT visit Realm.

	return
}

func (b *Block) VisitAssociated(vis Visitor) (stop bool) {
	// skip .uverse
	if pn, ok := b.Source.(*PackageNode); ok {
		if pn.PkgPath == ".uverse" {
			return
		}
	}

	// Visit each value.
	for i := 0; i < len(b.Values); i++ {
		v := b.Values[i].V
		if v == nil {
			continue
		}

		stop = vis(v)
		if stop {
			return
		}
	}

	// Visit the blank slot. `_ = expr` assigns through GetBlankRef, so the
	// block pins expr for its whole life even though nothing can read it
	// back. Skipping it made `_ = make([]byte, n)` per recursion level a
	// cap bypass: the value stays live on the Go heap while Reset()+Recount()
	// cannot see it (200 levels x 300KB held 60MB under a 1MB cap, while the
	// same expression bound to a name was correctly refused).
	if v := b.Blank.V; v != nil {
		stop = vis(v)
		if stop {
			return
		}
	}

	// Visit parent.
	switch v := b.Parent.(type) {
	case nil:
		return
	case *Block:
		if v != nil {
			stop = vis(v)
		}
	case RefValue:
		stop = vis(v)
	}

	return
}

func (hiv *HeapItemValue) VisitAssociated(vis Visitor) (stop bool) {
	v := hiv.Value.V
	if v != nil {
		stop = vis(hiv.Value.V)
	}
	return
}

func (pv PointerValue) VisitAssociated(vis Visitor) (stop bool) {
	// NOTE: *TV and Key will be visited along with base.
	v := pv.Base
	if v != nil {
		stop = vis(pv.Base)
	}
	return
}

func (sv StringValue) VisitAssociated(vis Visitor) (stop bool) {
	return false
}

func (biv BigintValue) VisitAssociated(vis Visitor) (stop bool) {
	return false
}

func (bdv BigdecValue) VisitAssociated(vis Visitor) (stop bool) {
	return false
}

func (dbv DataByteValue) VisitAssociated(vis Visitor) (stop bool) {
	return false
}

func (rv RefValue) VisitAssociated(vis Visitor) (stop bool) {
	return false
}

// Do not count the TypeValue, neither shallowly nor deeply.
func (tv TypeValue) VisitAssociated(vis Visitor) (stop bool) {
	return false
}

// -------------------------------------------------------------------
// Custom visit methods

func (fr *Frame) Visit(alloc *Allocator, vis Visitor) (stop bool) {
	// vis receiver
	if fr.Receiver.IsDefined() {
		alloc.Recount(allocTypedValue) // reclaim shallowly

		if v := fr.Receiver.V; v != nil {
			stop = vis(v)
			if stop {
				return
			}
		}
	}

	// vis FuncValue
	if fv := fr.Func; fv != nil {
		stop = vis(fv)
		if stop {
			return
		}
	}

	// vis defer
	for _, dfr := range fr.Defers {
		// visit dfr.Callable (the deferred func / bound method)
		if dfr.Callable != nil {
			stop = vis(dfr.Callable)
		}
		if stop {
			return
		}

		for _, arg := range dfr.Args {
			alloc.Recount(allocTypedValue)

			if arg.V != nil {
				stop = vis(arg.V)
			}
			if stop {
				return
			}
		}
	}

	// vis last package
	if fr.LastPackage != nil {
		stop = vis(fr.LastPackage)
	}
	if stop {
		return
	}

	return
}

func (e *Exception) Visit(alloc *Allocator, vis Visitor) (stop bool) {
	// vis value
	alloc.Recount(allocTypedValue)
	if v := e.Value.V; v != nil {
		stop = vis(v)
	}

	return
}
