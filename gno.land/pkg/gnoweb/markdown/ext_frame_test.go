package markdown

import (
	"bytes"
	"runtime"
	"strings"
	"testing"
	"time"

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

// BenchmarkFrame renders the same body with and without frame wrappers.
// The body has no `<`-leading line, so both cases do the same work.
func BenchmarkFrame(b *testing.B) {
	body := strings.Repeat("Some **text** with a [link](/r/demo).\n\n- item\n- item\n\n", 20)
	docs := map[string]string{
		"plain": strings.Repeat(body, 10),
		"frame": strings.Repeat("<gno-frame>\n"+body+"</gno-frame>\n\n", 10),
		// Many tiny frames: tag handling dominates.
		"pairs": strings.Repeat("<gno-frame>\nx\n</gno-frame>\n", 1000),
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

// TestFrameNestDepthBalanced checks that every frame shape leaves the
// shared gno-* depth counter where it found it, so a frame never eats
// nesting budget from blocks that follow it.
func TestFrameNestDepthBalanced(t *testing.T) {
	inputs := []string{
		"<gno-frame>\nx\n</gno-frame>\n",
		"<gno-frame>\nunclosed\n",
		"</gno-frame>\n",
		"<gno-frame class=\"x\">\n",
		"<gno-frame>\n<gno-frame>\n</gno-frame>\n</gno-frame>\n",
		"> <gno-frame>\n> x\n",
		"<gno-columns>\n<gno-frame>\na\n<gno-columns-sep>\nb\n</gno-columns>\n",
		"<gno-frame>\rx\r</gno-frame>\r",
		"<gno-foreign>\n<gno-frame>\nx\n</gno-foreign>\n</gno-frame>\n",
		"<gno-frame>\n<div></gno-frame>\nafter\n",
		"> [!NOTE]\n> <gno-frame>\n> x\n",
		"<gno-form>\n<gno-frame>\n<gno-input name=\"a\" />\n</gno-frame>\n</gno-form>\n",
		"<gno-frame>\n<gno-columns>\na\n<gno-columns-sep>\nb\n</gno-columns>\n</gno-frame>\n",
		"<gno-frame>\n<gno-columns>\n<gno-frame>\ncard\n</gno-frame>\n</gno-columns>\n</gno-frame>\n",
		"<gno-frame>\n<gno-columns>\n<gno-frame>\ncard\n<gno-columns-sep>\nb\n",
		"<gno-frame>\n<gno-columns>\na\n</gno-frame>\nb\n</gno-columns>\n",
		"<gno-frame>\n<gno-columns>\n<gno-frame>\n</gno-frame>\n</gno-frame>\n</gno-columns>\n",
		"<gno-frame>\n<gno-columns>\n<gno-columns>\n</gno-columns>\n</gno-frame>\n",
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

// TestParseFrameLineTagNoAlloc backs the ADR's perf claim: classifying a
// line costs no allocation, frame tag or not.
func TestParseFrameLineTagNoAlloc(t *testing.T) {
	for _, line := range []string{"<gno-frame>", "</gno-frame>", `<gno-frame class="x">`, "<div>", "plain"} {
		b := []byte(line)
		if n := testing.AllocsPerRun(100, func() { parseFrameLineTag(b) }); n != 0 {
			t.Errorf("parseFrameLineTag(%q) allocates %v times", line, n)
		}
	}
}

// TestFrameEOFWithoutNewline covers input that ends right after a tag,
// which a golden txtar cannot express.
func TestFrameEOFWithoutNewline(t *testing.T) {
	cases := map[string]string{
		"<gno-frame>\nx\n</gno-frame>": "<section class=\"gno-frame\">\n<p>x</p>\n</section>\n",
		"<gno-frame>":                  "<section class=\"gno-frame\">\n</section>\n",
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

// TestFrameNULInTag covers a NUL byte in a tag name, which would make a
// golden txtar a binary file for git: the line is no tag line, and
// goldmark renders the NUL as U+FFFD.
func TestFrameNULInTag(t *testing.T) {
	const in = "<gno-frame\x00>\n\nafter\n"
	const want = "<p>&lt;gno-frame\ufffd&gt;</p>\n<p>after</p>\n"
	var buf bytes.Buffer
	if err := convertGno(newGnoMarkdown(), []byte(in), &buf); err != nil {
		t.Fatal(err)
	}
	if buf.String() != want {
		t.Errorf("%q rendered %q, want %q", in, buf.String(), want)
	}
}

// TestFrameAllocsScaleLinearly guards against per-frame cost growing with
// the page: ten times the frames must cost about ten times the allocations.
// Allocation counts are deterministic, unlike wall-clock ratios; see
// BenchmarkFrame/pairs for time.
func TestFrameAllocsScaleLinearly(t *testing.T) {
	m := newGnoMarkdown()
	allocs := func(n int) float64 {
		src := []byte(strings.Repeat("<gno-frame>\nx\n</gno-frame>\n", n))
		var buf bytes.Buffer
		return testing.AllocsPerRun(5, func() {
			buf.Reset()
			_ = convertGno(m, src, &buf)
		})
	}
	small, large := allocs(200), allocs(2000)
	if ratio := large / small; ratio < 8 || ratio > 11 {
		t.Errorf("allocs for 2000 frames / 200 frames = %.2f, want about 10", ratio)
	}
}

// TestFrameNoAllocPaths pins the zero-allocation paths the ADR relies on:
// the transformer on a page without frame tags, and the HTML block wrapper
// ending a block on a frame close tag.
func TestFrameNoAllocPaths(t *testing.T) {
	doc := ast.NewDocument()
	doc.AppendChild(doc, ast.NewParagraph())
	pc := parser.NewContext()
	tr := &frameASTTransformer{}
	if n := testing.AllocsPerRun(100, func() { tr.Transform(doc, nil, pc) }); n != 0 {
		t.Errorf("transformer on a page without frames allocates %v times", n)
	}

	html := ast.NewHTMLBlock(ast.HTMLBlockType7)
	doc.AppendChild(doc, html)
	pc.Set(frameOpenKey, true)
	wrapper := frameHTMLBlockParser{parser.NewHTMLBlockParser()}
	reader := text.NewReader([]byte("</gno-frame>\n"))
	if n := testing.AllocsPerRun(100, func() {
		if wrapper.Continue(html, reader, pc) != parser.Close {
			t.Fatal("close tag did not end the HTML block")
		}
	}); n != 0 {
		t.Errorf("HTML block wrapper allocates %v times on a close tag", n)
	}
}

// TestFrameGridScanLinear guards the look-ahead that checks whether a grid
// opened in a frame closes inside it: on pages where it never does, or
// always does, four times the input must not cost much more than four
// times the time. A scan restarted per block would take about 16 times;
// the bound is 10 so a GC or scheduler pause on a CI runner does not trip it.
func TestFrameGridScanLinear(t *testing.T) {
	body := strings.Repeat("x\n\n", 50)
	pages := map[string]string{
		"half inside": "<gno-frame>\n<gno-columns>\n" + body + "</gno-frame>\n" + body + "</gno-columns>\n",
		"inside":      "<gno-frame>\n<gno-columns>\n<gno-frame>\n" + body + "</gno-frame>\n<gno-columns-sep>\n" + body + "</gno-columns>\n</gno-frame>\n",
		"unclosed":    "<gno-frame>\n<gno-columns>\n" + body,
	}
	m := newGnoMarkdown()
	render := func(src []byte) time.Duration {
		best := time.Duration(1<<63 - 1)
		for range 5 {
			var buf bytes.Buffer
			runtime.GC()
			start := time.Now()
			if err := convertGno(m, src, &buf); err != nil {
				t.Fatal(err)
			}
			best = min(best, time.Since(start))
		}
		return best
	}
	for name, page := range pages {
		small := render([]byte(strings.Repeat(page, 50)))
		large := render([]byte(strings.Repeat(page, 200)))
		if ratio := float64(large) / float64(small); ratio > 10 {
			t.Errorf("%s: 4x the input took %.1fx the time", name, ratio)
		}
	}
}
