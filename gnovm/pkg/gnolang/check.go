package gnolang

import (
	"fmt"
	"runtime"
)

// The preprocessor's checking half: the per-run context the rule files
// (check_*.go, one per Go spec chapter) report through, and the codes
// that identify their rules independently of the wording.

// ErrCode identifies a checking rule. Error text may change with the
// wording; the code does not.
type ErrCode string

const (
	CodeNone ErrCode = "" // a site not yet given a code

	// Types.
	CodeMapKey           ErrCode = "map-key"
	CodeEmbeddedField    ErrCode = "embedded-field"
	CodeArrayLength      ErrCode = "array-length"
	CodeArrayEllipsis    ErrCode = "array-ellipsis"
	CodeDuplicateCase    ErrCode = "duplicate-case"
	CodeMultipleDefaults ErrCode = "multiple-defaults"
	CodeImpossibleAssert ErrCode = "impossible-assertion"
	CodeImpossibleCase   ErrCode = "impossible-case"
	CodeTypeSwitchTag    ErrCode = "type-switch-tag"
	CodeNotAType         ErrCode = "not-a-type"
	CodeCopyArgs         ErrCode = "copy-args"
	CodeRecursiveType    ErrCode = "recursive-type"
	CodeRecursiveValue   ErrCode = "recursive-value"
	CodeInitCycle        ErrCode = "init-cycle"
	CodeFieldNotMethod   ErrCode = "field-not-method"
	CodeInterfaceEmbed   ErrCode = "interface-embed"

	// Expressions.
	CodeNotComparable   ErrCode = "not-comparable"
	CodeNilOnly         ErrCode = "nil-only"
	CodeLenCapArg       ErrCode = "len-cap-arg"
	CodeIndexBounds     ErrCode = "index-bounds"
	CodeSliceIndices    ErrCode = "slice-indices"
	CodeDuplicateIndex  ErrCode = "duplicate-index"
	CodeNotUsed         ErrCode = "not-used"
	CodeNotExpression   ErrCode = "not-expression"
	CodeLiteralKeyConst ErrCode = "literal-key-const"
	CodeAppendNil       ErrCode = "append-nil"
	CodeBuiltinSpread   ErrCode = "builtin-spread"
	CodeFuncAddress     ErrCode = "func-address"

	// Declarations and statements.
	CodeMissingConstInit   ErrCode = "missing-const-init"
	CodeVariadicLast       ErrCode = "variadic-last"
	CodeStructLiteralCount ErrCode = "struct-literal-count"
	CodeForPostDecl        ErrCode = "for-post-decl"
	CodeUnusedVar          ErrCode = "unused-var"
	CodeUnusedLabel        ErrCode = "unused-label"
	CodeImportPath         ErrCode = "import-path"
	CodeConstOverflow      ErrCode = "const-overflow"
	CodeBranchLabel        ErrCode = "branch-label"
	CodeNonNameDefine      ErrCode = "non-name-define"
	CodeRepeatedDefine     ErrCode = "repeated-define"
	CodeSwitchTag          ErrCode = "switch-tag"
	CodeConversionSpread   ErrCode = "conversion-spread"
	CodeMethodExprReceiver ErrCode = "method-expr-receiver"
	CodeNoMethod           ErrCode = "no-method"

	// Assignability.
	CodeMapValueAssign ErrCode = "map-value-assign"
	CodeNotAssignable  ErrCode = "not-assignable"
)

// spanError is a checking error raised by a rule: its code, its text, and
// the node it is positioned at (zero: the node being transcribed).
type spanError struct {
	code ErrCode
	span Span
	msg  string
}

func (p *spanError) Error() string { return p.msg }

// fail raises the checking error code with text msg, positioned at the
// first of at with a span; a folded constant borrows its source's.
func fail(code ErrCode, msg string, at ...Node) {
	for _, x := range at {
		for x != nil && x.GetSpan().IsZero() {
			switch cx := x.(type) {
			case *ConstExpr:
				x = cx.Source
			case *constTypeExpr:
				x = cx.Source
			default:
				x = nil
			}
		}
		if x != nil {
			panic(&spanError{code: code, span: x.GetSpan(), msg: msg})
		}
	}
	panic(&spanError{code: code, msg: msg})
}

// failf is fail with a format.
func failf(code ErrCode, at []Node, format string, args ...any) {
	fail(code, fmt.Sprintf(format, args...), at...)
}

// checkRun is the context of one preprocess walk: the store, the block
// stack the walk maintains (shared by pointer, the walk owns it) and, in
// accumulate mode, the error sink. On chain (acc == nil) the first error
// panics exactly as before.
type checkRun struct {
	store Store
	acc   *preprocessAccumulator
	stack *[]BlockNode
	last  *BlockNode
}

// report raises a checking error. In accumulate mode it is recorded and
// the walk goes on, so a node can report more than one.
func (c *checkRun) report(code ErrCode, msg string, at ...Node) {
	if c.acc == nil {
		fail(code, msg, at...)
	}
	func() {
		defer func() {
			c.acc.report(wrapPreprocessPanic(*c.stack, (*c.stack)[len(*c.stack)-1], recover()))
		}()
		fail(code, msg, at...)
	}()
}

// reportUnused raises the unused locals and labels of the function fn.
func (c *checkRun) reportUnused(parent BlockNode, fn Node) {
	for _, u := range checkUnused(preprocessGasMeterOf(c.store), parent, fn) {
		if c.acc != nil && (c.acc.poisoned[u.name] || c.acc.referenced[u.name]) {
			continue // a dropped statement declared or mentioned it
		}
		loc := (*c.last).GetLocation()
		loc.SetSpanOverride(u.span)
		pe := &PreprocessError{err: fmt.Errorf("%s: %s", loc.String(), u.msg), stack: *c.stack, code: u.code}
		if c.acc == nil {
			panic(pe)
		}
		c.acc.report(pe)
	}
}

// onPanic is the transcriber's handler in accumulate mode: a failed
// statement is dropped, the block stack rewound to its enclosing block,
// and the walk goes on. A non-checking panic (a VM bug) surfaces.
func (c *checkRun) onPanic(ns []Node, ftype TransField, index int, n Node, r any) bool {
	pe, ok := r.(*PreprocessError)
	if !ok {
		return false
	}
	stack := *c.stack
	for i := len(ns) - 1; i >= 0; i-- {
		if bn, ok := ns[i].(BlockNode); ok {
			for k := len(stack) - 1; k > 0; k-- {
				if stack[k] == bn {
					stack = stack[:k+1]
					break
				}
			}
			break
		}
	}
	*c.stack = stack
	*c.last = stack[len(stack)-1]
	c.acc.add(pe, n)
	return true
}

// predefineGuarded is predefineRecursively with the same drop-and-record
// rule for file-level declarations in accumulate mode; a dropped
// declaration is marked preprocessed so the walk skips it.
func (c *checkRun) predefineGuarded(last BlockNode, d Decl) {
	if c.acc == nil {
		predefineRecursively(c.store, last, d)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			if _, isRuntime := r.(runtime.Error); isRuntime {
				panic(r) // a VM bug, not a checking error
			}
			c.acc.add(wrapPreprocessPanic(*c.stack, d, r), d)
			d.SetAttribute(ATTR_PREPROCESSED, true)
		}
	}()
	predefineRecursively(c.store, last, d)
}
