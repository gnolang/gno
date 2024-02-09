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
		m.handleRedefinition(lb, nx.Path, ptr.TV)

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

// handleRedefinition handles a special case where a named symbol is redefined within
// the same block. This can happen in a loop block or a combination of labels and gotos.
// During redefinitions, we should check if there are closures from within this block referencing
// this symbol. If so, publish the value to the function value before it is redefined with a new value.
func (m *Machine) handleRedefinition(block *Block, valuePath ValuePath, tv *TypedValue) {
	// This is only a redifinition if the type is not nil. Also don't bother with
	// the special case of the blank identifier; the index of this identifier can only
	// have unintended side effects on actual zero indexed identifiers.
	if tv.T == nil || valuePath.Name == "_" {
		return
	}

	// Do nothing if the block has no closure references.
	var funcRefs []*funcValueWithDepth
	if funcRefs = block.closureRefs[int(valuePath.Index)]; funcRefs == nil {
		return
	}

	tvCopy := tv.unrefCopy(m.Alloc, m.Store)
	for _, ref := range funcRefs {
		newNameExpr := noAttrNameExpr{
			Name:      valuePath.Name,
			ValuePath: valuePath,
		}

		newNameExpr.Depth = ref.depth
		ref.fv.finalizedNamedTypes[newNameExpr] = &tvCopy
	}

	// This particular reference has been resolved.
	delete(block.closureRefs, int(valuePath.Index))
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
