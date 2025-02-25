package gnolang

import (
	"fmt"
	"reflect"
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
	debug2.Println2("=====GarbageCollect")
	debug2.Println2("m.Exceptions: ", m.Exceptions)
	// We don't need the old value anymore.
	m.Alloc.Reset()

	// This is the only place where it's bumped.
	m.GcCycle += 1

	debug2.Println2("m.GcCycle: ", m.GcCycle)

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
	stop := vis(m.Package)
	if stop {
		return -1, false
	}

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
		if o == nil {
			return false
		}
		if o == (*Block)(nil) {
			debug2.Println2("nil block: ", o) // XXX ???
			return false                      // stop
		}
		debug2.Printf2("o.GetLastGCCycle: %d, GcCycle: %d\n", o.GetLastGCCycle(), gcCycle)
		// Return if already measured.
		if o.GetLastGCCycle() == gcCycle {
			return false // but don't stop
		}

		// Add object size to alloc.
		size := o.GetShallowSize()
		fmt.Println("shallow size: ", size)
		alloc.Allocate(size)
		// Stop if alloc max exceeded.
		maxBytes, curBytes := alloc.Status()
		if maxBytes < curBytes {
			return true
		}
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
		v := av.List[i].V
		oo := unwrapObject(v, store)
		if oo == nil {
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
		v := sv.Fields[i].V
		oo := unwrapObject(v, store)
		if oo == nil {
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
	// bmv.Func cannot be a closure, it must be a method.
	// So we do not visit it (for garbage collection).

	// Visit receiver.
	oo := unwrapObject(bmv.Receiver.V, store)
	if oo == nil {
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
		// vis key
		oo := unwrapObject(cur.Key.V, store)
		if oo == nil {
			continue
		}
		stop = vis(oo)
		if stop {
			return
		}

		// vis value
		oo = unwrapObject(cur.Value.V, store)
		if oo == nil {
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
	oo := unwrapObject(pv.Block, store)
	if oo == nil {
		return
	}
	stop = vis(oo)
	if stop {
		return
	}

	for _, fb := range pv.FBlocks {
		debug2.Printf2("fb: %v \n", fb)
		oo := unwrapObject(fb, store)
		if oo == nil {
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
	debug2.Println2("len of values in block: ", len(b.Values))
	for i := 0; i < len(b.Values); i++ {
		v := b.Values[i].V
		oo := unwrapObject(v, store)
		if oo == nil {
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
		oo := unwrapObject(b.Parent, store)
		if oo == nil {
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
	oo := unwrapObject(hiv.Value.V, store)
	if oo == nil {
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
		oo := unwrapObject(tv.V, store)
		debug2.Println2("vis capture, oo: ", oo)
		if oo != nil {
			stop = vis(oo)
			if stop {
				return
			}
		}
	}
	// visit FuncValue's closure
	oo := unwrapObject(fv.Closure, store)
	debug2.Println2("vis Closure, oo: ", oo)
	if oo != nil {
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
	oo := unwrapObject(fr.Receiver.V, store)
	debug2.Println2("vis receiver oo: ", oo)
	if oo != nil {
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
			oo = unwrapObject(arg.V, store)
			if oo != nil {
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
	// vis value
	oo := unwrapObject(ex.Value.V, store)
	if oo != nil {
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
