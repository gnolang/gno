package gnolang

import (
	"fmt"
)

type staticAnalysis struct {
	// contexts for switch, for, functions, and lambdas
	contexts []any

	// here we accumulate errors
	// from lambdas defined in the function declaration
	errs []error
}

func newStaticAnalysis() *staticAnalysis {
	return &staticAnalysis{
		contexts: make([]any, 0),
		errs:     make([]error, 0),
	}
}

func Analyze(f *FuncDecl) []error {
	s := newStaticAnalysis()
	s.push(&FuncDeclContext{
		hasRet: false,
		f:      f,
	})
	term := s.staticAnalysisBlockStmt(f.Body)

	errs := make([]error, 0)
	if !term {
		errs = append(errs, fmt.Errorf("function %q does not terminate", f.Name))
	}

	errs = append(errs, s.errs...)

	return errs
}

func (s *staticAnalysis) push(ctx any) {
	s.contexts = append(s.contexts, ctx)
}

func (s *staticAnalysis) pop() any {
	if len(s.contexts) == 0 {
		return nil
	}
	last := s.contexts[len(s.contexts)-1]
	s.contexts = s.contexts[:len(s.contexts)-1]
	return last
}

// findCtxByLabel returns the last context if the label is empty
// otherwise it returns the context that matches the label
// if it doesn't exist, it returns nil
func (s *staticAnalysis) findCtxByLabel(label string) any {
	if len(label) == 0 {
		if len(s.contexts) > 0 {
			return s.contexts[len(s.contexts)-1]
		}
		return nil
	}

	for i := len(s.contexts) - 1; i >= 0; i-- {
		if ctx, ok := s.contexts[i].(contextLabeler); ok && ctx.label() == label {
			return s.contexts[i]
		}
	}

	return nil
}

func (s *staticAnalysis) staticAnalysisBlockStmt(stmts []Stmt) bool {
	if len(stmts) == 0 {
		return false
	}
	return s.staticAnalysisStmt(stmts[len(stmts)-1])
}

func (s *staticAnalysis) staticAnalysisExpr(expr Expr) bool {
	switch n := expr.(type) {
	case *CallExpr:
		for _, arg := range n.Args {
			term := s.staticAnalysisExpr(arg)
			if !term {
				return false
			}
		}
	case *FuncLitExpr:
		s.push(&FuncLitContext{
			hasRet: false,
			f:      n,
		})
		term := s.staticAnalysisBlockStmt(n.Body)
		if !term {
			ctx := s.pop().(*FuncLitContext)
			s.errs = append(s.errs, fmt.Errorf("lambda at %v does not terminate", ctx.f.GetLocation()))
		}
		return false
	case *NameExpr:
		return false
	}
	return false
}

// staticAnalysisStmt returns a boolean value,
// indicating whether a statement is terminating or not
func (s *staticAnalysis) staticAnalysisStmt(stmt Stmt) bool {
	switch n := stmt.(type) {
	case *BlockStmt:
		return s.staticAnalysisBlockStmt(n.Body)
	case *BranchStmt:
		switch n.Op {
		case BREAK:
			ctx := s.findCtxByLabel(string(n.Label))
			if ctx != nil {
				if c, ok := ctx.(breakPusher); ok {
					c.pushBreak(n)
				}
			}
		case CONTINUE:
			//
		case DEFAULT:
			//
		case FALLTHROUGH:
			return true
		}
	case *ExprStmt:
		x := n.X
		if cs, ok := x.(*CallExpr); ok {
			if nx, ok := cs.Func.(*NameExpr); ok {
				if nx.Name == "panic" {
					return true
				}
			}
		}
	case *IfStmt:
		terminates := s.staticAnalysisBlockStmt(n.Then.Body)

		var elseTerminates bool
		if len(n.Else.Body) > 0 {
			elseTerminates = s.staticAnalysisBlockStmt(n.Else.Body)
		}

		return terminates && elseTerminates
	case *ForStmt:
		s.push(&ForContext{forstmt: n})
		_ = s.staticAnalysisBlockStmt(n.BodyBlock.Body)
		ctx := s.pop().(*ForContext)
		// there are no "break" statements referring to the "for" statement
		hasNoBreaks := len(ctx.breakstmts) == 0
		// the loop condition is absent
		hasNoCond := n.Cond == nil
		terminates := hasNoBreaks && hasNoCond
		return terminates
	case *ReturnStmt:
		return true
	case *AssignStmt:
		for _, rh := range n.Rhs {
			term := s.staticAnalysisExpr(rh)
			if !term {
				return false
			}
		}
		return false
	case *SwitchStmt:
		// there is a default case, and
		var hasDefault bool
		for _, clause := range n.Clauses {
			// nil case means default
			if clause.Cases == nil {
				hasDefault = true
				break
			}
		}
		s.push(&SwitchContext{switchStmt: n})

		// the statement lists in each case,
		// including the default
		// end in a terminating statement,
		// or a possibly labeled "fallthrough" statement.
		casesTerm := true
		for _, clause := range n.Clauses {
			ct := s.staticAnalysisBlockStmt(clause.Body)
			casesTerm = ct
		}
		ctx := s.pop().(*SwitchContext)
		// there are no "break" statements referring to the "switch" statement
		hasNoBreaks := len(ctx.breakstmts) == 0
		terminates := hasNoBreaks && hasDefault && casesTerm
		return terminates
	}
	return false
}

type contextLabeler interface {
	label() string
}

type breakPusher interface {
	pushBreak(breakstmt *BranchStmt)
}

type FuncLitContext struct {
	hasRet bool
	f      *FuncLitExpr
}

type FuncDeclContext struct {
	hasRet bool
	f      *FuncDecl
}

type ForContext struct {
	forstmt    *ForStmt
	breakstmts []*BranchStmt
}

func (fc *ForContext) label() string {
	return string(fc.forstmt.Label)
}

func (fc *ForContext) pushBreak(breakstmt *BranchStmt) {
	fc.breakstmts = append(fc.breakstmts, breakstmt)
}

type SwitchContext struct {
	switchStmt *SwitchStmt
	breakstmts []*BranchStmt
}

func (sc *SwitchContext) label() string {
	return string(sc.switchStmt.Label)
}

func (sc *SwitchContext) pushBreak(breakstmt *BranchStmt) {
	sc.breakstmts = append(sc.breakstmts, breakstmt)
}
