package gno

import (
	"fmt"
	"reflect"
)

func (m *Machine) doOpPrecall() {
	cx := m.PopExpr().(*CallExpr)
	//fr := m.LastFrame()
	v := m.PeekValue(1).V
	if debug {
		if v == nil {
			// this may happen due to an undefined uverse or closure value
			// (which isn't supposed to happen but may happen due to
			// incomplete initialization).
			panic("should not happen")
		}
	}
	switch fv := v.(type) {
	case *FuncValue:
		m.PopValue()
		m.PushFrameCall(cx, fv, nil)
		// continuation #2
		m.PushOp(OpCall)
	case BoundMethodValue:
		m.PopValue()
		m.PushFrameCall(cx, fv.Func, fv.Receiver)
		// continuation #2
		m.PushOp(OpCall)
	case TypeValue:
		// do not pop type yet.
		// no need for frames.
		// continuation #2
		m.PushOp(OpConvert)
		if debug {
			if len(cx.Args) != 1 {
				panic("conversion expressions only take 1 argument")
			}
		}
	case *nativeValue:
		m.PopValue()
		m.PushFrameGoNative(cx, fv)
		// continuation #2
		m.PushOp(OpCallGoNative)
	default:
		panic(fmt.Sprintf(
			"unexpected function value type %s",
			reflect.TypeOf(v).String()))
	}
	// eval args
	args := cx.Args
	for i := len(args) - 1; 0 <= i; i-- {
		m.PushExpr(args[i])
		m.PushOp(OpEval)
	}
}

var gReturnStmt = &ReturnStmt{}

func (m *Machine) doOpCall() {
	// NOTE: Frame won't be popped until the statement is complete, to
	// discard the correct number of results for func calls in ExprStmts.
	fr := m.LastFrame()
	fv := fr.Func
	ft := fr.Func.Type
	pts := ft.Params
	numParams := len(pts)
	isMethod := 0 // 1 if true
	// Create new block scope
	b := NewBlock(fr.Func.Source, fr.Func.Closure)
	m.PushBlock(b)
	// continuation
	if fv.NativeBody == nil {
		if len(ft.Results) == 0 {
			// Push final empty *ReturnStmt;
			// TODO: transform in preprocessor instead to return only
			// when necessary.
			// NOTE: m.PushOp(OpReturn) doesn't handle defers.
			m.PushStmt(gReturnStmt)
			m.PushOp(OpExec)
		}
		// Exec body.
		b.bodyStmt = bodyStmt{
			Body:      fv.Body,
			BodyLen:   len(fv.Body),
			BodyIndex: 0,
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
	// Assign receiver as first parameter, if any.
	if fr.Receiver != nil {
		pt := pts[0]
		b.Values[0] = TypedValue{
			T: pt.Type,
			V: fr.Receiver,
		}
		isMethod = 1

	}
	// Convert variadic argument.
	// TODO: more optimizations may be possible here if varg is
	// unescaping.  NOTE: this logic is somwhat duplicated for
	// doOpReturnCallDefers().
	if ft.HasVarg() {
		nvar := fr.NumArgs - isMethod - (numParams - 1)
		if fr.IsVarg {
			// do nothing, last arg type is already slice type
			// called with form fncall(?, vargs...)
			if debug {
				if nvar != 1 {
					panic("should not happen")
				}
			}
		} else {
			list := make([]TypedValue, nvar)
			copy(list, m.PopValues(nvar))
			vart := pts[numParams-1].Type.(*SliceType)
			varg := newSliceFromList(list)
			m.PushValue(TypedValue{
				T: vart,
				V: varg,
			})
		}
	}
	// Assign non-receiver parameters in forward order.
	pvs := m.PopValues(numParams - isMethod)
	for i := isMethod; i < numParams; i++ {
		// pt := pts[i]
		pv := pvs[i-isMethod]
		if debug {
			// This is how run-time untyped const
			// conversions would work, but we
			// expect the preprocessor to convert
			// these to *constExpr.
			/*
				// Convert if untyped const.
				if isUntyped(pv.T) {
					ConvertUntypedTo(&pv, pv.Type)
				}
			*/
			if isUntyped(pv.T) {
				panic("unexpected untyped const type for assign during runtime")
			}
		}
		// TODO: some more pt <> pv.Type
		// reconciliations/conversions necessary.
		b.Values[i] = pv
	}
}

func (m *Machine) doOpCallNativeBody() {
	m.LastFrame().Func.NativeBody(m)
}

func (m *Machine) doOpCallDeferNativeBody() {
	fv := m.PopValue().V.(*FuncValue)
	fv.NativeBody(m)
}

// Assumes that result values are pushed onto the Values stack.
func (m *Machine) doOpReturn() {
	fr := m.LastFrame()
	// See if we are exiting a realm boundary.
	crlm := m.Realm
	if crlm != nil {
		lrlm := fr.LastRealm
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
			crlm.FinalizeRealmTransaction()
		}
	}
	// finalize
	m.PopFrameAndReturn()
}

// Like doOpReturn, but with results from the block;
// i.e. named result vars declared in func signatures.
func (m *Machine) doOpReturnFromBlock() {
	// copy results from block
	fr := m.LastFrame()
	numParams := len(fr.Func.Type.Params)
	numResults := len(fr.Func.Type.Results)
	fblock := m.Blocks[fr.NumBlocks] // frame +1
	for i := 0; i < numResults; i++ {
		rtv := fblock.Values[i+numParams]
		m.PushValue(rtv)
	}
	// finalize
	m.PopFrameAndReturn()
}

// Before defers during return, move results to block so that
// deferred statements can refer to results with name
// expressions.
func (m *Machine) doOpReturnToBlock() {
	fr := m.LastFrame()
	numParams := len(fr.Func.Type.Params)
	numResults := len(fr.Func.Type.Results)
	fblock := m.Blocks[fr.NumBlocks] // frame +1
	results := m.PopValues(numResults)
	for i := 0; i < numResults; i++ {
		rtv := results[i]
		fblock.Values[numParams+i] = rtv
	}
}

func (m *Machine) doOpReturnCallDefers() {
	fr := m.LastFrame()
	dfr, ok := fr.PopDefer()
	if !ok {
		// Done with defers.
		m.ForcePopOp()
		return
	}
	// Call last deferred call.
	// Get block of parent function.
	fb := m.Blocks[fr.NumBlocks]
	// NOTE: the following logic is largely duplicated in doOpCall().
	// Convert if variadic argument.
	if dfr.Func != nil {
		fv := dfr.Func
		ft := fv.Type
		pts := ft.Params
		numParams := len(ft.Params)
		// Create new block scope for defer.
		b := NewBlock(fv.Source, fb)
		m.PushBlock(b)
		// continuation
		if fv.NativeBody == nil {
			// Exec body.
			b.bodyStmt = bodyStmt{
				Body:      fv.Body,
				BodyLen:   len(fv.Body),
				BodyIndex: 0,
			}
			m.PushOp(OpBody)
			m.PushStmt(b.GetBodyStmt())
		} else {
			// Call native function.
			m.PushValue(TypedValue{
				T: fv.Type,
				V: fv,
			})
			m.PushOp(OpCallDeferNativeBody)
		}
		if ft.HasVarg() {
			numArgs := len(dfr.Args)
			nvar := numArgs - (numParams - 1)
			if dfr.Source.Call.Varg {
				if debug {
					if nvar != 1 {
						panic("should not happen")
					}
				}
				// do nothing, last arg type is already slice type
				// called with form fncall(?, vargs...)
			} else {
				// convert last nvar to slice.
				vart := pts[len(pts)-1].Type.(*SliceType)
				vargs := make([]TypedValue, nvar)
				copy(vargs, dfr.Args[numArgs-nvar:numArgs])
				varg := newSliceFromList(vargs)
				dfr.Args = dfr.Args[:numArgs-nvar]
				dfr.Args = append(dfr.Args, TypedValue{
					T: vart,
					V: varg,
				})
			}
		}
		b.Values = dfr.Args
	} else if dfr.GoFunc != nil {
		fv := dfr.GoFunc
		ptvs := dfr.Args
		prvs := make([]reflect.Value, len(ptvs))
		for i := 0; i < len(prvs); i++ {
			// TODO consider when declared types can be
			// converted, e.g. fmt.Println. See GoValue.
			prvs[i] = gno2GoValue(&ptvs[i], reflect.Value{})
		}
		// call and ignore results.
		fv.Value.Call(prvs)
		// cleanup
		m.NumResults = 0
	} else {
		panic("should not happen")
	}
}

func (m *Machine) doOpDefer() {
	fr := m.LastFrame()
	ds := m.PopStmt().(*DeferStmt)
	// pop arguments
	numArgs := len(ds.Call.Args)
	args0 := m.PopValues(numArgs)
	args := make([]TypedValue, len(args0))
	copy(args, args0)
	// pop func
	ftv := m.PopValue()
	// push defer.
	// NOTE: we let type be FuncValue and value nativeValue,
	// because native funcs can't be converted to gno anyways.
	switch cv := ftv.V.(type) {
	case *FuncValue:
		// TODO what if value is nativeValue?
		fr.PushDefer(Defer{
			Func:   cv,
			Args:   args,
			Source: ds,
		})
	case BoundMethodValue:
		args2 := make([]TypedValue, len(args)+1)
		args2[0] = TypedValue{
			T: cv.Func.Type.Params[0],
			V: cv.Receiver,
		}
		copy(args2[1:], args)
		fr.PushDefer(Defer{
			Func:   cv.Func,
			Args:   args2,
			Source: ds,
		})
	case *nativeValue:
		fr.PushDefer(Defer{
			GoFunc: cv,
			Args:   args,
			Source: ds,
		})
	default:
		panic("should not happen")
	}

}
