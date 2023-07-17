package gnolang

import (
	"fmt"
	// "sort"
	"strings"
	"testing"
	// "github.com/jaekwon/testify/assert"
)

const (
	fail = "\x1b[31mFAIL\x1b[0m"
	ok   = "\x1b[32m✓ \x1b[0m"
)

// Using X(s), asserts `s` gives an an ast type beginning with string `expect`.
// Two special cases:
//
//  1. When `expect` starts with "ERROR", it means a parse error is expected.
//     You can specify the exact error like so: "ERROR illegal: octal value over 255".
//
// 2. When `expect` is "", the test passes as long as parsing did not fail.
func test(t *testing.T, s string, expect string) {
	t.Helper()
	expr := X(s)
	// if j.IsParseError(ast) {
	// 	if strings.HasPrefix(expect, "ERROR") {
	// 		fmt.Printf("[32m%s[0m gave an error as expected [32m✓[0m\n", s)
	// 	} else {
	// 		t.Fatalf("Error parsing %s. Expected ast.ContentString() to contain '%s', got '%s'", s, expect, ast.ContentString())
	// 	}
	// } else {
	if strings.Contains(expr.String(), expect) {
		fmt.Printf("[32m%s[0m parsed as [33m%s[0m [32m✓[0m %s\n", s, expr.String(), expect)
	} else {
		t.Fatalf(
			"Error, \"[1m%s[0m\" [1;31mparsed[0m as %s [1;31mbut expected [0;31m%s[0m",
			s,
			expr.String(),
			expect,
		)
	}
	// }
}

func TestJoeson(t *testing.T) {
	initGrammar()
	// TODO group them into suite?
	tests := map[string]string{
		"992 + 293":            "992 + 293",
		"-1234":                "-1234", // UnaryExpr
		"- 1234":               "-1234",
		"+ 1234":               "+1234",
		"!0":                   "!0",
		"^0":                   "^0",
		"-7 -2":                "-7 - 2",
		"2398":                 "2398", // Operand Literal BasicLit decimal_lit
		"0":                    "0",
		"0b0":                  "0b0", // Operand Literal BasicLit binary_lit
		"0B1":                  "0B1",
		"0B_1":                 "0B_1",
		"0B_10":                "0B_10",
		"0O777":                "0O777", // Operand Literal BasicLit octal_lit
		"0o1":                  "0o1",
		"0xBadFace":            "0xBadFace",
		"0xBad_Face":           "0xBad_Face",
		"0x_67_7a_2f_cc_40_c6": "0x_67_7a_2f_cc_40_c6",
		// "func(a, b int, z float64) bool { return a*b < int(z) }": "func(a, b int, z float64) bool { return a*b < int(z) }", // FunctionLit
	}
	// sort.Strings(tests)
	for k, v := range tests {
		test(t, k, v)
	}
}
