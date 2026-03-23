package gnoweb

import (
	bytes "bytes"
	"errors"
	"log/slog"
	"testing"

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
		tc := tc
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
