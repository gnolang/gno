package gnolang

func (m *Machine) doOpDefine() {
	s := m.PopStmt().(*AssignStmt)
	// Define each value evaluated for Lhs.
	// NOTE: PopValues() returns a slice in
	// forward order, not the usual reverse.
	rvs := m.PopValues(len(s.Lhs))
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

		rtv := m.getTypeValueFromNX(nx, s.Rhs[i], rvs[i])
		ptr.Assign2(m.Alloc, m.Store, m.Realm, rtv, true)

		pv, is := ptr.TV.V.(PointerValue)
		if is && pv.TV.ShouldEscape {
			m.escape2Heap(nx, s.Rhs[i], ptr, pv.TV)
			pv.TV.OnHeap = true
			pv.TV.ShouldEscape = false
		}

		if is && pv.TV != nil && pv.TV.OnHeap {
			nx.IsRoot = true
		}
	}
}

func (m *Machine) doOpAssign() {
	s := m.PopStmt().(*AssignStmt)
	// Assign each value evaluated for Lhs.
	// NOTE: PopValues() returns a slice in
	// forward order, not the usual reverse.
	rvs := m.PopValues(len(s.Lhs))
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
		nx := s.Lhs[i].(*NameExpr)
		rtv := m.getTypeValueFromNX(nx, s.Rhs[i], rvs[i])

		lv.Assign2(m.Alloc, m.Store, m.Realm, rtv, true)

		pv, is := lv.TV.V.(PointerValue)
		if is && pv.TV.ShouldEscape {
			m.escape2Heap(nx, s.Rhs[i], lv, pv.TV)
			pv.TV.OnHeap = true
			pv.TV.ShouldEscape = false
		}

		if is && pv.TV != nil && pv.TV.OnHeap {
			nx.IsRoot = true
		}
	}
}

func (m *Machine) getTypeValueFromNX(lhs *NameExpr, rhs Expr, stackDefault TypedValue) TypedValue {
	var obj, root *GCObj
	var rname *NameExpr

	switch name := rhs.(type) {
	case *NameExpr:
		rname = name
	case *RefExpr:
		rname = name.X.(*NameExpr)
	}
	if rname != nil {
		root = m.GC.getRootByPath(rname.Path.String())
	}
	var rtv TypedValue

	if root != nil {
		obj = root.ref
	}

	if obj != nil {
		if ptr, is := obj.value.(PointerValue); is {
			rtv = *ptr.TV
		} else if t, iss := obj.value.(*TypedValue); iss {
			rtv = *t
		}
		m.GC.AddRoot(&GCObj{
			path:   lhs.Path.String(),
			marked: false,
			ref:    obj,
		})
		return rtv
	}
	return stackDefault
}

func (m *Machine) escape2Heap(nx *NameExpr, rhs Expr, pv PointerValue, tv *TypedValue) {
	obj := &GCObj{
		value:  tv,
		marked: false,
		ref:    nil,
	}
	root := &GCObj{
		path:   nx.Path.String(),
		value:  pv,
		marked: false,
		ref:    obj,
	}
	m.GC.AddRoot(root)
	m.GC.AddObject(obj)

	if refExpr, ok := rhs.(*RefExpr); ok {
		rn := refExpr.X.(*NameExpr)
		rn.IsRoot = true

		rroot := &GCObj{
			path:   rn.Path.String(),
			value:  pv,
			marked: false,
			ref:    obj,
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
