package gnolang

import (
	"fmt"
)

//----------------------------------------
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

//----------------------------------------
// Node.String()

func (p ValuePath) String() string {
	switch p.Type {
	case VPUverse:
		return fmt.Sprintf("VPUverse(%d)", p.Index)
	case VPBlock:
		return fmt.Sprintf("VPBlock(%d,%d)", p.Depth, p.Index)
	case VPField:
		return fmt.Sprintf("VPField(%d,%d,%s)", p.Depth, p.Index, p.Name)
	case VPSubrefField:
		return fmt.Sprintf("VPSubrefField(%d,%d,%s)", p.Depth, p.Index, p.Name)
	case VPValMethod:
		return fmt.Sprintf("VPValMethod(%d,%s)", p.Index, p.Name)
	case VPPtrMethod:
		return fmt.Sprintf("VPPtrMethod(%d,%s)", p.Index, p.Name)
	case VPInterface:
		return fmt.Sprintf("VPInterface(%s)", p.Name)
	case VPDerefField:
		return fmt.Sprintf("VPDerefField(%d,%d,%s)", p.Depth, p.Index, p.Name)
	case VPDerefValMethod:
		return fmt.Sprintf("VPDerefValMethod(%d,%s)", p.Index, p.Name)
	case VPDerefPtrMethod:
		return fmt.Sprintf("VPDerefPtrMethod(%d,%s)", p.Index, p.Name)
	case VPDerefInterface:
		return fmt.Sprintf("VPDerefInterface(%s)", p.Name)
	case VPNative:
		return fmt.Sprintf("VPNative(%s)", p.Name)
	default:
		panic("illegal_value_type")
	}
}

func (n NameExpr) String() string {
	return fmt.Sprintf("%s<%s>", n.Name, n.Path.String())
}

func (n BasicLitExpr) String() string {
	return n.Value
}

func (n BinaryExpr) String() string {
	return fmt.Sprintf("%s %s %s",
		n.Left.String(),
		n.Op.TokenString(),
		n.Right.String(),
	)
}

func (n CallExpr) String() string {
	if n.Varg {
		return fmt.Sprintf("%s(%s...)", n.Func, n.Args.String())
	} else {
		return fmt.Sprintf("%s(%s)", n.Func, n.Args.String())
	}
}

func (n IndexExpr) String() string {
	return fmt.Sprintf("%s[%s]", n.X, n.Index)
}

func (n SelectorExpr) String() string {
	// NOTE: for debugging selector issues:
	// return fmt.Sprintf("%s.(%v).%s", n.X, n.Path.Type, n.Sel)
	return fmt.Sprintf("%s.%s", n.X, n.Sel)
}

func (n SliceExpr) String() string {
	ls, hs, ms := "", "", ""
	if n.Low != nil {
		ls = n.Low.String()
	}
	if n.High != nil {
		hs = n.High.String()
	}
	if n.Max != nil {
		ms = n.Max.String()
	}
	if ms == "" {
		return fmt.Sprintf("%s[%s:%s]", n.X, ls, hs)
	} else {
		return fmt.Sprintf("%s[%s:%s:%s]", n.X, ls, hs, ms)
	}
}

func (n StarExpr) String() string {
	return fmt.Sprintf("*(%s)", n.X)
}

func (n RefExpr) String() string {
	return fmt.Sprintf("&(%s)", n.X)
}

func (n TypeAssertExpr) String() string {
	if n.Type == nil {
		return fmt.Sprintf("%s.(type)", n.X)
	} else {
		return fmt.Sprintf("%s.(%s)", n.X, n.Type)
	}
}

func (n UnaryExpr) String() string {
	return fmt.Sprintf("%s%s", n.Op.TokenString(), n.X)
}

func (n CompositeLitExpr) String() string {
	if n.Type == nil {
		return fmt.Sprintf("<elided>{%s}", n.Elts.String())
	} else {
		return fmt.Sprintf("%s{%s}", n.Type.String(), n.Elts.String())
	}
}

func (n FuncLitExpr) String() string {
	return fmt.Sprintf("func %s{ %s }", n.Type, n.Body.String())
}

func (n KeyValueExpr) String() string {
	if n.Key == nil {
		return fmt.Sprintf("%s", n.Value)
	} else {
		return fmt.Sprintf("%s: %s", n.Key, n.Value)
	}
}

func (n FieldTypeExpr) String() string {
	if n.Tag == nil {
		return fmt.Sprintf("%s %s", n.Name, n.Type)
	} else {
		return fmt.Sprintf("%s %s %s", n.Name, n.Type, n.Tag)
	}
}

func (n ArrayTypeExpr) String() string {
	if n.Len == nil {
		return fmt.Sprintf("[...]%s", n.Elt)
	} else {
		return fmt.Sprintf("[%s]%s", n.Len, n.Elt)
	}
}

func (n SliceTypeExpr) String() string {
	if n.Vrd {
		return fmt.Sprintf("...%s", n.Elt)
	} else {
		return fmt.Sprintf("[]%s", n.Elt)
	}
}

func (n InterfaceTypeExpr) String() string {
	return fmt.Sprintf("interface { %v }", n.Methods)
}

func (n ChanTypeExpr) String() string {
	switch n.Dir {
	case SEND:
		return fmt.Sprintf("<-chan %s", n.Value)
	case RECV:
		return fmt.Sprintf("chan<- %s", n.Value)
	case SEND | RECV:
		return fmt.Sprintf("chan %s", n.Value)
	default:
		panic("unexpected chan dir")
	}
}

func (n FuncTypeExpr) String() string {
	params := ""
	if 0 < len(n.Params) {
		params = n.Params.String()
	}
	results := ""
	if 0 < len(n.Results) {
		results = " " + n.Results.String()
	}
	return fmt.Sprintf("func(%s)%s", params, results)
}

func (n MapTypeExpr) String() string {
	return fmt.Sprintf("map[%s] %s", n.Key, n.Value)
}

func (n StructTypeExpr) String() string {
	return fmt.Sprintf("struct { %v }", n.Fields)
}

func (n MaybeNativeTypeExpr) String() string {
	return fmt.Sprintf("maybenative(%s)", n.Type.String())
}

func (n AssignStmt) String() string {
	return fmt.Sprintf("%v %s %v", n.Lhs, n.Op.TokenString(), n.Rhs)
}

func (n BlockStmt) String() string {
	return fmt.Sprintf("{ %s }", n.Body.String())
}

func (n BranchStmt) String() string {
	if n.Label == "" {
		return n.Op.TokenString()
	} else {
		return fmt.Sprintf("%s %s<%d,%d>",
			n.Op.TokenString(), string(n.Label),
			n.Depth, n.BodyIndex)
	}
}

func (n DeclStmt) String() string {
	return n.Body.String()
}

func (n DeferStmt) String() string {
	return "defer " + n.Call.String()
}

func (n EmptyStmt) String() string {
	return ""
}

func (n ExprStmt) String() string {
	return n.X.String()
}

func (n ForStmt) String() string {
	init, cond, post := "", "", ""
	if n.Init != nil {
		init = n.Init.String()
	}
	if n.Cond != nil {
		cond = n.Cond.String()
	}
	if n.Post != nil {
		post = n.Post.String()
	}
	return fmt.Sprintf("for %s; %s; %s { %s }",
		init, cond, post, n.Body.String())
}

func (n GoStmt) String() string {
	return "go " + n.Call.String()
}

func (n IfStmt) String() string {
	init := ""
	if n.Init != nil {
		init = n.Init.String() + "; "
	}
	cond := n.Cond.String()
	then := n.Then.String()
	els_ := n.Else.String()
	if n.Else.Body == nil {
		return fmt.Sprintf("if %s%s { %s }", init, cond, then)
	} else {
		return fmt.Sprintf("if %s%s { %s } else { %s }",
			init, cond, then, els_)
	}
}

func (n IfCaseStmt) String() string {
	return n.Body.String()
}

func (n IncDecStmt) String() string {
	if n.Op == INC {
		return n.X.String() + "++"
	} else if n.Op == DEC {
		return n.X.String() + "--"
	} else {
		panic("unexpected operator")
	}
}

func (n RangeStmt) String() string {
	if n.Key == nil {
		if n.Value != nil {
			panic("unexpected value in range stmt with no key")
		}
		return fmt.Sprintf("for range %s { %s }",
			n.X.String(), n.Body.String())
	} else if n.Value == nil {
		return fmt.Sprintf("for %s %s range %s { %s }",
			n.Key.String(), n.Op.TokenString(),
			n.X.String(), n.Body.String())
	} else {
		return fmt.Sprintf("for %s, %s %s range %s { %s }",
			n.Key.String(), n.Value.String(), n.Op.TokenString(),
			n.X.String(), n.Body.String())
	}
}

func (n ReturnStmt) String() string {
	if len(n.Results) == 0 {
		return fmt.Sprintf("return")
	} else {
		return fmt.Sprintf("return %v", n.Results)
	}
}

func (n PanicStmt) String() string {
	return fmt.Sprintf("panic(%s)", n.Exception.String())
}

func (n SelectStmt) String() string {
	cases := ""
	for i, s := range n.Cases {
		if i == 0 {
			cases += s.String()
		} else {
			cases += "; " + s.String()
		}
	}
	return fmt.Sprintf("select { %s }", cases)
}

func (n SelectCaseStmt) String() string {
	return fmt.Sprintf("case %v: %s", n.Comm.String(), n.Body.String())
}

func (n SendStmt) String() string {
	return fmt.Sprintf("%s <- %s", n.Chan.String(), n.Value.String())
}

func (n SwitchStmt) String() string {
	init := ""
	if n.Init != nil {
		init = n.Init.String() + "; "
	}
	varName := ""
	if n.VarName != "" {
		varName = string(n.VarName) + ":="
	}
	cases := ""
	for i, s := range n.Clauses {
		if i == 0 {
			cases += s.String()
		} else {
			cases += "; " + s.String()
		}
	}
	return fmt.Sprintf("switch %s%s%s { %s }",
		init, varName, n.X.String(), cases)
}

func (n SwitchClauseStmt) String() string {
	if len(n.Cases) == 0 {
		return fmt.Sprintf("default: %s", n.Body.String())
	} else {
		return fmt.Sprintf("case %v: %s", n.Cases, n.Body.String())
	}
}

func (n FuncDecl) String() string {
	recv := ""
	if n.IsMethod {
		recv = "(" + n.Recv.String() + ") "
	}
	return fmt.Sprintf("func %s%s%s { %s }",
		recv, n.Name, n.Type.String()[4:], n.Body.String())
}

func (n ImportDecl) String() string {
	if n.Name == "" {
		return fmt.Sprintf("import %s", n.PkgPath)
	} else {
		return fmt.Sprintf("import %s %s", n.Name, n.PkgPath)
	}
}

func (n ValueDecl) String() string {
	mod := "var"
	if n.Const {
		mod = "const"
	}
	names := n.NameExprs.String()
	type_ := ""
	if n.Type != nil {
		type_ = " " + n.Type.String()
	}
	value := ""
	if n.Values != nil {
		value = " = " + n.Values.String()
	}
	return fmt.Sprintf("%s %s%s%s", mod, names, type_, value)
}

func (n TypeDecl) String() string {
	if n.IsAlias {
		return fmt.Sprintf("type %s = %s", n.Name, n.Type.String())
	} else {
		return fmt.Sprintf("type %s %s", n.Name, n.Type.String())
	}
}

func (n FileNode) String() string {
	// return fmt.Sprintf("file{ package %s ... }", n.PkgName) // , n.Decls.String())
	return fmt.Sprintf("file{ package %s; %s }", n.PkgName, n.Decls.String())
}

func (n PackageNode) String() string {
	return fmt.Sprintf("package(%s)", n.PkgName)
}

func (n RefNode) String() string {
	return fmt.Sprintf("ref(%s)", n.Location.String())
}

//----------------------------------------
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

func (fts FieldTypeExprs) String() string {
	str := ""
	for i, x := range fts {
		if i == 0 {
			str += x.String()
		} else {
			str += ", " + x.String()
		}
	}
	return str
}

func (kvs KeyValueExprs) String() string {
	str := ""
	for i, x := range kvs {
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

func (cx ConstExpr) String() string {
	return fmt.Sprintf("(const %s)", cx.TypedValue.String())
}

func (ctx constTypeExpr) String() string {
	if ctx.Type == nil { // type switch case
		return fmt.Sprintf("(const-type nil)")
	} else {
		return fmt.Sprintf("(const-type %s)", ctx.Type.String())
	}
}
