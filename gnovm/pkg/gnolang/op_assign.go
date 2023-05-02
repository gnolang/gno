package gnolang

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/errors"
)

var ErrOpNotSupported = errors.New("operation not supported")

var assignOps map[Op]func(s *AssignStmt, m *Machine) (cpuCycles int64) = map[Op]func(s *AssignStmt, m *Machine) int64{
	OpDefine:      doOpDefine,
	OpAssign:      doOpAssign,
	OpAddAssign:   doOpAddAssign,
	OpSubAssign:   doOpSubAssign,
	OpMulAssign:   doOpMulAssign,
	OpQuoAssign:   doOpQuoAssign,
	OpRemAssign:   doOpRemAssign,
	OpBandAssign:  doOpBandAssign,
	OpBandnAssign: doOpBandnAssign,
	OpBorAssign:   doOpBorAssign,
	OpXorAssign:   doOpXorAssign,
	OpShlAssign:   doOpShlAssign,
	OpShrAssign:   doOpShrAssign,
}

func (s *AssignStmt) Exec(m *Machine, o Op) (int64, error) {
	op, ok := assignOps[o]

	if !ok {
		return 0, fmt.Errorf("%w: %q is not valid for Assign Statements", ErrOpNotSupported, o)
	}

	return op(s, m), nil
}

func doOpDefine(s *AssignStmt, m *Machine) int64 {
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

	return OpCPUDefine
}

func doOpAssign(s *AssignStmt, m *Machine) int64 {
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

	return OpCPUAssign
}

func doOpAddAssign(s *AssignStmt, m *Machine) int64 {
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

	return OpCPUAddAssign
}

func doOpSubAssign(s *AssignStmt, m *Machine) int64 {
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

	return OpCPUSubAssign
}

func doOpMulAssign(s *AssignStmt, m *Machine) int64 {
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

	return OpCPUMulAssign
}

func doOpQuoAssign(s *AssignStmt, m *Machine) int64 {
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

	return OpCPUQuoAssign
}

func doOpRemAssign(s *AssignStmt, m *Machine) int64 {
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

	return OpCPURemAssign
}

func doOpBandAssign(s *AssignStmt, m *Machine) int64 {
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

	return OpCPUBandAssign
}

func doOpBandnAssign(s *AssignStmt, m *Machine) int64 {
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

	return OpCPUBandnAssign
}

func doOpBorAssign(s *AssignStmt, m *Machine) int64 {
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

	return OpCPUBorAssign
}

func doOpXorAssign(s *AssignStmt, m *Machine) int64 {
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

	return OpCPUXorAssign
}

func doOpShlAssign(s *AssignStmt, m *Machine) int64 {
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

	return OpCPUShlAssign
}

func doOpShrAssign(s *AssignStmt, m *Machine) int64 {
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

	return OpCPUShrAssign
}
