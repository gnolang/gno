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
	// Go spec §Assignments: operands and RHS are evaluated first (done by
	// op_exec; values sit on m.Values), then assigned left-to-right. We resolve
	// and assign each LHS in increasing index order so the last write wins
	// (`a, a, a = 1, 2, 3` ⇒ 3) and a panic mid-assignment leaves earlier writes
	// committed (`m[k], *p = 42, 2`). See golang/go#23017.
	m.incrCPU(OpCPUSlopeAssign * int64(len(s.Lhs)))

	if len(s.Lhs) == 1 {
		// Fast path: one RHS value, resolve the LHS pointer off the stack top.
		rv := m.PopValue()
		lv := m.PopAsPointer(s.Lhs[0])
		if m.Stage != StagePre && isUntyped(rv.T) && rv.T.Kind() != BoolKind {
			panic("untyped conversion should not happen at runtime")
		}
		lv.Assign2(m, m.Alloc, m.Store, m.Realm, *rv, true)
		return
	}

	// NOTE: Gas: the loop below costs ~6-13% more CPU per LHS than
	// OpCPUSlopeAssign charges, depending on LHS shape (~6% IndexExpr, ~13%
	// NameExpr; see BenchmarkDoOpAssign_*_N*). Not recalibrated: the
	// constant also prices the unchanged single-LHS fast path above, so
	// raising it would overcharge the common case to fit a rare one. If
	// exact pricing is wanted, add a per-LHS multi-assign surcharge constant
	// instead (consensus-breaking, so best batched with other gas changes).

	// The RHS values sit on top in forward order, the LHS operand frames just
	// below, LHS_0 first. Both stay on m.Values until the op returns or panics,
	// so GC run by a Copy or an Assign2 still counts every destination, such as
	// the map `f()[i]` writes into.
	//
	// INVARIANT: the deferred pop assumes the top len(window) slots stay in
	// place, so no Go helper the op reaches may push to or pop from m.Values;
	// the conversion below, resolvePointer and Assign2 reach none.
	total := 0
	for _, lx := range s.Lhs {
		total += numStackValuesForPointer(lx)
	}
	stackLen := len(m.Values)
	window := m.PeekValues(total + len(s.Lhs))
	lhsOperands, rvs := window[:total], window[total:]
	defer m.PopValues(len(window))

	// `a, b = f()` where a result differs from its Lhs only by name: convert
	// it as `a = x` would, before any Lhs is written.
	if convs, ok := s.GetAttribute(ATTR_ASSIGN_CONVERT).([]Type); ok {
		m.convertTopValues(convs)
	}

	offset := 0
	for i, lx := range s.Lhs {
		sz := numStackValuesForPointer(lx)
		lv, ro := m.resolvePointer(lx, lhsOperands[offset:offset+sz])
		if ro {
			m.Panic(typedString(readonlyAccessPanic(lx)))
		}
		offset += sz
		if m.Stage != StagePre && isUntyped(rvs[i].T) && rvs[i].T.Kind() != BoolKind {
			panic("untyped conversion should not happen at runtime")
		}
		lv.Assign2(m, m.Alloc, m.Store, m.Realm, rvs[i], true)
		if debug && len(m.Values) != stackLen {
			panic("doOpAssign: value stack changed mid-loop")
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

	// Per-N gas for BigInt Shl: charge per-kilobit of both the shift amount
	// and the operand bit-width, mirroring doOpShl (and doOpShrAssign below).
	// This runs only at statement execution with a concrete-typed lvalue
	// (const targets are rejected; a variable can't hold an untyped bigint),
	// so the branch is currently unreachable — kept symmetric with doOpShl so
	// the operand-size undercharge can't creep back if that ever changes.
	if lv.TV.T == UntypedBigintType {
		// Clamped for the same overflow reason as doOpShl.
		shift := min(rv.GetUint(), uint64(maxBigintShift))
		m.incrCPU(int64(shift) * OpCPUSlopeBigIntShl / 1024)
		m.incrCPUBigUnary(lv.TV, OpCPUSlopeBigIntShl)
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
