package gnolang

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// ----------------------------------------
// AST Construction (Expr)
// These are copied over from go-amino-x, but produces Gno ASTs.

func N(n any) Name {
	switch n := n.(type) {
	case string:
		return Name(n)
	case Name:
		return n
	default:
		panic("unexpected name type")
	}
}

func Nx(n any) *NameExpr {
	return &NameExpr{Name: N(n)}
}

func ArrayT(l, elt any) *ArrayTypeExpr {
	return &ArrayTypeExpr{
		Len: X(l),
		Elt: X(elt),
	}
}

func SliceT(elt any) *SliceTypeExpr {
	return &SliceTypeExpr{
		Elt: X(elt),
		Vrd: false,
	}
}

func MapT(key, value any) *MapTypeExpr {
	return &MapTypeExpr{
		Key:   X(key),
		Value: X(value),
	}
}

func Vrd(elt any) *SliceTypeExpr {
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

func Flds(args ...any) FieldTypeExprs {
	list := FieldTypeExprs{}
	for i := 0; i < len(args); i += 2 {
		list = append(list, FieldTypeExpr{
			NameExpr: *Nx(args[i]),
			Type:     X(args[i+1]),
		})
	}
	return list
}

func Fld(n, t any) FieldTypeExpr {
	return FieldTypeExpr{
		NameExpr: *Nx(n),
		Type:     X(t),
	}
}

func Recv(n, t any) FieldTypeExpr {
	if n == "" {
		n = blankIdentifier
	}
	return FieldTypeExpr{
		NameExpr: *Nx(n),
		Type:     X(t),
	}
}

// FuncD creates a new function declaration.
//
// There is a difference between passing nil to body or passing []Stmt{}:
// nil means that the curly brackets are missing in the source code, indicating
// a declaration for an externally-defined function, while []Stmt{} is simply a
// functions with no statements (func() {}).
func FuncD(name any, params, results FieldTypeExprs, body []Stmt) *FuncDecl {
	return &FuncDecl{
		NameExpr: *Nx(name),
		Type: FuncTypeExpr{
			Params:  params,
			Results: results,
		},
		Body: body,
	}
}

func MthdD(name any, recv FieldTypeExpr, params, results FieldTypeExprs, body []Stmt) *FuncDecl {
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

func Kv(n, v any) KeyValueExpr {
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
func S(args ...any) Stmt {
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
func X(x any, args ...any) Expr {
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
		panic("unexpected input type for X()")
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

func Ss(b ...Stmt) []Stmt {
	return b
}

func Xs(exprs ...Expr) []Expr {
	return exprs
}

// Usage: A(lhs1, lhs2, ..., ":=", rhs1, rhs2, ...)
// Operation can be like ":=", "=", "+=", etc.
// Other strings are automatically parsed as X(arg).
func A(args ...any) *AssignStmt {
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
func Bx(lx any, op string, rx any) Expr {
	return &BinaryExpr{
		Left:  X(lx),
		Op:    Op2Word(op),
		Right: X(rx),
	}
}

func Call(fn any, args ...any) *CallExpr {
	argz := make([]Expr, len(args))
	for i := range args {
		argz[i] = X(args[i])
	}
	return &CallExpr{
		Func: X(fn),
		Args: argz,
	}
}

func TypeAssert(x any, t any) *TypeAssertExpr {
	return &TypeAssertExpr{
		X:    X(x),
		Type: X(t),
	}
}

func Sel(x any, sel any) *SelectorExpr {
	return &SelectorExpr{
		X:   X(x),
		Sel: N(sel),
	}
}

func Idx(x any, idx any) *IndexExpr {
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

func Ref(x any) *RefExpr {
	return &RefExpr{
		X: X(x),
	}
}

func Deref(x any) *StarExpr {
	return &StarExpr{
		X: X(x),
	}
}

// NOTE: Same as DEREF, but different context.
func Ptr(x any) *StarExpr {
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

func Continue(label any) *BranchStmt {
	return &BranchStmt{
		Op:    CONTINUE,
		Label: N(label),
	}
}

func Break(label any) *BranchStmt {
	return &BranchStmt{
		Op:    BREAK,
		Label: N(label),
	}
}

func Goto(label any) *BranchStmt {
	return &BranchStmt{
		Op:    GOTO,
		Label: N(label),
	}
}

func Fallthrough(label any) *BranchStmt {
	return &BranchStmt{
		Op:    FALLTHROUGH,
		Label: N(label),
	}
}

func ImportD(name any, path string) *ImportDecl {
	return &ImportDecl{
		NameExpr: *Nx(name),
		PkgPath:  path,
	}
}

func For(init, cond, post any, b ...Stmt) *ForStmt {
	return &ForStmt{
		Init:      S(init).(SimpleStmt),
		Cond:      X(cond),
		Post:      S(post).(SimpleStmt),
		BodyBlock: &BlockStmt{Body: b},
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

func Var(name any, typ Expr, value Expr) *DeclStmt {
	return &DeclStmt{
		Body: []Stmt{&ValueDecl{
			NameExprs: []NameExpr{*Nx(name)},
			Type:      typ,
			Values:    []Expr{value},
			Const:     false,
		}},
	}
}

func Inc(x any) *IncDecStmt {
	var xx Expr
	if xs, ok := x.(string); ok {
		xx = X(xs)
	}
	return &IncDecStmt{
		X:  xx,
		Op: INC,
	}
}

func Dec(x any) *IncDecStmt {
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
// AST Query

// Unwraps IndexExpr and SelectorExpr only.
// (defensive, in case malformed exprs that mix).
func LeftmostX(x Expr) Expr {
	for {
		switch x := x.(type) {
		case *IndexExpr:
			return x.X
		case *SelectorExpr:
			return x.X
		default:
			return x
		}
	}
}

// ----------------------------------------
// chop functions

// ----------------------------------------
func chopBinary(expr string) (left, op, right string, ok bool) {
	// 0 for prec1... -1 if no match.
	matchOp := func(op string) int {
		for i, prec := range precs {
			if slices.Contains(prec, op) {
				return i
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
					for range len(op2) {
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

type StmtInsertion struct {
	stmt Stmt
	idx  int // position to insert
}

func addStmtInsertionAttr(bn BlockNode, si *StmtInsertion) {
	var (
		sis   []*StmtInsertion
		found bool
	)
	if sis, found = bn.GetAttribute(ATTR_CONTINUE_INSERT).([]*StmtInsertion); !found {
		sis = append(sis, si)
	} else {
		if slices.Contains(sis, si) {
			return
		}
		sis = append(sis, si)
	}
	bn.SetAttribute(ATTR_CONTINUE_INSERT, sis)
}

func getStmtInsertionAttr(bn BlockNode) ([]*StmtInsertion, bool) {
	if sis, found := bn.GetAttribute(ATTR_CONTINUE_INSERT).([]*StmtInsertion); found {
		return sis, true
	}
	return nil, false
}

func addLoopvarAttrs(bn BlockNode, key GnoAttribute, names ...Name) {
	var (
		ns    []Name
		found bool
	)
	if ns, found = bn.GetAttribute(key).([]Name); !found {
		ns = append(ns, names...)
	} else {
		for _, n := range names {
			if slices.Contains(ns, n) {
				return
			} else {
				ns = append(ns, n)
			}
		}
	}

	bn.SetAttribute(key, ns)
}

func hasLoopvarAttrs(fs *ForStmt, n Name, key GnoAttribute) bool {
	if ns, ok := fs.GetAttribute(key).([]Name); ok {
		if slices.Contains(ns, n) {
			return true
		}
	}
	return false
}

func getLoopvarAttrs(bn BlockNode, key GnoAttribute) []Name {
	if names, ok := bn.GetAttribute(key).([]Name); ok {
		return names
	}
	return nil
}
