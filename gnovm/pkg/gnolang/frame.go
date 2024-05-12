package gnolang

import (
	"fmt"
)

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

	// call frame
	Func        *FuncValue    // function value
	GoFunc      *NativeValue  // go function value
	Receiver    TypedValue    // if bound method
	NumArgs     int           // number of arguments in call
	IsVarg      bool          // is form fncall(???, vargs...)
	Defers      []Defer       // deferred calls
	LastPackage *PackageValue // previous package context
	LastRealm   *Realm        // previous realm context

	Popped bool // true if frame has been popped
}

func (fr Frame) String() string {
	if fr.Func != nil {
		return fmt.Sprintf("[FRAME FUNC:%v RECV:%s (%d args) %d/%d/%d/%d/%d LASTPKG:%s LASTRLM:%v]",
			fr.Func,
			fr.Receiver,
			fr.NumArgs,
			fr.NumOps,
			fr.NumValues,
			fr.NumExprs,
			fr.NumStmts,
			fr.NumBlocks,
			fr.LastPackage.PkgPath,
			fr.LastRealm)
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

func (fr *Frame) PushDefer(dfr Defer) {
	fr.Defers = append(fr.Defers, dfr)
}

func (fr *Frame) PopDefer() (res Defer, ok bool) {
	if len(fr.Defers) > 0 {
		ok = true
		res = fr.Defers[len(fr.Defers)-1]
		fr.Defers = fr.Defers[:len(fr.Defers)-1]
	}
	return
}

//----------------------------------------
// Defer

type Defer struct {
	Func   *FuncValue   // function value
	GoFunc *NativeValue // go function value
	Args   []TypedValue // arguments
	Source *DeferStmt   // source
	Parent *Block

	// PanicScope is set to the value of the Machine's PanicScope when the
	// defer is created. The PanicScope of the Machine is incremented each time
	// a panic occurs and is decremented each time a panic is recovered.
	PanicScope uint
}
