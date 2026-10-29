package gnolang

import (
	"fmt"
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
	CodeMultiValue      ErrCode = "multi-value"

	// Declarations and statements.
	CodeMissingConstInit   ErrCode = "missing-const-init"
	CodeVariadicLast       ErrCode = "variadic-last"
	CodeStructLiteralCount ErrCode = "struct-literal-count"
	CodeStructFieldName    ErrCode = "struct-field-name"
	CodeForPostDecl        ErrCode = "for-post-decl"
	CodeUnusedVar          ErrCode = "unused-var"
	CodeUnusedLabel        ErrCode = "unused-label"
	CodeImportPath         ErrCode = "import-path"
	CodeConstOverflow      ErrCode = "const-overflow"
	CodeBranchLabel        ErrCode = "branch-label"
	CodeShiftCount         ErrCode = "shift-count"

	// Assignability.
	CodeMapValueAssign ErrCode = "map-value-assign"
	CodeNotAssignable  ErrCode = "not-assignable"
	CodeMissingMapKey  ErrCode = "missing-map-key"
	CodeDuplicateField ErrCode = "duplicate-field"
	CodeLiteralType    ErrCode = "literal-type"
	CodeUndefined      ErrCode = "undefined"
	CodeTypeGuard      ErrCode = "type-guard"
	CodeIndexOperand   ErrCode = "index-operand"
	CodeAppendArgs     ErrCode = "append-args"
	CodeImportName     ErrCode = "import-name"
	CodeStructField    ErrCode = "struct-field"
	CodeNoMethod       ErrCode = "no-method"
	CodeReceiverType   ErrCode = "receiver-type"
	CodeAssignability  ErrCode = "assignability"
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
	// fset and index: in accumulate mode declarations predefine through an
	// index carrying the accumulator, so undefined names and failing
	// dependencies resolve to invalidType instead of dropping the
	// dependent; the index is built on first use.
	fset  *FileSet
	index *predefineDeclIndex
}

// report raises a checking error. In accumulate mode it is recorded and
// the walk goes on, so a node can report more than one.
func (c *checkRun) report(code ErrCode, msg string, at ...Node) {
	if c.acc == nil {
		fail(code, msg, at...)
	}
	func() {
		defer func() {
			c.acc.record(wrapPreprocessPanic(*c.stack, (*c.stack)[len(*c.stack)-1], recover()))
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
		c.acc.record(pe)
	}
}

// onPanic is the transcriber's handler in accumulate mode: a failed
// statement is dropped, the block stack rewound to its enclosing block,
// and the walk goes on; a failed literal element or switch clause drops
// alone. A non-checking panic (a VM bug) surfaces.
func (c *checkRun) onPanic(ns []Node, ftype TransField, index int, n Node, r any) (Node, bool) {
	pe, ok := r.(*PreprocessError)
	if !ok || isRuntimePanic(pe) {
		return nil, false // not a checking error: a VM bug, let it surface
	}
	stack := *c.stack
	rewound := false
	for i := len(ns) - 1; i >= 0 && !rewound; i-- {
		if bn, ok := ns[i].(BlockNode); ok {
			for k := len(stack) - 1; k > 0; k-- {
				if stack[k] == bn {
					stack = stack[:k+1]
					break
				}
			}
			rewound = true
		}
	}
	if !rewound {
		stack = stack[:1] // every block above ctx was pushed below n
	}
	*c.stack = stack
	*c.last = stack[len(stack)-1]
	switch ftype {
	case TRANS_COMPOSITE_KEY, TRANS_COMPOSITE_VALUE:
		// the element alone drops: typed invalid, the literal's rule
		// passes it (a key so typed is removed there).
		c.acc.record(pe)
		return invalidConstExpr(n.(Expr)), true
	case TRANS_IF_COND, TRANS_FOR_COND, TRANS_SWITCH_X:
		// the condition alone drops, so the body is still checked; the
		// names it mentions count as used, as gc marks them before the
		// operation fails.
		c.acc.markReferenced(n)
		c.acc.record(pe)
		return invalidConstExpr(n.(Expr)), true
	case TRANS_SWITCH_CASE:
		n.SetAttribute(ATTR_PREPROCESS_DROPPED, true) // the switch skips it
	}
	c.acc.dropStmt(pe, n, *c.last)
	return n, true
}

// predefineAcc is predefineRecursively through the accumulator's index
// in accumulate mode (see checkRun.index); plain otherwise.
func (c *checkRun) predefineAcc(last BlockNode, d Decl) bool {
	if c.acc == nil {
		return predefineRecursively(c.store, last, d)
	}
	if c.index == nil {
		if c.fset == nil {
			return predefineRecursively(c.store, last, d)
		}
		c.index = newPredefineDeclIndex(c.fset)
		c.index.acc = c.acc
	}
	return predefineRecursivelyIndexed(c.store, last, d, c.index)
}

// predefineGuarded is predefineAcc with the drop-and-record rule for
// file-level declarations in accumulate mode: a dropped declaration is
// marked preprocessed so the walk skips it, and its names become invalid.
func (c *checkRun) predefineGuarded(last BlockNode, d Decl) {
	if c.acc == nil {
		predefineRecursively(c.store, last, d)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			if isRuntimePanic(r) {
				panic(r) // a VM bug, not a checking error
			}
			c.acc.dropDecl(wrapPreprocessPanic(*c.stack, d, r), last, d)
		}
	}()
	c.predefineAcc(last, d)
}
