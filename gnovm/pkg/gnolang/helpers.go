package gnolang

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	j "github.com/grepsuzette/joeson"
)

// ----------------------------------------
// AST Construction (Expr)
// These are copied over from go-amino-x, but produces Gno ASTs.

func N(n interface{}) Name {
	switch n := n.(type) {
	case string:
		return Name(n)
	case Name:
		return n
	default:
		panic("unexpected name arg")
	}
}

func Nx(n interface{}) *NameExpr {
	return &NameExpr{Name: N(n)}
}

func ArrayT(l, elt interface{}) *ArrayTypeExpr {
	return &ArrayTypeExpr{
		Len: X(l),
		Elt: X(elt),
	}
}

func SliceT(elt interface{}) *SliceTypeExpr {
	return &SliceTypeExpr{
		Elt: X(elt),
		Vrd: false,
	}
}

func MapT(key, value interface{}) *MapTypeExpr {
	return &MapTypeExpr{
		Key:   X(key),
		Value: X(value),
	}
}

func Vrd(elt interface{}) *SliceTypeExpr {
	return &SliceTypeExpr{
		Elt: X(elt),
		Vrd: true,
	}
}

func InterfaceT(methods FieldTypeExprs) *InterfaceTypeExpr {
	return &InterfaceTypeExpr{
		Methods: methods,
	}
}

func AnyT() *InterfaceTypeExpr {
	return InterfaceT(nil)
}

func GenT(generic Name, methods FieldTypeExprs) *InterfaceTypeExpr {
	return &InterfaceTypeExpr{
		Generic: generic,
		Methods: methods,
	}
}

func FuncT(params, results FieldTypeExprs) *FuncTypeExpr {
	return &FuncTypeExpr{
		Params:  params,
		Results: results,
	}
}

func Flds(args ...interface{}) FieldTypeExprs {
	list := FieldTypeExprs{}
	for i := 0; i < len(args); i += 2 {
		list = append(list, FieldTypeExpr{
			Name: N(args[i]),
			Type: X(args[i+1]),
		})
	}
	return list
}

func Recv(n, t interface{}) FieldTypeExpr {
	if n == "" {
		n = "_"
	}
	return FieldTypeExpr{
		Name: N(n),
		Type: X(t),
	}
}

func MaybeNativeT(tx interface{}) *MaybeNativeTypeExpr {
	return &MaybeNativeTypeExpr{
		Type: X(tx),
	}
}

func FuncD(name interface{}, params, results FieldTypeExprs, body []Stmt) *FuncDecl {
	return &FuncDecl{
		NameExpr: *Nx(name),
		Type: FuncTypeExpr{
			Params:  params,
			Results: results,
		},
		Body: body,
	}
}

func MthdD(name interface{}, recv FieldTypeExpr, params, results FieldTypeExprs, body []Stmt) *FuncDecl {
	return &FuncDecl{
		NameExpr: *Nx(name),
		Recv:     recv,
		Type: FuncTypeExpr{
			Params:  params,
			Results: results,
		},
		Body:     body,
		IsMethod: true,
	}
}

func Fn(params, results FieldTypeExprs, body []Stmt) *FuncLitExpr {
	return &FuncLitExpr{
		Type: *FuncT(params, results),
		Body: body,
	}
}

func Kv(n, v interface{}) KeyValueExpr {
	var kx, vx Expr
	if ns, ok := n.(string); ok {
		kx = X(ns) // key expr
	} else {
		kx = n.(Expr)
	}
	if vs, ok := v.(string); ok {
		vx = X(vs) // type expr
	} else {
		vx = v.(Expr)
	}
	return KeyValueExpr{
		Key:   kx,
		Value: vx,
	}
}

// Tries to infer statement from args.
func S(args ...interface{}) Stmt {
	if len(args) == 1 {
		switch arg0 := args[0].(type) {
		case Expr:
			return &ExprStmt{X: arg0}
		case Stmt:
			return arg0
		default:
			panic("dunno how to construct statement from argument")
		}
	}
	panic("dunno how to construct statement from arguments")
}

// Parses simple expressions (but not all).
// Useful for parsing strings to ast nodes, like foo.bar["qwe"](),
// new(bytes.Buffer), *bytes.Buffer, package.MyStruct{FieldA:1}, numeric
//
//   - num/char (e.g. e.g. 42, 0x7f, 3.14, 1e-9, 2.4i, 'a', '\x7f')
//   - strings (e.g. "foo" or `\m\n\o`), nil, function calls
//   - square bracket indexing
//   - dot notation
//   - star expression for pointers
//   - composite expressions
//   - nil
//   - type assertions, for EXPR.(EXPR) and also EXPR.(type)
//   - []type slice types
//   - [n]type array types
//   - &something referencing
//   - unary operations, namely
//     "+" | "-" | "!" | "^" | "*" | "&" | "<-" .
//   - binary operations, namely
//     "||", "&&",
//     "==" | "!=" | "<" | "<=" | ">" | ">="
//     "+" | "-" | "|" | "^"
//     "*" | "/" | "%" | "<<" | ">>" | "&" | "&^" .
//
// If the first argument is an expression, returns it.
// TODO replace this with rewrite of Joeson parser.
func Xold(x interface{}, args ...interface{}) Expr {
	switch cx := x.(type) {
	case Expr:
		return cx
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64:
		return X(fmt.Sprintf("%v", x))
	case string:
		if cx == "" {
			panic("input cannot be blank for X()")
		}
	case Name:
		if cx == "" {
			panic("input cannot be blank for X()")
		}
		x = string(cx)
	default:
		panic(fmt.Sprintf("unexpected input type for Xold(): %T", x))
		// panic("unexpected input type for X()")
	}
	expr := x.(string)
	expr = fmt.Sprintf(expr, args...)
	expr = strings.TrimSpace(expr)
	first := expr[0]

	// 1: Binary operators have a lower precedence than unary operators (or
	// monoids).
	left, op, right, ok := chopBinary(expr)
	if ok {
		return Bx(X(left), op, X(right))
	}

	// 2: Unary operators that depend on the first letter.
	switch first {
	case '*':
		return &StarExpr{
			X: X(expr[1:]),
		}
	case '&':
		return &RefExpr{
			X: X(expr[1:]),
		}
	case '+', '-', '!', '^':
		return &UnaryExpr{
			Op: Op2Word(expr[:1]),
			X:  X(expr[1:]),
		}
	case '<':
		second := expr[1] // is required.
		if second != '-' {
			panic("unparseable expression " + expr)
		}
		return &UnaryExpr{
			Op: Op2Word("<-"),
			X:  X(expr[2:]),
		}
	}

	// 3: Unary operators or literals that don't depend on the first letter,
	// and have some distinct suffix.
	if len(expr) > 1 {
		last := expr[len(expr)-1]
		switch last {
		case 'l':
			if expr == nilStr {
				return Nx(nilStr)
			}
		case 'i':
			if '0' <= expr[0] && expr[0] <= '9' {
				num := X(expr[:len(expr)-1]).(*BasicLitExpr)
				if num.Kind != INT && num.Kind != FLOAT {
					panic("expected int or float before 'i'")
				}
				num.Kind = IMAG
				return num
			}
		case '\'':
			if first != last {
				panic("unmatched quote")
			}
			return &BasicLitExpr{
				Kind:  CHAR,
				Value: expr[1 : len(expr)-1],
			}
		case '"', '`':
			if first != last {
				panic("unmatched quote")
			}
			return &BasicLitExpr{
				Kind:  STRING,
				Value: expr,
			}
		case ')':
			left, _, right := chopRight(expr)
			if left == "" {
				// Special case, not a function call.
				return X(right)
			} else if left[len(left)-1] == '.' {
				// Special case, a type assert.
				var x, t Expr = X(left[:len(left)-1]), nil
				if right == "type" {
					t = nil
				} else {
					t = X(right)
				}
				return &TypeAssertExpr{
					X:    x,
					Type: t,
				}
			}

			fn := X(left)
			args := []Expr{}
			parts := strings.Split(right, ",")
			for _, part := range parts {
				// NOTE: repeated commas have no effect,
				// nor do trailing commas.
				if len(part) > 0 {
					args = append(args, X(part))
				}
			}
			return &CallExpr{
				Func: fn,
				Args: args,
			}
		case '}':
			left, _, right := chopRight(expr)
			switch left {
			case "interface":
				panic("interface type expressions not supported, use InterfaceT(Flds(...)) instead")
			case "struct":
				panic("struct type expressions not supported")
			default:
				// composite type
				typ := X(left)
				kvs := []KeyValueExpr{}
				parts := strings.Split(right, ",")
				for _, part := range parts {
					if strings.TrimSpace(part) != "" {
						parts := strings.Split(part, ":")
						if len(parts) != 2 {
							panic("key:value requires 1 colon")
						}
						kvs = append(kvs, Kv(parts[0], parts[1]))
					}
				}
				return &CompositeLitExpr{
					Type: typ,
					Elts: kvs,
				}
			}
		case ']':
			left, _, right := chopRight(expr)
			return Idx(left, right)
		}
	}
	// 4.  Monoids of array or slice type.
	// NOTE: []foo.bar requires this to have lower precedence than dots.
	switch first {
	case '.': // variadic ... prefix.
		if expr[1] == '.' && expr[2] == '.' {
			return Vrd(expr[3:])
		} else {
			// nothing else should start with a dot.
			panic(fmt.Sprintf(
				"illegal expression %s",
				expr))
		}
	case '[': // array or slice prefix.
		if expr[1] == ']' {
			return SliceT(expr[2:])
		} else {
			idx := strings.Index(expr, "]")
			if idx == -1 {
				panic(fmt.Sprintf(
					"mismatched '[' in slice expr %v",
					expr))
			}
			return ArrayT(expr[1:idx], expr[idx+1:])
		}
	}
	// Numeric int?  We do these before dots, because dots are legal in numbers.
	isInt := isIntRegex.Match([]byte(expr))
	if isInt {
		return &BasicLitExpr{
			Kind:  INT,
			Value: expr,
		}
	}
	// Numeric float?  We do these before dots, because dots are legal in floats.
	isFloat := isFloatRegex.Match([]byte(expr))
	if isFloat {
		return &BasicLitExpr{
			Kind:  FLOAT,
			Value: expr,
		}
	}
	// Last case, handle dots.
	// It's last, meaning it's got the highest precedence.
	if idx := strings.LastIndex(expr, "."); idx != -1 {
		return &SelectorExpr{
			X:   X(expr[:idx]),
			Sel: N(expr[idx+1:]),
		}
	}
	return Nx(expr)
}

const (
	DGTS = `(?:[0-9]+)`
	HExX = `(?:0[xX][0-9a-fA-F]+)`
	PSCI = `(?:[eE]+?[0-9]+)`
	NSCI = `(?:[eE]-[1-9][0-9]+)`
	ASCI = `(?:[eE][-+]?[0-9]+)`
)

var isIntRegex = regexp.MustCompile(
	`^-?(?:` +
		DGTS + `|` +
		HExX + `)` + PSCI + `?$`,
)

var isFloatRegex = regexp.MustCompile(
	`^-?(?:` +
		DGTS + `\.` + DGTS + ASCI + `?|` +
		DGTS + NSCI + `)$`,
)

// Returns idx=-1 if not a binary operator.
// Precedence    Operator
//
//	5             *  /  %  <<  >>  &  &^
//	4             +  -  |  ^
//	3             ==  !=  <  <=  >  >=
//	2             &&
//	1             ||
var sp = " "

var (
	prec5 = strings.Split("*  /  %  <<  >>  &  &^", sp)
	prec4 = strings.Split("+ - | ^", sp)
	prec3 = strings.Split("== != < <= > >=", sp)
	prec2 = strings.Split("&&", sp)
	prec1 = strings.Split("||", sp)
	precs = [][]string{prec1, prec2, prec3, prec4, prec5}
)

// 0 for prec1... -1 if no match.
func lowestMatch(op string) int {
	for i, prec := range precs {
		for _, op2 := range prec {
			if op == op2 {
				return i
			}
		}
	}
	return -1
}

func Ss(b ...Stmt) []Stmt {
	return b
}

func Xs(exprs ...Expr) []Expr {
	return exprs
}

// Usage: A(lhs1, lhs2, ..., ":=", rhs1, rhs2, ...)
// Operation can be like ":=", "=", "+=", etc.
// Other strings are automatically parsed as X(arg).
func A(args ...interface{}) *AssignStmt {
	lhs := []Expr(nil)
	op := ILLEGAL
	rhs := []Expr(nil)

	setOp := func(w Word) {
		if op != ILLEGAL {
			panic("too many assignment operators")
		}
		op = w
	}

	for _, arg := range args {
		if s, ok := arg.(string); ok {
			switch s {
			case "=", ":=", "+=", "-=", "*=", "/=", "%=",
				"&=", "|=", "^=", "<<=", ">>=", "&^=":
				setOp(Op2Word(s))
				continue
			default:
				arg = X(s)
			}
		}
		// append to lhs or rhs depending on op.
		if op == ILLEGAL {
			lhs = append(lhs, arg.(Expr))
		} else {
			rhs = append(rhs, arg.(Expr))
		}
	}

	return &AssignStmt{
		Op:  op,
		Lhs: lhs,
		Rhs: rhs,
	}
}

func Not(x Expr) *UnaryExpr {
	return &UnaryExpr{
		Op: Op2Word("!"),
		X:  x,
	}
}

// Binary expression.  x, y can be Expr or string.
func Bx(lx interface{}, op string, rx interface{}) Expr {
	return &BinaryExpr{
		Left:  Xold(lx),
		Op:    Op2Word(op),
		Right: Xold(rx),
	}
}

func newBx(l Expr, op Word, r Expr) Expr { return &BinaryExpr{Left: l, Op: op, Right: r} }

func Call(fn interface{}, args ...interface{}) *CallExpr {
	argz := make([]Expr, len(args))
	for i := 0; i < len(args); i++ {
		argz[i] = X(args[i])
	}
	return &CallExpr{
		Func: X(fn),
		Args: argz,
	}
}

func TypeAssert(x interface{}, t interface{}) *TypeAssertExpr {
	return &TypeAssertExpr{
		X:    X(x),
		Type: X(t),
	}
}

func Sel(x interface{}, sel interface{}) *SelectorExpr {
	return &SelectorExpr{
		X:   X(x),
		Sel: N(sel),
	}
}

func Idx(x interface{}, idx interface{}) *IndexExpr {
	return &IndexExpr{
		X:     X(x),
		Index: X(idx),
	}
}

func Str(s string) *BasicLitExpr {
	return &BasicLitExpr{
		Kind:  STRING,
		Value: strconv.Quote(s),
	}
}

func Num(s string) *BasicLitExpr {
	return &BasicLitExpr{
		Kind:  INT,
		Value: s,
	}
}

func Ref(x interface{}) *RefExpr {
	return &RefExpr{
		X: X(x),
	}
}

func Deref(x interface{}) *StarExpr {
	return &StarExpr{
		X: X(x),
	}
}

// NOTE: Same as DEREF, but different context.
func Ptr(x interface{}) *StarExpr {
	return &StarExpr{
		X: X(x),
	}
}

// ----------------------------------------
// AST Construction (Stmt)

func If(cond Expr, b ...Stmt) *IfStmt {
	return &IfStmt{
		Cond: cond,
		Then: IfCaseStmt{Body: b},
	}
}

func IfElse(cond Expr, bdy, els Stmt) *IfStmt {
	var body []Stmt
	if bdy, ok := bdy.(*BlockStmt); !ok {
		body = bdy.Body
	} else {
		body = []Stmt{bdy}
	}
	var els_ []Stmt
	if els, ok := els.(*BlockStmt); !ok {
		els_ = els.Body
	} else {
		els_ = []Stmt{els}
	}
	return &IfStmt{
		Cond: cond,
		Then: IfCaseStmt{Body: body},
		Else: IfCaseStmt{Body: els_},
	}
}

func Return(results ...Expr) *ReturnStmt {
	return &ReturnStmt{
		Results: results,
	}
}

func Continue(label interface{}) *BranchStmt {
	return &BranchStmt{
		Op:    CONTINUE,
		Label: N(label),
	}
}

func Break(label interface{}) *BranchStmt {
	return &BranchStmt{
		Op:    BREAK,
		Label: N(label),
	}
}

func Goto(label interface{}) *BranchStmt {
	return &BranchStmt{
		Op:    GOTO,
		Label: N(label),
	}
}

func Fallthrough(label interface{}) *BranchStmt {
	return &BranchStmt{
		Op:    FALLTHROUGH,
		Label: N(label),
	}
}

func ImportD(name interface{}, path string) *ImportDecl {
	return &ImportDecl{
		NameExpr: *Nx(name),
		PkgPath:  path,
	}
}

func For(init, cond, post interface{}, b ...Stmt) *ForStmt {
	return &ForStmt{
		Init: S(init).(SimpleStmt),
		Cond: X(cond),
		Post: S(post).(SimpleStmt),
		Body: b,
	}
}

func Loop(b ...Stmt) *ForStmt {
	return For(nil, nil, nil, b...)
}

func Once(b ...Stmt) *ForStmt {
	b = append(b, Break(""))
	return For(nil, nil, nil, b...)
}

func Len(x Expr) *CallExpr {
	return Call(Nx("len"), x)
}

func Var(name interface{}, typ Expr, value Expr) *DeclStmt {
	return &DeclStmt{
		Body: []Stmt{&ValueDecl{
			NameExprs: []NameExpr{*Nx(name)},
			Type:      typ,
			Values:    []Expr{value},
			Const:     false,
		}},
	}
}

func Inc(x interface{}) *IncDecStmt {
	var xx Expr
	if xs, ok := x.(string); ok {
		xx = X(xs)
	}
	return &IncDecStmt{
		X:  xx,
		Op: INC,
	}
}

func Dec(x interface{}) *IncDecStmt {
	var xx Expr
	if xs, ok := x.(string); ok {
		xx = X(xs)
	}
	return &IncDecStmt{
		X:  xx,
		Op: DEC,
	}
}

func Op2Word(op string) Word {
	switch op {
	case "+":
		return ADD
	case "-":
		return SUB
	case "*":
		return MUL
	case "/":
		return QUO
	case "%":
		return REM
	case "&":
		return BAND
	case "|":
		return BOR
	case "^":
		return XOR
	case "<<":
		return SHL
	case ">>":
		return SHR
	case "&^":
		return BAND_NOT
	case "&&":
		return LAND
	case "||":
		return LOR
	case "<-":
		return ARROW
	case "++":
		return INC
	case "--":
		return DEC
	case "==":
		return EQL
	case "<":
		return LSS
	case ">":
		return GTR
	case "!":
		return NOT
	case "!=":
		return NEQ
	case "<=":
		return LEQ
	case ">=":
		return GEQ
	// Assignment
	case "=":
		return ASSIGN
	case ":=":
		return DEFINE
	case "+=":
		return ADD_ASSIGN
	case "-=":
		return SUB_ASSIGN
	case "*=":
		return MUL_ASSIGN
	case "/=":
		return QUO_ASSIGN
	case "%=":
		return REM_ASSIGN
	case "&=":
		return BAND_ASSIGN
	case "|=":
		return BOR_ASSIGN
	case "^=":
		return XOR_ASSIGN
	case "<<=":
		return SHL_ASSIGN
	case ">>=":
		return SHR_ASSIGN
	case "&^=":
		return BAND_NOT_ASSIGN
	default:
		panic("unrecognized binary/unary/assignment operator " + op)
	}
}

// ----------------------------------------
// AST Static (compile time)

func SIf(cond bool, then_, else_ Stmt) Stmt {
	if cond {
		return then_
	} else if else_ != nil {
		return else_
	} else {
		return &EmptyStmt{}
	}
}

// ----------------------------------------
// chop functions

// ----------------------------------------
func chopBinary(expr string) (left, op, right string, ok bool) {
	// 0 for prec1... -1 if no match.
	matchOp := func(op string) int {
		for i, prec := range precs {
			for _, op2 := range prec {
				if op == op2 {
					return i
				}
			}
		}
		return -1
	}
	ss := newScanner(expr)
	lowestMatch := 0xff
	for !ss.advance() {
		if ss.out() {
			// find match starting with longer operators.
			for ln := 2; ln > 0; ln-- {
				op2 := ss.peek(ln)
				match := matchOp(op2)
				if match != -1 {
					if match <= lowestMatch {
						ok = true
						lowestMatch = match
						left = string(ss.rnz[:ss.idx])
						op = op2
						// NOTE: `op2` may be shorter than ln.
						// NOTE: assumes operators are ascii chars.
						right = string(ss.rnz[ss.idx+len(op2):])
					}
					// advance, so we don't match a substring
					// operator.
					for i := 0; i < len(op2); i++ {
						ss.advance()
					}
					break
					// Do not return here, we want to find the last
					// match.  But don't consider shorter operators.
				}
				if len(op2) == 1 {
					// nothing more to read.
					break
				}
			}
		}
	}
	if !ss.out() {
		return "", "", "", false
	}
	return
}

// Given that 'in' ends with ')', '}', or ']',
// find the matching opener, while processing escape
// sequences of strings and rune literals.
// `tok` is the corresponding opening rune.
// `right` excludes the last character (closer).
func chopRight(in string) (left string, tok rune, right string) {
	switch in[len(in)-1] {
	case '}', ')', ']':
		// good
	default:
		panic("input doesn't start with brace: " + in)
	}
	ss := newScanner(in)
	lastOut := 0 // last position where out.
	for !ss.advance() {
		if ss.out() {
			lastOut = ss.idx
		}
	}
	if !ss.out() {
		panic("mismatched braces/brackets")
	} else {
		left = string(ss.rnz[:lastOut+1])
		tok = ss.rnz[lastOut+1]
		right = string(ss.rnz[lastOut+2 : len(in)-2])
		return
	}
}

// Rewrite of X() with Joeson
func X(x interface{}, args ...interface{}) Expr {
	fmt.Printf("| Initially, X(x=%s)\n", x)
	switch cx := x.(type) {
	case Expr:
		return cx
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64:
		return Xold(fmt.Sprintf("%v", x))
	case string:
		if cx == "" {
			panic("input cannot be blank for X()")
		}
	case Name:
		if cx == "" {
			panic("input cannot be blank for X()")
		}
		x = string(cx)
	default:
		panic(fmt.Sprintf("unexpected input type for X(): %T", x))
	}
	expr := x.(string)
	fmt.Printf("| x.(string)=%s\n", expr)
	expr = fmt.Sprintf(expr, args...)
	fmt.Printf("| fmt.Sprintf(expr, args...)=%s\n", expr)
	expr = strings.TrimSpace(expr)
	// first := expr[0]

	// return Xold(x, args...)
	//
	ast := grammar.ParseString(expr)
	if j.IsParseError(ast) {
		panic(ast.ContentString())
	} else {
		switch v := ast.(type) {
		case w:
			// just unwrap ast
			// fmt.Println(ast.ContentString())
			return v.expr

		default:
			panic(fmt.Sprintf("From X(): X() can not extract an expr from %T. \nContentString(): %s.", ast, ast.ContentString()))
		}
	}
}

// TODO find where to initialize it
func initGrammar() {
	grammar = j.GrammarFromLines(gnoRules, "GNO-grammar")
}

// wrap an Expr inside a joeson.Ast.
// to be Expr, one needs to be Node and have:
// - assertExpr()
// - assertNode()
// - String() string
// - Copy() Node
// - GetLine() int
// - SetLine(int)
// - GetLabel() Name
// - SetLabel(Name)
// - {Has/Get/Set}Attribute
// Here choose to just wrap it.
func expr2Ast(expr Expr) j.Ast {
	return w{expr}
}

type w struct {
	expr Expr
}

func (ww w) ContentString() string { return ww.expr.String() }

// rules and grammar for GNO

func i(a ...any) j.ILine                       { return j.I(a...) }
func o(a ...any) j.OLine                       { return j.O(a...) }
func rules(a ...j.Line) []j.Line               { return a }
func named(name string, thing any) j.NamedRule { return j.Named(name, thing) }

// let's have Expr satisfy joeson.Ast
// func (e *Expr) ContentString() string { return "TODO switch and show, BinaryExpr etc. See nodes.go" }

/*
Primary expressions

Primary expressions are the operands for unary and binary expressions.

PrimaryExpr =
        Operand |			// see spec/Operand.txt or https://go.dev/ref/spec#Operand
        Conversion |
        MethodExpr |
        PrimaryExpr Selector |
        PrimaryExpr Index |
        PrimaryExpr Slice |
        PrimaryExpr TypeAssertion |
        PrimaryExpr Arguments .

Selector       = "." identifier .
Index          = "[" Expression [ "," ] "]" .
Slice          = "[" [ Expression ] ":" [ Expression ] "]" |
                 "[" [ Expression ] ":" Expression ":" Expression "]" .
TypeAssertion  = "." "(" Type ")" .
Arguments      = "(" [ ( ExpressionList | Type [ "," ExpressionList ] ) [ "..." ] [ "," ] ] ")" .
*/

var (
	grammar  *j.Grammar
	gnoRules = rules(
		o(named("Input", "Expression")),
		o(named("Expression", "bx:(Expression _ binary_op _ Expression) | UnaryExpr"), fExpression),
		o(named("UnaryExpr", "PrimaryExpr | unary_op _ UnaryExpr")),
		o(named("unary_op", revQuote("+ - ! ^ * & <-"))),
		o(named("binary_op", "mul_op | add_op | rel_op | '&&' | '||'")),
		o(named("mul_op", revQuote("* / % << >> & &^"))),
		o(named("add_op", revQuote("+ - | ^"))),
		o(named("rel_op", revQuote("== != < <= > >="))),
		// o(named("PrimaryExpr", "Operand | Conversion | MethodExpr | PrimaryExpr _ ( Selector | Index | Slice | TypeAssertion | Arguments )")),
		o(named("PrimaryExpr", "Operand | PrimaryExpr _ ( Selector | Index | Slice | TypeAssertion | Arguments )")),

		o(named("Operand", rules(
			// o("'(' _ Expression _ ')' | OperandName TypeArgs? | Literal"), // TODO this is the original
			o("Literal | '(' _ Expression _ ')'"),
			// o(named("Literal", "BasicLit | CompositeLit | FunctionLit")),
			o(named("Literal", "BasicLit")),
			// TODO add float_lit and imaginary_lit
			// o(named("BasicLit", "int_lit | rune_lit | string_lit")),
			o(named("BasicLit", "decimal_lit")),                   // NOTE it is escamoted, normally there is int_lit layer
			i(named("decimal_lit", "/^0|[1-9](_?[0-9])*/"), fInt), // x("decimal_lit")),
			// o(named("OperandName", "QualifiedIdent | identifier")),
			// i(named("QualifiedIdent", "PackageName '.' identifier"), x("QualifiedIdent")), // https://go.dev/ref/spec#QualifiedIdent
			// i(named("PackageName", "identifier")),                                         // https://go.dev/ref/spec#PackageName
			// o(named("Block", "'{' Statement*';' '}'")),
		))),

		// o(named("Expr", "BinaryExpr | aadigits")),
		// o(named("BinaryExpr", "l:aadigits _ op:Word _ r:Expr"), fBinaryExpr),
		// o(named("Word", "'"+strings.Join(strings.Fields("+ - * / % & | ^ << >> &^ && || ++ -- == < > ! != <= >= = := += -= *= /= %= &= |= ^= <<= >>= &^="), "'|'")+"'")),
		// i(named("aadigits", "/^-?[0-9]+/"), fInt),

		// "White space, formed from spaces (U+0020), horizontal tabs (U+0009),
		// carriage returns (U+000D), and newlines (U+000A), is ignored except as
		// it separates tokens that would otherwise combine into a single token."
		i(named("_", "/[ \t\n\r]*/")),
		i(named("__", "/[ \t\n\r]+/")),
	)
)

// Facilitates writing rules for PEG grammars.
// It splits upon space, reverse order, adds single quotes, and joins upon '|'
// For example:
//
// "* / %"      becomes      "'%'|'/'|'*'".
func revQuote(spaceSeparatedElements string) string {
	a := strings.Fields(spaceSeparatedElements)
	s := ""
	for i := len(a) - 1; i >= 0; i-- {
		s += "'" + a[i] + "'|"
	}
	return s[:len(s)-1]
}

// builder function

func fInt(it j.Ast) j.Ast { return expr2Ast(Num(it.(j.NativeString).Str)) }

// func fBinaryExpr(it j.Ast) j.Ast {
// 	m := it.(j.NativeMap)
// 	lhs, b1 := m.GetExists("l")
// 	op_, b2 := m.GetStringExists("op")
// 	rhs, b3 := m.GetExists("r")
// 	if b1 && b2 && b3 {
// 		return expr2Ast(newBx(lhs.(w).expr, op_, rhs.(w).expr))
// 	} else {
// 		panic("assert")
// 	}
// }

func fExpression(it j.Ast) j.Ast {
	if m, ok := it.(j.NativeMap); ok {
		a := m.GetOrPanic("bx").(*j.NativeArray).Array
		return expr2Ast(newBx(a[0].(w).expr, Op2Word(a[1].(j.NativeString).Str), a[2].(w).expr))
	} else {
		return it // Unary
	}
}

// function x() helps to quickly write a grammar.
// Calling x("foo") returns a callback `func(τ Ast) Ast`.
// Calling cb.ContentString() gives "<foo:" + τ.ContentString() + ">"
//
// For example:
//
// var rules_tokens = rules(
//
//	o(named("token", "( keyword | identifier | operator | punctuation | literal )"), x("token")),
//	i(named("keyword", "( 'break' | 'default' | 'func' | 'interface' | 'select' | 'case' | 'defer' | 'go' | 'map' | 'struct' | 'chan' | 'else' | 'goto' | 'package' | 'switch' | 'const' | 'fallthrough' | 'if' | 'range' | 'type' | 'continue' | 'for' | 'import' | 'return' | 'var' )"), x("keyword")),
//	i(named("identifier", "[a-zA-Z_][a-zA-Z0-9_]*"), x("identifier")), // letter { letter | unicode_digit } .   We rewrite it so to accelerate parsing
//	i(named("operator", "( '+' | '&' | '+=' | '&=' | '&&' | '==' | '!=' | '(' | ')' | '-' | '|' | '-=' | '|=' | '||' | '<' | '<=' | '[' |  ']' | '*' | '^' | '*=' | '^=' | '<-' | '>' | '>=' | '{' | '}' | '/' | '<<' | '/=' | '<<=' | '++' | '=' | ':=' | '%' | '>>' | '%=' | '>>=' | '--' | '!' | '...' | '&^' | '&^=' | '~' )"), x("operator")),
//
// ...
// )
//
// Here, whichever of keyword, identifier etc gets built,
// its ContentString() will be like "<token:keyword>", "<token:identifier>" etc.
func x(typename string) func(j.Ast) j.Ast {
	return func(ast j.Ast) j.Ast {
		return dumb{typename, ast}
	}
}

// type dumb is used by x(). As the name hints, it's nothing too exciting
type dumb struct {
	typename string
	ast      j.Ast
}

func (dumb dumb) ContentString() string {
	return "<" + dumb.typename + ":" + dumb.ast.ContentString() + ">"
}

// ∎
