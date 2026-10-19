package test

import (
	"bytes"
	"io"
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
