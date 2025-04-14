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

func (x NameExpr) String(m *Machine) string {
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

func (x BasicLitExpr) String(m *Machine) string {
	return x.Value
}

func (x BinaryExpr) String(m *Machine) string {
	return fmt.Sprintf("%s %s %s",
		x.Left.String(m),
		x.Op.TokenString(),
		x.Right.String(m),
	)
}

func (x CallExpr) String(m *Machine) string {
	if x.Varg {
		return fmt.Sprintf("%s(%s...)", x.Func, x.Args.String(m))
	}
	return fmt.Sprintf("%s(%s)", x.Func, x.Args.String(m))
}

func (x IndexExpr) String(m *Machine) string {
	return fmt.Sprintf("%s[%s]", x.X, x.Index)
}

func (x SelectorExpr) String(m *Machine) string {
	// NOTE: for debugging selector issues:
	// return fmt.Sprintf("%s.(%v).%s", n.X, n.Path.Type, n.Sel)
	return fmt.Sprintf("%s.%s", x.X, x.Sel)
}

func (x SliceExpr) String(m *Machine) string {
	ls, hs, ms := "", "", ""
	if x.Low != nil {
		ls = x.Low.String(m)
	}
	if x.High != nil {
		hs = x.High.String(m)
	}
	if x.Max != nil {
		ms = x.Max.String(m)
	}
	if ms == "" {
		return fmt.Sprintf("%s[%s:%s]", x.X, ls, hs)
	}
	return fmt.Sprintf("%s[%s:%s:%s]", x.X, ls, hs, ms)
}

func (x StarExpr) String(m *Machine) string {
	return fmt.Sprintf("*(%s)", x.X)
}

func (x RefExpr) String(m *Machine) string {
	return fmt.Sprintf("&(%s)", x.X)
}

func (x TypeAssertExpr) String(m *Machine) string {
	if x.Type == nil {
		return fmt.Sprintf("%s.(type)", x.X)
	}
	return fmt.Sprintf("%s.(%s)", x.X, x.Type)
}

func (x UnaryExpr) String(m *Machine) string {
	return fmt.Sprintf("%s%s", x.Op.TokenString(), x.X)
}

func (x CompositeLitExpr) String(m *Machine) string {
	if x.Type == nil {
		return fmt.Sprintf("<elided>{%s}", x.Elts.String(m))
	}
	return fmt.Sprintf("%s{%s}", x.Type.String(m), x.Elts.String(m))
}

func (x FuncLitExpr) String(m *Machine) string {
	heapCaptures := ""
	if len(x.HeapCaptures) > 0 {
		heapCaptures = "<" + x.HeapCaptures.String(m) + ">"
	}
	return fmt.Sprintf("func %s{ %s }%s", x.Type.String(m), x.Body.String(m), heapCaptures)
}

func (x KeyValueExpr) String(m *Machine) string {
	if x.Key == nil {
		return fmt.Sprintf("%s", x.Value)
	}
	return fmt.Sprintf("%s: %s", x.Key, x.Value)
}

func (x FieldTypeExpr) String(m *Machine) string {
	if x.Tag == nil {
		return fmt.Sprintf("%s %s", x.Name, x.Type)
	}
	return fmt.Sprintf("%s %s %s", x.Name, x.Type, x.Tag)
}

func (x ArrayTypeExpr) String(m *Machine) string {
	if x.Len == nil {
		return fmt.Sprintf("[...]%s", x.Elt)
	}
	return fmt.Sprintf("[%s]%s", x.Len, x.Elt)
}

func (x SliceTypeExpr) String(m *Machine) string {
	if x.Vrd {
		return fmt.Sprintf("...%s", x.Elt)
	}
	return fmt.Sprintf("[]%s", x.Elt)
}

func (x InterfaceTypeExpr) String(m *Machine) string {
	return fmt.Sprintf("interface { %v }", x.Methods)
}

func (x ChanTypeExpr) String(m *Machine) string {
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

func (x FuncTypeExpr) String(m *Machine) string {
	params := ""
	if 0 < len(x.Params) {
		params = x.Params.String(m)
	}
	results := ""
	if 0 < len(x.Results) {
		results = " " + x.Results.String(m)
	}
	return fmt.Sprintf("func(%s)%s", params, results)
}

func (x MapTypeExpr) String(m *Machine) string {
	return fmt.Sprintf("map[%s] %s", x.Key, x.Value)
}

func (x StructTypeExpr) String(m *Machine) string {
	return fmt.Sprintf("struct { %v }", x.Fields)
}

func (x AssignStmt) String(m *Machine) string {
	return fmt.Sprintf("%v %s %v", x.Lhs, x.Op.TokenString(), x.Rhs)
}

func (x BlockStmt) String(m *Machine) string {
	return fmt.Sprintf("{ %s }", x.Body.String(m))
}

func (x BranchStmt) String(m *Machine) string {
	if x.Label == "" {
		return x.Op.TokenString()
	}
	return fmt.Sprintf("%s %s<%d,%d>",
		x.Op.TokenString(), string(x.Label),
		x.Depth, x.BodyIndex)
}

func (x DeclStmt) String(m *Machine) string {
	return x.Body.String(m)
}

func (x DeferStmt) String(m *Machine) string {
	return "defer " + x.Call.String(m)
}

func (x EmptyStmt) String(m *Machine) string {
	return ""
}

func (x ExprStmt) String(m *Machine) string {
	return x.X.String(m)
}

func (x ForStmt) String(m *Machine) string {
	init, cond, post := "", "", ""
	if x.Init != nil {
		init = x.Init.String(m)
	}
	if x.Cond != nil {
		cond = x.Cond.String(m)
	}
	if x.Post != nil {
		post = x.Post.String(m)
	}
	return fmt.Sprintf("for %s; %s; %s { %s }",
		init, cond, post, x.Body.String(m))
}

func (x GoStmt) String(m *Machine) string {
	return "go " + x.Call.String(m)
}

func (x IfStmt) String(m *Machine) string {
	init := ""
	if x.Init != nil {
		init = x.Init.String(m) + "; "
	}
	cond := x.Cond.String(m)
	then := x.Then.String(m)
	els_ := x.Else.String(m)
	if x.Else.Body == nil {
		return fmt.Sprintf("if %s%s { %s }", init, cond, then)
	}
	return fmt.Sprintf("if %s%s { %s } else { %s }",
		init, cond, then, els_)
}

func (x IfCaseStmt) String(m *Machine) string {
	return x.Body.String(m)
}

func (x IncDecStmt) String(m *Machine) string {
	switch x.Op {
	case INC:
		return x.X.String(m) + "++"
	case DEC:
		return x.X.String(m) + "--"
	default:
		panic("unexpected operator")
	}
}

func (x RangeStmt) String(m *Machine) string {
	if x.Key == nil {
		if x.Value != nil {
			panic("unexpected value in range stmt with no key")
		}
		return fmt.Sprintf("for range %s { %s }",
			x.X.String(m), x.Body.String(m))
	} else if x.Value == nil {
		return fmt.Sprintf("for %s %s range %s { %s }",
			x.Key.String(m), x.Op.TokenString(),
			x.X.String(m), x.Body.String(m))
	} else {
		return fmt.Sprintf("for %s, %s %s range %s { %s }",
			x.Key.String(m), x.Value.String(m), x.Op.TokenString(),
			x.X.String(m), x.Body.String(m))
	}
}

func (x ReturnStmt) String(m *Machine) string {
	if len(x.Results) == 0 {
		return fmt.Sprintf("return")
	}
	return fmt.Sprintf("return %v", x.Results)
}

func (x PanicStmt) String(m *Machine) string {
	return fmt.Sprintf("panic(%s)", x.Exception.String(m))
}

func (x SelectStmt) String(m *Machine) string {
	cases := ""
	for i, s := range x.Cases {
		if i == 0 {
			cases += s.String(m)
		} else {
			cases += "; " + s.String(m)
		}
	}
	return fmt.Sprintf("select { %s }", cases)
}

func (x SelectCaseStmt) String(m *Machine) string {
	return fmt.Sprintf("case %v: %s", x.Comm.String(m), x.Body.String(m))
}

func (x SendStmt) String(m *Machine) string {
	return fmt.Sprintf("%s <- %s", x.Chan.String(m), x.Value.String(m))
}

func (x SwitchStmt) String(m *Machine) string {
	init := ""
	if x.Init != nil {
		init = x.Init.String(m) + "; "
	}
	varName := ""
	if x.VarName != "" {
		varName = string(x.VarName) + ":="
	}
	cases := ""
	for i, s := range x.Clauses {
		if i == 0 {
			cases += s.String(m)
		} else {
			cases += "; " + s.String(m)
		}
	}
	return fmt.Sprintf("switch %s%s%s { %s }",
		init, varName, x.X.String(m), cases)
}

func (x SwitchClauseStmt) String(m *Machine) string {
	if len(x.Cases) == 0 {
		return fmt.Sprintf("default: %s", x.Body.String(m))
	}
	return fmt.Sprintf("case %v: %s", x.Cases, x.Body.String(m))
}

func (x FuncDecl) String(m *Machine) string {
	recv := ""
	if x.IsMethod {
		recv = "(" + x.Recv.String(m) + ") "
	}
	return fmt.Sprintf("func %s%s%s { %s }",
		recv, x.Name, x.Type.String(m)[4:], x.Body.String(m))
}

func (x ImportDecl) String(m *Machine) string {
	if x.Name == "" {
		return fmt.Sprintf("import %s", x.PkgPath)
	}
	return fmt.Sprintf("import %s %s", x.Name, x.PkgPath)
}

func (x ValueDecl) String(m *Machine) string {
	mod := "var"
	if x.Const {
		mod = "const"
	}
	names := x.NameExprs.String(m)
	type_ := ""
	if x.Type != nil {
		type_ = " " + x.Type.String(m)
	}
	value := ""
	if x.Values != nil {
		value = " = " + x.Values.String(m)
	}
	return fmt.Sprintf("%s %s%s%s", mod, names, type_, value)
}

func (x TypeDecl) String(m *Machine) string {
	if x.IsAlias {
		return fmt.Sprintf("type %s = %s", x.Name, x.Type.String(m))
	}
	return fmt.Sprintf("type %s %s", x.Name, x.Type.String(m))
}

func (x FileNode) String(m *Machine) string {
	// return fmt.Sprintf("file{ package %s ... }", n.PkgName) // , n.Decls.String())
	return fmt.Sprintf("file{ package %s; %s }", x.PkgName, x.Decls.String(m))
}

func (x PackageNode) String(m *Machine) string {
	return fmt.Sprintf("package(%s)", x.PkgName)
}

func (rn RefNode) String(m *Machine) string {
	return fmt.Sprintf("ref(%s)", rn.Location.String())
}

// ----------------------------------------
// Node slice strings
// NOTE: interface-generics or?

func (xs Exprs) String(m *Machine) string {
	str := ""
	for i, x := range xs {
		if i == 0 {
			str += x.String(m)
		} else {
			str += ", " + x.String(m)
		}
	}
	return str
}

func (nxs NameExprs) String(m *Machine) string {
	str := ""
	for i, nx := range nxs {
		if i == 0 {
			str += nx.String(m)
		} else {
			str += ", " + nx.String(m)
		}
	}
	return str
}

func (ftxz FieldTypeExprs) String(m *Machine) string {
	str := ""
	for i, x := range ftxz {
		if i == 0 {
			str += x.String(m)
		} else {
			str += ", " + x.String(m)
		}
	}
	return str
}

func (kvs KeyValueExprs) String(m *Machine) string {
	str := ""
	for i, x := range kvs {
		if i == 0 {
			str += x.String(m)
		} else {
			str += ", " + x.String(m)
		}
	}
	return str
}

func (ss Body) String(m *Machine) string {
	str := ""
	for i, s := range ss {
		if i == 0 {
			str += s.String(m)
		} else {
			str += "; " + s.String(m)
		}
	}
	return str
}

func (ds Decls) String(m *Machine) string {
	str := ""
	for i, s := range ds {
		if i == 0 {
			str += s.String(m)
		} else {
			str += "; " + s.String(m)
		}
	}
	return str
}

func (x ConstExpr) String(m *Machine) string {
	return fmt.Sprintf("(const %s)", x.TypedValue.String(m))
}

func (x constTypeExpr) String(m *Machine) string {
	if x.Type == nil { // type switch case
		return fmt.Sprintf("(const-type nil)")
	}
	return fmt.Sprintf("(const-type %s)", x.Type.String())
}
