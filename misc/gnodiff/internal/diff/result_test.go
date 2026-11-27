package diff

import "testing"

func TestNormalizePanicSynonym(t *testing.T) {
	goMsg := "panic: runtime error: invalid memory address or nil pointer dereference"
	gnoMsg := "runtime error: nil pointer dereference"
	if got, want := NormalizePanic(goMsg), NormalizePanic(gnoMsg); got != want {
		t.Fatalf("NormalizePanic() = %q, want %q", got, want)
	}
}

func TestClassifyPanicWording(t *testing.T) {
	goRes := Result{Kind: KindPanic, Stdout: "before", Panic: "runtime error: index out of range [3] with length 2"}
	gnoRes := Result{Kind: KindPanic, Stdout: "before", Panic: "runtime error: slice index out of bounds: 3 (len=2)"}
	class, reason := Classify(Program{}, goRes, gnoRes, Result{})
	if class != ClassPanicWording {
		t.Fatalf("class = %s (%s)", class, reason)
	}
	nilGo := Result{Kind: KindPanic, Stdout: "before", Panic: "runtime error: index out of range [0] with length 0"}
	nilGno := Result{Kind: KindPanic, Stdout: "before", Panic: "runtime error: nil slice index (out of bounds)"}
	class, _ = Classify(Program{}, nilGo, nilGno, Result{})
	if class != ClassPanicWording {
		t.Fatalf("nil slice class = %s", class)
	}
	gnoRes.Panic = "runtime error: slice index out of bounds: 4 (len=2)"
	class, _ = Classify(Program{}, goRes, gnoRes, Result{})
	if class != ClassReal {
		t.Fatalf("different index class = %s", class)
	}
	known := Program{Known: "cap", ExpectGo: "5 32", ExpectGno: "5 5"}
	class, _ = Classify(known,
		Result{Kind: KindOK, Stdout: "5 32"},
		Result{Kind: KindOK, Stdout: "5 5"},
		Result{})
	if class != ClassKnown {
		t.Fatalf("known class = %s", class)
	}
	class, _ = Classify(known,
		Result{Kind: KindOK, Stdout: "5 8"},
		Result{Kind: KindOK, Stdout: "5 5"},
		Result{})
	if class != ClassReal {
		t.Fatalf("drifted known class = %s, want real", class)
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
	rej := Result{Kind: KindReject, Detail: "./main.go:3:2: undefined: missing"}
	rejSame := Result{Kind: KindReject, Detail: "main.gno:3:2: undefined: missing"}
	if match, reason := Compare(rej, rejSame); !match {
		t.Fatalf("same reject diagnostic: %s", reason)
	}
	rejOther := Result{Kind: KindReject, Detail: "main.gno:1:1: undefined: foo"}
	if match, _ := Compare(rej, rejOther); match {
		t.Fatal("different reject diagnostics matched")
	}
	overflow := Result{Kind: KindReject, Detail: "./main.go:1:9: overflows uint64"}
	otherErr := Result{Kind: KindReject, Detail: "main.gno:1:1: undefined: foo"}
	if match, _ := Compare(overflow, otherErr); match {
		t.Fatal("overflow vs undefined matched")
	}
	empty := Result{Kind: KindNonTerm, Detail: "timed out"}
	if match, _ := Compare(empty, empty); !match {
		t.Fatal("both-empty nonterm should agree")
	}
	printed := Result{Kind: KindNonTerm, Stdout: "hi", Detail: "timed out"}
	printedOther := Result{Kind: KindNonTerm, Stdout: "yo", Detail: "timed out"}
	if match, _ := Compare(printed, printedOther); match {
		t.Fatal("different stdout before nonterm matched")
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
	prog, msg = splitGoStderr("panic: hello\n")
	if msg != "" {
		t.Fatalf("println of panic prefix classified as a panic: %q", msg)
	}
	if prog != "panic: hello" {
		t.Fatalf("prog = %q", prog)
	}
}

func TestPanicFoldsKeepOperands(t *testing.T) {
	class, _ := Classify(Program{},
		Result{Kind: KindPanic, Panic: "runtime error: negative shift amount"},
		Result{Kind: KindPanic, Panic: "runtime error: negative shift amount: (-1 int)"},
		Result{})
	if class != ClassPanicWording {
		t.Fatalf("gc negative shift vs numbered gno = %s", class)
	}
	class, _ = Classify(Program{},
		Result{Kind: KindPanic, Panic: "runtime error: negative shift amount: (-1 int)"},
		Result{Kind: KindPanic, Panic: "runtime error: negative shift amount: (-2 int)"},
		Result{})
	if class != ClassReal {
		t.Fatalf("different shift amounts class = %s", class)
	}
	if NormalizePanic("runtime error: negative shift amount: (-1 int)") ==
		NormalizePanic("runtime error: negative shift amount: (-2 int)") {
		t.Fatal("shift amounts were folded away")
	}
	class, _ = Classify(Program{},
		Result{Kind: KindPanic, Panic: "runtime error: index out of range [-5] with length 0"},
		Result{Kind: KindPanic, Panic: "runtime error: nil slice index (out of bounds)"},
		Result{})
	if class != ClassReal {
		t.Fatalf("nil slice with a non-zero index class = %s", class)
	}
}

func TestOverflowRejectWording(t *testing.T) {
	goRes := Result{Kind: KindReject, Detail: "./main.go:4:17: cannot use 128 (untyped int constant) as int8 value in constant declaration (overflows)"}
	gnoRes := Result{Kind: KindReject, Detail: "main.gno:4:8-20: bigint overflows target kind"}
	class, _ := Classify(Program{}, goRes, gnoRes, Result{})
	if class != ClassRejectWording {
		t.Fatalf("overflow wording class = %s", class)
	}
	class, _ = Classify(Program{},
		Result{Kind: KindReject, Detail: "./main.go:5:16: constant -2147483649 overflows int32"},
		Result{Kind: KindReject, Detail: "main.gno:5:10-28: bigint underflows target kind"},
		Result{})
	if class != ClassRejectWording {
		t.Fatalf("underflow wording class = %s", class)
	}
	class, _ = Classify(Program{},
		Result{Kind: KindReject, Detail: "overflows uint64"},
		Result{Kind: KindReject, Detail: "undefined: foo"},
		Result{})
	if class != ClassReal {
		t.Fatalf("overflow vs undefined class = %s", class)
	}
	class, _ = Classify(Program{},
		Result{Kind: KindReject, Detail: "cannot use 1 as int8 value (overflows)"},
		Result{Kind: KindReject, Detail: "cannot use 1 as int64 value (overflows)"},
		Result{})
	if class != ClassReal {
		t.Fatalf("different overflow types class = %s", class)
	}
}

func TestLintRunAndStaleKnown(t *testing.T) {
	class, reason := Classify(Program{ExpectGo: "reject: declared and not used", ExpectGno: "2"},
		Result{Kind: KindReject, Detail: "declared and not used: x"},
		Result{Kind: KindOK, Stdout: "2"},
		Result{Kind: KindReject, Detail: "x declared and not used"})
	if class != ClassLintRun {
		t.Fatalf("class = %s (%s)", class, reason)
	}
	class, _ = Classify(Program{ExpectGno: "9"},
		Result{Kind: KindReject, Detail: "overflows uint64"},
		Result{Kind: KindOK, Stdout: "2"},
		Result{Kind: KindReject, Detail: "overflows uint64"})
	if class != ClassReal {
		t.Fatalf("drifted lint-run class = %s", class)
	}
	class, reason = Classify(Program{Known: "old"},
		Result{Kind: KindOK, Stdout: "1"},
		Result{Kind: KindOK, Stdout: "1"},
		Result{})
	if class != ClassKnownAgrees || reason != "known now agrees" {
		t.Fatalf("stale known = %s (%s)", class, reason)
	}
}
