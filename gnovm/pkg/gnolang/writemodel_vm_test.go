package gnolang

import (
	"fmt"
	"io"
	"sync"
	"testing"
)

// Part A: every combination of the facts decideCallStart reads, with real
// minimal values, answered by the VM and by the model (writemodel_test.go).

const (
	wmRA = "gno.land/r/t/a"
	wmRB = "gno.land/r/t/b"
	wmPL = "gno.land/p/t/l"
	wmPM = "gno.land/p/t/m"
	wmEE = "gno.land/e/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5/run"
	wmSL = "strings"
)

// wmVMRealms are the realms NewPackage gives these packages (.uverse has none).
var wmVMRealms = map[string]*Realm{}

func init() {
	for _, p := range []string{wmRA, wmRB, wmPL, wmPM, wmEE, wmSL, "main", TestingBasePkgPath} {
		wmVMRealms[p] = NewRealm(p)
	}
}

func wmPID(p string) PkgID {
	if p == "" {
		return PkgID{}
	}
	return PkgIDFromPkgPath(p)
}

// wmSetMode puts the machine in md.
func wmSetMode(m *Machine, md wmMode) {
	m.setRealm(nil)
	switch md.kind {
	case wmWrite:
		m.setRealm(wmVMRealms[md.realm])
	case wmReadOnly:
		m.enterReadOnly(wmVMRealms[md.realm])
	case wmByValue:
		m.enterReadOnlyByS3(wmVMRealms[md.realm])
	case wmView:
		m.setViewOnly()
	}
}

// wmVMMode reads the machine's mode.
func wmVMMode(m *Machine) wmMode {
	switch {
	case m.viewOnly && m.viewRealm != nil && m.viewS3:
		return wmMode{wmByValue, m.viewRealm.Path}
	case m.viewOnly && m.viewRealm != nil:
		return wmMode{wmReadOnly, m.viewRealm.Path}
	case m.viewOnly:
		return wmMode{kind: wmView}
	case m.Realm != nil:
		return wmMode{wmWrite, m.Realm.Path}
	}
	return wmMode{}
}

func wmRealmPath(r *Realm) string {
	if r == nil {
		return ""
	}
	return r.Path
}

// wmAnswerOf turns the VM's answer into the mode it starts, from the caller's mode.
func wmAnswerOf(m *Machine, f *callFacts, cs callStart, asks wmAsk) wmAnswer {
	a := wmAnswer{mode: wmVMMode(m), asks: asks}
	switch cs.kind {
	case startNotCrossing:
		return wmAnswer{mode: wmMode{kind: wmErrNotCrossing}, asks: asks}
	case startExternalCur:
		return wmAnswer{mode: wmMode{kind: wmErrExternalCur}, asks: asks}
	case startCross:
		return wmAnswer{mode: wmWriteAs(wmRealmPath(f.pv.Realm)), cross: true, asks: asks}
	case startWrite:
		a.mode = wmWriteAs(wmRealmPath(cs.realm))
	case startView:
		a.mode = wmMode{kind: wmView}
	case startReadOnly:
		a.mode = wmMode{wmReadOnly, wmRealmPath(cs.realm)}
	}
	if cs.borrow {
		a.mode = wmWriteAs(wmRealmPath(cs.borrowTo))
	}
	if cs.byValue {
		a.mode = wmMode{wmByValue, a.mode.writer()}
		if a.mode.realm == "" {
			a.mode.kind = 99 // never: read-only by value with no realm
		}
	}
	return a
}

var wmCallExprs = func() (cxs [wmIface + 1]*CallExpr) {
	for i := range cxs {
		cxs[i] = wmCallExpr(wmForm(i))
	}
	return
}()

// wmCallExpr builds a call naming its callee as form does.
func wmCallExpr(form wmForm) *CallExpr {
	cx := &CallExpr{}
	switch form {
	case wmNamed, wmHelper, wmValue:
		cx.Func = &NameExpr{Path: ValuePath{Type: VPBlock}}
	case wmLit:
		cx.Func = &FuncLitExpr{}
	case wmBuiltin:
		cx.Func = &NameExpr{Path: ValuePath{Type: VPUverse}}
	case wmMethod:
		cx.Func = &SelectorExpr{Path: ValuePath{Type: VPPtrMethod}}
	case wmIface:
		cx.Func = &SelectorExpr{Path: ValuePath{Type: VPInterface}}
	}
	switch form {
	case wmNamed:
		cx.SetAttribute(ATTR_NAMED_CALLEE, true)
	case wmHelper:
		cx.SetAttribute(ATTR_LOCAL_HELPER, true)
	}
	return cx
}

// wmStack is the calls below the caller and the mode below them all.
type wmStack struct {
	base  wmMode
	below []wmFrame
	top   bool // the caller is a call (else it runs at the base)
}

// wmLoanCase is the callee's loan, if any.
type wmLoanCase struct {
	realm string
	old   bool // made in an earlier crossing
	dead  bool // its call returned
	has   bool
}

type wmCase struct {
	md                wmMode
	callee, caller    string
	cross, crossing   bool
	form              wmForm
	recv              bool
	owner, binder, mk string
	foreign           bool
	stack             wmStack
	loan              wmLoanCase
}

// wmRunVM answers c with decideCallStart on a machine set up as c says,
// filling each costly fact it asks for as askCallStart does.
func wmRunVM(m *Machine, c *wmCase, fv *FuncValue) wmAnswer {
	cur := &crossingCall{}
	m.Frames = m.Frames[:0]
	m.lends = nil
	wmSetMode(m, c.md)
	m.curCrossing = cur
	callerPV := &PackageValue{PkgPath: c.caller}
	if c.stack.top {
		mode := c.stack.base
		for _, b := range c.stack.below {
			m.Frames = append(m.Frames, Frame{Func: &FuncValue{PkgPath: b.pkg}, LastRealm: wmVMRealms[mode.writer()]})
			mode = b.mode
		}
		m.Frames = append(m.Frames, Frame{Func: &FuncValue{PkgPath: c.caller}, LastRealm: wmVMRealms[mode.writer()]})
	}
	m.Frames = append(m.Frames, Frame{Func: fv, LastRealm: m.Realm, LastPackage: callerPV, LastCrossing: cur})
	if c.loan.has {
		tok := &lendTok{lent: []*FuncValue{fv}}
		lt := &lendTag{tok: tok, depth: 0, realm: wmVMRealms[c.loan.realm], crossing: cur}
		if c.loan.old {
			lt.crossing = &crossingCall{}
		}
		if !c.loan.dead {
			m.Frames[0].lendTok = tok
		}
		m.lends = map[*FuncValue]*lendTag{fv: lt}
	}

	pv := &PackageValue{PkgPath: c.callee, Realm: wmVMRealms[c.callee]}
	f := callFacts{cx: wmCallExprs[c.form], fv: fv, pv: pv, withCross: c.cross,
		realmPkg: IsRealmPath(c.callee) || IsEphemeralPath(c.callee),
		last:     callerPV, recvSet: c.recv, bindPID: wmPID(c.binder)}
	if !c.cross && !c.crossing && !f.realmPkg {
		if c.recv && c.owner != "" {
			pid := wmPID(c.owner)
			f.pOwned = !pid.IsStdlibPkg() && pid.IsImmutablePkg()
			if f.pOwned {
				f.recvPID = pid
			}
		}
		f.viaValue = c.recv
		if sx, ok := f.cx.Func.(*SelectorExpr); ok && sx.Path.IsMethod() {
			f.viaIface = sx.Path.Type == VPInterface
			f.viaValue = false
		}
	}
	var asks wmAsk
	for {
		cs := m.decideCallStart(&f)
		switch cs.kind {
		case startAskCalledBy:
			asks |= wmAskBinder
			f.calledBy, f.haveCalledBy = m.calledBy(f.bindPID), true
		case startAskForeignLib:
			asks |= wmAskForeign
			f.foreign, f.haveForeign = c.foreign, true
		case startAskBorrow:
			asks |= wmAskOwner
			for p, r := range wmVMRealms {
				if wmPID(p) == f.recvPID {
					f.borrowTo = r
				}
			}
			f.haveBorrow = true
		default:
			return wmAnswerOf(m, &f, cs, asks)
		}
	}
}

// wmRunModel answers c with the model, its state built from c alone.
func wmRunModel(c *wmCase, fn any) wmAnswer {
	s := &wmState{base: c.md, crossing: 1}
	if c.stack.top {
		s.base = c.stack.base
		s.stack = append(s.stack, c.stack.below...)
		s.stack = append(s.stack, wmFrame{pkg: c.caller, mode: c.md})
	}
	if c.loan.has && !c.loan.dead {
		l := wmLoan{fn: fn, realm: c.loan.realm, crossing: 1}
		if c.loan.old {
			l.crossing = 0
		}
		s.loans = append(s.loans, l)
	}
	owner := c.owner
	if !c.recv {
		owner = ""
	}
	return s.start(wmFacts{cross: c.cross, crossing: c.crossing, callee: c.callee, caller: c.caller,
		form: c.form, recv: c.recv, owner: owner, binder: c.binder, maker: c.mk, fn: fn, foreign: c.foreign})
}

func TestWriteModelAllFacts(t *testing.T) {
	m := NewMachineWithOptions(MachineOptions{PkgPath: "", Store: foreignFuncStore(), Output: io.Discard})
	defer m.Release()

	modes := []wmMode{{}, {wmWrite, wmRA}, {wmWrite, wmRB}, {wmWrite, wmPL}, {wmWrite, "main"},
		{wmReadOnly, wmRA}, {wmByValue, wmRA}, {kind: wmView}}
	callees := []string{wmRA, wmEE, wmPL, wmSL, "main", uversePkgPath}
	callers := []string{wmRA, wmPL, TestingBasePkgPath}
	forms := []wmForm{wmNamed, wmHelper, wmLit, wmValue, wmBuiltin, wmMethod, wmIface}
	owners := []string{"-", "", wmPL, wmPM, wmSL, wmRA} // "-": no receiver
	binders := []string{"", wmRA, wmRB, wmPL}
	makers := []string{"", wmRA, wmRB, wmPL}
	stacks := []wmStack{
		{},
		{base: wmWriteAs(wmRA), top: true},
		{base: wmWriteAs(wmRB), top: true},
		{below: []wmFrame{{pkg: wmRA, mode: wmWriteAs(wmRA)}}, top: true},
		{base: wmWriteAs(wmRB), below: []wmFrame{{pkg: wmPL, mode: wmWriteAs(wmRB)}}, top: true},
		{below: []wmFrame{{pkg: TestingBasePkgPath, mode: wmMode{}}}, top: true},
	}
	loans := []wmLoanCase{{}, {realm: wmRA, has: true}, {realm: wmRB, has: true},
		{realm: wmRA, has: true, old: true}, {realm: wmRA, has: true, dead: true}}

	var n, diffs int
	for _, md := range modes {
		for _, callee := range callees {
			for xc := range 4 {
				for _, caller := range callers {
					for _, form := range forms {
						for _, owner := range owners {
							for _, binder := range binders {
								for _, mk := range makers {
									for _, foreign := range []bool{false, true} {
										for _, st := range stacks {
											for _, ln := range loans {
												c := wmCase{md: md, callee: callee, caller: caller,
													cross: xc&1 != 0, crossing: xc&2 != 0, form: form,
													recv: owner != "-", owner: owner, binder: binder, mk: mk,
													foreign: foreign, stack: st, loan: ln}
												if owner == "-" {
													c.owner = ""
												}
												fv := &FuncValue{PkgPath: callee, Crossing: c.crossing}
												fv.ObjectInfo.ID.PkgID = wmPID(mk)
												n++
												got, want := wmRunVM(m, &c, fv), wmRunModel(&c, fv)
												if got != want {
													if diffs++; diffs <= 10 {
														t.Errorf("%+v\n  vm %+v\n  model %+v", c, got, want)
													}
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	t.Logf("MODEL ALL FACTS: %d combinations, %d differences", n, diffs)
	if diffs > 0 {
		t.Fatalf("%d of %d combinations differ", diffs, n)
	}
}

// Part B glue: at each call start the VM makes, the model answers from its own
// stack and loans; the glue gathers only facts (TAKEN FROM VM in the report).

// WMCounts is what an observation saw.
type WMCounts struct {
	Calls, Mismatches, Drift, Unaligned int
	Broken                              int // the glue itself failed
	Deferred                            int // deferred calls, arguments read from their Defer
	Kinds                               map[wmKind]int
	Loans, LentHits                     int
	Samples                             []string
}

type wmObserver struct {
	mu     sync.Mutex
	states map[*Machine]*wmState
	paths  map[PkgID]string
	file   string
	WMCounts
}

// WMObserve runs the model beside every call start until stop; not concurrent.
func WMObserve() (setFile func(string), stop func() WMCounts) {
	o := &wmObserver{states: map[*Machine]*wmState{}, paths: map[PkgID]string{}}
	callStartHook = o.see
	return func(f string) { o.mu.Lock(); o.file = f; o.mu.Unlock() },
		func() WMCounts { callStartHook = nil; return o.WMCounts }
}

func (o *wmObserver) path(id PkgID) string {
	if id.IsZero() {
		return ""
	}
	p, ok := o.paths[id]
	if !ok {
		pkgIDFromPkgPathCache.Range(func(k, v any) bool {
			o.paths[*v.(*PkgID)] = k.(string)
			return true
		})
		if p, ok = o.paths[id]; !ok {
			p = id.String() // never named: still distinct from every path
			o.paths[id] = p
		}
	}
	return p
}

func (o *wmObserver) sample(format string, args ...any) {
	if len(o.Samples) < 20 {
		o.Samples = append(o.Samples, o.file+": "+fmt.Sprintf(format, args...))
	}
}

// wmFrameMode is the mode a frame's caller ran in, from the frame's record of it.
func wmFrameMode(fr *Frame) wmMode {
	switch {
	case fr.LastViewOnly && fr.LastViewRealm != nil && fr.LastViewS3:
		return wmMode{wmByValue, fr.LastViewRealm.Path}
	case fr.LastViewOnly && fr.LastViewRealm != nil:
		return wmMode{wmReadOnly, fr.LastViewRealm.Path}
	case fr.LastViewOnly:
		return wmMode{kind: wmView}
	case fr.LastRealm != nil:
		return wmMode{wmWrite, fr.LastRealm.Path}
	}
	return wmMode{}
}

func (o *wmObserver) see(m *Machine, f *callFacts, cs callStart) {
	o.mu.Lock()
	defer o.mu.Unlock()
	defer func() {
		if r := recover(); r != nil {
			o.Broken++
			o.sample("glue failed: %v", r)
			delete(o.states, m)
		}
	}()
	top := len(m.Frames) - 1
	callee := &m.Frames[top]
	n := 0
	for i := range top {
		if m.Frames[i].Func != nil {
			n++
		}
	}
	s := o.states[m]
	switch {
	case n == 0 || s == nil:
		if n > 0 {
			o.Unaligned++
			o.sample("no model state for %d running calls", n)
		}
		// TAKEN FROM VM: the mode below every call, when a run starts.
		s = &wmState{base: wmFrameMode(callee), baseHold: o.path(callee.callerHold())}
		o.states[m] = s
	case len(s.stack) < n:
		o.Unaligned++
		o.sample("model has %d running calls, VM %d", len(s.stack), n)
		s.base, s.baseHold, s.stack, s.loans = wmFrameMode(callee), o.path(callee.callerHold()), nil, nil
	default:
		s.pop(n)
	}
	if vm := wmVMMode(m); vm != s.mode() {
		o.Drift++
		o.sample("caller mode: model %v, VM %v, calling %s", s.mode(), vm, f.fv.String())
		s.stack = append(s.stack[:0:0], s.stack...)
		if len(s.stack) > 0 {
			s.stack[len(s.stack)-1].mode = vm
		} else {
			s.base = vm
		}
	}
	wf := wmFacts{cross: f.withCross, crossing: f.fv.IsCrossing(), callee: f.pv.PkgPath,
		form: wmFormOf(f.cx), recv: f.recvSet, binder: o.path(f.bindPID),
		maker: o.path(f.fv.GetObjectInfo().ID.PkgID), fn: f.fv, foreign: f.haveForeign && f.foreign}
	if f.last != nil {
		wf.caller = f.last.PkgPath
	}
	if f.pOwned {
		wf.owner = o.path(f.recvPID)
	}
	var asks wmAsk
	if f.haveCalledBy {
		asks |= wmAskBinder
	}
	if f.haveForeign {
		asks |= wmAskForeign
	}
	if f.haveBorrow {
		asks |= wmAskOwner
	}
	got, want := s.start(wf), wmAnswerOf(m, f, cs, asks)
	o.Calls++
	if o.Kinds == nil {
		o.Kinds = map[wmKind]int{}
	}
	o.Kinds[got.mode.kind]++
	if w := s.mode().writer(); w != "" && s.lent(f.fv, w) {
		o.LentHits++
	}
	nl := len(s.loans)
	defer func() { o.Loans += max(0, len(s.loans)-nl) }()
	if got != want {
		o.Mismatches++
		o.sample("%s from %s (form %d): model %+v, VM %+v", f.fv.String(), wf.caller, wf.form, got, want)
	}
	if callee.IsDefer {
		o.Deferred++
	}
	s.enter(wf.callee, got, wmArgsOf(m, f, callee), callee.IsDefer)
}

// wmFormOf reads how the call names its callee, as calledByValue does.
func wmFormOf(cx *CallExpr) wmForm {
	if sx, ok := cx.Func.(*SelectorExpr); ok && sx.Path.IsMethod() {
		if sx.Path.Type == VPInterface || sx.Path.Type == VPDerefInterface {
			return wmIface
		}
		return wmMethod
	}
	switch {
	case cx.GetAttribute(ATTR_NAMED_CALLEE) == true:
		return wmNamed
	case cx.GetAttribute(ATTR_LOCAL_HELPER) == true:
		return wmHelper
	}
	switch fx := cx.Func.(type) {
	case *FuncLitExpr:
		return wmLit
	case *NameExpr:
		if fx.Path.Type != VPBlock {
			return wmBuiltin
		}
	}
	return wmValue
}

// wmReach lists the function values fv reaches through captures, at any depth.
func wmReach(fv *FuncValue) []any {
	var out []any
	seen := map[*FuncValue]bool{fv: true}
	todo := []*FuncValue{fv}
	for len(todo) > 0 {
		cur := todo[len(todo)-1]
		todo = todo[:len(todo)-1]
		for i := range cur.Captures {
			if hiv, ok := cur.Captures[i].V.(*HeapItemValue); ok {
				if cfv, ok := hiv.Value.V.(*FuncValue); ok && !seen[cfv] {
					seen[cfv] = true
					out = append(out, cfv)
					todo = append(todo, cfv)
				}
			}
		}
	}
	return out
}

func wmArgOf(tv TypedValue) wmArg {
	switch v := tv.V.(type) {
	case *FuncValue:
		return wmArg{fn: v, reach: wmReach(v)}
	case *SliceValue:
		x := wmArg{slice: true}
		if av, ok := v.Base.(*ArrayValue); ok {
			for i := v.Offset; i < min(v.Offset+v.Length, len(av.List)); i++ {
				x.elems = append(x.elems, wmArgOf(av.List[i]))
			}
		}
		return x
	}
	return wmArg{}
}

// wmArgsOf reads the arguments the callee will receive, with its parameters'
// marks: from the value stack, or for a deferred call from the Defer just popped.
func wmArgsOf(m *Machine, f *callFacts, fr *Frame) []wmArg {
	var args []wmArg
	if fr.IsDefer {
		cfr := m.MustPeekCallFrame(2)
		d := cfr.Defers
		if cap(d) == len(d) {
			panic("model: the deferred call's Defer is not readable")
		}
		for _, tv := range d[:len(d)+1][len(d)].Args {
			args = append(args, wmArgOf(tv))
		}
		return args
	}
	if f.fv.PkgPath == uversePkgPath && (f.fv.Name == "len" || f.fv.Name == "cap") {
		return nil // the VM hands len and cap their argument without copying it
	}
	vals := m.Values[fr.NumValues+1 : fr.NumValues+1+fr.NumArgs]
	ft := f.fv.GetType(m.Store)
	if fr.Receiver.IsDefined() {
		ft = ft.BoundType()
		args = append(args, wmArgOf(fr.Receiver))
	}
	np := len(ft.Params)
	if ft.HasVarg() && !fr.IsVarg {
		for _, tv := range vals[:np-1] {
			args = append(args, wmArgOf(tv))
		}
		tail := wmArg{slice: true}
		for _, tv := range vals[np-1:] {
			tail.elems = append(tail.elems, wmArgOf(tv))
		}
		args = append(args, tail)
	} else {
		for _, tv := range vals {
			args = append(args, wmArgOf(tv))
		}
	}
	var nonesc []bool
	if fd, ok := f.fv.GetSource(m.Store).(*FuncDecl); ok {
		nonesc, _ = fd.GetAttribute(ATTR_NONESC_PARAMS).([]bool)
	}
	off, coff := len(args)-len(nonesc), 0
	if fr.Receiver.IsDefined() {
		coff = 1
	}
	for i := range args {
		if pi := i - off; pi >= 0 && pi < len(nonesc) {
			args[i].nonesc = nonesc[pi]
		}
		if ci := i - coff; ci >= 0 && ci < len(f.cx.Args) {
			args[i].lit = isLitArg(f.cx.Args[ci])
		}
	}
	return args
}
