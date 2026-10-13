package gnolang

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// Rules for import paths, constants and branch statements (Go spec:
// "Import declarations", "Constant expressions", "Break statements",
// "Continue statements").

// assertValidImportPath checks the spelling of an import path: not
// empty, valid UTF-8, graphic characters outside Go's excluded set, and
// not absolute.
func assertValidImportPath(path string) {
	if path == "" {
		fail(CodeImportPath, "invalid import path (empty string)")
	}
	if !utf8.ValidString(path) {
		failf(CodeImportPath, nil, "import path contains invalid UTF-8 character: %q", path)
	}
	for _, r := range path {
		if !unicode.IsGraphic(r) || unicode.IsSpace(r) ||
			strings.ContainsRune("!\"#$%&'()*,:;<=>?[\\]^`{|}", r) {
			failf(CodeImportPath, nil, "invalid import path %q: invalid character %q", path, r)
		}
	}
	if strings.HasPrefix(path, "/") {
		failf(CodeImportPath, nil, "import path cannot be absolute path: %q", path)
	}
}

// constWidth is the width go/types gives untyped integer constants.
const constWidth = 512

// assertConstWidth rejects a folded untyped integer constant wider than
// constWidth bits, naming the operation that produced it.
func assertConstWidth(x Expr, cv *TypedValue) {
	if cv.T != UntypedBigintType || cv.GetBigInt().BitLen() <= constWidth {
		return
	}
	op := ""
	switch x := x.(type) {
	case *BinaryExpr:
		op = constOpNames[x.Op]
	case *UnaryExpr:
		switch x.Op {
		case XOR:
			op = "bitwise complement"
		case SUB:
			op = "negation"
		}
	}
	if op == "" {
		fail(CodeConstOverflow, "constant overflow")
	}
	failf(CodeConstOverflow, nil, "constant %s overflow", op)
}

var constOpNames = map[Word]string{
	ADD: "addition", SUB: "subtraction", MUL: "multiplication", QUO: "division",
	SHL: "shift", SHR: "shift", XOR: "bitwise XOR", BAND: "bitwise AND",
	BOR: "bitwise OR", BAND_NOT: "bitwise AND NOT",
}

// assertBranchTarget checks that the statement a labeled break or
// continue names can be its target: a loop, or for break also a switch
// or select.
func assertBranchTarget(n *BranchStmt, bn BlockNode) {
	switch bn.(type) {
	case *ForStmt, *RangeStmt:
		return
	case *SwitchClauseStmt, *SelectCaseStmt:
		if n.Op == BREAK {
			return
		}
	}
	if n.Op == BREAK {
		failf(CodeBranchLabel, nil, "invalid break label %s", n.Label)
	}
	failf(CodeBranchLabel, nil, "invalid continue label %s", n.Label)
}
