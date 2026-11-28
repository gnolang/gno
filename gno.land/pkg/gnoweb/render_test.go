package gnoweb

import (
	"bufio"
	bytes "bytes"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"

	md "github.com/gnolang/gno/gno.land/pkg/gnoweb/markdown"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// errWriter always returns an error on Write.
type errWriter struct{ err error }

func (w *errWriter) Write([]byte) (int, error) { return 0, w.err }

// limitWriter allows writing up to limit bytes, then returns an error.
type limitWriter struct {
	limit   int
	written int
}

func (w *limitWriter) Write(p []byte) (int, error) {
	if w.written+len(p) > w.limit {
		return 0, errors.New("write limit exceeded")
	}
	w.written += len(p)
	return len(p), nil
}

func newTestRenderer() *HTMLRenderer {
	return NewHTMLRenderer(slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)), NewDefaultRenderConfig(), nil)
}

func TestHTMLRenderer_RenderDocumentation_FencedBlock(t *testing.T) {
	t.Parallel()
	r := newTestRenderer()
	var buf bytes.Buffer
	err := r.RenderDocumentation(&buf, []byte("Intro.\n\n```go\nfmt.Println(\"hi\")\n```\n"))
	require.NoError(t, err)
	out := buf.String()
	require.Contains(t, out, "Intro")
	require.Contains(t, out, `<details class="b-doc-example">`)
}

// A doc comment is its package author's: no link in it passes on authority,
// autolinks included.
func TestHTMLRenderer_RenderDocumentation_LinksAreUGC(t *testing.T) {
	t.Parallel()
	r := newTestRenderer()
	var buf bytes.Buffer
	require.NoError(t, r.RenderDocumentation(&buf, []byte("See [x](https://example.org/), <https://example.net/> and [in](/r/x).\n")))
	out := buf.String()
	require.Equal(t, 3, strings.Count(out, `rel="noopener nofollow ugc"`), out)
}

func TestHTMLRenderer_RenderDocumentation_StripsRawHTML(t *testing.T) {
	t.Parallel()
	// Raw HTML in doc strings is stripped by Goldmark's default safe mode
	// and replaced with an omitted-comment placeholder.
	r := newTestRenderer()
	var buf bytes.Buffer
	err := r.RenderDocumentation(&buf, []byte("<script>alert('xss')</script>"))
	require.NoError(t, err)
	out := buf.String()
	require.NotContains(t, out, "<script>")
	require.NotContains(t, out, "alert('xss')")
}

func TestHTMLRenderer_RenderDocumentation_EmptyInput(t *testing.T) {
	t.Parallel()
	r := newTestRenderer()
	var buf bytes.Buffer
	require.NoError(t, r.RenderDocumentation(&buf, nil))
	require.Empty(t, buf.String())
}

func TestHTMLRenderer_RenderDocumentation_EmphasisGuardActive(t *testing.T) {
	r := newTestRenderer()
	var buf bytes.Buffer
	src := []byte(strings.Repeat("*x* ", md.MaxEmphasisDelimitersPerBlock*3))
	require.NoError(t, r.RenderDocumentation(&buf, src))
	// Over-cap emphasis renders as literal text, so the <em> count is bounded by
	// the per-block cap rather than the span count — proving documentationGM
	// carries the guard.
	n := strings.Count(buf.String(), "<em>")
	assert.LessOrEqualf(t, n, md.MaxEmphasisDelimitersPerBlock, "documentationGM lacks the emphasis guard: <em>=%d", n)
}

func TestRenderer_RenderRealm_Markdown(t *testing.T) {
	r := newTestRenderer()
	w := &bytes.Buffer{}
	u := &weburl.GnoURL{Path: "/r/test"}
	src := []byte(`# Hello\n\nThis is a **test**.`)
	toc, err := r.RenderRealm(w, u, src, RealmRenderContext{
		ChainId: "dev",
		Remote:  "127.0.0.1:26657",
	})
	require.NoError(t, err)
	assert.Regexp(t, "<h1[^>]*>.*Hello.*</h1>", w.String())
	assert.Contains(t, w.String(), "<strong>test</strong>")
	assert.NotNil(t, toc)
}

// Display math ends at any line that would interrupt a paragraph, footnote
// definitions included: the default config hands the footnote extension's
// block parser to the math extension.
func TestRenderRealm_MathEndsAtFootnoteDefinition(t *testing.T) {
	r := newTestRenderer()
	w := &bytes.Buffer{}
	src := []byte("$$\nx\n[^1]: note\n$$\n\nsee[^1]")
	_, err := r.RenderRealm(w, &weburl.GnoURL{Path: "/r/test"}, src, RealmRenderContext{ChainId: "dev"})
	require.NoError(t, err)
	assert.NotContains(t, w.String(), "<math")
	assert.Contains(t, w.String(), "<p>$$\nx</p>")
	assert.Contains(t, w.String(), `<li id="fn:1">`)
}

// A ```math fence is display math; the syntax highlighter still renders
// every other fence, and a math fence the converter rejects.
func TestRenderRealm_MathFence(t *testing.T) {
	r := newTestRenderer()
	w := &bytes.Buffer{}
	src := []byte("```math\nx^2\n```\n\n```go\nx := 1\n```\n\n```math\n\\begin{matrix}\n```\n")
	_, err := r.RenderRealm(w, &weburl.GnoURL{Path: "/r/test"}, src, RealmRenderContext{ChainId: "dev"})
	require.NoError(t, err)
	out := w.String()
	assert.Equal(t, 1, strings.Count(out, "<math"), out)
	assert.Contains(t, out, `<span class="chroma-nx">x</span>`)
	assert.Contains(t, out, "<pre><code class=\"language-math\">\\begin{matrix}\n</code></pre>")
}

func TestRenderRealm_ForeignBudgetResetsPerCall(t *testing.T) {
	r := newTestRenderer()
	var src strings.Builder
	for range md.MaxGnoForeignBlocksPerConvert {
		src.WriteString("\n\n<gno-foreign>\nok\n</gno-foreign>\n")
	}
	u := &weburl.GnoURL{Path: "/r/test"}
	for i := range 2 {
		var out bytes.Buffer
		_, err := r.RenderRealm(&out, u, []byte(src.String()), RealmRenderContext{})
		require.NoError(t, err)
		assert.Equal(t, md.MaxGnoForeignBlocksPerConvert, strings.Count(out.String(), `class="gno-foreign"`), "render %d", i+1)
	}
}

func TestRenderRealm_OverSizeCapServesEscapedPlainText(t *testing.T) {
	r := newTestRenderer()
	// > 1 MiB. If goldmark ran, output would contain <h1>; if the raw bytes were
	// reflected unescaped, it would contain <script>. The fallback must do neither.
	src := []byte(strings.Repeat("# Heading <script>\n\n", (maxMarkdownRenderBytes/20)+100))
	require.Greater(t, len(src), maxMarkdownRenderBytes)

	var buf bytes.Buffer
	u := &weburl.GnoURL{Path: "/r/mock/path"}
	toc, err := r.RenderRealm(&buf, u, src, RealmRenderContext{})
	require.NoError(t, err)
	assert.Empty(t, toc.Toc.Items)
	out := buf.String()
	assert.NotContains(t, out, "<h1", "goldmark must not run on oversize input")
	assert.NotContains(t, out, "<script>", "raw content must be escaped")
	assert.Contains(t, out, "&lt;script&gt;", "content is shown as escaped plain text")
	assert.Contains(t, out, "gno-alert-warning", "the notice reuses the markdown warning-alert styling")
	assert.Contains(t, out, "too large to render", "a notice explains why the page is unstyled")
}

func TestRenderRealm_UnderSizeCapRendersNormally(t *testing.T) {
	r := newTestRenderer()
	var buf bytes.Buffer
	u := &weburl.GnoURL{Path: "/r/mock/path"}
	_, err := r.RenderRealm(&buf, u, []byte("# Title\n\nbody\n"), RealmRenderContext{})
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "<h1")
	assert.NotContains(t, buf.String(), "gno-render-plaintext")
}

// 4,000 empty forms are 132 KB of markdown and about 12 MB of HTML, since each
// form renders its own command block.
func TestRenderRealm_OverOutputCapServesPlainText(t *testing.T) {
	var logs bytes.Buffer
	r := NewHTMLRenderer(slog.New(slog.NewTextHandler(&logs, nil)), NewDefaultRenderConfig(), nil)
	src := []byte("## Forms\n\n" + strings.Repeat("<gno-form exec=\"F\">\n</gno-form>\n\n", 4000))
	require.Less(t, len(src), maxMarkdownRenderBytes)

	var buf bytes.Buffer
	_, err := r.RenderRealm(&buf, &weburl.GnoURL{Path: "/r/mock/path"}, src, RealmRenderContext{})
	require.NoError(t, err)
	out := buf.String()
	assert.Less(t, len(out), 2*len(src))
	assert.NotContains(t, out, "<form")
	assert.Contains(t, out, "gno-render-plaintext")
	assert.Contains(t, logs.String(), "path=/r/mock/path len=132010")
}

// 100 headings of 10 KB of quotes render 6 MB of article and 5 MB of table of
// contents, so only the two together pass the cap.
func TestRenderRealm_TocSharesOutputCap(t *testing.T) {
	r := newTestRenderer()
	var buf bytes.Buffer
	meta, err := r.RenderRealm(&buf, &weburl.GnoURL{Path: "/r/mock/path"}, []byte("## A\n"), RealmRenderContext{})
	require.NoError(t, err)
	assert.Len(t, meta.Toc.Items, 1)
	assert.NotContains(t, buf.String(), "b-toc", "the table of contents is charged, never written into the article")

	buf.Reset()
	src := []byte("# Title\n\nA paragraph long enough to be the page's description.\n\n" + strings.Repeat("## "+strings.Repeat(`"`, 10_000)+"\n\n", 100))
	meta, err = r.RenderRealm(&buf, &weburl.GnoURL{Path: "/r/mock/path"}, src, RealmRenderContext{})
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "gno-render-plaintext")
	assert.Empty(t, meta.Toc.Items, "the table of contents links headings the page no longer shows")
	assert.Empty(t, meta.Title, "the page's title comes from a document it shows as plain text")
	assert.Empty(t, meta.Description, "the page's description comes from a document it shows as plain text")
}

// 40,000 code blocks are 440 KB of doc comment and about 11 MB of HTML, since
// each block renders inside its own disclosure.
func TestRenderDocumentation_OverOutputCapServesPlainText(t *testing.T) {
	r := newTestRenderer()
	src := []byte(strings.Repeat("```\nx\n```\n\n", 40000))
	require.Less(t, len(src), maxMarkdownRenderBytes)

	var buf bytes.Buffer
	require.NoError(t, r.RenderDocumentation(&buf, src))
	out := buf.String()
	assert.Less(t, len(out), 2*len(src))
	assert.NotContains(t, out, `<details class="b-doc-example">`)
	assert.Contains(t, out, "gno-render-plaintext")
}

// 50,000 blank lines are 50 KB of source and about 9.6 MB of HTML, since each
// line renders its own numbered span.
func TestRenderSource_OverOutputCapServesPlainText(t *testing.T) {
	r := newTestRenderer()
	var buf bytes.Buffer
	require.NoError(t, r.RenderSource(&buf, "/r/mock/path/a.gno", bytes.Repeat([]byte("\n"), 50_000)))
	assert.Less(t, buf.Len(), 100_000)
	assert.Contains(t, buf.String(), "gno-render-plaintext")
	assert.Contains(t, buf.String(), "too large to render")
}

func TestRenderCapped(t *testing.T) {
	write := func(n int) func(*cappedBuffer) error {
		return func(w *cappedBuffer) error {
			_, err := w.Write(make([]byte, n))
			return err
		}
	}

	var out bytes.Buffer
	fellBack, err := renderCapped(&out, 16, write(16))
	require.NoError(t, err)
	assert.False(t, fellBack)
	assert.Equal(t, 16, out.Len())

	out.Reset()
	fellBack, err = renderCapped(&out, 16, write(17))
	require.NoError(t, err)
	assert.True(t, fellBack)
	assert.Zero(t, out.Len(), "nothing of a render past the cap reaches w")

	// bufio hands a string longer than its buffer to the writer's WriteString,
	// past Write, whenever the writer has one.
	fellBack, err = renderCapped(&out, 8000, func(w *cappedBuffer) error {
		bw := bufio.NewWriter(w)
		for range 3 {
			_, _ = bw.WriteString(strings.Repeat("x", 5000))
		}
		return bw.Flush()
	})
	require.NoError(t, err)
	assert.True(t, fellBack)

	assert.PanicsWithValue(t, "boom", func() {
		_, _ = renderCapped(&out, 16, func(*cappedBuffer) error { panic("boom") })
	})
	assert.Panics(t, func() {
		_, _ = renderCapped(&out, 16, func(*cappedBuffer) error {
			var m map[string]int
			m["k"] = 1 // a runtime.Error: an error, and still not the cap
			return nil
		})
	})
}

// A budget of three and a half docs renders one; a triple doc passes what is
// left and spends it, so the next small one shows as plain text too, under no
// second notice.
func TestPageDocRenderer_SharesOneBudget(t *testing.T) {
	var logs bytes.Buffer
	r := NewHTMLRenderer(slog.New(slog.NewTextHandler(&logs, nil)), NewDefaultRenderConfig(), nil)
	small := []byte("```\nx\n```\n\n```\ny\n```\n")
	var one bytes.Buffer
	require.NoError(t, r.RenderDocumentation(&one, small))

	budget := 3*one.Len() + one.Len()/2
	docs := &pageDocRenderer{HTMLRenderer: r, path: "/r/mock/path", left: budget}
	var buf bytes.Buffer
	require.NoError(t, docs.RenderDocumentation(&buf, small))
	assert.Equal(t, budget-one.Len(), docs.left)

	buf.Reset()
	require.NoError(t, docs.RenderDocumentation(&buf, bytes.Repeat(small, 3)))
	assert.Contains(t, buf.String(), "gno-render-plaintext")
	assert.Contains(t, buf.String(), "too large to render")

	buf.Reset()
	require.NoError(t, docs.RenderDocumentation(&buf, small))
	assert.Contains(t, buf.String(), "gno-render-plaintext")
	assert.NotContains(t, buf.String(), "too large to render")
	assert.Equal(t, 1, strings.Count(logs.String(), "path=/r/mock/path"))
}

// A part fits only with what it adds beside it, and is charged both, less what
// it embeds that the page already paid for.
func TestRenderPart_ChargesBeside(t *testing.T) {
	part := func(w io.Writer) error { _, err := w.Write(make([]byte, 100)); return err }

	docs := newDocPage(newTestRenderer(), "", maxRenderOutputBytes-150)
	fits, err := renderPart(docs, io.Discard, 0, 60, part)
	require.NoError(t, err)
	assert.False(t, fits)
	assert.Equal(t, 150, docs.left)

	fits, err = renderPart(docs, io.Discard, 10, 60, part)
	require.NoError(t, err)
	assert.True(t, fits)
	assert.Equal(t, 0, docs.left)
}

// A page whose budget a README used up logs its first doc comment's fallback,
// and only that one.
func TestPageDocRenderer_LogsBudgetUsedBeforeFirstDoc(t *testing.T) {
	var logs bytes.Buffer
	r := NewHTMLRenderer(slog.New(slog.NewTextHandler(&logs, nil)), NewDefaultRenderConfig(), nil)
	docs := newDocPage(r, "/r/demo/foo", maxRenderOutputBytes)
	var buf bytes.Buffer
	require.NoError(t, docs.RenderDocumentation(&buf, []byte("hello *world*")))
	require.NoError(t, docs.RenderDocumentation(&buf, []byte("hello *world*")))
	assert.Equal(t, 2, strings.Count(buf.String(), "gno-render-plaintext"))
	assert.Equal(t, 1, strings.Count(logs.String(), "path=/r/demo/foo "))
}

// A signature is charged to the page like a doc comment.
func TestPageDocRenderer_ChargesSignatures(t *testing.T) {
	docs := newDocPage(newTestRenderer(), "", 0)
	var buf bytes.Buffer
	require.NoError(t, docs.RenderSource(&buf, "sig.gno", []byte("func F()")))
	assert.Equal(t, maxRenderOutputBytes-buf.Len(), docs.left)
}

// A doc comment past the input cap is charged its plain text, not the rest of
// the page's budget.
func TestPageDocRenderer_OverSizeInputKeepsBudget(t *testing.T) {
	docs := newDocPage(newTestRenderer(), "", 0)
	var buf bytes.Buffer
	require.NoError(t, docs.RenderDocumentation(&buf, bytes.Repeat([]byte("a"), maxMarkdownRenderBytes+1)))
	assert.Equal(t, maxRenderOutputBytes-buf.Len(), docs.left)
	buf.Reset()
	require.NoError(t, docs.RenderDocumentation(&buf, []byte("hello *world*")))
	assert.Contains(t, buf.String(), "<em>world</em>")
}

func TestRenderDocumentation_OverSizeCapServesEscapedPlainText(t *testing.T) {
	r := newTestRenderer()
	src := []byte(strings.Repeat("# Heading <script>\n\n", (maxMarkdownRenderBytes/20)+100))
	require.Greater(t, len(src), maxMarkdownRenderBytes)

	var buf bytes.Buffer
	require.NoError(t, r.RenderDocumentation(&buf, src))
	out := buf.String()
	assert.NotContains(t, out, "<h1", "goldmark must not run on oversize input")
	assert.NotContains(t, out, "<script>", "raw content must be escaped")
	assert.Contains(t, out, "&lt;script&gt;", "content is shown as escaped plain text")
	assert.Contains(t, out, "gno-alert-warning", "the notice reuses the markdown warning-alert styling")
	assert.Contains(t, out, "too large to render", "a notice explains why the page is unstyled")
}

func TestRenderer_RenderSource_Gno(t *testing.T) {
	r := newTestRenderer()
	w := &bytes.Buffer{}
	src := []byte("package main\nfunc main() {}\n")
	err := r.RenderSource(w, "foo.gno", src)
	require.NoError(t, err)
	assert.Contains(t, w.String(), "chroma-") // chroma CSS classes
}

func TestRenderer_RenderSource_Markdown(t *testing.T) {
	r := newTestRenderer()
	w := &bytes.Buffer{}
	src := []byte("# Title\nSome Text\n")
	err := r.RenderSource(w, "foo.md", src)
	require.NoError(t, err)
	assert.Contains(t, w.String(), "chroma-")
	assert.Contains(t, w.String(), "Some Text")
}

func TestRenderer_RenderSource_Mod(t *testing.T) {
	r := newTestRenderer()
	w := &bytes.Buffer{}
	src := []byte("module foo\n")
	err := r.RenderSource(w, "foo.mod", src)
	require.NoError(t, err)
	assert.Contains(t, w.String(), "chroma-")
}

func TestRenderer_RenderSource_Toml(t *testing.T) {
	r := newTestRenderer()
	w := &bytes.Buffer{}
	src := []byte("[section]\nkey = 'value'\n")
	err := r.RenderSource(w, "foo.toml", src)
	require.NoError(t, err)
	assert.Contains(t, w.String(), "chroma-")
}

func TestRenderer_RenderSource_Unknown(t *testing.T) {
	r := newTestRenderer()
	w := &bytes.Buffer{}
	src := []byte("plain text\n")
	err := r.RenderSource(w, "foo.txt", src)
	require.NoError(t, err)
	assert.Contains(t, w.String(), "chroma-")
}

func TestRenderer_RenderSource_UnsupportedLexer(t *testing.T) {
	r := newTestRenderer()
	w := &bytes.Buffer{}
	src := []byte("plain text\n")
	err := r.RenderSource(w, "foo.unknownext", src)
	// Should not error, should use txt lexer
	require.NoError(t, err)
	assert.Contains(t, w.String(), "chroma-")
}

func TestRenderer_WriteChromaCSS(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name          string
		darkStyle     bool
		wantLightCSS  bool
		wantDarkScope bool
	}{
		{"success: default config outputs light and dark CSS", true, true, true},
		{"success: nil dark style outputs light CSS only", false, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := NewDefaultRenderConfig()
			if !tc.darkStyle {
				cfg.ChromaDarkStyle = nil
			}
			r := NewHTMLRenderer(slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)), cfg, nil)

			w := &bytes.Buffer{}
			err := r.WriteChromaCSS(w)
			require.NoError(t, err)

			css := w.String()
			assert.True(t, tc.wantLightCSS, "light CSS should always be present")
			assert.Contains(t, css, ".chroma-")

			if tc.wantDarkScope {
				_, dark, ok := strings.Cut(css, `[data-theme="dark"]`)
				require.True(t, ok)
				// Dark token rules target the class the HTML carries, and
				// leave the frame backgrounds to gnoweb.
				assert.Contains(t, dark, ".chroma-light .chroma-k ")
				assert.NotContains(t, dark, ".chroma-dark")
				assert.NotContains(t, dark, "PreWrapper")
				assert.NotContains(t, dark, "/* Background */")
			} else {
				assert.NotContains(t, css, `[data-theme="dark"]`)
			}
		})
	}

	t.Run("error: light CSS write failure", func(t *testing.T) {
		t.Parallel()

		r := newTestRenderer()
		err := r.WriteChromaCSS(&errWriter{err: errors.New("boom")})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "writing light chroma CSS")
	})

	t.Run("error: dark CSS WriteTo failure", func(t *testing.T) {
		t.Parallel()

		// Measure light CSS size by rendering without dark mode.
		lightCfg := NewDefaultRenderConfig()
		lightCfg.ChromaDarkStyle = nil
		lightR := NewHTMLRenderer(slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)), lightCfg, nil)
		var lightBuf bytes.Buffer
		require.NoError(t, lightR.WriteChromaCSS(&lightBuf))

		// Allow exactly the light CSS bytes, then fail on dark CSS WriteTo.
		r := newTestRenderer()
		err := r.WriteChromaCSS(&limitWriter{limit: lightBuf.Len()})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "writing dark chroma CSS")
	})
}
