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

		lb.Roots = append(lb.Roots, ptr)

		if ppv, ok := ptr.TV.V.(PointerValue); ok {
			if _, ok := ppv.Base.(*HeapItemValue); ok {
				root := NewObject(*ptr.TV)
				m.Alloc.heap.RemoveRoot(root)
			}
		}

		// XXX HACK (until value persistence impl'd)
		if m.ReadOnly {
			if oo, ok := ptr.Base.(Object); ok {
				if oo.GetIsReal() {
					panic("readonly violation")
				}
			}
		}
		ptr.Assign2(m.Alloc, m.Store, m.Realm, rvs[i], true)
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
		lv.Assign2(m.Alloc, m.Store, m.Realm, rvs[i], true)
	}
}

func (m *Machine) doOpAddAssign() {
	s := m.PopStmt().(*AssignStmt)
	rv := m.PopValue() // only one.
	lv := m.PopAsPointer(s.Lhs[0])
	if debug {
		debugAssertSameTypes(lv.TV.T, rv.T)
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
		debugAssertSameTypes(lv.TV.T, rv.T)
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
		debugAssertSameTypes(lv.TV.T, rv.T)
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
		debugAssertSameTypes(lv.TV.T, rv.T)
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
		debugAssertSameTypes(lv.TV.T, rv.T)
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
		debugAssertSameTypes(lv.TV.T, rv.T)
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
		debugAssertSameTypes(lv.TV.T, rv.T)
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
		debugAssertSameTypes(lv.TV.T, rv.T)
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
		debugAssertSameTypes(lv.TV.T, rv.T)
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
