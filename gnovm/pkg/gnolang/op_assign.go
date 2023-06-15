package gnolang

func (m *Machine) doOpDefine() {
	s := m.PopStmt().(*AssignStmt)
	// Define each value evaluated for Lhs.
	// NOTE: PopValues() returns a slice in
	// forward order, not the usual reverse.

	lb := m.LastBlock()
	for i := 0; i < len(s.Lhs); i++ {
		// Get name and value of i'th term.
		nx := s.Lhs[i].(*NameExpr)
		// Finally, define (or assign if loop block).
		ptr := lb.GetPointerTo(m.Store, nx.Path)
		// XXX HACK (until value persistence impl'd)
		if m.ReadOnly {
			if oo, ok := ptr.Base.(Object); ok {
				if oo.GetIsReal() {
					panic("readonly violation")
				}
			}
		}

		val := *m.PopValue()

		ptr.Assign2(m.Alloc, m.Store, m.Realm, val, true)

		if ptr.TV.OnHeap {
			m.GC.setEmptyRootPath(&nx.Path)
		}
	}
}

func (m *Machine) doOpAssign() {
	s := m.PopStmt().(*AssignStmt)
	// Assign each value evaluated for Lhs.
	// NOTE: PopValues() returns a slice in
	// forward order, not the usual reverse.
	for i := len(s.Lhs) - 1; 0 <= i; i-- {
		// Pop lhs value and desired type.
		lv := m.PopAsPointer(s.Lhs[i])
		// XXX HACK (until value persistence impl'd)
		if m.ReadOnly {
			if oo, ok := lv.Base.(Object); ok {
				if oo.GetIsReal() {
					panic("readonly violation")
				}
			}
		}

		// Get name and value of i'th term.
		//nx := s.Lhs[i].(*NameExpr)
		val := *m.PopValue()

		lv.Assign2(m.Alloc, m.Store, m.Realm, val, true)
	}
}

func (m *Machine) getTypeValueFromNX(nx *NameExpr, rhs Expr) *TypedValue {
	var obj, root *GCObj
	var rname *NameExpr
	var shouldCopy bool

	switch name := rhs.(type) {
	case *NameExpr:
		rname = name
		shouldCopy = true
	case *RefExpr:
		rname = name.X.(*NameExpr)
	}
	if rname == nil {
		return nil
	}

	if !rname.IsRoot {
		return nil
	}

	root = m.GC.getRootByPath(&rname.Path)

	if root == nil {
		return nil
	}
	obj = root.ref

	if obj == nil {
		return nil
	}

	if shouldCopy {
		newCopy := *obj
		m.GC.AddRoot(&GCObj{
			ref:  &newCopy,
			path: &nx.Path,
		})
		m.GC.AddObject(&newCopy)
		obj = &newCopy
	}

	return &obj.value
}

func (m *Machine) escape2Heap(nx *NameExpr, rhs Expr, rp PointerValue) {
	obj := &GCObj{value: TypedValue{
		T:      &PointerType{Elt: rp.TV.T},
		V:      rp,
		OnHeap: true,
	}, path: &nx.Path}

	root := &GCObj{
		path: &nx.Path,
		ref:  obj,
	}
	m.GC.AddRoot(root)
	m.GC.AddObject(obj)

	if refExpr, ok := rhs.(*RefExpr); ok {
		rn := refExpr.X.(*NameExpr)
		rn.IsRoot = true

		rroot := &GCObj{
			path: &rn.Path,
			ref:  obj,
		}
		m.GC.AddRoot(rroot)
	}
}

func (m *Machine) doOpAddAssign() {
	s := m.PopStmt().(*AssignStmt)
	rv := m.PopValue() // only one.
	lv := m.PopAsPointer(s.Lhs[0])
	if debug {
		assertSameTypes(lv.TV.T, rv.T)
	}

	// XXX HACK (until value persistence impl'd)
	if m.ReadOnly {
		if oo, ok := lv.Base.(Object); ok {
			if oo.GetIsReal() {
				panic("readonly violation")
			}
		}
	}
	// add rv to lv.
	addAssign(m.Alloc, lv.TV, rv)
	if lv.Base != nil {
		m.Realm.DidUpdate(lv.Base.(Object), nil, nil)
	}
}

func (m *Machine) doOpSubAssign() {
	s := m.PopStmt().(*AssignStmt)
	rv := m.PopValue() // only one.
	lv := m.PopAsPointer(s.Lhs[0])
	if debug {
		assertSameTypes(lv.TV.T, rv.T)
	}

	// XXX HACK (until value persistence impl'd)
	if m.ReadOnly {
		if oo, ok := lv.Base.(Object); ok {
			if oo.GetIsReal() {
				panic("readonly violation")
			}
		}
	}
	// sub rv from lv.
	subAssign(lv.TV, rv)
	if lv.Base != nil {
		m.Realm.DidUpdate(lv.Base.(Object), nil, nil)
	}
}

func (m *Machine) doOpMulAssign() {
	s := m.PopStmt().(*AssignStmt)
	rv := m.PopValue() // only one.
	lv := m.PopAsPointer(s.Lhs[0])
	if debug {
		assertSameTypes(lv.TV.T, rv.T)
	}

	// XXX HACK (until value persistence impl'd)
	if m.ReadOnly {
		if oo, ok := lv.Base.(Object); ok {
			if oo.GetIsReal() {
				panic("readonly violation")
			}
		}
	}
	// lv *= rv
	mulAssign(lv.TV, rv)
	if lv.Base != nil {
		m.Realm.DidUpdate(lv.Base.(Object), nil, nil)
	}
}

func (m *Machine) doOpQuoAssign() {
	s := m.PopStmt().(*AssignStmt)
	rv := m.PopValue() // only one.
	lv := m.PopAsPointer(s.Lhs[0])
	if debug {
		assertSameTypes(lv.TV.T, rv.T)
	}

	// XXX HACK (until value persistence impl'd)
	if m.ReadOnly {
		if oo, ok := lv.Base.(Object); ok {
			if oo.GetIsReal() {
				panic("readonly violation")
			}
		}
	}
	// lv /= rv
	quoAssign(lv.TV, rv)
	if lv.Base != nil {
		m.Realm.DidUpdate(lv.Base.(Object), nil, nil)
	}
}

func (m *Machine) doOpRemAssign() {
	s := m.PopStmt().(*AssignStmt)
	rv := m.PopValue() // only one.
	lv := m.PopAsPointer(s.Lhs[0])
	if debug {
		assertSameTypes(lv.TV.T, rv.T)
	}

	// XXX HACK (until value persistence impl'd)
	if m.ReadOnly {
		if oo, ok := lv.Base.(Object); ok {
			if oo.GetIsReal() {
				panic("readonly violation")
			}
		}
	}
	// lv %= rv
	remAssign(lv.TV, rv)
	if lv.Base != nil {
		m.Realm.DidUpdate(lv.Base.(Object), nil, nil)
	}
}

func (m *Machine) doOpBandAssign() {
	s := m.PopStmt().(*AssignStmt)
	rv := m.PopValue() // only one.
	lv := m.PopAsPointer(s.Lhs[0])
	if debug {
		assertSameTypes(lv.TV.T, rv.T)
	}

	// XXX HACK (until value persistence impl'd)
	if m.ReadOnly {
		if oo, ok := lv.Base.(Object); ok {
			if oo.GetIsReal() {
				panic("readonly violation")
			}
		}
	}
	// lv &= rv
	bandAssign(lv.TV, rv)
	if lv.Base != nil {
		m.Realm.DidUpdate(lv.Base.(Object), nil, nil)
	}
}

func (m *Machine) doOpBandnAssign() {
	s := m.PopStmt().(*AssignStmt)
	rv := m.PopValue() // only one.
	lv := m.PopAsPointer(s.Lhs[0])
	if debug {
		assertSameTypes(lv.TV.T, rv.T)
	}

	// XXX HACK (until value persistence impl'd)
	if m.ReadOnly {
		if oo, ok := lv.Base.(Object); ok {
			if oo.GetIsReal() {
				panic("readonly violation")
			}
		}
	}
	// lv &^= rv
	bandnAssign(lv.TV, rv)
	if lv.Base != nil {
		m.Realm.DidUpdate(lv.Base.(Object), nil, nil)
	}
}

func (m *Machine) doOpBorAssign() {
	s := m.PopStmt().(*AssignStmt)
	rv := m.PopValue() // only one.
	lv := m.PopAsPointer(s.Lhs[0])
	if debug {
		assertSameTypes(lv.TV.T, rv.T)
	}

	// XXX HACK (until value persistence impl'd)
	if m.ReadOnly {
		if oo, ok := lv.Base.(Object); ok {
			if oo.GetIsReal() {
				panic("readonly violation")
			}
		}
	}
	// lv |= rv
	borAssign(lv.TV, rv)
	if lv.Base != nil {
		m.Realm.DidUpdate(lv.Base.(Object), nil, nil)
	}
}

func (m *Machine) doOpXorAssign() {
	s := m.PopStmt().(*AssignStmt)
	rv := m.PopValue() // only one.
	lv := m.PopAsPointer(s.Lhs[0])
	if debug {
		assertSameTypes(lv.TV.T, rv.T)
	}

	// XXX HACK (until value persistence impl'd)
	if m.ReadOnly {
		if oo, ok := lv.Base.(Object); ok {
			if oo.GetIsReal() {
				panic("readonly violation")
			}
		}
	}
	// lv ^= rv
	xorAssign(lv.TV, rv)
	if lv.Base != nil {
		m.Realm.DidUpdate(lv.Base.(Object), nil, nil)
	}
}

func (m *Machine) doOpShlAssign() {
	s := m.PopStmt().(*AssignStmt)
	rv := m.PopValue() // only one.
	lv := m.PopAsPointer(s.Lhs[0])

	// XXX HACK (until value persistence impl'd)
	if m.ReadOnly {
		if oo, ok := lv.Base.(Object); ok {
			if oo.GetIsReal() {
				panic("readonly violation")
			}
		}
	}
	// lv <<= rv
	shlAssign(lv.TV, rv)
	if lv.Base != nil {
		m.Realm.DidUpdate(lv.Base.(Object), nil, nil)
	}
}

func (m *Machine) doOpShrAssign() {
	s := m.PopStmt().(*AssignStmt)
	rv := m.PopValue() // only one.
	lv := m.PopAsPointer(s.Lhs[0])

	// XXX HACK (until value persistence impl'd)
	if m.ReadOnly {
		if oo, ok := lv.Base.(Object); ok {
			if oo.GetIsReal() {
				panic("readonly violation")
			}
		}
	}
	// lv >>= rv
	shrAssign(lv.TV, rv)
	if lv.Base != nil {
		m.Realm.DidUpdate(lv.Base.(Object), nil, nil)
	}
}
