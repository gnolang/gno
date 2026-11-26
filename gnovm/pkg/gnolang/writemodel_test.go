package gnolang

import "strings"

// A small model of the write rule (pr6495), written apart from the VM and
// checked against it. It keeps its own stack of modes and its own loans, and
// takes from the VM only the facts listed on wmFacts.

type wmKind uint8

const (
	wmNone           wmKind = iota // no realm at all: genesis, tooling
	wmWrite                        // writes as realm
	wmReadOnly                     // realm's code entered without crossing
	wmByValue                      // read-only as realm: a function value outside its loan
	wmView                         // acts for no realm
	wmErrNotCrossing               // error: cross call of a non-crossing function
	wmErrExternalCur               // error: another realm's crossing function without cross
)

type wmMode struct {
	kind  wmKind
	realm string
}

func wmWriteAs(r string) wmMode {
	if r == "" {
		return wmMode{}
	}
	return wmMode{wmWrite, r}
}

// writer is the realm whose stored state the mode may change, or "".
func (md wmMode) writer() string {
	if md.kind == wmWrite {
		return md.realm
	}
	return ""
}

// readOnlyAs reports whether the mode runs realm r's own code read-only.
func (md wmMode) readOnlyAs(r string) bool {
	return r != "" && md.realm == r && (md.kind == wmReadOnly || md.kind == wmByValue)
}

// wmForm is how the call names its callee in the source.
type wmForm uint8

const (
	wmNamed   wmForm = iota // a declared function, named at the call
	wmHelper                // a local helper proved never handed out
	wmLit                   // a function literal called where it is written
	wmValue                 // a function value held in a variable, slot or result
	wmBuiltin               // a builtin
	wmMethod                // x.M() on a concrete receiver
	wmIface                 // x.M() through an interface
)

// wmFacts is one call. owner, maker and foreign come from the VM.
type wmFacts struct {
	cross, crossing bool
	callee, caller  string // package paths
	form            wmForm
	recv            bool   // a receiver is bound
	owner           string // the package of the receiver's first object, "" if none
	fresh           bool   // the callee gets a fresh copy of a struct or array receiver
	binder          string // the package that bound a method value, "" if none
	maker           string // the package the function value was made under
	fn              any    // the function value, to find its loan
	foreign         bool   // the callee's package is outside the realm's trust
}

// wmAsk names the costly facts an answer needed.
type wmAsk uint8

const (
	wmAskBinder wmAsk = 1 << iota
	wmAskForeign
	wmAskOwner
)

type wmAnswer struct {
	mode  wmMode
	cross bool // a new crossing starts
	asks  wmAsk
}

type wmFrame struct {
	pkg      string
	mode     wmMode
	crossing int    // the crossing the call runs in
	hold     string // the realm new values belong to while it runs
}

// wmLoan: fn was written in place as an argument while realm wrote; it ends
// when the receiving call returns.
type wmLoan struct {
	fn       any
	realm    string
	crossing int
	depth    int // the receiving call's place on the stack
}

type wmState struct {
	base     wmMode    // the mode below every call
	baseHold string    // the hold below every call
	baseX    int       // the crossing below every call
	stack    []wmFrame // running calls, the caller on top
	loans    []wmLoan  // live loans
	crossing int       // the current crossing
	next     int       // the last crossing number used
}

func (s *wmState) mode() wmMode {
	if n := len(s.stack); n > 0 {
		return s.stack[n-1].mode
	}
	return s.base
}

// start answers the mode a call starts in.
func (s *wmState) start(f wmFacts) wmAnswer {
	cur, home := s.mode(), wmRealmOf(f.callee)
	switch {
	// 1. A cross call reaches only a crossing function, which writes as its realm.
	case f.cross && !f.crossing:
		return wmAnswer{mode: wmMode{kind: wmErrNotCrossing}}
	case f.cross:
		return wmAnswer{mode: wmWriteAs(home), cross: true}
	// 2. Without cross, only the callee's own realm calls its crossing function.
	case f.crossing:
		if cur.writer() != home && !cur.readOnlyAs(home) {
			return wmAnswer{mode: wmMode{kind: wmErrExternalCur}}
		}
		return wmAnswer{mode: cur}
	// 3. Realm code: who may write here is its own realm, and only that.
	case wmRealmCode(f.callee):
		switch {
		case cur.kind == wmNone: // 3a. first code to run: writes as its realm
			return wmAnswer{mode: wmWriteAs(home)}
		case cur.writer() == home: // 3b. called by its realm: writes, but see rule 5
			return s.byValue(f, cur, 0)
		case cur.readOnlyAs(home): // 3c. except: its own read-only code stays read-only
			return wmAnswer{mode: cur}
		}
		return wmAnswer{mode: wmMode{wmReadOnly, home}} // 3d. except: anyone else, read-only
	}
	// 4. Library code: who may write here.
	w := cur.writer()
	viaValue := f.recv && f.form != wmMethod && f.form != wmIface
	unbound := viaValue && f.binder == "" && f.callee != uversePkgPath // bound while acting for none
	// 4a. A library's own receiver object: that library, as its realm.
	if !f.fresh && wmLibPkg(f.owner) {
		if w == f.owner && !unbound { // the caller already writes as it: nothing to ask
			return s.byValue(f, cur, 0)
		}
		return s.byValue(f, wmWriteAs(wmRealmOf(f.owner)), wmAskOwner)
	}
	// 4b. Otherwise the caller's realm, if it writes, with these exceptions.
	var asks wmAsk
	view := func() wmAnswer { return wmAnswer{mode: wmMode{kind: wmView}, asks: asks} }
	// 4c. Except: a method value bound by a realm, while that realm's calls do
	// not lead here, acts for none, even under a caller that writes nothing.
	if wmRealmPkg(f.binder) {
		if asks |= wmAskBinder; !s.ledBy(f.binder) {
			return view()
		}
	}
	if w == "" {
		return wmAnswer{mode: cur, asks: asks}
	}
	// 4d. Except: a function value made under another realm, or none, and not lent to w.
	// 4e. Except: a method value bound while acting for no realm.
	if (!f.recv && f.form != wmNamed && f.maker != w && !s.lent(f.fn, w)) || unbound {
		return view()
	}
	// 4f. Except: code the caller picked from an untrusted package.
	if (viaValue || f.form == wmIface) && f.callee != w && f.callee != f.caller {
		if asks |= wmAskForeign; f.foreign {
			return view()
		}
	}
	return s.byValue(f, cur, asks)
}

// 5. A function value called outside its loan runs read-only, unless the test runner called it.
func (s *wmState) byValue(f wmFacts, md wmMode, asks wmAsk) wmAnswer {
	if w := md.writer(); w != "" && f.form == wmValue && !s.lent(f.fn, w) && !s.testRunner(f) {
		md = wmMode{wmByValue, w}
	}
	return wmAnswer{mode: md, asks: asks}
}

// 6. A loan lets its function value write only for its realm, in its crossing.
func (s *wmState) lent(fn any, realm string) bool {
	for _, l := range s.loans {
		if l.fn == fn && l.realm == realm && l.crossing == s.crossing {
			return true
		}
	}
	return false
}

// ledBy reports whether every running call down to the nearest realm code acts for r.
func (s *wmState) ledBy(r string) bool {
	for i := len(s.stack) - 1; i >= 0; i-- {
		if s.stack[i].mode.writer() != r {
			return false
		}
		if wmRealmCode(s.stack[i].pkg) {
			return true
		}
	}
	return s.base.writer() == r
}

// testRunner reports whether only the "testing" package's calls lead here.
func (s *wmState) testRunner(f wmFacts) bool {
	if f.caller != TestingBasePkgPath || len(s.stack) == 0 {
		return false
	}
	for _, fr := range s.stack {
		if fr.pkg != TestingBasePkgPath {
			return false
		}
	}
	return true
}

func wmRealmCode(p string) bool { return IsRealmPath(p) || IsEphemeralPath(p) }

// wmLibPkg: an object of a /p/ library (or test overlay) is a frozen library's.
func wmLibPkg(p string) bool {
	return p != "" && !IsStdlib(p) && (IsPPackagePath(p) || IsTestOverlayPath(p) || p == uversePkgPath)
}

// wmRealmPkg: a package whose objects are a realm's state.
func wmRealmPkg(p string) bool {
	return p != "" && !IsStdlib(p) && !IsPPackagePath(p) && !IsTestOverlayPath(p) && p != uversePkgPath
}

// wmRealmOf is the realm a package's code acts as, "" if none.
func wmRealmOf(p string) string {
	_, run := IsGnoRunPath(p)
	lib := !strings.HasSuffix(p, "_test") && (IsPPackagePath(p) || IsStdlib(p))
	if IsRealmPath(p) || p == "main" || run || lib {
		return p
	}
	return ""
}

// wmArg is one argument as the callee receives it; a variadic tail is one slice.
type wmArg struct {
	fn     any     // a function value, or nil
	reach  []any   // the function values fn reaches through captures, at any depth
	slice  bool    // a slice: its elements escape like stored values
	elems  []wmArg // the slice's elements
	nonesc bool    // the parameter is marked as never letting its value escape
	lit    bool    // a function literal written in place
}

// readOnly reports whether md writes no stored state.
func (md wmMode) readOnly() bool {
	return md.kind == wmView || md.kind == wmReadOnly || md.kind == wmByValue
}

func (s *wmState) hold() string {
	if n := len(s.stack); n > 0 {
		return s.stack[n-1].hold
	}
	return s.baseHold
}

// enter pushes a call that starts in a. Its arguments then hand over loans
// (rule 7) and, unless the call was deferred, lend (rule 8).
func (s *wmState) enter(callee string, a wmAnswer, args []wmArg, deferred bool) {
	caller, callerHold, callerX := s.mode(), s.hold(), s.crossing
	fr := wmFrame{pkg: callee, mode: a.mode, crossing: callerX, hold: a.mode.realm}
	if a.mode.kind == wmErrNotCrossing || a.mode.kind == wmErrExternalCur {
		fr.mode, fr.hold = caller, callerHold // the call fails before it runs
		s.stack = append(s.stack, fr)
		return
	}
	if a.mode.kind == wmView {
		fr.hold = callerHold
	}
	if a.cross {
		s.next++
		fr.crossing = s.next
	}
	s.stack, s.crossing = append(s.stack, fr), fr.crossing
	// 7. Leaving the caller's hold, or entering read-only code, ends the loans an argument captures.
	if (fr.mode.readOnly() && !caller.readOnly()) || fr.hold != callerHold {
		for _, x := range args {
			if x.fn != nil {
				s.drop(x.reach...)
			}
		}
	}
	if deferred {
		return
	}
	// 8. A literal written in place for a non-escaping parameter is lent to the call
	// while its caller writes; any other function value handed over loses its loans.
	for _, x := range args {
		switch {
		case x.slice:
			for _, e := range x.elems {
				s.drop(e.fn)
				s.drop(e.reach...)
			}
		case x.fn == nil:
		case s.hasLoan(x.fn):
			if !x.nonesc {
				s.drop(x.fn)
				s.drop(x.reach...)
			}
		case x.nonesc && x.lit && caller.kind == wmWrite:
			s.loans = append(s.loans, wmLoan{fn: x.fn, realm: caller.realm, crossing: callerX, depth: len(s.stack) - 1})
		default:
			s.drop(x.reach...)
		}
	}
}

// pop leaves n calls running: the loans of the others end with them.
func (s *wmState) pop(n int) {
	s.stack = s.stack[:n]
	s.crossing = s.baseX
	if n > 0 {
		s.crossing = s.stack[n-1].crossing
	}
	kept := s.loans[:0]
	for _, l := range s.loans {
		if l.depth < n {
			kept = append(kept, l)
		}
	}
	s.loans = kept
}

func (s *wmState) hasLoan(fn any) bool {
	for _, l := range s.loans {
		if l.fn == fn {
			return true
		}
	}
	return false
}

func (s *wmState) drop(fns ...any) {
	for _, fn := range fns {
		for i := 0; i < len(s.loans); i++ {
			if s.loans[i].fn == fn {
				s.loans = append(s.loans[:i], s.loans[i+1:]...)
				i--
			}
		}
	}
}
