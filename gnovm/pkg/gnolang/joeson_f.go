package gnolang

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	j "github.com/grepsuzette/joeson"
)

func fExpression(it j.Ast) j.Ast {
	// bx:(Expression _ binary_op _ Expression) | UnaryExpr
	//                                             ^-- done in mtUnaryExpr
	if m, ok := it.(j.NativeMap); ok {
		// bx: create a BinaryExpr with Bx
		a := m.GetOrPanic("bx").(*j.NativeArray).Array
		return &BinaryExpr{
			Left:  a[0].(Expr),
			Op:    Op2Word(a[1].(j.NativeString).Str),
			Right: a[2].(Expr),
		}
	} else {
		return it
	}
}

func fUnary(it j.Ast) j.Ast {
	if m, ok := it.(j.NativeMap); ok {
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
		s := ""
		switch v := it.(type) {
		case *j.NativeArray:
			s = v.Concat()
		case j.NativeString:
			s = v.Str
		default:
			panic(fmt.Sprintf("Unexpected type in fInt %s", reflect.TypeOf(it).String()))
		}
		var e error
		var i int64
		var prefix string
		switch base {
		case 2:
			// i, e = strconv.ParseInt(s[2:], 2, 64)
			i, e = strconv.ParseInt(s, 2, 64)
			prefix = "0b"
		case 8:
			// i, e = strconv.ParseInt(s, 8, 64)
			if strings.HasPrefix(s, "0o") || strings.HasPrefix(s, "0O") {
				i, e = strconv.ParseInt(s[2:], 8, 64)
			} else if strings.HasPrefix(s, "0") {
				i, e = strconv.ParseInt(s[1:], 8, 64) // 0177
			}
			prefix = "0o"
		case 10:
			// i, e = strconv.ParseInt(s, 10, 64)
			i, e = strconv.ParseInt(s, 10, 64)
			prefix = ""
		case 16:
			// i, e = strconv.ParseInt(s[2:], 16, 64)
			i, e = strconv.ParseInt(s, 16, 64)
			prefix = "0x"
		default:
			panic("impossible base, expecting 2,8,10,16")
		}
		if e != nil {
			// it may have overflowed
			// or faulty grammar.
			fmt.Println(e.Error())
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
