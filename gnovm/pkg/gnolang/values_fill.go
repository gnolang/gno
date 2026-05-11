package gnolang

import "fmt"

func (sv StringValue) DeepFill(store Store) Value {
	return sv
}

func (biv BigintValue) DeepFill(store Store) Value {
	return biv
}

func (bdv BigdecValue) DeepFill(store Store) Value {
	return bdv
}

func (dbv DataByteValue) DeepFill(store Store) Value {
	dbv.Base.DeepFill(store)
	return dbv
}

func (pv PointerValue) DeepFill(store Store) Value {
	// No need to fill pv.TV.V because
	// either it will be filled with .Base,
	// or, it was never persisted anyways.
	if pv.Base != nil {
		return PointerValue{
			TV:    pv.TV,
			Base:  pv.Base.DeepFill(store),
			Index: pv.Index,
		}
	}
	return pv
}

func (av *ArrayValue) DeepFill(store Store) Value {
	if av.List != nil {
		for i := range len(av.List) {
			tv := &av.List[i]
			if tv.V != nil {
				tv.V = tv.V.DeepFill(store)
			}
		}
	}
	return av
}

func (sv *SliceValue) DeepFill(store Store) Value {
	if sv.Base != nil {
		sv.Base = sv.Base.DeepFill(store)
	}
	return sv
}

func (sv *StructValue) DeepFill(store Store) Value {
	for i := range len(sv.Fields) {
		tv := &sv.Fields[i]
		if tv.V != nil {
			tv.V = tv.V.DeepFill(store)
		}
	}
	return sv
}

// XXX implement these too
func (fv *FuncValue) DeepFill(store Store) Value         { panic("not yet implemented") }
func (mv *MapValue) DeepFill(store Store) Value          { panic("not yet implemented") }
func (bmv *BoundMethodValue) DeepFill(store Store) Value { panic("not yet implemented") }
func (tv TypeValue) DeepFill(store Store) Value          { panic("not yet implemented") }
func (pv *PackageValue) DeepFill(store Store) Value      { panic("not yet implemented") }
func (b *Block) DeepFill(store Store) Value              { panic("not yet implemented") }

func (rv RefValue) DeepFill(store Store) Value {
	obj := store.GetObject(rv.ObjectID)
	if debugAssert {
		// Verify hash chain: parent's RefValue hash must match child's stored hash.
		// Escaped objects carry zero RefValue hash (resolved via IAVL).
		if childHash := obj.GetHash(); !rv.Hash.IsZero() && rv.Hash != childHash {
			panic(fmt.Sprintf(
				"hash chain broken at %s: parent claims child hash %X, but child has %X",
				rv.ObjectID, rv.Hash.Bytes(), childHash.Bytes()))
		}
	}
	return obj
}

func (erv ExportRefValue) DeepFill(_ Store) Value {
	return erv // export-only; no store lookup
}

func (hiv *HeapItemValue) DeepFill(store Store) Value {
	if hiv.Value.V != nil {
		hiv.Value.V = hiv.Value.V.DeepFill(store)
	}
	return hiv
}
