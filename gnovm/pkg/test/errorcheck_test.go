package test

import (
	"io"
	"path/filepath"
	"strings"
	"testing"
)

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
