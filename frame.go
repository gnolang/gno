package gno

import "fmt"

//----------------------------------------
// (runtime) Frame

type Frame struct {

	// general
	Label     Name // for break/continue
	Source    Node // source of frame
	NumOps    int  // number of ops in stack
	NumValues int  // number of values in stack
	NumExprs  int  // number of exprs in stack
	NumStmts  int  // number of statements in stack
	NumBlocks int  // number of blocks in stack
	BodyIndex int  // for call and for stmts

	// call frame
	Func        *FuncValue    // function value
	GoFunc      *nativeValue  // go function value
	Receiver    Value         // if bound method
	NumArgs     int           // number of arguments in call
	IsVarg      bool          // is form fncall(???, vargs...)
	Defers      []Defer       // deferred calls
	LastPackage *PackageValue // previous package context
}

func (fr Frame) String() string {
	if fr.Func != nil {
		return fmt.Sprintf("[FRAME FUNC:%v RECV:%s (%d args) %d/%d/%d/%d/%d LASTPKG:%s]",
			fr.Func,
			fr.Receiver,
			fr.NumArgs,
			fr.NumOps,
			fr.NumValues,
			fr.NumExprs,
			fr.NumStmts,
			fr.NumBlocks,
			fr.LastPackage.PkgPath)
	} else if fr.GoFunc != nil {
		return fmt.Sprintf("[FRAME GOFUNC:%v RECV:%s (%d args) %d/%d/%d/%d/%d]",
			fr.GoFunc.Value,
			fr.Receiver,
			fr.NumArgs,
			fr.NumOps,
			fr.NumValues,
			fr.NumExprs,
			fr.NumStmts,
			fr.NumBlocks)
	} else {
		return fmt.Sprintf("[FRAME LABEL: %s %d/%d/%d/%d/%d]",
			fr.Label,
			fr.NumOps,
			fr.NumValues,
			fr.NumExprs,
			fr.NumStmts,
			fr.NumBlocks)
	}
}

//----------------------------------------
// Defer

type Defer struct {
	Func   *FuncValue   // function value
	GoFunc *nativeValue // go function value
	Args   []TypedValue // arguments
}
