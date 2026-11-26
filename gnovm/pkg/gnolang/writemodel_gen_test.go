package gnolang

import (
	"fmt"
	"math/rand"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/std"
)

// A seeded generator of small multi-realm programs for TestWriteModelRandom.
// Programs are built from typed pieces: realm 0 is the caller (Main), realms
// 1..nR the callees, libraries l0..; calls go only to later workers (a DAG).

type wgKind uint8

const (
	wgPlain  wgKind = iota // func F<id>()
	wgCross                // func C<id>(cur realm)
	wgMethod               // func (o *ObjT) M<id>()
)

// wgShape is how a call statement reaches its target.
type wgShape uint8

const (
	shCross    wgShape = iota // rB.C(cross(cur))
	shCur                     // C(cur), same realm
	shCurValue                // g := rB.C; g(cur)
	shNamed                   // rB.F()
	shMethod                  // rB.Obj.M()
	shHook                    // rB.H() with var H = F
	shIface                   // var x interface{ M() } = rB.Obj; x.M()
	shLibIface                // l.RunIt(rB.Obj): a library calls Run through an interface
	shHelper                  // h := func() { cb }; h()
	shEach                    // rB.Each(arg): a callback the callee calls
	shLib                     // l.Call/Pass/Catch/Defer(arg)
	shKeep                    // rB.Keep(arg); rB.RunKept(): kept, called after the call
	shGive                    // g := rB.Give(); g(): a returned value
	shCb                      // rB.Cb(cross(cur), arg)
	shBox                     // box.Inc() on a library type the realm owns
	shBoxValue                // mv := box.Inc; mv()
	shBoxer                   // var x interface{ Inc() } = rB.Boxer(); x.Inc()
	shRelay                   // rB.Relay(arg): the callee hands arg on to Keep, then runs it
	shToss                    // rB.Toss(arg) panics with arg; the caller recovers it and runs it
	shOut                     // rB.Out(arg): arg wrapped in a literal handed to the next realm or a library, then run
	shCapt                    // rB.Capt(arg): arg captured by a literal handed to Keep, then run
	shSpread                  // rB.Spread(arg): arg handed to a ...func() helper, then run
	shSlot                    // rB.Slot(arg): arg put in a slice, run from there
	wgShapes
)

var wgShapeNames = [...]string{"cross", "cur", "cur-value", "named", "method", "hook", "iface",
	"lib-iface", "helper", "each", "lib", "keep", "give", "cb", "box", "box-value", "boxer", "relay", "toss", "out", "capt", "spread", "slot"}

// wgArgKind is a callback argument.
type wgArgKind uint8

const (
	argLit    wgArgKind = iota // func() { cb }
	argWrap                    // l.Wrap(func() { cb }): a library closure
	argHold                    // l.Hold(func() { cb }): an escaped literal
	argBox                     // box.Inc: a method value
	argWorker                  // rB.F: a declared function as a value
	argGive                    // rB.Give(): another call's result
	wgArgKinds
)

var wgArgNames = [...]string{"lit", "wrap", "hold", "box", "worker", "give"}

// wgCB is a callback body: one write to its realm, then maybe a named call.
type wgCB struct {
	site   int
	target int // a plain worker called by name, or -1
}

type wgArg struct {
	kind   wgArgKind
	lib    int
	realm  int // argWorker, argGive
	target int // argWorker
	cb     wgCB
}

type wgCall struct {
	shape    wgShape
	realm    int // the target's realm
	target   int // a worker id, or -1
	lib      int
	libFn    string
	arg      *wgArg
	cb       wgCB // shHelper
	deferred bool
	guarded  bool // wrapped in a literal that recovers
}

type wgFunc struct {
	id, realm int
	kind      wgKind
	site      int
	panics    bool
	calls     []*wgCall
}

type wgProg struct {
	seed    int64
	nR, nL  int       // realms after the caller, libraries
	workers []*wgFunc // by id; nil once dropped; 0 is Main
	shapes  [wgShapes]int
}

const wgRoot = "gno.land/r/g/"

func wgRealmName(r int) string {
	if r == 0 {
		return "caller"
	}
	return fmt.Sprintf("r%d", r)
}

func wgRealmPath(r int) string { return wgRoot + wgRealmName(r) }
func wgLibPath(l int) string   { return fmt.Sprintf("gno.land/p/g/l%d", l) }

// Fixed sites of each realm's helpers and each library's Box.Inc.
const (
	hEach = iota
	hKeep
	hRunKept
	hCb
	hGive
	hRun
	hRelay
	hOut
	hCapt
	hSpread
	hEachV
	hSlot
)

func wgHelperSite(r, h int) int { return 10000 + 20*r + h }
func wgBoxSite(l int) int       { return 20000 + l }

// boxLib is the library whose Box realm r holds: realms differ when there are two.
func (p *wgProg) boxLib(r int) int { return r % p.nL }

func (f *wgFunc) name() string {
	switch {
	case f.id == 0:
		return "Main"
	case f.kind == wgCross:
		return fmt.Sprintf("C%d", f.id)
	case f.kind == wgMethod:
		return fmt.Sprintf("M%d", f.id)
	}
	return fmt.Sprintf("F%d", f.id)
}

// wgGenerate builds one program from seed.
func wgGenerate(seed int64) *wgProg {
	rng := rand.New(rand.NewSource(seed))
	p := &wgProg{seed: seed, nR: 2 + rng.Intn(2), nL: 1 + rng.Intn(2)}
	site := 1
	p.workers = append(p.workers, &wgFunc{id: 0, realm: 0, kind: wgCross, site: site})
	for r := 0; r <= p.nR; r++ {
		n := 2 + rng.Intn(3)
		if r == 0 {
			n = rng.Intn(2)
		}
		for range n {
			site++
			f := &wgFunc{id: len(p.workers), realm: r, kind: wgKind(rng.Intn(3)), site: site}
			f.panics = f.kind != wgCross && rng.Intn(4) == 0
			p.workers = append(p.workers, f)
		}
	}
	keepFree := make([]bool, len(p.workers))
	for id := len(p.workers) - 1; id >= 0; id-- {
		f := p.workers[id]
		n := rng.Intn(4)
		if id == 0 {
			n = 4 + rng.Intn(5)
		}
		for range n {
			if c := p.genCall(rng, f, keepFree, &site); c != nil {
				f.calls = append(f.calls, c)
			}
		}
		keepFree[id] = p.keepFree(f, keepFree)
	}
	return p
}

// keepFree reports whether f, and all it reaches, never runs a kept callback.
func (p *wgProg) keepFree(f *wgFunc, kf []bool) bool {
	for _, c := range f.calls {
		if c.shape == shKeep || c.shape == shRelay || c.shape == shCapt {
			return false
		}
		for _, t := range c.refs() {
			if !kf[t] {
				return false
			}
		}
	}
	return true
}

// refs lists the workers a call names.
func (c *wgCall) refs() []int {
	var out []int
	add := func(t int) {
		if t >= 0 {
			out = append(out, t)
		}
	}
	add(c.target)
	add(c.cb.target)
	if c.arg != nil {
		add(c.arg.target)
		add(c.arg.cb.target)
	}
	return out
}

// pick returns a random worker after f in a realm f may reach that ok accepts, or -1.
func (p *wgProg) pick(rng *rand.Rand, f *wgFunc, ok func(*wgFunc) bool) int {
	var cand []int
	for _, g := range p.workers[f.id+1:] {
		if g.realm >= f.realm && ok(g) {
			cand = append(cand, g.id)
		}
	}
	if len(cand) == 0 {
		return -1
	}
	return cand[rng.Intn(len(cand))]
}

// leaf accepts a plain worker that neither panics nor runs a kept callback.
func leaf(kf []bool) func(*wgFunc) bool {
	return func(g *wgFunc) bool { return g.kind == wgPlain && !g.panics && kf[g.id] }
}

func (p *wgProg) genCB(rng *rand.Rand, f *wgFunc, kf []bool, site *int) wgCB {
	*site++
	cb := wgCB{site: *site, target: -1}
	if rng.Intn(2) == 0 {
		cb.target = p.pick(rng, f, leaf(kf))
	}
	return cb
}

func (p *wgProg) genArg(rng *rand.Rand, f *wgFunc, kf []bool, site *int) *wgArg {
	a := &wgArg{kind: wgArgKind(rng.Intn(int(wgArgKinds) + 1)), lib: rng.Intn(p.nL), target: -1}
	if a.kind == wgArgKinds {
		a.kind = argLit // a literal in place is the only kind that gets a loan: draw it twice as often
	}
	a.cb = p.genCB(rng, f, kf, site)
	a.realm = f.realm + rng.Intn(p.nR+1-f.realm)
	if a.kind == argWorker {
		if a.target = p.pick(rng, f, leaf(kf)); a.target < 0 {
			a.kind = argLit
		} else {
			a.realm = p.workers[a.target].realm
		}
	}
	return a
}

func (p *wgProg) genCall(rng *rand.Rand, f *wgFunc, kf []bool, site *int) *wgCall {
	for range 8 {
		c := &wgCall{shape: wgShape(rng.Intn(int(wgShapes))), target: -1, cb: wgCB{target: -1},
			lib: rng.Intn(p.nL), deferred: rng.Intn(4) == 0, guarded: rng.Intn(6) == 0}
		c.realm = f.realm + rng.Intn(p.nR+1-f.realm)
		crossing := f.kind == wgCross
		kind := func(k wgKind) func(*wgFunc) bool { return func(g *wgFunc) bool { return g.kind == k } }
		switch c.shape {
		case shCross, shCurValue, shCb:
			if !crossing {
				continue
			}
			if c.shape == shCurValue && rng.Intn(2) != 0 {
				// Another realm's crossing function called with cur ends the run: keep it rare.
				c.target = p.pick(rng, f, func(g *wgFunc) bool { return g.kind == wgCross && g.realm == f.realm })
			} else if c.shape != shCb {
				c.target = p.pick(rng, f, kind(wgCross))
			}
		case shCur:
			if !crossing {
				continue
			}
			c.target = p.pick(rng, f, func(g *wgFunc) bool { return g.kind == wgCross && g.realm == f.realm })
		case shNamed, shHook:
			c.target = p.pick(rng, f, kind(wgPlain))
		case shMethod, shIface:
			c.target = p.pick(rng, f, kind(wgMethod))
		case shHelper:
			c.cb = p.genCB(rng, f, kf, site)
		case shEach, shLib, shKeep, shRelay, shToss, shOut, shCapt, shSpread, shSlot:
			c.arg = p.genArg(rng, f, kf, site)
		}
		if c.shape == shCb {
			c.arg = p.genArg(rng, f, kf, site)
		}
		switch c.shape {
		case shCross, shCur, shCurValue, shNamed, shHook, shMethod, shIface:
			if c.target < 0 {
				continue
			}
			c.realm = p.workers[c.target].realm
		case shBox, shBoxValue, shHelper, shLib:
			c.realm = f.realm
		}
		if (c.shape == shOut || c.shape == shCapt) && rng.Intn(2) == 0 {
			c.realm = f.realm // the callee writes, so its loan can end by a hand-off
		}
		c.libFn = [...]string{"Call", "Pass", "Catch", "Defer"}[rng.Intn(4)]
		if c.target >= 0 && p.workers[c.target].panics {
			c.guarded = true
		}
		if c.shape == shCurValue && c.realm != f.realm {
			c.guarded = true
		}
		if c.shape == shToss {
			c.deferred, c.guarded = false, true
		}
		p.shapes[c.shape]++
		return c
	}
	return nil
}

// clone deep-copies p for shrinking.
func (p *wgProg) clone() *wgProg {
	q := *p
	q.workers = make([]*wgFunc, len(p.workers))
	for i, f := range p.workers {
		if f == nil {
			continue
		}
		g := *f
		g.calls = make([]*wgCall, len(f.calls))
		for j, c := range f.calls {
			d := *c
			if c.arg != nil {
				a := *c.arg
				d.arg = &a
			}
			g.calls[j] = &d
		}
		q.workers[i] = &g
	}
	return &q
}

// dropWorker removes worker id and every call that names it.
func (p *wgProg) dropWorker(id int) {
	p.workers[id] = nil
	for _, f := range p.workers {
		if f == nil {
			continue
		}
		kept := f.calls[:0]
		for _, c := range f.calls {
			names := false
			for _, t := range c.refs() {
				names = names || t == id
			}
			if !names {
				kept = append(kept, c)
			}
		}
		f.calls = kept
	}
}

// wgSrc renders a program as its memory packages, dependencies first.
func (p *wgProg) packages() []*std.MemPackage {
	var out []*std.MemPackage
	for l := p.nL - 1; l >= 0; l-- {
		out = append(out, wgMemPkg(fmt.Sprintf("l%d", l), wgLibPath(l), p.libSrc(l)))
	}
	for r := p.nR; r >= 0; r-- {
		out = append(out, wgMemPkg(wgRealmName(r), wgRealmPath(r), p.realmSrc(r)))
	}
	return out
}

func wgMemPkg(name, path, body string) *std.MemPackage {
	return &std.MemPackage{Type: MPUserProd, Name: name, Path: path, Files: []*std.MemFile{
		{Name: "gnomod.toml", Body: GenGnoModLatest(path)},
		{Name: name + ".gno", Body: body},
	}}
}

// text is the program as one listing, for failure reports.
func (p *wgProg) text() string {
	var sb strings.Builder
	for _, mp := range p.packages() {
		fmt.Fprintf(&sb, "// ---- %s\n%s\n", mp.Path, mp.Files[1].Body)
	}
	return sb.String()
}

// wgWrite is the write wrapper: it records the try, then the outcome; a
// refused write that still changed probe prints X. owner is a Go expression.
func wgWrite(ind string, site int, wf wgWriteForm, owner, stmt, probe string) string {
	return fmt.Sprintf(`%[1]sfunc() {
%[1]s	println("T", %[2]d, %[3]d, %[4]s)
%[1]s	p0 := %[6]s
%[1]s	defer func() {
%[1]s		if r := recover(); r != nil {
%[1]s			if %[6]s != p0 {
%[1]s				println("X", %[2]d)
%[1]s			}
%[1]s			println("R", %[2]d, r)
%[1]s		}
%[1]s	}()
%[1]s	%[5]s
%[1]s	println("K", %[2]d)
%[1]s}()
`, ind, site, wf, owner, stmt, probe)
}

// wgWriteForm is how a write site changes its realm's state.
type wgWriteForm uint8

const (
	wfVar    wgWriteForm = iota // count++ or kept = f
	wfField                     // Obj.N++
	wfPair                      // count, Obj.N = ...: two targets
	wfMap                       // M[len(M)] = 1: a new key
	wfAppend                    // S = append(S[:1], k), within capacity
	wfCopy                      // copy(S, []int{k})
	wfDelete                    // delete(D, -1)
	wfPtr                       // setN(Obj): through a pointer parameter
	wfFresh                     // a new object stored nowhere: goes through
	wfRecv                      // o.N++ in a method
	wgWriteForms
)

var wgWriteFormNames = [...]string{"variable", "field", "two targets", "map index", "append", "copy", "delete",
	"pointer parameter", "fresh object", "receiver field"}

// wgWriteSrc is the wrapped write at a worker's or callback's site for the realm at path.
func wgWriteSrc(ind string, site int, path string, method bool) string {
	wf := wgWriteForm(site % int(wfRecv))
	if method {
		wf = wfRecv
	}
	sp := map[wgWriteForm][2]string{
		wfVar:    {"count++", "count"},
		wfField:  {"Obj.N++", "Obj.N"},
		wfPair:   {"count, Obj.N = count+1, Obj.N+1", "count + Obj.N"},
		wfMap:    {"M[len(M)] = 1", "len(M)"},
		wfAppend: {fmt.Sprintf("S = append(S[:1], %d)", site), "S[1]"},
		wfCopy:   {fmt.Sprintf("copy(S, []int{%d})", site), "S[0]"},
		wfDelete: {"delete(D, -1)", "len(D)"},
		wfPtr:    {"setN(Obj)", "Obj.N"},
		wfFresh:  {"x := &ObjT{}; x.N = 1", "0"},
		wfRecv:   {"o.N++", "o.N"},
	}[wf]
	owner := strconv.Quote(path)
	if wf == wfFresh {
		owner = `""`
	}
	return wgWrite(ind, site, wf, owner, sp[0], sp[1])
}

func (p *wgProg) libSrc(l int) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "package l%d\n\n", l)
	pass := "Call(f)"
	if l+1 < p.nL {
		fmt.Fprintf(&sb, "import \"%s\"\n\n", wgLibPath(l+1))
		pass = fmt.Sprintf("l%d.Call(f)", l+1)
	}
	sb.WriteString("type Box struct {\n\tTag string\n\tN   int\n}\n\nfunc (b *Box) Inc() {\n")
	sb.WriteString(wgWrite("\t", wgBoxSite(l), wfRecv, "b.Tag", "b.N++", "b.N"))
	sb.WriteString("}\n\n")
	sb.WriteString("func Call(f func()) { f() }\n\n")
	fmt.Fprintf(&sb, "func Pass(f func()) { %s }\n\n", pass)
	sb.WriteString("func Defer(f func()) { defer f() }\n\n")
	sb.WriteString("func Catch(f func()) (r any) {\n\tdefer func() { r = recover() }()\n\tf()\n\treturn nil\n}\n\n")
	sb.WriteString("func Hold(f func()) func() { return f }\n\n")
	sb.WriteString("func Wrap(f func()) func() { return func() { f() } }\n\n")
	sb.WriteString("func RunIt(r interface{ Run() }) { r.Run() }\n")
	return sb.String()
}

func (p *wgProg) realmSrc(r int) string {
	path := wgRealmPath(r)
	var body strings.Builder
	q := strconv.Quote(path)
	w := func(ind string, site int, stmt string) string { return wgWrite(ind, site, wfVar, q, stmt, "0") }
	fmt.Fprintf(&body, "type ObjT struct{ N int }\n\nvar count int\nvar Obj = &ObjT{}\nvar kept func()\nvar box = &l%d.Box{Tag: %q}\n\n", p.boxLib(r), path)
	body.WriteString("var M = map[int]int{}\nvar D = map[int]int{-1: 0}\nvar S = make([]int, 2, 4)\n\nfunc setN(o *ObjT) { o.N = o.N + 1 }\n\n")
	fmt.Fprintf(&body, "func (o *ObjT) Run() {\n%s}\n\n", wgWrite("\t", wgHelperSite(r, hRun), wfRecv, q, "o.N++", "o.N"))
	fmt.Fprintf(&body, "func Each(f func()) {\n%s\tf()\n}\n\n", w("\t", wgHelperSite(r, hEach), "count++"))
	fmt.Fprintf(&body, "func EachV(fs ...func()) {\n%s\tfor _, g := range fs {\n\t\tg()\n\t}\n}\n\n", w("\t", wgHelperSite(r, hEachV), "count++"))
	fmt.Fprintf(&body, "func Keep(f func()) {\n%s}\n\n", w("\t", wgHelperSite(r, hKeep), "kept = f"))
	fmt.Fprintf(&body, "func RunKept() {\n%s\tif kept != nil {\n\t\tkept()\n\t}\n}\n\n", w("\t", wgHelperSite(r, hRunKept), "count++"))
	fmt.Fprintf(&body, "func Cb(cur realm, f func()) {\n%s\tf()\n}\n\n", w("\t", wgHelperSite(r, hCb), "count++"))
	body.WriteString("func Boxer() interface{ Inc() } { return box }\n\n")
	fmt.Fprintf(&body, "func Relay(f func()) {\n%s\tKeep(f)\n\tRunKept()\n}\n\n", w("\t", wgHelperSite(r, hRelay), "count++"))
	body.WriteString("func Toss(f func()) {\n\tpanic(f)\n}\n\n")
	out := fmt.Sprintf("l%d.Call(func() { f() })", p.boxLib(r))
	if r < p.nR {
		out = fmt.Sprintf("%s.Each(func() { f() })\n\t%s", wgRealmName(r+1), out)
	}
	fmt.Fprintf(&body, "func Out(f func()) {\n%s\t%s\n\tf()\n}\n\n", w("\t", wgHelperSite(r, hOut), "count++"), out)
	fmt.Fprintf(&body, "func Capt(f func()) {\n%s\tKeep(func() { f() })\n\tRunKept()\n\tf()\n}\n\n", w("\t", wgHelperSite(r, hCapt), "count++"))
	fmt.Fprintf(&body, "func Spread(f func()) {\n%s\tEachV(f)\n\tf()\n}\n\n", w("\t", wgHelperSite(r, hSpread), "count++"))
	fmt.Fprintf(&body, "func Slot(f func()) {\n%s\ts := []func(){f}\n\ts[0]()\n\tf()\n}\n\n", w("\t", wgHelperSite(r, hSlot), "count++"))
	fmt.Fprintf(&body, "func Give() func() {\n\treturn func() {\n%s\t}\n}\n\n", w("\t\t", wgHelperSite(r, hGive), "count++"))
	hooks := map[int]bool{}
	for _, f := range p.workers {
		if f == nil {
			continue
		}
		for _, c := range f.calls {
			if c.shape == shHook && p.workers[c.target].realm == r {
				hooks[c.target] = true
			}
		}
	}
	var hk []int
	for id := range hooks {
		hk = append(hk, id)
	}
	sort.Ints(hk)
	for _, id := range hk {
		fmt.Fprintf(&body, "var H%d = F%d\n\n", id, id)
	}
	for _, f := range p.workers {
		if f == nil || f.realm != r {
			continue
		}
		switch f.kind {
		case wgCross:
			fmt.Fprintf(&body, "func %s(cur realm) {\n", f.name())
		case wgMethod:
			fmt.Fprintf(&body, "func (o *ObjT) %s() {\n", f.name())
		default:
			fmt.Fprintf(&body, "func %s() {\n", f.name())
		}
		body.WriteString(wgWriteSrc("\t", f.site, path, f.kind == wgMethod))
		for _, c := range f.calls {
			body.WriteString(p.callSrc(f, c))
		}
		if f.panics {
			fmt.Fprintf(&body, "\tpanic(\"p%d\")\n", f.id)
		}
		body.WriteString("}\n\n")
	}
	src := body.String()
	var sb strings.Builder
	fmt.Fprintf(&sb, "package %s\n\nimport (\n", wgRealmName(r))
	for l := 0; l < p.nL; l++ {
		if strings.Contains(src, fmt.Sprintf("l%d.", l)) {
			fmt.Fprintf(&sb, "\t\"%s\"\n", wgLibPath(l))
		}
	}
	for q := r + 1; q <= p.nR; q++ {
		if strings.Contains(src, wgRealmName(q)+".") {
			fmt.Fprintf(&sb, "\t\"%s\"\n", wgRealmPath(q))
		}
	}
	sb.WriteString(")\n\n")
	sb.WriteString(src)
	return sb.String()
}

// pfx names realm b from code of realm a.
func pfx(a, b int) string {
	if a == b {
		return ""
	}
	return wgRealmName(b) + "."
}

func (p *wgProg) cbSrc(f *wgFunc, cb wgCB, ind string) string {
	s := wgWriteSrc(ind+"\t", cb.site, wgRealmPath(f.realm), false)
	if cb.target >= 0 {
		t := p.workers[cb.target]
		s += fmt.Sprintf("%s\t%s%s()\n", ind, pfx(f.realm, t.realm), t.name())
	}
	return "func() {\n" + s + ind + "}"
}

func (p *wgProg) argSrc(f *wgFunc, a *wgArg, ind string) string {
	switch a.kind {
	case argWrap:
		return fmt.Sprintf("l%d.Wrap(%s)", a.lib, p.cbSrc(f, a.cb, ind))
	case argHold:
		return fmt.Sprintf("l%d.Hold(%s)", a.lib, p.cbSrc(f, a.cb, ind))
	case argBox:
		return "box.Inc"
	case argWorker:
		t := p.workers[a.target]
		return pfx(f.realm, t.realm) + t.name()
	case argGive:
		return pfx(f.realm, a.realm) + "Give()"
	}
	return p.cbSrc(f, a.cb, ind)
}

// callSrc renders one call statement of f, deferred or guarded as c says.
func (p *wgProg) callSrc(f *wgFunc, c *wgCall) string {
	ind := "\t"
	if c.guarded {
		ind = "\t\t"
	}
	var pre []string // statements before the (maybe deferred) last call
	var last string
	var t *wgFunc
	if c.target >= 0 {
		t = p.workers[c.target]
	}
	px := pfx(f.realm, c.realm)
	switch c.shape {
	case shCross:
		last = fmt.Sprintf("%s%s(cross(cur))", px, t.name())
	case shCur:
		last = fmt.Sprintf("%s(cur)", t.name())
	case shCurValue:
		pre = append(pre, fmt.Sprintf("g := %s%s", px, t.name()))
		last = "g(cur)"
	case shNamed:
		last = fmt.Sprintf("%s%s()", px, t.name())
	case shMethod:
		last = fmt.Sprintf("%sObj.%s()", px, t.name())
	case shHook:
		last = fmt.Sprintf("%sH%d()", px, t.id)
	case shIface:
		pre = append(pre, fmt.Sprintf("var x interface{ %s() } = %sObj", t.name(), px))
		last = fmt.Sprintf("x.%s()", t.name())
	case shLibIface:
		last = fmt.Sprintf("l%d.RunIt(%sObj)", c.lib, px)
	case shHelper:
		pre = append(pre, "h := "+p.cbSrc(f, c.cb, ind))
		last = "h()"
	case shRelay:
		last = fmt.Sprintf("%sRelay(%s)", px, p.argSrc(f, c.arg, ind))
	case shToss:
		pre = append(pre, "defer func() {\n"+ind+"\tif g, ok := recover().(func()); ok {\n"+ind+"\t\tg()\n"+ind+"\t}\n"+ind+"}()")
		last = fmt.Sprintf("%sToss(%s)", px, p.argSrc(f, c.arg, ind))
	case shOut, shCapt, shSpread, shSlot:
		last = fmt.Sprintf("%s%s(%s)", px, [...]string{shOut: "Out", shCapt: "Capt", shSpread: "Spread", shSlot: "Slot"}[c.shape], p.argSrc(f, c.arg, ind))
	case shEach:
		last = fmt.Sprintf("%sEach(%s)", px, p.argSrc(f, c.arg, ind))
	case shLib:
		last = fmt.Sprintf("l%d.%s(%s)", c.lib, c.libFn, p.argSrc(f, c.arg, ind))
	case shKeep:
		pre = append(pre, fmt.Sprintf("%sKeep(%s)", px, p.argSrc(f, c.arg, ind)))
		last = px + "RunKept()"
	case shGive:
		pre = append(pre, fmt.Sprintf("g := %sGive()", px))
		last = "g()"
	case shCb:
		last = fmt.Sprintf("%sCb(cross(cur), %s)", px, p.argSrc(f, c.arg, ind))
	case shBox:
		last = "box.Inc()"
	case shBoxValue:
		pre = append(pre, "mv := box.Inc")
		last = "mv()"
	case shBoxer:
		pre = append(pre, fmt.Sprintf("var x interface{ Inc() } = %sBoxer()", px))
		last = "x.Inc()"
	}
	switch {
	case c.deferred && (c.shape == shCross || c.shape == shCb):
		last = "defer func() { " + last + " }()" // a cross call cannot be deferred directly
	case c.deferred:
		last = "defer " + last
	}
	var sb strings.Builder
	for _, s := range append(pre, last) {
		sb.WriteString(ind + s + "\n")
	}
	if c.guarded {
		return "\tfunc() {\n\t\tdefer func() {\n\t\t\tif r := recover(); r != nil {\n\t\t\t\tprintln(\"P\", r)\n\t\t\t}\n\t\t}()\n" +
			sb.String() + "\t}()\n"
	}
	return "\t{\n" + strings.ReplaceAll(sb.String(), "\n\t", "\n\t\t") + "\t}\n"
}

// wgNonEsc is, per helper, which parameters the generator wrote so that they
// never let a function value escape: only called, or handed on as a plain
// argument, maybe inside a literal written in place.
var wgNonEsc = map[string][]bool{
	"p.Call": {true}, "p.Pass": {true}, "p.Defer": {true}, "p.Catch": {true},
	"p.Hold": {false}, "p.Wrap": {false}, "p.RunIt": {false},
	"r.Each": {true}, "r.Relay": {true}, "r.Out": {true}, "r.Capt": {true},
	"r.Keep": {false}, "r.Toss": {false}, "r.Cb": {false, true}, "r.Spread": {false},
	"r.EachV": {false}, "r.Slot": {false}, "r.setN": {false},
}

// nonesc answers wgNonEsc for a function of a generated package.
func (p *wgProg) nonesc(pkg, name string) []bool {
	k := "r."
	if strings.HasPrefix(pkg, "gno.land/p/") {
		k = "p."
	}
	return wgNonEsc[k+name]
}

var wgImportLine = regexp.MustCompile(`(?m)^\s*(?:import\s+)?"([^"]+)"\s*$`)

// foreignOf returns the generator's own answer to whether code of package callee,
// called under realm root by package caller, is untrusted: outside root's
// import tree, or through an interface, imported directly by neither.
func (p *wgProg) foreignOf() func(root, caller, callee string, iface bool) bool {
	direct := map[string]map[string]bool{}
	for _, mp := range p.packages() {
		direct[mp.Path] = map[string]bool{}
		for _, f := range mp.Files {
			for _, m := range wgImportLine.FindAllStringSubmatch(f.Body, -1) {
				direct[mp.Path][m[1]] = true
			}
		}
	}
	closure := func(root string) map[string]bool {
		c := map[string]bool{}
		var walk func(string)
		walk = func(q string) {
			for d := range direct[q] {
				if !c[d] {
					c[d] = true
					walk(d)
				}
			}
		}
		walk(root)
		return c
	}
	return func(root, caller, callee string, iface bool) bool {
		switch {
		case callee == uversePkgPath:
			return false
		case !iface:
			return !closure(root)[callee]
		case direct[root][callee]:
			return false
		case caller != root && direct[caller] != nil:
			return !direct[caller][callee]
		}
		return true
	}
}
