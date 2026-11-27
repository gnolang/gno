package diff

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gnolang/gno/misc/gnodiff/internal/gen"
	"github.com/gnolang/gno/tm2/pkg/store"
)

func TestSmokeBothSides(t *testing.T) {
	if testing.Short() {
		t.Skip("runs gc and GnoVM")
	}
	src := "package main\n\nfunc main() {\n\tprintln(1 + 1)\n}\n"
	goRes := RunGo(context.Background(), src, Options{})
	gnoRes := RunGno(src, Options{})
	match, reason := Compare(goRes, gnoRes)
	if !match {
		t.Fatalf("%s\n--- go ---\n%s\n--- gno ---\n%s", reason, goRes.Summary(), gnoRes.Summary())
	}
	if goRes.Stdout != "2" {
		t.Fatalf("stdout = %q, want 2", goRes.Stdout)
	}
	if gnoRes.Gas <= 0 {
		t.Fatalf("gno gas = %d, want a recorded charge", gnoRes.Gas)
	}
}

func TestBothReject(t *testing.T) {
	if testing.Short() {
		t.Skip("runs gc and GnoVM")
	}
	src := "package main\n\nfunc main() {\n\tmissing()\n}\n"
	goRes := RunGo(context.Background(), src, Options{})
	gnoRes := RunGno(src, Options{})
	match, reason := Compare(goRes, gnoRes)
	if !match {
		t.Fatalf("%s\n--- go ---\n%s\n--- gno ---\n%s", reason, goRes.Summary(), gnoRes.Summary())
	}
	if goRes.Kind != KindReject || gnoRes.Kind != KindReject {
		t.Fatalf("go %s, gno %s", goRes.Kind, gnoRes.Kind)
	}
}

func TestGenerateSmoke(t *testing.T) {
	if testing.Short() {
		t.Skip("runs gc and GnoVM")
	}
	for _, seed := range []uint64{1, 2} {
		src := gen.Generate(seed)
		goRes := RunGo(context.Background(), src, Options{Timeout: 15 * time.Second})
		gnoRes := RunGno(src, Options{})
		lintRes := RunLint(src, Options{})
		class, reason := Classify(Program{Source: src}, goRes, gnoRes, lintRes)
		if class == ClassReal {
			t.Fatalf("seed %d %s\n--- go ---\n%s\n--- gno ---\n%s\n%s", seed, reason, goRes.Summary(), gnoRes.Summary(), src)
		}
	}
}

func TestLoadCorpus(t *testing.T) {
	progs, err := LoadCorpus()
	if err != nil {
		t.Fatal(err)
	}
	if len(progs) < 250 {
		t.Fatalf("corpus has %d programs, want at least 250", len(progs))
	}
	seen := map[string]bool{}
	known := 0
	for _, prog := range progs {
		if seen[prog.Name] {
			t.Fatalf("duplicate %s", prog.Name)
		}
		seen[prog.Name] = true
		if prog.Known != "" {
			known++
			switch prog.Name {
			case "runcap", "control/unused_local", "const/unsigned_sub":
			default:
				t.Fatalf("unexpected known annotation on %s", prog.Name)
			}
		}
		if !strings.Contains(prog.Source, "package main") || !strings.Contains(prog.Source, "func main(") {
			t.Fatalf("%s is not a main program", prog.Name)
		}
		if why := Unsupported(prog.Source); why != "" {
			t.Fatalf("%s: %s", prog.Name, why)
		}
	}
	if known != 3 {
		t.Fatalf("known annotations = %d, want 3", known)
	}
}

func TestReportText(t *testing.T) {
	rep := Report{
		Total: 2,
		Exact: 1,
		Real:  1,
		Cases: []CaseResult{
			{Name: "ok", Class: ClassExact, Go: Result{Kind: KindOK, Stdout: "1"}, Gno: Result{Kind: KindOK, Stdout: "1"}},
			{
				Name:   "bad",
				Class:  ClassReal,
				Reason: "stdout",
				Go:     Result{Kind: KindOK, Stdout: "1"},
				Gno:    Result{Kind: KindOK, Stdout: "2"},
			},
		},
	}
	text := rep.Text()
	if !strings.Contains(text, "REAL bad") {
		t.Fatalf("text:\n%s", text)
	}
	if strings.Contains(text, "REAL ok") {
		t.Fatalf("exact case leaked into the report:\n%s", text)
	}
	if !rep.Failed() {
		t.Fatal("real divergence should fail the run")
	}
	known := Report{Total: 1, Known: 1, Cases: []CaseResult{{
		Name: "z", Class: ClassKnown, Known: "cap", Reason: "stdout",
		Go:  Result{Kind: KindOK, Stdout: "5 32"},
		Gno: Result{Kind: KindOK, Stdout: "5 5"},
	}}}
	if known.Failed() {
		t.Fatal("known divergences are reported and do not fail the run")
	}
	wording := Report{Total: 1, PanicWording: 1, Cases: []CaseResult{{
		Name: "d", Class: ClassPanicWording, Reason: "panic wording",
		Go:  Result{Kind: KindPanic, Panic: "runtime error: integer divide by zero"},
		Gno: Result{Kind: KindPanic, Panic: "runtime error: division by zero"},
	}}}
	if wording.Failed() {
		t.Fatal("panic wording is reported and does not fail the run")
	}
	if !strings.Contains(wording.wordingText(), "integer divide by zero") {
		t.Fatalf("wording report:\n%s", wording.wordingText())
	}
	// Known-agrees, reject-wording, lint-run, and internal must appear in Text().
	stale := Report{Total: 1, KnownAgrees: 1, Cases: []CaseResult{{
		Name: "stale", Class: ClassKnownAgrees, Known: "old", Reason: "known now agrees",
		Go: Result{Kind: KindOK, Stdout: "1"}, Gno: Result{Kind: KindOK, Stdout: "1"},
	}}}
	if !strings.Contains(stale.Text(), "KNOWN-AGREES stale") {
		t.Fatalf("stale known hidden from report:\n%s", stale.Text())
	}
	internal := Report{Total: 1, Internal: 1, Cases: []CaseResult{{
		Name: "crash", Class: ClassInternal, Reason: "runner failed",
		Go: Result{Kind: KindInternal, Detail: "boom"}, Gno: Result{Kind: KindOK},
	}}}
	if !strings.Contains(internal.Text(), "INTERNAL crash") || !internal.Failed() {
		t.Fatalf("internal report:\n%s", internal.Text())
	}
}

func TestReportWriteSavesUnminimizedRealRepro(t *testing.T) {
	dir := t.TempDir()
	rep := Report{
		Total:    2,
		Real:     1,
		Internal: 1,
		Cases: []CaseResult{
			{
				Name:   "a/b",
				Class:  ClassReal,
				Reason: "stdout",
				Go:     Result{Kind: KindOK, Stdout: "1"},
				Gno:    Result{Kind: KindOK, Stdout: "2"},
				Source: "package main\nfunc main() { println(1) }\n",
			},
			{
				Name:   "crash",
				Class:  ClassInternal,
				Reason: "runner failed",
				Go:     Result{Kind: KindInternal, Detail: "boom"},
				Source: "package main\nfunc main() {}\n",
			},
		},
	}
	if err := rep.Write(dir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "repro", "a__b.go")); err != nil {
		t.Fatalf("expected unminimized repro for real divergence: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "repro", "crash.go")); !os.IsNotExist(err) {
		t.Fatalf("internal runner failure must not write a repro: %v", err)
	}
}

func TestDirectiveParsing(t *testing.T) {
	src := "// gnodiff:known\n// gnodiff:expect-go\n// gnodiff:expect-gno 2\npackage main\n"
	if got := directive(src, "known"); got != "annotated" {
		t.Fatalf("bare known = %q, want annotated", got)
	}
	if got := directive(src, "expect-go"); got != "" {
		t.Fatalf("empty expect-go = %q, want empty", got)
	}
	if got := directive(src, "expect-gno"); got != "2" {
		t.Fatalf("expect-gno = %q, want 2", got)
	}
}

func TestRunGnoNonTermPreservesStdout(t *testing.T) {
	if testing.Short() {
		t.Skip("runs GnoVM")
	}
	src := "package main\n\nfunc main() {\n\tprintln(\"hi\")\n\tfor {\n\t}\n}\n"
	res := RunGno(src, Options{GasLimit: 50_000})
	if res.Kind != KindNonTerm {
		t.Fatalf("kind = %s (%s), want nonterm", res.Kind, res.Detail)
	}
	if res.Stdout != "hi" {
		t.Fatalf("stdout before nonterm = %q, want hi", res.Stdout)
	}
}

func TestResultFromRecoverStages(t *testing.T) {
	// A non-Exception string panic during RunMain is an internal VM failure,
	// never a compile rejection.
	res := resultFromRecover("invalid operation: internal bug", stageRun)
	if res.Kind != KindInternal {
		t.Fatalf("stageRun string panic kind = %s, want internal", res.Kind)
	}
	res = resultFromRecover("invalid operation: bad", stageLoad)
	if res.Kind != KindReject {
		t.Fatalf("stageLoad compile panic kind = %s, want reject", res.Kind)
	}
	res = resultFromRecover(store.OutOfGasError{Descriptor: "CPUCycles"}, stageRun)
	if res.Kind != KindNonTerm {
		t.Fatalf("oog kind = %s, want nonterm", res.Kind)
	}
}

func TestRunGnoPreprocessErrorUnwrapped(t *testing.T) {
	if testing.Short() {
		t.Skip("runs GnoVM")
	}
	src := "package main\n\nfunc main() {\n\tconst x int8 = 128\n\tprintln(x)\n}\n"
	res := RunGno(src, Options{})
	if res.Kind != KindReject {
		t.Fatalf("kind = %s (%s), want reject", res.Kind, res.Detail)
	}
	if strings.Contains(res.Detail, "preprocess stack") {
		t.Fatalf("preprocess stack leaked into reject detail: %q", res.Detail)
	}
}

func TestMinimizeDivergencePreservesKind(t *testing.T) {
	if testing.Short() {
		t.Skip("runs gc and GnoVM")
	}
	// A real runtime stdout divergence with an extra unused line; removing the
	// function header would cause a syntax error on both sides with different
	// wording, which must NOT be accepted as a valid minimization step.
	src := "package main\n\nfunc unused() {}\n\nfunc main() {\n\ts := \"héllo\"\n\trb := []rune(s)\n\tprintln(len(rb), cap(rb) == len(rb))\n}\n"
	opt := Options{Timeout: 15 * time.Second}
	goRes := RunGo(context.Background(), src, opt)
	gnoRes := RunGno(src, opt)
	if goRes.Kind != KindOK || gnoRes.Kind != KindOK || goRes.Stdout == gnoRes.Stdout {
		t.Skip("toolchain agreed on []rune cap")
	}
	reduced := minimizeDivergence(src, goRes, gnoRes, opt, 20)
	if !parses(reduced) {
		t.Fatalf("minimizer broke syntax:\n%s", reduced)
	}
	if strings.Contains(reduced, "unused") {
		t.Fatalf("minimizer kept unused function:\n%s", reduced)
	}
	if !strings.Contains(reduced, "func main()") {
		t.Fatalf("minimizer ate func main():\n%s", reduced)
	}
}
