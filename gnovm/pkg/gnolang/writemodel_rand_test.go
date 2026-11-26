package gnolang

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	storetypes "github.com/gnolang/gno/tm2/pkg/store/types"
)

// TestWriteModelRandom runs seeded random programs (writemodel_gen_test.go) in
// the VM with the model beside it: every call start and every write must agree.

type wgWriteRec struct {
	site      int
	owner     string
	mode      wmMode
	want, got bool
	text      string
	form      wgWriteForm
	worker    bool // the site is a worker's or a callback's, not a helper's
	deepCross bool // under 3 worker calls spanning 2 realms
}

type wgResult struct {
	refused, aborted string // deploy failed; the run ended by an unrecovered panic
	writes           []wgWriteRec
	diffs            []string
	counts           WMCounts
	forms            map[wmForm]int
	borderRecovers   int // a worker's panic recovered in another realm
	lines            int
}

// wgPrint is the model's mode and package of the code that called println.
type wgPrint struct {
	mode  wmMode
	pkg   string
	depth int
	cross bool
}

var wgWorkerName = regexp.MustCompile(`^(Main|[FCM]\d+)$`)

// wgOut reads the program's output one line at a time.
type wgOut struct {
	r       *wgResult
	p       *wgProg
	buf     []byte
	q       []wgPrint
	pending *wgWriteRec
}

func (w *wgOut) Write(b []byte) (int, error) {
	w.buf = append(w.buf, b...)
	for {
		i := bytes.IndexByte(w.buf, '\n')
		if i < 0 {
			return len(b), nil
		}
		w.line(string(w.buf[:i]))
		w.buf = w.buf[i+1:]
	}
}

// wgRefusals are the texts a refused write may carry; a builtin refuses by its own.
var wgRefusals = []string{refuseLibWrite, refuseByValue, refuseUncrossed, refuseTainted}

var wgBuiltinRefusal = map[wgWriteForm]string{wfAppend: refuseAppend, wfCopy: refuseCopy, wfDelete: refuseDelete}

func (w *wgOut) line(s string) {
	w.r.lines++
	if len(w.q) == 0 {
		w.r.diffs = append(w.r.diffs, "output line with no println seen: "+s)
		return
	}
	pr := w.q[0]
	w.q = w.q[1:]
	f := strings.SplitN(s, " ", 3)
	switch f[0] {
	case "T":
		f = strings.SplitN(s, " ", 4)
		site, _ := strconv.Atoi(f[1])
		form, _ := strconv.Atoi(f[2])
		w.pending = &wgWriteRec{site: site, owner: f[3], mode: pr.mode, want: pr.mode.mayWrite(f[3]),
			form: wgWriteForm(form), worker: site < 10000, deepCross: pr.depth >= 3 && pr.cross}
	case "X":
		w.r.diffs = append(w.r.diffs, "refused write changed state: "+s)
	case "K", "R":
		site, _ := strconv.Atoi(f[1])
		wr := w.pending
		w.pending = nil
		if wr == nil || wr.site != site {
			w.r.diffs = append(w.r.diffs, "write outcome with no try: "+s)
			return
		}
		wr.got = f[0] == "K"
		if !wr.got {
			wr.text = f[2]
			known := false
			for _, t := range wgRefusals {
				known = known || strings.Contains(wr.text, t)
			}
			if t, ok := wgBuiltinRefusal[wr.form]; ok {
				known = strings.Contains(wr.text, t)
			}
			if !known {
				w.r.diffs = append(w.r.diffs, fmt.Sprintf("site %d: refused by an unknown text: %s", site, wr.text))
			}
		}
		if wr.got != wr.want {
			w.r.diffs = append(w.r.diffs, fmt.Sprintf("site %d (owner %s): model mode %v says through=%v, VM through=%v %s",
				site, wr.owner, wr.mode, wr.want, wr.got, wr.text))
		}
		w.r.writes = append(w.r.writes, *wr)
	case "P":
		if len(f) > 1 && len(f[1]) > 1 && f[1][0] == 'p' {
			if id, err := strconv.Atoi(f[1][1:]); err == nil && id < len(w.p.workers) && w.p.workers[id] != nil &&
				wgRealmPath(w.p.workers[id].realm) != pr.pkg {
				w.r.borderRecovers++
			}
		}
	}
}

func wgNewStore() Store {
	db := memdb.NewMemDB()
	backend := dbadapter.StoreConstructor(db, storetypes.StoreOptions{})
	return NewStore(nil, backend, backend)
}

// wgRun deploys and runs p once, with the model beside the VM.
func wgRun(p *wgProg, typecheck bool) (res *wgResult) {
	res = &wgResult{forms: map[wmForm]int{}}
	store := wgNewStore()
	for _, mp := range p.packages() {
		if typecheck {
			if _, err := TypeCheckMemPackage(mp, TypeCheckOptions{Getter: store, TestGetter: store, Mode: TCLatestRelaxed}); err != nil {
				res.refused = fmt.Sprintf("type check %s: %v", mp.Path, err)
				return res
			}
		}
		if err := wgDeploy(store, mp); err != "" {
			res.refused = fmt.Sprintf("deploy %s: %s", mp.Path, err)
			return res
		}
	}
	o := &wmObserver{states: map[*Machine]*wmState{}, paths: map[PkgID]string{},
		nonesc: p.nonesc, foreign: p.foreignOf()}
	out := &wgOut{r: res, p: p}
	callStartHook = func(m *Machine, f *callFacts, cs callStart) {
		o.see(m, f, cs)
		res.forms[wmFormOf(f.cx)]++
		if f.fv.PkgPath != uversePkgPath || f.fv.Name != "println" {
			return
		}
		s := o.states[m]
		pr := wgPrint{}
		if s != nil && len(s.stack) >= 2 {
			pr.mode = s.stack[len(s.stack)-2].mode
		} else if s != nil {
			pr.mode = s.base
		}
		if f.last != nil {
			pr.pkg = f.last.PkgPath
		}
		realms := map[string]bool{}
		for i := range m.Frames {
			if fv := m.Frames[i].Func; fv != nil && wgWorkerName.MatchString(string(fv.Name)) {
				pr.depth++
				realms[fv.PkgPath] = true
			}
		}
		pr.cross = len(realms) >= 2
		out.q = append(out.q, pr)
	}
	defer func() { callStartHook = nil; res.counts = o.WMCounts }()
	m := NewMachineWithOptions(MachineOptions{PkgPath: wgRealmPath(0), Store: store, Output: out})
	defer m.Release()
	func() {
		defer func() {
			if r := recover(); r != nil {
				res.aborted = fmt.Sprint(r)
			}
		}()
		x := Call(X("Main"))
		m.MaybeInjectCurForEval(x)
		m.Eval(x)
	}()
	c := o.WMCounts
	if n := c.Mismatches + c.Drift + c.Unaligned + c.Broken; n > 0 {
		res.diffs = append(res.diffs, fmt.Sprintf("call starts: %d mismatches, %d drift, %d unaligned, %d glue failures: %s",
			c.Mismatches, c.Drift, c.Unaligned, c.Broken, strings.Join(c.Samples, " | ")))
	}
	return res
}

func wgDeploy(store Store, mp *std.MemPackage) (err string) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Sprint(r)
		}
	}()
	m := NewMachineWithOptions(MachineOptions{Store: store, Output: io.Discard})
	defer m.Release()
	m.RunMemPackage(mp, true)
	return ""
}

// wgShrink drops workers, then calls, while a difference stays.
func wgShrink(p *wgProg) *wgProg {
	bad := func(q *wgProg) bool { r := wgRun(q, false); return r.refused == "" && len(r.diffs) > 0 }
	for changed := true; changed; {
		changed = false
		for id := len(p.workers) - 1; id > 0; id-- {
			if p.workers[id] == nil {
				continue
			}
			q := p.clone()
			q.dropWorker(id)
			if bad(q) {
				p, changed = q, true
			}
		}
		for id := range p.workers {
			for j := 0; p.workers[id] != nil && j < len(p.workers[id].calls); j++ {
				q := p.clone()
				f := q.workers[id]
				f.calls = append(f.calls[:j:j], f.calls[j+1:]...)
				if bad(q) {
					p, changed = q, true
					j--
				}
			}
		}
	}
	return p
}

func wgEnvInt(name string, def int64) int64 {
	if v := os.Getenv(name); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}
	return def
}

// wgSeeds is the default seed list: 300 programs from one fixed start.
const wgDefaultSeed, wgDefaultN = 80_000, 300

func TestWriteModelRandom(t *testing.T) {
	n := wgEnvInt("WM_RAND_N", wgDefaultN)
	seed0 := wgEnvInt("WM_RAND_SEED", wgDefaultSeed)
	typecheck := os.Getenv("WM_RAND_NOTC") == ""
	start := time.Now()
	var (
		calls, writes, through, refused, diffs, refusedBefore, aborted, border, deep int
		forms                                                                        = map[wmForm]int{}
		kinds                                                                        = map[wmKind]int{}
		shapes                                                                       [wgShapes]int
		deferred, guarded                                                            int
		args                                                                         [wgArgKinds]int
		loans, lentHits, endedHandOff, endedParam                                    int
		wforms                                                                       [wgWriteForms]int
	)
	for i := range n {
		seed := seed0 + i
		p := wgGenerate(seed)
		for _, f := range p.workers {
			for _, c := range f.calls {
				shapes[c.shape]++
				if c.deferred {
					deferred++
				}
				if c.guarded {
					guarded++
				}
				if c.arg != nil {
					args[c.arg.kind]++
				}
			}
		}
		r := wgRun(p, typecheck)
		if r.refused != "" {
			refusedBefore++
			if refusedBefore <= 3 {
				t.Logf("seed %d refused before running: %s\n%s", seed, r.refused, p.text())
			}
			continue
		}
		if r.aborted != "" {
			aborted++
			if os.Getenv("WM_RAND_V") != "" {
				t.Logf("seed %d aborted: %s", seed, r.aborted)
			}
		}
		calls += r.counts.Calls
		loans += r.counts.Loans
		lentHits += r.counts.LentHits
		endedHandOff += r.counts.EndedHandOff
		endedParam += r.counts.EndedParam
		for k, v := range r.counts.Kinds {
			kinds[k] += v
		}
		for k, v := range r.forms {
			forms[k] += v
		}
		border += r.borderRecovers
		for _, w := range r.writes {
			writes++
			wforms[w.form]++
			if w.got {
				through++
			} else {
				refused++
			}
			if w.deepCross {
				deep++
			}
		}
		if len(r.diffs) > 0 {
			diffs++
			t.Errorf("seed %d: %d differences, first: %s\nprogram:\n%s", seed, len(r.diffs), r.diffs[0], p.text())
			if diffs <= 3 {
				small := wgShrink(p.clone())
				rs := wgRun(small, false)
				t.Errorf("seed %d shrunk: %s\nsmall program:\n%s", seed, strings.Join(rs.diffs, "\n"), small.text())
			}
		}
	}
	fmt.Printf("RANDOM: %d programs, %d calls, %d writes, %d differences (seed %d, %s)\n",
		n, calls, writes, diffs, seed0, time.Since(start).Round(time.Second))

	// The reach table: every line must be at least 20 unless marked.
	type row struct {
		name string
		n    int
		min  int
	}
	var rows []row
	formNames := []string{"named", "helper", "lit", "value", "builtin", "method", "iface"}
	for f, name := range formNames {
		need := 20
		if wmForm(f) == wmBuiltin {
			need = 0 // namedCallee marks every builtin call named (preprocess.go)
		}
		rows = append(rows, row{"form " + name, forms[wmForm(f)], need})
	}
	kindNames := map[wmKind]string{wmWrite: "write", wmReadOnly: "read-only", wmByValue: "by-value",
		wmView: "view", wmErrExternalCur: "external-cur error"}
	var kk []int
	for k := range kindNames {
		kk = append(kk, int(k))
	}
	sort.Ints(kk)
	for _, k := range kk {
		rows = append(rows, row{"answer " + kindNames[wmKind(k)], kinds[wmKind(k)], 20})
	}
	// Not reachable from a run: no realm at all (genesis, tooling), and a cross
	// call of a plain function, which the type checker refuses since gno 0.9.
	rows = append(rows, row{"answer none (not reachable)", kinds[wmNone], 0},
		row{"answer not-crossing error (not reachable)", kinds[wmErrNotCrossing], 0})
	for s := range wgShapes {
		rows = append(rows, row{"shape " + wgShapeNames[s], shapes[s], 20})
	}
	for a := range wgArgKinds {
		rows = append(rows, row{"callback " + wgArgNames[a], args[a], 20})
	}
	rows = append(rows,
		row{"deferred calls", deferred, 20},
		row{"guarded calls", guarded, 20},
		row{"panics recovered in another realm", border, 20},
		row{"writes 3 worker calls deep across 2 realms", deep, 20},
		row{"loans made", loans, 20},
		row{"calls that ran lent", lentHits, 20},
		row{"loans ended by hand-off", endedHandOff, 20},
		row{"loans ended by a non-lending parameter", endedParam, 20},
		row{"programs aborted by a panic", aborted, 0},
	)
	for wf := range wgWriteForms {
		rows = append(rows, row{"write form " + wgWriteFormNames[wf], wforms[wf], 20})
	}
	var sb strings.Builder
	for _, r := range rows {
		fmt.Fprintf(&sb, "  %-45s %8d\n", r.name, r.n)
		if r.n < r.min {
			t.Errorf("reach: %s reached %d times, want at least %d", r.name, r.n, r.min)
		}
	}
	fmt.Fprintf(&sb, "  %-45s %8d (%.1f%%)\n", "writes through", through, pct(through, writes))
	fmt.Fprintf(&sb, "  %-45s %8d (%.1f%%)\n", "writes refused", refused, pct(refused, writes))
	fmt.Fprintf(&sb, "  %-45s %8d (%.1f%%)\n", "programs refused before running", refusedBefore, pct(refusedBefore, int(n)))
	fmt.Printf("REACH:\n%s", sb.String())
	if pct(through, writes) < 15 || pct(refused, writes) < 15 {
		t.Errorf("writes: %d through, %d refused; each must be at least 15%%", through, refused)
	}
	if pct(refusedBefore, int(n)) > 20 {
		t.Errorf("%d of %d programs refused before running; at most 20%% allowed", refusedBefore, n)
	}
}

func pct(a, b int) float64 {
	if b == 0 {
		return 0
	}
	return 100 * float64(a) / float64(b)
}
