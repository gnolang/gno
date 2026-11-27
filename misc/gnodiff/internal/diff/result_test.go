package diff

import "testing"

func TestNormalizePanicSynonym(t *testing.T) {
	goMsg := "panic: runtime error: invalid memory address or nil pointer dereference"
	gnoMsg := "runtime error: nil pointer dereference"
	if got, want := NormalizePanic(goMsg), NormalizePanic(gnoMsg); got != want {
		t.Fatalf("NormalizePanic() = %q, want %q", got, want)
	}
}

func TestCanonicalRuntimePanics(t *testing.T) {
	pairs := [][2]string{
		{"runtime error: integer divide by zero", "runtime error: division by zero"},
		{
			"runtime error: index out of range [3] with length 2",
			"runtime error: slice index out of bounds: 3 (len=2)",
		},
		{
			"runtime error: negative shift amount",
			"runtime error: negative shift amount: (-1 int)",
		},
		{
			"runtime error: slice bounds out of range [2:1]",
			"runtime error: invalid slice index 2 > 1",
		},
	}
	for _, pair := range pairs {
		if got, want := NormalizePanic(pair[0]), NormalizePanic(pair[1]); got != want {
			t.Fatalf("NormalizePanic(%q) = %q, NormalizePanic(%q) = %q", pair[0], got, pair[1], want)
		}
	}
	// Same family, different index: still a divergence.
	if NormalizePanic("runtime error: index out of range [3] with length 2") ==
		NormalizePanic("runtime error: slice index out of bounds: 4 (len=2)") {
		t.Fatal("different indexes were treated as the same panic")
	}
}

func TestNormalizePanicKeepsIndexes(t *testing.T) {
	a := NormalizePanic("runtime error: index out of range [3] with length 1")
	b := NormalizePanic("runtime error: index out of range [2] with length 1")
	if a == b {
		t.Fatalf("distinct indexes collapsed to %q", a)
	}
}

func TestNormalizeOutput(t *testing.T) {
	got := NormalizeOutput("a \r\nptr 0xcafebabe\n\n")
	want := "a\nptr 0xPTR"
	if got != want {
		t.Fatalf("NormalizeOutput() = %q, want %q", got, want)
	}
}

func TestCompareOutcomes(t *testing.T) {
	ok := Result{Kind: KindOK, Stdout: "1"}
	other := Result{Kind: KindOK, Stdout: "2"}
	if match, reason := Compare(ok, ok); !match || reason != "" {
		t.Fatalf("identical ok: match=%v reason=%q", match, reason)
	}
	if match, _ := Compare(ok, other); match {
		t.Fatal("different stdout matched")
	}
	rej := Result{Kind: KindReject, Detail: "nope"}
	rej2 := Result{Kind: KindReject, Detail: "different wording"}
	if match, _ := Compare(rej, rej2); !match {
		t.Fatal("both-reject should agree without comparing text")
	}
	pan := Result{Kind: KindPanic, Panic: "runtime error: nil pointer dereference", Stdout: "before"}
	panGo := Result{Kind: KindPanic, Panic: "runtime error: invalid memory address or nil pointer dereference", Stdout: "before"}
	if match, reason := Compare(panGo, pan); !match {
		t.Fatalf("synonym panics: %s", reason)
	}
	if match, _ := Compare(ok, pan); match {
		t.Fatal("ok vs panic matched")
	}
	internal := Result{Kind: KindInternal, Detail: "exec"}
	if match, _ := Compare(internal, internal); match {
		t.Fatal("internal errors must not count as agreement")
	}
}

func TestSplitGoStderr(t *testing.T) {
	stderr := "before\npanic: runtime error: integer divide by zero\n\ngoroutine 1 [running]:\nmain.main()\n\t/tmp/main.go:4 +0x20\nexit status 2\n"
	prog, msg := splitGoStderr(stderr)
	if prog != "before" {
		t.Fatalf("prog = %q", prog)
	}
	if msg != "runtime error: integer divide by zero" {
		t.Fatalf("panic = %q", msg)
	}
}
