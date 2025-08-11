package gnolang

import (
	"fmt"
	"math"
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
	Func          *FuncValue    // function value
	Receiver      TypedValue    // if bound method
	NumArgs       int           // number of arguments in call
	IsVarg        bool          // is form fncall(???, vargs...)
	LastPackage   *PackageValue // previous frame's package
	LastRealm     *Realm        // previous frame's realm
	WithCross     bool          // true if called like cross(fn)(...). expects crossing() after.
	DidCross      bool          // true if crossing() was called.
	Defers        []Defer       // deferred calls
	IsDefer       bool          // was func defer called
	LastException *Exception    // previous m.exception

	// test info
	TestOverridden bool // bool if overridden by test SetContext.
}

func (fr Frame) String() string {
	if fr.Func != nil {
		return fmt.Sprintf("[FRAME FUNC:%v RECV:%s (%d args) %d/%d/%d/%d/%d LASTPKG:%s LASTRLM:%v WSW:%v DSW:%v ISDEFER:%v LASTEX:%v]",
			fr.Func,
			fr.Receiver,
			fr.NumArgs,
			fr.NumOps,
			fr.NumValues,
			fr.NumExprs,
			fr.NumStmts,
			fr.NumBlocks,
			fr.LastPackage.PkgPath,
			fr.LastRealm,
			fr.WithCross,
			fr.DidCross,
			fr.IsDefer,
			fr.LastException,
		)
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
	return fr.Func != nil
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

func (fr *Frame) SetWithCross() {
	if fr.WithCross {
		panic("fr.WithCross already set")
	}
	fr.WithCross = true
}

func (fr *Frame) SetDidCross() {
	if fr.DidCross {
		panic("fr.DidCross already set")
	}
	fr.DidCross = true
}

//----------------------------------------
// Defer

type Defer struct {
	Func   *FuncValue   // function value
	Args   []TypedValue // arguments
	Source *DeferStmt   // source
	Parent *Block
}

type StacktraceCall struct {
	CallExpr *CallExpr
	IsDefer  bool
	FuncLoc  Location // func loc in which CallExpr is declared
}
type Stacktrace struct {
	Calls           []StacktraceCall
	NumFramesElided int
	LastLine        int
}

func (s Stacktrace) IsZero() bool {
	return s.Calls == nil && s.NumFramesElided == 0 && s.LastLine == 0
}

func (s Stacktrace) String() string {
	var builder strings.Builder

	for i, call := range s.Calls {
		if s.NumFramesElided > 0 && i == maxStacktraceSize/2 {
			fmt.Fprintf(&builder, "...%d frame(s) elided...\n", s.NumFramesElided)
		}
		var line int
		if i == 0 {
			line = s.LastLine
		} else {
			line = s.Calls[i-1].CallExpr.GetLine()
		}

		if call.IsDefer {
			fmt.Fprintf(&builder, "defer ")
		}
		fmt.Fprintf(&builder, "%s\n", toExprTrace(call.CallExpr))
		if line == -1 { // special line for native
			fmt.Fprintf(&builder, "    gonative:%s/%s\n", call.FuncLoc.PkgPath, call.FuncLoc.File)
		} else {
			fmt.Fprintf(&builder, "    %s/%s:%d\n", call.FuncLoc.PkgPath, call.FuncLoc.File, line)
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
			return fmt.Sprintf("%v", math.Float32frombits(tv.GetFloat32()))
		case Float64Type:
			return fmt.Sprintf("%v", math.Float64frombits(tv.GetFloat64()))
		}
	}

	return tv.V.String()
}

//----------------------------------------
// Exception

// Exception represents a panic that originates from a gno program.
type Exception struct {
	Value      TypedValue
	Stacktrace Stacktrace
	Previous   *Exception
	Next       *Exception
}

func (e *Exception) StringWithStacktrace(m *Machine) string {
	return "panic: " + e.Value.Sprint(m) + "\n" + e.Stacktrace.String()
}

func (e *Exception) Sprint(m *Machine) string {
	res := e.Value.Sprint(m)
	return res
}

func (e *Exception) NumExceptions() int {
	if e == nil {
		return 0
	}
	num := 1
	for ; e.Previous != nil; e = e.Previous {
		num++
	}
	return num
}

func (e *Exception) WithPrevious(e2 *Exception) *Exception {
	if e == nil {
		panic("missing exception")
	}
	if e.Previous != nil {
		panic("previous exception already exists")
	}
	if e2 == nil {
		return e
	}
	e.Previous = e2
	e2.Next = e
	return e
}

// UnhandledPanicError represents an error thrown when a panic is not handled in the realm.
type UnhandledPanicError struct {
	Descriptor string // Description of the unhandled panic.
}

func (e UnhandledPanicError) Error() string {
	return e.Descriptor
}
