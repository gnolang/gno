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
