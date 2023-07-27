package gnolang

import (
	"strings"

	j "github.com/grepsuzette/joeson"
)

// rule names prefixed by a _ (e.g. _BinaryExpr)
// do not appear in spec/spec.html, and have been thus distinguished for
// clarity.

// rules are named after https://go.dev/ref/spec
// labels (such as "bx:") are used by rules callbacks.
var (
	gm       *j.Grammar
	gnoRules = rules(
		o(named("Input", "Expression")),
		o(named("Expression", "bx:(Expression _ binary_op _ Expression) | UnaryExpr"), fExpression),
		o(named("UnaryExpr", "PrimaryExpr | ux:(unary_op _ UnaryExpr)"), fUnary),
		o(named("unary_op", revQuote("+ - ! ^ * & <-"))),
		o(named("binary_op", "mul_op | add_op | rel_op | '&&' | '||'")),
		o(named("mul_op", revQuote("* / % << >> & &^"))),
		o(named("add_op", revQuote("+ - | ^"))),
		o(named("rel_op", revQuote("== != < <= > >="))),
		// o(named("PrimaryExpr", "Operand | Conversion | MethodExpr | PrimaryExpr _ ( Selector | Index | Slice | TypeAssertion | Arguments )")),
		o(named("PrimaryExpr", "Operand")),

		o(named("Operand", rules(
			// o("'(' _ Expression _ ')' | OperandName TypeArgs? | Literal"), // TODO this is the original
			o("Literal | '(' _ Expression _ ')'"), // 𝘧, func(it j.Ast) j.Ast { return it.(j.NativeMap).GetWhicheverOrPanic([]string{"lit", "expr"}) }),
			// o(named("Literal", "BasicLit | CompositeLit | FunctionLit")),
			i(named("Literal", "BasicLit")),
			i(named("BasicLit", rules(
				o("rune_lit | imaginary_lit | float_lit | int_lit"), //| string_lit")),
				i(named("rune_lit", rules(
					o("'\\'' ( byte_value | unicode_value ) '\\'' | '\\'' [^\n] '\\''"),
					i(named("byte_value", rules(
						o("octal_byte_value | hex_byte_value"),
						i(named("octal_byte_value", `a:'\\' b:octal_digit{3,3}`)),
						i(named("hex_byte_value", `a:'\\x' b:hex_digit{2,2}`)),
					))),
					i(named("unicode_value", rules(
						o("escaped_char | little_u_value | big_u_value | unicode_char | _error_unicode_char_toomany"),
						i(named("escaped_char", `[\a\f\n\r\t\v]`)),                                          // TODO NOTE: we skipped \b as for some reason it doesn't work in the regex.
						i(named("little_u_value", `a:'\\u' b:hex_digit*`), ff_u_value("little_u_value", 4)), // 4 hex_digit or error
						i(named("big_u_value", `a:'\\U' b:hex_digit*`), ff_u_value("big_u_value", 8)),       // 8 hex digit or error
						i(named("_error_unicode_char_toomany", "[^\\x{0a}]{2,}"), func(it j.Ast, ctx *j.ParseContext) j.Ast { return ctx.Error("too many characters") }),
					))),
				)), f_rune_lit),
				i(named("int_lit", rules(
					// the order is critical for PEG grammars
					o(named("binary_lit", "('0b'|'0B') '_'? binary_digits"), ffInt(2)),
					o(named("hex_lit", "('0x'|'0X') '_'? hex_digits"), ffInt(16)),
					o(named("octal_lit", "[0] [oO]? '_'? octal_digit octal_digits?"), ffInt(8)),
					o(named("decimal_lit", "[0] | [1-9] ( '_'? decimal_digits)?"), ffInt(10)),
				))),
				i(named("float_lit", rules(
					o("decimal_float_lit | hex_float_lit"),
					i(named("decimal_float_lit",
						"DOT decimal_digits decimal_exponent? | "+
							"decimal_digits DOT decimal_digits? decimal_exponent? | "+
							"decimal_digits decimal_exponent"), ffFloatFormat("%g")),
					i(named("hex_float_lit", "[0] [xX] hex_mantissa hex_exponent"), ffFloatFormat("%x")),
					i(named("decimal_exponent", "[eE] [+-]? decimal_digits")),
					i(named("hex_mantissa", "'_'? hex_digits DOT hex_digits? |"+
						"'_'? hex_digits | DOT hex_digits",
					)),
					i(named("hex_exponent", "[pP] [+-]? decimal_digits")),
				))),
				i(named("imaginary_lit", "(float_lit | int_lit | decimal_digits ) [i]"), fImaginary),
				// avoid regexes with PEG in general, regexes are greedy and this can
				// create ambiguity and buggy grammars. As a special case, character classes are OK.
				// Regexes can be used to optimize but again avoid them unless
				// you know what you're doing.
				i(named("decimal_digits", "decimal_digit ( '_'? decimal_digit )*")),
				i(named("binary_digits", "binary_digit ( '_'? binary_digit )*")),
				i(named("octal_digits", "octal_digit ( '_'? octal_digit )*")),
				i(named("hex_digits", "hex_digit ( '_'? hex_digit )*")),
				i(named("decimal_digit", "[0-9]")),
				i(named("binary_digit", "[01]")),
				i(named("octal_digit", "[0-7]")),
				i(named("hex_digit", "[0-9a-fA-F]")),
			))),
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
		i(named("characters", "(newline | unicode_char | unicode_letter | unicode_digit)")),
		i(named("newline", "[\\x{0a}]")),
		i(named("unicode_char", "[^\\x{0a}]")), // "an arbitrary Unicode code point except newline"
		i(named("unicode_letter", "[a-zA-Z]")), // "a Unicode code point categorized as "Letter" TODO it misses all non ASCII
		i(named("unicode_digit", "[0-9]")),     // "a Unicode code point categorized as "Number, decimal digit" TODO it misses all non ASCII
		//                         ^^^
		// For now we'll stick to ANSI for letters and digits. It can later be improved
		// We looked into unicode specs for them but there are not defined
		// in https://www.unicode.org/versions/Unicode8.0.0/ch04.pdf Section 4.5

		// "White space, formed from spaces (U+0020), horizontal tabs (U+0009),
		// carriage returns (U+000D), and newlines (U+000A), is ignored except as
		// it separates tokens that would otherwise combine into a single token."
		i(named("comma", "',' | _")),
		i(named("DOT", "'.'")), // when it needs to get captured (by default '.' in a sequence is not captured)
		i(named("_", "[ \t\n\r]*")),
		i(named("__", "[ \t\n\r]+")),
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

// vim: fdm=indent
