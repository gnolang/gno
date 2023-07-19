package gnolang

import (
	"fmt"
	// "reflect"
	"strings"
	"testing"

	// "sort"
	j "github.com/grepsuzette/joeson"
	"github.com/grepsuzette/joeson/helpers"
	// "github.com/jaekwon/testify/assert"
)

// Using X(s), asserts `s` gives an an ast type beginning with strings in `expects`.
// All values in `expects` must be contained in rendered parsed expression.
// In the rest of this comment we will use `expect` to designate iterated values.
// Two special cases:
//
//  1. When `expect` starts with "ERROR", it means a parse error is expected.
//     You can specify the exact error like so: "ERROR illegal: octal value over 255".
//
// 2. When `expect` is "", the test passes as long as parsing did not fail.
func test(t *testing.T, s string, expects ...string) {
	t.Helper()
	expr := X(s)
	// TODO figure out what to do. X() currently panics if there was a ParseError.
	// (  Maybe we can use Attributes ? )
	// if j.IsParseError(ast) {
	// 	if strings.HasPrefix(expect, "ERROR") {
	// 		fmt.Printf("[32m%s[0m gave an error as expected [32m✓[0m\n", s)
	// 	} else {
	// 		t.Fatalf("Error parsing %s. Expected ast.ContentString() to contain '%s', got '%s'", s, expect, ast.ContentString())
	// 	}
	// } else {
	resultString := ""
	if w, isWrapped := expr.(wrapped); isWrapped {
		resultString = w.String() // + " [wrapped]"
	} else if expr.HasAttribute("joeson") {
		ast := expr.GetAttribute("joeson")
		resultString = StringWithRulenames(ast.(j.Ast)) // + " [wraPPed]"
	} else {
		resultString = expr.String()
	}
	allOk := true
	for _, expect := range expects {
		if !strings.Contains(resultString, expect) {
			t.Fatalf(
				"Error, %s "+j.BoldRed("parsed as")+" %s "+j.BoldRed("but expected ")+"%s",
				j.Bold(`"`+s+`"`),
				resultString,
				j.Magenta(expect),
			)
			allOk = false
			break
		}
	}
	if allOk {
		fmt.Printf(
			"%s parsed as %s "+j.Green("✓")+" %s\n",
			j.Green(s),
			j.Yellow(resultString),
			strings.Join(helpers.AMap(expects, func(s string) string { return j.Magenta(s) }), ", "),
		)
	}
}

func init() { initGrammar() }

func TestJoesonUnaryExpr(t *testing.T) {
	tests := [][]string{
		{"992 + 293", "Expression", "bx"},
		{"-1234", "UnaryExpr", "decimal_lit"},
		{"- 1234", "UnaryExpr", "decimal_lit"},
		{"+ 1234", "UnaryExpr", "decimal_lit"},
		{"!0", "UnaryExpr"},
		{"^0", "UnaryExpr"},
		{"-7 -2", "Expression", "UnaryExpr", "decimal_lit"},
		{"2398", "decimal_lit"},
		{"0", "decimal_lit", "0"},
		{"0b0", "binary_lit"},
		{"0B1", "binary_lit"},
		{"0B_1", "binary_lit"},
		{"0B_10", "binary_lit"},
		{"0O777", "octal_lit"},
		{"0o1", "octal_lit"},
		{"0xBadFace", "hex_lit"},
		{"0xBad_Face", "hex_lit"},
		{"0x_67_7a_2f_cc_40_c6", "hex_lit"},
		{"1e043", "float_lit"},
		{"1.e+3", "float_lit"},
		{".4e+33493", "float_lit"},
		{".4e-33493", "float_lit"},
		// "func(a, b int, z float64) bool { return a*b < int(z) }": "func(a, b int, z float64) bool { return a*b < int(z) }", // FunctionLit
	}
	// sort.Strings(tests)
	for _, a := range tests {
		test(t, a[0], a[1:]...)
	}
}
