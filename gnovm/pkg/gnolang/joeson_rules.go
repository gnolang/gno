package gnolang

import (
	"strings"

	j "github.com/grepsuzette/joeson"
)

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

// rules are named after https://go.dev/ref/spec
// labels (such as "bx:") are used by rules callbacks.
var (
	grammar  *j.Grammar
	gnoRules = rules(
		o(named("Input", "Expression")),
		o(named("Expression", "bx:(Expression _ binary_op _ Expression) | UnaryExpr"), fExpression),
		o(named("UnaryExpr", "PrimaryExpr | ux:(unary_op _ UnaryExpr)"), fUnaryExpr),
		o(named("unary_op", revQuote("+ - ! ^ * & <-"))),
		o(named("binary_op", "mul_op | add_op | rel_op | '&&' | '||'")),
		o(named("mul_op", revQuote("* / % << >> & &^"))),
		o(named("add_op", revQuote("+ - | ^"))),
		o(named("rel_op", revQuote("== != < <= > >="))),
		// o(named("PrimaryExpr", "Operand | Conversion | MethodExpr | PrimaryExpr _ ( Selector | Index | Slice | TypeAssertion | Arguments )")),
		o(named("PrimaryExpr", "Operand")),

		o(named("Operand", rules(
			// o("'(' _ Expression _ ')' | OperandName TypeArgs? | Literal"), // TODO this is the original
			o("lit:Literal | '(' _ expr:Expression _ ')'", func(it j.Ast) j.Ast { return it.(j.NativeMap).GetWhicheverOrPanic([]string{"lit", "expr"}) }),
			// o(named("Literal", "BasicLit | CompositeLit | FunctionLit")),
			o(named("Literal", "BasicLit")),
			// TODO add float_lit and imaginary_lit
			// o(named("BasicLit", "int_lit | rune_lit | string_lit")),
			o(named("BasicLit", "float_lit | int_lit")),
			o(named("int_lit", "hex_lit | octal_lit | binary_lit | decimal_lit"), fInt),
			i(named("decimal_lit", "/^0|[1-9][_0-9]*[0-9]?/")), // x("decimal_lit")),
			i(named("binary_lit", "/^0[bB]_?([01_])*[01]/")),
			i(named("octal_lit", "/^0[oO]_?([01234567_])*[01234567]/")),
			i(named("hex_lit", "/^0[xX]_?([0123456789a-fA-F_])*[0123456789a-fA-F]/")),
			// float_lit
			o(named("float_lit", "decimal_float_lit"), fFloat), // TODO | hex_float_lit")),
			o(named("decimal_float_lit", "decimal_digits '.' decimal_digits? decimal_exponent?"+
				"| decimal_digits decimal_exponent"+
				"| '.' decimal_digits decimal_exponent?",
			)),
			o(named("decimal_exponent", "/[eE][+-]?/ decimal_digits")),
			// o(named("OperandName", "QualifiedIdent | identifier")),
			// i(named("QualifiedIdent", "PackageName '.' identifier"), x("QualifiedIdent")), // https://go.dev/ref/spec#QualifiedIdent
			// i(named("PackageName", "identifier")),                                         // https://go.dev/ref/spec#PackageName
			// o(named("Block", "'{' Statement*';' '}'")),

			/*
				o(named("FunctionLit", "'func' _ Signature _ FunctionBody")),
				o(named("Signature", "Parameters _ Result?")),
				o(named("Result", "Parameters | Type")),
				o(named("Parameters", "'(' _ ParameterList*comma _ ')'")),
				o(named("ParameterList", "ParameterDecl _ (',' _ ParameterDecl)*")),
				o(named("ParameterDecl", "IdentifierList? _ '...'? _ Type")),

				o(named("IdentifierList", "identifier _ comma*identifier")),
				o(named("ExpressionList", "Expression _ comma*Expression")),
				// o(named("identifier", "letter (letter | unicode_digit)*")),
				i(named("identifier", "[a-zA-Z_][a-zA-Z0-9_]*")), //, x("identifier")), // letter { letter | unicode_digit } . FIXME We rewrite it for now to accelerate parsing
			*/

			/*
				// spec/Type.txt
				o("TypeName TypeArgs? | TypeLit | '(' Type ')'"),
				o(named("TypeLit", rules(
					// "The length is part of the array's type; it must evaluate to
					// a non-negative constant representable by a value of type int.
					// The length of array a can be discovered using the built-in
					// function len. The elements can be addressed by integer indices
					// 0 through len(a)-1. Array types are always one-dimensional but
					// may be composed to form multi-dimensional types."
					o(named("ArrayType", "'[' length:Expression ']' elementType:Type")), //, x("ArrayType")),
				// o("StructType"),
				// o("PointerType"),
				// o("FunctionType"),
				// o("InterfaceType"),
				// o("SliceType"),
				// o("MapType"),
				// o("ChannelType"),
				))),
				i(named("TypeName", "QualifiedIdent | identifier")),
				i(named("TypeArgs", "'[' TypeList ','? ']'")),
				i(named("TypeList", "Type*','")),
			*/
		))),
		// "White space, formed from spaces (U+0020), horizontal tabs (U+0009),
		// carriage returns (U+000D), and newlines (U+000A), is ignored except as
		// it separates tokens that would otherwise combine into a single token."
		i(named("comma", "',' | _")),
		i(named("_", "/[ \t\n\r]*/")),
		i(named("__", "/[ \t\n\r]+/")),
		i(named("decimal_digits", "/[0-9](_?[0-9])*/")),
	)
)

// helps writing rules for PEG grammars.
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
