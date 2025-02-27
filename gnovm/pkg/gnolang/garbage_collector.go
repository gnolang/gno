package gnolang

import (
	"fmt"
	"reflect"

	bm "github.com/gnolang/gno/gnovm/pkg/benchops"
	"github.com/gnolang/gno/tm2/pkg/overflow"
)

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
		gasCPU := overflow.Mul64p(m.Alloc.visitCount, GasFactorCPU)
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
	for i, frame := range m.Frames {
		// XXX implement for frames.
		// XXX Frame is not an object,
		// so implement a custom method and pass in vis.
		fmt.Printf("===visit Frame[%d] is: %v \n", i, frame)
		stop := frame.Visit(vis, m.Store)
		if stop {
			return -1, false
		}
	}

	// Visit package
	bm.StartGCCode(bm.VisitObject)
	stop := vis(m.Package)
	if stop {
		return -1, false
	}

	debug2.Println2("m.Exceptions: ", m.Exceptions)
	// Visit exceptions
	for i, exception := range m.Exceptions {
		// XXX implement for exceptions.
		// XXX Exception is not an object,
		// so implement a custom method and pass in vis.
		fmt.Printf("Exception[%d] is: %v \n", i, exception)
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
	var vis func(Object) bool // Declare `vis` first
	vis = func(o Object) bool {
		debug2.Printf2("===visit o: %v (type: %v) \n", o, reflect.TypeOf(o))
		debug2.Printf2("o.GetLastGCCycle: %d, GcCycle: %d\n", o.GetLastGCCycle(), gcCycle)
		// Return if already measured.
		if o.GetLastGCCycle() == gcCycle {
			bm.StopGCCode() // stop metric
			return false    // but don't stop
		}

		// Add object size to alloc.
		size := o.GetShallowSize()
		fmt.Println("shallow size: ", size)
		alloc.visitCount++ // count for gas calculation
		alloc.Allocate(size)
		// Stop if alloc max exceeded.
		maxBytes, curBytes := alloc.Status()
		if maxBytes < curBytes {
			bm.StopGCCode()
			return true
		}

		// stop metric for this visit.
		bm.StopGCCode()

		// Invote the traverser on o.
		stop := o.VisitAssociated(vis, alloc.m.Store)
		// Finally bump cycle.
		o.SetLastGCCycle(gcCycle)
		return stop
	}
	return vis
}

func (av *ArrayValue) VisitAssociated(vis Visitor, store Store) (stop bool) {
	debug2.Println2("VisitAssociated, av: ", av)
	// Visit each value.
	for i := 0; i < len(av.List); i++ {
		if bm.GCEnabled {
			bm.StartGCCode(bm.VisitObject)
		}

		v := av.List[i].V
		oo := unwrapObject(v, store)
		if oo == nil {
			if bm.GCEnabled {
				bm.StopGCCode()
			}
			continue
		}
		stop = vis(oo)
		if stop {
			return
		}
	}
	return
}

func (sv *StructValue) VisitAssociated(vis Visitor, store Store) (stop bool) {
	debug2.Println2("VisitAssociated, sv: ", sv)
	// Visit each value.
	for i := 0; i < len(sv.Fields); i++ {
		if bm.GCEnabled {
			bm.StartGCCode(bm.VisitObject)
		}
		v := sv.Fields[i].V
		oo := unwrapObject(v, store)
		if oo == nil {
			if bm.GCEnabled {
				bm.StopGCCode()
			}
			continue
		}
		stop = vis(oo)
		if stop {
			return
		}
	}
	return
}

func (bmv *BoundMethodValue) VisitAssociated(vis Visitor, store Store) (stop bool) {
	debug2.Println2("VisitAssociated, bmv: ", bmv)
	if bm.GCEnabled {
		bm.StartGCCode(bm.VisitObject)
	}
	// bmv.Func cannot be a closure, it must be a method.
	// So we do not visit it (for garbage collection).

	// Visit receiver.
	oo := unwrapObject(bmv.Receiver.V, store)
	if oo == nil {
		if bm.GCEnabled {
			bm.StopGCCode()
		}
		return
	}
	stop = vis(oo)
	if stop {
		return
	}
	return
}

func (mv *MapValue) VisitAssociated(vis Visitor, store Store) (stop bool) {
	debug2.Println2("VisitAssociated, mv: ", mv)
	// Visit values.
	// XXX visit mv.List.
	// XXX do NOT visit mv.vmap.
	for cur := mv.List.Head; cur != nil; cur = cur.Next {
		if bm.GCEnabled {
			bm.StartGCCode(bm.VisitObject)
		}
		// vis key
		oo := unwrapObject(cur.Key.V, store)
		if oo == nil {
			if bm.GCEnabled {
				bm.StopGCCode()
			}
			continue
		}
		stop = vis(oo)
		if stop {
			return
		}

		if bm.GCEnabled {
			bm.StartGCCode(bm.VisitObject)
		}
		// vis value
		oo = unwrapObject(cur.Value.V, store)
		if oo == nil {
			if bm.GCEnabled {
				bm.StopGCCode()
			}
			continue
		}
		stop = vis(oo)
		if stop {
			return
		}

	}
	return
}

func (pv *PackageValue) VisitAssociated(vis Visitor, store Store) (stop bool) {
	debug2.Println2("VisitAssociated, pv: ", pv)
	// XXX visit pv.Block
	// XXX visit pv.FBlocks
	// XXX do NOT visit Realm.

	if bm.GCEnabled {
		bm.StartGCCode(bm.VisitObject)
	}
	oo := unwrapObject(pv.Block, store)
	if oo == nil {
		if bm.GCEnabled {
			bm.StopGCCode()
		}
		return
	}
	stop = vis(oo)
	if stop {
		return
	}

	for _, fb := range pv.FBlocks {
		debug2.Printf2("fb: %v \n", fb)
		if bm.GCEnabled {
			bm.StartGCCode(bm.VisitObject)
		}
		oo := unwrapObject(fb, store)
		if oo == nil {
			if bm.GCEnabled {
				bm.StopGCCode()
			}
			continue
		}
		stop = vis(oo)
		if stop {
			return
		}
	}
	return
}

func (b *Block) VisitAssociated(vis Visitor, store Store) (stop bool) {
	debug2.Println2("VisitAssociated, block: ", b)
	// Visit each value.
	for i := 0; i < len(b.Values); i++ {
		if bm.GCEnabled {
			bm.StartGCCode(bm.VisitObject)
		}
		v := b.Values[i].V
		oo := unwrapObject(v, store)
		if oo == nil {
			if bm.GCEnabled {
				bm.StopGCCode()
			}
			continue
		}
		stop = vis(oo)
		if stop {
			return
		}
	}
	// Visit parent.
	if b.Parent != nil {
		debug2.Println2("visit parent block: ", b.Parent)
		if bm.GCEnabled {
			bm.StartGCCode(bm.VisitObject)
		}
		oo := unwrapObject(b.Parent, store)
		if oo == (*Block)(nil) {
			if bm.GCEnabled {
				bm.StopGCCode()
			}
			return
		}
		stop = vis(oo)
		if stop {
			return
		}
	}
	return
}

func (hiv *HeapItemValue) VisitAssociated(vis Visitor, store Store) (stop bool) {
	debug2.Println2("VisitAssociated, hiv: ", hiv)
	if bm.GCEnabled {
		bm.StartGCCode(bm.VisitObject)
	}
	oo := unwrapObject(hiv.Value.V, store)
	if oo == nil {
		if bm.GCEnabled {
			bm.StopGCCode()
		}
		return
	}
	stop = vis(oo)
	if stop {
		return
	}
	return
}

func (fv *FuncValue) Visit(vis Visitor, store Store) (stop bool) {
	// visit captures
	for _, tv := range fv.Captures {
		if bm.GCEnabled {
			bm.StartGCCode(bm.VisitObject)
		}
		oo := unwrapObject(tv.V, store)
		debug2.Println2("vis capture, oo: ", oo)
		if oo == nil {
			if bm.GCEnabled {
				bm.StopGCCode()
			}
			continue
		} else {
			stop = vis(oo)
			if stop {
				return
			}
		}
	}
	// visit FuncValue's closure
	oo := unwrapObject(fv.Closure, store)
	debug2.Println2("vis Closure, oo: ", oo, reflect.TypeOf(oo))

	if oo == (*Block)(nil) {
		if bm.GCEnabled {
			bm.StopGCCode()
			return
		}
	} else {
		stop = vis(oo)
		if stop {
			return
		}
	}
	return false
}

func (fr *Frame) Visit(vis Visitor, store Store) (stop bool) {
	// vis receiver
	// TODO: how about receiver define in other file
	// also check gcCount...

	if bm.GCEnabled {
		bm.StartGCCode(bm.VisitObject)
	}
	oo := unwrapObject(fr.Receiver.V, store)
	debug2.Println2("vis receiver oo: ", oo)
	if oo == nil {
		if bm.GCEnabled {
			bm.StopGCCode()
		}
		return
	} else {
		stop = vis(oo)
		if stop {
			return
		}
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
		debug2.Println2("vis defer: ", dfr)
		// visit dfr.Func
		stop = dfr.Func.Visit(vis, store)
		if stop {
			return
		}

		for _, arg := range dfr.Args {
			debug2.Println2("vis arg: ", arg)
			if bm.GCEnabled {
				bm.StartGCCode(bm.VisitObject)
			}
			oo = unwrapObject(arg.V, store)
			if oo == nil {
				if bm.GCEnabled {
					bm.StopGCCode()
				}
				return
			} else {
				stop = vis(oo)
				if stop {
					return
				}
			}
		}
	}
	return false
}

func (ex *Exception) Visit(vis Visitor, store Store) (stop bool) {
	debug2.Println2("vis exception: ", ex)
	if bm.GCEnabled {
		bm.StartGCCode(bm.VisitObject)
	}
	// vis value
	oo := unwrapObject(ex.Value.V, store)
	if oo == nil {
		if bm.GCEnabled {
			bm.StopGCCode()
		}
		return
	} else {
		stop = vis(oo)
		if stop {
			return
		}
	}
	// Max, this should be unnecessary
	stop = ex.Frame.Visit(vis, store)
	if stop {
		return
	}
	return
}

func unwrapObject(v Value, store Store) Object {
	//debug2.Println2("unwrapReference, v: ", v, reflect.TypeOf(v))
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
