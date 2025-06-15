package gnolang

import (
	"fmt"
)

func (m *Machine) doOpValueDecl() {
	s := m.PopStmt().(*ValueDecl)
	lb := m.LastBlock()
	nt := Type(nil)
	if s.Type != nil {
		nt = m.PopValue().GetType()
	}
	var rvs []TypedValue
	if s.Values != nil {
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
		ptr.Assign2(m.Alloc, m.Store, m.Realm, tv, false)
	}
}

func (m *Machine) doOpTypeDecl() {
	s := m.PopStmt().(*TypeDecl)
	t := m.PopValue().GetType()
	tv := asValue(t)
	last := m.LastBlock()
	ptr := last.GetPointerTo(m.Store, s.Path)
	ptr.Assign2(m.Alloc, m.Store, m.Realm, tv, false)
}
