package gnolang

import (
	"errors"
	"fmt"
)

type StaticAnalysis struct {
	// contexts for switch and for
	contexts []Context
	// funcContexts for functions and lambdas
	funcContexts []FuncContext

	// here we accumulate errors
	// from lambdas defined in the function declaration
	lambdaErrs []error
}

func NewStaticAnalysis() *StaticAnalysis {
	return &StaticAnalysis{contexts: make([]Context, 0), funcContexts: make([]FuncContext, 0), lambdaErrs: make([]error, 0)}
}

func (s *StaticAnalysis) pushContext(ctx Context) {
	s.contexts = append(s.contexts, ctx)
}

func (s *StaticAnalysis) popContext() Context {
	last := s.contexts[len(s.contexts)-1]
	s.contexts = s.contexts[0 : len(s.contexts)-1]
	return last
}

func (s *StaticAnalysis) pushFuncContext(fc FuncContext) {
	s.funcContexts = append(s.funcContexts, fc)
}

func (s *StaticAnalysis) popFuncContext() FuncContext {
	last := s.funcContexts[len(s.funcContexts)-1]
	s.funcContexts = s.funcContexts[0 : len(s.funcContexts)-1]
	return last
}

// findCtxByLabel returns the last context if the label is empty
// otherwise it returns the context that matches the label
// if it doesn't exist, it returns nil
func (s *StaticAnalysis) findCtxByLabel(label string) Context {
	if len(label) == 0 {
		return s.contexts[len(s.contexts)-1]
	}

	for i := len(s.contexts) - 1; i > -1; i-- {
		if s.contexts[i].label() == label {
			return s.contexts[i]
		}
	}

	return nil
}

func (s *StaticAnalysis) Analyse(f *FuncDecl) []error {
	s.pushFuncContext(&FuncDeclContext{
		hasRet: false,
		f:      f,
	})
	term := s.staticAnalysisBlockStmt(f.Body)

	//todo use later maybe?
	_ = s.popFuncContext().(*FuncDeclContext)

	errs := make([]error, 0)
	if !term {
		errs = append(errs, errors.New(fmt.Sprintf("function %+v: does not terminate", f.Name)))
	}

	errs = append(errs, s.lambdaErrs...)

	return errs
}

func (s *StaticAnalysis) staticAnalysisBlockStmt(stmts []Stmt) bool {
	if len(stmts) > 0 {
		return s.staticAnalysisStmt(stmts[len(stmts)-1])
	}
	return false
}

func (s *StaticAnalysis) staticAnalysisExpr(expr Expr) (bool, bool) {
	switch n := expr.(type) {
	case *CallExpr:
		for _, arg := range n.Args {
			term, is := s.staticAnalysisExpr(arg)

			if is && !term {
				return true, true
			}
		}
	case *FuncLitExpr:
		s.pushFuncContext(&FuncLitContext{
			hasRet: false,
			f:      n,
		})
		term := s.staticAnalysisBlockStmt(n.Body)
		ctx := s.popFuncContext().(*FuncLitContext)

		if !term {
			s.lambdaErrs = append(s.lambdaErrs, errors.New(fmt.Sprintf("lambda at %v does not terminate\n", ctx.f.Loc)))
		}
		return false, false
	case *NameExpr:
		return false, false
	}
	return false, false
}

// staticAnalysisStmt returns a boolean value,
// indicating weather a statement is terminating or not
func (s *StaticAnalysis) staticAnalysisStmt(stmt Stmt) bool {
	switch n := stmt.(type) {
	case *BranchStmt:
		switch n.Op {
		case BREAK:
			ctx := s.findCtxByLabel(string(n.Label))
			ctx.pushBreak(n)
		case CONTINUE:
			//
		case DEFAULT:
			//
		case FALLTHROUGH:
			return true
		}
	case *IfStmt:
		terminates := s.staticAnalysisBlockStmt(n.Then.Body)

		var elseTerminates bool
		if len(n.Else.Body) > 0 {
			elseTerminates = s.staticAnalysisBlockStmt(n.Else.Body)
		}

		return terminates && elseTerminates
	case *ForStmt:
		s.pushContext(&ForContext{forstmt: n})
		_ = s.staticAnalysisBlockStmt(n.Body)

		ctx := s.popContext().(*ForContext)

		//there are no "break" statements referring to the "for" statement, and
		hasNoBreaks := len(ctx.breakstmts) == 0
		//the loop condition is absent, and
		hasNoCond := n.Cond == nil

		//the "for" statement does not use a range clause.
		// this one is always false because in our nodes
		// the range loop is a different data structure
		hasRange := false

		terminates := hasNoBreaks && hasNoCond && !hasRange

		if !terminates {
			return false
		}

		return true
	//for statement
	case *ReturnStmt:
		//n.Results
		return true
	case *AssignStmt:
		for _, rh := range n.Rhs {
			term, is := s.staticAnalysisExpr(rh)

			if is && !term {
				return true
			}
		}
		return false
	case *SwitchStmt:
		//there is a default case, and
		var hasDefault bool
		for _, clause := range n.Clauses {
			// nil case means default
			if clause.Cases == nil {
				hasDefault = true
				break
			}
		}

		s.pushContext(&SwitchContext{switchStmt: n})

		//the statement lists in each case,
		//including the default
		//end in a terminating statement,
		//or a possibly labeled "fallthrough" statement.
		casesTerm := true

		for _, clause := range n.Clauses {
			ct := s.staticAnalysisBlockStmt(clause.Body)
			casesTerm = casesTerm && ct
		}

		ctx := s.popContext().(*SwitchContext)
		//there are no "break" statements referring to the "switch" statement
		hasNoBreaks := len(ctx.breakstmts) == 0

		terminates := hasNoBreaks && hasDefault && casesTerm

		if !terminates {
			return false
		}

		return true
	case *PanicStmt:
		return true
	}
	return false
}

type FuncContext interface {
	isLastExprRet() bool
}

type FuncLitContext struct {
	hasRet bool
	f      *FuncLitExpr
}

func (fdc *FuncLitContext) isLastExprRet() bool {
	return fdc.hasRet
}

type FuncDeclContext struct {
	hasRet bool
	f      *FuncDecl
}

func (fdc *FuncDeclContext) isLastExprRet() bool {
	return fdc.hasRet
}

type Context interface {
	label() string
	pushBreak(breakstmt *BranchStmt)
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
