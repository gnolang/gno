package gnolang

import (
	"fmt"
	"strings"
)

const maxStacktraceSize = 128

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

func (fr *Frame) IsCall() bool {
	return fr.Func != nil || fr.GoFunc != nil
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
	Func     *FuncValue   // function value
	GoFunc   *NativeValue // go function value
	Receiver TypedValue   // for methods
	Args     []TypedValue // arguments
	Source   *DeferStmt   // source
	Parent   *Block

	// PanicScope is set to the value of the Machine's PanicScope when the
	// defer is created. The PanicScope of the Machine is incremented each time
	// a panic occurs and is decremented each time a panic is recovered.
	PanicScope uint
}

//----------------------------------------
// Exception

// Exception represents a panic that originates from a gno program.
type Exception struct {
	// Value is the value passed to panic.
	Value TypedValue
	// Frames is a snapshot of the Machine frames at the time panic() is called.
	// It is used to determine whether recover() can be called (it may only be
	// called directly by a referred function) and to determine whether the
	// current execution of a deferred function is happening in the context of
	// a panic.
	Frames []*Frame

	Stacktrace Stacktrace
}

func (e Exception) Sprint(m *Machine) string {
	return e.Value.Sprint(m)
}

// UnhandledPanicError represents an error thrown when a panic is not handled in the realm.
type UnhandledPanicError struct {
	Descriptor string // Description of the unhandled panic.
}

func (e UnhandledPanicError) Error() string {
	return e.Descriptor
}

type StacktraceCall struct {
	Stmt  Stmt
	Frame *Frame
}
type Stacktrace struct {
	Calls           []StacktraceCall
	NumFramesElided int
}

func (s Stacktrace) String() string {
	var builder strings.Builder

	for i := 0; i < len(s.Calls); i++ {
		if s.NumFramesElided > 0 && i == maxStacktraceSize/2 {
			fmt.Fprintf(&builder, "...%d frame(s) elided...\n", s.NumFramesElided)
		}

		call := s.Calls[i]
		cx := call.Frame.Source.(*CallExpr)
		switch {
		case call.Frame.Func != nil && call.Frame.Func.IsNative():
			fmt.Fprintf(&builder, "%s\n", toExprTrace(cx))
			fmt.Fprintf(&builder, "    gonative:%s.%s\n", call.Frame.Func.NativePkg, call.Frame.Func.NativeName)
		case call.Frame.Func != nil:
			fmt.Fprintf(&builder, "%s\n", toExprTrace(cx))
			fmt.Fprintf(&builder, "    %s/%s:%d\n", call.Frame.Func.PkgPath, call.Frame.Func.FileName, call.Stmt.GetLine())
		case call.Frame.GoFunc != nil:
			fmt.Fprintf(&builder, "%s\n", toExprTrace(cx))
			fmt.Fprintf(&builder, "    gofunction:%s\n", call.Frame.GoFunc.Value.Type())
		default:
			panic("StacktraceCall has a non-call Frame")
		}
	}
	return builder.String()
}

func toExprTrace(ex Expr) string {
	switch ex := ex.(type) {
	case *CallExpr:
		s := make([]string, len(ex.Args))
		for i, arg := range ex.Args {
			s[i] = toExprTrace(arg)
		}
		return fmt.Sprintf("%s(%s)", toExprTrace(ex.Func), strings.Join(s, ","))
	case *BinaryExpr:
		return fmt.Sprintf("%s %s %s", toExprTrace(ex.Left), ex.Op.TokenString(), toExprTrace(ex.Right))
	case *UnaryExpr:
		return fmt.Sprintf("%s%s", ex.Op.TokenString(), toExprTrace(ex.X))
	case *SelectorExpr:
		return fmt.Sprintf("%s.%s", toExprTrace(ex.X), ex.Sel)
	case *IndexExpr:
		return fmt.Sprintf("%s[%s]", toExprTrace(ex.X), toExprTrace(ex.Index))
	case *StarExpr:
		return fmt.Sprintf("*%s", toExprTrace(ex.X))
	case *RefExpr:
		return fmt.Sprintf("&%s", toExprTrace(ex.X))
	case *CompositeLitExpr:
		lenEl := len(ex.Elts)
		if ex.Type == nil {
			return fmt.Sprintf("<elided><len=%d>", lenEl)
		}

		return fmt.Sprintf("%s<len=%d>", toExprTrace(ex.Type), lenEl)
	case *FuncLitExpr:
		return fmt.Sprintf("%s{ ... }", toExprTrace(&ex.Type))
	case *TypeAssertExpr:
		return fmt.Sprintf("%s.(%s)", toExprTrace(ex.X), toExprTrace(ex.Type))
	case *ConstExpr:
		return toConstExpTrace(ex)
	case *NameExpr, *BasicLitExpr, *SliceExpr:
		return ex.String()
	}

	return ex.String()
}

func toConstExpTrace(cte *ConstExpr) string {
	tv := cte.TypedValue

	switch bt := baseOf(tv.T).(type) {
	case PrimitiveType:
		switch bt {
		case UntypedBoolType, BoolType:
			return fmt.Sprintf("%t", tv.GetBool())
		case UntypedStringType, StringType:
			return tv.GetString()
		case IntType:
			return fmt.Sprintf("%d", tv.GetInt())
		case Int8Type:
			return fmt.Sprintf("%d", tv.GetInt8())
		case Int16Type:
			return fmt.Sprintf("%d", tv.GetInt16())
		case UntypedRuneType, Int32Type:
			return fmt.Sprintf("%d", tv.GetInt32())
		case Int64Type:
			return fmt.Sprintf("%d", tv.GetInt64())
		case UintType:
			return fmt.Sprintf("%d", tv.GetUint())
		case Uint8Type:
			return fmt.Sprintf("%d", tv.GetUint8())
		case Uint16Type:
			return fmt.Sprintf("%d", tv.GetUint16())
		case Uint32Type:
			return fmt.Sprintf("%d", tv.GetUint32())
		case Uint64Type:
			return fmt.Sprintf("%d", tv.GetUint64())
		case Float32Type:
			return fmt.Sprintf("%v", tv.GetFloat32())
		case Float64Type:
			return fmt.Sprintf("%v", tv.GetFloat64())
		}
	}

	return tv.T.String()
}
