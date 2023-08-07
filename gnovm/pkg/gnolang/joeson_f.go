package gnolang

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	j "github.com/grepsuzette/joeson"
)

func stringIt(it j.Ast) (string, error) {
	switch v := it.(type) {
	case *j.NativeArray:
		return v.Concat(), nil
	case *j.NativeMap:
		return v.Concat(), nil
	case j.NativeString:
		return v.Str, nil
	default:
		return "", errors.New(fmt.Sprintf("Unexpected type in stringIt: %s", reflect.TypeOf(it).String()))
	}
}

func fExpression(it j.Ast, ctx *j.ParseContext, org j.Ast) j.Ast {
	// bx:(Expression _ binary_op _ Expression) _T? | UnaryExpr _T?
	//                                                  ^-- done in mtUnaryExpr
	if m, ok := it.(*j.NativeMap); ok {
		if m.Exists("ux") {
			return m.GetOrPanic("ux")
		} else if m.Exists("bx") {
			// bx: create a BinaryExpr with Bx
			a := m.GetOrPanic("bx").(*j.NativeArray).Array
			return &BinaryExpr{
				Left:  a[0].(Expr),
				Op:    Op2Word(a[1].(j.NativeString).Str),
				Right: a[2].(Expr),
			}
		} else {
			panic("assert")
		}
	} else {
		// panic(reflect.TypeOf(it).String())
		// panic(it.String())
		return it
	}
}

func fUnary(it j.Ast) j.Ast {
	if m, ok := it.(*j.NativeMap); ok {
		// ux:(unary_op _ UnaryExpr)
		a := m.GetOrPanic("ux").(*j.NativeArray).Array
		return &UnaryExpr{
			Op: Op2Word(a[0].(j.NativeString).Str),
			X:  a[1].(Expr),
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
		panic("assert: imaginary_lit ends with 'i'")
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
		h := it.(*j.NativeMap)
		if h.GetOrPanic("b").(*j.NativeArray).Length() != hexDigits {
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

// when PackageName is the blank identifier panic with a ParseError
func fPackageName(it j.Ast, ctx *j.ParseContext) j.Ast {
	if it.String() == "" {
		panic(ctx.Error("PackageName must not be the blank identifier"))
	}
	return Nx(it.(*j.NativeArray).Concat())
}

func fQualifiedIdent(it j.Ast, ctx *j.ParseContext) j.Ast {
	m := it.(*j.NativeMap)
	packageName := m.GetOrPanic("p")
	identifier := m.GetOrPanic("i")
	return &SelectorExpr{
		X:   packageName.(*NameExpr),
		Sel: N(identifier.(*j.NativeArray).Concat()),
	}
}
