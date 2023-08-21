package gnolang

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	j "github.com/grepsuzette/joeson"
)

// parse functions
// Naming convention:
// - fXxx(it Ast, *ParseContext) Ast
// - ffXxx(someArg) func(it Ast, *ParseContext) Ast

func stringIt(it j.Ast) (string, error) {
	switch v := it.(type) {
	case *j.NativeArray:
		return v.Concat(), nil
	case *j.NativeMap:
		return v.Concat(), nil
	case j.NativeString:
		return string(v), nil
	default:
		return "", errors.New(fmt.Sprintf("Unexpected type in stringIt: %s", reflect.TypeOf(it).String()))
	}
}

// peel([[a,b,...]]) -> [a,b,...]
// Asserts `it` is NativeArray
// Useful when rules would create two or more levels of NativeArray.
func peel(it j.Ast, ctx *j.ParseContext) j.Ast {
	// (don't assert size of exactly 1, we want
	// cases like [[a,b,..], NativeUndefined{}]) to also work)
	return it.(*j.NativeArray).Get(0)
}

func fExpression(it j.Ast, ctx *j.ParseContext, org j.Ast) j.Ast {
	// bx:(Expression binary_op Expression) | ux:UnaryExpr
	if m, ok := it.(*j.NativeMap); ok {
		if m.Exists("ux") {
			return m.GetOrPanic("ux")
		} else if m.Exists("bx") {
			// bx: create a BinaryExpr with Bx
			a := m.GetOrPanic("bx").(*j.NativeArray).Array()
			return &BinaryExpr{
				Left:  a[0].(Expr),
				Op:    Op2Word(a[1].(j.NativeString).String()),
				Right: a[2].(Expr),
			}
		} else {
			panic("assert")
		}
	} else {
		return it
	}
}

func fUnaryExpr(it j.Ast) j.Ast {
	// PrimaryExpr | ux:(unary_op _ UnaryExpr)
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

func ffInt(base int) func(j.Ast, *j.ParseContext) j.Ast {
	return func(it j.Ast, ctx *j.ParseContext) j.Ast {
		var e error
		var s string
		s, e = stringIt(it)
		if e != nil {
			return ctx.Error(e.Error())
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

// Loosely based on "the imagination song", by South Park
func fImaginary(it j.Ast, ctx *j.ParseContext) j.Ast {
	a := it.(*j.NativeArray)
	s := a.Concat()
	if s[len(s)-1:] != "i" {
		panic("assert") // imaginary_lit ends with 'i' by rule
	}
	return &BasicLitExpr{
		Kind:  IMAG,
		Value: s,
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

func finterpreted_string_lit(it j.Ast, ctx *j.ParseContext) j.Ast {
	if j.IsParseError(it) {
		return it
	}
	if s, e := stringIt(it); e == nil {
		return &BasicLitExpr{
			Kind:  STRING,
			Value: `"` + s + `"`,
		}
	} else {
		return ctx.Error(e.Error())
	}
}

func fraw_string_lit(it j.Ast, ctx *j.ParseContext) j.Ast {
	if j.IsParseError(it) {
		return it
	}
	if s, e := stringIt(it); e == nil {
		return &BasicLitExpr{
			Kind:  STRING,
			Value: "`" + s + "`",
		}
	} else {
		return ctx.Error(e.Error())
	}
}

func ffBasicLit(kind Word) func(j.Ast, *j.ParseContext) j.Ast {
	return func(it j.Ast, ctx *j.ParseContext) j.Ast {
		if j.IsParseError(it) {
			return it
		}
		if s, e := stringIt(it); e == nil {
			return &BasicLitExpr{
				Kind:  kind,
				Value: s,
			}
		} else {
			return ctx.Error(e.Error())
		}
	}
}

func ffErr(msg string) func(j.Ast, *j.ParseContext) j.Ast {
	return func(it j.Ast, ctx *j.ParseContext) j.Ast {
		return ctx.Error(msg)
	}
}

// same as ffErr but with postpended colon and near context
func ffErrNearContext(msg string) func(j.Ast, *j.ParseContext) j.Ast {
	return func(it j.Ast, ctx *j.ParseContext) j.Ast {
		return ctx.Error(msg + ": " + ctx.Code.PeekLines(-1, 1))
	}
}

// Panic with a ParseError made from msg string.
// ParseErrors panics are recovered higher up, in parseX()
func ffPanic(msg string) func(j.Ast, *j.ParseContext) j.Ast {
	return func(it j.Ast, ctx *j.ParseContext) j.Ast {
		panic(ctx.Error(msg))
	}
}

// Like ffPanic, but the text of the current line from the parse context is
// postpended to `msg`. This panics with a ParseError that should be
// recovered in parseX()
func ffPanicNearContext(msg string) func(j.Ast, *j.ParseContext) j.Ast {
	return func(it j.Ast, ctx *j.ParseContext) j.Ast {
		panic(ctx.Error(msg + ": " + ctx.Code.PeekLines(-1, 1)))
	}
}

func fSimpleStmt(it j.Ast) j.Ast {
	// TODO "The following built-in functions are not permitted in statement context:
	// append cap complex imag len make new real
	// unsafe.Add unsafe.Alignof unsafe.Offsetof unsafe.Sizeof unsafe.Slice"
	return it
}

// same as identifier (*NameExpr), but when Name is the blank identifier panic with a ParseError
func fPackageName(it j.Ast, ctx *j.ParseContext) j.Ast {
	if it.(*NameExpr).String() == "_" {
		panic(ctx.Error("PackageName must not be the blank identifier"))
	} else {
		// experiment, trying to differentiate `identifier` from package name
		// here, as they are both NameExpr. Until a better idea.
		// See fPrimaryExprSelector()
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

// returns a &KeyValueExpr, with possibly nil Key when just a Value is available.
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

// returns a NativeMap,
// It resembles a CallExpr without Func:
// "Args"    Exprs        function arguments, if any.
// "Varg"	 NativeInt    if 1, final arg is variadic.
// "NumArgs" NativeInt    len(Args) or len(Args[0].Results)
func fArguments(it j.Ast, ctx *j.ParseContext) j.Ast {
	switch m := it.(type) {
	case j.NativeUndefined:
		return j.NewNativeMap(map[string]j.Ast{
			"Args":    j.NewNativeArray([]j.Ast{}),
			"NumArgs": j.NewNativeInt(0),
			"Varg":    j.NewNativeInt(0),
		})
	case *j.NativeMap:
		args := m.GetOrPanic("Args").(*j.NativeArray)
		m.Set("NumArgs", j.NewNativeInt(args.Length()))
		return m // this will be used in e.g. fPrimaryExprArguments()
	default:
		panic("assert")
	}
}

// This returns a &CallExpr
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

// This returns a &IndexExpr
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

// This returns a &SliceExpr if it's valid
// 2 cases are allowed by go/spec (square brackets denote optionality):
// - '[' [Expression] ':' [Expression]               ']'
// - '[' [Expression] ':'  Expression ':' Expression ']'
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

// This returns a &SelectorExpr
// this builds selector for expr like `x.f` where x
// is a primary expression that is not a package name
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
		// Type is a gnolang.Type, an interface which
		// boasts a TypeID(). See the top of types.go
		Type: pt,
	}
}

// This returns an &ArrayTypeExpr
func fArrayType(it j.Ast, ctx *j.ParseContext) j.Ast {
	a := it.(*j.NativeArray).Array()
	return &ArrayTypeExpr{
		Len: a[0].(Expr),
		Elt: a[1].(Expr),
	}
}

// This returns &MapTypeExpr
func fMapType(it j.Ast, ctx *j.ParseContext) j.Ast {
	a := it.(*j.NativeArray).Array()
	return &MapTypeExpr{
		Key:   a[0].(Expr),
		Value: a[1].(Expr),
	}
}

// This returns &ChanTypeExpr
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

// This returns a &TypeAssertExpr
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

// rune parser against unicode.IsLetter(rune), also '_'
// letter = unicode_letter | '_' .
func fLetter(_ j.Ast, ctx *j.ParseContext) j.Ast {
	// OPTIM NativeString probably not good idea anymore
	if isLetter, rune := ctx.Code.MatchRune(unicode.IsLetter); isLetter {
		return j.NewNativeString(string(rune))
	} else if is, _ := ctx.Code.MatchRune(isUnderscore); is {
		return j.NewNativeString(string('_'))
	} else {
		return nil
	}
}

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
// `(a, b int, s string, d ...float)` would be
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

// This returns a NativeArray<*FieldTypeExpr>
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

// This returns a &FieldTypeExpr
// "A field declared with a type but no explicit field
// name is called an embedded field." (e.g. Attr, Name)
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
// from a list of fields sharing the same type to which we add a Tag.
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
// from a *FieldTypeExpr to which we merely add a Tag.
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
	// almost ready from fEmbeddedField was still waiting tag.
	fte.Tag = tag
	ret.Append(fte)
	return ret
}

// This returns a &StructTypeExpr
// from a NativeArray<NativeArray<*FieldTypeExpr>>
// yes there are two levels, because `a,b,c int` returns 3 fields.
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
