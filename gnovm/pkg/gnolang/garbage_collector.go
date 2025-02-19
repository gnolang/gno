package gnolang

import (
	"reflect"

	"github.com/gnolang/gno/tm2/pkg/overflow"
)

// Represents the "time unit" cost for
// a single garbage collection visit.
// It's similar to "CPU cycles" and is
// calculated based on a rough benchmarking
// results.
// TODO: more accurate benchmark.
const VisitCpuFactor = 8

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
func (m *Machine) GarbageCollect() (left int64, ok bool) {
	defer func() {
		gasCPU := overflow.Mul64p(m.Alloc.visitCount*VisitCpuFactor, GasFactorCPU)
		m.Alloc.visitCount = 0
		if m.GasMeter != nil {
			m.GasMeter.ConsumeGas(gasCPU, "GC")
		}
	}()

	// We don't need the old value anymore.
	m.Alloc.Reset()

	// This is the only place where it's bumped.
	m.GCCycle += 1

	// Construct visitor callback.
	vis := GCVisitorFn(m.GCCycle, m.Alloc)

	// Visit blocks
	for _, block := range m.Blocks {
		stop := vis(block)
		if stop {
			return -1, false
		}
	}

	// Visit frames
	for _, frame := range m.Frames {
		stop := frame.Visit(vis)
		if stop {
			return -1, false
		}
	}

	// Visit package
	stop := vis(m.Package)
	if stop {
		return -1, false
	}

	// Visit exceptions
	for _, exception := range m.Exceptions {
		stop = exception.Visit(vis)
		if stop {
			return -1, false
		}
	}

	// Return bytes remaining.
	maxBytes, bytes := m.Alloc.Status()
	return maxBytes - bytes, true
}

// Returns a visitor that bumps the GCCycle counter
// and stops if alloc is out of memory.
func GCVisitorFn(gcCycle int64, alloc *Allocator) Visitor {
	var vis func(value Value) bool

	vis = func(v Value) bool {
		if v == nil || v == (*Block)(nil) || v == (*PackageValue)(nil) {
			return false
		}

		if debug {
			debug.Printf("Visit, v: %v (type: %v)\n", v, reflect.TypeOf(v))
		}

		oo, isObject := v.(Object)

		if isObject {
			defer func() {
				// Finally bump cycle for object.
				oo.SetLastGCCycle(gcCycle)
			}()

			// Return if already measured.
			if debug {
				debug.Printf("oo.GetLastGCCycle: %d, gcCycle: %d\n", oo.GetLastGCCycle(), gcCycle)
			}
			if oo.GetLastGCCycle() == gcCycle {
				return false // but don't stop
			}
		}

		alloc.visitCount++ // Count operations for gas calculation

		// Add object size to alloc.
		size := v.GetShallowSize()
		alloc.Allocate(size)

		// Stop if alloc max exceeded.
		// NOTE: Unlikely to occur, but keep it here for
		// now to handle potential edge cases.
		// Consider removing it later if no issues arise.
		maxBytes, curBytes := alloc.Status()
		if maxBytes < curBytes {
			return true
		}

		// Invoke the traverser on v.
		stop := v.VisitAssociated(vis)

		return stop
	}
	return vis
}

// ---------------------------------------------------------------
// visit associated

func (sv *SliceValue) VisitAssociated(vis Visitor) (stop bool) {
	// Visit base.
	stop = vis(sv.Base)
	return stop
}

func (av *ArrayValue) VisitAssociated(vis Visitor) (stop bool) {
	// Visit each value.
	for i := 0; i < len(av.List); i++ {
		stop = vis(av.List[i].V)
		if stop {
			return
		}
	}
	return
}

func (fv *FuncValue) VisitAssociated(vis Visitor) (stop bool) {
	// visit captures
	for _, tv := range fv.Captures {
		stop = vis(tv.V)
		if stop {
			return
		}
	}

	return
}

func (sv *StructValue) VisitAssociated(vis Visitor) (stop bool) {
	// Visit each value.
	for i := 0; i < len(sv.Fields); i++ {
		stop = vis(sv.Fields[i].V)
		if stop {
			return
		}
	}
	return
}

func (bmv *BoundMethodValue) VisitAssociated(vis Visitor) (stop bool) {
	// bmv.Func cannot be a closure, it must be a method.
	// So we do not visit it (for garbage collection).

	// Visit receiver.
	stop = vis(bmv.Receiver.V)
	return
}

func (mv *MapValue) VisitAssociated(vis Visitor) (stop bool) {
	// visit mv.List.
	for cur := mv.List.Head; cur != nil; cur = cur.Next {
		// vis key
		stop = vis(cur.Key.V)
		if stop {
			return
		}

		// vis value
		stop = vis(cur.Value.V)
		if stop {
			return
		}
	}
	return
}

func (pv *PackageValue) VisitAssociated(vis Visitor) (stop bool) {
	// visit pv.Block
	stop = vis(pv.Block)
	if stop {
		return
	}

	// visit pv.FBlocks
	for _, fb := range pv.FBlocks {
		stop = vis(fb)
		if stop {
			return
		}
	}

	// do NOT visit Realm.

	return
}

func (b *Block) VisitAssociated(vis Visitor) (stop bool) {
	// Visit each value.
	for i := 0; i < len(b.Values); i++ {
		stop = vis(b.Values[i].V)
		if stop {
			return
		}
	}

	// Visit parent.
	stop = vis(b.Parent)
	return
}

func (hiv *HeapItemValue) VisitAssociated(vis Visitor) (stop bool) {
	stop = vis(hiv.Value.V)
	return
}

func (pv PointerValue) VisitAssociated(vis Visitor) (stop bool) {
	// NOTE: *TV and Key will be visited along with base.
	stop = vis(pv.Base)
	return
}

func (sv StringValue) VisitAssociated(vis Visitor) (stop bool) {
	return false
}

func (bv BigintValue) VisitAssociated(vis Visitor) (stop bool) {
	return false
}

func (bv BigdecValue) VisitAssociated(vis Visitor) (stop bool) {
	return false
}

func (dbv DataByteValue) VisitAssociated(vis Visitor) (stop bool) {
	return false
}

func (nv *NativeValue) VisitAssociated(vis Visitor) (stop bool) {
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
// custom visit methods

func (fr *Frame) Visit(vis Visitor) (stop bool) {
	// vis receiver
	stop = vis(fr.Receiver.V)
	if stop {
		return
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
		stop = vis(dfr.Func)
		if stop {
			return
		}

		for _, arg := range dfr.Args {
			stop = vis(arg.V)
			if stop {
				return
			}
		}

		stop = vis(dfr.Parent)
		if stop {
			return
		}
	}

	// vis last package
	stop = vis(fr.LastPackage)
	if stop {
		return
	}

	return
}

func (ex *Exception) Visit(vis Visitor) (stop bool) {
	// vis value
	stop = vis(ex.Value.V)
	if stop {
		return
	}

	// The frame should have been visited elsewhere.
	// This ensures integrity and improves readability.
	stop = ex.Frame.Visit(vis)
	return
}
