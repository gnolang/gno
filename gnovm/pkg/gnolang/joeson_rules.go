package gnolang

import (
	"strings"

	j "github.com/grepsuzette/joeson"
)

// A Golang PEG grammar.
//   - This grammar is written against the *formal go syntax* with semicolons ";"
//     as terminators, as is produced for example by "go/scanner".
//   - The following rules are named after https://go.dev/ref/spec
//
// (?...) are lookaheads, they can optimize away certain branches but consume nothing
// Unsourced quotes come from gospec.html.
var (
	gm       *j.Grammar
	gnoRules = rules(
		o(named("Input", `SimpleStmt _SEMICOLON?`)),
		o(named("Block", rules(
			o(`'{' Statement*_SEMICOLON '}'`),
			i(named("Statement", `SimpleStmt`)),
			i(named("SimpleStmt", `ExpressionStmt`), fSimpleStmt),
			i(named("ExpressionStmt", `Expression`)),
			i(named("Expression", rules(
				// (see study in joeson/examples/precedence)
				// "Expression" is normally `Expression binary_op Expression | UnaryExpr`
				// because right recursion would affect precedence
				// as explained by Laurence Tratt in http://tratt.net/laurie/research/publications/html/tratt__direct_left_recursive_parsing_expression_grammars/
				// we simply express like this instead: `Expression binary_op UnaryExpr`
				// In practice, we must also deal with precedence,
				// each precedence level is a moss familie growing laterally,
				// the moss families don't intermix.
				// So it ends up being like this:
				o(named("BinaryExpr", rules(
					// o(`bxTerm (add_op bxTerm)*`, growMoss),

					o(`bxLOr (opLOR bxLOr)*`, growMoss),

					i(named("bxLOr", `bxLAnd (opLAND bxLAnd)*`), growMoss),
					i(named("bxLAnd", `bxRel (rel_op bxRel)*`), growMoss),
					i(named("bxRel", `bxTerm (add_op bxTerm)*`), growMoss),

					i(named("bxTerm", `bxFactor (mul_op bxFactor)*`), growMoss),
					i(named("bxFactor", `'(' Expression ')' | UnaryExpr`)),
					i(named("opLOR", `'||'`)),
					i(named("opLAND", `'&&'`)),
				))),
				o(named("UnaryExpr", `PrimaryExpr | ux:(unary_op UnaryExpr)`), fUnaryExpr),
				i(named("unary_op", `'+' | '-' | '!' | '^' | '*' | ([&] !'&') | '<-'`)),
				i(named("binary_op", rules(
					o(`'||' | '&&' | rel_op | add_op | mul_op`),
					i(named("mul_op", `'*' | '/' | '%' | '<<' | '>>' | '&^' | ([&] !'&')`)),
					i(named("add_op", `'+' | '-' | ([|] !'|') | '^'`)),
					i(named("rel_op", `op:('==' | '!=' | '<=' | '>=' | '<' | '>') _:_`), func(it j.Ast) j.Ast { return it.(*j.NativeMap).GetOrPanic("op") }),
				))),
				i(named("PrimaryExpr", rules(
					o(`p:PrimaryExpr a:Arguments`, fPrimaryExprArguments),                            // e.g. `math.Atan2(x, y)`
					o(`p:PrimaryExpr i:Index`, fPrimaryExprIndex),                                    // e.g. `something[1]`
					o(`p:PrimaryExpr s:Slice`, fPrimaryExprSlice),                                    // e.g. `a[23 : 87]`
					o(`p:PrimaryExpr s:Selector`, fPrimaryExprSelector),                              // e.g. `x.f` for a PrimaryExpr x that is not a package name
					o(`primaryExpr:PrimaryExpr typeAssertion:TypeAssertion`, fPrimaryExprTypeAssert), // e.g. `x.(T)`
					// o(named("Conversion", rules( o(`Type '(' Expression ','? ')'`),))), // NOT Conversion: are not in helpers.go X() and would create ambiguity
					o(named("Operand", rules(
						// o("'(' Expression ')' | OperandName TypeArgs? | Literal"), // TODO this is the original
						o(`Literal | OperandName | '(' Expression ')'`),
						i(named("Literal", rules(
							o(`BasicLit | FunctionLit | CompositeLit`),
							i(named("BasicLit", rules(
								o("(?'\\'') rune_lit | (?[\"`]) string_lit | (?[0-9.]) imaginary_lit | (?[0-9.]) float_lit | (?[0-9]) int_lit"), // (?xx) is lookahead, quickly ruling out rules away
								i(named("rune_lit", `'\'' ( byte_value | unicode_value | [^\n] ) '\''`), f_rune_lit),
								i(named("string_lit", rules(
									o(named("raw_string_lit", "'`' [^`]* '`'"), fraw_string_lit),
									o(named("interpreted_string_lit", `'"' (!'\"' ('\\' [\s\S] | unicode_value | byte_value))* '"'`), finterpreted_string_lit),
								))),
								i(named("int_lit", rules(
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
								i(named("byte_value", rules(
									o(`  (?'\\' octal_digit) (octal_byte_value_err1 | octal_byte_value | octal_byte_value_err2) |`+
										`(?'\\x'           ) (hex_byte_value_err1 | hex_byte_value | hex_byte_value_err2)`),
									i(named("octal_byte_value_err1", `a:'\\' (?octal_digit{4,})`), ffPanic("illegal: too many octal digits")),
									i(named("octal_byte_value", `a:'\\' b:octal_digit{3,3}`), foctal_byte_value), // passthru unless "illegal: octal value over 255"
									i(named("octal_byte_value_err2", `a:'\\' (?octal_digit{1,})`), ffPanic("illegal: too few octal digits")),
									i(named("hex_byte_value_err1", `a:'\\x' b:hex_digit{3,}`), ffPanic("illegal: too many hexadecimal digits")),
									i(named("hex_byte_value", `a:'\\x' b:hex_digit{2,2}`)),
									i(named("hex_byte_value_err2", `a:'\\x' b:hex_digit{0,1}`), ffPanic("illegal: too few hexadecimal digits")),
								))),
								i(named("unicode_value", rules(
									o("escaped_char | little_u_value | big_u_value | unicode_char | _error_unicode_char_toomany"),
									i(named("escaped_char", `esc:'\\' char:[abfnrtv\\\'"]`)),
									i(named("little_u_value", `a:'\\u' b:hex_digit*`), ff_u_value("little_u_value", 4)), // 4 hex_digit or error
									i(named("big_u_value", `a:'\\U' b:hex_digit*`), ff_u_value("big_u_value", 8)),       // 8 hex digit or error
									i(named("_error_unicode_char_toomany", "[^\\x{0a}]{2,}"), func(it j.Ast, ctx *j.ParseContext) j.Ast { return ctx.Error("too many characters") }),
								))),
							))),
							i(named("FunctionLit", rules( // "A function literal represents an anonymous function"
								o(`'func' Signature FunctionBody`),
								i(named("FunctionBody", `Block`)), // TODO Block only allows SimpleStmt for now
							))),
							i(named("CompositeLit", rules(
								o(`LiteralType LiteralValue`, fCompositeLit),
								i(named("LiteralType", `StructType | ArrayType | AutoLengthElementType | SliceType | MapType | TypeName `)), // TODO add TypeArgs?  OPTIM consider `[]int{1,2,3}`, there could be an EmptyArrayType to accelerate parsing.
								i(named("AutoLengthElementType", `'[' '...' ']' ElementType`)),
								i(named("ElementType", `Element`)),
								i(named("LiteralValue", `'{' ElementList*_COMMA '}'`), peel),
								i(named("ElementList", `KeyedElement*_COMMA`)),
								i(named("KeyedElement", `(Key ':')? Element`), fKeyedElement), // OPTIM could benefit from lookahead (Key is quite broad)
								i(named("Key", `FieldName | Expression | LiteralValue`)),
								i(named("FieldName", `identifier`)),
								i(named("Element", `Expression | LiteralValue`)),
							))),
						))),
						i(named("OperandName", rules(
							o(`QualifiedIdent | identifier`),
							i(named("QualifiedIdent", `p:PackageName DOT i:identifier`), fQualifiedIdent),
						))),
					))),
					// TODO add to below alternation: Type [ "," ExpressionList ]
					i(named("Arguments", `'(' (Args:(ExpressionList) Varg:MaybeVariadic ','? )? ')'`), fArguments),
					i(named("Index", `'[' Expression ','? ']'`)),
					i(named("Slice", `'[' (Expression?)*':'{2,3} ']'`)),
					i(named("Selector", `'.' identifier`)),
					i(named("TypeAssertion", `'.' '(' Type ')'`)),
					i(named("Type", rules(
						o(named("TypeLit", rules( // "A type may also be specified using a type literal, which composes a type from existing types."
							o(named("MapType", `'map[' Type ']' Type`), fMapType),
							o(named("SliceType", `'[]' Type`), func(it j.Ast) j.Ast { return &SliceTypeExpr{Elt: it.(Expr), Vrd: false} }),
							o(named("ArrayType", `'[' Expression ']' Type`), fArrayType),
							o(named("ChannelType", `chanDir:('chan<-' | '<-chan' | 'chan') _:' '? type:Type`), fChannelType),
							o(named("PointerType", `'*' Type`), func(it j.Ast) j.Ast { return &StarExpr{X: it.(Expr)} }), // nodes.go: "[StarExpr] semantically (...) could be unary * expression, or a pointer type."
							o(named("FunctionType", rules(
								o(`'func' &:Signature`),
								i(named("Signature", `params:Parameters result:Result?`), fSignature),
								i(named("Parameters", `'(' ParameterDecl*_COMMA _COMMA? ')'`), fParameters), // "Within a list of parameters or results, the names must be all present or all absent"
								i(named("ParameterDecl", `IdentifierList? _ MaybeVariadic Type`), fParameterDecl),
								i(named("Result", `Parameters | Type`), fResult),
							))),
							o(named("StructType", rules(
								o(`'struct' '{' FieldDecl*_SEMICOLON '}'`, fStructType),
								i(named("FieldDecl", rules(
									o(`IdentifierList _ Type Tag?`, fFieldDecl1),
									o(`EmbeddedField Tag?`, fFieldDecl2),
									i(named("EmbeddedField", `star:MaybeStar typename:TypeName typeargs:TypeArgs?`), fEmbeddedField),
									i(named("Tag", `string_lit`)),
								))),
							))),
						// o(named("InterfaceType", "")),
						))),
						o(`TypeName TypeArgs?`, fTypeName),
						o(`'(' Type ')'`),
						i(named("TypeName", `identifier | QualifiedIdent`)),
						i(named("TypeArgs", `'[' Type*',' ','? ']'`)), // note: TypeArgs seems not supported by X() ATM
					))),
				))),
			)) /*, fExpression*/),
			i(named("ExpressionList", `Expression+_COMMA`)),
		))),
		i(named("PackageClause", `'package' PackageName`)),
		i(named("PackageName", `identifier`), fPackageName),
		i(named("identifier", `letter (letter | unicode_digit)*`), fIdentifier),
		i(named("IdentifierList", `identifier+_COMMA`)),
		i(named("characters", `(newline | unicode_char | unicode_letter | unicode_digit)`)),
		i(named("newline", `[\x{0a}]`)),
		i(named("letter", `(?[^0-9 \t\n\r+(){}[\]<>-])`), fLetter),                 // lookahead next rune, if not impossibly a letter try to parse with fLetter using unicode.IsLetter(). gospec = "unicode_letter | '_'"
		i(named("unicode_char", `[^\x{0a}]`)),                                      // "an arbitrary Unicode code point except newline"
		i(named("unicode_letter", `(?[^0-9 \t\n\r+(){}[\]<>-])`), funicode_letter), // lookahead next rune, if not impossibly a letter try etc. "a Unicode code point categorized as "Letter" TODO it misses all non ASCII
		i(named("unicode_digit", `(?[^a-zA-Z \t\n\r-])`), funicode_digit),          // lookahead next rune, if not impossibly a digit, try to unicode.IsDigit(). "a Unicode code point categorized as "Number, decimal digit" TODO it misses all non ASCII

		i(named("MaybeVariadic", `'...'?`), func(it j.Ast) j.Ast { return j.NewNativeIntFromBool(!j.IsUndefined(it)) }), // NativeInt 0 or 1
		i(named("MaybeStar", `'*'?`), func(it j.Ast) j.Ast { return j.NewNativeIntFromBool(!j.IsUndefined(it)) }),       // NativeInt 0 or 1
		i(named("_COMMA", `_ ','`)),
		i(named("_", `( ' ' | '\t' | '\n' | '\r' )*`)),
		i(named("DOT", `'.'`)), // using an inline ref such as DOT will capture '.', as opposed to writing '.' which would not.
		i(named("_SEMICOLON", "';' '\n'?")),
	)
)

// "quote me on this" helps writing rules for PEG grammars.
// Split upon space, add single quotes and join upon '|'.
// E.g. qmot("* / %") -> "'*'|'/'|'%'".
func qmot(spaceSeparatedElements string) string {
	a := strings.Fields(spaceSeparatedElements)
	s := ""
	for i := 0; i < len(a); i++ {
		s += "'" + a[i] + "'|"
	}
	return s[:len(s)-1]
}

// tip: with vim, fold open and close the rules with 'zo' and 'zc', by
// indentation level
// vim: fdm=indent fdl=4
