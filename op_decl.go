package gno

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
	fmt.Println("RRR s", s)
	fmt.Println("RRR nt", nt)
	var rvs []TypedValue
	if s.Values != nil {
		rvs = m.PopValues(len(s.NameExprs))
	}
	fmt.Println("RRR rvs", rvs)
	for i := 0; i < len(s.NameExprs); i++ {
		var tv TypedValue
		if rvs == nil {
			// NOTE: Go/Gno wart.
			// implicit interface casting could
			// requiring the consideration of the typed-nil case.
			if nt == nil {
				fmt.Println("RRR1", nt, nt == nil)
				tv = TypedValue{}
			} else {
				fmt.Println("RRR2", nt, nt == nil)
				tv = TypedValue{T: nt, V: defaultValue(nt)}
			}
		} else {
			tv = rvs[i]
		}
		if nt != nil {
			if nt.Kind() == InterfaceKind {
				if isUntyped(tv.T) {
					ConvertUntypedTo(&tv, nil)
				} else {
					// keep type as is.
				}
			} else {
				if isUntyped(tv.T) {
					ConvertUntypedTo(&tv, nt)
				} else {
					if debug {
						fmt.Println("RRR nt", nt)
						fmt.Println("RRR baseOf(nt)", baseOf(nt))
						fmt.Println("RRR tv.T", tv.T)
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
		} else if s.Const {
			// leave untyped as is.
		} else if isUntyped(tv.T) {
			ConvertUntypedTo(&tv, nil)
		}
		nx := s.NameExprs[i]
		ptr := lb.GetPointerTo(nx.Path)
		ptr.Assign2(m.Realm, tv, false)
	}
}

func (m *Machine) doOpTypeDecl() {
	s := m.PopStmt().(*TypeDecl)
	t := m.PopValue().GetType()
	tv := asValue(t)
	last := m.LastBlock()
	ptr := last.GetPointerTo(s.Path)
	ptr.Assign2(m.Realm, tv, false)
}
