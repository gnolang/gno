package gnolang

import (
	"fmt"
	"github.com/gnolang/gno/tm2/pkg/overflow"
)

// Represents the "time unit" cost for
// a single garbage collection visit.
// It's similar to "CPU cycles" and is
// calculated based on benchmarking results.
const VisitCpuFactor = 8

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
	fmt.Println("===GarbageCollect===")
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
	var vis func(value Value) bool

	vis = func(v Value) bool {
		switch v.(type) {
		case *SliceValue, *ArrayValue, *MapValue, *StructValue, *Block, *FuncValue,
			*BoundMethodValue, *HeapItemValue, *PackageValue, PointerValue:
		default:
			return false
		}
	
		// Return if already measured.
		if v.GetLastGCCycle() == gcCycle {
			return false // but don't stop
		}

		// Add object size to alloc.
		size := v.GetShallowSize()
		alloc.visitCount++ // count for gas calculation

		alloc.Allocate(size)
		// Stop if alloc max exceeded.
		maxBytes, curBytes := alloc.Status()
		if maxBytes < curBytes {
			return true
		}

		// Invote the traverser on o.
		stop := v.VisitAssociated(vis)
		// Finally bump cycle.
		v.SetLastGCCycle(gcCycle)
		return stop
	}
	return vis
}

// ---------------------------------------------------------------
// visit associated

func (sv *SliceValue) VisitAssociated(vis Visitor) (stop bool) {
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

	// visit FuncValue's closure
	stop = vis(fv.Closure)
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
	// Visit values.
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
func (tv TypeValue) VisitAssociated(vis Visitor) (stop bool) {
	return false
}

// -------------------------------------------------------------------
// custom visit methods

func (fr *Frame) Visit(vis Visitor, store Store) (stop bool) {
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

func (ex *Exception) Visit(vis Visitor, store Store) (stop bool) {
	// vis value
	stop = vis(ex.Value.V)
	if stop {
		return
	}

	// the frame should've been
	// visited in other places
	// this ensures integrity.
	stop = ex.Frame.Visit(vis, store)
	return
}

// -------------------------------------------------------------
// gc cycle

func (sv StringValue) GetLastGCCycle() int64        { return sv.lastGCCycle }
func (bv BigintValue) GetLastGCCycle() int64        { return bv.lastGCCycle }
func (bv BigdecValue) GetLastGCCycle() int64        { return bv.lastGCCycle }
func (dbv DataByteValue) GetLastGCCycle() int64     { return dbv.lastGCCycle }
func (pv PointerValue) GetLastGCCycle() int64       { return pv.lastGCCycle }
func (av *ArrayValue) GetLastGCCycle() int64        { return av.lastGCCycle }
func (sv *SliceValue) GetLastGCCycle() int64        { return sv.lastGCCycle }
func (sv *StructValue) GetLastGCCycle() int64       { return sv.lastGCCycle }
func (fv *FuncValue) GetLastGCCycle() int64         { return fv.lastGCCycle }
func (mv *MapValue) GetLastGCCycle() int64          { return mv.lastGCCycle }
func (bmv *BoundMethodValue) GetLastGCCycle() int64 { return bmv.lastGCCycle }
func (pv *PackageValue) GetLastGCCycle() int64      { return pv.lastGCCycle }
func (nv *NativeValue) GetLastGCCycle() int64       { return nv.lastGCCycle }
func (b *Block) GetLastGCCycle() int64              { return b.lastGCCycle }
func (rv RefValue) GetLastGCCycle() int64           { return rv.lastGCCycle }
func (hiv *HeapItemValue) GetLastGCCycle() int64    { return hiv.lastGCCycle }
func (tv TypeValue) GetLastGCCycle() int64          { return tv.lastGCCycle }

func (sv StringValue) SetLastGCCycle(cycle int64)        { sv.lastGCCycle = cycle }
func (bv BigintValue) SetLastGCCycle(cycle int64)        { bv.lastGCCycle = cycle }
func (bv BigdecValue) SetLastGCCycle(cycle int64)        { bv.lastGCCycle = cycle }
func (dbv DataByteValue) SetLastGCCycle(cycle int64)     { dbv.lastGCCycle = cycle }
func (pv PointerValue) SetLastGCCycle(cycle int64)       { pv.lastGCCycle = cycle }
func (av *ArrayValue) SetLastGCCycle(cycle int64)        { av.lastGCCycle = cycle }
func (sv *SliceValue) SetLastGCCycle(cycle int64)        { sv.lastGCCycle = cycle }
func (sv *StructValue) SetLastGCCycle(cycle int64)       { sv.lastGCCycle = cycle }
func (fv *FuncValue) SetLastGCCycle(cycle int64)         { fv.lastGCCycle = cycle }
func (mv *MapValue) SetLastGCCycle(cycle int64)          { mv.lastGCCycle = cycle }
func (bmv *BoundMethodValue) SetLastGCCycle(cycle int64) { bmv.lastGCCycle = cycle }
func (pv *PackageValue) SetLastGCCycle(cycle int64)      { pv.lastGCCycle = cycle }
func (nv *NativeValue) SetLastGCCycle(cycle int64)       { nv.lastGCCycle = cycle }
func (b *Block) SetLastGCCycle(cycle int64)              { b.lastGCCycle = cycle }
func (rv RefValue) SetLastGCCycle(cycle int64)           { rv.lastGCCycle = cycle }
func (hiv *HeapItemValue) SetLastGCCycle(cycle int64)    { hiv.lastGCCycle = cycle }
func (tv TypeValue) SetLastGCCycle(cycle int64)          { tv.lastGCCycle = cycle }
