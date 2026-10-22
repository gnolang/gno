package test

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
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

// preprocessUmbrella is the tracking issue every errorcheck gap cites:
// a marker line Gno's preprocess misses, or an unmarked line it rejects,
// is by definition a preprocess conformance gap, fixed in place by the
// type-checker PR series.
const preprocessUmbrella = "#6344"

// An errorcheck reason is `;`-separated parts. A gap group is
// `[#NNNN ]<kind>: lines a, b, c-d[ (note)]` (kind in gapKinds); without
// a ref it belongs to the last ref in force. `[#NNNN ]on-chain path …`
// is recomputed each run. Any other part is prose, kept verbatim.
type gapGroup struct {
	issue string // the #NNNN owning the lines
	kind  string // one of gapKinds, or onChainKind
	lines []int
	note  string // "(…)" after the list, kept with the group; the text of an on-chain part
	prose string // set instead of the fields above for a non-group part
}

var gapKinds = []string{"uncaught markers", "over-strict", "internal errors", "types differ"}

const onChainKind = "on-chain path"

var (
	reGapGroup     = regexp.MustCompile(`^(?:(#\d+) )?(uncaught markers|over-strict|internal errors|types differ): lines (\d+(?:\s*-\s*\d+)?(?:\s*,\s*\d+(?:\s*-\s*\d+)?)*)\s*(\(.*\))?$`)
	reOnChainPart  = regexp.MustCompile(`^(?:(#\d+) )?on-chain path (.+)$`)
	reLeadingIssue = regexp.MustCompile(`^#\d+`)
)

// splitReasonParts splits a reason on `;` outside parentheses.
func splitReasonParts(reason string) []string {
	var parts []string
	depth, start := 0, 0
	for i, c := range reason {
		switch c {
		case '(':
			depth++
		case ')':
			depth = max(depth-1, 0)
		case ';':
			if depth == 0 {
				parts = append(parts, reason[start:i])
				start = i + 1
			}
		}
	}
	parts = append(parts, reason[start:])
	var out []string
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// parseLineList expands "3, 5-7" to 3, 5, 6, 7.
func parseLineList(s string) []int {
	var ns []int
	for item := range strings.SplitSeq(s, ",") {
		lo, hi, isRange := strings.Cut(strings.TrimSpace(item), "-")
		a, _ := strconv.Atoi(strings.TrimSpace(lo))
		b := a
		if isRange {
			b, _ = strconv.Atoi(strings.TrimSpace(hi))
		}
		for n := a; n <= b; n++ {
			ns = append(ns, n)
		}
	}
	return ns
}

// parseGapGroups reads a reason into its parts, in order.
func parseGapGroups(reason string) []gapGroup {
	var gs []gapGroup
	issue := ""
	for _, part := range splitReasonParts(reason) {
		if m := reGapGroup.FindStringSubmatch(part); m != nil {
			if m[1] != "" {
				issue = m[1]
			}
			gs = append(gs, gapGroup{issue: issue, kind: m[2], lines: parseLineList(m[3]), note: m[4]})
			continue
		}
		if m := reOnChainPart.FindStringSubmatch(part); m != nil {
			if m[1] != "" {
				issue = m[1]
			}
			gs = append(gs, gapGroup{issue: issue, kind: onChainKind, note: m[2]})
			continue
		}
		if ref := reLeadingIssue.FindString(part); ref != "" {
			issue = ref
		}
		gs = append(gs, gapGroup{prose: part})
	}
	return gs
}

// formatGapGroups writes the parts back; a group names its ref only when
// it differs from the one in force (`#6344 uncaught markers: …; over-strict: …`).
func formatGapGroups(gs []gapGroup) string {
	inForce := ""
	parts := make([]string, 0, len(gs))
	for _, g := range gs {
		if g.prose != "" {
			if ref := reLeadingIssue.FindString(g.prose); ref != "" {
				inForce = ref
			}
			parts = append(parts, g.prose)
			continue
		}
		s := g.kind + ": lines " + lineList(g.lines)
		if g.kind == onChainKind {
			s = onChainKind + " " + g.note
		} else if g.note != "" {
			s += " " + g.note
		}
		if g.issue != "" && g.issue != inForce {
			s = g.issue + " " + s
			inForce = g.issue
		}
		parts = append(parts, s)
	}
	return strings.Join(parts, "; ")
}

// gapLabel is how a kind reads in a stale-verdict message.
func gapLabel(kind string) string {
	if kind == "uncaught markers" {
		return "uncaught"
	}
	return kind
}

// reconcileGapReason prunes each group of a hand-written reason to the
// lines still gaps of its kind (an emptied group is dropped, a note stays),
// gives unowned gap lines to the umbrella issue and rewrites the on-chain
// part; stale lists what verify must refuse (lines no longer gaps, gaps
// unlisted, an on-chain part that changed). The `types differ` kind is the
// types check's (typediff.go), reconciled in every mode.
func reconcileGapReason(prev string, uncaught, overStrict, internal, typesDiffer []int, onChain string) (reason string, stale []string) {
	gaps := map[string][]int{
		gapKinds[0]: slices.Sorted(slices.Values(uncaught)), gapKinds[1]: slices.Sorted(slices.Values(overStrict)),
		gapKinds[2]: slices.Sorted(slices.Values(internal)), gapKinds[3]: slices.Sorted(slices.Values(typesDiffer)),
	}
	listed := map[string]map[int]bool{}
	noLonger := map[string][]int{}
	prevOnChain, hadOnChain := "", false
	var out []gapGroup
	for _, g := range parseGapGroups(prev) {
		if g.kind == onChainKind {
			prevOnChain, hadOnChain = g.note, true
			continue
		}
		if g.prose != "" {
			out = append(out, g)
			continue
		}
		var keep []int
		for _, l := range g.lines {
			if !slices.Contains(gaps[g.kind], l) {
				noLonger[g.kind] = append(noLonger[g.kind], l)
				continue
			}
			keep = append(keep, l)
			if listed[g.kind] == nil {
				listed[g.kind] = map[int]bool{}
			}
			listed[g.kind][l] = true
		}
		if len(keep) > 0 {
			g.lines = keep
			out = append(out, g)
		}
	}
	for _, kind := range gapKinds {
		if ls := noLonger[kind]; len(ls) > 0 {
			stale = append(stale, fmt.Sprintf("lines %s are no longer %s", lineList(ls), gapLabel(kind)))
		}
	}
	for _, kind := range gapKinds {
		var unlisted []int
		for _, l := range gaps[kind] {
			if !listed[kind][l] {
				unlisted = append(unlisted, l)
			}
		}
		if len(unlisted) == 0 {
			continue
		}
		stale = append(stale, fmt.Sprintf("lines %s are %s but unlisted", lineList(unlisted), gapLabel(kind)))
		merged := false
		for i := len(out) - 1; i >= 0 && !merged; i-- {
			if out[i].prose == "" && out[i].issue == preprocessUmbrella && out[i].kind == kind {
				out[i].lines = slices.Sorted(slices.Values(slices.Concat(out[i].lines, unlisted)))
				merged = true
			}
		}
		if !merged {
			out = append(out, gapGroup{issue: preprocessUmbrella, kind: kind, lines: unlisted})
		}
	}
	if onChain != "" {
		out = append(out, gapGroup{issue: preprocessUmbrella, kind: onChainKind, note: onChain})
	}
	if hadOnChain != (onChain != "") || prevOnChain != onChain {
		stale = append(stale, "the on-chain path part changed")
	}
	return formatGapGroups(out), stale
}

// typesVerdict folds the types check into v, the verdict a run or compile
// file reached from its own evidence: a Clean file with mismatches becomes
// the umbrella KnownIssue, a KnownIssue reason gets its `types differ`
// group reconciled (sync rewrites it, verify reports it stale) and a
// KnownDivergence stands as decided.
func typesVerdict(v verdict, typesDiffer []int, sync bool) (verdict, []string) {
	switch {
	case v.Kind == VerdictClean && len(typesDiffer) > 0:
		v = verdict{Kind: VerdictKnownIssue, Mode: v.Mode, Reason: errorcheckGapReason(nil, nil, nil, typesDiffer, "")}
	case v.Kind == VerdictKnownIssue && isSyncWrittenReason(v.Reason):
		if want := errorcheckGapReason(nil, nil, nil, typesDiffer, ""); sync || legacySyncReason(v.Reason) != want {
			v.Reason = want
		}
	case v.Kind == VerdictKnownIssue:
		reason, stale := reconcileGapReason(v.Reason, nil, nil, nil, typesDiffer, "")
		if sync {
			v.Reason = reason
		}
		return v, stale
	}
	return v, nil
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
