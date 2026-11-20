package gnolang

// Rule S3 (pr6495_cross_to_write.md): a non-crossing function value called BY
// VALUE writes only during the crossing call that made it. The crossing id
// lives in memory only: never stored, encoded or hashed.

// crossingCallee reports which crossing call made the callee value; 0 for a
// stored location and for any value a copy or reload did not record.
func (m *Machine) crossingCallee(tv *TypedValue) uint64 {
	switch cv := tv.V.(type) {
	case *FuncValue:
		return cv.madeIn()
	case *BoundMethodValue:
		return cv.madeIn()
	}
	return 0
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

// restrictByValueCall applies S3 to a call that would otherwise run writable:
// a non-crossing callee reached by value writes only in the crossing call that
// made the value.
func (m *Machine) restrictByValueCall(cx *CallExpr, fv *FuncValue, withCross bool, madeIn uint64) {
	if withCross || fv.IsCrossing() || m.Realm == nil || !calledByValue(cx) {
		return
	}
	if madeIn != 0 && madeIn == m.curCrossing {
		return // made during the crossing call that is calling it
	}
	m.enterReadOnlyByS3(m.Realm)
}
