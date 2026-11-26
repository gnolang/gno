package gnolang

// Rule R2a (pr6495_cross_to_write.md): a non-crossing function value called BY
// VALUE runs read-only, unless it is a literal lent in place to a live call.
// The crossing-call marker lives in memory only: never stored, encoded or hashed.

// calledByValue reports whether cx reaches its callee as a value rather than by
// its declared name. ATTR_NAMED_CALLEE is the static answer: a realm's own
// variable holding a function is not named, so an alias cannot dodge the rule.
func calledByValue(cx *CallExpr) bool {
	if cx.GetAttribute(ATTR_NAMED_CALLEE) == true || cx.GetAttribute(ATTR_LOCAL_HELPER) == true {
		return false
	}
	switch fx := cx.Func.(type) {
	case *NameExpr:
		return fx.Path.Type == VPBlock
	case *SelectorExpr:
		return !fx.Path.IsMethod()
	case *IndexExpr, *StarExpr:
		return true
	case *FuncLitExpr:
		// Called where it is written: made here, run once in the mode of the
		// code around it, never held by anyone else.
		return false
	}
	return true
}

// restrictByValueCall applies R2a to a call that would otherwise run writable:
// a non-crossing callee reached by value runs read-only unless its loan is live.
func (m *Machine) restrictByValueCall(cx *CallExpr, fv *FuncValue, withCross bool) {
	if withCross || fv.IsCrossing() || m.Realm == nil || !calledByValue(cx) || m.lendLive(fv) {
		return
	}
	if !m.testRunnerCall() {
		m.enterReadOnlyByS3(m.Realm)
	}
}

// testRunnerCall reports whether the test runner itself made the call just
// entered (TestXxx(t), an example): every call frame below it is "testing"
// code. A value a test hands to a testing helper has the test's frame below.
func (m *Machine) testRunnerCall() bool {
	n := len(m.Frames) - 1
	if n < 1 || m.Frames[n].LastPackage == nil || m.Frames[n].LastPackage.PkgPath != "testing" {
		return false
	}
	for i := 0; i < n; i++ {
		if fv := m.Frames[i].Func; fv != nil && fv.PkgPath != "testing" {
			return false
		}
	}
	return true
}

// shareHandedArgs shares the arguments of a call that left the caller's realm
// or entered read-only code (a stranger's library, or any function value).
func (m *Machine) shareHandedArgs(fr *Frame, args []TypedValue) {
	if len(args) == 0 {
		return
	}
	if !m.viewOnly || fr.LastViewOnly {
		if now, _ := m.Alloc.holdRealm(); now == fr.callerHold() {
			return
		}
	}
	for i := range args {
		m.Alloc.shareSlot(&args[i])
	}
}

// unmarked returns v, or a copy with no maker when v is a function value that
// still has one, so what is stored or handed out carries no maker. A value
// already kept is wiped in place instead, so no second object shares its ID;
// a marked value was made outside any read-only stretch, so the fresh
// ObjectInfo drops no scratch number.
func (alloc *Allocator) unmarked(v Value) Value {
	switch cv := v.(type) {
	case *FuncValue:
		if cv.madeIn != nil {
			if kept(cv) {
				cv.setMadeIn(nil)
				return v
			}
			alloc.AllocateFunc()
			cp := *cv
			cp.ObjectInfo = ObjectInfo{}
			return &cp
		}
	case *BoundMethodValue:
		if cv.madeIn != nil {
			if kept(cv) {
				cv.setMadeIn(nil)
				return v
			}
			alloc.AllocateBoundMethod()
			cp := *cv
			cp.ObjectInfo = ObjectInfo{}
			if fn := cv.Func; fn != nil && !kept(fn) {
				alloc.AllocateFunc()
				fc := *fn
				fc.ObjectInfo = ObjectInfo{}
				cp.Func = &fc
			}
			return &cp
		}
	}
	return v
}

// shareSlot replaces a marked function value in tv by an unmarked copy and
// shares what tv reaches.
func (alloc *Allocator) shareSlot(tv *TypedValue) {
	if tv.V == nil {
		return
	}
	tv.V = alloc.unmarked(tv.V)
	alloc.shareDeep(tv.V)
}

// shareInto is shareSlot for tv, a slot of base, when base is saved or shared:
// every write into such a slot goes through here, or it could plant a writer.
func (alloc *Allocator) shareInto(base Object, tv *TypedValue) {
	if sharedOrSaved(base) {
		alloc.shareSlot(tv)
	}
}

// kept reports whether o is saved or already referenced by an object.
func kept(o Object) bool { return o.GetIsReal() || o.GetRefCount() > 0 }

// sharedOrSaved reports whether what is written into o can be seen outside
// the call.
func sharedOrSaved(o Object) bool {
	oi := o.GetObjectInfo()
	return oi.shared || oi.GetIsReal()
}

// shareDeep replaces, in every unsaved object reachable from v, each function
// value that has a maker by an unmarked copy, and flags those objects shared.
// Saved and already-shared objects end the walk; a function's captures are
// skipped. A function value given directly has no slot: it is wiped in place.
// It keeps its own work list: a long pointer chain must not grow the Go stack.
func (alloc *Allocator) shareDeep(v Value) {
	var buf [8]Value
	todo := append(buf[:0], v)
	slot := func(tv *TypedValue) {
		switch tv.V.(type) {
		case nil:
		case *FuncValue, *BoundMethodValue:
			tv.V = alloc.unmarked(tv.V)
		default:
			todo = append(todo, tv.V)
		}
	}
	for len(todo) > 0 {
		v, todo = todo[len(todo)-1], todo[:len(todo)-1]
		switch cv := v.(type) {
		case PointerValue:
			switch cv.Base.(type) {
			case nil, *Block:
				if cv.TV != nil {
					slot(cv.TV)
				}
			default:
				todo = append(todo, cv.Base)
			}
		case *SliceValue:
			todo = append(todo, cv.Base)
		case *FuncValue:
			cv.setMadeIn(nil)
		case *BoundMethodValue:
			cv.setMadeIn(nil)
		case *ArrayValue:
			if markShared(cv) {
				for i := range cv.List {
					slot(&cv.List[i])
				}
			}
		case *StructValue:
			if markShared(cv) {
				for i := range cv.Fields {
					slot(&cv.Fields[i])
				}
			}
		case *MapValue:
			if markShared(cv) {
				for cur := cv.List.Head; cur != nil; cur = cur.Next {
					todo = append(todo, cur.Key.V) // a key holds no function itself
					slot(&cur.Value)
				}
			}
		case *HeapItemValue:
			if markShared(cv) {
				slot(&cv.Value)
			}
		}
	}
}

// markShared flags o shared; false if it was already, or is saved.
func markShared(o Object) bool {
	if sharedOrSaved(o) {
		return false
	}
	o.GetObjectInfo().shared = true
	return true
}

// handOffLends ends the loans of what each function value in args captures
// when the call leaves the caller's hold or enters read-only code: what is
// handed over there must not keep a writer alive through its captures.
func (m *Machine) handOffLends(fr *Frame, args []TypedValue) {
	if len(m.lends) == 0 {
		return
	}
	if !m.viewOnly || fr.LastViewOnly {
		if now, _ := m.Alloc.holdRealm(); now == fr.callerHold() {
			return
		}
	}
	for i := range args {
		if fv, ok := args[i].V.(*FuncValue); ok {
			m.dropCapturedLends(fv)
		}
	}
}
