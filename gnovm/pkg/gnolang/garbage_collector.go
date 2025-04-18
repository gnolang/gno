package gnolang

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
	// We don't need the old value anymore.
	m.Alloc.Reset()

	// This is the only place where it's bumped.
	m.gcCycle += 1

	// Construct visitor callback.
	vis := GCVisitorFn(m.gcCycle, m.Alloc)

	// Visit blocks
	for block := range m.Blocks {
		stop := vis(block)
		if stop {
			return -1, false
		}
	}

	// Visit frames
	for frame := range m.Frames {
		// XXX implement for frames.
		// XXX Frame is not an object,
		// so implement a custom method and pass in vis.
	}

	// Visit package
	stop := vis(m.Package)
	if stop {
		return -1, false
	}

	// Visit exceptions
	for exception := range m.Exceptions {
		// XXX implement for exceptions.
		// XXX Exception is not an object,
		// so implement a custom method and pass in vis.
	}

	// Return bytes remaining.
	maxBytes, bytes := m.Alloc.Status()
	return maxBytes - bytes, true
}

// Returns a visitor that bumps the gcCycle counter
// and stops if alloc is out of memory.
func GCVisitorFn(gcCycle int64, alloc *Alloc) Visitor {
	vis := func(o Object) bool {
		// Return if already measured.
		if o.GetLastGCCycle() == gcCycle {
			return false // but don't stop
		}
		// Add object size to alloc.
		size := o.GetShallowSize()
		alloc.Allocate(size)
		// Stop if alloc max exceeded.
		maxBytes, curBytes := alloc.Status()
		if maxBytes < curBytes {
			return true
		}
		// Invote the traverser on o.
		stop := o.VisitAssociated(vis)
		// Finally bump cycle.
		o.SetLastGCCycle(gcCycle)
		return stop
	}
	return tr
}

func (av *ArrayValue) VisitAssociated(vis Visitor) (stop bool) {
	// Visit each value.
	for i := 0; i < len(av.List); i++ {
		v := av.List[i].Value
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

func (sv *StructValue) VisitAssociated(vis Visitor) (stop bool) {
	// Visit each value.
	for i := 0; i < len(sv.Fields); i++ {
		v := sv.Fields[i].Value
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
	// bmv.Func cannot be a closure, it must be a method.
	// So we do not visit it (for garbage collection).

	// Visit receiver.
	stop = vis(bmv.Receiver.V)
	return
}

func (mv *MapValue) VisitAssociated(vis Visitor) (stop bool) {
	// Visit values.
	// XXX visit mv.List.
	// XXX do NOT visit mv.vmap.
	return
}

func (pv *PackageValue) VisitAssociated(vis Visitor) (stop bool) {
	// XXX visit pv.Block
	// XXX visit pv.FBlocks
	// XXX do NOT visit Realm.
}

func (b *Block) VisitAssociated(vis Visitor) (stop bool) {
	// Visit each value.
	for i := 0; i < len(b.Values); i++ {
		v := b.Values[i].Value
		if v == nil {
			continue
		}
		stop = vis(v)
		if stop {
			return
		}
	}
	// Visit parent.
	if b.Parent != nil {
		stop = vis(b.Parent)
		if stop {
			return
		}
	}
}

func (hiv *HeapItemValue) VisitAssociated(vis Visitor) (stop bool) {
	stop = vis(hiv.Value.V)
	return
}
