package gnolang

func (sv StringValue) Fill(store Store) Value {
	return sv
}

func (biv BigintValue) Fill(store Store) Value {
	return biv
}

func (bdv BigdecValue) Fill(store Store) Value {
	return bdv
}

func (dbv DataByteValue) Fill(store Store) Value {
	dbv.Base.Fill(store)
	return dbv
}

func (pv PointerValue) Fill(store Store) Value {
	if pv.Key != nil {
		// only used transiently for assignment!
		panic("should not happen")
	}
	// No need to fill pv.TV.V because
	// either it will be filled with .Base,
	// or, it was never persisted anyways.
	if pv.Base != nil {
		return PointerValue{
			TV:    pv.TV,
			Base:  pv.Base.Fill(store),
			Index: pv.Index,
			Key:   nil,
		}
	}
	return pv
}

func (av *ArrayValue) Fill(store Store) Value {
	if av.List != nil {
		for i := range len(av.List) {
			tv := &av.List[i]
			if tv.V != nil {
				tv.V = tv.V.Fill(store)
			}
		}
	}
	return av
}

func (sv *SliceValue) Fill(store Store) Value {
	if sv.Base != nil {
		sv.Base = sv.Base.Fill(store)
	}
	return sv
}

func (sv *StructValue) Fill(store Store) Value {
	for i := range len(sv.Fields) {
		tv := &sv.Fields[i]
		if tv.V != nil {
			tv.V = tv.V.Fill(store)
		}
	}
	return sv
}

// XXX implement these too
func (fv *FuncValue) Fill(store Store) Value         { panic("not yet implemented") }
func (mv *MapValue) Fill(store Store) Value          { panic("not yet implemented") }
func (bmv *BoundMethodValue) Fill(store Store) Value { panic("not yet implemented") }
func (tv TypeValue) Fill(store Store) Value          { panic("not yet implemented") }
func (pv *PackageValue) Fill(store Store) Value      { panic("not yet implemented") }
func (nv *NativeValue) Fill(store Store) Value       { panic("not yet implemented") }
func (b *Block) Fill(store Store) Value              { panic("not yet implemented") }

func (rv RefValue) Fill(store Store) Value {
	return store.GetObject(rv.ObjectID)
}

func (hiv *HeapItemValue) Fill(store Store) Value {
	if hiv.Value.V != nil {
		hiv.Value.V = hiv.Value.V.Fill(store)
	}
	return hiv
}
