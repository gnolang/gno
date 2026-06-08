package gnolang

func (m *Machine) doOpDefine() {
	s := m.PopStmt().(*AssignStmt)
	// Define each value evaluated for Lhs.
	// NOTE: PopValues() returns a slice in
	// forward order, not the usual reverse.
	rvs := m.PopValues(len(s.Lhs))
	m.incrCPU(OpCPUSlopeDefine * int64(len(s.Lhs)))
	lb := m.LastBlock()
	for i := range s.Lhs {
		// Get name and value of i'th term.
		nx := s.Lhs[i].(*NameExpr)
		// Finally, define (or assign if loop block).
		ptr := lb.GetPointerToMaybeHeapDefine(m.Store, nx)
		if m.Stage != StagePre && isUntyped(rvs[i].T) && rvs[i].T.Kind() != BoolKind {
			panic("untyped conversion should not happen at runtime")
		}
		ptr.Assign2(m, m.Alloc, m.Store, m.Realm, rvs[i], true)
	}
}

func (m *Machine) doOpAssign() {
	s := m.PopStmt().(*AssignStmt)
	// Go spec (§Assignments): in a tuple assignment, all operands of the LHS
	// and all RHS expressions are evaluated FIRST (op_exec already did this:
	// their values sit on m.Values), then the assignments are carried out in
	// left-to-right order. Two consequences we must preserve here:
	//   - left-to-right: `a, a, a = 1, 2, 3` must leave a == 3 (the last write
	//     wins), so we Assign2 in increasing LHS index order.
	//   - panic atomicity: resolving an LHS pointer (resolvePointer) or the
	//     Assign2 itself may panic (e.g. `m[k], *p = 42, 2` nil-derefs on *p);
	//     by then the earlier writes (m[k] = 42) must already be committed.
	//     So we interleave resolve+assign per LHS left-to-right, NOT
	//     resolve-all-then-assign-all. See golang/go#23017.
	//
	// NOTE: PopValues() returns a slice in forward order, not the usual
	// reverse. rvs[0] is the leftmost RHS value.
	rvs := m.PopValues(len(s.Lhs))
	m.incrCPU(OpCPUSlopeAssign * int64(len(s.Lhs)))

	if len(s.Lhs) == 1 {
		// Single-LHS fast path: no operand-frame slicing needed; PopAsPointer
		// consumes this LHS's operand frame (now on top of the stack) directly.
		lv := m.PopAsPointer(s.Lhs[0])
		if m.Stage != StagePre && isUntyped(rvs[0].T) && rvs[0].T.Kind() != BoolKind {
			panic("untyped conversion should not happen at runtime")
		}
		lv.Assign2(m, m.Alloc, m.Store, m.Realm, rvs[0], true)
		return
	}

	// Multi-LHS: after popping the RHS values above, m.Values' top region holds
	// the LHS operand frames, oldest first — LHS_0's frame at the bottom up to
	// LHS_{n-1}'s frame on top (op_exec pushes PushForPointer for the LHS in
	// reverse, so they execute LHS_0 first). Take ONE slice window over that
	// whole region and truncate the stack once, then resolve each LHS in place
	// from its sub-slice. This avoids the per-statement make([][]TypedValue, N)
	// heap allocation of the buffer-and-repush approach.
	//
	// INVARIANT: `ops` (and `rvs` above it) alias the live m.Values backing
	// array with its full capacity retained. Nothing in the loop below may push
	// onto m.Values — an append would write in place into those windows and
	// corrupt not-yet-resolved frames. This holds because resolvePointer and
	// Assign2 are leaf operations that execute no bytecode (the rvs alias is
	// pre-existing precedent: the old loop read rvs[i] across Assign2 too). The
	// debug assertion below guards against a future callee breaking it.
	total := 0
	for _, lx := range s.Lhs {
		total += numStackValuesForPointer(lx)
	}
	ops := m.Values[len(m.Values)-total:]
	m.Values = m.Values[:len(m.Values)-total]
	stackLen := len(m.Values)

	offset := 0
	for i, lx := range s.Lhs {
		sz := numStackValuesForPointer(lx)
		// Resolve LHS_i's pointer from its operand frame, then immediately
		// assign — keeping the left-to-right, fail-with-earlier-writes-committed
		// discipline described above.
		lv, ro := m.resolvePointer(lx, ops[offset:offset+sz])
		if ro {
			m.Panic(typedString(readonlyAccessPanic(lx)))
		}
		offset += sz
		if m.Stage != StagePre && isUntyped(rvs[i].T) && rvs[i].T.Kind() != BoolKind {
			panic("untyped conversion should not happen at runtime")
		}
		lv.Assign2(m, m.Alloc, m.Store, m.Realm, rvs[i], true)
		if debug && len(m.Values) != stackLen {
			panic("doOpAssign: value stack grew mid-loop; ops/rvs aliases corrupted")
		}
	}
}

func (m *Machine) doOpAddAssign() {
	s := m.PopStmt().(*AssignStmt)
	rv := m.PopValue() // only one.
	lv := m.PopAsPointer(s.Lhs[0])
	if debug {
		debugAssertSameTypes(lv.TV.T, rv.T)
	}

	// Per-N gas for BigInt/BigDec.
	m.incrCPUBigInt(lv.TV, rv, OpCPUSlopeBigIntAdd)
	m.incrCPUBigDec(lv.TV, rv, OpCPUSlopeBigDecAdd)

	// add rv to lv.
	addAssign(m.Alloc, lv.TV, rv)
	if lv.Base != nil {
		m.Realm.DidUpdate(m, lv.Base.(Object), nil, nil)
	}
}

func (m *Machine) doOpSubAssign() {
	s := m.PopStmt().(*AssignStmt)
	rv := m.PopValue() // only one.
	lv := m.PopAsPointer(s.Lhs[0])
	if debug {
		debugAssertSameTypes(lv.TV.T, rv.T)
	}

	m.incrCPUBigInt(lv.TV, rv, OpCPUSlopeBigIntSub)
	m.incrCPUBigDec(lv.TV, rv, OpCPUSlopeBigDecSub)

	// sub rv from lv.
	subAssign(lv.TV, rv)
	if lv.Base != nil {
		m.Realm.DidUpdate(m, lv.Base.(Object), nil, nil)
	}
}

func (m *Machine) doOpMulAssign() {
	s := m.PopStmt().(*AssignStmt)
	rv := m.PopValue() // only one.
	lv := m.PopAsPointer(s.Lhs[0])
	if debug {
		debugAssertSameTypes(lv.TV.T, rv.T)
	}

	m.incrCPUBigIntQuad(lv.TV, rv, OpCPUSlopeBigIntMulQ)
	m.incrCPUBigDecQuad(lv.TV, rv, OpCPUSlopeBigDecMulQ)

	// lv *= rv
	mulAssign(lv.TV, rv)
	if lv.Base != nil {
		m.Realm.DidUpdate(m, lv.Base.(Object), nil, nil)
	}
}

func (m *Machine) doOpQuoAssign() {
	s := m.PopStmt().(*AssignStmt)
	rv := m.PopValue() // only one.
	lv := m.PopAsPointer(s.Lhs[0])
	if debug {
		debugAssertSameTypes(lv.TV.T, rv.T)
	}

	m.incrCPUBigIntQuad(lv.TV, rv, OpCPUSlopeBigIntQuoQ)
	m.incrCPUBigDecQuad(lv.TV, rv, OpCPUSlopeBigDecQuoQ)

	// lv /= rv
	err := quoAssign(lv.TV, rv)
	if err != nil {
		panic(err)
	}

	if lv.Base != nil {
		m.Realm.DidUpdate(m, lv.Base.(Object), nil, nil)
	}
}

func (m *Machine) doOpRemAssign() {
	s := m.PopStmt().(*AssignStmt)
	rv := m.PopValue() // only one.
	lv := m.PopAsPointer(s.Lhs[0])
	if debug {
		debugAssertSameTypes(lv.TV.T, rv.T)
	}

	m.incrCPUBigIntQuad(lv.TV, rv, OpCPUSlopeBigIntRemQ)

	// lv %= rv
	err := remAssign(lv.TV, rv)
	if err != nil {
		panic(err)
	}

	if lv.Base != nil {
		m.Realm.DidUpdate(m, lv.Base.(Object), nil, nil)
	}
}

func (m *Machine) doOpBandAssign() {
	s := m.PopStmt().(*AssignStmt)
	rv := m.PopValue() // only one.
	lv := m.PopAsPointer(s.Lhs[0])
	if debug {
		debugAssertSameTypes(lv.TV.T, rv.T)
	}

	m.incrCPUBigInt(lv.TV, rv, OpCPUSlopeBigIntBand)

	// lv &= rv
	bandAssign(lv.TV, rv)
	if lv.Base != nil {
		m.Realm.DidUpdate(m, lv.Base.(Object), nil, nil)
	}
}

func (m *Machine) doOpBandnAssign() {
	s := m.PopStmt().(*AssignStmt)
	rv := m.PopValue() // only one.
	lv := m.PopAsPointer(s.Lhs[0])
	if debug {
		debugAssertSameTypes(lv.TV.T, rv.T)
	}

	m.incrCPUBigInt(lv.TV, rv, OpCPUSlopeBigIntBandn)

	// lv &^= rv
	bandnAssign(lv.TV, rv)
	if lv.Base != nil {
		m.Realm.DidUpdate(m, lv.Base.(Object), nil, nil)
	}
}

func (m *Machine) doOpBorAssign() {
	s := m.PopStmt().(*AssignStmt)
	rv := m.PopValue() // only one.
	lv := m.PopAsPointer(s.Lhs[0])
	if debug {
		debugAssertSameTypes(lv.TV.T, rv.T)
	}

	m.incrCPUBigInt(lv.TV, rv, OpCPUSlopeBigIntBor)

	// lv |= rv
	borAssign(lv.TV, rv)
	if lv.Base != nil {
		m.Realm.DidUpdate(m, lv.Base.(Object), nil, nil)
	}
}

func (m *Machine) doOpXorAssign() {
	s := m.PopStmt().(*AssignStmt)
	rv := m.PopValue() // only one.
	lv := m.PopAsPointer(s.Lhs[0])
	if debug {
		debugAssertSameTypes(lv.TV.T, rv.T)
	}

	m.incrCPUBigInt(lv.TV, rv, OpCPUSlopeBigIntXor)

	// lv ^= rv
	xorAssign(lv.TV, rv)
	if lv.Base != nil {
		m.Realm.DidUpdate(m, lv.Base.(Object), nil, nil)
	}
}

func (m *Machine) doOpShlAssign() {
	s := m.PopStmt().(*AssignStmt)
	rv := m.PopValue() // only one.
	lv := m.PopAsPointer(s.Lhs[0])

	// Per-N gas for BigInt Shl: charge per-kilobit of shift amount.
	if lv.TV.T == UntypedBigintType {
		m.incrCPU(int64(rv.GetUint()) * OpCPUSlopeBigIntShl / 1024)
	}

	// lv <<= rv
	shlAssign(m, lv.TV, rv)
	if lv.Base != nil {
		m.Realm.DidUpdate(m, lv.Base.(Object), nil, nil)
	}
}

func (m *Machine) doOpShrAssign() {
	s := m.PopStmt().(*AssignStmt)
	rv := m.PopValue() // only one.
	lv := m.PopAsPointer(s.Lhs[0])

	m.incrCPUBigUnary(lv.TV, OpCPUSlopeBigIntShr)

	// lv >>= rv
	shrAssign(m, lv.TV, rv)
	if lv.Base != nil {
		m.Realm.DidUpdate(m, lv.Base.(Object), nil, nil)
	}
}
