package gno

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
		ptr.Assign2(m.Store, m.Realm, rvs[i], true)
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
		lv.Assign2(m.Store, m.Realm, rvs[i], true)
	}
}

func (m *Machine) doOpAddAssign() {
	s := m.PopStmt().(*AssignStmt)
	rv := m.PopValue() // only one.
	lv := m.PopAsPointer(s.Lhs[0])
	if debug {
		assertSameTypes(lv.TV.T, rv.T)
	}

	// add rv to lv.
	addAssign(lv.TV, rv)
}

func (m *Machine) doOpSubAssign() {
	s := m.PopStmt().(*AssignStmt)
	rv := m.PopValue() // only one.
	lv := m.PopAsPointer(s.Lhs[0])
	if debug {
		assertSameTypes(lv.TV.T, rv.T)
	}

	// sub rv from lv.
	subAssign(lv.TV, rv)
}

func (m *Machine) doOpMulAssign() {
	s := m.PopStmt().(*AssignStmt)
	rv := m.PopValue() // only one.
	lv := m.PopAsPointer(s.Lhs[0])
	if debug {
		assertSameTypes(lv.TV.T, rv.T)
	}

	// lv *= rv
	mulAssign(lv.TV, rv)
}

func (m *Machine) doOpQuoAssign() {
	s := m.PopStmt().(*AssignStmt)
	rv := m.PopValue() // only one.
	lv := m.PopAsPointer(s.Lhs[0])
	if debug {
		assertSameTypes(lv.TV.T, rv.T)
	}

	// lv /= rv
	quoAssign(lv.TV, rv)
}

func (m *Machine) doOpRemAssign() {
	s := m.PopStmt().(*AssignStmt)
	rv := m.PopValue() // only one.
	lv := m.PopAsPointer(s.Lhs[0])
	if debug {
		assertSameTypes(lv.TV.T, rv.T)
	}

	// lv %= rv
	remAssign(lv.TV, rv)
}

func (m *Machine) doOpBandAssign() {
	s := m.PopStmt().(*AssignStmt)
	rv := m.PopValue() // only one.
	lv := m.PopAsPointer(s.Lhs[0])
	if debug {
		assertSameTypes(lv.TV.T, rv.T)
	}

	// lv &= rv
	bandAssign(lv.TV, rv)
}

func (m *Machine) doOpBandnAssign() {
	s := m.PopStmt().(*AssignStmt)
	rv := m.PopValue() // only one.
	lv := m.PopAsPointer(s.Lhs[0])
	if debug {
		assertSameTypes(lv.TV.T, rv.T)
	}

	// lv &^= rv
	bandnAssign(lv.TV, rv)
}

func (m *Machine) doOpBorAssign() {
	s := m.PopStmt().(*AssignStmt)
	rv := m.PopValue() // only one.
	lv := m.PopAsPointer(s.Lhs[0])
	if debug {
		assertSameTypes(lv.TV.T, rv.T)
	}

	// lv |= rv
	borAssign(lv.TV, rv)
}

func (m *Machine) doOpXorAssign() {
	s := m.PopStmt().(*AssignStmt)
	rv := m.PopValue() // only one.
	lv := m.PopAsPointer(s.Lhs[0])
	if debug {
		assertSameTypes(lv.TV.T, rv.T)
	}

	// lv ^= rv
	xorAssign(lv.TV, rv)
}

func (m *Machine) doOpShlAssign() {
	panic("not yet implemented")
}

func (m *Machine) doOpShrAssign() {
	panic("not yet implemented")
}
