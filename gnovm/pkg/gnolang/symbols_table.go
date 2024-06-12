package gnolang

type SymbolTable struct {
	scopes []*Scope
}

type Scope struct {
	symbols map[string]struct{}
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		scopes: []*Scope{newScope()},
	}
}

func newScope() *Scope {
	return &Scope{
		symbols: make(map[string]struct{}),
	}
}

func (st *SymbolTable) EnterScope() {
	st.scopes = append(st.scopes, newScope())
}

func (st *SymbolTable) ExitScope() {
	if len(st.scopes) > 1 {
		st.scopes = st.scopes[:len(st.scopes)-1]
	}
}

func (st *SymbolTable) AddIdentifier(name string) {
	if len(st.scopes) > 0 {
		st.scopes[len(st.scopes)-1].symbols[name] = struct{}{}
	}
}

func (st *SymbolTable) IdentifierExists(name string) bool {
	for i := len(st.scopes) - 1; i >= 0; i-- {
		if _, exists := st.scopes[i].symbols[name]; exists {
			return true
		}
	}
	return false
}
