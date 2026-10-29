package test

import (
	"bytes"
	"io"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/store"
)

func TestAttributeErrorcheck(t *testing.T) {
	markers := []InlineError{{Line: 3}, {Line: 5}, {Line: 9}, {Line: 12, GcOnly: true}}
	cov := attributeErrorcheck(markers, []string{
		// One error spanning both marked lines 3 and 5 credits line 3 only.
		"main/a.go:3:2-6:3: duplicate type int in type switch",
		// A function that does not terminate credits its closing brace.
		"main/a.go:7:1-9:2: [function \"f\" does not terminate]",
		// Starts on an unmarked line inside a marked span: over-strict.
		"main/a.go:4:9: cannot use x as int",
		// Internal: credits nothing, listed on its own.
		"main/a.go:12:1-14:2: should not happen",
		// Feature gap: not over-strict.
		"main/a.go:10:1: channel send is not permitted",
	}, 0, 20)
	if !slices.Equal(cov.uncaught, []int{5}) {
		t.Errorf("uncaught = %v, want [5]", cov.uncaught)
	}
	if !slices.Equal(cov.overStrict, []int{4}) {
		t.Errorf("overStrict = %v, want [4]", cov.overStrict)
	}
	if !slices.Equal(cov.internal, []int{12}) {
		t.Errorf("internal = %v, want [12]", cov.internal)
	}
	if cov.gnoErr[3] != "duplicate type int in type switch" || cov.gnoErr[12] != "should not happen" {
		t.Errorf("gnoErr = %v", cov.gnoErr)
	}
	// EOF error, one past the last line, maps to the last line (with the
	// rescue's prepended line taken off).
	cov = attributeErrorcheck([]InlineError{{Line: 4}}, []string{"main/a.go:6:1: expected '}', found 'EOF'"}, 1, 4)
	if len(cov.uncaught) != 0 || cov.gnoErr[4] == "" {
		t.Errorf("EOF: uncaught=%v gnoErr=%v", cov.uncaught, cov.gnoErr)
	}
}

func TestErrorSpan(t *testing.T) {
	for in, want := range map[string][2]int{
		"main/a.go:12:3: msg":                       {12, 12},
		"main/a.go:12:3-9: msg":                     {12, 12},
		"main/a.go:23:3-24:3: msg":                  {23, 24},
		"main/a.go:7: msg":                          {7, 7},
		"main/a.go:3:5'': msg":                      {3, 3},
		"main/a.go:1:1-4:15': initialization cycle": {1, 4},
		"a.go:5:1: x[main/a.go:1:1-2:2].S":          {5, 5},
		"no location here":                          {0, 0},
	} {
		lo, hi := errorSpan(in)
		if lo != want[0] || hi != want[1] {
			t.Errorf("errorSpan(%q) = %d, %d, want %d, %d", in, lo, hi, want[0], want[1])
		}
	}
}

func TestStripPreprocessStack(t *testing.T) {
	in := "main/a.go:10:1-22:2: decl node already predefined! const A:\n--- preprocess stack ---\nstack 0: file main/a.go:10:1-22:2\n------------------------"
	if got := stripPreprocessStack(in); got != "main/a.go:10:1-22:2: decl node already predefined! const A" {
		t.Errorf("got %q", got)
	}
	if got := stripPreprocessStack("plain"); got != "plain" {
		t.Errorf("got %q", got)
	}
}

func TestErrorcheckGapReason(t *testing.T) {
	got := errorcheckGapReason([]int{3, 9}, []int{5}, []int{7}, nil, "accepts the file")
	want := "#6344 uncaught markers: lines 3, 9; over-strict: lines 5; internal errors: lines 7; on-chain path accepts the file"
	if got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
	if got := errorcheckGapReason(nil, nil, nil, nil, ""); got != "#6344" {
		t.Errorf("got %q", got)
	}
	for r, sync := range map[string]bool{
		want:                                   true,
		"#6344":                                true,
		"#6344; uncaught markers: lines 3":     true,
		"#6344 over-strict: lines 5":           true,
		"#1234 wrong receiver type":            false,
		"#6344 is wrong about this, see #1234": false,
	} {
		if isSyncWrittenReason(r) != sync {
			t.Errorf("isSyncWrittenReason(%q) = %v", r, !sync)
		}
	}
	if got := legacySyncReason("#6344; uncaught markers: lines 3"); got != "#6344 uncaught markers: lines 3" {
		t.Errorf("legacy: %q", got)
	}
}

func TestGcOnlyErrorcheckReason_LineDirectives(t *testing.T) {
	// Both directive forms remap the positions the markers refer to, so
	// the file is gc-specific rather than attributed against the wrong
	// line; a directive-free file is checkable.
	markers := []InlineError{{Line: 3, Patterns: []ErrorPattern{{Text: "x"}}}}
	for src, want := range map[string]string{
		"// errorcheck\n\npackage p\n//line :20:1\nvar _ = x // ERROR \"x\"\n":                  "//line",
		"// errorcheck\n\npackage p\nvar _ = /*line :20:1*/x // ERROR \"x\"\n":                  "/*line*/",
		"// errorcheck\n\npackage p\nvar _ = struct{}{ /*line :20:1*/foo: 0 } // ERROR \"x\"\n": "/*line*/",
		"// errorcheck\n\npackage p\nvar _ = x // ERROR \"x\"\n":                                "",
	} {
		got := gcOnlyErrorcheckReason([]byte(src), markers)
		if want == "" {
			if got != "" {
				t.Errorf("%q: got reason %q, want none", src, got)
			}
			continue
		}
		if !strings.HasPrefix(got, "gc-specific: "+want+" directives") {
			t.Errorf("%q: got %q, want a %s reason", src, got, want)
		}
	}
}

func errorcheckMachine(t *testing.T, out io.Writer) (*TestOptions, *gno.Machine) {
	t.Helper()
	opts := NewTestOptions(gnoenv.RootDir(), out, out, nil)
	gasMeter := store.NewInfiniteGasMeter()
	cw := opts.BaseStore.CacheWrap()
	m := gno.NewMachineWithOptions(gno.MachineOptions{
		Output:        out,
		Store:         opts.TestStore.BeginTransaction(cw, cw, nil, gasMeter),
		Context:       Context("", "main", nil),
		MaxAllocBytes: 1 << 30,
		GasMeter:      gasMeter,
		ReviveEnabled: true,
	})
	return opts, m
}

func TestFirstPreprocessErrorDoesNotExecute(t *testing.T) {
	var out bytes.Buffer
	_, m := errorcheckMachine(t, &out)
	defer m.Release()
	src := "package main\n\nvar x = f()\n\nfunc f() int { println(\"EXECUTED\"); return 1 }\n\nfunc init() { println(\"INIT\") }\n\nfunc main() {}\n"
	if msg := firstPreprocessError(m, "main", "main", "a.go", []byte(src)); msg != "" {
		t.Fatalf("accepted file rejected: %q", msg)
	}
	if out.Len() != 0 {
		t.Fatalf("the on-chain probe executed the file: %q", out.String())
	}
	msg := firstPreprocessError(m, "main", "main", "b.go", []byte("package main\n\nvar y = undeclared\n\nfunc main() {}\n"))
	if !strings.Contains(msg, "b.go:3:") || !strings.Contains(msg, "undeclared") {
		t.Fatalf("checking error not reported: %q", msg)
	}
}

func TestErrorcheckVerifyDemandsOneVerdict(t *testing.T) {
	opts, m := errorcheckMachine(t, io.Discard)
	m.Release()
	src := "package p\n\nvar x = undeclared // ERROR \"undeclared\"\n\n// GnoError:\n// line 3: name not defined: undeclared\n\n// Verdict: Clean errorcheck\n"
	if _, _, _, err := opts.RunFiletest("canary.go", []byte(src), opts.TestStore); err != nil {
		t.Fatalf("one verdict: %v", err)
	}
	two := src + "\n// Verdict: KnownIssue errorcheck: #6344 uncaught markers: lines 3\n"
	_, gas, diffs, err := opts.RunFiletest("canary.go", []byte(two), opts.TestStore)
	_, _ = gas, diffs
	if err == nil || !strings.Contains(err.Error(), "exactly one") {
		t.Fatalf("two verdicts accepted: %v", err)
	}
}

// errorcheckOpts is a sync-mode harness over a fresh store: RunFiletest
// returns the file as sync would write it, verdict included.
func errorcheckOpts(t *testing.T) *TestOptions {
	t.Helper()
	rootDir, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	o := &TestOptions{RootDir: rootDir, Output: io.Discard, Error: io.Discard, Sync: true}
	o.BaseStore, o.TestStore = StoreWithOptions(rootDir, o.WriterForStore(), StoreOptions{Testing: true})
	return o
}

func TestErrorcheck_ParseErrorsDoNotBlockTheChecker(t *testing.T) {
	// The parser's error credits line 4; the checker still runs on the
	// partial tree and credits line 6, so the file is Clean.
	src := []byte(`package p

func f() {
	if { // ERROR "missing condition"
	}
	var x int = "s" // ERROR "cannot use"
	_ = x
}
`)
	o := errorcheckOpts(t)
	out, gas, diffs, err := o.RunFiletest("testdata/errorcheck/pb.go", src, o.TestStore)
	_, _ = gas, diffs
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"// Verdict: Clean errorcheck", "line 4: missing condition", "line 6: cannot use"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
	for _, bad := range []string{"!syntax", "uncaught"} {
		if strings.Contains(out, bad) {
			t.Errorf("unexpected %q in:\n%s", bad, out)
		}
	}
}

func TestParseInlineErrors_GccgoMarkerIsOptional(t *testing.T) {
	src := []byte(`package p
var _ = x // GCCGO_ERROR "undefined"
var _ = y // ERROR "undefined"
`)
	got := ParseInlineErrors(src, false)
	if len(got) != 2 || !got[0].GcOnly || got[1].GcOnly {
		t.Fatalf("want an optional marker on line 2 and a required one on line 3, got %+v", got)
	}
	cov := attributeErrorcheck(got, []string{"a.go:2:9: name x not declared"}, 0, 3)
	if len(cov.overStrict) != 0 || !slices.Equal(cov.uncaught, []int{3}) {
		t.Errorf("gccgo line must not be over-strict and line 3 stays uncaught: %+v", cov)
	}
}

func TestErrorcheck_MissingPackageClauseIsPinned(t *testing.T) {
	// The parser's "expected 'package'" error credits gc's marker line;
	// the checker never runs, so there is no on-chain comparison.
	src := []byte(`// errorcheck

func main() { // ERROR "package"
}
`)
	o := errorcheckOpts(t)
	out, gas, diffs, err := o.RunFiletest("testdata/errorcheck/pb.go", src, o.TestStore)
	_, _ = gas, diffs
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"// Verdict: Clean errorcheck", "line 3: expected 'package', found 'func'"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

func TestParseInlineErrors_BlockMarkers(t *testing.T) {
	src := []byte(`package p
var _ = x /* ERROR "undefined: x" */ + y /* ERRORx "undefined.*y" */ // two markers
var _ = a /* ERROR "first" */ + "quoted" /* ERROR "second" */
/* ERROR "leading" */ var _ = 1
// ERROR "own line"
var _ = q // ERROR "line form"
`)
	got := ParseInlineErrors(src, true)
	want := []InlineError{
		{Line: 2, Patterns: []ErrorPattern{{Text: "undefined: x"}, {Text: "undefined.*y", Regex: true}}},
		{Line: 3, Patterns: []ErrorPattern{{Text: "first"}, {Text: "second"}}},
		{Line: 4, Patterns: []ErrorPattern{{Text: "leading"}}, LineLeading: true},
		{Line: 5, Patterns: []ErrorPattern{{Text: "own line"}}, LineLeading: true},
		{Line: 6, Patterns: []ErrorPattern{{Text: "line form"}}},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d markers %+v, want %d", len(got), got, len(want))
	}
	for i := range want {
		if got[i].Line != want[i].Line || got[i].LineLeading != want[i].LineLeading ||
			got[i].GcOnly != want[i].GcOnly || len(got[i].Patterns) != len(want[i].Patterns) {
			t.Fatalf("marker %d: got %+v, want %+v", i, got[i], want[i])
		}
		for j := range want[i].Patterns {
			if got[i].Patterns[j] != want[i].Patterns[j] {
				t.Errorf("marker %d pattern %d: got %+v, want %+v", i, j, got[i].Patterns[j], want[i].Patterns[j])
			}
		}
	}
	// gc's line form is a regexp.
	gc := ParseInlineErrors(src, false)
	if p := gc[4].Patterns[0]; !p.Regex {
		t.Errorf("gc // ERROR pattern should be a regexp: %+v", p)
	}
	if p := gc[0].Patterns[0]; p.Regex {
		t.Errorf("/* ERROR */ pattern is a substring in every file: %+v", p)
	}
}

func TestWordingMatches(t *testing.T) {
	cases := []struct {
		pat  ErrorPattern
		msg  string
		want bool
	}{
		// go/types ERROR literals with regexp metacharacters.
		{ErrorPattern{Text: "cannot assign to m[key].a[0] (neither addressable nor a map index expression)"},
			"cannot assign to m[key].a[0] (neither addressable nor a map index expression)", true},
		{ErrorPattern{Text: "invalid use of [...] array"}, "invalid use of [...] array (outside a composite literal)", true},
		// The same text read as a regexp does not match: the flag matters.
		{ErrorPattern{Text: "invalid use of [...] array", Regex: true}, "invalid use of [...] array (outside a composite literal)", false},
		{ErrorPattern{Text: "undefined.*y", Regex: true}, "undefined: y", true},
		{ErrorPattern{Text: "undefined.*y"}, "undefined: y", false},
		// A regexp that does not compile is tried as a substring.
		{ErrorPattern{Text: "a(b", Regex: true}, "x a(b y", true},
	}
	for _, c := range cases {
		if got := wordingMatches([]ErrorPattern{c.pat}, c.msg); got != c.want {
			t.Errorf("wordingMatches(%+v, %q) = %v, want %v", c.pat, c.msg, got, c.want)
		}
	}
}

func TestErrorcheck_EveryParserErrorCounts(t *testing.T) {
	// go/parser reports both lines; the second must not be dropped as a
	// cascade of the first.
	src := []byte(`package p

func f() {
	if { /* ERROR "missing condition" */
	}
	if x := 1 { /* ERROR "missing condition" */
	}
}
`)
	o := errorcheckOpts(t)
	out, _, _, err := o.RunFiletest("testdata/gotypes/check/pc.go", src, o.TestStore)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "// Verdict: Clean errorcheck") || strings.Contains(out, "ParseBlocked") {
		t.Errorf("want Clean with no ParseBlocked block, got:\n%s", out)
	}
}

func TestErrorcheck_LineLeadingMarkerRefused(t *testing.T) {
	src := []byte(`package p

var _ = undefined
/* ERROR "undefined" */
`)
	o := errorcheckOpts(t)
	_, gas, diffs, err := o.RunFiletest("testdata/gotypes/check/ll.go", src, o.TestStore)
	_, _ = gas, diffs
	if err == nil || !strings.Contains(err.Error(), "line-leading marker") {
		t.Fatalf("want a line-leading marker error, got %v", err)
	}
}

// A bodiless function (gc: assembly-backed) is a per-line gap in
// errorcheck mode: its error is evidence, not over-strictness, and the
// function stays declared, so the call on the marker line is checked
// (its argument folded). A compile file still routes to Unsupported.
func TestErrorcheck_BodilessFuncIsPerLineGap(t *testing.T) {
	o := errorcheckOpts(t)
	src := []byte("package p\n\nfunc f(int)\n\nconst c = 1.5\n\nfunc g() {\n\tf(c) // ERROR \"truncated\"\n\tf(\"s\") // ERROR \"cannot use\"\n}\n")
	out, _, _, err := o.RunFiletest("testdata/errorcheck/bodiless.go", src, o.TestStore)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"// line 3: function f does not have a body", "// line 8: ", "// line 9: ", "// Verdict: Clean errorcheck\n"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
	compile := []byte("// compile\n\npackage p\n\nfunc f(int)\n")
	out, _, _, err = o.RunFiletest("testdata/compile/bodiless.go", compile, o.TestStore)
	if err != nil || !strings.Contains(out, "// Verdict: Unsupported bodiless-func:") {
		t.Fatalf("compile file: err=%v out:\n%s", err, out)
	}
}

func TestParseGapGroups(t *testing.T) {
	reason := "fixed by #1234; #6392 uncaught markers: lines 3, 5-7 (see the note; it has a semicolon); over-strict: lines 9; #6344 uncaught markers: lines 11; on-chain path accepts the file"
	got := parseGapGroups(reason)
	want := []gapGroup{
		{prose: "fixed by #1234"},
		{issue: "#6392", kind: "uncaught markers", lines: []int{3, 5, 6, 7}, note: "(see the note; it has a semicolon)"},
		{issue: "#6392", kind: "over-strict", lines: []int{9}},
		{issue: "#6344", kind: "uncaught markers", lines: []int{11}},
		{issue: "#6344", kind: onChainKind, note: "accepts the file"},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d parts %+v, want %d", len(got), got, len(want))
	}
	for i := range want {
		if got[i].issue != want[i].issue || got[i].kind != want[i].kind || got[i].note != want[i].note || got[i].prose != want[i].prose || !slices.Equal(got[i].lines, want[i].lines) {
			t.Errorf("part %d: got %+v, want %+v", i, got[i], want[i])
		}
	}
	// A round trip keeps the text (ranges expanded); a group names its
	// ref only when it differs from the one in force.
	if s := formatGapGroups(got); s != strings.Replace(reason, "5-7", "5, 6, 7", 1) {
		t.Errorf("format: %q", s)
	}
	// A hand-written group of the umbrella issue, a note or prose is not
	// the sync-written form, which is recomputed wholesale.
	for r, sync := range map[string]bool{
		"#6344 uncaught markers: lines 3; #6392 uncaught markers: lines 5": false,
		"#6344 uncaught markers: lines 3 (why)":                            false,
		"#6344 uncaught markers: lines 3; on-chain path accepts the file":  true,
		"#6344 on-chain path accepts the file":                             true,
		"#63440":                                                           false,
	} {
		if isSyncWrittenReason(r) != sync {
			t.Errorf("isSyncWrittenReason(%q) = %v", r, !sync)
		}
	}
}

func TestReconcileGapReason(t *testing.T) {
	prev := "#6392 uncaught markers: lines 11, 14, 17 (typed overflow); #6344 uncaught markers: lines 38"
	// Verify: a listed line no longer a gap, and a gap listed nowhere.
	reason, stale := reconcileGapReason(prev, []int{17, 38, 40}, []int{5}, nil, nil, "")
	wantStale := []string{"lines 11, 14 are no longer uncaught", "lines 40 are uncaught but unlisted", "lines 5 are over-strict but unlisted"}
	if !slices.Equal(stale, wantStale) {
		t.Errorf("stale = %q, want %q", stale, wantStale)
	}
	// Sync: the group is pruned, its note kept; the unowned uncaught line
	// joins the umbrella group of that kind, the over-strict one is appended.
	if want := "#6392 uncaught markers: lines 17 (typed overflow); #6344 uncaught markers: lines 38, 40; over-strict: lines 5"; reason != want {
		t.Errorf("reason = %q, want %q", reason, want)
	}
	// An emptied group is dropped and the umbrella form appended; prose
	// stays, and the on-chain part is recomputed.
	reason, stale = reconcileGapReason("#1234 wrong receiver; #6392 uncaught markers: lines 11; on-chain path accepts the file", []int{3}, nil, []int{7}, nil, "")
	if want := "#1234 wrong receiver; #6344 uncaught markers: lines 3; internal errors: lines 7"; reason != want {
		t.Errorf("reason = %q, want %q", reason, want)
	}
	if len(stale) != 4 || stale[3] != "the on-chain path part changed" {
		t.Errorf("stale = %q", stale)
	}
	// Up to date: nothing stale, text unchanged.
	if reason, stale = reconcileGapReason(prev, []int{11, 14, 17, 38}, nil, nil, nil, ""); len(stale) != 0 || reason != prev {
		t.Errorf("up to date: reason = %q, stale = %q", reason, stale)
	}
}

// A hand-written verdict's line lists are reconciled with the gaps: verify
// refuses a stale list, sync rewrites it, and the verdict is Clean once
// every marker is caught.
func TestErrorcheck_HandWrittenVerdictReconciled(t *testing.T) {
	o := errorcheckOpts(t)
	o.Sync = false
	// Line 3 is caught now, line 4 is a gap no group lists.
	body := "package p\n\nvar x = undeclared // ERROR \"undeclared\"\nvar y = 1 // ERROR \"nope\"\n\n// GnoError:\n// line 3: name not defined: undeclared\n\n"
	stale := body + "// Verdict: KnownIssue errorcheck: #1234 uncaught markers: lines 3 (why)\n"
	_, err := runFiletest(o, "testdata/errorcheck/hand.go", stale)
	if err == nil || !strings.Contains(err.Error(), "hand-written verdict is stale: lines 3 are no longer uncaught; lines 4 are uncaught but unlisted") {
		t.Fatalf("stale list accepted: %v", err)
	}
	o.Sync = true
	out, err := runFiletest(o, "testdata/errorcheck/hand.go", stale)
	if err != nil || !strings.HasSuffix(out, "// Verdict: KnownIssue errorcheck: #6344 uncaught markers: lines 4\n") {
		t.Fatalf("sync: err=%v out:\n%s", err, out)
	}
	// Every marker caught: Clean, whatever the hand-written lists said.
	clean := "package p\n\nvar x = undeclared // ERROR \"undeclared\"\n\n// GnoError:\n// line 3: name not defined: undeclared\n\n// Verdict: KnownIssue errorcheck: #1234 uncaught markers: lines 3 (why)\n"
	out, err = runFiletest(o, "testdata/errorcheck/hand.go", clean)
	if err != nil || !strings.HasSuffix(out, "// Verdict: Clean errorcheck\n") {
		t.Fatalf("clean: err=%v out:\n%s", err, out)
	}
}

// runFiletest runs one in-memory corpus file and returns the content sync
// would write (empty when verify passes) and the error.
func runFiletest(opts *TestOptions, name, src string) (string, error) {
	out, _, _, err := opts.RunFiletest(name, []byte(src), opts.TestStore)
	return out, err
}
