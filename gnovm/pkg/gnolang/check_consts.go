package gnolang

import (
	"fmt"
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
			failf(CodeImportPath, nil, "invalid import path (invalid character %#U)", r)
		}
	}
	if strings.HasPrefix(path, "/") {
		failf(CodeImportPath, nil, "import path cannot be absolute path: %q", path)
	}
}

// constWidth is the width go/types gives untyped integer constants.
const constWidth = 512

// maxConstShift is the largest count go/types accepts in a constant
// shift of a constant: enough to express the smallest float64.
const maxConstShift = 1023 - 1 + 52

// assertConstShiftCount rejects a constant shift of a constant by more
// than maxConstShift, before the fold, as go/types does. A negative
// count has its own error.
func assertConstShiftCount(rcx *ConstExpr) {
	if rcx.T == nil {
		return
	}
	over := false
	switch rcx.T.Kind() {
	case BigintKind:
		b := rcx.GetBigInt()
		over = b.Sign() >= 0 && (b.BitLen() > 63 || b.Int64() > maxConstShift)
	case Uint64Kind:
		over = rcx.GetUint64() > maxConstShift
	case IntKind, Int8Kind, Int16Kind, Int32Kind, Int64Kind, UintKind, Uint8Kind, Uint16Kind, Uint32Kind:
		tv := rcx.TypedValue // ConvertGetInt converts in place
		over = tv.ConvertGetInt() > maxConstShift
	}
	if !over {
		return
	}
	what := "untyped int constant"
	if !isUntyped(rcx.T) {
		what = fmt.Sprintf("constant %s of type %s", rcx.TypedValue.Sprint(nil), rcx.T)
	}
	expr := exprString(rcx)
	if nx, ok := rcx.Source.(*NameExpr); ok {
		expr = string(nx.Name)
	}
	failf(CodeShiftCount, nil, "invalid operation: invalid shift count %s (%s)", expr, what)
}

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
		// A shifted float literal is converted first; go/types names no
		// operation.
		if lcx, ok := x.Left.(*ConstExpr); ok && (x.Op == SHL || x.Op == SHR) {
			if bl, ok := lcx.Source.(*BasicLitExpr); ok && bl.Kind == FLOAT {
				op = ""
			}
		}
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
	failBranchLabel(n)
}

// failBranchLabel rejects the label of a break or continue whose target
// cannot be branched to.
func failBranchLabel(n *BranchStmt) {
	if n.Op == BREAK {
		failf(CodeBranchLabel, nil, "invalid break label %s", n.Label)
	}
	failf(CodeBranchLabel, nil, "invalid continue label %s", n.Label)
}
