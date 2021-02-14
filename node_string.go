package gno

import (
	"fmt"
	"strings"
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
	case VPTypeUverse:
		return fmt.Sprintf("#%d", p.Index)
	case VPTypeDefault:
		return fmt.Sprintf("%s%d", strings.Repeat("@", int(p.Depth)), p.Index)
	case VPTypeInterface:
		return fmt.Sprintf("@%s", p.Name)
	case VPTypeNative:
		return fmt.Sprintf("@%s", p.Name)
	default:
		panic("illegal_value_type")
	}
}

func (n NameExpr) String() string {
	return fmt.Sprintf("%s%s", n.Name, n.Path.String())
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
	return fmt.Sprintf("%s.%s", n.X, n.Sel)
}

func (n SliceExpr) String() string {
	return fmt.Sprintf("%s[%s:%s:%s]", n.X, n.Low, n.High, n.Max)
}

func (n StarExpr) String() string {
	return fmt.Sprintf("*%s", n.X)
}

func (n RefExpr) String() string {
	return fmt.Sprintf("&%s", n.X)
}

func (n TypeAssertExpr) String() string {
	return fmt.Sprintf("%s.(%s)", n.X, n.Type)
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
		return fmt.Sprintf("%s %s", n.Op.TokenString(), string(n.Label))
	}
}

func (n DeclStmt) String() string {
	return n.Decls.String()
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
	return fmt.Sprintf("for %s; %s; %s { %s }",
		n.Init, n.Cond, n.Post, n.Body.String())
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
	body := n.Body.String()
	els_ := n.Else.String()
	if n.Else == nil {
		return fmt.Sprintf("if %s%s { %s }", init, cond, body)
	} else {
		return fmt.Sprintf("if %s%s { %s } else { %s }",
			init, cond, body, els_)
	}
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

func (n LabeledStmt) String() string {
	return string(n.Label) + ": " + n.Stmt.String()
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
	for i, s := range n.Cases {
		if i == 0 {
			cases += s.String()
		} else {
			cases += "; " + s.String()
		}
	}
	return fmt.Sprintf("switch %s%s%s { %s }",
		init, varName, n.X.String(), cases)
}

func (n SwitchCaseStmt) String() string {
	return fmt.Sprintf("case %v: %s", n.Cases, n.Body.String())
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
		return fmt.Sprintf("import %s", n.Path)
	} else {
		return fmt.Sprintf("import %s %s", n.Name, n.Path)
	}
}

func (n ValueDecl) String() string {
	mod := "var"
	if n.Const {
		mod = "const"
	}
	type_ := ""
	if n.Type != nil {
		type_ = " " + n.Type.String()
	}
	value := ""
	if n.Value != nil {
		value = " = " + n.Value.String()
	}
	return fmt.Sprintf("%s %s%s%s", mod, n.Name, type_, value)
}

func (n TypeDecl) String() string {
	if n.IsAlias {
		return fmt.Sprintf("type %s = %s", n.Name, n.Type.String())
	} else {
		return fmt.Sprintf("type %s %s", n.Name, n.Type.String())
	}
}

func (n FileNode) String() string {
	return fmt.Sprintf("file{ package %s; %s }", n.PkgName, n.Body.String())
}

func (n PackageNode) String() string {
	return fmt.Sprintf("package(%s)", n.PkgName)
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

func (ss Stmts) String() string {
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

func (ds SimpleDecls) String() string {
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
