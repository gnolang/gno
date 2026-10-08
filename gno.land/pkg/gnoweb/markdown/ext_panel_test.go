package markdown

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// newGnoMarkdown returns goldmark with the Gno extensions.
func newGnoMarkdown() goldmark.Markdown {
	m := goldmark.New()
	NewGnoExtension().Extend(m)
	return m
}

// convertGno renders src with a fresh Gno parser context, as gnoweb does
// for each page.
func convertGno(m goldmark.Markdown, src []byte, buf *bytes.Buffer) error {
	return m.Convert(src, buf, parser.WithContext(NewGnoParserContext(GnoContext{})))
}

// BenchmarkPanel renders the same body with and without panel wrappers.
// The body has no `<`-leading line, so both cases do the same work.
func BenchmarkPanel(b *testing.B) {
	body := strings.Repeat("Some **text** with a [link](/r/demo).\n\n- item\n- item\n\n", 20)
	docs := map[string]string{
		"plain": strings.Repeat(body, 10),
		"panel": strings.Repeat("<gno-panel>\n"+body+"</gno-panel>\n\n", 10),
		// Many tiny panels: tag handling dominates.
		"pairs": strings.Repeat("<gno-panel>\nx\n</gno-panel>\n", 1000),
	}
	for name, doc := range docs {
		src := []byte(doc)
		b.Run(name, func(b *testing.B) {
			m := newGnoMarkdown()
			var buf bytes.Buffer
			b.ReportAllocs()
			for b.Loop() {
				buf.Reset()
				if err := convertGno(m, src, &buf); err != nil {
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
		"<gno-panel>\rx\r</gno-panel>\r",
		"<gno-foreign>\n<gno-panel>\nx\n</gno-foreign>\n</gno-panel>\n",
		"<gno-panel>\n<div></gno-panel>\nafter\n",
		"> [!NOTE]\n> <gno-panel>\n> x\n",
		"<gno-form>\n<gno-panel>\n<gno-input name=\"a\" />\n</gno-panel>\n</gno-form>\n",
	}
	m := newGnoMarkdown()
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

// TestPanelEOFWithoutNewline covers input that ends right after a tag,
// which a golden txtar cannot express.
func TestPanelEOFWithoutNewline(t *testing.T) {
	cases := map[string]string{
		"<gno-panel>\nx\n</gno-panel>": "<section class=\"gno-panel\">\n<p>x</p>\n</section>\n",
		"<gno-panel>":                  "<section class=\"gno-panel\">\n</section>\n",
	}
	m := newGnoMarkdown()
	for in, want := range cases {
		var buf bytes.Buffer
		if err := convertGno(m, []byte(in), &buf); err != nil {
			t.Fatal(err)
		}
		if buf.String() != want {
			t.Errorf("%q rendered %q, want %q", in, buf.String(), want)
		}
	}
}

// TestPanelAllocsScaleLinearly guards against per-panel cost growing with
// the page: ten times the panels must cost about ten times the allocations.
// Allocation counts are deterministic, unlike wall-clock ratios; see
// BenchmarkPanel/pairs for time.
func TestPanelAllocsScaleLinearly(t *testing.T) {
	m := newGnoMarkdown()
	allocs := func(n int) float64 {
		src := []byte(strings.Repeat("<gno-panel>\nx\n</gno-panel>\n", n))
		var buf bytes.Buffer
		return testing.AllocsPerRun(5, func() {
			buf.Reset()
			_ = convertGno(m, src, &buf)
		})
	}
	small, large := allocs(200), allocs(2000)
	if ratio := large / small; ratio < 8 || ratio > 11 {
		t.Errorf("allocs for 2000 panels / 200 panels = %.2f, want about 10", ratio)
	}
}

// TestPanelNoAllocPaths pins the zero-allocation paths the ADR relies on:
// the transformer on a page without panel tags, and the HTML block wrapper
// ending a block on a panel close tag.
func TestPanelNoAllocPaths(t *testing.T) {
	doc := ast.NewDocument()
	doc.AppendChild(doc, ast.NewParagraph())
	pc := parser.NewContext()
	tr := &panelASTTransformer{}
	if n := testing.AllocsPerRun(100, func() { tr.Transform(doc, nil, pc) }); n != 0 {
		t.Errorf("transformer on a page without panels allocates %v times", n)
	}

	html := ast.NewHTMLBlock(ast.HTMLBlockType7)
	doc.AppendChild(doc, html)
	pc.Set(panelOpenKey, true)
	wrapper := panelHTMLBlockParser{parser.NewHTMLBlockParser()}
	reader := text.NewReader([]byte("</gno-panel>\n"))
	if n := testing.AllocsPerRun(100, func() {
		if wrapper.Continue(html, reader, pc) != parser.Close {
			t.Fatal("close tag did not end the HTML block")
		}
	}); n != 0 {
		t.Errorf("HTML block wrapper allocates %v times on a close tag", n)
	}
}
