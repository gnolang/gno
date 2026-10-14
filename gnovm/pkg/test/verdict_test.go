package test

import (
	"strings"
	"testing"
)

func TestParseVerdict(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		in   string
		want verdict
		ok   bool
	}{
		{"Clean run", verdict{Kind: "Clean", Mode: "run"}, true},
		{"Unsupported channel: channels not supported", verdict{Kind: "Unsupported", Mode: "channel", Reason: "channels not supported"}, true},
		{"KnownIssue run: init order", verdict{Kind: "KnownIssue", Mode: "run", Reason: "init order"}, true},
		{"KnownDivergence run: println format", verdict{Kind: "KnownDivergence", Mode: "run", Reason: "println format"}, true},
		{"Clean", verdict{}, false},        // no mode
		{"Broken run", verdict{}, false},   // unknown kind
		{"run: no kind", verdict{}, false}, // kind missing
	} {
		got, ok := parseVerdict(tc.in)
		if ok != tc.ok || got != tc.want {
			t.Errorf("parseVerdict(%q) = %+v, %v; want %+v, %v", tc.in, got, ok, tc.want, tc.ok)
		}
		if ok && got.String() != tc.in {
			t.Errorf("round trip: %q -> %q", tc.in, got.String())
		}
	}
}

// goldens builds the pinned tail of a run-mode .go corpus file.
func goldens(gnoOut, gnoErr, goOut, verdictLine string) Directives {
	dirs := Directives{{Name: DirectiveGnoOutput, Content: gnoOut}}
	if gnoErr != "" {
		dirs = append(dirs, Directive{Name: DirectiveGnoError, Content: gnoErr})
	}
	dirs = append(dirs, Directive{Name: DirectiveGoOutput, Content: goOut})
	if verdictLine != "" {
		dirs = append(dirs, Directive{Name: DirectiveVerdict, Content: verdictLine, Complete: true})
	}
	return dirs
}

func TestFinalizeGoRunDivergence_Verify(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name            string
		dirs            Directives
		gnoOut, gnoErr  string
		goOut           string
		wantErrContains string // "" means verify passes
	}{
		{"clean", goldens("x", "", "x", "Clean run"), "x", "", "x", ""},
		{"known divergence", goldens("1.5", "", "+1.5e+000", "KnownDivergence run: fmt"), "1.5", "", "+1.5e+000", ""},
		{"known issue, gno errors", goldens("", "boom", "x", "KnownIssue run: bug"), "", "boom", "x", ""},
		{"divergence without verdict", goldens("a", "", "b", ""), "a", "", "b", "classify as"},
		{"stale clean", goldens("a", "", "b", "Clean run"), "a", "", "b", "stale `// Verdict: Clean run`"},
		{"stale known issue", goldens("x", "", "x", "KnownIssue run: fixed"), "x", "", "x", "stale verdict"},
		{"gno output drifted", goldens("old", "", "b", "KnownDivergence run: r"), "new", "", "b", "GnoOutput: diff"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			newDirs, err := finalizeGoRunDivergence(tc.dirs, tc.gnoOut, tc.gnoErr, tc.goOut, "", nil, false, true)
			if newDirs != nil {
				t.Fatalf("verify must not rewrite directives, got %v", newDirs)
			}
			switch {
			case tc.wantErrContains == "" && err != nil:
				t.Fatalf("unexpected error: %v", err)
			case tc.wantErrContains != "" && (err == nil || !strings.Contains(err.Error(), tc.wantErrContains)):
				t.Fatalf("want error containing %q, got %v", tc.wantErrContains, err)
			}
		})
	}
}

func TestFinalizeGoRunDivergence_Sync(t *testing.T) {
	t.Parallel()
	verdictOf := func(dirs Directives) string {
		if d := dirs.First(DirectiveVerdict); d != nil {
			return d.Content
		}
		return ""
	}
	for _, tc := range []struct {
		name           string
		dirs           Directives
		gnoOut, gnoErr string
		goOut          string
		isGoFile       bool
		wantVerdict    string
	}{
		{"match writes Clean", nil, "x", "", "x", true, "Clean run"},
		{"match writes Clean for opted-in .gno", nil, "x", "", "x", false, "Clean run"},
		{"gno error proposes KnownIssue", nil, "", "boom", "x", true, "KnownIssue run: TODO: explain the Gno bug (Gno errors where Go runs clean)"},
		{"output diff proposes KnownDivergence", nil, "a", "", "b", true, "KnownDivergence run: TODO: <category>: explain why this divergence is acceptable"},
		{"human reason survives re-sync", goldens("a", "", "b", "KnownDivergence run: on purpose"), "a", "", "b", true, "KnownDivergence run: on purpose"},
		{"stale KnownIssue replaced by Clean", goldens("x", "", "x", "KnownIssue run: fixed"), "x", "", "x", true, "Clean run"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			out, _ := finalizeGoRunDivergence(tc.dirs, tc.gnoOut, tc.gnoErr, tc.goOut, "", nil, true, tc.isGoFile)
			if got := verdictOf(out); got != tc.wantVerdict {
				t.Fatalf("verdict = %q, want %q", got, tc.wantVerdict)
			}
			n := 0
			for _, d := range out {
				if d.Name == DirectiveVerdict {
					n++
				}
			}
			if n != 1 {
				t.Fatalf("want exactly one verdict line, got %d: %v", n, out)
			}
		})
	}
}
