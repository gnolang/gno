package test

import (
	"fmt"
	"regexp"
	"strings"
)

// The verdict line of a .go corpus file: exactly one, last in the file,
//
//	// Verdict: <Kind> <mode>[: <reason>]
//
// Kinds: Clean (both checkers reach the same verdict), Unsupported
// (feature gap or gc-only test; mode is the tag from unsupportedTags),
// KnownIssue (Gno is wrong; the reason cites the tracking issue) and
// KnownDivergence (Gno differs on purpose). Wording differences are
// evidence, never a verdict.
const (
	VerdictClean           = "Clean"
	VerdictUnsupported     = "Unsupported"
	VerdictKnownIssue      = "KnownIssue"
	VerdictKnownDivergence = "KnownDivergence"
)

type verdict struct {
	Kind   string
	Mode   string // run | compile | errorcheck, or the Unsupported tag
	Reason string
}

func (v verdict) String() string {
	s := v.Kind + " " + v.Mode
	if v.Reason != "" {
		s += ": " + v.Reason
	}
	return s
}

func (v verdict) directive() Directive {
	return Directive{Name: DirectiveVerdict, Content: v.String(), Complete: true}
}

// parseVerdict reads the content of a `// Verdict:` line.
func parseVerdict(content string) (verdict, bool) {
	content = strings.TrimSpace(content)
	kind, rest, _ := strings.Cut(content, " ")
	switch kind {
	case VerdictClean, VerdictUnsupported, VerdictKnownIssue, VerdictKnownDivergence:
	default:
		return verdict{}, false
	}
	mode, reason, _ := strings.Cut(rest, ":")
	mode = strings.TrimSpace(mode)
	if mode == "" {
		return verdict{}, false
	}
	return verdict{Kind: kind, Mode: mode, Reason: strings.TrimSpace(reason)}, true
}

// verdictIn returns the file's verdict, if it has a well-formed one.
func verdictIn(dirs Directives) (verdict, bool) {
	d := dirs.First(DirectiveVerdict)
	if d == nil {
		return verdict{}, false
	}
	return parseVerdict(d.Content)
}

var (
	reIssueRef = regexp.MustCompile(`#\d+|github\.com/[\w.-]+/[\w.-]+/issues/\d+`)
	// Where a Gno decision can be recorded: a docs page, or an issue /
	// PR / URL. The Go spec permitting a difference is not a decision.
	reDecisionRef = regexp.MustCompile(`\bdocs/[\w./-]+|#\d+|https?://`)
)

// check enforces that a Known* verdict is a human decision: sync's TODO
// placeholder never passes verify, a KnownIssue names its issue, and a
// KnownDivergence cites the decision it rests on — without one it is an
// issue, not a blessing.
func (v verdict) check() error {
	switch v.Kind {
	case VerdictKnownIssue, VerdictKnownDivergence:
		if v.Reason == "" || strings.HasPrefix(v.Reason, "TODO") {
			return fmt.Errorf("`// Verdict: %s` needs a reason (sync's TODO placeholder does not count)", v.Kind)
		}
	}
	switch {
	case v.Kind == VerdictKnownIssue && !reIssueRef.MatchString(v.Reason):
		return fmt.Errorf("`// Verdict: KnownIssue` must cite its tracking issue (#NNNN or an issue URL): %q", v.Reason)
	case v.Kind == VerdictKnownDivergence && !reDecisionRef.MatchString(v.Reason):
		return fmt.Errorf("`// Verdict: KnownDivergence` must cite the Gno decision it rests on (a docs/ page, #NNNN or a URL), else file it as a KnownIssue: %q", v.Reason)
	}
	return nil
}

// verdictKind returns the file's verdict kind, or "" without one.
func verdictKind(dirs Directives) string {
	v, _ := verdictIn(dirs)
	return v.Kind
}
