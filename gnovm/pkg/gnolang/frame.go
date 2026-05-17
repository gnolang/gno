package gnolang

import (
	"fmt"
	"math"
	"runtime"
	"strconv"
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
	DidCrossing   bool          // true if crossing() was called.
	Defers        []Defer       // deferred calls
	IsDefer       bool          // was func defer called
	IsRevive      bool          // calling revive()
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
			fr.DidCrossing,
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
	lenDefers := len(fr.Defers)
	if lenDefers > 0 {
		ok = true
		res = fr.Defers[lenDefers-1]
		fr.Defers = fr.Defers[:lenDefers-1]
	}
	return
}

func (fr *Frame) SetDidCrossing() {
	if fr.DidCrossing {
		panic("fr.DidCrossing already set")
	}
	fr.DidCrossing = true
}

func (fr *Frame) SetIsRevive() {
	if fr.IsRevive {
		panic("fr.IsRevive already set")
	}
	fr.IsRevive = true
}

//----------------------------------------
// Defer

type Defer struct {
	Func          *FuncValue   // function value
	IsBoundMethod bool         // if true, args[0] is receiver
	Args          []TypedValue // arguments
	Source        *DeferStmt   // source
	Parent        *Block
}

type StacktraceCall struct {
	CallExpr *CallExpr
	IsDefer  bool
	FuncLoc  Location // func loc in which CallExpr is declared
	// FuncName is a pre-rendered display name including receiver
	// type prefix for methods (e.g. "Counter.Inc",
	// "(*pkg.Counter).Inc"); empty for anonymous functions. Used
	// only by the bounded stacktrace renderer
	// (BoundedStacktrace) — Stacktrace.String() retains the
	// existing toExprTrace-based output.
	FuncName string
}
type Stacktrace struct {
	Calls           []StacktraceCall
	NumFramesElided int
	LastLine        int
}

func (s Stacktrace) IsZero() bool {
	return s.Calls == nil && s.NumFramesElided == 0 && s.LastLine == 0
}

const stackRecursiveTraceLimit = 100

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
		fmt.Fprintf(&builder, "%s\n", toExprTrace(call.CallExpr, stackRecursiveTraceLimit))
		if line == -1 { // special line for native
			fmt.Fprintf(&builder, "    gonative:%s/%s\n", call.FuncLoc.PkgPath, call.FuncLoc.File)
		} else {
			fmt.Fprintf(&builder, "    %s/%s:%d\n", call.FuncLoc.PkgPath, call.FuncLoc.File, line)
		}
	}
	return builder.String()
}

func toExprTrace(ex Expr, limit int) string {
	if limit < 0 {
		return "..." // Reached recursivity limit.
	}
	switch ex := ex.(type) {
	case *CallExpr:
		s := make([]string, len(ex.Args))
		for i, arg := range ex.Args {
			s[i] = toExprTrace(arg, limit-1)
		}
		return fmt.Sprintf("%s(%s)", toExprTrace(ex.Func, limit-1), strings.Join(s, ","))
	case *BinaryExpr:
		return fmt.Sprintf("%s %s %s", toExprTrace(ex.Left, limit-1), ex.Op.TokenString(), toExprTrace(ex.Right, limit-1))
	case *UnaryExpr:
		return fmt.Sprintf("%s%s", ex.Op.TokenString(), toExprTrace(ex.X, limit-1))
	case *SelectorExpr:
		return fmt.Sprintf("%s.%s", toExprTrace(ex.X, limit-1), ex.Sel)
	case *IndexExpr:
		return fmt.Sprintf("%s[%s]", toExprTrace(ex.X, limit-1), toExprTrace(ex.Index, limit-1))
	case *StarExpr:
		return fmt.Sprintf("*%s", toExprTrace(ex.X, limit-1))
	case *RefExpr:
		return fmt.Sprintf("&%s", toExprTrace(ex.X, limit-1))
	case *CompositeLitExpr:
		lenEl := len(ex.Elts)
		if ex.Type == nil {
			return fmt.Sprintf("<elided><len=%d>", lenEl)
		}

		return fmt.Sprintf("%s<len=%d>", toExprTrace(ex.Type, limit-1), lenEl)
	case *FuncLitExpr:
		return fmt.Sprintf("%s{ ... }", toExprTrace(&ex.Type, limit-1))
	case *TypeAssertExpr:
		return fmt.Sprintf("%s.(%s)", toExprTrace(ex.X, limit-1), toExprTrace(ex.Type, limit-1))
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

	if tv.IsTypedNil() {
		return "typed-nil"
	} else if tv.IsUndefined() {
		return "undefined"
	} else {
		return tv.V.String()
	}
}

//----------------------------------------
// Exception

// Exception represents a panic that originates from a gno program.
// Constructed at the raise site via NewException, which captures the
// VM-internal Go call chain in GoStack; the helper returns the value to
// its doOp* caller, which pushes it cooperatively via m.pushPanicException.
// No code path panics with *Exception — recovery infrastructure is unused
// for VM signals, retaining its purpose only for true Go runtime bugs.
//
//nolint:errname // predates the error interface; renaming would touch the VM core
type Exception struct {
	Value      TypedValue
	Stacktrace Stacktrace
	Previous   *Exception
	Next       *Exception

	// Abort marks the unhandled-panic terminal state (defers exhausted
	// without recover). markAbort sets it; Run surfaces the *Exception
	// directly to the outer recoverer instead of re-entering pushPanic.
	Abort bool
	// Descriptor caches the joined-chain message for external recoverers
	// that lack a *Machine for Sprint.
	Descriptor string
	// GoStack is the VM-internal Go call chain at the raise site, captured
	// by NewException via runtime.Callers. Formatted "funcName\n\tfile:line"
	// per frame, joined by '\n', stopping at (*Machine).Run.
	GoStack string
}

// NewException constructs an Exception and captures the Go call chain
// from the raise site up to (*Machine).Run. Helpers without *Machine
// access (values.go, alloc.go) use this in place of panic(&Exception{...}).
func NewException(value TypedValue) *Exception {
	ex := &Exception{Value: value}
	ex.GoStack = captureExceptionStack(2)
	return ex
}

func captureExceptionStack(skip int) string {
	var pcs [32]uintptr
	n := runtime.Callers(skip+1, pcs[:])
	if n == 0 {
		return ""
	}
	frames := runtime.CallersFrames(pcs[:n])
	var sb strings.Builder
	for {
		f, more := frames.Next()
		if !strings.HasPrefix(f.Function, "runtime.") {
			sb.WriteString(f.Function)
			sb.WriteString("\n\t")
			sb.WriteString(TrimOriginFile(f.File))
			sb.WriteByte(':')
			sb.WriteString(strconv.Itoa(f.Line))
			sb.WriteByte('\n')
			if strings.HasSuffix(f.Function, ".(*Machine).Run") {
				break
			}
		}
		if !more {
			break
		}
	}
	return sb.String()
}

func TrimOriginFile(file string) string {
	for _, marker := range []string{"/gnovm/", "/src/"} {
		if i := strings.Index(file, marker); i >= 0 {
			return file[i+1:]
		}
	}
	if i := strings.LastIndexByte(file, '/'); i >= 0 {
		return file[i+1:]
	}
	return file
}

// Error makes *Exception satisfy `error` for recoverer ergonomics.
// Returns Descriptor when Abort is set (joined-chain format matching
// legacy UnhandledPanicError.Error()), else the head exception's value.
func (e *Exception) Error() string {
	if e.Abort && e.Descriptor != "" {
		return e.Descriptor
	}
	return e.Value.String()
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
