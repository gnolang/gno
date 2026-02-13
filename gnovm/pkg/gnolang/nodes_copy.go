package gnolang

// Copy should happen before any preprocessing.
// * Attributes are not copied.
// * Paths are not copied.
// * *ConstExpr, *constTypeExpr, *bodyStmt not yet supported.

func (x *ConstExpr) Copy() Node {
	panic("*ConstExpr.Copy() not yet implemented")
}

func (x *constTypeExpr) Copy() Node {
	panic("*constTypeExpr.Copy() not yet implemented")
}

func (x *bodyStmt) Copy() Node {
	panic("*bodyStmt.Copy() not yet implemented")
}

func (x *NameExpr) Copy() Node {
	return &NameExpr{
		Name: x.Name,
	}
}

func (x *BasicLitExpr) Copy() Node {
	return &BasicLitExpr{
		Kind:  x.Kind,
		Value: x.Value,
	}
}

func (x *BinaryExpr) Copy() Node {
	return &BinaryExpr{
		Left:  x.Left.Copy().(Expr),
		Op:    x.Op,
		Right: x.Right.Copy().(Expr),
	}
}

func (x *CallExpr) Copy() Node {
	return &CallExpr{
		Func: x.Func.Copy().(Expr),
		Args: copyExprs(x.Args),
		Varg: x.Varg,
	}
}

func (x *IndexExpr) Copy() Node {
	return &IndexExpr{
		X:     x.X.Copy().(Expr),
		Index: x.Index.Copy().(Expr),
	}
}

func (x *SelectorExpr) Copy() Node {
	return &SelectorExpr{
		X:   x.X.Copy().(Expr),
		Sel: x.Sel,
	}
}

func (x *SliceExpr) Copy() Node {
	return &SliceExpr{
		X:    x.X.Copy().(Expr),
		Low:  copyExpr(x.Low),
		High: copyExpr(x.High),
		Max:  copyExpr(x.Max),
	}
}

func (x *StarExpr) Copy() Node {
	return &StarExpr{
		X: x.X.Copy().(Expr),
	}
}

func (x *RefExpr) Copy() Node {
	return &RefExpr{
		X: x.X.Copy().(Expr),
	}
}

func (x *TypeAssertExpr) Copy() Node {
	return &TypeAssertExpr{
		X:    x.X.Copy().(Expr),
		Type: x.Type.Copy().(Expr),
	}
}

func (x *UnaryExpr) Copy() Node {
	return &UnaryExpr{
		X:  x.X.Copy().(Expr),
		Op: x.Op,
	}
}

func (x *CompositeLitExpr) Copy() Node {
	return &CompositeLitExpr{
		Type: x.Type.Copy().(Expr),
		Elts: copyKVs(x.Elts),
	}
}

func (x *KeyValueExpr) Copy() Node {
	return &KeyValueExpr{
		Key:   copyExpr(x.Key),
		Value: x.Value.Copy().(Expr),
	}
}

func (fle *FuncLitExpr) Copy() Node {
	return &FuncLitExpr{
		Type: *(fle.Type.Copy().(*FuncTypeExpr)),
		Body: copyStmts(fle.Body),
	}
}

func (x *FieldTypeExpr) Copy() Node {
	return &FieldTypeExpr{
		NameExpr: *(x.NameExpr.Copy().(*NameExpr)),
		Type:     x.Type.Copy().(Expr),
		Tag:      copyExpr(x.Tag),
	}
}

func (x *ArrayTypeExpr) Copy() Node {
	return &ArrayTypeExpr{
		Len: copyExpr(x.Len),
		Elt: x.Elt.Copy().(Expr),
	}
}

func (x *SliceTypeExpr) Copy() Node {
	return &SliceTypeExpr{
		Elt: x.Elt.Copy().(Expr),
		Vrd: x.Vrd,
	}
}

func (x *InterfaceTypeExpr) Copy() Node {
	return &InterfaceTypeExpr{
		Methods: copyFTs(x.Methods),
	}
}

func (x *ChanTypeExpr) Copy() Node {
	return &ChanTypeExpr{
		Dir:   x.Dir,
		Value: x.Value.Copy().(Expr),
	}
}

func (x *FuncTypeExpr) Copy() Node {
	return &FuncTypeExpr{
		Params:  copyFTs(x.Params),
		Results: copyFTs(x.Results),
	}
}

func (x *MapTypeExpr) Copy() Node {
	return &MapTypeExpr{
		Key:   x.Key.Copy().(Expr),
		Value: x.Value.Copy().(Expr),
	}
}

func (x *StructTypeExpr) Copy() Node {
	return &StructTypeExpr{
		Fields: copyFTs(x.Fields),
	}
}

func (x *AssignStmt) Copy() Node {
	return &AssignStmt{
		Lhs: copyExprs(x.Lhs),
		Op:  x.Op,
		Rhs: copyExprs(x.Rhs),
	}
}

func (x *BlockStmt) Copy() Node {
	return &BlockStmt{
		Body: copyStmts(x.Body),
	}
}

func (x *BranchStmt) Copy() Node {
	return &BranchStmt{
		Op:    x.Op,
		Label: x.Label,
	}
}

func (x *DeclStmt) Copy() Node {
	return &DeclStmt{
		Body: copyStmts(x.Body),
	}
}

func (x *DeferStmt) Copy() Node {
	return &DeferStmt{
		Call: *(x.Call.Copy().(*CallExpr)),
	}
}

func (x *EmptyStmt) Copy() Node {
	return &EmptyStmt{}
}

func (x *ExprStmt) Copy() Node {
	return &ExprStmt{
		X: x.X.Copy().(Expr),
	}
}

func (x *ForStmt) Copy() Node {
	return &ForStmt{
		Init:      copyStmt(x.Init),
		Cond:      copyExpr(x.Cond),
		Post:      copyStmt(x.Post),
		BodyBlock: &BlockStmt{Body: copyStmts(x.GetBody())},
	}
}

func (x *GoStmt) Copy() Node {
	return &GoStmt{
		Call: *(x.Call.Copy().(*CallExpr)),
	}
}

func (x *IfStmt) Copy() Node {
	return &IfStmt{
		Init: copyStmt(x.Init),
		Cond: copyExpr(x.Cond),
		Then: *copyStmt(&x.Then).(*IfCaseStmt),
		Else: *copyStmt(&x.Else).(*IfCaseStmt),
	}
}

func (x *IfCaseStmt) Copy() Node {
	return &IfCaseStmt{
		Body: copyStmts(x.Body),
	}
}

func (x *IncDecStmt) Copy() Node {
	return &IncDecStmt{
		X:  x.X.Copy().(Expr),
		Op: x.Op,
	}
}

func (x *RangeStmt) Copy() Node {
	return &RangeStmt{
		X:     x.X.Copy().(Expr),
		Key:   copyExpr(x.Key),
		Value: copyExpr(x.Value),
		Op:    x.Op,
		Body:  copyStmts(x.Body),
	}
}

func (x *ReturnStmt) Copy() Node {
	return &ReturnStmt{
		Results: copyExprs(x.Results),
	}
}

func (x *SelectStmt) Copy() Node {
	return &SelectStmt{
		Cases: copySelectCases(x.Cases),
	}
}

func (x *SelectCaseStmt) Copy() Node {
	return &SelectCaseStmt{
		Comm: x.Comm.Copy().(Stmt),
		Body: copyStmts(x.Body),
	}
}

func (x *SendStmt) Copy() Node {
	return &SendStmt{
		Chan:  x.Chan.Copy().(Expr),
		Value: x.Value.Copy().(Expr),
	}
}

func (x *SwitchStmt) Copy() Node {
	return &SwitchStmt{
		Init:    copyStmt(x.Init),
		X:       x.X.Copy().(Expr),
		Clauses: copyCaseClauses(x.Clauses),
		VarName: x.VarName,
	}
}

func (x *SwitchClauseStmt) Copy() Node {
	return &SwitchClauseStmt{
		Cases: copyExprs(x.Cases),
		Body:  copyStmts(x.Body),
	}
}

func (x *FuncDecl) Copy() Node {
	funcDecl := &FuncDecl{
		NameExpr: *(x.NameExpr.Copy().(*NameExpr)),
		IsMethod: x.IsMethod,
		Type:     *(x.Type.Copy().(*FuncTypeExpr)),
		Body:     copyStmts(x.Body),
	}
	if x.IsMethod {
		funcDecl.Recv = *(x.Recv.Copy().(*FieldTypeExpr))
	}
	return funcDecl
}

func (x *ImportDecl) Copy() Node {
	return &ImportDecl{
		NameExpr: *(x.NameExpr.Copy().(*NameExpr)),
		PkgPath:  x.PkgPath,
	}
}

func (x *ValueDecl) Copy() Node {
	return &ValueDecl{
		NameExprs: copyNameExprs(x.NameExprs),
		Type:      copyExpr(x.Type),
		Values:    copyExprs(x.Values),
		Const:     x.Const,
	}
}

func (x *TypeDecl) Copy() Node {
	return &TypeDecl{
		NameExpr: *(x.NameExpr.Copy().(*NameExpr)),
		Type:     x.Type.Copy().(Expr),
		IsAlias:  x.IsAlias,
	}
}

func (fs *FileSet) CopyFileSet() *FileSet {
	files := make([]*FileNode, len(fs.Files))
	for i, file := range fs.Files {
		files[i] = file.Copy().(*FileNode)
	}
	return &FileSet{files}
}

func (x *FileNode) Copy() Node {
	return &FileNode{
		PkgName: x.PkgName,
		Decls:   copyDecls(x.Decls),
	}
}

func (pn *PackageNode) Copy() Node {
	return &PackageNode{
		PkgPath: pn.PkgPath,
		PkgName: pn.PkgName,
		FileSet: pn.FileSet.CopyFileSet(),
	}
}

// ----------------------------------------
// misc

func copyExpr(x Expr) Expr {
	if x == nil {
		return nil
	}
	return x.Copy().(Expr)
}

func copyExprs(xs []Expr) []Expr {
	res := make([]Expr, len(xs))
	for i, x := range xs {
		res[i] = x.Copy().(Expr)
	}
	return res
}

func copyNameExprs(nxs NameExprs) NameExprs {
	res := make([]NameExpr, len(nxs))
	for i, nx := range nxs {
		res[i] = *(nx.Copy().(*NameExpr))
	}
	return res
}

func copyKVs(kvs []KeyValueExpr) []KeyValueExpr {
	res := make([]KeyValueExpr, len(kvs))
	for i, kv := range kvs {
		res[i] = *(kv.Copy().(*KeyValueExpr))
	}
	return res
}

func copyStmt(x Stmt) Stmt {
	if x == nil {
		return nil
	}
	return x.Copy().(Stmt)
}

func copyStmts(ss []Stmt) []Stmt {
	res := make([]Stmt, len(ss))
	for i, s := range ss {
		res[i] = s.Copy().(Stmt)
	}
	return res
}

func copyFTs(fts []FieldTypeExpr) []FieldTypeExpr {
	res := make([]FieldTypeExpr, len(fts))
	for i, ft := range fts {
		res[i] = *(ft.Copy().(*FieldTypeExpr))
	}
	return res
}

func copyDecls(ds []Decl) []Decl {
	res := make([]Decl, len(ds))
	for i, d := range ds {
		res[i] = d.Copy().(Decl)
	}
	return res
}

func copySelectCases(scs []SelectCaseStmt) []SelectCaseStmt {
	res := make([]SelectCaseStmt, len(scs))
	for i, sc := range scs {
		res[i] = *(sc.Copy().(*SelectCaseStmt))
	}
	return res
}

func copyCaseClauses(scs []SwitchClauseStmt) []SwitchClauseStmt {
	res := make([]SwitchClauseStmt, len(scs))
	for i, sc := range scs {
		res[i] = *(sc.Copy().(*SwitchClauseStmt))
	}
	return res
}
