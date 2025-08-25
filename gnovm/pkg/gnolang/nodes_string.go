package gnolang

import (
	"fmt"
)

// ----------------------------------------
// Word.TokenString()

var wordTokenStrings = map[Word]string{
	// Operations
	ADD:             "+",
	SUB:             "-",
	MUL:             "*",
	QUO:             "/",
	REM:             "%",
	BAND:            "&",
	BOR:             "|",
	XOR:             "^",
	SHL:             "<<",
	SHR:             ">>",
	BAND_NOT:        "&^",
	ADD_ASSIGN:      "+=",
	SUB_ASSIGN:      "-=",
	MUL_ASSIGN:      "*=",
	QUO_ASSIGN:      "/=",
	REM_ASSIGN:      "%=",
	BAND_ASSIGN:     "&=",
	BOR_ASSIGN:      "|=",
	XOR_ASSIGN:      "^=",
	SHL_ASSIGN:      "<<=",
	SHR_ASSIGN:      ">>=",
	BAND_NOT_ASSIGN: "&^=",
	LAND:            "&&",
	LOR:             "||",
	ARROW:           "<-",
	INC:             "++",
	DEC:             "--",
	EQL:             "==",
	LSS:             "<",
	GTR:             ">",
	ASSIGN:          "=",
	NOT:             "!",
	NEQ:             "!=",
	LEQ:             "<=",
	GEQ:             ">=",
	DEFINE:          ":=",

	// Branch operations
	BREAK:       "break",
	CONTINUE:    "continue",
	FALLTHROUGH: "fallthrough",
	GOTO:        "goto",
}

func (w Word) TokenString() string {
	s, ok := wordTokenStrings[w]
	if !ok {
		panic(fmt.Sprintf("no token repr for %s (%d)", w.String(), w))
	}
	return s
}

// ----------------------------------------
// Node.String()

func (vp ValuePath) String() string {
	switch vp.Type {
	case VPInvalid:
		// index doesn't matter but useful for debugging.
		return fmt.Sprintf("VPInvalid(%d)", vp.Index)
	case VPUverse:
		return fmt.Sprintf("VPUverse(%d)", vp.Index)
	case VPBlock:
		return fmt.Sprintf("VPBlock(%d,%d)", vp.Depth, vp.Index)
	case VPField:
		return fmt.Sprintf("VPField(%d,%d,%s)", vp.Depth, vp.Index, vp.Name)
	case VPSubrefField:
		return fmt.Sprintf("VPSubrefField(%d,%d,%s)", vp.Depth, vp.Index, vp.Name)
	case VPValMethod:
		return fmt.Sprintf("VPValMethod(%d,%s)", vp.Index, vp.Name)
	case VPPtrMethod:
		return fmt.Sprintf("VPPtrMethod(%d,%s)", vp.Index, vp.Name)
	case VPInterface:
		return fmt.Sprintf("VPInterface(%s)", vp.Name)
	case VPDerefField:
		return fmt.Sprintf("VPDerefField(%d,%d,%s)", vp.Depth, vp.Index, vp.Name)
	case VPDerefValMethod:
		return fmt.Sprintf("VPDerefValMethod(%d,%s)", vp.Index, vp.Name)
	case VPDerefPtrMethod:
		return fmt.Sprintf("VPDerefPtrMethod(%d,%s)", vp.Index, vp.Name)
	case VPDerefInterface:
		return fmt.Sprintf("VPDerefInterface(%s)", vp.Name)
	default:
		panic("illegal_value_type")
	}
}

func (x NameExpr) String() string {
	switch x.Type {
	case NameExprTypeNormal:
		return fmt.Sprintf("%s<%s>", x.Name, x.Path.String())
	case NameExprTypeDefine:
		return fmt.Sprintf("%s<!%s>", x.Name, x.Path.String())
	case NameExprTypeHeapDefine:
		return fmt.Sprintf("%s<!~%s>", x.Name, x.Path.String())
	case NameExprTypeHeapUse:
		return fmt.Sprintf("%s<~%s>", x.Name, x.Path.String())
	case NameExprTypeHeapClosure:
		return fmt.Sprintf("%s<()~%s>", x.Name, x.Path.String())
	default:
		panic("unexpected NameExpr type")
	}
}

func (x BasicLitExpr) String() string {
	return x.Value
}

func (x BinaryExpr) String() string {
	return fmt.Sprintf("%s %s %s",
		x.Left.String(),
		x.Op.TokenString(),
		x.Right.String(),
	)
}

func (x CallExpr) String() string {
	if x.Varg {
		return fmt.Sprintf("%s(%s...)", x.Func, x.Args.String())
	}
	return fmt.Sprintf("%s(%s)", x.Func, x.Args.String())
}

func (x IndexExpr) String() string {
	return fmt.Sprintf("%s[%s]", x.X, x.Index)
}

func (x SelectorExpr) String() string {
	return fmt.Sprintf("%s.%s", x.X, x.Sel)
}

func (x SliceExpr) String() string {
	ls, hs, ms := "", "", ""
	if x.Low != nil {
		ls = x.Low.String()
	}
	if x.High != nil {
		hs = x.High.String()
	}
	if x.Max != nil {
		ms = x.Max.String()
	}
	if ms == "" {
		return fmt.Sprintf("%s[%s:%s]", x.X, ls, hs)
	}
	return fmt.Sprintf("%s[%s:%s:%s]", x.X, ls, hs, ms)
}

func (x StarExpr) String() string {
	return fmt.Sprintf("*(%s)", x.X)
}

func (x RefExpr) String() string {
	return fmt.Sprintf("&(%s)", x.X)
}

func (x TypeAssertExpr) String() string {
	if x.Type == nil {
		return fmt.Sprintf("%s.(type)", x.X)
	}
	return fmt.Sprintf("%s.(%s)", x.X, x.Type)
}

func (x UnaryExpr) String() string {
	return fmt.Sprintf("%s%s", x.Op.TokenString(), x.X)
}

func (x CompositeLitExpr) String() string {
	if x.Type == nil {
		return fmt.Sprintf("<elided>{%s}", x.Elts.String())
	}
	return fmt.Sprintf("%s{%s}", x.Type.String(), x.Elts.String())
}

func (fle FuncLitExpr) String() string {
	heapCaptures := ""
	if len(fle.HeapCaptures) > 0 {
		heapCaptures = "<" + fle.HeapCaptures.String() + ">"
	}
	return fmt.Sprintf("func %s{ %s }%s", fle.Type, fle.Body.String(), heapCaptures)
}

func (x KeyValueExpr) String() string {
	if x.Key == nil {
		return x.Value.String()
	}
	return fmt.Sprintf("%s: %s", x.Key, x.Value)
}

func (x FieldTypeExpr) String() string {
	hd := ""
	if x.NameExpr.Type == NameExprTypeHeapDefine {
		hd = "~"
	}
	if x.Tag == nil {
		return fmt.Sprintf("%s%s %s", x.Name, hd, x.Type)
	}
	return fmt.Sprintf("%s%s %s %s", x.Name, hd, x.Type, x.Tag)
}

func (x ArrayTypeExpr) String() string {
	if x.Len == nil {
		return fmt.Sprintf("[...]%s", x.Elt)
	}
	return fmt.Sprintf("[%s]%s", x.Len, x.Elt)
}

func (x SliceTypeExpr) String() string {
	if x.Vrd {
		return fmt.Sprintf("...%s", x.Elt)
	}
	return fmt.Sprintf("[]%s", x.Elt)
}

func (x InterfaceTypeExpr) String() string {
	return fmt.Sprintf("interface { %v }", x.Methods)
}

func (x ChanTypeExpr) String() string {
	switch x.Dir {
	case SEND:
		return fmt.Sprintf("<-chan %s", x.Value)
	case RECV:
		return fmt.Sprintf("chan<- %s", x.Value)
	case SEND | RECV:
		return fmt.Sprintf("chan %s", x.Value)
	default:
		panic("unexpected chan dir")
	}
}

func (x FuncTypeExpr) String() string {
	params := ""
	if 0 < len(x.Params) {
		params = x.Params.String()
	}
	results := ""
	if 0 < len(x.Results) {
		results = " " + x.Results.String()
	}
	return fmt.Sprintf("func(%s)%s", params, results)
}

func (x MapTypeExpr) String() string {
	return fmt.Sprintf("map[%s] %s", x.Key, x.Value)
}

func (x StructTypeExpr) String() string {
	return fmt.Sprintf("struct { %v }", x.Fields)
}

func (x AssignStmt) String() string {
	return fmt.Sprintf("%v %s %v", x.Lhs, x.Op.TokenString(), x.Rhs)
}

func (x BlockStmt) String() string {
	return fmt.Sprintf("{ %s }", x.Body.String())
}

func (x BranchStmt) String() string {
	if x.Label == "" {
		return x.Op.TokenString()
	}
	return fmt.Sprintf("%s %s<%d,%d,%d>",
		x.Op.TokenString(), string(x.Label),
		x.BlockDepth, x.FrameDepth, x.BodyIndex)
}

func (x DeclStmt) String() string {
	return x.Body.String()
}

func (x DeferStmt) String() string {
	return "defer " + x.Call.String()
}

func (x EmptyStmt) String() string {
	return ""
}

func (x ExprStmt) String() string {
	return x.X.String()
}

func (x ForStmt) String() string {
	init, cond, post := "", "", ""
	if x.Init != nil {
		init = x.Init.String()
	}
	if x.Cond != nil {
		cond = x.Cond.String()
	}
	if x.Post != nil {
		post = x.Post.String()
	}
	return fmt.Sprintf("for %s; %s; %s { %s }",
		init, cond, post, x.Body.String())
}

func (x GoStmt) String() string {
	return "go " + x.Call.String()
}

func (x IfStmt) String() string {
	init := ""
	if x.Init != nil {
		init = x.Init.String() + "; "
	}
	cond := x.Cond.String()
	then := x.Then.String()
	els_ := x.Else.String()
	if x.Else.Body == nil {
		return fmt.Sprintf("if %s%s %s", init, cond, then)
	}
	return fmt.Sprintf("if %s%s %s else %s",
		init, cond, then, els_)
}

func (x IfCaseStmt) String() string {
	return "{ " + x.Body.String() + " }"
}

func (x IncDecStmt) String() string {
	switch x.Op {
	case INC:
		return x.X.String() + "++"
	case DEC:
		return x.X.String() + "--"
	default:
		panic("unexpected operator")
	}
}

func (x RangeStmt) String() string {
	if x.Key == nil {
		if x.Value != nil {
			panic("unexpected value in range stmt with no key")
		}
		return fmt.Sprintf("for range %s { %s }",
			x.X.String(), x.Body.String())
	} else if x.Value == nil {
		return fmt.Sprintf("for %s %s range %s { %s }",
			x.Key.String(), x.Op.TokenString(),
			x.X.String(), x.Body.String())
	} else {
		return fmt.Sprintf("for %s, %s %s range %s { %s }",
			x.Key.String(), x.Value.String(), x.Op.TokenString(),
			x.X.String(), x.Body.String())
	}
}

func (x ReturnStmt) String() string {
	if len(x.Results) == 0 {
		return "return"
	}
	return fmt.Sprintf("return %v", x.Results)
}

func (x SelectStmt) String() string {
	cases := ""
	for i, s := range x.Cases {
		if i == 0 {
			cases += s.String()
		} else {
			cases += "; " + s.String()
		}
	}
	return fmt.Sprintf("select { %s }", cases)
}

func (x SelectCaseStmt) String() string {
	return fmt.Sprintf("case %v: %s", x.Comm.String(), x.Body.String())
}

func (x SendStmt) String() string {
	return fmt.Sprintf("%s <- %s", x.Chan.String(), x.Value.String())
}

func (x SwitchStmt) String() string {
	init := ""
	if x.Init != nil {
		init = x.Init.String() + "; "
	}
	varName := ""
	if x.VarName != "" {
		varName = string(x.VarName) + ":="
	}
	cases := ""
	for i, s := range x.Clauses {
		if i == 0 {
			cases += s.String()
		} else {
			cases += "; " + s.String()
		}
	}
	return fmt.Sprintf("switch %s%s%s { %s }",
		init, varName, x.X.String(), cases)
}

func (x SwitchClauseStmt) String() string {
	if len(x.Cases) == 0 {
		return fmt.Sprintf("default: %s", x.Body.String())
	}
	return fmt.Sprintf("case %v: %s", x.Cases, x.Body.String())
}

func (x FuncDecl) String() string {
	recv := ""
	if x.IsMethod {
		recv = "(" + x.Recv.String() + ") "
	}
	return fmt.Sprintf("func %s%s%s { %s }",
		recv, x.Name, x.Type.String()[4:], x.Body.String())
}

func (x ImportDecl) String() string {
	if x.Name == "" {
		return fmt.Sprintf("import %s", x.PkgPath)
	}
	return fmt.Sprintf("import %s %s", x.Name, x.PkgPath)
}

func (x ValueDecl) String() string {
	mod := "var"
	if x.Const {
		mod = "const"
	}
	names := x.NameExprs.String()
	type_ := ""
	if x.Type != nil {
		type_ = " " + x.Type.String()
	}
	value := ""
	if x.Values != nil {
		value = " = " + x.Values.String()
	}
	return fmt.Sprintf("%s %s%s%s", mod, names, type_, value)
}

func (x TypeDecl) String() string {
	if x.IsAlias {
		return fmt.Sprintf("type %s = %s", x.Name, x.Type.String())
	}
	return fmt.Sprintf("type %s %s", x.Name, x.Type.String())
}

func (x FileNode) String() string {
	// return fmt.Sprintf("file{ package %s ... }", n.PkgName) // , n.Decls.String())
	return fmt.Sprintf("file{ package %s; %s }", x.PkgName, x.Decls.String())
}

func (pn PackageNode) String() string {
	return fmt.Sprintf("package(%s %s)", pn.PkgName, pn.PkgPath)
}

func (ref RefNode) String() string {
	return fmt.Sprintf("ref(%s)", ref.Location.String())
}

// ----------------------------------------
// Node slice strings
// NOTE: interface-generics or?

func (xs Exprs) String() string {
	str := ""
	for i, x := range xs {
		if i == 0 {
			str += x.String()
		} else {
			str += ", " + x.String()
		}
	}
	return str
}

func (nxs NameExprs) String() string {
	str := ""
	for i, nx := range nxs {
		if i == 0 {
			str += nx.String()
		} else {
			str += ", " + nx.String()
		}
	}
	return str
}

func (ftxz FieldTypeExprs) String() string {
	str := ""
	for i, x := range ftxz {
		if i == 0 {
			str += x.String()
		} else {
			str += ", " + x.String()
		}
	}
	return str
}

func (kvxs KeyValueExprs) String() string {
	str := ""
	for i, x := range kvxs {
		if i == 0 {
			str += x.String()
		} else {
			str += ", " + x.String()
		}
	}
	return str
}

func (ss Body) String() string {
	str := ""
	for i, s := range ss {
		if i == 0 {
			str += s.String()
		} else {
			str += "; " + s.String()
		}
	}
	return str
}

func (ds Decls) String() string {
	str := ""
	for i, s := range ds {
		if i == 0 {
			str += s.String()
		} else {
			str += "; " + s.String()
		}
	}
	return str
}

func (x ConstExpr) String() string {
	if x.TypedValue.HasKind(TypeKind) {
		return x.TypedValue.V.String()
	} else {
		return fmt.Sprintf("(const %s)", x.TypedValue.String())
	}
}

func (x constTypeExpr) String() string {
	if x.Type == nil { // type switch case
		return "(const-type nil)"
	}
	return fmt.Sprintf("(const-type %s)", x.Type.String())
}
