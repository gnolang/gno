package gnolang

// The mode a call starts in (pr6495_cross_to_write.md): decideCallStart is the
// one place that answers it; pushFrameCall gathers the facts and applies it.

// startKind is the first step of a call start.
type startKind uint8

const (
	startKeep          startKind = iota // run in the caller's mode
	startCross                          // write as the callee's realm, in a new crossing
	startWrite                          // write as callStart.realm
	startView                           // view-only: no realm, every stored object refused
	startReadOnly                       // read-only as callStart.realm (entered without crossing)
	startNotCrossing                    // error: cross-call of a non-crossing function
	startExternalCur                    // error: cur-call into another realm's crossing function
	startAskCalledBy                    // fill callFacts.calledBy, then ask again
	startAskForeignLib                  // fill callFacts.foreign, then ask again
	startAskBorrow                      // fill callFacts.borrowTo, then ask again
)

// callStart is decideCallStart's answer, applied in field order.
type callStart struct {
	kind     startKind
	realm    *Realm // startWrite, startReadOnly
	borrow   bool   // then write as borrowTo (a /p/-owned receiver's package)
	borrowTo *Realm
	byValue  bool // then read-only by value as the realm then current
}

// callFacts are what decideCallStart reads about one call besides the machine.
// calledBy, foreign and borrowTo cost gas or fill a cache: the caller computes
// each only when an answer asks for it, as the call path always has.
type callFacts struct {
	cx        *CallExpr
	fv        *FuncValue
	pv        *PackageValue // the callee's package
	withCross bool
	realmPkg  bool   // the callee is /r/- or /e/-declared
	lastPkg   string // the calling package
	recvSet   bool   // a receiver is bound (method call or method value)
	pOwned    bool   // the receiver is a /p/ package's object
	recvPID   PkgID  // the receiver's package, when pOwned
	bindPID   PkgID  // a method value's binder stamp
	viaIface  bool
	viaValue  bool

	calledBy, haveCalledBy bool
	foreign, haveForeign   bool
	borrowTo               *Realm
	haveBorrow             bool
}

// decideCallStart answers whether a call starts writable, read-only, view-only
// or in the caller's mode. It reads the machine as it is right after the
// callee's frame is pushed and changes nothing; an ask* answer names a costly
// fact to fill before asking again.
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
	// Not cross nor crossing. A realm's stored state changes only during a
	// crossing call into it:
	//
	//   #1. /r/- or /e/-declared function/method/closure entered from another
	//       realm (or from read-only code) runs read-only: it constructs its
	//       realm's objects and writes only objects allocated in the current
	//       read-only stretch, never stored ones. Its own realm's calls are
	//       unchanged.
	//
	//   #2. Otherwise (/p/-declared) → only a /p/-owned receiver borrows,
	//       so the post-init immutability gate reports the write. A
	//       receiver owned by another realm is a view, real or not, and the
	//       method's writes hit the readonly gate.
	//
	//       A method value (x.M bound earlier, then called) writes for its
	//       binder only while every frame down to the nearest /r/ or /e/
	//       one runs as the binder; otherwise it runs view-only, so neither
	//       the binder's nor a caller's stored objects can be written
	//       through it. A /p/-owned receiver borrows as in a direct call.
	//
	//   A /p/- or stdlib-declared closure borrows nothing. Called while the
	//   realm it was built under is current, it runs as that realm;
	//   otherwise it runs view-only (it arrived as another realm's value).
	//
	// Stdlib-stamped receivers are EXCLUDED from borrow rule #2: stdlib is
	// a trusted, frozen transform library that owns no mutable state, so
	// its methods run with the CALLER's authority (m.Realm unchanged), not
	// stdlib's. This lets a stdlib method write a caller-supplied out-param
	// buffer (e.g. base64.Encode(dst, src) where dst is caller-owned), while
	// still blocking writes to a third realm's data (foreign != caller). It
	// also keeps the nil-realm write hole closed: a tx's entry realm is
	// non-nil, so a stdlib method writing a foreign /r/ object is rejected
	// by the readonly gate against the caller's realm.
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
	// /p/ or stdlib callee: each rule below sees the mode the ones above left.
	cs := callStart{kind: startKeep}
	realm, view := m.Realm, m.viewOnly
	// Rule #2 prelude (see "A method value" above).
	if f.bindPID.IsRealmPkg() && !f.pOwned {
		if !f.haveCalledBy {
			return callStart{kind: startAskCalledBy}
		}
		if !f.calledBy {
			cs.kind, realm, view = startView, nil, true
		}
	}
	// A library closure or handle made under another realm, or with no maker,
	// reached through a value acts for no realm: it runs view-only. A value
	// in a library's package variable named by the call counts as named.
	if !f.recvSet && realm != nil && f.fv.GetObjectInfo().ID.PkgID != realm.ID &&
		f.cx.GetAttribute(ATTR_NAMED_CALLEE) != true && !m.lendLiveAs(f.fv, realm, view) {
		cs.kind, realm, view = startView, nil, true
	}
	// A /p/ method reached through an interface or a method value runs code the
	// caller picked; it acts for the realm only if foreignLib trusts its
	// package. A method value built while view-only (zero stamp) acts for no
	// realm at all.
	if f.viaValue && f.bindPID.IsZero() && realm != nil && f.fv.PkgPath != uversePkgPath {
		cs.kind, realm, view = startView, nil, true
	}
	if (f.viaIface || f.viaValue) && !f.pOwned && realm != nil &&
		f.fv.PkgPath != realm.Path && f.fv.PkgPath != f.lastPkg {
		if !f.haveForeign {
			return callStart{kind: startAskForeignLib}
		}
		if f.foreign {
			cs.kind, realm, view = startView, nil, true
		}
	}
	if f.pOwned && (realm == nil || f.recvPID != realm.ID) {
		if !f.haveBorrow {
			return callStart{kind: startAskBorrow}
		}
		cs.borrow, cs.borrowTo = true, f.borrowTo
		realm, view = f.borrowTo, false
	}
	cs.byValue = m.byValueReadOnly(f, realm, view)
	return cs
}

// byValueReadOnly reports whether a call that would run as realm (viewOnly as
// given) runs read-only instead: a non-crossing callee reached by value whose
// loan is not live, unless the test runner itself made the call.
func (m *Machine) byValueReadOnly(f *callFacts, realm *Realm, viewOnly bool) bool {
	return !f.withCross && !f.fv.IsCrossing() && realm != nil && calledByValue(f.cx) &&
		!m.lendLiveAs(f.fv, realm, viewOnly) && !m.testRunnerCall()
}

// askCallStart asks decideCallStart until it answers, filling each costly
// fact it asks for once, in the order it asks.
func (m *Machine) askCallStart(f *callFacts) callStart {
	for {
		cs := m.decideCallStart(f)
		switch cs.kind {
		case startAskCalledBy:
			f.calledBy, f.haveCalledBy = m.calledBy(f.bindPID), true
		case startAskForeignLib:
			f.foreign, f.haveForeign = m.foreignLib(f.fv, f.lastPkg, f.viaIface), true
		case startAskBorrow:
			objpv := m.Store.GetObject(ObjectIDFromPkgID(f.recvPID)).(*PackageValue)
			f.borrowTo, f.haveBorrow = objpv.GetRealm(), true
			assertBorrowedRealm(objpv.PkgPath, f.borrowTo)
		default:
			return cs
		}
	}
}
