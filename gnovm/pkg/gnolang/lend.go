package gnolang

// Loans: a function literal written in
// place as the argument for a parameter its callee only calls or passes on to
// another such parameter writes as its maker while that call's frame is live.

// lendTok marks one frame that received a lent literal and lists the
// literals lent to it: their loans end when the frame is popped.
type lendTok struct{ lent []*FuncValue }

// lendArgsHook, when set by a test, runs after each lendArgs.
var lendArgsHook func(m *Machine)

// lendTag is a lent literal's loan, kept in Machine.lends only: a copy of
// the value has none, and nothing stores, encodes or hashes it.
type lendTag struct {
	tok      *lendTok
	depth    int // index in m.Frames of the frame holding tok
	realm    *Realm
	crossing *crossingCall
}

// markNonEscParams sets ATTR_NONESC_PARAMS on each FuncDecl: per parameter,
// true when every use is a call of it or a plain (not deferred) argument of a
// call that is no builtin or non-func conversion, possibly inside a literal used
// so. An interface parameter may also be type-switched into variables held to
// the same rule, or asserted and called at once.
func markNonEscParams(fn *FileNode) {
	for _, d := range fn.Decls {
		fd, ok := d.(*FuncDecl)
		if !ok || fd.Body == nil {
			continue
		}
		// Tracked variables: the parameters, then each type-switch variable over one.
		vars := map[helperKey]int{}
		for i, p := range fd.Type.Params {
			if p.Name != "" && p.Name != blankIdentifier {
				vars[helperKey{fd, p.Name}] = i
			}
		}
		res := make([]bool, len(fd.Type.Params))
		if len(vars) == 0 {
			fd.SetAttribute(ATTR_NONESC_PARAMS, res)
			continue
		}
		bad := map[int]bool{}
		inPlace := map[*FuncLitExpr]bool{}
		TranscribeB(fn, fd, func(ns []Node, stack []BlockNode, last BlockNode, ftype TransField, index int, n Node, stage TransStage) (Node, TransCtrl) {
			if stage != TRANS_ENTER {
				return n, TRANS_CONTINUE
			}
			switch n := n.(type) {
			case *FuncLitExpr:
				inPlace[n] = len(ns) > 0 && lendPlace(ns, ftype, index)
			case *NameExpr:
				if n.Path.Type != VPBlock {
					break
				}
				bn := last.GetBlockNodeForPath(nil, n.Path)
				pi, ok := vars[helperKey{bn, n.Name}]
				if !ok {
					break
				}
				if len(ns) == 0 || !lendUse(ns, ftype, index, vars, pi) {
					bad[pi] = true
					break
				}
				for i := len(stack) - 1; i >= 0 && stack[i] != bn; i-- {
					if fl, ok := stack[i].(*FuncLitExpr); ok && !inPlace[fl] {
						bad[pi] = true
					}
				}
			}
			return n, TRANS_CONTINUE
		})
		for _, i := range vars {
			res[i] = !bad[i]
		}
		fd.SetAttribute(ATTR_NONESC_PARAMS, res)
	}
}

// lendUse reports whether a tracked variable of parameter pi, whose parent is
// ns[len(ns)-1], is used where a lent value may be: a lend place, a func-type
// conversion in one, the subject of a type switch (its variable is then
// tracked too), or asserted and called at once. Any other use escapes.
func lendUse(ns []Node, ftype TransField, index int, vars map[helperKey]int, pi int) bool {
	if lendPlace(ns, ftype, index) {
		return true
	}
	switch p := ns[len(ns)-1].(type) {
	case *CallExpr: // T(x) with T a func type: the same value, retyped
		ct := convType(p.Func)
		if ct == nil || ftype != TRANS_CALL_ARG || len(p.Args) != 1 || ct.Kind() != FuncKind || len(ns) < 2 {
			return false
		}
		return lendPlaceOf(ns[:len(ns)-1], p)
	case *SwitchStmt:
		if ftype != TRANS_SWITCH_X || !p.IsTypeSwitch {
			return false
		}
		if p.VarName != "" && p.VarName != blankIdentifier {
			for i := range p.Clauses {
				vars[helperKey{&p.Clauses[i], p.VarName}] = pi
			}
		}
		return true
	case *TypeAssertExpr:
		if ftype != TRANS_TYPEASSERT_X || p.HasOK || len(ns) < 2 {
			return false
		}
		cx, ok := ns[len(ns)-2].(*CallExpr)
		return ok && cx.Func == p
	}
	return false
}

// convType returns the type x names when x is a conversion's callee, else nil.
func convType(x Expr) Type {
	switch fx := x.(type) {
	case *constTypeExpr:
		return fx.Type
	case *ConstExpr:
		if tv, ok := fx.V.(TypeValue); ok {
			return tv.Type
		}
	}
	return nil
}

// lendPlaceOf is lendPlace for x, a child of ns[len(ns)-1].
func lendPlaceOf(ns []Node, x Expr) bool {
	cx, ok := ns[len(ns)-1].(*CallExpr)
	if !ok {
		return false
	}
	if cx.Func == x {
		return lendPlace(ns, TRANS_CALL_FUNC, 0)
	}
	for i, a := range cx.Args {
		if a == x {
			return lendPlace(ns, TRANS_CALL_ARG, i)
		}
	}
	return false
}

// lendPlace reports whether the node whose parent is ns[len(ns)-1] sits where
// a lent value may: called (if callee is allowed) or a plain call argument.
func lendPlace(ns []Node, ftype TransField, index int) bool {
	cx, ok := ns[len(ns)-1].(*CallExpr)
	if !ok {
		return false
	}
	if ftype == TRANS_CALL_FUNC {
		return true
	}
	if ftype != TRANS_CALL_ARG || (cx.Varg && index == len(cx.Args)-1) {
		return false
	}
	if len(ns) > 1 {
		if _, ok := ns[len(ns)-2].(*DeferStmt); ok {
			return false
		}
	}
	switch fx := cx.Func.(type) {
	case *NameExpr, *SelectorExpr:
		return true
	case *ConstExpr:
		fv, ok := fx.V.(*FuncValue)
		return ok && fv.PkgPath != uversePkgPath
	}
	return false
}

// lendArgs runs in doOpCall once the callee's arguments are copied: a literal
// written in place for a non-escaping parameter is lent to this frame; a lent
// value handed to any other parameter loses its loan, and so does a lent value
// captured by a function argument that is not lent.
func (m *Machine) lendArgs(fr *Frame, fs BlockNode, args []TypedValue) {
	cx, _ := fr.Source.(*CallExpr)
	var nonesc []bool
	if fd, ok := fs.(*FuncDecl); ok {
		nonesc, _ = fd.GetAttribute(ATTR_NONESC_PARAMS).([]bool)
	}
	off := len(args) - len(nonesc)
	var cargs []Expr
	if cx != nil {
		cargs = cx.Args
	}
	// A bound receiver is in args but not in cx.Args; a method expression's
	// receiver is in both; a variadic slot is a slice, never a FuncValue.
	coff := 0
	if fr.Receiver.IsDefined() {
		coff = 1
	}
	for i := range args {
		if sv, ok := args[i].V.(*SliceValue); ok {
			m.dropSliceLends(sv) // a variadic element escapes like a stored value
			continue
		}
		afv, ok := args[i].V.(*FuncValue)
		if !ok {
			continue
		}
		pi := i - off
		okParam := nonesc != nil && pi >= 0 && pi < len(nonesc) && nonesc[pi]
		if m.lends[afv] != nil {
			if !okParam {
				delete(m.lends, afv)
			}
			continue
		}
		ci := i - coff
		lit := false
		if ci >= 0 && ci < len(cargs) {
			lit = isLitArg(cargs[ci])
		}
		if !okParam || !lit || fr.LastViewOnly || fr.LastRealm == nil {
			m.dropCapturedLends(afv)
			continue
		}
		if fr.lendTok == nil {
			fr.lendTok = &lendTok{}
		}
		if m.lends == nil {
			m.lends = map[*FuncValue]*lendTag{}
		}
		m.lends[afv] = &lendTag{tok: fr.lendTok, depth: len(m.Frames) - 1, realm: fr.LastRealm, crossing: fr.LastCrossing}
		fr.lendTok.lent = append(fr.lendTok.lent, afv)
	}
	if lendArgsHook != nil {
		lendArgsHook(m)
	}
}

// dropFrameLends ends the loans of frames m.Frames[from:], about to be popped
// (return, panic, goto): m.lends holds only loans of live frames.
func (m *Machine) dropFrameLends(from int) {
	if len(m.lends) == 0 {
		return
	}
	for i := from; i < len(m.Frames); i++ {
		tok := m.Frames[i].lendTok
		if tok == nil {
			continue
		}
		for _, fv := range tok.lent {
			if l := m.lends[fv]; l != nil && l.tok == tok {
				delete(m.lends, fv)
			}
		}
	}
}

// dropCapturedLends ends the loan of every lent value fv captured.
func (m *Machine) dropCapturedLends(fv *FuncValue) {
	if len(m.lends) == 0 {
		return
	}
	for i := range fv.Captures {
		hiv, ok := fv.Captures[i].V.(*HeapItemValue)
		if !ok {
			continue
		}
		if cfv, ok := hiv.Value.V.(*FuncValue); ok {
			delete(m.lends, cfv)
		}
	}
}

// lendLive reports whether fv is a lent literal whose loan still holds: the
// frame it was lent to is live, and its maker's realm is writing, uncrossed.
func (m *Machine) lendLive(fv *FuncValue) bool {
	l := m.lends[fv]
	return l != nil && !m.viewOnly && m.Realm != nil && m.Realm == l.realm &&
		m.curCrossing == l.crossing && l.depth < len(m.Frames) &&
		m.Frames[l.depth].lendTok == l.tok
}

// isLitArg reports whether x is a function literal, possibly converted to a
// named function type (preprocess inserts that conversion).
func isLitArg(x Expr) bool {
	if cx, ok := x.(*CallExpr); ok && len(cx.Args) == 1 {
		if _, ok := cx.Func.(*constTypeExpr); ok {
			x = cx.Args[0]
		}
	}
	_, ok := x.(*FuncLitExpr)
	return ok
}

// dropSliceLends ends the loan of every lent value held in sv's elements.
func (m *Machine) dropSliceLends(sv *SliceValue) {
	if len(m.lends) == 0 || sv.Base == nil {
		return
	}
	av, ok := sv.Base.(*ArrayValue)
	if !ok {
		return
	}
	for i := sv.Offset; i < sv.Offset+sv.Length && i < len(av.List); i++ {
		if fv, ok := av.List[i].V.(*FuncValue); ok {
			delete(m.lends, fv)
			m.dropCapturedLends(fv)
		}
	}
}
