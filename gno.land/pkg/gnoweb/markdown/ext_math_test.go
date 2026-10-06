package markdown

import (
	"bytes"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
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
	for _, c := range []struct{ name, src, want string }{
		{"inline math after the delimiter", "$$\nx\n$$y$$ trailing\n", "<p>$$\nx\n"},
		{"text after the delimiter", "$$\nx\n$$ y\n", "<p>$$\nx\n$$ y</p>"},
		{"text after brackets", "\\\\[\nx\n\\\\] y\n", "<p>\\[\nx\n\\] y</p>"},
	} {
		t.Run(c.name, func(t *testing.T) {
			out := renderMathMarkdown(t, c.src)
			assert.Contains(t, out, c.want)
			assert.NotContains(t, out, "<mi>x</mi>")
			assert.NotContains(t, out, "<p>y")
		})
	}
	out := renderMathMarkdown(t, "$$\nx\n$$y$$ trailing\n")
	assert.Contains(t, out, `<annotation encoding="application/x-tex">y</annotation>`)
	assert.Contains(t, out, " trailing</p>")

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
		find func([]byte) int
	}{
		{"$a ", 1, closeDollarInlineKey, findDollarClose},
		{`\\(a `, 3, closeInlineKey, func(b []byte) int { return bytes.Index(b, _inlineclose) }},
	} {
		unit, open, key := tc.unit, tc.open, tc.key
		line := []byte(strings.Repeat(unit, 1<<12))
		next := []byte("no closer on the next line either\n")
		pc := parser.NewContext()
		scanned := 0
		counting := func(b []byte) int { scanned += len(b); return tc.find(b) }
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
		"\\\\[\n\\\\]\n": "<p>\\\\[\n\\\\]</p>",
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

	// Cell alignment is a short attribute, never inline CSS. Empty cells
	// keep it: it spaces the column pairs of an aligned environment.
	out := renderMathMarkdown(t, "$\\begin{aligned}&x\\end{aligned}$")
	assert.Contains(t, out, `<mtd columnalign="right"></mtd>`)
	assert.NotContains(t, out, "text-align")
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
		strings.Repeat("$", 64),     // empty $$$$ expressions
		strings.Repeat("a$$", 64),   // tiny inline expressions
		strings.Repeat("$a$$$", 64), // tiny display expressions
	} {
		f.Add(seed)
	}
	gm := goldmark.New(goldmark.WithExtensions(NewGnoExtension()))
	f.Fuzz(func(t *testing.T, tex string) {
		for _, src := range []string{"$" + tex + "$", "$$" + tex + "$$", "$$\n" + tex + "\n$$"} {
			var buf bytes.Buffer
			if err := gm.Convert([]byte(src), &buf); err != nil {
				return
			}
			out := buf.String()
			// The MathML of a page is bounded like one expression of the
			// page's size; the HTML around it, and the escaped source of
			// expressions left unconverted, by a small multiple of the input.
			if limit := maxMathOutputLen(len(src)) + maxMarkupRatio*len(src); len(out) > limit {
				t.Fatalf("output too large: %d bytes for %d bytes of input (limit %d)", len(out), len(src), limit)
			}
			z := html.NewTokenizer(strings.NewReader(out))
			for tt := z.Next(); tt != html.ErrorToken; tt = z.Next() {
				if tt != html.StartTagToken && tt != html.SelfClosingTagToken {
					continue
				}
				tok := z.Token()
				switch tok.Data {
				case "script", "style", "iframe", "object", "embed", "svg":
					t.Fatalf("unexpected <%s> in output for %q:\n%s", tok.Data, src, out)
				}
				for _, a := range tok.Attr {
					if strings.HasPrefix(a.Key, "on") ||
						strings.HasPrefix(strings.ToLower(strings.TrimSpace(a.Val)), "javascript:") {
						t.Fatalf("unexpected %s attribute in output for %q:\n%s", a.Key, src, out)
					}
				}
			}
		}
	})
}
