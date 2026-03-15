package gnolang

import (
	"math/bits"
	"reflect"

	"github.com/gnolang/gno/tm2/pkg/overflow"
)

// gcVisitGasTable[k] = gas per GC visit when log2(visitCount) == k.
// Per-visit cost increases with heap size due to CPU cache effects:
// small heaps fit in L2/L3 (~30ns/visit), large heaps hit DRAM (~700ns/visit).
//
// Calibrated from BenchmarkGCVisit on DigitalOcean Dedicated (2-core),
// Intel Xeon Platinum 8168 @ 2.70GHz, cpuBaseNs = 5.2 ns/gas.
//
// See gnovm/pkg/gnolang/bench_gc_test.go for benchmarks.
var gcVisitGasTable = [25]int64{
	6, 6, 6, 6, 6, 6, 6, // 2^0 - 2^6:   1-64 visits       (~29ns, L1/L2)
	8, 8, 8,              // 2^7 - 2^9:   128-512 visits     (~40ns, L2/L3)
	18, 18, 18,           // 2^10 - 2^12: 1K-4K visits       (~91ns, L3)
	30, 30, 30, 38,       // 2^13 - 2^16: 8K-64K visits      (~160-197ns, L3/DRAM)
	55, 55, 73,           // 2^17 - 2^19: 128K-512K visits   (~290-380ns, DRAM)
	73, 73, 100,          // 2^20 - 2^22: 1M-4M visits       (~380-520ns, DRAM)
	135, 135,             // 2^23 - 2^24: 8M-16M visits      (~700ns, DRAM+TLB)
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

	defer func() {
		gasCPU := gcVisitGas(visitCount)
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

	// Return bytes remaining.
	maxBytes, bytes := m.Alloc.Status()
	return maxBytes - bytes, true
}

// Returns a visitor that bumps the GCCycle counter
// and stops if alloc is out of memory.
func GCVisitorFn(gcCycle int64, alloc *Allocator, visitCount *int64) Visitor {
	var vis func(value Value) bool

	vis = func(v Value) bool {
		if debug {
			debug.Printf("Visit, v: %v (type: %v)\n", v, reflect.TypeOf(v))
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
		// visit dfr.Func
		if dfr.Func != nil {
			stop = vis(dfr.Func)
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

		if dfr.Parent != nil {
			stop = vis(dfr.Parent)
		}
		if stop {
			return
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
