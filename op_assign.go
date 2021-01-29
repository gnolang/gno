package gno

func (m *Machine) doOpDefine() {
	s := m.PopStmt().(*AssignStmt)
	// For each value evaluated from Rhs,
	// define in LastBlock according to Lhs.
	// NOTE: PopValues() returns a slice in
	// forward order, not the usual reverse.
	rvs := m.PopValues(len(s.Rhs))
	for i := 0; i < len(s.Rhs); i++ {
		// Get name and value of i'th term.
		nx := s.Lhs[i].(*NameExpr)
		rv := rvs[i]
		if debug {
			// This is how run-time untyped const
			// conversions would work, but we
			// expect the preprocessor to convert
			// these to *constExpr.
			/*
				// Convert if untyped const.
				if isUntyped(rv.T) {
					ConvertUntypedTo(&rv, defaultTypeOf(rv.T))
				}
			*/
			if isUntyped(rv.T) {
				panic("unexpected untyped const type for assign during runtime")
			}
		}
		// Define in block in forward order.
		lv := m.LastBlock().GetValueRefAt(nx.Path)
		lv.Assign(rv)
	}
}

func (m *Machine) doOpAssign() {
	s := m.PopStmt().(*AssignStmt)
	// For each value evaluated from Rhs,
	// assign in LastBlock according to Lhs.
	// NOTE: PopValues() returns a slice in
	// forward order, not the usual reverse.
	rvs := m.PopValues(len(s.Rhs))
	for i := len(s.Rhs) - 1; 0 <= i; i-- {
		rv := rvs[i]
		// Pop lhs value and desired type.
		lv := m.PopForAssign(s.Lhs[i])
		if debug {
			// This is how run-time untyped const
			// conversions would work, but we
			// expect the preprocessor to convert
			// these to *constExpr.
			/*
				// Convert if untyped const.
				if isUntyped(rv.T) {
					ConvertUntypedTo(&rv, lt)
				}
			*/
			if isUntyped(rv.T) {
				panic("unexpected untyped const type for assign during runtime")
			}
		}
		// Finally, assign.
		lv.Assign(rv)
	}
}

func (m *Machine) doOpAddAssign() {
	s := m.PopStmt().(*AssignStmt)
	rv := m.PopValue() // only one.
	lv := m.PopForAssign(s.Lhs[0])
	if debug {
		assertTypes(lv.T, rv.T)
	}

	// add rv to lv.
	addAssign(lv, rv)
}

func (m *Machine) doOpSubAssign() {
	s := m.PopStmt().(*AssignStmt)
	rv := m.PopValue() // only one.
	lv := m.PopForAssign(s.Lhs[0])
	if debug {
		assertTypes(lv.T, rv.T)
	}

	// sub rv from lv.
	subAssign(lv, rv)
}

func (m *Machine) doOpMulAssign() {
	s := m.PopStmt().(*AssignStmt)
	rv := m.PopValue() // only one.
	lv := m.PopForAssign(s.Lhs[0])
	if debug {
		assertTypes(lv.T, rv.T)
	}

	// lv *= rv
	mulAssign(lv, rv)
}

func (m *Machine) doOpQuoAssign() {
	s := m.PopStmt().(*AssignStmt)
	rv := m.PopValue() // only one.
	lv := m.PopForAssign(s.Lhs[0])
	if debug {
		assertTypes(lv.T, rv.T)
	}

	// lv /= rv
	quoAssign(lv, rv)
}

func (m *Machine) doOpRemAssign() {
	s := m.PopStmt().(*AssignStmt)
	rv := m.PopValue() // only one.
	lv := m.PopForAssign(s.Lhs[0])
	if debug {
		assertTypes(lv.T, rv.T)
	}

	// lv %= rv
	remAssign(lv, rv)
}

func (m *Machine) doOpBandAssign() {
	s := m.PopStmt().(*AssignStmt)
	rv := m.PopValue() // only one.
	lv := m.PopForAssign(s.Lhs[0])
	if debug {
		assertTypes(lv.T, rv.T)
	}

	// lv &= rv
	bandAssign(lv, rv)
}

func (m *Machine) doOpBandnAssign() {
	s := m.PopStmt().(*AssignStmt)
	rv := m.PopValue() // only one.
	lv := m.PopForAssign(s.Lhs[0])
	if debug {
		assertTypes(lv.T, rv.T)
	}

	// lv &^= rv
	bandnAssign(lv, rv)
}

func (m *Machine) doOpBorAssign() {
	s := m.PopStmt().(*AssignStmt)
	rv := m.PopValue() // only one.
	lv := m.PopForAssign(s.Lhs[0])
	if debug {
		assertTypes(lv.T, rv.T)
	}

	// lv |= rv
	borAssign(lv, rv)
}

func (m *Machine) doOpXorAssign() {
	s := m.PopStmt().(*AssignStmt)
	rv := m.PopValue() // only one.
	lv := m.PopForAssign(s.Lhs[0])
	if debug {
		assertTypes(lv.T, rv.T)
	}

	// lv ^= rv
	xorAssign(lv, rv)
}

func (m *Machine) doOpShlAssign() {
	panic("not yet implemented")
}

func (m *Machine) doOpShrAssign() {
	panic("not yet implemented")
}
