package gnolang

import (
	"fmt"
)

func (m *Machine) doOpValueDecl() {
	s := m.PopStmt().(*ValueDecl)
	m.incrCPU(OpCPUSlopeValueDecl * int64(len(s.NameExprs)))
	lb := m.LastBlock()
	nt := Type(nil)
	if s.Type != nil {
		nt = m.PopValue().GetType()
	}
	var rvs []TypedValue
	if s.Values != nil {
		// `var a, b T = f()`: a result whose declared type differs from T
		// only by name gets the conversion, and the checks, that `var a T = x`
		// gets; preprocess records which in ATTR_ASSIGN_CONVERT. Convert in
		// place, while the results are still on m.Values, as doOpAssign does:
		// Copy can allocate and so run GC, which counts live bytes from
		// m.Values and would miss popped ones.
		if len(s.Values) == 1 && len(s.NameExprs) > 1 {
			if convs, ok := s.GetAttribute(ATTR_ASSIGN_CONVERT).([]Type); ok {
				results := m.PeekValues(len(s.NameExprs))
				for i, t := range convs {
					if t != nil {
						results[i] = results[i].Copy(m.Alloc, m.Store)
						m.convertValue(&results[i], t)
					}
				}
			}
		}
		rvs = m.PopValues(len(s.NameExprs))
	}
	for i := range s.NameExprs {
		var tv TypedValue
		if rvs == nil {
			// NOTE: Go/Gno wart.
			// implicit interface casting could
			// requiring the consideration of the typed-nil case.
			if nt == nil {
				tv = TypedValue{}
			} else if nt.Kind() == InterfaceKind {
				tv = TypedValue{}
			} else {
				tv = defaultTypedValue(m.Alloc, nt)
			}
		} else {
			tv = rvs[i]
		}

		if isUntyped(tv.T) {
			if !s.Const {
				if m.Stage != StagePre && rvs[i].T.Kind() != BoolKind {
					panic("untyped conversion should not happen at runtime")
				}
				ConvertUntypedTo(&tv, nil)
			}
		} else if nt != nil {
			// if nt.T is an interface, maintain tv.T as-is.
			if nt.Kind() != InterfaceKind {
				if debug {
					if nt.TypeID() != tv.T.TypeID() &&
						baseOf(nt).TypeID() != tv.T.TypeID() {
						panic(fmt.Sprintf(
							"type mismatch: %s vs %s",
							nt.TypeID(),
							tv.T.TypeID(),
						))
					}
				}
				tv.T = nt
			}
		}

		nx := &s.NameExprs[i]
		ptr := lb.GetPointerToMaybeHeapDefine(m.Store, nx)
		ptr.Assign2(m, m.Alloc, m.Store, m.Realm, tv, false)
	}
}

func (m *Machine) doOpTypeDecl() {
	s := m.PopStmt().(*TypeDecl)
	t := m.PopValue().GetType()
	tv := asValue(t)
	last := m.LastBlock()
	ptr := last.GetPointerTo(m.Store, s.Path)
	ptr.Assign2(m, m.Alloc, m.Store, m.Realm, tv, false)
}
