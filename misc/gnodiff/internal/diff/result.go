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

// NormalizePanic reduces a panic value to the text that should agree across
// toolchains. Wording that describes the same runtime event is rewritten to
// a canonical form that still carries the indexes, so a wrong bound stays a
// divergence.
func NormalizePanic(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "panic: ")
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	s = strings.TrimSpace(s)
	for _, pair := range panicSynonyms {
		s = strings.ReplaceAll(s, pair[0], pair[1])
	}
	s = spaceRE.ReplaceAllString(s, " ")
	s = strings.TrimSpace(s)
	if canon, ok := canonicalRuntimePanic(s); ok {
		return canon
	}
	return s
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

// Compare reports whether the two results agree.
// match is false when a runner itself failed (KindInternal): that is a
// problem with the fuzzer, and it must not look like language agreement.
func Compare(goRes, gnoRes Result) (match bool, reason string) {
	if goRes.Kind == KindInternal || gnoRes.Kind == KindInternal {
		return false, "runner failed"
	}
	if goRes.Truncated || gnoRes.Truncated {
		if goRes.Truncated != gnoRes.Truncated || goRes.Stdout != gnoRes.Stdout || goRes.Kind != gnoRes.Kind {
			return false, "output truncated"
		}
	}
	if goRes.Kind != gnoRes.Kind {
		return false, fmt.Sprintf("outcome %s vs %s", goRes.Kind, gnoRes.Kind)
	}
	switch goRes.Kind {
	case KindOK:
		if goRes.Stdout != gnoRes.Stdout {
			return false, "stdout"
		}
	case KindPanic:
		if NormalizePanic(goRes.Panic) != NormalizePanic(gnoRes.Panic) {
			return false, "panic message"
		}
		if goRes.Stdout != gnoRes.Stdout {
			return false, "stdout before panic"
		}
	case KindReject, KindNonTerm:
		return true, ""
	default:
		return false, "unknown outcome"
	}
	return true, ""
}
