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
		m.incrCPU(OpCPUPrecallFunc)
		m.PushFrameCall(cx, fv, TypedValue{}, false)
		m.PushOp(OpCall)
		isCrossing := fv.IsCrossing()
		if isCrossing {
			m.PushOp(OpEnterCrossing)
		}
		if cx.IsWithCross() {
			m.installCrossingCur(cx, isCrossing, fv.PkgPath)
		}
	case *BoundMethodValue:
		m.incrCPU(OpCPUPrecallBoundMethod)
		recv := fv.Receiver
		m.PushFrameCall(cx, fv.Func, recv, false)
		m.PushOp(OpCall)
		isCrossing := fv.IsCrossing()
		if isCrossing {
			m.PushOp(OpEnterCrossing)
		}
		if cx.IsWithCross() {
			m.installCrossingCur(cx, isCrossing, fv.Func.PkgPath)
		}
	case TypeValue:
		m.incrCPU(OpCPUPrecallTypeConv)
		// Do not pop type yet.
		// No need for frames.
		xv := m.PeekValue(1)
		// When the preprocessor wraps a shift RHS in uint(),
		// it sets ATTR_SHIFT_RHS so we can reject negative
		// values before the conversion.
		if cx.GetAttribute(ATTR_SHIFT_RHS) == true {
			if xv.Sign() < 0 {
				m.Panic(typedString(fmt.Sprintf("runtime error: negative shift amount: %v", xv)))
			}
		}
		m.PushOp(OpConvert)
		if debug {
			if len(cx.Args) != 1 {
				panic("conversion expressions only take 1 argument")
			}
		}
	default:
		// e.g. when *CallExpr.NumArgs is wrong.
		panic(fmt.Sprintf(
			"unexpected function value type %s %v",
			reflect.TypeOf(v).String(), v))
	}
}

// installCrossingCur replaces the cross-arg slot with a freshly minted
// cur realm and records it on the just-pushed frame.
//
// Two paths, distinguished by what Args[0] evaluated to on the value
// stack:
//
//   - Compiler-synthesized `.origin` (MsgCall chain root) or the
//     legacy `cross1` migration sentinel: preprocessor replaced
//     Args[0] with a constNil, so its stack slot is undefined. The
//     new cur's prev comes from m.callingCurOrOrigin() — a frame
//     walk that finds the topmost crossing frame's Cur (or the
//     per-tx origin).
//
//   - Explicit `cross(rlm)`: Args[0] is the inner cross CallExpr. At
//     runtime cross's native body validates IsCurrent-strict on rlm
//     and pushes it back unchanged, so the stack slot holds the
//     validated realm value. We use it directly as the new cur's
//     prev — no second IsCurrent check needed here.
func (m *Machine) installCrossingCur(cx *CallExpr, isCrossing bool, pkgPath string) {
	if !isCrossing {
		panic("non-crossing function in cross call")
	}
	argtv := m.PeekValue(cx.NumArgs)
	var prev TypedValue
	if argtv.IsUndefined() {
		// .origin / cross1 path.
		prev = m.callingCurOrOrigin()
	} else {
		// cross(rlm) form: argtv is the realm value cross pushed
		// back after its own IsCurrent-strict check.
		prev = *argtv
	}
	crlm := NewConcreteRealm(m.Alloc, pkgPath, prev)
	argtv.Assign(m.Alloc, crlm, false)
	m.LastFrame().Cur = crlm
}

// curUsesPreprocessOrigin reports whether tv is a captured realm whose
// prev field is the preprocess-time placeholder origin (addr=""). The
// placeholder is baked into the `.cur` ConstExpr by preprocess.go for
// main(cur realm) / init(cur realm); at runtime we detect it so the
// doOpCall fix can swap in the per-tx origin carrying the real
// OriginCaller addr. Fully structural — survives AST persistence,
// because the already-swapped per-tx origin always has a non-empty
// addr and is naturally suppressed.
func (m *Machine) curUsesPreprocessOrigin(tv *TypedValue) bool {
	sv := derefRealmStruct(tv)
	if sv == nil || len(sv.Fields) < 3 {
		return false
	}
	prev, ok := sv.Fields[2].V.(PointerValue)
	if !ok {
		return false
	}
	hiv, _ := prev.Base.(*HeapItemValue)
	if !isOriginRealmHIV(hiv) {
		return false
	}
	prevSV := hiv.Value.V.(*StructValue)
	return prevSV.Fields[0].GetString() == ""
}

// callingCurOrOrigin returns the captured cur TypedValue of the most recent
// crossing call frame on the stack, or the per-tx origin realm when none
// exists. The origin realm mirrors runtime.PreviousRealm() at the chain
// root (addr=OriginCaller, pkgPath="").
//
// Skips:
//   - non-call frames (loops, blocks).
//   - non-crossing call frames (no WithCross or DidCrossing).
//   - frames whose Cur has not been set yet (the just-pushed frame at the
//     top during doOpPrecall — its Cur is assigned right after we return).
//
// The walk is intentionally simpler than execctx.GetRealm: we only need
// the immediate captured prev, not a height-based ancestor selection.
func (m *Machine) callingCurOrOrigin() TypedValue {
	for i := len(m.Frames) - 1; i >= 0; i-- {
		fr := &m.Frames[i]
		if !fr.IsCall() {
			continue
		}
		if !(fr.WithCross || fr.DidCrossing) {
			continue
		}
		if fr.Cur.T == nil {
			continue
		}
		return fr.Cur
	}
	return buildOriginRealm(m)
}

var gReturnStmt = &ReturnStmt{}

// crossingFromTestFile reports whether fv is a crossing function declared
// in a *_test.gno file. Used by doOpEnterCrossing to allow `p/` test files
// to declare and call crossing functions (production p/ code still can't —
// see crossingAllowed in preprocess.go).
//
// fv.FileName is populated for top-level FuncDecls but empty for function
// literals (closures). For literals we walk to the Source AST node and
// read its Location.File.
func crossingFromTestFile(fv *FuncValue) bool {
	if strings.HasSuffix(fv.FileName, "_test.gno") {
		return true
	}
	if fv.Source == nil {
		return false
	}
	return strings.HasSuffix(fv.Source.GetLocation().File, "_test.gno")
}

// This used to be the crossing() uverse function.
// It should be run once upon calling every crossing function,
// whether or not it was cross-called.
func (m *Machine) doOpEnterCrossing() {
	// Sanity check.
	fr1 := m.PeekCallFrame(1) // fr1.LastPackage called to create fr1.
	if !m.Package.IsRealm() {
		// Allow crossing functions declared in *_test.gno files so p/
		// package tests can declare `TestXxx(cur realm, t *testing.T)`
		// and drive migrated methods. Also allow the top-level `main`
		// in ephemeral /e/ run packages so MsgRun scripts can opt into
		// `func main(cur realm)`. Preprocess already enforces both
		// carve-outs; this runtime check is the matching gate.
		if !IsEphemeralPath(m.Package.PkgPath) &&
			(fr1 == nil || fr1.Func == nil || !crossingFromTestFile(fr1.Func)) {
			panic("expected crossing function in a realm package")
		}
	}

	// Verify prior fr.WithCross or fr.DidCrossing.
	// NOTE: fr.WithCross may or may not be true,
	// crossing() (which sets fr.DidCrossing) can be
	// stacked.
	//
	// PERF: O(n^2) in call-stack depth. PeekCallFrame(i) restarts from the
	// top of m.Frames every iteration; outer loop runs until the first
	// crossing ancestor, visiting 1+2+...+D = O(D^2) frames. Fix is to
	// walk m.Frames once with a cursor, yielding each call frame in
	// order, which makes the handler O(D). If/when that lands, drop
	// OpCPUSlopeEnterCrossingQuad and switch this handler to a linear
	// per-depth charge (OpCPUSlopeEnterCrossing * depth).
	for i := 1; ; i++ {
		fri := m.PeekCallFrame(i) // see PERF note above.
		if 1 < i && fri == nil {
			// For stage add, meaning init() AND
			// global var decls inherit a faux
			// frame of index -1 which crossed from
			// the package deployer.
			// For stage run, main() does the same,
			// so main() can be crossing or not, it
			// doesn't matter. This applies for
			// MsgRun() as well as tests. MsgCall()
			// runs like cross(fn)(...) which
			// meains fri.WithCross would have been
			// found below.
			m.incrCPU(int64(i) * int64(i) * OpCPUSlopeEnterCrossingQuad / 10)
			fr1.SetDidCrossing()
			return
		}
		if fri.WithCross || fri.DidCrossing {
			// NOTE: fri.DidCrossing implies
			// everything under it is also valid.
			// fri.DidCrossing && !fri.WithCross
			// can happen with an implicit switch.
			m.incrCPU(int64(i) * int64(i) * OpCPUSlopeEnterCrossingQuad / 10)
			fr1.SetDidCrossing()
			return
		}
		// Neither fri.WithCross nor fri.DidCrossing, yet
		// Realm already switched implicitly.
		if fri.LastRealm != m.Realm {
			panic("crossing could not find corresponding cross(fn)(...) call")
		}
	}
	// NOTE: this loop must never exit without setting fr1.DidCrossing or panicking.
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
	m.incrCPU(OpCPUSlopeCallParam*int64(len(ft.Params)) +
		OpCPUSlopeCallCapture*int64(len(fv.Captures)))
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
	// Inherit fr.Cur from the block for crossing functions entered without
	// a cross-call (doOpPrecall sets fr.Cur only for cross-call entries).
	// Bound-method receivers occupy block[0], so cur is at block[1] for
	// methods and block[0] otherwise. If the inherited cur's prev is the
	// preprocess-time placeholder (built without OriginCaller knowledge),
	// rebuild with the per-tx origin so cur.Previous() carries the EOA
	// addr that runtime.PreviousRealm() surfaces.
	//
	// Skip natives: uverse helpers like cross(rlm realm) realm satisfy
	// ft.IsCrossing() at runtime because their generic X param resolves
	// to realm, but they were never registered with a preprocess-time
	// origin placeholder. Inheriting + rebuilding here would replace
	// the caller-supplied rlm with a fresh uverse-pkgPath realm and
	// trip cross's IsCurrent-strict check.
	curIdx := 0
	if !fr.Receiver.IsUndefined() {
		curIdx = 1
	}
	if ft.IsCrossing() && fv.nativeBody == nil && fr.Cur.T == nil && len(b.Values) > curIdx {
		// Unwrap a heap-promoted slot: when cur is captured by a nested
		// closure, the preprocessor heap-promotes its block slot, so
		// b.Values[curIdx] is a *HeapItemValue wrapper rather than the
		// realm PointerValue itself. Storing the wrapper on fr.Cur would
		// hide the underlying HIV from realmHIV (used by .grealm.IsCurrent
		// and cross's strict check), making the frame invisible to the
		// HIV-identity walk. Deref to keep fr.Cur shaped like a normal
		// PointerValue+HIV realm. The block slot itself must stay heap-
		// wrapped (the closure-capture preprocess check at doOpFuncLit
		// expects ptr.TV.T to be heapItemType), so the preprocess-origin
		// rebuild writes into the HIV's Value field rather than replacing
		// the slot entry.
		bvSlot := &b.Values[curIdx]
		hiv, isHeap := bvSlot.V.(*HeapItemValue)
		if isHeap {
			fr.Cur = hiv.Value
		} else {
			fr.Cur = *bvSlot
		}
		if m.curUsesPreprocessOrigin(&fr.Cur) {
			fresh := NewConcreteRealm(m.Alloc, fv.PkgPath, buildOriginRealm(m))
			fr.Cur = fresh
			if isHeap {
				hiv.Value = fresh
			} else {
				*bvSlot = fresh
			}
		}
	}
}

func (m *Machine) doOpCallNativeBody() {
	fv := m.LastFrame().Func
	gi := m.chargeNativeGas(fv)
	fv.nativeBody(m)
	if gi != nil {
		m.chargeNativeGasPost(gi)
	}
}

func (m *Machine) doOpCallDeferNativeBody() {
	fv := m.PopValue().V.(*FuncValue)
	gi := m.chargeNativeGas(fv)
	fv.nativeBody(m)
	if gi != nil {
		m.chargeNativeGasPost(gi)
	}
}

// Used by return and panic operation handlers.
// Must finalize for returns, and must abort for panics.
func (m *Machine) isRealmBoundary(cfr *Frame) bool {
	// Explicit cross-call always marks a realm boundary, regardless
	// of whether m.Realm is tracked. /p/ test code that wraps calls
	// in `func(cur){...}(cross(cur))` relies on this so panics
	// propagating back through the cross frame route to revive(),
	// not all the way up — even though /p/ packages have no Realm
	// (pre-Phase-3 / post-Phase-3-revert state). Pulled out of the
	// `crlm != nil` guard so it fires on m.Realm==nil too.
	if cfr.WithCross {
		return true
	}
	crlm := m.Realm
	if crlm != nil {
		prlm := cfr.LastRealm
		if crlm != prlm {
			// .WithCross was already handled;
			// This is for implicitly crossed
			// borrow-realms, the storage realm
			// of a method's receiver.
			return true
		} else if m.NumFrames() == 1 {
			// We are exiting the machine's realm.
			if m.Stage == StageAdd {
				// Unless StageAdd, where functions are called
				// during var decls. e.g.
				// // in _test.gno
				// var (
				//   x = struct{}{}
				//   alice = testutils.TestAddress("alice")
				// )
				// Since the package is real (created before
				// RunFiles() w/ _test.gno files), x = 1 will
				// pv.DidUpdate and mark pv.Block as dirty, and
				// when returning from frame 1 TestAddress
				// there will be an unexpected unreal object in
				// pv.Block. RunFiles() will finalize manually
				// after.
				return false
			}
			return true
		}
	}
	return false
}

// Finalize realm updates if realm boundary.
// NOTE: resource intensive
func (m *Machine) maybeFinalize(cfr *Frame) {
	if m.isRealmBoundary(cfr) && m.Realm != nil {
		// m.Realm==nil only happens for /p/ and stdlib (no real realm),
		// where there's nothing to finalize even though isRealmBoundary
		// reports the cross-call frame as a boundary for panic-routing.
		m.Realm.FinalizeRealmTransaction(m.Store)
	}
}

// Assumes that result values are pushed onto the Values stack.
func (m *Machine) doOpReturn() {
	// Unwind stack.
	cfr := m.PopUntilLastCallFrame()
	// Finalize if exiting realm boundary.
	m.maybeFinalize(cfr)
	// Reset to before frame.
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
	// Finalize if exiting realm boundary.
	m.maybeFinalize(cfr)
	// Reset to before frame.
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

	// Finalize if exiting realm boundary.
	m.maybeFinalize(cfr)
	// Reset to before frame.
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
			// If crossing a realm boundary find the revive frame
			// for transaction revival.
			if m.isRealmBoundary(cfr) {
				cfr := m.PopUntilLastReviveFrame()
				if cfr == nil {
					// or abort the transaction.
					panic(m.makeUnhandledPanicError())
				}
				m.PopFrameAndReturn()
				// assign exception as return of revive().
				resx := m.PeekValue(1)
				resx.Assign(m.Alloc, m.Exception.Value, false)
				m.Exception = nil // reset
				return
			}
			// Handle panic by calling OpReturnCallDefers on
			// the next (last) call frame)
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
		m.pushPanic(typedString("defer called a nil function"))
		return
	}

	// Call last deferred call.
	fv := dfr.Func
	ft := fv.GetType(m.Store)
	// Push frame for defer.
	if dfr.IsBoundMethod {
		// args[0] is the receiver, per popCopyArgs bound-method invariant.
		m.PushFrameCall(&dfr.Source.Call, fv, dfr.Args[0], true)
	} else {
		m.PushFrameCall(&dfr.Source.Call, fv, TypedValue{}, true)
	}
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
			Func:          fv,
			IsBoundMethod: true,
			Args:          args,
			Source:        ds,
			Parent:        lb,
		})
	case nil:
		cfr.PushDefer(Defer{
			Func: nil,
		})
	default:
		m.pushPanic(typedString(fmt.Sprintf("invalid defer function call: %v", cv)))
		return
	}
	m.PopValue() // pop func
}

// Build exception string just as go, separated by \n\t.
// TODO: deprecate UnhandledPanicError and just use the Exception.
// (use a field to mark transaction abort)
func (m *Machine) makeUnhandledPanicError() UnhandledPanicError {
	if m.BoundedPanicRender {
		return UnhandledPanicError{
			Descriptor: BoundedSprintException(m.Exception, m, BoundedRenderBytes),
		}
	}
	numExceptions := m.Exception.NumExceptions()
	exs := make([]string, numExceptions)
	last := m.Exception
	for i := 0; i < numExceptions; i++ {
		exs[numExceptions-1-i] = last.Sprint(m)
		last = last.Previous
	}
	return UnhandledPanicError{
		Descriptor: strings.Join(exs, "\n\t"),
	}
}

func (m *Machine) doOpPanic2() {
	if m.Exception == nil {
		panic("should not happen")
	}
	cfr := m.PopUntilLastCallFrame()
	if cfr == nil {
		// If we can't find a call frame, we're in a corrupted state.
		// This can happen during init functions with realm calls.
		// Return the original exception as an unhandled panic.
		panic(m.makeUnhandledPanicError())
	}
	m.PushOp(OpReturnCallDefers)
}
