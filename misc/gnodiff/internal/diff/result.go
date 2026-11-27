// Package diff runs the same small Go program on the standard toolchain and
// on GnoVM and reports observable disagreements.
package diff

import (
	"fmt"
	"regexp"
	"strings"
)

// Kind is the coarse outcome of one run. Message text is compared only for
// KindOK and KindPanic; KindReject and KindNonTerm match on the kind alone,
// because compiler wording and resource-limit timing are not the language
// behavior under test.
type Kind string

const (
	KindOK       Kind = "ok"
	KindPanic    Kind = "panic"
	KindReject   Kind = "reject"
	KindNonTerm  Kind = "nonterm"
	KindInternal Kind = "internal"
)

// Result is one side of a differential run.
type Result struct {
	Kind      Kind   `json:"kind"`
	Stdout    string `json:"stdout,omitempty"`
	Panic     string `json:"panic,omitempty"`
	Detail    string `json:"detail,omitempty"`
	Gas       int64  `json:"gas,omitempty"`
	Truncated bool   `json:"truncated,omitempty"`
}

// Summary is a short multi-line description for reports.
func (r Result) Summary() string {
	switch r.Kind {
	case KindOK:
		if r.Stdout == "" {
			return "ok (empty output)"
		}
		return "ok\n" + r.Stdout
	case KindPanic:
		msg := "panic: " + r.Panic
		if r.Stdout != "" {
			return "output before panic:\n" + r.Stdout + "\n" + msg
		}
		return msg
	case KindReject:
		return "reject: " + oneLine(r.Detail)
	case KindNonTerm:
		return "nonterm: " + oneLine(r.Detail)
	default:
		return "internal: " + oneLine(r.Detail)
	}
}

func oneLine(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	const limit = 240
	if len(s) > limit {
		return s[:limit] + "..."
	}
	return s
}

func clip(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "\n...[truncated]"
}

// panicSynonyms rewrites wording that both toolchains use for the same
// runtime event. The longer phrase is replaced with the shorter one.
// Behavioral differences (different indexes, one side not panicking) stay
// visible because the rest of the message is compared verbatim.
var panicSynonyms = [][2]string{
	{"invalid memory address or nil pointer dereference", "nil pointer dereference"},
}

var ptrRE = regexp.MustCompile(`0x[0-9a-fA-F]+`)
var spaceRE = regexp.MustCompile(`[ \t]+`)

// NormalizeOutput canonicalizes program output: newline style, trailing
// whitespace, a single trimmed end, and pointer addresses. Addresses are
// masked because println of a pointer is process-specific on both sides;
// programs that want to compare identity should print a boolean.
func NormalizeOutput(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = ptrRE.ReplaceAllString(s, "0xPTR")
	lines := strings.Split(s, "\n")
	for i, ln := range lines {
		lines[i] = strings.TrimRight(ln, " \t")
	}
	return strings.TrimRight(strings.Join(lines, "\n"), "\n")
}

// cleanPanic is the panic text with only whitespace and a leading "panic:"
// removed. It does not fold wording.
func cleanPanic(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "panic: ")
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	s = spaceRE.ReplaceAllString(strings.TrimSpace(s), " ")
	return strings.TrimSpace(s)
}

// NormalizePanic reduces a panic value to the event it names. Wording for the
// same event is folded, and indexes are kept, so a wrong bound stays distinct.
func NormalizePanic(s string) string {
	s = cleanPanic(s)
	folded := s
	for _, pair := range panicSynonyms {
		folded = strings.ReplaceAll(folded, pair[0], pair[1])
	}
	if canon, ok := canonicalRuntimePanic(folded); ok {
		return canon
	}
	return folded
}

// Class is how a pair of results relates.
type Class string

const (
	// ClassExact is the same outcome and the same text.
	ClassExact Class = "exact"
	// ClassKnown is a documented difference (// gnodiff:known), and the
	// outputs still match that seed's expect lines when it has them.
	ClassKnown Class = "known"
	// ClassPanicWording is the same panic event (kind, indexes, output
	// before the panic) with different message text.
	ClassPanicWording Class = "panic-wording"
	// ClassReal is a behavioral disagreement, or a known seed whose output
	// no longer matches the expect lines.
	ClassReal Class = "real"
)

// Classify compares one gc result with one GnoVM result.
func Classify(prog Program, goRes, gnoRes Result) (Class, string) {
	if goRes.Kind == KindInternal || gnoRes.Kind == KindInternal {
		return ClassReal, "runner failed"
	}
	if goRes.Truncated || gnoRes.Truncated {
		if goRes.Truncated != gnoRes.Truncated || goRes.Stdout != gnoRes.Stdout || goRes.Kind != gnoRes.Kind {
			return knownOrReal(prog, goRes, gnoRes, "output truncated")
		}
	}
	if goRes.Kind != gnoRes.Kind {
		return knownOrReal(prog, goRes, gnoRes, fmt.Sprintf("outcome %s vs %s", goRes.Kind, gnoRes.Kind))
	}
	switch goRes.Kind {
	case KindOK:
		if goRes.Stdout == gnoRes.Stdout {
			return ClassExact, ""
		}
		return knownOrReal(prog, goRes, gnoRes, "stdout")
	case KindPanic:
		if goRes.Stdout != gnoRes.Stdout {
			return knownOrReal(prog, goRes, gnoRes, "stdout before panic")
		}
		if cleanPanic(goRes.Panic) == cleanPanic(gnoRes.Panic) {
			return ClassExact, ""
		}
		if NormalizePanic(goRes.Panic) == NormalizePanic(gnoRes.Panic) || nilSliceWording(goRes.Panic, gnoRes.Panic) {
			return ClassPanicWording, "panic wording"
		}
		return knownOrReal(prog, goRes, gnoRes, "panic message")
	case KindReject, KindNonTerm:
		return ClassExact, ""
	default:
		return ClassReal, "unknown outcome"
	}
}

func knownOrReal(prog Program, goRes, gnoRes Result, reason string) (Class, string) {
	if prog.Known == "" {
		return ClassReal, reason
	}
	if prog.ExpectGo != "" && !expectMatch(goRes, prog.ExpectGo) {
		return ClassReal, "known go output changed"
	}
	if prog.ExpectGno != "" && !expectMatch(gnoRes, prog.ExpectGno) {
		return ClassReal, "known gno output changed"
	}
	return ClassKnown, reason
}

func expectMatch(res Result, want string) bool {
	if after, ok := strings.CutPrefix(want, "reject:"); ok {
		return res.Kind == KindReject && strings.Contains(res.Detail, strings.TrimSpace(after))
	}
	if after, ok := strings.CutPrefix(want, "panic:"); ok {
		return res.Kind == KindPanic && strings.Contains(cleanPanic(res.Panic), strings.TrimSpace(after))
	}
	want = NormalizeOutput(want)
	switch res.Kind {
	case KindOK:
		return res.Stdout == want
	case KindPanic:
		got := NormalizeOutput(res.Stdout)
		if got != "" {
			got += "\n"
		}
		got += cleanPanic(res.Panic)
		return NormalizeOutput(got) == want || cleanPanic(res.Panic) == want
	default:
		return oneLine(res.Detail) == want || res.Detail == want
	}
}

var (
	indexGoRE   = regexp.MustCompile(`index out of range \[(-?\d+)\] with length (-?\d+)`)
	indexGnoRE  = regexp.MustCompile(`slice index out of bounds: (-?\d+) \(len=(-?\d+)\)`)
	boundsGoRE  = regexp.MustCompile(`slice bounds out of range \[(-?\d+):(-?\d+)\]`)
	boundsGnoRE = regexp.MustCompile(`invalid slice index (-?\d+) > (-?\d+)`)
)

// canonicalRuntimePanic maps panic texts that name the same event. The
// numbers are kept: index 3 with length 2 must not match index 4.
func canonicalRuntimePanic(s string) (string, bool) {
	switch {
	case strings.Contains(s, "divide by zero"), strings.Contains(s, "division by zero"):
		return "division by zero", true
	case strings.Contains(s, "negative shift"):
		return "negative shift", true
	}
	for _, spec := range []struct {
		re  *regexp.Regexp
		fmt string
	}{
		{indexGoRE, "index %s len %s"},
		{indexGnoRE, "index %s len %s"},
		{boundsGoRE, "slice bounds %s:%s"},
		{boundsGnoRE, "slice bounds %s:%s"},
	} {
		if m := spec.re.FindStringSubmatch(s); m != nil {
			return fmt.Sprintf(spec.fmt, m[1], m[2]), true
		}
	}
	return "", false
}

// nilSliceWording is the nil-slice index panic. gc reports it as an ordinary
// index panic with length 0; Gno names the nil slice and does not repeat the
// index. Both sides panicked on the same operation.
func nilSliceWording(goMsg, gnoMsg string) bool {
	if !strings.Contains(cleanPanic(gnoMsg), "nil slice index") {
		return false
	}
	m := indexGoRE.FindStringSubmatch(cleanPanic(goMsg))
	return m != nil && m[2] == "0"
}

// Compare reports whether the results are free of a real behavioral
// divergence. Panic-wording differences count as agreement here; Classify
// records them separately.
func Compare(goRes, gnoRes Result) (match bool, reason string) {
	cl, reason := Classify(Program{}, goRes, gnoRes)
	return cl == ClassExact || cl == ClassPanicWording, reason
}
