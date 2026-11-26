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
	owner           string // the receiver's owning package, "" if none
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
	pkg  string
	mode wmMode
}

// wmLoan: fn was written in place as an argument while realm wrote; it ends
// when the receiving call returns.
type wmLoan struct {
	fn       any
	realm    string
	crossing int
}

type wmState struct {
	base     wmMode    // the mode below every call
	stack    []wmFrame // running calls, the caller on top
	loans    []wmLoan  // live loans
	crossing int       // the current crossing
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
	// 3. Realm code writes only for its own realm; from elsewhere it is read-only.
	case wmRealmCode(f.callee):
		switch {
		case cur.kind == wmNone:
			return wmAnswer{mode: wmWriteAs(home)}
		case cur.writer() == home:
			return s.byValue(f, cur, 0)
		case cur.readOnlyAs(home):
			return wmAnswer{mode: cur}
		}
		return wmAnswer{mode: wmMode{wmReadOnly, home}}
	}
	// 4. Library code acts for the caller's realm unless that power was not handed to it.
	w, asks, libOwned := cur.writer(), wmAsk(0), wmLibPkg(f.owner)
	viaValue := f.recv && f.form != wmMethod && f.form != wmIface
	view := func() { w, cur = "", wmMode{kind: wmView} }
	// 4a. A method value bound by a realm acts only while that realm's calls lead here.
	if wmRealmPkg(f.binder) && !libOwned {
		if asks |= wmAskBinder; !s.ledBy(f.binder) {
			view()
		}
	}
	// 4b. A function value made under another realm, or none, acts for no realm.
	if !f.recv && w != "" && f.maker != w && f.form != wmNamed && !s.lent(f.fn, w) {
		view()
	}
	// 4c. A method value bound while acting for no realm acts for none.
	if viaValue && f.binder == "" && w != "" && f.callee != uversePkgPath {
		view()
	}
	// 4d. Code the caller picked from an untrusted package acts for no realm.
	if (viaValue || f.form == wmIface) && !libOwned && w != "" && f.callee != w && f.callee != f.caller {
		if asks |= wmAskForeign; f.foreign {
			view()
		}
	}
	// 4e. A library's own object is written as that library's realm.
	if libOwned && w != f.owner {
		return s.byValue(f, wmWriteAs(wmRealmOf(f.owner)), asks|wmAskOwner)
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
