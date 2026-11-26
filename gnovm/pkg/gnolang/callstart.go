package gnolang

// The mode a call starts in (pr6495_cross_to_write.md): decideCallStart is the
// one place that answers it; pushFrameCall gathers the facts and applies it.

// startKind is the first step of a call start.
type startKind uint8

const (
	startKeep        startKind = iota // run in the caller's mode
	startCross                        // write as the callee's realm, in a new crossing
	startWrite                        // write as callStart.realm
	startView                         // view-only: no realm, every stored object refused
	startReadOnly                     // read-only as callStart.realm (entered without crossing)
	startNotCrossing                  // error: cross-call of a non-crossing function
	startExternalCur                  // error: cur-call into another realm's crossing function
	startUnrouted                     // not decided here yet
)

// callStart is decideCallStart's answer, applied in field order.
type callStart struct {
	kind    startKind
	realm   *Realm // startWrite, startReadOnly
	byValue bool   // then read-only by value as the realm then current
}

// callFacts are what decideCallStart reads about one call besides the machine.
type callFacts struct {
	cx        *CallExpr
	fv        *FuncValue
	pv        *PackageValue // the callee's package
	withCross bool
	realmPkg  bool // the callee is /r/- or /e/-declared
}

// decideCallStart answers whether a call starts writable, read-only or in the
// caller's mode. It reads the machine as it is right after the callee's frame
// is pushed and changes nothing.
func (m *Machine) decideCallStart(f *callFacts) callStart {
	if f.withCross {
		if !f.fv.IsCrossing() {
			return callStart{kind: startNotCrossing}
		}
		return callStart{kind: startCross}
	}
	if f.fv.IsCrossing() {
		// A read-only frame of the callee's realm is still that realm's code:
		// doOpCall's identity check reports a stale cur.
		if m.Realm != f.pv.Realm && !m.readOnlyOf(f.pv.Realm) {
			return callStart{kind: startExternalCur}
		}
		return callStart{kind: startKeep}
	}
	if f.realmPkg {
		switch {
		case m.Realm == nil && !m.viewOnly:
			return callStart{kind: startWrite, realm: f.pv.GetRealm()} // no realm context (genesis, tooling)
		case m.Realm != nil && m.Realm.Path == f.pv.PkgPath:
			// the realm's own code; a function value called by value writes
			// only while lent in place to a live call
			return callStart{kind: startKeep, byValue: m.byValueReadOnly(f, m.Realm, m.viewOnly)}
		case m.viewOnly && m.viewRealm != nil && m.viewRealm.Path == f.pv.PkgPath:
			return callStart{kind: startKeep} // the same realm's code inside its read-only entry
		default:
			// Entered without crossing: a realm's stored state changes only
			// during a crossing call into it.
			return callStart{kind: startReadOnly, realm: f.pv.GetRealm()}
		}
	}
	return callStart{kind: startUnrouted}
}

// byValueReadOnly reports whether a call that would run as realm (viewOnly as
// given) runs read-only instead: a non-crossing callee reached by value whose
// loan is not live, unless the test runner itself made the call.
func (m *Machine) byValueReadOnly(f *callFacts, realm *Realm, viewOnly bool) bool {
	return !f.withCross && !f.fv.IsCrossing() && realm != nil && calledByValue(f.cx) &&
		!m.lendLiveAs(f.fv, realm, viewOnly) && !m.testRunnerCall()
}
