package gnolang

// Local helpers: a local declared once from a literal written here (or
// `var f T` then one `f = literal`, for recursion) and only ever called never
// travels, so calls through it are not calls by value. Runs after preprocess1:
// a name is keyed by the variable its path resolves to, not its spelling.

type helperKey struct {
	bn   BlockNode
	name Name
}

type helperUse struct {
	defs  int  // := or var declarations of this very variable
	lit   bool // every declaration gives it a function literal
	bad   bool // used other than as a callee, or assigned
	calls []*CallExpr
	caps  []*FuncLitExpr // literals between its declaring block and a use
	bare  int            // `var f T` declarations without a value
	asg   int            // `f = literal` statements in the declaring block
}

func markLocalHelpers(fn *FileNode) {
	for _, d := range fn.Decls {
		switch d := d.(type) {
		case *FuncDecl:
			markLocalHelpersIn(fn, d)
		case *ValueDecl:
			// A package-level initialiser's literals hold helpers too.
			if !d.Const {
				markLocalHelpersIn(fn, d)
			}
		}
	}
}

func markLocalHelpersIn(fn *FileNode, root Decl) {
	uses := map[helperKey]*helperUse{}
	inPlace := map[*FuncLitExpr]bool{}     // called where written, or crossing
	litOf := map[*FuncLitExpr]*helperUse{} // literal declaring a helper
	// Parameters, receiver and named results are declared without a NameExpr,
	// so a later `:=` reusing one would count as its only declaration: fail closed.
	seedParams := func(bn BlockNode, ft *FuncTypeExpr) {
		for _, fs := range [][]FieldTypeExpr{ft.Params, ft.Results} {
			for _, f := range fs {
				uses[helperKey{bn, f.Name}] = &helperUse{bad: true}
			}
		}
	}
	if fd, ok := root.(*FuncDecl); ok {
		seedParams(fd, &fd.Type)
		if fd.IsMethod {
			uses[helperKey{fd, fd.Recv.Name}] = &helperUse{bad: true}
		}
	}
	TranscribeB(fn, root, func(ns []Node, stack []BlockNode, last BlockNode, ftype TransField, index int, n Node, stage TransStage) (Node, TransCtrl) {
		if stage != TRANS_ENTER {
			return n, TRANS_CONTINUE
		}
		switch n := n.(type) {
		case *FuncLitExpr:
			ft, _ := n.Type.GetAttribute(ATTR_TYPE_VALUE).(*FuncType)
			_, isCall := ns[len(ns)-1].(*CallExpr)
			inPlace[n] = ft != nil && ft.IsCrossing() || isCall && ftype == TRANS_CALL_FUNC
			seedParams(n, &n.Type)
		case *NameExpr:
			if n.Name == blankIdentifier || n.Path.Type != VPBlock {
				return n, TRANS_CONTINUE
			}
			dbn := last.GetBlockNodeForPath(nil, n.Path)
			k := helperKey{dbn, n.Name}
			u := uses[k]
			if u == nil {
				u = &helperUse{lit: true}
				uses[k] = u
			}
			// Every literal entered since the declaring block captures it.
			for i := len(stack) - 1; i >= 0 && stack[i] != dbn; i-- {
				if fl, ok := stack[i].(*FuncLitExpr); ok {
					u.caps = append(u.caps, fl)
				}
			}
			if len(ns) == 0 {
				u.bad = true
				break
			}
			switch parent := ns[len(ns)-1].(type) {
			case *CallExpr:
				if ftype == TRANS_CALL_FUNC {
					u.calls = append(u.calls, parent)
				} else {
					u.bad = true
				}
			case *AssignStmt:
				// One plain `f = literal` in the declaring block; any other assignment is bad.
				if ftype == TRANS_ASSIGN_LHS && parent.Op == ASSIGN && last == dbn && len(parent.Lhs) == 1 && len(parent.Rhs) == 1 {
					u.asg++
					u.lit = u.lit && litDef(parent.Rhs[0], u, litOf)
					break
				}
				if ftype != TRANS_ASSIGN_LHS || parent.Op != DEFINE {
					u.bad = true
					break
				}
				u.defs++
				u.lit = u.lit && len(parent.Lhs) == len(parent.Rhs) && litDef(parent.Rhs[index], u, litOf)
			case *ValueDecl:
				if ftype != TRANS_VAR_NAME || parent.Const {
					u.bad = true
					break
				}
				u.defs++
				if len(parent.Values) == 0 {
					u.bare++
					break
				}
				u.lit = u.lit && len(parent.Values) == len(parent.NameExprs) && litDef(parent.Values[index], u, litOf)
			default:
				u.bad = true
			}
		}
		return n, TRANS_CONTINUE
	})
	// Good: declared once from a literal, only called, and every literal
	// capturing it stays (called in place, crossing, or a good helper's own
	// literal). A := literal cannot name what is declared later: no cycle.
	good := func(u *helperUse) bool { return u.defs == 1 && u.lit && !u.bad && u.asg == u.bare }
	for changed := true; changed; {
		changed = false
		for _, u := range uses {
			if !good(u) {
				continue
			}
			for _, fl := range u.caps {
				if owner := litOf[fl]; !inPlace[fl] && (owner == nil || !good(owner)) {
					u.bad, changed = true, true
					break
				}
			}
		}
	}
	for _, u := range uses {
		if good(u) {
			for _, cx := range u.calls {
				cx.SetAttribute(ATTR_LOCAL_HELPER, true)
			}
		}
	}
}

// litDef records x as the literal declaring u; false if x is not a literal.
func litDef(x Expr, u *helperUse, litOf map[*FuncLitExpr]*helperUse) bool {
	fl, ok := x.(*FuncLitExpr)
	if ok {
		litOf[fl] = u
	}
	return ok
}
