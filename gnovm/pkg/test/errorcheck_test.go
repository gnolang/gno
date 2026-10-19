package test

import (
	"bytes"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/store"
	"io"
	"path/filepath"
	"slices"
	"strings"
	"testing"
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
	got := errorcheckGapReason([]int{3, 9}, []int{5}, []int{7}, "accepts the file")
	want := "#6344 uncaught markers: lines 3, 9; over-strict: lines 5; internal errors: lines 7; on-chain path accepts the file"
	if got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
	if got := errorcheckGapReason(nil, nil, nil, ""); got != "#6344" {
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
	src := "package p\n\nvar x = undeclared // ERROR \"undeclared\"\n\n// GnoError:\n// line 3: name undeclared not defined in fileset with files [canary.go]\n\n// Verdict: Clean errorcheck\n"
	if _, _, _, err := opts.RunFiletest("canary.go", []byte(src), opts.TestStore); err != nil {
		t.Fatalf("one verdict: %v", err)
	}
	two := src + "\n// Verdict: KnownIssue errorcheck: #6344 uncaught markers: lines 3\n"
	_, _, _, err := opts.RunFiletest("canary.go", []byte(two), opts.TestStore)
	if err == nil || !strings.Contains(err.Error(), "exactly one") {
		t.Fatalf("two verdicts accepted: %v", err)
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

func TestErrorcheck_ParseBlockedMarkersStayUncaught(t *testing.T) {
	src := []byte(`package p

func f() {
	if { /* ERROR "missing condition" */
	}
	var x int = "s" /* ERROR "cannot use" */
	_ = x
}
`)
	o := errorcheckOpts(t)
	out, _, _, err := o.RunFiletest("testdata/gotypes/check/pb.go", src, o.TestStore)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"// Verdict: KnownIssue errorcheck:",
		"uncaught markers: lines 6",
		"// ParseBlocked:\n// line 6: gc expects: cannot use\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
	for _, bad := range []string{"line 4: gc expects", "on-chain path", "Clean"} {
		if strings.Contains(out, bad) {
			t.Errorf("unexpected %q in:\n%s", bad, out)
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
	_, _, _, err := o.RunFiletest("testdata/gotypes/check/ll.go", src, o.TestStore)
	if err == nil || !strings.Contains(err.Error(), "line-leading marker") {
		t.Fatalf("want a line-leading marker error, got %v", err)
	}
}
