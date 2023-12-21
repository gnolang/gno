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

func (vp ValuePath) String(_ *Debugging) string {
	switch vp.Type {
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
	case VPNative:
		return fmt.Sprintf("VPNative(%s)", vp.Name)
	default:
		panic("illegal_value_type")
	}
}

func (x NameExpr) String(debugging *Debugging) string {
	return fmt.Sprintf("%s<%s>", x.Name, x.Path.String(debugging))
}

func (x BasicLitExpr) String(_ *Debugging) string {
	return x.Value
}

func (x BinaryExpr) String(debugging *Debugging) string {
	return fmt.Sprintf("%s %s %s",
		x.Left.String(debugging),
		x.Op.TokenString(),
		x.Right.String(debugging),
	)
}

func (x CallExpr) String(debugging *Debugging) string {
	if x.Varg {
		return fmt.Sprintf("%s(%s...)", x.Func, x.Args.String(debugging))
	}
	return fmt.Sprintf("%s(%s)", x.Func, x.Args.String(debugging))
}

func (x IndexExpr) String(_ *Debugging) string {
	return fmt.Sprintf("%s[%s]", x.X, x.Index)
}

func (x SelectorExpr) String(_ *Debugging) string {
	// NOTE: for debugging selector issues:
	// return fmt.Sprintf("%s.(%v).%s", n.X, n.Path.Type, n.Sel)
	return fmt.Sprintf("%s.%s", x.X, x.Sel)
}

func (x SliceExpr) String(debugging *Debugging) string {
	ls, hs, ms := "", "", ""
	if x.Low != nil {
		ls = x.Low.String(debugging)
	}
	if x.High != nil {
		hs = x.High.String(debugging)
	}
	if x.Max != nil {
		ms = x.Max.String(debugging)
	}
	if ms == "" {
		return fmt.Sprintf("%s[%s:%s]", x.X, ls, hs)
	}
	return fmt.Sprintf("%s[%s:%s:%s]", x.X, ls, hs, ms)
}

func (x StarExpr) String(_ *Debugging) string {
	return fmt.Sprintf("*(%s)", x.X)
}

func (x RefExpr) String(_ *Debugging) string {
	return fmt.Sprintf("&(%s)", x.X)
}

func (x TypeAssertExpr) String(_ *Debugging) string {
	if x.Type == nil {
		return fmt.Sprintf("%s.(type)", x.X)
	}
	return fmt.Sprintf("%s.(%s)", x.X, x.Type)
}

func (x UnaryExpr) String(_ *Debugging) string {
	return fmt.Sprintf("%s%s", x.Op.TokenString(), x.X)
}

func (x CompositeLitExpr) String(debugging *Debugging) string {
	if x.Type == nil {
		return fmt.Sprintf("<elided>{%s}", x.Elts.String(debugging))
	}
	return fmt.Sprintf("%s{%s}", x.Type.String(debugging), x.Elts.String(debugging))
}

func (x FuncLitExpr) String(debugging *Debugging) string {
	return fmt.Sprintf("func %s{ %s }", x.Type, x.Body.String(debugging))
}

func (x KeyValueExpr) String(_ *Debugging) string {
	if x.Key == nil {
		return fmt.Sprintf("%s", x.Value)
	}
	return fmt.Sprintf("%s: %s", x.Key, x.Value)
}

func (x FieldTypeExpr) String(_ *Debugging) string {
	if x.Tag == nil {
		return fmt.Sprintf("%s %s", x.Name, x.Type)
	}
	return fmt.Sprintf("%s %s %s", x.Name, x.Type, x.Tag)
}

func (x ArrayTypeExpr) String(_ *Debugging) string {
	if x.Len == nil {
		return fmt.Sprintf("[...]%s", x.Elt)
	}
	return fmt.Sprintf("[%s]%s", x.Len, x.Elt)
}

func (x SliceTypeExpr) String(_ *Debugging) string {
	if x.Vrd {
		return fmt.Sprintf("...%s", x.Elt)
	}
	return fmt.Sprintf("[]%s", x.Elt)
}

func (x InterfaceTypeExpr) String(_ *Debugging) string {
	return fmt.Sprintf("interface { %v }", x.Methods)
}

func (x ChanTypeExpr) String(_ *Debugging) string {
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

func (x FuncTypeExpr) String(debugging *Debugging) string {
	params := ""
	if 0 < len(x.Params) {
		params = x.Params.String(debugging)
	}
	results := ""
	if 0 < len(x.Results) {
		results = " " + x.Results.String(debugging)
	}
	return fmt.Sprintf("func(%s)%s", params, results)
}

func (x MapTypeExpr) String(_ *Debugging) string {
	return fmt.Sprintf("map[%s] %s", x.Key, x.Value)
}

func (x StructTypeExpr) String(_ *Debugging) string {
	return fmt.Sprintf("struct { %v }", x.Fields)
}

func (x MaybeNativeTypeExpr) String(debugging *Debugging) string {
	return fmt.Sprintf("maybenative(%s)", x.Type.String(debugging))
}

func (x AssignStmt) String(_ *Debugging) string {
	return fmt.Sprintf("%v %s %v", x.Lhs, x.Op.TokenString(), x.Rhs)
}

func (x BlockStmt) String(debugging *Debugging) string {
	return fmt.Sprintf("{ %s }", x.Body.String(debugging))
}

func (x BranchStmt) String(_ *Debugging) string {
	if x.Label == "" {
		return x.Op.TokenString()
	}
	return fmt.Sprintf("%s %s<%d,%d>",
		x.Op.TokenString(), string(x.Label),
		x.Depth, x.BodyIndex)
}

func (x DeclStmt) String(debugging *Debugging) string {
	return x.Body.String(debugging)
}

func (x DeferStmt) String(debugging *Debugging) string {
	return "defer " + x.Call.String(debugging)
}

func (x EmptyStmt) String(_ *Debugging) string {
	return ""
}

func (x ExprStmt) String(debugging *Debugging) string {
	return x.X.String(debugging)
}

func (x ForStmt) String(debugging *Debugging) string {
	init, cond, post := "", "", ""
	if x.Init != nil {
		init = x.Init.String(debugging)
	}
	if x.Cond != nil {
		cond = x.Cond.String(debugging)
	}
	if x.Post != nil {
		post = x.Post.String(debugging)
	}
	return fmt.Sprintf("for %s; %s; %s { %s }",
		init, cond, post, x.Body.String(debugging))
}

func (x GoStmt) String(debugging *Debugging) string {
	return "go " + x.Call.String(debugging)
}

func (x IfStmt) String(debugging *Debugging) string {
	init := ""
	if x.Init != nil {
		init = x.Init.String(debugging) + "; "
	}
	cond := x.Cond.String(debugging)
	then := x.Then.String(debugging)
	els_ := x.Else.String(debugging)
	if x.Else.Body == nil {
		return fmt.Sprintf("if %s%s { %s }", init, cond, then)
	}
	return fmt.Sprintf("if %s%s { %s } else { %s }",
		init, cond, then, els_)
}

func (x IfCaseStmt) String(debugging *Debugging) string {
	return x.Body.String(debugging)
}

func (x IncDecStmt) String(debugging *Debugging) string {
	switch x.Op {
	case INC:
		return x.X.String(debugging) + "++"
	case DEC:
		return x.X.String(debugging) + "--"
	default:
		panic("unexpected operator")
	}
}

func (x RangeStmt) String(debugging *Debugging) string {
	if x.Key == nil {
		if x.Value != nil {
			panic("unexpected value in range stmt with no key")
		}
		return fmt.Sprintf("for range %s { %s }",
			x.X.String(debugging), x.Body.String(debugging))
	} else if x.Value == nil {
		return fmt.Sprintf("for %s %s range %s { %s }",
			x.Key.String(debugging), x.Op.TokenString(),
			x.X.String(debugging), x.Body.String(debugging))
	} else {
		return fmt.Sprintf("for %s, %s %s range %s { %s }",
			x.Key.String(debugging), x.Value.String(debugging), x.Op.TokenString(),
			x.X.String(debugging), x.Body.String(debugging))
	}
}

func (x ReturnStmt) String(_ *Debugging) string {
	if len(x.Results) == 0 {
		return fmt.Sprintf("return")
	}
	return fmt.Sprintf("return %v", x.Results)
}

func (x PanicStmt) String(debugging *Debugging) string {
	return fmt.Sprintf("panic(%s)", x.Exception.String(debugging))
}

func (x SelectStmt) String(debugging *Debugging) string {
	cases := ""
	for i, s := range x.Cases {
		if i == 0 {
			cases += s.String(debugging)
		} else {
			cases += "; " + s.String(debugging)
		}
	}
	return fmt.Sprintf("select { %s }", cases)
}

func (x SelectCaseStmt) String(debugging *Debugging) string {
	return fmt.Sprintf("case %v: %s", x.Comm.String(debugging), x.Body.String(debugging))
}

func (x SendStmt) String(debugging *Debugging) string {
	return fmt.Sprintf("%s <- %s", x.Chan.String(debugging), x.Value.String(debugging))
}

func (x SwitchStmt) String(debugging *Debugging) string {
	init := ""
	if x.Init != nil {
		init = x.Init.String(debugging) + "; "
	}
	varName := ""
	if x.VarName != "" {
		varName = string(x.VarName) + ":="
	}
	cases := ""
	for i, s := range x.Clauses {
		if i == 0 {
			cases += s.String(debugging)
		} else {
			cases += "; " + s.String(debugging)
		}
	}
	return fmt.Sprintf("switch %s%s%s { %s }",
		init, varName, x.X.String(debugging), cases)
}

func (x SwitchClauseStmt) String(debugging *Debugging) string {
	if len(x.Cases) == 0 {
		return fmt.Sprintf("default: %s", x.Body.String(debugging))
	}
	return fmt.Sprintf("case %v: %s", x.Cases, x.Body.String(debugging))
}

func (x FuncDecl) String(debugging *Debugging) string {
	recv := ""
	if x.IsMethod {
		recv = "(" + x.Recv.String(debugging) + ") "
	}
	return fmt.Sprintf("func %s%s%s { %s }",
		recv, x.Name, x.Type.String(debugging)[4:], x.Body.String(debugging))
}

func (x ImportDecl) String(_ *Debugging) string {
	if x.Name == "" {
		return fmt.Sprintf("import %s", x.PkgPath)
	}
	return fmt.Sprintf("import %s %s", x.Name, x.PkgPath)
}

func (x ValueDecl) String(debugging *Debugging) string {
	mod := "var"
	if x.Const {
		mod = "const"
	}
	names := x.NameExprs.String(debugging)
	type_ := ""
	if x.Type != nil {
		type_ = " " + x.Type.String(debugging)
	}
	value := ""
	if x.Values != nil {
		value = " = " + x.Values.String(debugging)
	}
	return fmt.Sprintf("%s %s%s%s", mod, names, type_, value)
}

func (x TypeDecl) String(debugging *Debugging) string {
	if x.IsAlias {
		return fmt.Sprintf("type %s = %s", x.Name, x.Type.String(debugging))
	}
	return fmt.Sprintf("type %s %s", x.Name, x.Type.String(debugging))
}

func (x FileNode) String(debugging *Debugging) string {
	// return fmt.Sprintf("file{ package %s ... }", n.PkgName) // , n.Decls.String())
	return fmt.Sprintf("file{ package %s; %s }", x.PkgName, x.Decls.String(debugging))
}

func (x PackageNode) String(_ *Debugging) string {
	return fmt.Sprintf("package(%s)", x.PkgName)
}

func (rn RefNode) String(debugging *Debugging) string {
	return fmt.Sprintf("ref(%s)", rn.Location.String(debugging))
}

// ----------------------------------------
// Node slice strings
// NOTE: interface-generics or?

func (xs Exprs) String(debugging *Debugging) string {
	str := ""
	for i, x := range xs {
		if i == 0 {
			str += x.String(debugging)
		} else {
			str += ", " + x.String(debugging)
		}
	}
	return str
}

func (nxs NameExprs) String(debugging *Debugging) string {
	str := ""
	for i, nx := range nxs {
		if i == 0 {
			str += nx.String(debugging)
		} else {
			str += ", " + nx.String(debugging)
		}
	}
	return str
}

func (ftxz FieldTypeExprs) String(debugging *Debugging) string {
	str := ""
	for i, x := range ftxz {
		if i == 0 {
			str += x.String(debugging)
		} else {
			str += ", " + x.String(debugging)
		}
	}
	return str
}

func (kvs KeyValueExprs) String(debugging *Debugging) string {
	str := ""
	for i, x := range kvs {
		if i == 0 {
			str += x.String(debugging)
		} else {
			str += ", " + x.String(debugging)
		}
	}
	return str
}

func (ss Body) String(debugging *Debugging) string {
	str := ""
	for i, s := range ss {
		if i == 0 {
			str += s.String(debugging)
		} else {
			str += "; " + s.String(debugging)
		}
	}
	return str
}

func (ds Decls) String(debugging *Debugging) string {
	str := ""
	for i, s := range ds {
		if i == 0 {
			str += s.String(debugging)
		} else {
			str += "; " + s.String(debugging)
		}
	}
	return str
}

func (x ConstExpr) String(_ *Debugging) string {
	return fmt.Sprintf("(const %s)", x.TypedValue.String())
}

func (x constTypeExpr) String(debugging *Debugging) string {
	if x.Type == nil { // type switch case
		return fmt.Sprintf("(const-type nil)")
	}
	return fmt.Sprintf("(const-type %s)", x.Type.String(debugging))
}
