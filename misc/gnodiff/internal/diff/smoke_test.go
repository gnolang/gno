package diff

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gnolang/gno/misc/gnodiff/internal/gen"
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
		class, reason := Classify(Program{Source: src}, goRes, gnoRes)
		if class == ClassReal {
			t.Fatalf("seed %d %s\n--- go ---\n%s\n--- gno ---\n%s\n%s", seed, reason, goRes.Summary(), gnoRes.Summary(), src)
		}
	}
}

func TestLoadCorpus(t *testing.T) {
	progs, err := LoadDir(CorpusDir())
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
}
