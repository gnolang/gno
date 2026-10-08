package markdown

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// BenchmarkPanel renders the same body with and without panel wrappers.
// The body has no `<`-leading line, so both cases do the same work.
func BenchmarkPanel(b *testing.B) {
	body := strings.Repeat("Some **text** with a [link](/r/demo).\n\n- item\n- item\n\n", 20)
	docs := map[string]string{
		"plain": strings.Repeat(body, 10),
		"panel": strings.Repeat("<gno-panel>\n"+body+"</gno-panel>\n\n", 10),
	}
	for name, doc := range docs {
		src := []byte(doc)
		b.Run(name, func(b *testing.B) {
			m := goldmark.New()
			NewGnoExtension().Extend(m)
			var buf bytes.Buffer
			b.ReportAllocs()
			for b.Loop() {
				buf.Reset()
				ctx := parser.WithContext(NewGnoParserContext(GnoContext{}))
				if err := m.Convert(src, &buf, ctx); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// TestPanelNestDepthBalanced checks that every panel shape leaves the
// shared gno-* depth counter where it found it, so a panel never eats
// nesting budget from blocks that follow it.
func TestPanelNestDepthBalanced(t *testing.T) {
	inputs := []string{
		"<gno-panel>\nx\n</gno-panel>\n",
		"<gno-panel>\nunclosed\n",
		"</gno-panel>\n",
		"<gno-panel class=\"x\">\n",
		"<gno-panel>\n<gno-panel>\n</gno-panel>\n</gno-panel>\n",
		"> <gno-panel>\n> x\n",
		"<gno-columns>\n<gno-panel>\na\n<gno-columns-sep>\nb\n</gno-columns>\n",
	}
	m := goldmark.New()
	NewGnoExtension().Extend(m)
	for _, in := range inputs {
		pc := NewGnoParserContext(GnoContext{})
		m.Parser().Parse(text.NewReader([]byte(in)), parser.WithContext(pc))
		if d := Get(pc); d != 0 {
			t.Errorf("depth after %q = %d, want 0", in, d)
		}
	}
}

// TestParsePanelLineTagNoAlloc backs the ADR's perf claim: classifying a
// line costs no allocation, panel tag or not.
func TestParsePanelLineTagNoAlloc(t *testing.T) {
	for _, line := range []string{"<gno-panel>", "</gno-panel>", `<gno-panel class="x">`, "<div>", "plain"} {
		b := []byte(line)
		if n := testing.AllocsPerRun(100, func() { parsePanelLineTag(b) }); n != 0 {
			t.Errorf("parsePanelLineTag(%q) allocates %v times", line, n)
		}
	}
}
