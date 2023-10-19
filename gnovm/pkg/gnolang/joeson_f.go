package gnolang

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	j "github.com/grepsuzette/joeson"
)

// Parser functions (rule callbacks)
// Naming convention:
// - fXxx(it Ast[, *ParseContext]) Ast // for a rule named "Xxx".
// - fxxx(it Ast[, *ParseContext]) Ast // for a rule named "xxx".
// - ffXxx(someArg) func(it Ast, *ParseContext) Ast // this returns a function

// About panics:
//
// - panic("assert") denotes an unreachable code (i.e. it's useless to elaborate).
//
// - panic(msg) is an unexpected error but will panic with a message to help
//  the rule implementors. One day, those may as well become panic("assert").
//
// - panic(ctx.Error(msg)) is for ParseError, and will not panic (recovered). It
// will stop parsing rules right away.  A rule callback may simply employ
// ffPanic below that uses this way but inline:
// ```
// i(named(
//	"octal_byte_value_err1", `a:'\\' (?octal_digit{4,})`),
//	ffPanic("illegal: too many octal digits"),
// ),
// ```
//
// Rules written with joeson can produce NativeString, *NativeMap, *NativeArray
// and NativeInt depending on the rule itself. For instance for the following
// rule: o(`p:PrimaryExpr a:Arguments`, fPrimaryExprArguments)
//
// The fPrimaryExprArguments parser function will receive a NativeMap
// with keys "p" and "a". It is customary to access them in a forceful way
// like this: `primary := it.(*j.NativeMap).GetOrPanic("p")`.
//
// This is because both rules and parser functions are supposed to
// go hand in hand and hard assumptions are therefore perfectly normal
// (a parser function receiving an unexpected input is facing a broken grammar,
// and there is no way to recover from this, at least from the rules + parser
// functions themselves).

// Panic with a ParseError made from msg string.
// ParseErrors panics are recovered higher up, in parseX().
// This is a way to give a custom, fine error inline from
// a rule, or you may choose to panic(ctx.Error(msg)) if
// within a parser function.
func ffPanic(msg string) func(j.Ast, *j.ParseContext) j.Ast {
	return func(it j.Ast, ctx *j.ParseContext) j.Ast {
		panic(ctx.Error(msg))
	}
}

// Like ffPanic, but the text of the current line
// from the parse context is postpended to `msg`.
func ffPanicNearContext(msg string) func(j.Ast, *j.ParseContext) j.Ast {
	return func(it j.Ast, ctx *j.ParseContext) j.Ast {
		panic(ctx.Error(msg + ": " + ctx.Code.PeekLines(-1, 1)))
	}
}

// Helper
// peel([ [ a,b,... ] ]) -> [a,b,...]
// Assert `it` is NativeArray.
// Useful when rules would create two or more levels of NativeArray.
func peel(it j.Ast, ctx *j.ParseContext) j.Ast {
	// (don't assert size of exactly 1,
	// case like [[a,b,..], NativeUndefined{}]) must also work)
	return it.(*j.NativeArray).Get(0)
}

// grow the moss. Different levels of precedence
// must call this function separately (if you
// have 5 levels of binary operator precedence
// then you must have 5 different rules calling
// growMoss).  In go that number is 5 indeed:
//
// Precedence    Operator
//
//	5             *  /  %  <<  >>  &  &^
//	4             +  -  |  ^
//	3             ==  !=  <  <=  >  >=
//	2             &&
//	1             ||
//
// The moss grows laterally.
// The moss families don't intermix.
//
// Given `(first op1 e1 op2 e2 op3 e3 ...)`
// Where first, e1, e2, e3, ... ∈ Expr
// Where op1, op2, op3, ... are binary operators of the SAME precedence
// This creates the following recursive BinaryExpr
//
// :	       [op1 first e1]
// :	  [op2 [op1 first e1] e2]
// : [op3 [op2 [op1 first e1] e2] e3]       etc.
//
// `[op a b]` here is not a array but
// a short way for `BinaryExpr{Left:a, Op:op, Right:b}`.
// See study in joeson/examples/precedence
// See joeson_rules.go
func growMoss(it j.Ast) j.Ast {
	switch v := it.(type) {
	case j.NativeString:
		return v
	}
	first := it.(*j.NativeArray).Get(0)
	rest := it.(*j.NativeArray).Get(1)
	operations := []*BinaryExpr{}
	for _, v := range rest.(*j.NativeArray).Array() {
		switch w := v.(type) {
		case *j.NativeArray:
			a := w.Array()
			operations = append(operations, &BinaryExpr{
				Left:  nil,
				Op:    Op2Word(a[0].String()),
				Right: a[1].(Expr),
			})
		case *BinaryExpr:
			return w
		default:
			panic("assert type=" + reflect.TypeOf(w).String() + " String()=" + w.String())
		}
	}
	if len(operations) == 0 {
		return first
	}
	// The moss grows laterally:
	//
	//           [op1 first e1]
	//      [op2 [op1 first e1] e2]
	// [op3 [op2 [op1 first e1] e2] e3]       etc.
	moss := first
	for _, operation := range operations {
		moss = &BinaryExpr{
			Left:  moss.(Expr),
			Op:    operation.Op,
			Right: operation.Right,
		}
	}
	return moss
}

// parser functions (rule callbacks) ------------------
// these allow to map the Ast produced automatically
// by the joeson rules into something else (usually,
// something like a gnolang.Expr, that implements both
// joeson.Ast and gnolang.Node).

func ffInt(base int) func(j.Ast, *j.ParseContext) j.Ast {
	return func(it j.Ast, ctx *j.ParseContext) j.Ast {
		var s string
		var e error
		switch v := it.(type) {
		case j.NativeString:
			s = v.String()
		case *j.NativeArray:
			s = v.Concat()
		default:
			panic("unsupported type " + reflect.TypeOf(it).String() + " in ffInt")
		}
		if strings.HasSuffix(s, "_") {
			panic(ctx.Error("invalid: _ must separate successive digits"))
		}
		s = strings.ReplaceAll(s, "_", "")
		var i int64
		var prefix string
		switch base {
		case 2:
			i, e = strconv.ParseInt(s, 2, 64)
			prefix = "0b"
		case 8:
			if strings.HasPrefix(s, "0o") || strings.HasPrefix(s, "0O") {
				i, e = strconv.ParseInt(s[2:], 8, 64)
			} else if strings.HasPrefix(s, "0") {
				i, e = strconv.ParseInt(s[1:], 8, 64) // 0177
			}
			prefix = "0o"
		case 10:
			i, e = strconv.ParseInt(s, 10, 64)
			prefix = ""
		case 16:
			i, e = strconv.ParseInt(s, 16, 64)
			prefix = "0x"
		default:
			panic("impossible base, expecting 2,8,10,16")
		}
		if e != nil {
			// it may have overflowed or faulty grammar.
			return ctx.Error(e.Error())
		}
		return &BasicLitExpr{
			Kind:  INT,
			Value: prefix + strconv.FormatInt(i, base),
		}
	}
}

// it simply creates shortcut functions for FLOAT BasicLitExpr
// using Sprintf with several formats like "%g" etc.
func ffFloatFormat(format string) func(j.Ast, *j.ParseContext) j.Ast {
	// format for floats:
	// %f	decimal point but no exponent, e.g. 123.456
	// %F	synonym for %f
	// %e	scientific notation, e.g. -1.234456e+78
	// %E	scientific notation, e.g. -1.234456E+78
	// %g	%e for large exponents, %f otherwise. Precision is discussed below.
	// %G	%E for large exponents, %F otherwise
	return func(it j.Ast, ctx *j.ParseContext) j.Ast {
		s := it.(*j.NativeArray).Concat()
		if f, err := strconv.ParseFloat(s, 64); err != nil {
			return ctx.Error(fmt.Sprintf("%s did not parse as a Float, err=%s", s, err.Error()))
		} else {
			return &BasicLitExpr{
				Kind:  FLOAT,
				Value: fmt.Sprintf(format, f),
			}
		}
	}
}

func fImaginary(it j.Ast, ctx *j.ParseContext) j.Ast {
	return &BasicLitExpr{
		Kind:  IMAG,
		Value: it.(*j.NativeArray).Concat(),
	}
}

func f_rune_lit(it j.Ast, ctx *j.ParseContext) j.Ast {
	return ffBasicLit(CHAR)(it, ctx)
}

func ff_u_value(rule string, hexDigits int) func(j.Ast, *j.ParseContext) j.Ast {
	return func(it j.Ast, ctx *j.ParseContext) j.Ast {
		if it.(*j.NativeMap).GetOrPanic("b").(*j.NativeArray).Length() != hexDigits {
			return ctx.Error(fmt.Sprintf("%s requires %d hex", rule, hexDigits))
		} else {
			return it
		}
	}
}

func f_raw_string_lit(it j.Ast, ctx *j.ParseContext) j.Ast {
	return ffBasicLit(STRING)(it, ctx)
}

func foctal_byte_value(it j.Ast, ctx *j.ParseContext) j.Ast {
	if n, ok := it.(*j.NativeMap).GetIntExists("b"); ok {
		if n < 0 || n > 255 {
			return ffPanic("illegal: octal value over 255")(it, ctx)
		} else {
			return it
		}
	}
	panic(fmt.Sprintf("unexpected type %s", reflect.TypeOf(it).String()))
}

// Returns &BasicLitExpr
func finterpreted_string_lit(it j.Ast, ctx *j.ParseContext) j.Ast {
	return &BasicLitExpr{
		Kind:  STRING,
		Value: `"` + it.(*j.NativeArray).Concat() + `"`,
	}
}

// Returns &BasicLitExpr
func fraw_string_lit(it j.Ast, ctx *j.ParseContext) j.Ast {
	return &BasicLitExpr{
		Kind:  STRING,
		Value: "`" + it.(*j.NativeArray).Concat() + "`",
	}
}

func ffBasicLit(kind Word) func(j.Ast, *j.ParseContext) j.Ast {
	return func(it j.Ast, ctx *j.ParseContext) j.Ast {
		if j.IsParseError(it) {
			return it
		}
		s := ""
		switch v := it.(type) {
		case *j.NativeArray:
			s = v.Concat()
		case *j.NativeMap:
			s = v.Concat()
		case j.NativeString:
			s = string(v)
		default:
			panic(fmt.Sprintf("Unexpected type in ffBasicLit: %s", reflect.TypeOf(it).String()))
		}
		return &BasicLitExpr{
			Kind:  kind,
			Value: s,
		}
	}
}

func fSimpleStmt(it j.Ast) j.Ast {
	// TODO "The following built-in functions are not permitted in statement context:
	// append cap complex imag len make new real
	// unsafe.Add unsafe.Alignof unsafe.Offsetof unsafe.Sizeof unsafe.Slice"
	return it
}

// same as identifier (*NameExpr), but when Name
// is the blank identifier panic with a ParseError
func fPackageName(it j.Ast, ctx *j.ParseContext) j.Ast {
	if it.(*NameExpr).String() == "_" {
		panic(ctx.Error("PackageName must not be the blank identifier"))
	} else {
		// So we can check the "selector operate on primary expression
		// that is not a package name" spec from golang.
		// See fPrimaryExprSelector() or "i_m_a_package_name" in this file
		it.(*NameExpr).SetAttribute("i_m_a_package_name", "1")
		return it
	}
}

// returns a &FuncLitExpr
// "A function literal represents an anonymous function"
// TODO we have to stop here and come back after we write rules for Blocks and Stmt
// func fFunctionLit(it j.Ast) j.Ast {
// 	a := it.(*j.NativeArray)
// 	return &FuncLitExpr{
// 		// StaticBlock
// 		Type: a.Get(0).(*FuncTypeExpr), // function type
// 		Body: a.Get(1).)             // function body
// 	}
// }

// returns a &CompositeLitExpr from NativeArray<*KeyValueExpr>
func fCompositeLit(it j.Ast) j.Ast {
	a := it.(*j.NativeArray).Array()
	var elts []KeyValueExpr
	for _, elt := range a[1].(*j.NativeArray).Array() {
		elts = append(elts, *elt.(*KeyValueExpr))
	}
	return &CompositeLitExpr{
		Type: a[0].(TypeExpr),
		Elts: elts,
	}
}

// returns a &KeyValueExpr.
// The Key can be nil if it has just a Value.
func fKeyedElement(it j.Ast, ctx *j.ParseContext) j.Ast {
	a := it.(*j.NativeArray).Array()
	var k Expr
	if len(a) != 2 {
		panic("assert")
	}
	if !j.IsUndefined(a[0]) {
		k = a[0].(Expr)
	}
	return &KeyValueExpr{
		Key:   k,
		Value: a[1].(Expr),
	}
}

func fQualifiedIdent(it j.Ast, ctx *j.ParseContext) j.Ast {
	m := it.(*j.NativeMap)
	// p:PackageName DOT i:identifier
	packageName := m.GetOrPanic("p")
	identifier := m.GetOrPanic("i").(*NameExpr)
	return &SelectorExpr{
		X:   packageName.(*NameExpr),
		Sel: identifier.Name,
	}
}

// Prepare information for fPrimaryExprArguments. Returns a NativeMap.
// It resembles a CallExpr without Func:
// "Args"    Exprs        function arguments, if any.
// "Varg"	 NativeInt    if 1, final arg is variadic.
// "NumArgs" NativeInt    len(Args) or len(Args[0].Results)
func fArguments(it j.Ast, ctx *j.ParseContext) j.Ast {
	switch m := it.(type) {
	case j.NativeUndefined:
		// empty arguments, okay
		return j.NewNativeMap(map[string]j.Ast{
			"Args":    j.NewNativeArray([]j.Ast{}),
			"NumArgs": j.NewNativeInt(0),
			"Varg":    j.NewNativeInt(0),
		})
	case *j.NativeMap:
		// actual arguments.
		// Prepare information for fPrimaryExprArguments
		args := m.GetOrPanic("Args").(*j.NativeArray)
		m.Set("NumArgs", j.NewNativeInt(args.Length()))
		return m
	default:
		panic("assert")
	}
}

func fUnaryExpr(it j.Ast) j.Ast {
	// Rule: `PrimaryExpr | ux:(unary_op _ UnaryExpr)`
	if m, ok := it.(*j.NativeMap); ok {
		if ux, ok := m.GetExists("ux"); !ok {
			panic(fmt.Sprintf("key ux not found in %s", m.String()))
		} else {
			a := ux.(*j.NativeArray).Array()
			op := a[0].(j.NativeString).String()
			switch op {
			case "*":
				return &StarExpr{
					X: a[1].(Expr),
				}
			case "&":
				return &RefExpr{
					X: a[1].(Expr),
				}
			case "+", "-", "!", "^", "<-":
				return &UnaryExpr{
					Op: Op2Word(op),
					X:  a[1].(Expr),
				}
			default:
				panic("assert")
			}
		}
	} else {
		return it
	}
}

// Returns a &CallExpr
// Information ("a:Arguments") as been prepared by fArguments,
// including variadicity
func fPrimaryExprArguments(it j.Ast, ctx *j.ParseContext) j.Ast {
	m := it.(*j.NativeMap)
	primaryExpr := m.GetOrPanic("p").(Expr)
	arguments := m.GetOrPanic("a").(*j.NativeMap)
	var exprs []Expr
	for _, v := range arguments.GetOrPanic("Args").(*j.NativeArray).Array() {
		exprs = append(exprs, v.(Expr))
	}
	lastIsVariadic := false
	varg := arguments.GetOrPanic("Varg")
	if !j.IsUndefined(varg) {
		lastIsVariadic = varg.(j.NativeInt).Int() == 1
	}
	return &CallExpr{
		Func:    primaryExpr,    // Expr   function expression
		Args:    exprs,          // Exprs  function arguments, if any.
		Varg:    lastIsVariadic, // if true, final arg is variadic.
		NumArgs: len(exprs),     // len(Args) or len(Args[0].Results)
	}
}

// Returns &IndexExpr
func fPrimaryExprIndex(it j.Ast, ctx *j.ParseContext) j.Ast {
	m := it.(*j.NativeMap)
	primaryExpr := m.GetOrPanic("p").(Expr)
	index := m.GetOrPanic("i")
	return &IndexExpr{
		X:     primaryExpr,
		Index: index.(Expr),
		HasOK: false, // TODO if true, is form: `value, ok := <X>[<Key>]
	}
}

// This returns a &SliceExpr
// 2 cases are allowed by go/spec (square brackets denote optionality):
// - '[' [Expression] ':' [Expression]               ']'
// - '[' [Expression] ':'  Expression ':' Expression ']'
// It can panic with custom ParseError.
func fPrimaryExprSlice(it j.Ast, ctx *j.ParseContext) j.Ast {
	m := it.(*j.NativeMap)
	primaryExpr := m.GetOrPanic("p").(Expr)
	aSlice := m.GetOrPanic("s").(*j.NativeArray).Array()
	var low, high, max Expr
	if len(aSlice) < 2 || len(aSlice) > 3 {
		panic("assert") // impossible, by rule
	}
	if len(aSlice) == 3 {
		if j.IsUndefined(aSlice[2]) {
			panic(ctx.Error("3rd argument 'max' is mandatory in full slice expressions foo[low:high:max]"))
		}
		if j.IsUndefined(aSlice[1]) {
			panic(ctx.Error("2nd argument 'high' is mandatory in full slice expressions foo[low:high:max]"))
		}
		max = aSlice[2].(Expr)
	}
	if !j.IsUndefined(aSlice[0]) {
		low = aSlice[0].(Expr)
	}
	if !j.IsUndefined(aSlice[1]) {
		high = aSlice[1].(Expr)
	}
	return &SliceExpr{
		X:    primaryExpr,
		Low:  low,
		High: high,
		Max:  max,
	}
}

// This returns a &SelectorExpr.
// Selectors are expr like `x.f`,
// Where `x`, according to go/spec
// is a primary expression that MUST NOT be a package name
func fPrimaryExprSelector(it j.Ast, ctx *j.ParseContext) j.Ast {
	m := it.(*j.NativeMap)
	primaryExpr := m.GetOrPanic("p").(Expr)
	if v, is := primaryExpr.(*NameExpr); is {
		if v.HasAttribute("i_m_a_package_name") {
			panic(ctx.Error("selector operate on primary expression that is NOT a package name"))
		}
	}
	selector := m.GetOrPanic("s").(*NameExpr)
	return &SelectorExpr{
		X:   primaryExpr,
		Sel: selector.Name,
	}
}

// This returns a gnolang.Type (a constTypeExpr)
func fTypeName(it j.Ast, ctx *j.ParseContext) j.Ast {
	tname := it.(*j.NativeArray).Get(0).(*NameExpr)
	var pt PrimitiveType
	switch tname.Name {
	case "bool":
		pt = BoolType
	case "string":
		pt = StringType
	case "int":
		pt = IntType
	case "int8":
		pt = Int8Type
	case "int16":
		pt = Int16Type
	case "int32":
		pt = Int32Type
	case "int64":
		pt = Int64Type
	case "uint":
		pt = UintType
	case "uint8":
		pt = Uint8Type
	case "uint16":
		pt = Uint16Type
	case "uint32":
		pt = Uint32Type
	case "uint64":
		pt = Uint64Type
	case "float32":
		pt = Float32Type
	case "float64":
		pt = Float64Type
	default:
		// NativeType { Type reflect.Type /*go*/; typeid TypeID; gnoType Type // /*gno*/ }
		// DeclaredType { Name; Base Type; Methods[]TypedValue; typeid TypeID }
		// blockType {}
		// tupleType {}
		// RefType {}
		// MaybeNativeType { Type }
		// FuncType { Params []FieldType; Results []FieldType; typeid TypeID; bound *FuncType }
		//  func declareWith(pkgPath string, Name, b Type) // not for aliases
		//  func (ft *FuncType) Specify(store Store, argTVs []TypedValue, isVarg bool) *FuncType {
		// omitted are {Untyped*|DataByte|Bigint|Bigdec}Type
		panic(fmt.Sprintf("unsupported %q", tname.Name))
	}
	return &constTypeExpr{
		Source: tname,
		// Type is a gnolang.Type, an interface with TypeID().
		// See types.go
		Type: pt,
	}
}

// Returns &ArrayTypeExpr
func fArrayType(it j.Ast, ctx *j.ParseContext) j.Ast {
	a := it.(*j.NativeArray).Array()
	return &ArrayTypeExpr{
		Len: a[0].(Expr),
		Elt: a[1].(Expr),
	}
}

// Returns &MapTypeExpr
func fMapType(it j.Ast, ctx *j.ParseContext) j.Ast {
	a := it.(*j.NativeArray).Array()
	return &MapTypeExpr{
		Key:   a[0].(Expr),
		Value: a[1].(Expr),
	}
}

// Returns &ChanTypeExpr
func fChannelType(it j.Ast, ctx *j.ParseContext) j.Ast {
	m := it.(*j.NativeMap)
	var dir ChanDir
	switch strings.TrimSuffix(m.GetOrPanic("chanDir").(j.NativeString).String(), " ") {
	case "chan":
		dir = BOTH
	case "chan<-":
		dir = SEND
	case "<-chan":
		dir = RECV
	}
	return &ChanTypeExpr{
		Dir:   dir,
		Value: m.GetOrPanic("type").(Expr),
	}
}

// Returns &TypeAssertExpr
func fPrimaryExprTypeAssert(it j.Ast, ctx *j.ParseContext) j.Ast {
	// note: type args appear unsupported in X(), so we ignore typeargs in
	// o("typename:TypeName typeargs:TypeArgs?"),
	m := it.(*j.NativeMap)
	primaryExpr := m.GetOrPanic("primaryExpr").(Expr)
	typeAssertion := m.GetOrPanic("typeAssertion")
	return &TypeAssertExpr{
		X:     primaryExpr,
		Type:  typeAssertion.(Expr),
		HasOK: false, // TODO if true, is form: `_, ok := <X>(<Type>)
	}
}

func isUnderscore(c rune) bool { return c == '_' }

func fIdentifier(it j.Ast) j.Ast {
	return Nx(it.(*j.NativeArray).Concat())
}

// rune parser against unicode.IsLetter(rune),
// also allowing '_'. Rule being `letter = unicode_letter | '_' .`
func fLetter(_ j.Ast, ctx *j.ParseContext) j.Ast {
	// OPTIM NativeString being a buffer?
	if isLetter, rune := ctx.Code.MatchRune(unicode.IsLetter); isLetter {
		return j.NewNativeString(string(rune))
	} else if is, _ := ctx.Code.MatchRune(isUnderscore); is {
		return j.NewNativeString(string('_'))
	} else {
		return nil
	}
}

// rune parser against unicode.IsLetter(rune)
func funicode_letter(_ j.Ast, ctx *j.ParseContext) j.Ast {
	if is, rune := ctx.Code.MatchRune(unicode.IsLetter); is {
		return j.NewNativeString(string(rune))
	}
	return nil
}

// rune parser against unicode.IsDigit(rune)
func funicode_digit(_ j.Ast, ctx *j.ParseContext) j.Ast {
	if is, rune := ctx.Code.MatchRune(unicode.IsDigit); is {
		return j.NewNativeString(string(rune))
	}
	return nil
}

// return &FuncTypeExpr
// input: NativeMap with keys:
//   - "params": NativeArray<FieldTypeExpr>
//   - "results": NativeArray<FieldTypeExpr> or NativeUndefined
//
// The validity of variadic arguments is checked here
func fSignature(it j.Ast, ctx *j.ParseContext) j.Ast {
	m := it.(*j.NativeMap)
	// for params, only allow variadic for the last argument
	params := []FieldTypeExpr{}
	a := m.GetOrPanic("params").(*j.NativeArray).Array()
	for i, p := range a {
		// variadic args are represented as SliceTypeExpr with true Vrd
		if slice, is := p.(*FieldTypeExpr).Type.(*SliceTypeExpr); is && i < len(a)-1 && slice.Vrd {
			panic(ctx.Error(fmt.Sprintf("only the final parameter can be variadic. %d", i)))
		}
		params = append(params, *p.(*FieldTypeExpr))
	}
	// don't allow variadic in results
	results := []FieldTypeExpr{}
	if b, exists := m.GetExists("result"); exists && !j.IsUndefined(b) {
		for _, p := range b.(*j.NativeArray).Array() {
			if slice, is := p.(*FieldTypeExpr).Type.(*SliceTypeExpr); is && slice.Vrd {
				panic(ctx.Error("function results can not be variadic"))
			}
			results = append(results, *p.(*FieldTypeExpr))
		}
	}
	return &FuncTypeExpr{
		Params:  params,
		Results: results,
	}
}

// returns NativeArray<FieldTypeExpr>.
// Main task therefore is to linearize as in [[a,b], [c]] -> [a,b,c].
//
// Input:
// `(a, b int, s string, d ...float)` would develop as
// [[const-type int, const-type int], [const-type string], [... const-type float]].
func fParameters(it j.Ast, ctx *j.ParseContext) j.Ast {
	ret := j.NewEmptyNativeArray()
	groups := it.(*j.NativeArray).Array()
	if len(groups) == 0 {
		return j.NewNativeArray([]j.Ast{})
	}
	if groups[0].(*j.NativeArray).Length() == 0 {
		panic("assert")
	}
	// check Namedness: "Within a list of parameters or results,
	// the names must either all be present or all be absent",
	// i.e. (first.Name != "") == (el.Name != "") ∀ el
	firstIsSet := false
	firstIsNamed := false
	for _, group := range groups {
		for _, t := range group.(*j.NativeArray).Array() {
			el := t.(*FieldTypeExpr)
			if !firstIsSet {
				firstIsNamed = el.Name != ""
				firstIsSet = true
			} else {
				if firstIsNamed != (el.Name != "") {
					panic(ctx.Error(
						// note: could have used nodes.go (ftxz FieldTypeExprs)
						// IsNamed(), which panics when inconsistent namedness.
						// panic with a ParseError and custom message instead.
						fmt.Sprintf(
							"within a list of parameters the names must either"+
								" all be present or all be absent: %v", it.String(),
						)))
				}
			}
		}
	}
	// "all non-blank names in the signature must be unique."
	names := map[Name]bool{} // bool doesn't matter, it's a "set"
	for _, group := range groups {
		for _, t := range group.(*j.NativeArray).Array() {
			el := t.(*FieldTypeExpr)
			if el.Name != "" && el.Name != "_" {
				if _, already := names[el.Name]; already {
					panic(ctx.Error(
						fmt.Sprintf(
							"all non-blank names in the signature must be unique: %v",
							it.String(),
						)))
				} else {
					names[el.Name] = true
					ret.Append(el)
				}
			}
		}
	}
	// variadic will be checked in fSignature,
	// as in this function, we still lack the knowledge
	// of whether the array of args is for `params` or `results`.
	return ret
}

// Returns NativeArray<*FieldTypeExpr>
func fParameterDecl(it j.Ast, ctx *j.ParseContext) j.Ast {
	a := it.(*j.NativeArray).Array()
	r := []j.Ast{}
	for _, identifier := range a[0].(*j.NativeArray).Array() {
		// we substitute variadic here by SliceTypeExpr with Vrd set to true
		// namedness is checked in fParameters
		// variadicity validity is checked in fSignature
		var fte *FieldTypeExpr
		isVariadic := a[1].(j.NativeInt).Bool()
		if !isVariadic {
			fte = &FieldTypeExpr{
				Name: identifier.(*NameExpr).Name,
				Type: a[2].(Expr),
			}
		} else {
			fte = &FieldTypeExpr{
				Name: identifier.(*NameExpr).Name,
				Type: &SliceTypeExpr{
					Elt: a[2].(Expr),
					Vrd: true,
				},
			}
		}
		r = append(r, fte)
	}
	return j.NewNativeArray(r)
}

// This returns a NativeArray<*FieldTypeExpr>
// and expects either (because of `Parameters | Type`):
// - a NativeArray<FieldTypeExpr> (when `Parameters`)
// - a TypeExpr such as *FieldTypeExpr (when `Type`)
func fResult(it j.Ast, ctx *j.ParseContext) j.Ast {
	r := []j.Ast{}
	switch v := it.(type) {
	case TypeExpr:
		// a single type
		r = append(r, &FieldTypeExpr{
			Name: "",
			Type: v,
		})
	case *j.NativeArray:
		for _, field := range v.Array() {
			r = append(r, field.(*FieldTypeExpr))
		}
	default:
		panic(fmt.Sprintf("unhandled %s", v.String()))
	}
	// - namedness is checked in fParameters
	// - variadicity validity is checked in fSignature
	return j.NewNativeArray(r)
}

// Returns &FieldTypeExpr
// "A field declared with a type but no explicit field
// name is called an embedded field."
// (i.e. the `Name` field is called "embedded" in the following snippet:
// `type Foo struct { Name; age int }`)
func fEmbeddedField(it j.Ast, ctx *j.ParseContext) j.Ast {
	m := it.(*j.NativeMap)
	isStar := m.GetOrPanic("star").(j.NativeInt).Bool()
	typeName := m.GetOrPanic("typename").(Expr)
	// TODO when we support TypeArgs
	// typeArgs := m.GetOrPanic("typeargs").(Expr)
	if isStar {
		typeName = &StarExpr{X: typeName}
	}
	// "An embedded field must be specified as a type name T or as a pointer to
	// a non-interface type name *T, and T itself may not be a pointer type."
	return &FieldTypeExpr{
		Name: Name(""),
		Type: typeName.(Expr),
		Tag:  nil, // Tag overwritten in fFieldDecl
	}
}

// This returns a NativeArray<*FieldTypeExpr>
// The rule calling this is the first of those:
// : o(`IdentifierList _ Type Tag?`, fFieldDecl1),
// : o(`EmbeddedField Tag?`, fFieldDecl2),
//
// In other words, this is a list of fields
// that all have the same Type, as in `a, b, c int`.
// To which a Tag (that can be nil) is added.
func fFieldDecl1(it j.Ast, ctx *j.ParseContext) j.Ast {
	a := it.(*j.NativeArray).Array()
	ret := j.NewEmptyNativeArray()
	identifiers := a[0].(*j.NativeArray).Array()
	theType := a[1].(TypeExpr)
	var tag *BasicLitExpr = nil
	if !j.IsUndefined(a[2]) {
		tag = &BasicLitExpr{
			Kind:  STRING,
			Value: a[2].(j.NativeString).String(),
		}
	}
	for _, identifier := range identifiers {
		// identifier begin by a letter, by rule.
		ret.Append(&FieldTypeExpr{
			Name: identifier.(*NameExpr).Name,
			Type: theType.(TypeExpr),
			Tag:  tag,
		})
	}
	return ret
}

// This returns a NativeArray<*FieldTypeExpr>
// The rule calling this is the second of those:
// : o(`IdentifierList _ Type Tag?`, fFieldDecl1),
// : o(`EmbeddedField Tag?`, fFieldDecl2),
//
// In other words, the input is a *FieldTypeExpr
// (prepared from fEmbeddedField) to which optional Tag is added.
func fFieldDecl2(it j.Ast, ctx *j.ParseContext) j.Ast {
	a := it.(*j.NativeArray).Array()
	ret := j.NewEmptyNativeArray()
	fte := a[0].(*FieldTypeExpr)
	var tag *BasicLitExpr = nil
	if !j.IsUndefined(a[1]) {
		tag = &BasicLitExpr{
			Kind:  STRING,
			Value: a[1].(j.NativeString).String(),
		}
	}
	// `fte` was almost fully prepared by fEmbeddedField.
	// Just add the Tag now, or nil.
	fte.Tag = tag
	ret.Append(fte)
	return ret
}

// This returns a &StructTypeExpr
// from a NativeArray<NativeArray<*FieldTypeExpr>>
// There are two levels, because `a,b,c int` returns 3 fields.
// This is where it get linearized
//
// TODO struct require some work to check conditions
// defined in https://go.dev/ref/spec#StructType (e.g. uniqueness of field
// names)
func fStructType(it j.Ast) j.Ast {
	fields := FieldTypeExprs{}
	for _, aa := range it.(*j.NativeArray).Array() {
		for _, field := range aa.(*j.NativeArray).Array() {
			fields = append(fields, *field.(*FieldTypeExpr))
		}
	}
	return &StructTypeExpr{Fields: fields}
}
