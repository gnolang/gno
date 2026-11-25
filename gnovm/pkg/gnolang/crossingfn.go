package gnolang

// Rule R2a (pr6495_cross_to_write.md): a non-crossing function value called BY
// VALUE writes only during the crossing call that made it. The crossing-call
// marker lives in memory only: never stored, encoded or hashed.

// crossingCallee reports which crossing call made the callee value; nil for a
// stored location and for any value a copy or reload did not record.
func (m *Machine) crossingCallee(tv *TypedValue) *crossingCall {
	switch cv := tv.V.(type) {
	case *FuncValue:
		return cv.madeInCall()
	case *BoundMethodValue:
		return cv.madeInCall()
	}
	return nil
}

// calledByValue reports whether cx reaches its callee as a value rather than by
// its declared name. ATTR_NAMED_CALLEE is the static answer: a realm's own
// variable holding a function is not named, so an alias cannot dodge the rule.
func calledByValue(cx *CallExpr) bool {
	if cx.GetAttribute(ATTR_NAMED_CALLEE) == true {
		return false
	}
	switch fx := cx.Func.(type) {
	case *NameExpr:
		return fx.Path.Type == VPBlock
	case *SelectorExpr:
		return !fx.Path.IsMethod()
	case *IndexExpr, *StarExpr:
		return true
	}
	return true
}

// restrictByValueCall applies R2a to a call that would otherwise run writable:
// a non-crossing callee reached by value writes only in the crossing call that
// made the value.
func (m *Machine) restrictByValueCall(cx *CallExpr, fv *FuncValue, withCross bool, madeIn *crossingCall) {
	if withCross || fv.IsCrossing() || m.Realm == nil || !calledByValue(cx) {
		return
	}
	if madeIn != nil && madeIn == m.curCrossing {
		return // made during the crossing call that is calling it
	}
	m.enterReadOnlyByS3(m.Realm)
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
		shareDeep(args[i].V)
	}
}

// shareDeep clears the maker of every function value reachable from v
// through unsaved objects and flags those objects shared, so R2a holds only
// for a function value that stayed private to the call that made it. Saved
// and already-shared objects end the walk; a function's captures are skipped.
// It keeps its own work list: a long pointer chain must not grow the Go stack.
func shareDeep(v Value) {
	var buf [8]Value
	todo := append(buf[:0], v)
	for len(todo) > 0 {
		v, todo = todo[len(todo)-1], todo[:len(todo)-1]
		switch cv := v.(type) {
		case PointerValue:
			switch cv.Base.(type) {
			case nil, *Block:
				if cv.TV != nil {
					todo = append(todo, cv.TV.V)
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
					todo = append(todo, cv.List[i].V)
				}
			}
		case *StructValue:
			if markShared(cv) {
				for i := range cv.Fields {
					todo = append(todo, cv.Fields[i].V)
				}
			}
		case *MapValue:
			if markShared(cv) {
				for cur := cv.List.Head; cur != nil; cur = cur.Next {
					todo = append(todo, cur.Key.V, cur.Value.V)
				}
			}
		case *HeapItemValue:
			if markShared(cv) {
				todo = append(todo, cv.Value.V)
			}
		}
	}
}

// markShared flags o shared; false if it was already, or is saved.
func markShared(o Object) bool {
	oi := o.GetObjectInfo()
	if oi.shared || oi.GetIsReal() {
		return false
	}
	oi.shared = true
	return true
}
