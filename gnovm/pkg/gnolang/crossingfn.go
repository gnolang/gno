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
