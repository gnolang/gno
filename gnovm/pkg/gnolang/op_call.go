package gnolang

import (
	"fmt"
	"reflect"
	"strings"
)

func (m *Machine) doOpPrecall() {
	cx := m.PopExpr().(*CallExpr)
	v := m.PeekValue(1 + cx.NumArgs).V
	if debug {
		if v == nil {
			// This may happen due to an undefined uverse or
			// closure value (which isn't supposed to happen but
			// may happen due to incomplete initialization).
			panic("should not happen")
		}
	}

	switch fv := v.(type) {
	case *FuncValue:
		m.PushFrameCall(cx, fv, TypedValue{}, false)
		m.PushOp(OpCall)
	case *BoundMethodValue:
		recv := fv.Receiver
		m.PushFrameCall(cx, fv.Func, recv, false)
		m.PushOp(OpCall)
	case TypeValue:
		// Do not pop type yet.
		// No need for frames.
		xv := m.PeekValue(1)
		if cx.GetAttribute(ATTR_SHIFT_RHS) == true {
			xv.AssertNonNegative("runtime error: negative shift amount")
		}
		m.PushOp(OpConvert)
		if debug {
			if len(cx.Args) != 1 {
				panic("conversion expressions only take 1 argument")
			}
		}
	default:
		panic(fmt.Sprintf(
			"unexpected function value type %s",
			reflect.TypeOf(v).String()))
	}
}

var gReturnStmt = &ReturnStmt{}

func getFuncTypeExprFromSource(source Node) *FuncTypeExpr {
	if fd, ok := source.(*FuncDecl); ok {
		return fd.GetUnboundTypeExpr()
	} else {
		return &source.(*FuncLitExpr).Type
	}
}

func (m *Machine) doOpCall() {
	// NOTE: Frame won't be popped until the statement is complete, to
	// discard the correct number of results for func calls in ExprStmts.
	fr := m.LastFrame()
	fv := fr.Func
	fs := fv.GetSource(m.Store)
	ft := fr.Func.GetType(m.Store)
	// Create new block scope.
	pb := fr.Func.GetParent(m.Store)
	b := m.Alloc.NewBlock(fs, pb)

	// Copy *FuncValue.Captures into block
	// NOTE: addHeapCapture in preprocess ensures order.
	if len(fv.Captures) != 0 {
		if len(fv.Captures) > len(b.Values) {
			panic("should not happen, length of captured variables must not exceed the number of values")
		}
		for i := range fv.Captures {
			b.Values[len(b.Values)-len(fv.Captures)+i] = fv.Captures[i].Copy(m.Alloc)
		}
	}

	m.PushBlock(b)
	if fv.nativeBody == nil && fv.NativePkg != "" {
		// native function, unmarshaled so doesn't have nativeBody yet
		fv.nativeBody = m.Store.GetNative(fv.NativePkg, fv.NativeName)
		if fv.nativeBody == nil {
			panic(fmt.Sprintf("natively defined function (%q).%s could not be resolved", fv.NativePkg, fv.NativeName))
		}
	}
	if fv.nativeBody == nil {
		fbody := fv.GetBodyFromSource(m.Store)
		if len(ft.Results) == 0 {
			// Push final empty *ReturnStmt;
			// TODO: transform in preprocessor instead.
			// NOTE: m.PushOp(OpReturn) doesn't handle defers.
			m.PushStmt(gReturnStmt)
			m.PushOp(OpExec)
		} else {
			// NOTE: not a bound method.
			numParams := len(ft.Params)
			// Initialize return variables with default value.
			for i, rt := range ft.Results {
				dtv := defaultTypedValue(m.Alloc, rt.Type)
				ptr := b.GetPointerToInt(nil, numParams+i)
				// Write to existing heap item if result is heap defined.
				ptr.TV.AssignToBlock(dtv)
			}
		}
		// Exec body.
		b.bodyStmt = bodyStmt{
			Body:          fbody,
			BodyLen:       len(fbody),
			NextBodyIndex: -2,
		}
		m.PushOp(OpBody)
		m.PushStmt(b.GetBodyStmt())
	} else {
		// No return exprs and no defers, safe to skip OpEval.
		// NOTE: m.PushOp(OpReturn) doesn't handle defers.
		m.PushOp(OpReturn)
		// Call native function.
		// It reads the native function from the frame,
		// so this op follows (this) OpCall.
		m.PushOp(OpCallNativeBody)
	}
	// Construct arg values.
	bft := ft
	if !fr.Receiver.IsUndefined() {
		bft = ft.BoundType()
	}
	args := m.popCopyArgs(bft, fr.NumArgs, fr.IsVarg, fr.Receiver)
	// Assign parameters in forward order.
	for i, argtv := range args {
		b.Values[i].AssignToBlock(argtv)
	}
}

func (m *Machine) doOpCallNativeBody() {
	m.LastFrame().Func.nativeBody(m)
}

func (m *Machine) doOpCallDeferNativeBody() {
	fv := m.PopValue().V.(*FuncValue)
	fv.nativeBody(m)
}

// Assumes that result values are pushed onto the Values stack.
func (m *Machine) doOpReturn() {
	// Unwind stack.
	cfr := m.PopUntilLastCallFrame()

	// See if we are exiting a realm boundary.
	crlm := m.Realm
	if crlm != nil {
		if cfr.DidCross {
			// Finalize realm updates!
			// NOTE: This is a resource intensive undertaking.
			crlm.FinalizeRealmTransaction(m.Store)
		}
	}

	// Finalize
	m.PopFrameAndReturn()
}

// Like doOpReturn but first copies results to block.
func (m *Machine) doOpReturnAfterCopy() {
	// If there are named results that are heap defined,
	// need to write to those from stack before returning.
	cfr := m.MustPeekCallFrame(1)
	fv := cfr.Func
	ft := fv.GetType(m.Store)
	numParams := len(ft.Params)
	numResults := len(ft.Results)
	fblock := m.Blocks[cfr.NumBlocks] // frame +1
	results := m.PeekValues(numResults)
	for i := 0; i < numResults; i++ {
		rtv := results[i].Copy(m.Alloc)
		fblock.Values[numParams+i].AssignToBlock(rtv)
	}

	// Unwind stack.
	cfr = m.PopUntilLastCallFrame()

	// See if we are exiting a realm boundary.
	crlm := m.Realm
	if crlm != nil {
		lrlm := cfr.LastRealm
		finalize := false
		if m.NumFrames() == 1 {
			// We are exiting the machine's realm.
			finalize = true
		} else if crlm != lrlm {
			// We are changing realms or exiting a realm.
			finalize = true
		}
		if finalize {
			// Finalize realm updates!
			// NOTE: This is a resource intensive undertaking.
			crlm.FinalizeRealmTransaction(m.Store)
		}
	}

	// Finalize
	m.PopFrameAndReturn()
}

// Like doOpReturn, but with results from the block;
// i.e. named result vars declared in func signatures,
// because return was called with no return arguments.
func (m *Machine) doOpReturnFromBlock() {
	// Copy results from block.
	cfr := m.PopUntilLastCallFrame()
	fv := cfr.Func
	ft := fv.GetType(m.Store)
	numParams := len(ft.Params)
	numResults := len(ft.Results)
	fblock := m.Blocks[cfr.NumBlocks] // frame +1
	for i := range numResults {
		rtv := *fillValueTV(m.Store, &fblock.Values[i+numParams])
		m.PushValueFromBlock(rtv)
	}
	// See if we are exiting a realm boundary.
	crlm := m.Realm
	if crlm != nil {
		lrlm := cfr.LastRealm
		finalize := false
		if m.NumFrames() == 1 {
			// We are exiting the machine's realm.
			finalize = true
		} else if crlm != lrlm {
			// We are changing realms or exiting a realm.
			finalize = true
		}
		if finalize {
			// Finalize realm updates!
			// NOTE: This is a resource intensive undertaking.
			crlm.FinalizeRealmTransaction(m.Store)
		}
	}
	// finalize
	m.PopFrameAndReturn()
}

// Before defers during return, move results to block so that
// deferred statements can refer to results with name
// expressions.
func (m *Machine) doOpReturnToBlock() {
	cfr := m.MustPeekCallFrame(1)
	fv := cfr.Func
	ft := fv.GetType(m.Store)
	numParams := len(ft.Params)
	numResults := len(ft.Results)
	fblock := m.Blocks[cfr.NumBlocks] // frame +1
	results := m.PopValues(numResults)
	for i := range numResults {
		rtv := results[i]
		fblock.Values[numParams+i].AssignToBlock(rtv)
	}
}

func (m *Machine) doOpReturnCallDefers() {
	cfr := m.MustPeekCallFrame(1)
	dfr, ok := cfr.PopDefer()
	if !ok {
		// Done with defers.
		m.ForcePopOp()
		// If still in panic state pop this frame so doOpPanic2() will
		// try doOpReturnCallDefers() in the previous frame.
		if m.Exception != nil {
			m.PopFrame()
			m.PushOp(OpPanic2)
		} else {
			// Otherwise continue with the return process,
			// OpReturnFromBlock needs frame, don't pop here.
			m.PushOp(OpReturnFromBlock)
		}
		return
	}

	if dfr.Func == nil {
		m.Panic(typedString("defer called a nil function"))
		return
	}

	// Call last deferred call.
	fv := dfr.Func
	ft := fv.GetType(m.Store)
	// Push frame for defer.
	m.PushFrameCall(&dfr.Source.Call, fv, TypedValue{}, true)
	// NOTE: the following logic is largely duplicated in doOpCall().
	// Push final empty *ReturnStmt;
	// TODO: transform in preprocessor instead.
	// NOTE: m.PushOp(OpReturn) doesn't handle defers.
	m.PushStmt(gReturnStmt)
	m.PushOp(OpExec)
	// Convert if variadic argument.
	// Create new block scope for defer.
	pb := dfr.Func.GetParent(m.Store)
	b := m.Alloc.NewBlock(fv.GetSource(m.Store), pb)
	// Copy values from captures.
	if len(fv.Captures) != 0 {
		if len(fv.Captures) > len(b.Values) {
			panic("should not happen, length of captured variables must not exceed the number of values")
		}
		for i := range fv.Captures {
			b.Values[len(b.Values)-len(fv.Captures)+i] = fv.Captures[i].Copy(m.Alloc)
		}
	}
	m.PushBlock(b)
	if fv.nativeBody == nil {
		fbody := fv.GetBodyFromSource(m.Store)
		// Exec body.
		b.bodyStmt = bodyStmt{
			Body:          fbody,
			BodyLen:       len(fbody),
			NextBodyIndex: -2,
		}
		m.PushOp(OpBody)
		m.PushStmt(b.GetBodyStmt())
	} else {
		// Call native function.
		m.PushValue(TypedValue{
			T: ft,
			V: fv,
		})
		m.PushOp(OpCallDeferNativeBody)
	}
	// Assign parameters in forward order.
	for i, arg := range dfr.Args {
		// We need to define, but b was already populated
		// with new empty heap items, so AssignToBlock is
		// faster.
		b.Values[i].AssignToBlock(arg)
	}
}

// ft: the (bound) func type.
// numArgs: number of arguments provided.
// isVarg: true if called with ...varg.
// recv: receiver if bound otherwise undefined.
// Returns a slice of parameters with receiver (if any) and varg conversion.
// For bound method calls the returned slice is 1 greater than len(ft.Params).
// Constructed varg slice is allocated, but the result slice is not.
func (m *Machine) popCopyArgs(ft *FuncType, numArgs int, isVarg bool, recv TypedValue) []TypedValue {
	pts := ft.Params
	numParams := len(pts)
	isMethod := 0
	if !recv.IsUndefined() {
		isMethod = 1
	}
	args := make([]TypedValue, isMethod+numParams)
	if isMethod == 1 {
		args[0] = recv
	}
	nvar := numArgs - (numParams - 1)
	if ft.HasVarg() {
		if isVarg {
			// Do nothing special, last arg type is already slice
			// type called with form fncall(?, vargs...)
			if debug {
				if nvar != 1 {
					panic("should not happen")
				}
			}
		} else {
			// Convert variadic argument to slice argument.
			// Convert last nvar to slice.
			list := make([]TypedValue, nvar)
			m.PopCopyValues(list)
			varg := m.Alloc.NewSliceFromList(list)
			// Pop non-receiver non-varg args.
			m.PopCopyValues(args[isMethod : isMethod+numParams-1])
			// Set varg slice.
			vart := pts[numParams-1].Type.(*SliceType)
			args[isMethod+numParams-1] = TypedValue{
				T: vart,
				V: varg,
			}
			return args
		}
	}
	// Pop non-receiver args.
	m.PopCopyValues(args[isMethod:])
	return args
}

func (m *Machine) doOpDefer() {
	lb := m.LastBlock()
	cfr := m.MustPeekCallFrame(1)
	ds := m.PopStmt().(*DeferStmt)
	numArgs := len(ds.Call.Args)
	// Peek func to get type.
	ftv := m.PeekValue(numArgs + 1)
	// Push defer.
	switch cv := ftv.V.(type) {
	case *FuncValue:
		fv := cv
		args := m.popCopyArgs(
			baseOf(ftv.T).(*FuncType),
			numArgs,
			ds.Call.Varg,
			TypedValue{})
		cfr.PushDefer(Defer{
			Func:   fv,
			Args:   args,
			Source: ds,
			Parent: lb,
		})
	case *BoundMethodValue:
		fv := cv.Func
		recv := cv.Receiver
		args := m.popCopyArgs(
			baseOf(ftv.T).(*FuncType),
			numArgs,
			ds.Call.Varg,
			recv)
		cfr.PushDefer(Defer{
			Func:   fv,
			Args:   args,
			Source: ds,
			Parent: lb,
		})
	case nil:
		cfr.PushDefer(Defer{
			Func: nil,
		})
	default:
		m.Panic(typedString(fmt.Sprintf("invalid defer function call: %v", cv)))
		return
	}
	m.PopValue() // pop func
}

// XXX DEPRECATED
func (m *Machine) doOpPanic1() {
	panic("doOpPanic1 is deprecated")
	// Pop exception
	var ex TypedValue = m.PopValue().Copy(m.Alloc)
	// Panic
	m.Panic(ex)
	return
}

func (m *Machine) doOpPanic2() {
	if m.Exception == nil {
		panic("should not happen")
	}
	last := m.PopUntilLastCallFrame()
	if last == nil {
		// Build exception string just as go, separated by \n\t.
		numExceptions := m.Exception.NumExceptions()
		exs := make([]string, numExceptions)
		last := m.Exception
		for i := 0; i < numExceptions; i++ {
			exs[numExceptions-1-i] = last.Sprint(m)
			last = last.Previous
		}
		panic(UnhandledPanicError{
			Descriptor: strings.Join(exs, "\n\t"),
		})
	}
	m.PushOp(OpReturnCallDefers)
}
