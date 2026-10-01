// Regression guard for the goldmark emphasis-parsing DoS (yuin/goldmark#555)
// against gnoweb's production markdown pipeline. gnoweb fetches a realm's
// attacker-controlled Render() output and drives the production goldmark
// instance (NewDefaultGoldmarkOptions, wired through NewHTMLRenderer) over it
// server-side per HTTP request. BenchmarkEmphasisStress (markdown_dos_profile_test.go)
// characterizes the raw parse cost; this test pins the guard's effect.

package gnoweb

import (
	"bytes"
	"strings"
	"testing"

	md "github.com/gnolang/gno/gno.land/pkg/gnoweb/markdown"
	"github.com/stretchr/testify/require"
	"github.com/yuin/goldmark"
)

// makeProductionRenderer builds the exact goldmark instance gnoweb uses for
// realm rendering — NewDefaultGoldmarkOptions wired through goldmark.New, the
// same construction NewHTMLRenderer performs.
func makeProductionRenderer() goldmark.Markdown {
	cfg := NewDefaultRenderConfig()
	return goldmark.New(cfg.GoldmarkOptions...)
}

// payloadEmphasis builds a string of overlapping emphasis runs designed to
// stress goldmark's delimiter-stack scanner. The mix of `*`, `_`, `**` and
// `__` with intentionally unmatched markers forces the parser to consider
// many open/close pairings.
func payloadEmphasis(n int) []byte {
	var sb strings.Builder
	sb.Grow(n)
	stress := "*a_b*c_d**e_f*g**h*i_j**"
	for sb.Len() < n {
		sb.WriteString(stress)
	}
	out := sb.String()
	if len(out) > n {
		out = out[:n]
	}
	return []byte(out)
}

// TestEmphasisGuardBoundsProductionRender is the deterministic regression guard
// for yuin/goldmark#555: over-cap emphasis renders as literal text through the
// full production renderer, so the <em> count stays bounded by the per-block
// cap instead of growing with the span count. Unguarded, goldmark's default
// parser would emphasize every span and reopen the quadratic parse.
func TestEmphasisGuardBoundsProductionRender(t *testing.T) {
	t.Parallel()
	gm := makeProductionRenderer()

	var out bytes.Buffer
	require.NoError(t, gm.Convert([]byte(strings.Repeat("*x* ", md.MaxEmphasisDelimitersPerBlock*3)), &out))
	n := strings.Count(out.String(), "<em>")
	require.LessOrEqualf(t, n, md.MaxEmphasisDelimitersPerBlock, "<em> count = %d (emphasis guard not active in production renderer)", n)
}
