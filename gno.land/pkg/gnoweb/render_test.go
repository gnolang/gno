package gnoweb

import (
	bytes "bytes"
	"log/slog"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func TestRenderer_WriteFormatterCSS(t *testing.T) {
	r := newTestRenderer()
	w := &bytes.Buffer{}
	err := r.WriteChromaCSS(w)
	require.NoError(t, err)
	assert.Contains(t, w.String(), ".chroma-")
}
