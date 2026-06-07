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

// doOpAssign desugars a (possibly multi-target) assignment into a sequence of
// per-slot store ops, mirroring how the Go compiler lowers OAS2 in
// cmd/compile/internal/walk: all RHS values and all LHS address operands are
// already evaluated (they sit on the value stack); here we only re-arrange
// them and schedule one OpAssignSlot per LHS so that the actual
// pointer-resolution + store of each target runs as an independent step, in
// left-to-right order.
//
// This matters because resolving the pointer for L_i (PopAsPointer) can itself
// panic — nil-deref (*p), out-of-range index, nil-map write. Per the Go spec
// (§Assignment statements, §Order of evaluation) the stores are carried out
// left-to-right after all operands are evaluated, so a panic while storing L_i
// must leave the stores to L_0..L_{i-1} committed. Scheduling each store as a
// separate op (rather than one right-to-left resolve-then-store loop) gives
// that atomicity for free: when L_i's op panics, the ops for L_0..L_{i-1} have
// already run to completion and been popped.
//
// Value-stack layout on entry (bottom -> top), as pushed by op_exec.go:
//
//	[ L_0 operands ][ L_1 operands ]...[ L_{n-1} operands ][ R_0 ][ R_1 ]...[ R_{n-1} ]
//
// We pop the RHS values, then pop each LHS operand frame, then re-push one
// self-contained group per slot in reverse store order plus an OpAssignSlot
// for each. Because ops execute LIFO, the slot-0 group/op ends up on top and
// runs first. Each group is (bottom -> top): R_i, L_i operands, i — the slot
// index i is read by doOpAssignSlot to recover s.Lhs[i] for PopAsPointer.
func (m *Machine) doOpAssign() {
	s := m.PopStmt().(*AssignStmt)
	m.incrCPU(OpCPUSlopeAssign * int64(len(s.Lhs)))

	n := len(s.Lhs)
	// Single-target assignment (Go's OAS, not OAS2) has no left-to-right
	// ordering to preserve: there is no earlier store to leave committed if the
	// lone PopAsPointer panics. Store it directly, avoiding the per-slot
	// re-arrange and its allocations on this very hot path.
	if n == 1 {
		rv := *m.PopValue()
		lv := m.PopAsPointer(s.Lhs[0])
		if m.Stage != StagePre && isUntyped(rv.T) && rv.T.Kind() != BoolKind {
			panic("untyped conversion should not happen at runtime")
		}
		lv.Assign2(m, m.Alloc, m.Store, m.Realm, rv, true)
		return
	}

	// NOTE: PopValues() returns values in forward (stack, oldest-first) order,
	// aliasing the stack's backing array; snapshot into independent Go slices
	// because we immediately re-push (which would overwrite those slots).
	// These are plain struct copies (no Allocator/Copy involved), matching the
	// original code, which never copied the assigned values either.
	rvs := append([]TypedValue(nil), m.PopValues(n)...)
	// Pop each LHS operand frame (right-to-left, since they sit on top of the
	// stack in left-to-right push order).
	frames := make([][]TypedValue, n)
	for i := n - 1; 0 <= i; i-- {
		k := numStackValuesForPointer(s.Lhs[i])
		frames[i] = append([]TypedValue(nil), m.PopValues(k)...)
	}
	// Re-push one group + op per slot, slot-0 last so it is on top and stored
	// first.
	for i := n - 1; 0 <= i; i-- {
		m.PushValue(rvs[i])
		for _, tv := range frames[i] {
			m.PushValue(tv)
		}
		var idx TypedValue
		idx.T = IntType
		idx.SetInt(int64(i))
		m.PushValue(idx)
		m.PushStmt(s)
		m.PushOp(OpAssignSlot)
	}
}

// doOpAssignSlot stores a single LHS target. See doOpAssign for the layout it
// consumes: top-of-stack is the slot index i, below it L_i's address operands,
// and below those the RHS value R_i.
func (m *Machine) doOpAssignSlot() {
	s := m.PopStmt().(*AssignStmt)
	i := int(m.PopValue().GetInt())
	// PopAsPointer consumes L_i's operands (now on top) and may panic.
	lv := m.PopAsPointer(s.Lhs[i])
	rv := *m.PopValue()
	if m.Stage != StagePre && isUntyped(rv.T) && rv.T.Kind() != BoolKind {
		panic("untyped conversion should not happen at runtime")
	}
	lv.Assign2(m, m.Alloc, m.Store, m.Realm, rv, true)
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
