package gnolang

func (sv StringValue) DeepFill(store Store) Value {
	//if debug {
	panic("DeepFill should not be called on StringValue")
	//}
	//	return sv
}

func (biv BigintValue) DeepFill(store Store) Value {
	//if debug {
	panic("DeepFill should not be called on BigintValue")
	//}
	//return biv
}

func (bdv BigdecValue) DeepFill(store Store) Value {
	//if debug {
	panic("DeepFill should not be called on BigdecValue")
	//}
	//return bdv
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
	return store.GetObject(rv.ObjectID)
}

func (hiv *HeapItemValue) DeepFill(store Store) Value {
	if hiv.Value.V != nil {
		hiv.Value.V = hiv.Value.V.DeepFill(store)
	}
	return hiv
}
