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

func (m *Machine) doOpCall() {
	// NOTE: Frame won't be popped until the statement is complete, to
	// discard the correct number of results for func calls in ExprStmts.
	fr := m.LastFrame()
	fv := fr.Func
	ft := fr.Func.Type
	pts := ft.Params
	numParams := len(pts)
	isMethod := 0 // 1 if true
	// continuation
	if fv.NativeBody == nil {
		// If a function has return values, this is not necessary.
		// TODO: transform in preprocessor instead.
		if len(ft.Results) == 0 {
			// no return exprs, safe to skip OpEval.
			m.PushOp(OpReturn)
		}
		// Queue body statements.
		for i := len(fv.Body) - 1; 0 <= i; i-- {
			s := fv.Body[i]
			m.PushStmt(s)
			m.PushOp(OpExec)
		}
	} else {
		// No return exprs, safe to skip OpEval.
		m.PushOp(OpReturn)
		// Call native function.
		// It reads the native function from the frame,
		// so this op follows (this) OpCall.
		m.PushOp(OpCallNativeBody)
	}
	// Create new block scope
	b := NewBlock(fr.Func.Source, fr.Func.Closure)
	m.PushBlock(b)
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
	// TODO: more optimizations may be possible here if varg is unescaping.
	if ft.HasVarg() {
		nvar := fr.NumArgs - isMethod - (numParams - 1)
		if nvar == 1 && fr.IsVarg {
			// do nothing, last arg type is already slice type
			// called with form fncall(?, vargs...)
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

// Like doOpReturn after pushing results to values stack.
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

// Before defers during return, pop results to block so that
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
	panic("not yet implemented")
	// TODO sticky, so force pop once
	// deferred statements are gone.
}
