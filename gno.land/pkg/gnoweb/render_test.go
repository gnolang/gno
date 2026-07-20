package gnoweb

import (
	bytes "bytes"
	"errors"
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
	require.Contains(t, out, `<details class="doc-example">`)
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
	assert.Empty(t, toc.Items)
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
				assert.Contains(t, css, `[data-theme="dark"]`)
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
