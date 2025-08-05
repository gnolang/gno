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
	return NewHTMLRenderer(slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)), NewDefaultRenderConfig())
}

func TestRenderer_RenderRealm_Markdown(t *testing.T) {
	r := newTestRenderer()
	w := &bytes.Buffer{}
	u := &weburl.GnoURL{Path: "/r/test"}
	src := []byte(`# Hello\n\nThis is a **test**.`)
	toc, err := r.RenderRealm(w, u, src)
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

func TestRenderer_RenderDocumentation(t *testing.T) {
	r := newTestRenderer()
	w := &bytes.Buffer{}
	u := &weburl.GnoURL{Path: "/r/test"}
	src := []byte(`# Documentation

This is documentation with **markdown** support.

## Code Example

    func example() {
        return "hello"
    }

## Another Section

More documentation content.`)

	err := r.RenderDocumentation(w, u, src)
	require.NoError(t, err)

	html := w.String()
	assert.Contains(t, html, "<h1")
	assert.Contains(t, html, "Documentation")
	assert.Contains(t, html, "<strong>markdown</strong>")
	assert.Contains(t, html, "<h2")
	assert.Contains(t, html, "Code Example")
	assert.Contains(t, html, "Another Section")
}

func TestRenderer_RenderDocumentation_WithCodeBlocks(t *testing.T) {
	r := newTestRenderer()
	w := &bytes.Buffer{}
	u := &weburl.GnoURL{Path: "/r/test"}
	src := []byte(`# Test Documentation

Here is a code example:

    func test() {
        return true
    }

And another:

    func another() {
        return false
    }

End of documentation.`)

	err := r.RenderDocumentation(w, u, src)
	require.NoError(t, err)

	html := w.String()
	assert.Contains(t, html, "Test Documentation")
	assert.Contains(t, html, "func")
	assert.Contains(t, html, "test")
	assert.Contains(t, html, "return")
	assert.Contains(t, html, "true")
	assert.Contains(t, html, "another")
	assert.Contains(t, html, "false")
	assert.Contains(t, html, "End of documentation")
}

func TestRenderer_RenderDocumentation_EmptyInput(t *testing.T) {
	r := newTestRenderer()
	w := &bytes.Buffer{}
	u := &weburl.GnoURL{Path: "/r/test"}
	src := []byte("")

	err := r.RenderDocumentation(w, u, src)
	require.NoError(t, err)

	// Should not error with empty input
	assert.NotNil(t, w.String())
}

func TestRenderer_RenderDocumentation_ComplexMarkdown(t *testing.T) {
	r := newTestRenderer()
	w := &bytes.Buffer{}
	u := &weburl.GnoURL{Path: "/r/test"}
	src := []byte(`# Complex Documentation

This documentation has multiple features:

## Lists

- Item 1
- Item 2
- Item 3

## Code

    func complex() {
        if true {
            return "complex"
        }
        return "simple"
    }

## Links

Visit [Gno.land](https://gno.land) for more information.

## Emphasis

*Italic text* and **bold text** are supported.`)

	err := r.RenderDocumentation(w, u, src)
	require.NoError(t, err)

	html := w.String()
	assert.Contains(t, html, "Complex Documentation")
	assert.Contains(t, html, "Lists")
	assert.Contains(t, html, "Item 1")
	assert.Contains(t, html, "Item 2")
	assert.Contains(t, html, "Item 3")
	assert.Contains(t, html, "func")
	assert.Contains(t, html, "complex")
	assert.Contains(t, html, "if")
	assert.Contains(t, html, "true")
	assert.Contains(t, html, "return")
	assert.Contains(t, html, "complex")
	assert.Contains(t, html, "Links")
	assert.Contains(t, html, "Gno.land")
	assert.Contains(t, html, "Emphasis")
	assert.Contains(t, html, "Italic text")
	assert.Contains(t, html, "bold text")
}
