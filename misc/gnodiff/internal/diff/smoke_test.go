package diff

import (
	"context"
	"strings"
	"testing"
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

func TestLoadCorpus(t *testing.T) {
	progs, err := LoadDir(CorpusDir())
	if err != nil {
		t.Fatal(err)
	}
	if len(progs) < 15 {
		t.Fatalf("corpus has %d programs", len(progs))
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
			if prog.Name != "runcap" {
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
	if known != 1 {
		t.Fatalf("known annotations = %d, want 1 (runcap)", known)
	}
}

func TestReportText(t *testing.T) {
	rep := Report{
		Total:     2,
		Matched:   1,
		Divergent: 1,
		Cases: []CaseResult{
			{Name: "ok", Match: true, Go: Result{Kind: KindOK, Stdout: "1"}, Gno: Result{Kind: KindOK, Stdout: "1"}},
			{
				Name:   "bad",
				Match:  false,
				Reason: "stdout",
				Go:     Result{Kind: KindOK, Stdout: "1"},
				Gno:    Result{Kind: KindOK, Stdout: "2"},
			},
		},
	}
	text := rep.Text()
	if !strings.Contains(text, "DIVERGENCE bad") {
		t.Fatalf("text:\n%s", text)
	}
	if strings.Contains(text, "DIVERGENCE ok") {
		t.Fatalf("matched case leaked into the report:\n%s", text)
	}
	if !rep.Failed() {
		t.Fatal("unexpected divergence should fail the run")
	}
	known := Report{Total: 1, Known: 1, Cases: []CaseResult{{
		Name: "z", Known: "wording", Match: false, Reason: "panic message",
		Go:  Result{Kind: KindPanic, Panic: "a"},
		Gno: Result{Kind: KindPanic, Panic: "b"},
	}}}
	if known.Failed() {
		t.Fatal("annotated divergences are reported and do not fail the run")
	}
}
