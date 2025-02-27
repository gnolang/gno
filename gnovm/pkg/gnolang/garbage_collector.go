package gnolang

import (
	"reflect"

	bm "github.com/gnolang/gno/gnovm/pkg/benchops"
	"github.com/gnolang/gno/tm2/pkg/overflow"
)

const VisitCpuFactor = 8 // calculated based on benchmark

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

	debug2.Println2("=====GarbageCollect")
	defer func() {
		gasCPU := overflow.Mul64p(m.Alloc.visitCount*VisitCpuFactor, GasFactorCPU)
		debug2.Println2("gasCPU:", gasCPU)
		if m.GasMeter != nil { //  no gas meter for test
			m.GasMeter.ConsumeGas(gasCPU, "GC")
		}
	}()

	// We don't need the old value anymore.
	m.Alloc.Reset()

	// This is the only place where it's bumped.
	m.GcCycle += 1

	debug2.Println2("m.GcCycle: ", m.GcCycle)

	// Construct visitor callback.
	vis := GCVisitorFn(m.GcCycle, m.Alloc)

	// Visit blocks
	for _, block := range m.Blocks {
		if bm.GCEnabled {
			bm.StartGCCode(bm.VisitObject)
		}
		stop := vis(block)
		if stop {
			return -1, false
		}
	}

	// Visit frames
	for _, frame := range m.Frames {
		// XXX implement for frames.
		// XXX Frame is not an object,
		// so implement a custom method and pass in vis.
		if bm.GCEnabled {
			bm.StartGCCode(bm.VisitObject)
		}
		stop := frame.Visit(vis, m.Store)
		if stop {
			return -1, false
		}
	}

	// Visit package
	if bm.GCEnabled {
		bm.StartGCCode(bm.VisitObject)
	}
	stop := vis(m.Package)
	if stop {
		return -1, false
	}

	// Visit exceptions
	for _, exception := range m.Exceptions {
		if bm.GCEnabled {
			bm.StartGCCode(bm.VisitObject)
		}
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
		if o == nil || o == (*Block)(nil) {
			if bm.GCEnabled {
				bm.StopGCCode()
			}
			return false
		}

		debug2.Printf2("===visit o: %v (type: %v) \n", o, reflect.TypeOf(o))
		debug2.Printf2("o.GetLastGCCycle: %d, GcCycle: %d\n", o.GetLastGCCycle(), gcCycle)

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

		// stop metric for this visit.
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
		if bm.GCEnabled {
			bm.StartGCCode(bm.VisitObject)
		}
		v := av.List[i].V
		oo := unwrapObject(v, store)
		stop = vis(oo)
		if stop {
			return
		}
	}
	return
}

func (sv *StructValue) VisitAssociated(vis Visitor, store Store) (stop bool) {
	// Visit each value.
	for i := 0; i < len(sv.Fields); i++ {
		if bm.GCEnabled {
			bm.StartGCCode(bm.VisitObject)
		}
		v := sv.Fields[i].V
		oo := unwrapObject(v, store)
		stop = vis(oo)
		if stop {
			return
		}
	}
	return
}

func (bmv *BoundMethodValue) VisitAssociated(vis Visitor, store Store) (stop bool) {
	if bm.GCEnabled {
		bm.StartGCCode(bm.VisitObject)
	}
	// bmv.Func cannot be a closure, it must be a method.
	// So we do not visit it (for garbage collection).

	// Visit receiver.
	oo := unwrapObject(bmv.Receiver.V, store)
	stop = vis(oo)
	return
}

func (mv *MapValue) VisitAssociated(vis Visitor, store Store) (stop bool) {
	// Visit values.
	// XXX visit mv.List.
	// XXX do NOT visit mv.vmap.
	for cur := mv.List.Head; cur != nil; cur = cur.Next {
		if bm.GCEnabled {
			bm.StartGCCode(bm.VisitObject)
		}
		// vis key
		oo := unwrapObject(cur.Key.V, store)
		stop = vis(oo)
		if stop {
			return
		}

		if bm.GCEnabled {
			bm.StartGCCode(bm.VisitObject)
		}
		// vis value
		oo = unwrapObject(cur.Value.V, store)
		stop = vis(oo)
		if stop {
			return
		}

	}
	return
}

func (pv *PackageValue) VisitAssociated(vis Visitor, store Store) (stop bool) {
	// XXX visit pv.Block
	// XXX visit pv.FBlocks
	// XXX do NOT visit Realm.

	if bm.GCEnabled {
		bm.StartGCCode(bm.VisitObject)
	}
	oo := unwrapObject(pv.Block, store)
	stop = vis(oo)
	if stop {
		return
	}

	for _, fb := range pv.FBlocks {
		if bm.GCEnabled {
			bm.StartGCCode(bm.VisitObject)
		}
		oo := unwrapObject(fb, store)
		stop = vis(oo)
		if stop {
			return
		}
	}
	return
}

func (b *Block) VisitAssociated(vis Visitor, store Store) (stop bool) {
	// Visit each value.
	for i := 0; i < len(b.Values); i++ {
		if bm.GCEnabled {
			bm.StartGCCode(bm.VisitObject)
		}
		v := b.Values[i].V
		oo := unwrapObject(v, store)
		stop = vis(oo)
		if stop {
			return
		}
	}
	// Visit parent.
	if b.Parent != nil {
		if bm.GCEnabled {
			bm.StartGCCode(bm.VisitObject)
		}
		oo := unwrapObject(b.Parent, store)
		stop = vis(oo)
		return
	}
	return
}

func (hiv *HeapItemValue) VisitAssociated(vis Visitor, store Store) (stop bool) {
	if bm.GCEnabled {
		bm.StartGCCode(bm.VisitObject)
	}
	oo := unwrapObject(hiv.Value.V, store)
	stop = vis(oo)
	return
}

func (fv *FuncValue) Visit(vis Visitor, store Store) (stop bool) {
	// visit captures
	for _, tv := range fv.Captures {
		if bm.GCEnabled {
			bm.StartGCCode(bm.VisitObject)
		}
		oo := unwrapObject(tv.V, store)
		stop = vis(oo)
		if stop {
			return
		}
	}

	// visit FuncValue's closure
	if bm.GCEnabled {
		bm.StartGCCode(bm.VisitObject)
	}
	oo := unwrapObject(fv.Closure, store)

	stop = vis(oo)
	return
}

func (fr *Frame) Visit(vis Visitor, store Store) (stop bool) {
	// vis receiver
	oo := unwrapObject(fr.Receiver.V, store)
	stop = vis(oo)
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
			if bm.GCEnabled {
				bm.StartGCCode(bm.VisitObject)
			}
			oo = unwrapObject(arg.V, store)
			stop = vis(oo)
			if stop {
				return
			}
		}
	}
	return
}

func (ex *Exception) Visit(vis Visitor, store Store) (stop bool) {
	// vis value
	oo := unwrapObject(ex.Value.V, store)
	stop = vis(oo)
	if stop {
		return
	}

	// Max, this should be unnecessary
	if bm.GCEnabled {
		bm.StartGCCode(bm.VisitObject)
	}
	stop = ex.Frame.Visit(vis, store)
	return
}

func unwrapObject(v Value, store Store) Object {
	switch v := v.(type) {
	case *SliceValue:
		return v.GetBase(store)
	case PointerValue:
		return v.GetBase(store)
	case RefValue:
		oo := store.GetObject(v.ObjectID)
		return oo
	}

	if o, ok := v.(Object); ok {
		return o
	}

	return nil
}
