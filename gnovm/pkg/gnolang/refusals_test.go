package gnolang

import (
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// The write rule's refusal texts; each must say what to do instead.
var refusalTexts = []string{
	"cannot modify object owned by another realm from library code (call a function of the owning realm): ",
	"'s code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): ",
	"'s code was entered without crossing and is read-only (call its crossing function): ",
	"cannot directly modify readonly tainted object (call a crossing function of the owning realm): ",
	"%s: an object written by read-only code would be stored (write it inside a crossing call of %s)",
	"cannot cross-call a non-crossing function %s.%v from %s",
	"cannot cur-call to external realm function %s.%v from %s",
	"cannot append to readonly tainted slice",
	"cannot copy to readonly tainted slice",
	"cannot delete from readonly tainted map",
	"illegal conversion of readonly or externally stored value",
	"illegal conversion to external realm type",
	"cannot mutate %s: package is immutable post-init",
}

// refusalRuleFiles hold every site that raises a refusal of the write rule.
var refusalRuleFiles = []string{
	"machine.go", "op_assign.go", "realm.go", "callstart.go", "uverse.go",
	"op_expressions.go", "crossingfn.go", "lend.go", "birthrule.go",
}

var (
	// a bracketed "do this instead": at least two words, not a format verb
	refusalFixRe = regexp.MustCompile(`\([a-z][^()%]*\s[^()]*\)`)
	// words only a refusal of the write rule uses
	refusalWordRe = regexp.MustCompile(`(?i)read-?only|tainted|cross-call|cur-call|immutable post-init|external realm|another realm`)
)

func TestRefusalTextsNameFix(t *testing.T) {
	for _, text := range refusalTexts {
		if !refusalFixRe.MatchString(text) {
			t.Errorf("refusal text names no fix in brackets: %q", text)
		}
	}
	fset := token.NewFileSet()
	for _, name := range refusalRuleFiles {
		f, err := parser.ParseFile(fset, name, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(f, func(n ast.Node) bool {
			lit, ok := n.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			s, err := strconv.Unquote(lit.Value)
			if err != nil || !refusalWordRe.MatchString(s) || strings.HasPrefix(s, "invariant violation:") {
				return true
			}
			for _, text := range refusalTexts {
				if strings.Contains(text, s) {
					return true
				}
			}
			t.Errorf("%s: refusal text not in the list: %q", fset.Position(lit.Pos()), s)
			return true
		})
	}
}
