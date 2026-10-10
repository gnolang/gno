package markdown

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"golang.org/x/net/html"
)

func renderMathMarkdown(t *testing.T, src string) string {
	t.Helper()
	var buf bytes.Buffer
	gm := goldmark.New(goldmark.WithExtensions(NewGnoExtension()))
	require.NoError(t, gm.Convert([]byte(src), &buf))
	return buf.String()
}

func TestMathEscapesUserContent(t *testing.T) {
	cases := map[string]string{
		"text breaks out of math":     `$$\text{</math><script>alert('XSS')</script>}$$`,
		"class attribute injection":   `$\class{x" onclick="alert(1)}{y}$`,
		"color attribute injection":   `$\color{red" onmouseover="alert(1)}{y}$`,
		"fallback breaks out of span": `$\begin{matrix}</span><script>alert(1)</script>$`,
		"fallback breaks out of div":  "$$\n\\begin{matrix}</div><script>alert(1)</script>\n$$",
	}
	for name, src := range cases {
		t.Run(name, func(t *testing.T) {
			out := renderMathMarkdown(t, src)
			assert.NotContains(t, out, "<script")
			assert.NotContains(t, out, `" onclick=`)
			assert.NotContains(t, out, `" onmouseover=`)
		})
	}
}

// An entity reference typed in math source must display as typed: the
// browser must not decode it into markup-significant characters.
func TestMathTypedEntityIsNotDecoded(t *testing.T) {
	out := renderMathMarkdown(t, `$\text{&lt;b&gt;} \text{&#34;}$`)
	assert.Contains(t, out, "<mtext>&amp;lt;b&amp;gt;</mtext>")
	assert.NotContains(t, out, "<mtext>&#34;</mtext>")
	assert.NotContains(t, out, "<mtext>&lt;b&gt;</mtext>")
	assert.Contains(t, out, `\text{&amp;lt;b&amp;gt;} \text{&amp;#34;}</annotation>`)
}

func TestMathDoesNotSwallowText(t *testing.T) {
	cases := map[string]string{
		`\alpha is greek`:                    `\alpha is greek`,
		`$100 is the price`:                  `$100 is the price`,
		`\_underscore`:                       `_underscore`,
		`$5 and $10`:                         `$5 and $10`,
		`costs \$5 and \$6`:                  `costs $5 and $6`,
		"$$\nnever closed\n\nnext paragraph": "next paragraph",
	}
	for src, want := range cases {
		t.Run(src, func(t *testing.T) {
			out := renderMathMarkdown(t, src)
			assert.Contains(t, out, want)
			assert.NotContains(t, out, "<math")
		})
	}
}

// An unclosed $$ must not turn the blocks after it into math, even when a
// later line holds $$ (reviewer repros on PR #4879).
func TestMathUnclosedDisplayDoesNotSwallowBlocks(t *testing.T) {
	cases := []struct{ name, src, want string }{
		{"heading", "$$\noops forgot to close\n\n## Section 2\n\nMore text with $$y$$ here.\n", "<h2>Section 2</h2>"},
		{"fence", "$$\nunclosed\n\n```go\n$$\n```\n\nlast paragraph\n", "<p>last paragraph</p>"},
		{"fence without blank line", "$$\nunclosed\n```go\n$$\n```\nlast paragraph\n", "<p>last paragraph</p>"},
		{"heading without blank line", "$$\nunclosed\n# Title\n$$\n", "<h1>Title</h1>"},
		{"inline $$ is not a closing line", "$$\nunclosed\nsee $$y$$ here\n", "<p>$$\nunclosed"},
		{"blockquote ends first", "> $$\n> x^2\n$$\nafter\n", "<blockquote>\n<p>$$\nx^2"},
		{"blank line inside blockquote", "> $$\n> x\n>\n> $$\n", "<p>$$\nx"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out := renderMathMarkdown(t, c.src)
			assert.Contains(t, out, c.want)
			// Whatever was swallowed would show up in the TeX annotation.
			assert.NotContains(t, out, "unclosed\n</annotation>")
			assert.NotContains(t, out, "<mi>u</mi>")
		})
	}
}

// An unclosed $$ must not swallow a block that would interrupt a paragraph
// when a later line holds or ends with $$: display math reads like a
// paragraph, and any CommonMark block start ends it.
func TestMathUnclosedDisplayDoesNotSwallowInterruptingBlocks(t *testing.T) {
	cases := []struct{ name, src, want string }{
		{"bullet list", "$$\nx+1\n- item one\n- item $$\n", "<ul>\n<li>item one</li>\n<li>item $$</li>\n</ul>"},
		{"plus list", "$$\nx+1\n+ item $$\n", "<li>item $$</li>"},
		{"ordered list", "$$\nx+1\n1. item one\n2. item $$\n", "<ol>\n<li>item one</li>\n<li>item $$</li>\n</ol>"},
		{"blockquote", "$$\nx+1\n> quote $$\n", "<blockquote>\n<p>quote $$</p>"},
		{"html block", "$$\nx+1\n<div>\n$$\n</div>\n", "<!-- raw HTML omitted -->"},
		{"setext heading", "$$\nx+1\nTitle\n===\nmore $$\n", "<h1>$$\nx+1\nTitle</h1>"},
		{"thematic break ***", "$$\nx+1\n***\ny $$\n", "<hr>\n<p>y $$</p>"},
		{"thematic break ---", "$$\nx+1\n\n---\ny $$\n", "<hr>\n<p>y $$</p>"},
		{"line starting with $$", "$$\nx+1\n- item\n$$\n", "<li>item\n$$</li>"},
		{"brackets", "\\\\[\nx+1\n- item \\\\]\n", "<li>item \\]</li>"},
		{"inside a blockquote", "> $$\n> x+1\n> - item\n> $$\n", "<blockquote>\n<p>$$\nx+1"},
		{"inside a list item", "- $$\n  x+1\n  - item $$\n", "<li>item $$</li>"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out := renderMathMarkdown(t, c.src)
			assert.NotContains(t, out, "<math")
			assert.Contains(t, out, c.want)
		})
	}
}

// The gnoweb blocks that interrupt a paragraph end display math too: an
// unclosed $$ cannot swallow columns, a form or an alert up to a later $$.
func TestMathUnclosedDisplayDoesNotSwallowGnoBlocks(t *testing.T) {
	cases := []struct{ name, src, want string }{
		{"columns", "$$\nx+1\n<gno-columns>\ny $$\n</gno-columns>\n", `<div class="gno-columns">`},
		{"form", "$$\nx+1\n<gno-form>\n<gno-input name=\"a\" />\n</gno-form>\n$$\n", `<form class="gno-form"`},
		{"alert", "$$\nx+1\n> [!NOTE]\n> y\n$$\n", `<details class="gno-alert gno-alert-note"`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out := renderMathMarkdown(t, c.src)
			assert.NotContains(t, out, "<math")
			assert.Contains(t, out, "<p>$$\nx+1")
			assert.Contains(t, out, c.want)
		})
	}
}

// Lines of display math that do not read as a block start stay math. A line
// that does ("+ b", "- x", ">0") ends the block like it would end a
// paragraph; "{}+ b" or a four-space indent keeps it in the math.
func TestMathDisplayLinesThatLookLikeBlocks(t *testing.T) {
	for _, src := range []string{
		"$$\n-x^2 + 1\n$$\n",
		"$$\na\n{}+ b\n$$\n",
		"$$\na\n    + b\n$$\n",
		"$$\na\n2. b\n$$\n",
		"$$\na\n#b\n$$\n",
		"$$\n\\begin{aligned}\nf &= x \\\\\n  &- y\n\\end{aligned}\n$$\n",
	} {
		assert.Contains(t, renderMathMarkdown(t, src), "<math", "%q", src)
	}
	for _, src := range []string{
		"$$\na\n+ b\n$$\n",
		"$$\na\n- x\n$$\n",
		"$$\na\n> 0\n$$\n",
		"$$\na\n>0\n$$\n",
	} {
		assert.NotContains(t, renderMathMarkdown(t, src), "<math", "%q", src)
	}
}

func TestMathDisplayBlockClosing(t *testing.T) {
	for name, src := range map[string]string{
		"own line":         "$$\nx^2\n$$\n",
		"end of last line": "$$\nx^2 $$\n",
		"indented close":   "$$\nx^2\n  $$\n",
		"brackets":         "\\\\[\nx^2\n\\\\]\n",
		"blockquote":       "> $$\n> x^2\n> $$\n",
		"list item":        "- $$\n  x^2\n  $$\n- b\n",
	} {
		t.Run(name, func(t *testing.T) {
			out := renderMathMarkdown(t, src)
			assert.Contains(t, out, `<math class="math-displaystyle"`)
			assert.Contains(t, out, "<msup>")
		})
	}
}

// Like a closing code fence, a closing delimiter may be followed only by
// spaces. A line with text after it does not close the block, so the text
// is not split off into a paragraph with a stray $$: here the block never
// closes and the lines stay one paragraph, in which $$y$$ is inline math.
func TestMathDisplayCloseIsLastOnItsLine(t *testing.T) {
	for _, c := range []struct {
		name, src string
		wants     []string
	}{
		{"inline math after the delimiter", "$$\nx\n$$y$$ trailing\n", []string{
			"<p>$$\nx\n",
			`<annotation encoding="application/x-tex">y</annotation>`,
			" trailing</p>",
		}},
		{"text after the delimiter", "$$\nx\n$$ y\n", []string{"<p>$$\nx\n$$ y</p>"}},
		{"text after brackets", "\\\\[\nx\n\\\\] y\n", []string{"<p>\\[\nx\n\\] y</p>"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			out := renderMathMarkdown(t, c.src)
			for _, want := range c.wants {
				assert.Contains(t, out, want)
			}
			assert.NotContains(t, out, "<mi>x</mi>")
			assert.NotContains(t, out, "<p>y")
		})
	}

	// Spaces after the delimiter, or before it, still close the block.
	for _, src := range []string{"$$\nx\n$$  \n", "$$\nx\n  $$\n", "$$\nx $$ \n"} {
		out := renderMathMarkdown(t, src)
		assert.Contains(t, out, `<math class="math-displaystyle"`, "%q", src)
		assert.NotContains(t, out, "<p>", "%q", src)
	}
}

// Display math that fails to convert falls back to its escaped source. Inside
// a paragraph the fallback must be phrasing content: a <div> there is invalid
// HTML, and the browser closes the paragraph before it.
func TestMathDisplayFallbackElement(t *testing.T) {
	for _, c := range []struct{ name, src, want string }{
		{"dollars in a paragraph", "text $$\\begin{matrix}$$ text", `<p>text <span class="math-display">\begin{matrix}</span> text</p>`},
		{"brackets in a paragraph", "text \\\\[\\begin{matrix}\\\\] text", `<p>text <span class="math-display">\begin{matrix}</span> text</p>`},
		{"alone on its line", "$$\\begin{matrix}$$", `<p><span class="math-display">\begin{matrix}</span></p>`},
		{"block", "$$\n\\begin{matrix}\n$$\n", "<div class=\"math-display\">\n\\begin{matrix}\n</div>"},
	} {
		t.Run(c.name, func(t *testing.T) {
			out := renderMathMarkdown(t, c.src)
			assert.Contains(t, out, c.want)
			if strings.HasPrefix(out, "<p>") {
				assert.NotContains(t, out, "<div")
			}
		})
	}
}

// Code spans and fences keep their $ literally.
func TestMathNotInCode(t *testing.T) {
	for _, src := range []string{
		"a `$x$` b `$$y$$` c `\\(z\\)`",
		"```\n$$\nx\n$$\n```\n",
		"```\n$x$\n```\n",
		"    $$x$$\n",
	} {
		out := renderMathMarkdown(t, src)
		assert.NotContains(t, out, "<math", src)
		assert.Contains(t, out, "<code", src)
	}
}

// A $ inside a code span does not close an expression opened before it.
func TestMathDoesNotCrossCodeSpan(t *testing.T) {
	for src, code := range map[string]string{
		"Costs $5: see `foo$bar` in the code.":              "<code>foo$bar</code>",
		"Costs $5, then open `/r/demo/foo$help` to donate.": "<code>/r/demo/foo$help</code>",
	} {
		out := renderMathMarkdown(t, src)
		assert.NotContains(t, out, "<math", src)
		assert.Contains(t, out, code, src)
	}
	assert.Equal(t, 2, strings.Count(renderMathMarkdown(t, "$n$th and `x$y` and $a$"), "<math"))
}

func TestMathStillRenders(t *testing.T) {
	for _, src := range []string{
		`$E=mc^2$`,
		`price $x^2$ here`,
		`$$\int_0^1 x^2 dx$$`,
		"$$\n\\int_0^1 x^2 dx\n$$",
		`\\(A = \\pi r^2\\)`,
		`a \\$x$ b`,
		`$a\\$`,
	} {
		assert.Contains(t, renderMathMarkdown(t, src), "<math", src)
	}
}

// Run with -race: the renderer used to share one converter across renders.
func TestMathConcurrentRender(t *testing.T) {
	gm := goldmark.New(goldmark.WithExtensions(NewGnoExtension()))
	in := [][]byte{[]byte(`$E=mc^2$`), []byte(`$$\int_0^1 x^2 dx$$`)}
	var wg sync.WaitGroup
	for i := range 50 {
		wg.Go(func() {
			var b bytes.Buffer
			assert.NoError(t, gm.Convert(in[i%2], &b))
			assert.Contains(t, b.String(), "<math")
		})
	}
	wg.Wait()
}

func TestMathOutputIsBounded(t *testing.T) {
	t.Run("deep nesting", func(t *testing.T) {
		for _, depth := range []int{50, 1000} {
			src := "$" + strings.Repeat(`\sqrt{`, depth) + "x" + strings.Repeat("}", depth) + "$"
			out := renderMathMarkdown(t, src)
			assert.Less(t, len(out), 64*len(src)+4096, "depth %d", depth)
		}
	})
	t.Run("many unclosed openers", func(t *testing.T) {
		// Each unclosed opener looks ahead for a closing delimiter; the
		// lookahead is cached so this stays linear.
		start := time.Now()
		out := renderMathMarkdown(t, strings.Repeat("\\\\[\n", 1<<16))
		assert.NotContains(t, out, "<math")
		assert.Less(t, time.Since(start), 10*time.Second)
		// The cache is kept per container: a long quote or list item
		// full of openers is scanned once too.
		for _, src := range []string{
			strings.Repeat("> \\\\[\n", 1<<15),
			"- a\n" + strings.Repeat("  \\\\[\n", 1<<15),
		} {
			start := time.Now()
			assert.NotContains(t, renderMathMarkdown(t, src), "<math")
			assert.Less(t, time.Since(start), 10*time.Second)
		}
	})
	t.Run("too long", func(t *testing.T) {
		src := "$" + strings.Repeat("x+", MaxMathInputLen) + "x$"
		out := renderMathMarkdown(t, src)
		assert.NotContains(t, out, "<math")
		assert.Contains(t, out, `<span class="math-inline">`)
	})
}

// Each unclosed inline opener used to rescan the rest of its line, which is
// quadratic: a 1 MiB line of "$a " took about a minute to render.
func TestMathUnclosedInlineOpenersAreLinear(t *testing.T) {
	// Deterministic part: count the bytes the closing-delimiter search reads
	// when every opener on a line is unclosed, as the inline parser calls it.
	for _, tc := range []struct {
		unit string
		open int
		key  parser.ContextKey
		find closeFinder
	}{
		{"$a ", 1, closeDollarInlineKey, findDollarClose},
		{"[$a ](", 2, closeDollarInlineKey, findDollarClose},
		{`\\(a `, 3, closeInlineKey, func(b []byte) (int, int) { return bytes.Index(b, _inlineclose), len(b) }},
	} {
		unit, open, key := tc.unit, tc.open, tc.key
		line := []byte(strings.Repeat(unit, 1<<12))
		next := []byte("no closer on the next line either\n")
		pc := parser.NewContext()
		scanned := 0
		counting := func(b []byte) (int, int) {
			idx, end := tc.find(b)
			scanned += max(idx+1, end) // the bytes read
			return idx, end
		}
		lineStop, nextStart := len(line), len(line)
		for i := 0; i < len(line); i += len(unit) {
			if findCloseCached(pc, key, line[i+open:], i+open, lineStop, counting) >= 0 {
				t.Fatalf("%q: unexpected close", unit)
			}
			if findCloseCached(pc, key, next, nextStart, nextStart+len(next), counting) >= 0 {
				t.Fatalf("%q: unexpected close on next line", unit)
			}
		}
		assert.LessOrEqual(t, scanned, len(line)+len(next), "%q", unit)
	}

	// End to end, with a generous bound: the quadratic version took ~56s.
	for _, unit := range []string{"$a ", `\\(a `, `\\[a `, "$a\n"} {
		src := strings.Repeat(unit, (1<<20)/len(unit))
		start := time.Now()
		out := renderMathMarkdown(t, src)
		assert.NotContains(t, out, "<math", "%q", unit)
		assert.Less(t, time.Since(start), 10*time.Second, "%q", unit)
	}
}

func TestMathCachedCloseSearchStillPairs(t *testing.T) {
	for src, n := range map[string]int{
		"$a $b$ $c$ $d":        2,
		`\\(a \\(b\\) \\(c\\)`: 2,
		"$a\nb$ $c\nd$":        2,
		"$5 and $x$ and $y$":   2,
		"$a \\$ b$":            1,
	} {
		assert.Equal(t, n, strings.Count(renderMathMarkdown(t, src), "<math"), src)
	}
}

// MaxMathInputLen bounds one expression, not a page: a 1 MiB page of
// aligned environments of bare & used to render to 112 MB of HTML.
func TestMathPageOutputIsBounded(t *testing.T) {
	for name, para := range map[string]string{
		"aligned of bare &": "$\\begin{aligned}" + strings.Repeat("&", MaxMathInputLen-32) + "\\end{aligned}$\n\n",
		"aligned of &x":     "$\\begin{aligned}" + strings.Repeat("&x", MaxMathInputLen/2-32) + "\\end{aligned}$\n\n",
		"pmatrix of bare &": "$$\\begin{pmatrix}" + strings.Repeat("&", MaxMathInputLen-32) + "\\end{pmatrix}$$\n\n",
		"tiny inline math":  "$a$ ",
		"tiny display math": "$$a$$ ",
		"long superscripts": "$" + strings.Repeat("x^", 4000) + "x$\n\n",
	} {
		t.Run(name, func(t *testing.T) {
			src := strings.Repeat(para, (1<<20)/len(para))
			out := renderMathMarkdown(t, src)
			// The MathML is capped; past the cap every expression is
			// escaped source, which costs a few bytes per input byte.
			assert.Less(t, len(out), MaxMathPageOutput+12*len(src))
			assert.Contains(t, out, `class="math-`)
		})
	}
}

// An empty expression is not math: $$$$ used to cost a whole <math> element
// per four input bytes.
func TestMathEmptyExpressionIsText(t *testing.T) {
	for src, want := range map[string]string{
		"$$$$":           "<p>$$$$</p>",
		"a $$ $$ b":      "<p>a $$ $$ b</p>",
		`\\(\\)`:         `<p>\(\)</p>`,
		`\\( \\)`:        `<p>\( \)</p>`,
		"$$\n$$\n":       "<p>$$\n$$</p>",
		"$$\n \t \n$$\n": "<p>$$</p>",
		"\\\\[\n\\\\]\n": "<p>\\[\n\\]</p>",
	} {
		out := renderMathMarkdown(t, src)
		assert.NotContains(t, out, "<math", "%q", src)
		assert.Contains(t, out, want, "%q", src)
	}
	// Non-empty neighbours still render.
	assert.Equal(t, 1, strings.Count(renderMathMarkdown(t, "$$$$ $x$"), "<math"))
}

// Every <math> element carries a few hundred bytes of fixed markup, so a page
// of nothing but tiny expressions must be bounded by its own size, like one
// expression, and not only by MaxMathPageOutput.
func TestMathTinyExpressionsPageIsLinear(t *testing.T) {
	for _, unit := range []string{"$a$", "$a$$b$ ", "$$a$$", `\\(a\\)`} {
		src := strings.Repeat(unit, 4096/len(unit))
		out := renderMathMarkdown(t, src)
		assert.LessOrEqual(t, len(out), maxMathOutputLen(len(src))+maxMarkupRatio*len(src), "%q", unit)
		assert.Contains(t, out, "<math", "%q", unit)
	}
	// At three bytes an expression, the budget runs out partway: earlier
	// expressions render, later ones fall back to escaped text.
	out := renderMathMarkdown(t, strings.Repeat("$a$", 1000))
	converted := strings.Count(out, "<math")
	assert.Greater(t, converted, 500)
	assert.Equal(t, 1000-converted, strings.Count(out, `<span class="math-inline">a</span>`))
	// Prose between expressions leaves room for all of them.
	src := strings.Repeat("the value $x_i$ is positive, ", 200)
	assert.Equal(t, 200, strings.Count(renderMathMarkdown(t, src), "<math"))
}

func TestMathPageBudgetFallsBackToText(t *testing.T) {
	expr := "$" + strings.Repeat("x^", 1000) + "x$\n\n"
	out := renderMathMarkdown(t, strings.Repeat(expr, 200))
	converted := strings.Count(out, "<math")
	assert.Greater(t, converted, 0)
	assert.Less(t, converted, 200)
	assert.Equal(t, 200-converted, strings.Count(out, `<span class="math-inline">`))
	// Once the budget is spent, nothing after is converted.
	last := strings.LastIndex(out, "<math")
	assert.NotContains(t, out[:last], `<span class="math-inline">`)
}

func TestMathExpressionAmplificationIsBounded(t *testing.T) {
	// The densest inputs known; the past-the-cap fallback itself is
	// exercised by TestMathPageBudgetFallsBackToText.
	for _, src := range []string{
		"$\\begin{aligned}" + strings.Repeat("&", MaxMathInputLen-32) + "\\end{aligned}$",
		"$\\begin{aligned}" + strings.Repeat("&x", MaxMathInputLen/2-32) + "\\end{aligned}$",
		"$" + strings.Repeat("x^", MaxMathInputLen/2-32) + "x$",
	} {
		out := renderMathMarkdown(t, src)
		assert.LessOrEqual(t, len(out), maxMathOutputLen(len(src)), "%.30s", src)
	}
}

// maxMarkupRatio bounds the HTML a page produces outside MathML, per input
// byte: an unconverted $&$ is <span class="math-inline">&amp;</span>.
const maxMarkupRatio = 16

// FuzzMathRender checks that math input cannot inject script, event handlers
// or javascript: URLs, and that the output size stays linear in the input.
func FuzzMathRender(f *testing.F) {
	for _, seed := range []string{
		`E=mc^2`,
		`\text{</math><script>alert(1)</script>}`,
		`\class{x" onclick="alert(1)}{y}`,
		`\color{red" onmouseover="alert(1)}{y}`,
		`\begin{matrix}</span><script>alert(1)</script>`,
		`\begin{pmatrix} a & b \\ c & d \end{pmatrix}`,
		`\sqrt[3]{\frac{a}{b}} \overset{!}{=} \mathop{lim}`,
		`\raisebox{1em}{x} \textcolor{red}{y} \multirow{2}{a}`,
		`\begin{aligned}&x&x\\x&x\end{aligned}`,
		strings.Repeat("$", 64),                                            // empty $$$$ expressions
		strings.Repeat("a$$", 64),                                          // tiny inline expressions
		strings.Repeat("$a$$$", 64),                                        // tiny display expressions
		">[!0]",                                                            // an alert, which draws an <svg> icon
		`\left(\color x{\right)}`,                                          // a group crossing a fence
		strings.Repeat(`{\left(`, 8) + "x" + strings.Repeat(`}\right)`, 8), // crossing groups, nested
	} {
		f.Add(seed)
	}
	gm := goldmark.New(goldmark.WithExtensions(NewGnoExtension()))
	f.Fuzz(func(t *testing.T, tex string) {
		for _, src := range []string{"$" + tex + "$", "$$" + tex + "$$", "$$\n" + tex + "\n$$"} {
			out, ok := convertWithin(t, gm, src, 10*time.Second)
			if !ok {
				return
			}
			// The MathML of a page is bounded like one expression of the
			// page's size; the HTML around it, and the escaped source of
			// expressions left unconverted, by a small multiple of the input.
			if limit := maxMathOutputLen(len(src)) + maxMarkupRatio*len(src); len(out) > limit {
				t.Fatalf("output too large: %d bytes for %d bytes of input (limit %d)", len(out), len(src), limit)
			}
			if err := checkNoActiveContent(out); err != nil {
				t.Fatalf("%v in output for %q:\n%s", err, src, out)
			}
		}
	})
}

// convertWithin renders src, failing the test if that takes longer than
// timeout: fuzzing would otherwise sit on an input that loops forever
// without reporting it. ok is false if goldmark returned an error.
func convertWithin(t *testing.T, gm goldmark.Markdown, src string, timeout time.Duration) (out string, ok bool) {
	t.Helper()
	type result struct {
		out string
		err error
	}
	done := make(chan result, 1)
	go func() {
		var buf bytes.Buffer
		err := gm.Convert([]byte(src), &buf)
		done <- result{buf.String(), err}
	}()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case r := <-done:
		return r.out, r.err == nil
	case <-timer.C:
		t.Fatalf("rendering %q took over %v", src, timeout)
		return "", false
	}
}

// checkNoActiveContent returns an error if html holds an element or
// attribute that runs script or embeds content. The one <svg> allowed is the
// icon gnoweb's alerts, links and forms draw, <svg><use
// href="#ico-..."></use></svg>: an <svg> immediately holding a <use> that
// refers to one of the page's own icons, and nothing else.
func checkNoActiveContent(out string) error {
	z := html.NewTokenizer(strings.NewReader(out))
	for tt := z.Next(); tt != html.ErrorToken; tt = z.Next() {
		if tt != html.StartTagToken && tt != html.SelfClosingTagToken {
			continue
		}
		tok := z.Token()
		for _, a := range tok.Attr {
			if strings.HasPrefix(a.Key, "on") ||
				strings.HasPrefix(strings.ToLower(strings.TrimSpace(a.Val)), "javascript:") {
				return fmt.Errorf("unexpected %s attribute", a.Key)
			}
		}
		switch tok.Data {
		case "svg":
			if !isIconSVG(z) {
				return fmt.Errorf("unexpected <svg>")
			}
		case "script", "style", "iframe", "object", "embed", "use":
			return fmt.Errorf("unexpected <%s>", tok.Data)
		}
	}
	return nil
}

// isIconSVG reads the tokens after an <svg> start tag and reports whether
// they are exactly <use href="#ico-..."></use></svg>.
func isIconSVG(z *html.Tokenizer) bool {
	if z.Next() != html.StartTagToken {
		return false
	}
	use := z.Token()
	if use.Data != "use" || len(use.Attr) != 1 || use.Attr[0].Key != "href" || !strings.HasPrefix(use.Attr[0].Val, "#ico-") {
		return false
	}
	if z.Next() != html.EndTagToken || z.Token().Data != "use" {
		return false
	}
	return z.Next() == html.EndTagToken && z.Token().Data == "svg"
}

func TestMathFuzzOracle(t *testing.T) {
	for _, ok := range []string{
		`<p>a</p>`,
		`<summary><svg><use href="#ico-note"></use></svg>Note</summary>`,
		`<a href="u">x<svg class="c-icon"><use href="#ico-external-link"></use></svg></a>`,
	} {
		assert.NoError(t, checkNoActiveContent(ok), ok)
	}
	for _, bad := range []string{
		`<svg></svg>`,
		`<svg><use href="#ico-note"></use><script>alert(1)</script></svg>`,
		`<svg><use href="https://evil.example/x.svg#a"></use></svg>`,
		`<svg><use href="#ico-x" onload="alert(1)"></use></svg>`,
		`<svg><image href="x"></image></svg>`,
		`<svg onload="alert(1)"><use href="#ico-x"></use></svg>`,
		`<use href="#ico-x"></use>`,
		`<math><mi onclick="alert(1)">x</mi></math>`,
		`<a href=" JavaScript:alert(1)">x</a>`,
		`<script>alert(1)</script>`,
	} {
		assert.Error(t, checkNoActiveContent(bad), bad)
	}
}

// BenchmarkMathDisplayBlockLines parses a 200-line display block: every line
// is checked for a block start twice, by the lookahead in Open and by
// Continue.
func BenchmarkMathDisplayBlockLines(b *testing.B) {
	src := []byte("$$\n" + strings.Repeat("x_{1} + y^{2} \\\\\n", 200) + "$$\n")
	gm := goldmark.New(goldmark.WithExtensions(NewGnoExtension()))
	b.ReportAllocs()
	for b.Loop() {
		gm.Parser().Parse(text.NewReader(src))
	}
}

// A paragraph of one expression per line used to be quadratic: the inline
// parser read each expression back through block.Value, which walks the
// paragraph's lines back from its last one. 1 MiB of "$a$\n" took ~20s.
func TestMathOneExpressionPerLineIsLinear(t *testing.T) {
	units := []string{"$a$\n", `\\(a\\)` + "\n", "$$a$$\n", "$a\nb$\n", "$$ $$\n", "> $a$\n", "> $a\n> b$\n"}
	// Rendering eight times the lines must cost about eight times as much:
	// the quadratic version cost sixty-four times as much. Inputs are large
	// enough to dominate timer noise, and the best of three runs after a GC
	// keeps scheduler and collector noise out of the ratio.
	best := func(src string) time.Duration {
		d := time.Duration(1<<63 - 1)
		for range 3 {
			runtime.GC()
			start := time.Now()
			renderMathMarkdown(t, src)
			d = min(d, time.Since(start))
		}
		return d
	}
	for _, unit := range units {
		small, large := strings.Repeat(unit, 1<<12), strings.Repeat(unit, 1<<15)
		ratio := float64(best(large)) / float64(best(small))
		assert.Less(t, ratio, 24.0, "%q: 8x the input took %.1fx the time", unit, ratio)
	}
	// End to end, with a generous bound.
	for _, unit := range []string{"$a$\n", "> $a$\n"} {
		src := strings.Repeat(unit, (1<<20)/len(unit))
		start := time.Now()
		out := renderMathMarkdown(t, src)
		assert.Less(t, time.Since(start), 10*time.Second, "%q", unit)
		assert.Contains(t, out, "<math", "%q", unit)
	}
}

// Finding the containers of a $$ opener compared every open block with every
// ancestor, which is quadratic in the nesting depth: 1 MiB of "$$" in 2000
// nested list items took 22s, against 2s without math. goldmark is itself
// quadratic in the depth, so the cost is compared with the same lines
// without math: the quadratic version took over twenty times as long.
func TestMathDisplayOpenerInDeepContainersIsLinear(t *testing.T) {
	best := func(src string) time.Duration {
		d := time.Duration(1<<63 - 1)
		for range 3 {
			runtime.GC()
			start := time.Now()
			renderMathMarkdown(t, src)
			d = min(d, time.Since(start))
		}
		return d
	}
	for _, prefix := range []string{"- ", "> "} {
		line := strings.Repeat(prefix, 1000) + "%s\n\n"
		math := strings.Repeat(fmt.Sprintf(line, "$$"), 64)
		text := strings.Repeat(fmt.Sprintf(line, "xx"), 64)
		ratio := float64(best(math)) / float64(best(text))
		assert.Less(t, ratio, 4.0, "%q: $$ took %.1fx the time of text", prefix, ratio)
	}
}

// The expression is read from the parsed lines, without the container
// prefix of the line it continues on.
func TestMathInlineAcrossLinesInContainers(t *testing.T) {
	for src, want := range map[string]string{
		"> $a\n> b$":           "a\nb</annotation>",
		"- $a\n  b$":           "a\nb</annotation>",
		"> - $a\n>   b$":       "a\nb</annotation>",
		"$a\nb$":               "a\nb</annotation>",
		`\\(a` + "\n" + `b\\)`: "a\nb</annotation>",
	} {
		assert.Contains(t, renderMathMarkdown(t, src), want, "%q", src)
	}
}

// A math block opens only if its closing line is in the same container
// (blockquote, list item, ...). Otherwise the opener is paragraph text, and
// the lines after it are read exactly as without math: lazy continuation
// lines stay in the quote or item, and no text is lost to an HTML block
// started outside it (reviewer repros on PR #4879).
func TestMathDisplayCloseInSameContainer(t *testing.T) {
	for _, c := range []struct{ src, want string }{
		{"> $$\n> a\n<br>\nlost words\n$$", "<blockquote>\n<p>$$\na\n<!-- raw HTML omitted -->\nlost words\n$$</p>\n</blockquote>\n"},
		{"- $$\n  a\n<span>\nlost\n$$", "<ul>\n<li>$$\na\n<!-- raw HTML omitted -->\nlost\n$$</li>\n</ul>\n"},
		{"> $$\n> a\nlazy words\n$$", "<blockquote>\n<p>$$\na\nlazy words\n$$</p>\n</blockquote>\n"},
		{"> $$\nx\n> $$", "<blockquote>\n<p>$$\nx\n$$</p>\n</blockquote>\n"},
		// Inline markdown in the lines is rendered, not shown as source.
		{"> $$\n> a &amp; b \\* c **d**\nlazy\n> $$", "<blockquote>\n<p>$$\na &amp; b * c <strong>d</strong>\nlazy\n$$</p>\n</blockquote>\n"},
		{"> > $$\n> > a\n> $$", "<blockquote>\n<blockquote>\n<p>$$\na\n$$</p>\n</blockquote>\n</blockquote>\n"},
		{"- $$\n  a\n- $$", "<ul>\n<li>$$\na</li>\n<li>$$</li>\n</ul>\n"},
		// The paragraph the list interrupts is not a container of the item.
		{"a\nb\n- \\\\[\nc", "<p>a\nb</p>\n<ul>\n<li>\\[\nc</li>\n</ul>\n"},
		{"> a\n> - $$\n> c", "<blockquote>\n<p>a</p>\n<ul>\n<li>$$\nc</li>\n</ul>\n</blockquote>\n"},
		// Probing the empty item for the first opener must not leave the
		// list parser's state behind for the second one's probes.
		{"- $$\n-\n$$\n- $$", "<ul>\n<li>$$</li>\n<li></li>\n</ul>\n<p>$$</p>\n<ul>\n<li>$$</li>\n</ul>\n"},
	} {
		assert.Equal(t, c.want, renderMathMarkdown(t, c.src), "%q", c.src)
	}

	// A closing line in the same container closes the block, however deep.
	for _, src := range []string{
		"> $$\n> x^2\n> $$\n",
		"> > $$\n> > x^2\n> > $$\n",
		"- $$\n  x^2\n  $$\n",
		"> - $$\n>   x^2\n>   $$\n",
		"- > $$\n  > x^2\n  > $$\n",
		"1. a\n\n   $$\n   x^2\n   $$\n",
		"> [!NOTE]\n> $$\n> x^2\n> $$\n",
		"> para\n> $$\n> x^2\n> $$\n",
		"para\n- $$\n  x^2\n  $$\n",
	} {
		out := renderMathMarkdown(t, src)
		assert.Contains(t, out, "<msup>", "%q", src)
		assert.NotContains(t, out, "$$", "%q", src)
	}
}

// A block closed on the line after its opener holds no math and stays
// paragraph text, so the lines after it are read as without math.
func TestMathEmptyDisplayBlockIsParagraph(t *testing.T) {
	for src, want := range map[string]string{
		"$$\n$$\n<br>\nkept":         "<p>$$\n$$\n<!-- raw HTML omitted -->\nkept</p>\n",
		"> $$\n> $$\n> <br>\n> kept": "<blockquote>\n<p>$$\n$$\n<!-- raw HTML omitted -->\nkept</p>\n</blockquote>\n",
		"\\\\[\n  \\\\]\nkept":       "<p>\\[\n\\]\nkept</p>\n",
	} {
		assert.Equal(t, want, renderMathMarkdown(t, src), "%q", src)
	}
}

// After a closed math block, as after any block that is not a paragraph, a
// line of raw HTML starts an HTML block (CommonMark type 7), which gnoweb
// omits up to the next blank line. That is what a fenced code block in the
// same place does, so the math does not change it.
func TestMathHTMLBlockAfterDisplayMath(t *testing.T) {
	const after = "<span>\ntext after\n\nnext paragraph\n"
	fence := renderMathMarkdown(t, "```\nx\n```\n"+after)
	math := renderMathMarkdown(t, "$$\nx\n$$\n"+after)
	assert.True(t, strings.HasPrefix(fence, "<pre><code>x\n</code></pre>\n"), fence)
	assert.True(t, strings.HasSuffix(math, strings.TrimPrefix(fence, "<pre><code>x\n</code></pre>\n")), math)
	assert.Contains(t, math, "<!-- raw HTML omitted -->\n<p>next paragraph</p>")
}

// The lookahead in Open makes sure the block closes, but should Continue end
// it first, Close turns it back into a paragraph of the same lines rather
// than showing its source.
func TestMathUnclosedBlockBecomesParagraph(t *testing.T) {
	src := []byte("$$\na **b**\n")
	doc := ast.NewDocument()
	n := &mathBlockNode{openLen: 2, closeTag: _dollarDisplay}
	n.Lines().Append(text.NewSegment(0, 3))
	n.Lines().Append(text.NewSegment(3, len(src)))
	doc.AppendChild(doc, n)
	(&texBlockRegionParser{}).Close(n, text.NewReader(src), parser.NewContext())
	para, ok := doc.FirstChild().(*ast.Paragraph)
	require.True(t, ok, "got %T", doc.FirstChild())
	assert.Equal(t, "$$\na **b**", string(para.Lines().Value(src)))
}

// Dollars around a reference to a defined footnote are not math: the
// reference would be swallowed, and goldmark drops a footnote nothing refers
// to, so the note itself would vanish. A display block holding one is the
// paragraph it would be without math.
func TestMathKeepsFootnoteReferences(t *testing.T) {
	gm := goldmark.New(goldmark.WithExtensions(extension.Footnote, NewGnoExtension()))
	render := func(src string) string {
		var buf bytes.Buffer
		require.NoError(t, gm.Convert([]byte(src), &buf))
		return buf.String()
	}
	for _, src := range []string{
		"It costs $5[^1] or 4$ here.\n\n[^1]: the note",
		"It costs $$5[^1] or 4$$ here.\n\n[^1]: the note",
		"a $x[^1]\ny$ b\n\n[^1]: the note",
		"> [^1]: the note\n\n$a [^x] [^1]$",
		"$$\nx[^1]\n$$\n\n[^1]: the note",
		"> $$\n>   x [^1]\n> $$\n\n[^1]: the note",
	} {
		out := render(src)
		assert.NotContains(t, out, "<math", "%q", src)
		assert.Contains(t, out, `class="footnote-ref"`, "%q", src)
		assert.Contains(t, out, "the note", "%q", src)
	}
	// Brackets that are not a defined footnote stay math.
	for _, src := range []string{"$[^1]$", "$x[^2]$\n\n[^1]: note", "$[0,1]^2$", "$$\nx[^2]\n$$\n\n[^1]: note"} {
		assert.Contains(t, render(src), "<math", "%q", src)
	}
}

// Display math ends where a block of a peer extension's parser would
// interrupt a paragraph, once the parser is declared.
func TestMathEndsAtPeerBlocks(t *testing.T) {
	src := []byte("$$\nx\n[^1]: note\n$$\n")
	for _, c := range []struct {
		ext  *GnoExtension
		math bool
	}{
		{NewGnoExtension(), true},
		{NewGnoExtension(WithPeerBlockParsers(extension.NewFootnoteBlockParser())), false},
	} {
		var buf bytes.Buffer
		gm := goldmark.New(goldmark.WithExtensions(extension.Footnote, c.ext))
		require.NoError(t, gm.Convert(src, &buf))
		assert.Equal(t, c.math, strings.Contains(buf.String(), "<math"), buf.String())
	}
}

// A fenced code block whose info string is exactly "math" is display math,
// as on GitHub. Any other fence is code.
func TestMathFence(t *testing.T) {
	for _, src := range []string{
		"```math\nx^2\n```\n",
		"~~~math\nx^2\n~~~\n",
		"```math\n\\begin{aligned}\na &= x^2 \\\\\n\nb &= y\n\\end{aligned}\n```\n",
		"> ```math\n> x^2\n> ```\n",
		"- item\n\n  ```math\n  x^2\n  ```\n",
	} {
		out := renderMathMarkdown(t, src)
		assert.Contains(t, out, `<math class="math-displaystyle"`, "%q", src)
		assert.Contains(t, out, "<msup>", "%q", src)
		assert.NotContains(t, out, "<pre>", "%q", src)
	}
	for _, src := range []string{
		"```Math\nx^2\n```\n",
		"```math x\nx^2\n```\n",
		"```latex\nx^2\n```\n",
		"```\nx^2\n```\n",
		"```math\n```\n",
		"    math\n    x^2\n",
	} {
		out := renderMathMarkdown(t, src)
		assert.NotContains(t, out, "<math", "%q", src)
		assert.Contains(t, out, "<pre><code", "%q", src)
	}
}

// A math fence that is not converted (too long, a conversion error, the
// page's budget spent) is shown as the code block it is without math, not
// as a run of escaped source: a fence keeps its lines.
func TestMathFenceFallsBackToCode(t *testing.T) {
	for name, tex := range map[string]string{
		"conversion error": "\\begin{matrix}\na</pre>\n",
		"too long":         strings.Repeat("x+", MaxMathInputLen) + "x\n",
	} {
		t.Run(name, func(t *testing.T) {
			out := renderMathMarkdown(t, "```math\n"+tex+"```\n")
			assert.Equal(t, `<pre><code class="language-math">`+html.EscapeString(tex)+"</code></pre>\n", out)
		})
	}
	t.Run("page budget", func(t *testing.T) {
		fence := "```math\n" + strings.Repeat("x^", 1000) + "x\n```\n\n"
		out := renderMathMarkdown(t, strings.Repeat(fence, 200))
		converted := strings.Count(out, "<math")
		assert.Greater(t, converted, 0)
		assert.Equal(t, 200-converted, strings.Count(out, `<pre><code class="language-math">`))
	})
}

// Math shows its source, delimiters included, where gnoweb reads a node's
// text instead of rendering it: the table of contents and image alt text,
// which read as without math.
func TestMathTextIsItsSource(t *testing.T) {
	src := []byte("## Energy $E=mc^2$ law\n\n## $E=mc^2$\n\n## a \\\\(x\\\\) b $$y$$\n")
	gm := goldmark.New(goldmark.WithExtensions(NewGnoExtension()), goldmark.WithParserOptions(parser.WithAutoHeadingID()))
	doc := gm.Parser().Parse(text.NewReader(src))
	toc, err := TocInspect(doc, src, TocOptions{MaxDepth: 6})
	require.NoError(t, err)
	var titles []string
	for _, item := range toc.Items {
		titles = append(titles, item.Title)
	}
	assert.Equal(t, []string{"Energy $E=mc^2$ law", "$E=mc^2$", `a \(x\) b $$y$$`}, titles)

	for src, want := range map[string]string{
		"![a $x$ b](u)":           `alt="a $x$ b"`,
		"![$a<b$ \\\\(c\\\\)](u)": `alt="$a&lt;b$ \(c\)"`,
		"![a $x\nb$ c](u)":        "alt=\"a $x\nb$ c\"",
	} {
		assert.Contains(t, renderMathMarkdown(t, src), want, "%q", src)
	}
}
