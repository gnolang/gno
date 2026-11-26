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
