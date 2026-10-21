package test

import (
	"io"
	"slices"
	"strings"
	"testing"
)

func TestParseGnoTrim(t *testing.T) {
	t.Parallel()
	got, err := parseGnoTrim(" lines 12, 40-42,7 ")
	if err != nil || !slices.Equal(got.lines, []int{12, 40, 41, 42, 7}) || got.text != "lines 12, 40-42,7" {
		t.Fatalf("got %+v, %v", got, err)
	}
	for in, want := range map[string]string{
		"12":           "wants `lines",
		"lines":        "wants `lines",
		"lines 1-":     "wants `lines",
		"lines 0":      "bad range",
		"lines 5-3":    "bad range",
		"lines 3, 1-4": "line 3 twice",
	} {
		if _, err := parseGnoTrim(in); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%q: err = %v, want %q", in, err, want)
		}
	}
}

func TestGnoTrimApply(t *testing.T) {
	t.Parallel()
	body := []byte("package p\n\nimport \"unsafe\"\n\nvar a = unsafe.Sizeof(0) // ERROR \"x\"\nvar b = 1 // ERROR \"y\"\nvar c complex64\n")
	tr, _ := parseGnoTrim("lines 3, 5")
	out, err := tr.apply(body)
	if err != nil {
		t.Fatal(err)
	}
	// Blanked, not removed: line 6 is still line 6.
	want := "package p\n\n\n\n\nvar b = 1 // ERROR \"y\"\nvar c complex64\n"
	if string(out) != want {
		t.Errorf("apply = %q, want %q", out, want)
	}
	if mk := ParseInlineErrors(out); len(mk) != 1 || mk[0].Line != 6 {
		t.Errorf("markers after trim = %+v, want line 6 only", mk)
	}
	if ls := tr.markedLines(body); !slices.Equal(ls, []int{5}) {
		t.Errorf("markedLines = %v, want [5]", ls)
	}
	for content, want := range map[string]string{
		"lines 8": "has 7 lines",
		"lines 1": "carries no unsafe",
	} {
		tr, _ := parseGnoTrim(content)
		if _, err := tr.apply(body); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%q: err = %v, want %q", content, err, want)
		}
	}
	tr, _ = parseGnoTrim("lines 3, 5")
	if _, err := tr.apply(body[:len(body)-len("var b = 1 // ERROR \"y\"\nvar c complex64\n")]); err == nil || !strings.Contains(err.Error(), "every `// ERROR` marker") {
		t.Errorf("trimming the only marker: err = %v", err)
	}
	// The other three constructs are trimmable too.
	for _, line := range []string{"var c complex64", "x := 2.5i", "p := uintptr(0)", "\t_ = imag(z)"} {
		if trimTag(line) == "" {
			t.Errorf("%q: no trim tag", line)
		}
	}
	if trimTag("\t// a comment about nothing") != "" {
		t.Error("a plain comment is not trimmable")
	}
}

// Sync on a file skipped for an incidental `unsafe`: the trim removes the
// gap, the dropped marker is reported, the surviving marker is checked in
// its own line coordinates, and the written file passes verify.
func TestGnoTrim_ErrorcheckSync(t *testing.T) {
	src := "package p\n\nimport \"unsafe\"\n\nvar x = undeclared // ERROR \"undefined\"\n\nvar _ = unsafe.Sizeof(0) // ERROR \"nope\"\n\n// Verdict: Unsupported unsafe: unknown import path unsafe\n"
	o := errorcheckOpts(t)
	// Without the directive the classifier sees the import: still a skip.
	if _, _, _, err := o.RunFiletest("testdata/errorcheck/trim.go", []byte(src), o.TestStore); err == nil || !strings.Contains(err.Error(), "skipped: unsafe") {
		t.Fatalf("untrimmed: err = %v, want a skip", err)
	}
	withTrim := strings.Replace(src, "\n// Verdict:", "\n// GnoTrim: lines 3, 7\n\n// Verdict:", 1)
	out, _, _, err := o.RunFiletest("testdata/errorcheck/trim.go", []byte(withTrim), o.TestStore)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"\n// GnoTrim: lines 3, 7\n", "\n// GnoTrimmed: lines 7\n",
		"// GnoError:\n// line 5: name not defined: undeclared\n", "// Verdict: Clean errorcheck\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
	if strings.Count(out, "// Verdict:") != 1 || !strings.HasPrefix(out, src[:strings.Index(src, "\n// Verdict:")]) {
		t.Errorf("body or verdict count changed:\n%s", out)
	}
	o.Sync = false
	if _, _, _, err := o.RunFiletest("testdata/errorcheck/trim.go", []byte(out), o.TestStore); err != nil {
		t.Fatalf("verify of the synced file: %v", err)
	}
	// A trim that leaves another gap: the file stays Unsupported, directive kept.
	o.Sync = true
	gap := strings.Replace(withTrim, "var x = undeclared", "var x = make(chan int)", 1)
	out, _, _, err = o.RunFiletest("testdata/errorcheck/trim.go", []byte(gap), o.TestStore)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(out, "\n// GnoTrim: lines 3, 7\n\n// Verdict: Unsupported channel: channels not supported in Gno\n") {
		t.Errorf("gap left by the trim:\n%s", out)
	}
}

// Verify fails loudly on a bad directive, even on a file it would skip.
func TestGnoTrim_VerifyRejectsBadDirective(t *testing.T) {
	opts, m := errorcheckMachine(t, io.Discard)
	m.Release()
	body := "package p\n\nimport \"unsafe\"\n\nvar x = undeclared // ERROR \"undefined\"\n\nvar _ = unsafe.Sizeof(0)\n"
	for name, tc := range map[string]struct{ tail, want string }{
		"line outside the file":      {"// GnoTrim: lines 3, 70\n\n// Verdict: Unsupported unsafe: x\n", "names line 70"},
		"line without the construct": {"// GnoTrim: lines 3, 5\n\n// Verdict: Unsupported unsafe: x\n", "line 5 carries no"},
		"malformed":                  {"// GnoTrim: 3\n\n// Verdict: Unsupported unsafe: x\n", "wants `lines"},
		"stale evidence":             {"// GnoTrim: lines 3, 7\n\n// GnoTrimmed: lines 7\n\n// GnoError:\n// line 5: name not defined: undeclared\n\n// Verdict: Clean errorcheck\n", "stale `// GnoTrimmed:`"},
	} {
		t.Run(name, func(t *testing.T) {
			src := body + "\n" + tc.tail
			_, _, _, err := opts.RunFiletest("testdata/errorcheck/bad.go", []byte(src), opts.TestStore)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want %q", err, tc.want)
			}
		})
	}
}

// A trimmed run file would not be the upstream program.
func TestGnoTrim_RunFileRefused(t *testing.T) {
	opts, m := errorcheckMachine(t, io.Discard)
	m.Release()
	src := "package main\n\nimport \"unsafe\"\n\nvar _ = unsafe.Sizeof(0)\n\nfunc main() {}\n\n// GnoTrim: lines 3, 5\n\n// Verdict: Clean run\n"
	_, _, _, err := opts.RunFiletest("testdata/run/trim.go", []byte(src), opts.TestStore)
	if err == nil || !strings.Contains(err.Error(), "errorcheck and compile files") {
		t.Fatalf("err = %v", err)
	}
}

func TestStripTrailingGoldenRegion_GnoTrimOpensTheRegion(t *testing.T) {
	t.Parallel()
	const body = "package p\n\nvar _ = 1\n"
	src := body + "\n// GnoTrim: lines 3\n\n// Verdict: Unsupported unsafe: x\n"
	if got := upstreamBody(src); got != body {
		t.Errorf("upstreamBody = %q, want %q", got, body)
	}
}
