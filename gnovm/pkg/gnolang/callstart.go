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
	kind  startKind
	realm *Realm // startWrite, startReadOnly
}

// callFacts are what decideCallStart reads about one call besides the machine.
type callFacts struct {
	cx        *CallExpr
	fv        *FuncValue
	pv        *PackageValue // the callee's package
	withCross bool
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
	return callStart{kind: startUnrouted}
}
