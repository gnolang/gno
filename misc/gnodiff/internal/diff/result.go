// Package diff runs the same small Go program on the standard toolchain and
// on GnoVM and reports observable disagreements.
package diff

import (
	"fmt"
	"regexp"
	"strings"
)

// Kind is the coarse outcome of one run. Successful runs and panics compare
// text. Two rejections compare a path-stripped diagnostic. Two
// non-terminations compare stdout when either side printed something.
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
	{"runtime error: invalid memory address or nil pointer dereference", "runtime error: nil pointer dereference"},
}

var ptrRE = regexp.MustCompile(`\b0x[0-9a-fA-F]{6,16}\b`)
var spaceRE = regexp.MustCompile(`[ \t]+`)

// NormalizeOutput canonicalizes program output: newline style, trailing
// whitespace, a single trimmed end, and 6- to 16-digit hex pointer addresses.
// Addresses are masked because println of a pointer is process-specific on
// both sides; shorter hex values (including 0x0 for nil) are preserved so
// nil and non-nil pointers do not collapse together.
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
	// ClassLintRun means gc and gno lint both reject the program, and the
	// VM (gno run) still executes it.
	ClassLintRun Class = "lint-run"
	// ClassRejectWording means both sides reject the same constant overflow.
	// The VM says "bigint overflows"; gc names the type. A different error
	// (overflows versus undefined) is real. Two named types that differ are real.
	ClassRejectWording Class = "reject-wording"
	// ClassKnownAgrees means a // gnodiff:known seed now matches. The
	// annotation is stale.
	ClassKnownAgrees Class = "known-agrees"
	// ClassInternal means one of the runners (gc, GnoVM, or lint) failed
	// internally rather than producing a language result.
	ClassInternal Class = "internal"
)

// Classify compares one gc result with one GnoVM result. lintRes is the
// go/types pass gno lint uses. A zero Result means that pass was not run.
func Classify(prog Program, goRes, gnoRes, lintRes Result) (Class, string) {
	if lintRes.Kind == KindInternal {
		return ClassInternal, "lint runner failed"
	}
	if goRes.Kind == KindInternal || gnoRes.Kind == KindInternal {
		return ClassInternal, "runner failed"
	}
	if goRes.Truncated || gnoRes.Truncated {
		if goRes.Truncated != gnoRes.Truncated || goRes.Stdout != gnoRes.Stdout || goRes.Kind != gnoRes.Kind {
			return knownOrReal(prog, goRes, gnoRes, "output truncated")
		}
	}
	if class, reason, ok := lintRun(prog, goRes, gnoRes, lintRes); ok {
		return class, reason
	}
	if goRes.Kind != gnoRes.Kind {
		return knownOrReal(prog, goRes, gnoRes, fmt.Sprintf("outcome %s vs %s", goRes.Kind, gnoRes.Kind))
	}
	switch goRes.Kind {
	case KindOK:
		if goRes.Stdout == gnoRes.Stdout {
			return exactOrStale(prog)
		}
		return knownOrReal(prog, goRes, gnoRes, "stdout")
	case KindPanic:
		if goRes.Stdout != gnoRes.Stdout {
			return knownOrReal(prog, goRes, gnoRes, "stdout before panic")
		}
		if cleanPanic(goRes.Panic) == cleanPanic(gnoRes.Panic) {
			return exactOrStale(prog)
		}
		if panicWording(goRes.Panic, gnoRes.Panic) {
			return ClassPanicWording, "panic wording"
		}
		return knownOrReal(prog, goRes, gnoRes, "panic message")
	case KindReject:
		if normalizeDiag(goRes.Detail) == normalizeDiag(gnoRes.Detail) {
			return exactOrStale(prog)
		}
		if overflowWording(goRes.Detail, gnoRes.Detail) {
			return ClassRejectWording, "reject wording"
		}
		return knownOrReal(prog, goRes, gnoRes, "reject diagnostic")
	case KindNonTerm:
		if goRes.Stdout == "" && gnoRes.Stdout == "" || goRes.Stdout == gnoRes.Stdout {
			return exactOrStale(prog)
		}
		return knownOrReal(prog, goRes, gnoRes, "stdout before nonterm")
	default:
		return ClassReal, "unknown outcome"
	}
}

// lintRun classifies a program gc and gno lint both reject while the VM runs.
func lintRun(prog Program, goRes, gnoRes, lintRes Result) (Class, string, bool) {
	if lintRes.Kind != KindReject || goRes.Kind != KindReject || gnoRes.Kind == KindReject {
		return "", "", false
	}
	if prog.ExpectGo != "" && !expectMatch(goRes, prog.ExpectGo) {
		return ClassReal, "known go output changed", true
	}
	if prog.ExpectGno != "" && !expectMatch(gnoRes, prog.ExpectGno) {
		return ClassReal, "known gno output changed", true
	}
	return ClassLintRun, "run accepts what lint rejects", true
}

func exactOrStale(prog Program) (Class, string) {
	if prog.Known != "" {
		return ClassKnownAgrees, "known now agrees"
	}
	return ClassExact, ""
}

func panicWording(goMsg, gnoMsg string) bool {
	if NormalizePanic(goMsg) == NormalizePanic(gnoMsg) {
		return true
	}
	return nilSliceWording(goMsg, gnoMsg) || negShiftWording(goMsg, gnoMsg)
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
	indexGoRE   = regexp.MustCompile(`^runtime error: index out of range \[(-?\d+)\] with length (-?\d+)$`)
	indexGnoRE  = regexp.MustCompile(`^runtime error: slice index out of bounds: (-?\d+) \(len=(-?\d+)\)$`)
	boundsGoRE  = regexp.MustCompile(`^runtime error: slice bounds out of range \[(-?\d+):(-?\d+)\]$`)
	boundsGnoRE = regexp.MustCompile(`^runtime error: invalid slice index (-?\d+) > (-?\d+)$`)
)

var negShiftRE = regexp.MustCompile(`^runtime error: negative shift amount(?:: \((-?\d+)\b[^)]*\))?$`)

// canonicalRuntimePanic maps runtime panic texts that name the same event.
// User panic("...") messages without the "runtime error: " prefix are never
// folded. Numbers are kept: index 3 with length 2 must not match index 4,
// and a negative shift of -1 must not match a shift of -2.
func canonicalRuntimePanic(s string) (string, bool) {
	if !strings.HasPrefix(s, "runtime error: ") {
		return "", false
	}
	switch s {
	case "runtime error: integer divide by zero", "runtime error: division by zero":
		return "division by zero", true
	}
	if m := negShiftRE.FindStringSubmatch(s); m != nil {
		if m[1] != "" {
			return "negative shift " + m[1], true
		}
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

// nilSliceWording is the nil-slice index panic at index 0. gc reports
// `index out of range [0] with length 0`; Gno names the nil slice and does
// not repeat the index. Any other index is a different event.
func nilSliceWording(goMsg, gnoMsg string) bool {
	if cleanPanic(gnoMsg) != "runtime error: nil slice index (out of bounds)" {
		return false
	}
	m := indexGoRE.FindStringSubmatch(cleanPanic(goMsg))
	return m != nil && m[1] == "0" && m[2] == "0"
}

func shiftAmount(s string) (string, bool) {
	m := negShiftRE.FindStringSubmatch(s)
	if m == nil || m[1] == "" {
		return "", false
	}
	return m[1], true
}

// negShiftWording matches gc's operand-free "negative shift amount" with
// Gno's "negative shift amount: (N type)". When both sides print an
// operand, the operands must be equal. gc does not print one, so that
// specific pair cannot confirm the operand.
func negShiftWording(goMsg, gnoMsg string) bool {
	g := cleanPanic(goMsg)
	n := cleanPanic(gnoMsg)
	if !negShiftRE.MatchString(g) || !negShiftRE.MatchString(n) {
		return false
	}
	ga, gok := shiftAmount(g)
	na, nok := shiftAmount(n)
	if gok && nok {
		return ga == na
	}
	return true
}

const numTypePattern = `(?:u?int(?:8|16|32|64)?|uintptr|byte|rune|float(?:32|64))`

var (
	diagFileRE     = regexp.MustCompile(`(?:\./)?(?:[A-Za-z]:)?[\w./\\-]+\.g(?:no|o):\d+(?::\d+)?(?:-\d+(?::\d+)?)?:`)
	notDeclRE      = regexp.MustCompile(`name (\w+) not declared`)
	undefRE        = regexp.MustCompile(`undefined: (\w+)`)
	unusedGoRE     = regexp.MustCompile(`declared and not used: (\w+)`)
	unusedAltRE    = regexp.MustCompile(`(\w+) declared and not used`)
	overflowAsRE   = regexp.MustCompile(`\bas (` + numTypePattern + `)\b`)
	overflowTailRE = regexp.MustCompile(`(?:overflows|underflows) (` + numTypePattern + `)\b`)
)

// normalizeDiag strips file:line prefixes and process noise so two
// rejections can be compared on the diagnostic itself. The same missing
// name ("undefined: n" and "name n not declared") is one diagnostic.
// "overflows uint64" and "undefined: foo" stay different.
func normalizeDiag(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	var ids []string
	var other []string
	for line := range strings.SplitSeq(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "# ") || strings.HasPrefix(line, "exit status") {
			continue
		}
		line = strings.TrimSpace(diagFileRE.ReplaceAllString(line, ""))
		if line == "" {
			continue
		}
		line = spaceRE.ReplaceAllString(line, " ")
		if id, ok := diagIdentity(line); ok {
			ids = append(ids, id)
			continue
		}
		other = append(other, line)
	}
	// A recognized identity is the diagnostic. Stack frames after it are not.
	if len(ids) > 0 {
		return strings.Join(ids, "\n")
	}
	return strings.Join(other, "\n")
}

// overflowWording is a constant that does not fit, rejected on both sides,
// with different text. gc names the type ("overflows int32", or "as int8
// value ... (overflows)"). The VM says "bigint overflows" or, for a
// negative constant, "bigint underflows", and does not repeat the type.
// When both messages name a type, the types must match.
func overflowWording(goDetail, gnoDetail string) bool {
	g, gok := overflowIdentity(goDetail)
	n, nok := overflowIdentity(gnoDetail)
	if !gok || !nok {
		return false
	}
	if g == n {
		return true
	}
	return g == "overflows" || n == "overflows"
}

func overflowIdentity(detail string) (string, bool) {
	var found string
	ok := false
	for line := range strings.SplitSeq(normalizeDiag(detail), "\n") {
		id, matched := overflowLine(line)
		if !matched {
			continue
		}
		ok = true
		if found == "" || found == "overflows" {
			found = id
		} else if id != "overflows" && id != found {
			return "", false
		}
	}
	return found, ok
}

func overflowLine(line string) (string, bool) {
	if strings.HasPrefix(line, "undefined: ") || strings.HasPrefix(line, "declared and not used: ") {
		return "", false
	}
	if m := overflowTailRE.FindStringSubmatch(line); m != nil {
		return "overflows " + m[1], true
	}
	if m := overflowAsRE.FindStringSubmatch(line); m != nil &&
		(strings.Contains(line, "overflows") || strings.Contains(line, "underflows")) {
		return "overflows " + m[1], true
	}
	if strings.Contains(line, "bigint overflows") || strings.Contains(line, "bigint underflows") ||
		strings.Contains(line, "shift overflow") {
		return "overflows", true
	}
	return "", false
}

func diagIdentity(line string) (string, bool) {
	switch {
	case notDeclRE.MatchString(line):
		return "undefined: " + notDeclRE.FindStringSubmatch(line)[1], true
	case undefRE.MatchString(line):
		return "undefined: " + undefRE.FindStringSubmatch(line)[1], true
	case unusedGoRE.MatchString(line):
		return "declared and not used: " + unusedGoRE.FindStringSubmatch(line)[1], true
	case unusedAltRE.MatchString(line):
		return "declared and not used: " + unusedAltRE.FindStringSubmatch(line)[1], true
	default:
		return "", false
	}
}

// Compare reports whether the results are free of a real behavioral
// divergence. Panic-wording differences count as agreement here. The lint
// oracle is not consulted; RunCorpus classifies with it.
func Compare(goRes, gnoRes Result) (match bool, reason string) {
	cl, reason := Classify(Program{}, goRes, gnoRes, Result{})
	return cl == ClassExact || cl == ClassPanicWording || cl == ClassRejectWording, reason
}
