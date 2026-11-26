package gnolang

import (
	"fmt"
	"sort"
)

// LendWarning is one `gno lint` warning: a function value that writes the
// realm's package state but sits where the VM will run it read-only.
type LendWarning struct {
	File         string
	Line, Column int
	Msg          string
}

// LentCallbackWarnings checks a preprocessed realm package. It only reads the
// marks preprocess left (ATTR_NONESC_PARAMS, ATTR_NAMED_CALLEE,
// ATTR_LOCAL_HELPER); the run-time rule stays the only enforcement.
func LentCallbackWarnings(store Store, pn *PackageNode) (ws []LendWarning) {
	if pn == nil || pn.FileSet == nil || !IsRealmPath(pn.PkgPath) {
		return nil
	}
	defer func() {
		if r := recover(); r != nil {
			ws = nil // a lint aid never fails lint
		}
	}()
	c := &lendLint{store: store, pn: pn, bodies: map[BlockNode]*llBody{}}
	for _, fn := range pn.FileSet.Files {
		c.collectLocals(fn)
		for _, d := range fn.Decls {
			c.summarize(fn, d)
		}
	}
	c.fixpoint()
	for _, fn := range pn.FileSet.Files {
		for _, d := range fn.Decls {
			c.place(fn, d)
		}
	}
	sort.SliceStable(c.ws, func(i, j int) bool {
		a, b := c.ws[i], c.ws[j]
		if a.File != b.File {
			return a.File < b.File
		}
		if a.Line != b.Line {
			return a.Line < b.Line
		}
		return a.Column < b.Column
	})
	return c.ws
}

type llBody struct {
	writes  Name // first package variable written, directly or by a callee
	callees []BlockNode
}

type lendLint struct {
	store  Store
	pn     *PackageNode
	bodies map[BlockNode]*llBody
	// Local names whose calls preprocess marked as by name or as a helper,
	// and the one value each was declared with.
	named, helper map[helperKey]bool
	defs          map[helperKey]Expr
	ws            []LendWarning
}

func (c *lendLint) body(b BlockNode) *llBody {
	if c.bodies[b] == nil {
		c.bodies[b] = &llBody{}
	}
	return c.bodies[b]
}

// localKey keys a block name the way markLocalHelpers does.
func localKey(last BlockNode, nx *NameExpr) (helperKey, bool) {
	if nx.Path.Type != VPBlock || nx.Name == blankIdentifier {
		return helperKey{}, false
	}
	dbn := last.GetBlockNodeForPath(nil, nx.Path)
	switch dbn.(type) {
	case *PackageNode, *FileNode:
		return helperKey{}, false
	}
	names := dbn.GetBlockNames()
	if int(nx.Path.Index) >= len(names) {
		return helperKey{}, false
	}
	name := names[nx.Path.Index]
	// A capture ("~x") keys as the variable it captures, as preprocess marks it.
	if fle, ok := dbn.(*FuncLitExpr); ok && len(name) > 1 && name[0] == '~' {
		for i := range fle.HeapCaptures {
			if cx := &fle.HeapCaptures[i]; cx.Name == name[1:] {
				if parent, ok := fle.GetParentNode(nil).(BlockNode); ok {
					return localKey(parent, &NameExpr{Name: cx.Name, Path: cx.Path})
				}
			}
		}
		return helperKey{}, false
	}
	return helperKey{dbn, name}, true
}

func (c *lendLint) collectLocals(fn *FileNode) {
	if c.named == nil {
		c.named, c.helper, c.defs = map[helperKey]bool{}, map[helperKey]bool{}, map[helperKey]Expr{}
	}
	for _, d := range fn.Decls {
		TranscribeB(fn, d, func(ns []Node, stack []BlockNode, last BlockNode, ftype TransField, index int, n Node, stage TransStage) (Node, TransCtrl) {
			if stage != TRANS_ENTER {
				return n, TRANS_CONTINUE
			}
			nx, ok := n.(*NameExpr)
			if !ok || len(ns) == 0 {
				return n, TRANS_CONTINUE
			}
			k, ok := localKey(last, nx)
			if !ok {
				return n, TRANS_CONTINUE
			}
			switch p := ns[len(ns)-1].(type) {
			case *CallExpr:
				if ftype == TRANS_CALL_FUNC {
					if p.GetAttribute(ATTR_NAMED_CALLEE) == true {
						c.named[k] = true
					}
					if p.GetAttribute(ATTR_LOCAL_HELPER) == true {
						c.helper[k] = true
					}
				}
			case *AssignStmt:
				if ftype == TRANS_ASSIGN_LHS && len(p.Lhs) == len(p.Rhs) {
					c.defs[k] = p.Rhs[index]
				}
			case *ValueDecl:
				if ftype == TRANS_VAR_NAME && len(p.Values) == len(p.NameExprs) {
					c.defs[k] = p.Values[index]
				}
			}
			return n, TRANS_CONTINUE
		})
	}
}

// summarize records, for each function body under d, the first package
// variable it writes and the same-package plain functions it calls by name.
// A nested literal's body counts for every enclosing body too.
func (c *lendLint) summarize(fn *FileNode, d Decl) {
	TranscribeB(fn, d, func(ns []Node, stack []BlockNode, last BlockNode, ftype TransField, index int, n Node, stage TransStage) (Node, TransCtrl) {
		if stage != TRANS_ENTER {
			return n, TRANS_CONTINUE
		}
		var encl []BlockNode
		for _, b := range stack {
			switch b.(type) {
			case *FuncDecl, *FuncLitExpr:
				encl = append(encl, b)
			}
		}
		switch n.(type) {
		case *FuncDecl, *FuncLitExpr:
			c.body(n.(BlockNode))
		}
		if len(encl) == 0 {
			return n, TRANS_CONTINUE
		}
		var targets []Expr
		switch n := n.(type) {
		case *AssignStmt:
			if n.Op != DEFINE {
				targets = n.Lhs
			}
		case *IncDecStmt:
			targets = []Expr{n.X}
		case *CallExpr:
			if nx, ok := n.Func.(*NameExpr); ok && nx.Path.Type == VPUverse && len(n.Args) > 0 {
				switch nx.Name {
				case "delete", "clear", "copy":
					targets = []Expr{n.Args[0]}
				}
			}
			if n.WithCross {
				break
			}
			if fd, path := c.calleeDecl(last, n); fd != nil && path == c.pn.PkgPath && !fd.isCrossing() {
				for _, b := range encl {
					c.body(b).callees = append(c.body(b).callees, fd)
				}
			}
		}
		for _, t := range targets {
			if name := c.pkgVarRoot(last, t); name != "" {
				for _, b := range encl {
					if bd := c.body(b); bd.writes == "" {
						bd.writes = name
					}
				}
			}
		}
		return n, TRANS_CONTINUE
	})
}

func (fd *FuncDecl) isCrossing() bool {
	ft, ok := fd.Type.GetAttribute(ATTR_TYPE_VALUE).(*FuncType)
	return ok && ft.IsCrossing()
}

func (c *lendLint) fixpoint() {
	for changed := true; changed; {
		changed = false
		for _, bd := range c.bodies {
			if bd.writes != "" {
				continue
			}
			for _, cb := range bd.callees {
				if w := c.body(cb).writes; w != "" {
					bd.writes, changed = w, true
					break
				}
			}
		}
	}
}

// pkgVarRoot returns the package variable of this realm that x, an
// assignment target, is rooted at (through fields, elements, pointers).
func (c *lendLint) pkgVarRoot(last BlockNode, x Expr) Name {
	for {
		switch e := x.(type) {
		case *NameExpr:
			if e.Path.Type != VPBlock || e.Name == blankIdentifier {
				return ""
			}
			switch last.GetBlockNodeForPath(c.store, e.Path).(type) {
			case *PackageNode, *FileNode:
				return e.Name
			}
			return ""
		case *SelectorExpr:
			if c.ownPkgRef(e.X) {
				return e.Sel
			}
			x = e.X
		case *IndexExpr:
			x = e.X
		case *StarExpr:
			x = e.X
		default:
			return ""
		}
	}
}

// pkgRefPath returns the package path x refers to, if x is a package.
func pkgRefPath(x Expr) string {
	cx, ok := x.(*ConstExpr)
	if !ok {
		return ""
	}
	switch v := cx.V.(type) {
	case RefValue:
		return v.PkgPath
	case *PackageValue:
		return v.PkgPath
	}
	return ""
}

func (c *lendLint) ownPkgRef(x Expr) bool {
	p := pkgRefPath(x)
	return p != "" && p == c.pn.PkgPath
}

func (c *lendLint) pkgNode(path string) *PackageNode {
	if path == c.pn.PkgPath {
		return c.pn
	}
	pn, _ := c.store.GetBlockNode(PackageNodeLocation(path)).(*PackageNode)
	return pn
}

// topFuncDecl returns the non-method function declared as name in path.
func (c *lendLint) topFuncDecl(path string, name Name) *FuncDecl {
	pn := c.pkgNode(path)
	if pn == nil || pn.FileSet == nil {
		return nil
	}
	_, d, ok := pn.FileSet.GetDeclForSafe(name)
	if !ok || d == nil {
		return nil
	}
	fd, _ := (*d).(*FuncDecl)
	if fd != nil && fd.IsMethod {
		return nil
	}
	return fd
}

func (c *lendLint) staticType(last BlockNode, x Expr) (t Type) {
	defer func() {
		if recover() != nil {
			t = nil
		}
	}()
	if ct := convType(x); ct != nil {
		return ct // a type: method expression T.M
	}
	// Inside a literal a package name reads as pkgref.name.
	if sx, ok := x.(*SelectorExpr); ok && pkgRefPath(sx.X) != "" {
		if pn := c.pkgNode(pkgRefPath(sx.X)); pn != nil {
			return pn.GetStaticTypeOf(c.store, sx.Sel)
		}
	}
	return evalStaticTypeOf(c.store, last, x)
}

// declOf resolves x, a callee or a function used as a value, to the
// function declaration it names and that declaration's package path.
func (c *lendLint) declOf(last BlockNode, x Expr) (*FuncDecl, string) {
	switch e := x.(type) {
	case *FuncHandleExpr:
		return c.declOf(last, e.X)
	case *ConstExpr:
		if fv, ok := e.V.(*FuncValue); ok && fv.PkgPath != uversePkgPath {
			fd, _ := fv.GetSource(c.store).(*FuncDecl)
			return fd, fv.PkgPath
		}
	case *NameExpr:
		if e.Path.Type != VPBlock {
			return nil, ""
		}
		if bn := last.GetBlockNodeForPath(c.store, e.Path); bn != nil {
			if pn, ok := bn.(*PackageNode); ok || isFileNode(bn) {
				if !ok {
					pn = packageOf(bn)
				}
				return c.topFuncDecl(pn.PkgPath, e.Name), pn.PkgPath
			}
		}
		// A local bound once to a declared function, called by name.
		if k, ok := localKey(last, e); ok && c.named[k] && c.defs[k] != nil {
			return c.declOf(last, c.defs[k])
		}
	case *SelectorExpr:
		if p := pkgRefPath(e.X); p != "" {
			return c.topFuncDecl(p, e.Sel), p
		}
		if !e.Path.IsMethod() || e.Path.Type == VPInterface || e.Path.Type == VPDerefInterface || e.Path.Depth > 1 {
			return nil, ""
		}
		t := c.staticType(last, e.X)
		if pt, ok := t.(*PointerType); ok {
			t = pt.Elt
		}
		dt, ok := t.(*DeclaredType)
		if !ok {
			return nil, ""
		}
		for _, mv := range dt.Methods {
			if fv, ok := mv.V.(*FuncValue); ok && fv.Name == e.Sel {
				fd, _ := fv.GetSource(c.store).(*FuncDecl)
				return fd, dt.PkgPath
			}
		}
	}
	return nil, ""
}

func isFileNode(bn BlockNode) bool {
	_, ok := bn.(*FileNode)
	return ok
}

func (c *lendLint) calleeDecl(last BlockNode, cx *CallExpr) (*FuncDecl, string) {
	return c.declOf(last, cx.Func)
}

// Why a function value runs read-only where it sits, and the fix to name.
const (
	llStored    = "it is kept to be called later, so it runs read-only (call it where it is written, pass it in place to a function called by name, or make it func(cur realm) and call it with cross)"
	llReturned  = "it is returned, so it runs read-only (make it func(cur realm) and call it with cross)"
	llDeferred  = "it is the argument of a deferred call, so it runs read-only (defer a literal that makes the call, or make it func(cur realm))"
	llVariadic  = "it goes in a variadic list, so it runs read-only (pass it as a plain parameter, or make it func(cur realm))"
	llNotNamed  = "the function it is passed to is not called by name, so it runs read-only (call a declared function by name, or make it func(cur realm))"
	llOther     = "it is passed to another realm, so it runs read-only (make it func(cur realm) and call it with cross)"
	llEscaping  = "the function it is passed to keeps it or hands it on, so it runs read-only (pass it to a function that only calls it, or make it func(cur realm))"
	llValue     = "it is used as a value, so it runs read-only (call it by name, or make it func(cur realm) and call it with cross)"
	llOtherSpot = "it is not passed in place to a function called by name, so it runs read-only (call it where it is written, or make it func(cur realm))"
)

func (c *lendLint) warn(fn *FileNode, n Node, what string, w Name, why string) {
	c.ws = append(c.ws, LendWarning{
		File:   string(fn.FileName),
		Line:   n.GetLine(),
		Column: n.GetColumn(),
		Msg:    fmt.Sprintf("%s writes %s but %s", what, w, why),
	})
}

// place flags each writing literal or declared-function value that the VM
// will run read-only where it sits.
func (c *lendLint) place(fn *FileNode, d Decl) {
	TranscribeB(fn, d, func(ns []Node, stack []BlockNode, last BlockNode, ftype TransField, index int, n Node, stage TransStage) (Node, TransCtrl) {
		if stage != TRANS_ENTER {
			return n, TRANS_CONTINUE
		}
		switch n := n.(type) {
		case *FuncLitExpr:
			ft, _ := n.Type.GetAttribute(ATTR_TYPE_VALUE).(*FuncType)
			if ft != nil && ft.IsCrossing() {
				return n, TRANS_CONTINUE
			}
			w := c.body(n).writes
			if w == "" {
				return n, TRANS_CONTINUE
			}
			if why := c.litPlace(ns, last, ftype, index); why != "" {
				c.warn(fn, n, "this function", w, why)
			}
		case *FuncHandleExpr:
			// A declared function or method expression used as a value.
			fd, path := c.declOf(last, n)
			if fd == nil || path != c.pn.PkgPath || fd.isCrossing() || c.body(fd).writes == "" {
				break
			}
			if c.aliasDef(ns, last, ftype, index) {
				break
			}
			c.warn(fn, n, string(fd.Name), c.body(fd).writes, llValue)
		case *SelectorExpr:
			// A method value of this realm's type (not called here).
			if len(ns) == 0 || !n.Path.IsMethod() || n.Path.Type == VPInterface || n.Path.Type == VPDerefInterface {
				break
			}
			if cx, ok := ns[len(ns)-1].(*CallExpr); ok && ftype == TRANS_CALL_FUNC && cx != nil {
				break
			}
			if _, ok := ns[len(ns)-1].(*FuncHandleExpr); ok {
				break
			}
			if _, isType := c.staticType(last, n.X).(*TypeType); isType || convType(n.X) != nil {
				break
			}
			fd, path := c.declOf(last, n)
			if fd == nil || path != c.pn.PkgPath || fd.isCrossing() || c.body(fd).writes == "" {
				break
			}
			c.warn(fn, n, string(fd.Name), c.body(fd).writes, llValue)
		}
		return n, TRANS_CONTINUE
	})
}

// aliasDef reports whether the node sits as the value declaring a local that
// preprocess marked as called by name (`f := F`, only ever called).
func (c *lendLint) aliasDef(ns []Node, last BlockNode, ftype TransField, index int) bool {
	k, ok := c.declaredLocal(ns, last, ftype, index)
	return ok && c.named[k]
}

// declaredLocal returns the local whose declaration or assignment has the
// node as its value.
func (c *lendLint) declaredLocal(ns []Node, last BlockNode, ftype TransField, index int) (helperKey, bool) {
	if len(ns) == 0 {
		return helperKey{}, false
	}
	switch p := ns[len(ns)-1].(type) {
	case *AssignStmt:
		if ftype == TRANS_ASSIGN_RHS && len(p.Lhs) == len(p.Rhs) {
			if nx, ok := p.Lhs[index].(*NameExpr); ok {
				return localKey(last, nx)
			}
		}
	case *ValueDecl:
		if ftype == TRANS_VAR_VALUE && len(p.NameExprs) == len(p.Values) {
			return localKey(last, &p.NameExprs[index])
		}
	}
	return helperKey{}, false
}

// litPlace returns "" when a literal written here may write (called where
// written, a local helper, or lent in place), else why it runs read-only.
func (c *lendLint) litPlace(ns []Node, last BlockNode, ftype TransField, index int) string {
	if len(ns) == 0 {
		return llStored
	}
	parent := ns[len(ns)-1]
	switch p := parent.(type) {
	case *CallExpr:
		if ftype == TRANS_CALL_FUNC {
			return ""
		}
	case *AssignStmt, *ValueDecl:
		if k, ok := c.declaredLocal(ns, last, ftype, index); ok && c.helper[k] {
			return ""
		}
		return llStored
	case *ReturnStmt:
		return llReturned
	default:
		_ = p
		return llStored
	}
	// Strip a conversion to a named func type.
	cx := parent.(*CallExpr)
	if ct := convType(cx.Func); ct != nil {
		if ct.Kind() != FuncKind || len(ns) < 2 {
			return llStored
		}
		gp, ok := ns[len(ns)-2].(*CallExpr)
		if !ok {
			return llStored
		}
		found := false
		for i, a := range gp.Args {
			if a == cx {
				ftype, index, found = TRANS_CALL_ARG, i, true
			}
		}
		if !found {
			return llStored
		}
		ns, cx = ns[:len(ns)-1], gp
	}
	if _, ok := cx.Func.(*IndexExpr); ok && cx.GetAttribute(ATTR_NAMED_CALLEE) == true {
		return "" // a library table entry called in place: decided at run time
	}
	if !lendPlace(ns, ftype, index) {
		if ftype != TRANS_CALL_ARG {
			return llOtherSpot
		}
		if cx.Varg && index == len(cx.Args)-1 {
			return llVariadic
		}
		if len(ns) > 1 {
			if _, ok := ns[len(ns)-2].(*DeferStmt); ok {
				return llDeferred
			}
		}
		return llNotNamed
	}
	if cx.WithCross {
		return "" // a crossing call: decided at run time
	}
	if sx, ok := cx.Func.(*SelectorExpr); ok && (sx.Path.Type == VPInterface || sx.Path.Type == VPDerefInterface) {
		return "" // interface method: decided at run time
	}
	if nx, ok := cx.Func.(*NameExpr); ok {
		if _, local := localKey(last, nx); local && cx.GetAttribute(ATTR_NAMED_CALLEE) != true {
			return llNotNamed // a local, a parameter or a helper literal: never a lending callee
		}
	}
	fd, cpath := c.calleeDecl(last, cx)
	if fd == nil {
		return "" // not resolved here: decided at run time
	}
	if cpath != c.pn.PkgPath && IsRealmPath(cpath) {
		return llOther
	}
	if cpath != c.pn.PkgPath && cx.GetAttribute(ATTR_NAMED_CALLEE) != true {
		_, isConst := cx.Func.(*ConstExpr)
		sx, isSel := cx.Func.(*SelectorExpr)
		if !isConst && !(isSel && sx.Path.IsMethod()) {
			return llNotNamed
		}
	}
	pi := index
	if sx, ok := cx.Func.(*SelectorExpr); ok && sx.Path.IsMethod() {
		if _, isType := c.staticType(last, sx.X).(*TypeType); isType || convType(sx.X) != nil {
			pi = index - 1 // method expression: the receiver is an argument
		}
	}
	nonesc, _ := fd.GetAttribute(ATTR_NONESC_PARAMS).([]bool)
	if ft, ok := fd.Type.GetAttribute(ATTR_TYPE_VALUE).(*FuncType); ok && ft.HasVarg() && pi >= len(ft.Params)-1 {
		return llVariadic
	}
	if pi < 0 || pi >= len(nonesc) || !nonesc[pi] {
		return llEscaping
	}
	return ""
}
