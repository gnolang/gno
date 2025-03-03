package gnolang

import (
	bm "github.com/gnolang/gno/gnovm/pkg/benchops"
	"github.com/gnolang/gno/tm2/pkg/overflow"
)

// Represents the "time unit" cost for
// a single garbage collection visit.
// It's similar to "CPU cycles" and is
// calculated based on benchmarking results.
const VisitCpuFactor = 8

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
	if bm.GCEnabled {
		defer func() {
			bm.FinishGC()
		}()
	}

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
	m.GcCycle += 1

	// Construct visitor callback.
	vis := GCVisitorFn(m.GcCycle, m.Alloc)

	// Visit blocks
	for _, block := range m.Blocks {
		stop := vis(block)
		if stop {
			return -1, false
		}
	}

	// Visit frames
	for _, frame := range m.Frames {
		stop := frame.Visit(vis, m.Store)
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
		// XXX implement for exceptions.
		// XXX Exception is not an object,
		// so implement a custom method and pass in vis.
		stop = exception.Visit(vis, m.Store)
		if stop {
			return -1, false
		}
	}

	// Return bytes remaining.
	maxBytes, bytes := m.Alloc.Status()
	return maxBytes - bytes, true
}

// Returns a visitor that bumps the GcCycle counter
// and stops if alloc is out of memory.
func GCVisitorFn(gcCycle int64, alloc *Allocator) Visitor {
	var vis func(Object) bool
	vis = func(o Object) bool {
		if bm.GCEnabled {
			// The metric may either be initialized at this stage
			// or during the previous unwrap step, depending on the flow.
			if !bm.IsGCMeasureStarted() {
				bm.StartGCCode(bm.VisitObject)
			}
		}

		// Return if already measured.
		if o.GetLastGCCycle() == gcCycle {
			if bm.GCEnabled {
				bm.StopGCCode()
			}
			return false // but don't stop
		}

		// Add object size to alloc.
		size := o.GetShallowSize()
		alloc.visitCount++ // count for gas calculation

		alloc.Allocate(size)
		// Stop if alloc max exceeded.
		maxBytes, curBytes := alloc.Status()
		if maxBytes < curBytes {
			if bm.GCEnabled {
				bm.StopGCCode()
			}
			return true
		}

		// stop metric before next visit.
		if bm.GCEnabled {
			bm.StopGCCode()
		}
		// Invote the traverser on o.
		stop := o.VisitAssociated(vis, alloc.m.Store)
		// Finally bump cycle.
		o.SetLastGCCycle(gcCycle)
		return stop
	}
	return vis
}

// ---------------------------------------------------------------
// visit associated

func (av *ArrayValue) VisitAssociated(vis Visitor, store Store) (stop bool) {
	// Visit each value.
	for i := 0; i < len(av.List); i++ {
		stop = visitChild(vis, av.List[i].V, store)
		if stop {
			return
		}
	}
	return
}

func (sv *StructValue) VisitAssociated(vis Visitor, store Store) (stop bool) {
	// Visit each value.
	for i := 0; i < len(sv.Fields); i++ {
		stop = visitChild(vis, sv.Fields[i].V, store)
		if stop {
			return
		}
	}
	return
}

func (bmv *BoundMethodValue) VisitAssociated(vis Visitor, store Store) (stop bool) {
	// bmv.Func cannot be a closure, it must be a method.
	// So we do not visit it (for garbage collection).

	// Visit receiver.
	stop = visitChild(vis, bmv.Receiver.V, store)
	return
}

func (mv *MapValue) VisitAssociated(vis Visitor, store Store) (stop bool) {
	// Visit values.
	// visit mv.List.
	for cur := mv.List.Head; cur != nil; cur = cur.Next {
		// vis key
		stop = visitChild(vis, cur.Key.V, store)
		if stop {
			return
		}

		// vis value
		stop = visitChild(vis, cur.Value.V, store)
		if stop {
			return
		}
	}
	return
}

func (pv *PackageValue) VisitAssociated(vis Visitor, store Store) (stop bool) {
	// visit pv.Block
	stop = visitChild(vis, pv.Block, store)
	if stop {
		return
	}

	// visit pv.FBlocks
	for _, fb := range pv.FBlocks {
		stop = visitChild(vis, fb, store)
		if stop {
			return
		}
	}

	// do NOT visit Realm.

	return
}

func (b *Block) VisitAssociated(vis Visitor, store Store) (stop bool) {
	// Visit each value.
	for i := 0; i < len(b.Values); i++ {
		stop = visitChild(vis, b.Values[i].V, store)
		if stop {
			return
		}
	}

	// Visit parent.
	stop = visitChild(vis, b.Parent, store)
	return
}

func (hiv *HeapItemValue) VisitAssociated(vis Visitor, store Store) (stop bool) {
	stop = visitChild(vis, hiv.Value.V, store)
	return
}

func (fv *FuncValue) Visit(vis Visitor, store Store) (stop bool) {
	// visit captures
	for _, tv := range fv.Captures {
		stop = visitChild(vis, tv.V, store)
		if stop {
			return
		}
	}

	// visit FuncValue's closure
	stop = visitChild(vis, fv.Closure, store)
	return
}

func (fr *Frame) Visit(vis Visitor, store Store) (stop bool) {
	// vis receiver
	stop = visitChild(vis, fr.Receiver.V, store)
	if stop {
		return
	}

	// vis FuncValue
	if fv := fr.Func; fv != nil {
		stop = fv.Visit(vis, store)
		if stop {
			return
		}
	}

	// vis defer
	for _, dfr := range fr.Defers {
		// visit dfr.Func
		stop = dfr.Func.Visit(vis, store)
		if stop {
			return
		}

		for _, arg := range dfr.Args {
			stop = visitChild(vis, arg.V, store)
			if stop {
				return
			}
		}

		stop = visitChild(vis, dfr.Parent, store)
		if stop {
			return
		}
	}

	// vis last package
	stop = visitChild(vis, fr.LastPackage, store)
	if stop {
		return
	}

	return
}

func (ex *Exception) Visit(vis Visitor, store Store) (stop bool) {
	// vis value
	stop = visitChild(vis, ex.Value.V, store)
	if stop {
		return
	}

	// the frame should've been
	// visited in other places
	// this ensures integrity.
	stop = ex.Frame.Visit(vis, store)
	return
}

func visitChild(vis Visitor, v Value, store Store) (stop bool) {
	// start metric before unwrap,
	// and ends metric in visit
	if bm.GCEnabled {
		bm.StartGCCode(bm.VisitObject)
	}
	var oo Object
	switch v := v.(type) {
	case *SliceValue:
		oo = v.GetBase(store)
	case PointerValue:
		oo = v.GetBase(store)
	case RefValue:
		oo = store.GetObject(v.ObjectID)
	case Object:
		oo = v
	}

	if oo == nil || oo == (*Block)(nil) || oo == (*PackageValue)(nil) {
		if bm.GCEnabled {
			bm.StopGCCode()
		}
		return false
	}

	stop = vis(oo)
	return
}
