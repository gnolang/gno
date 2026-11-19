package components

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/gnolang/gno/gnovm/pkg/doc"
	"github.com/stretchr/testify/require"
)

func TestOverviewView_BuildsFullPayload(t *testing.T) {
	t.Parallel()
	u, err := weburl.Parse("/r/demo/foo")
	require.NoError(t, err)

	data := BuildOverview(OverviewInput{
		URL:         u,
		Files:       []string{"foo.gno", "foo_test.gno", "README.md", "LICENSE"},
		Doc:         &doc.JSONDocumentation{PackageDoc: "Package foo.", Funcs: []*doc.JSONFunc{{Name: "Hello"}}},
		Readme:      nil,
		DocRenderer: noopRenderer{},
		Domain:      "gno.land",
	})

	require.Equal(t, "foo", data.Title)
	require.Equal(t, "Package foo.", data.Synopsis)
	require.Equal(t, "realm", data.Info.PackageType)
	require.Equal(t, "", data.Info.License.Kind, "license file present but content not provided → Kind empty")
	require.Equal(t, "LICENSE", data.Info.License.FileName)
	require.Len(t, data.Funcs, 1)

	view := OverviewView(data)
	require.Equal(t, OverviewViewType, view.Type)
}

// TestOverviewView_TypeCardFoldsMethodNamesIntoDataName guards the symbol filter:
// method cards are nested inside their type card, which is what the filter hides,
// so the type card's data-name must also carry its method names to stay matchable.
func TestOverviewView_TypeCardFoldsMethodNamesIntoDataName(t *testing.T) {
	t.Parallel()
	u, err := weburl.Parse("/r/demo/foo")
	require.NoError(t, err)

	data := BuildOverview(OverviewInput{
		URL:   u,
		Files: []string{"foo.gno"},
		Doc: &doc.JSONDocumentation{
			Types: []*doc.JSONType{{Name: "Config", Type: "type Config struct{}", Kind: "struct"}},
			Funcs: []*doc.JSONFunc{{Name: "Load", Signature: "func (c *Config) Load() error", Type: "Config"}},
		},
		DocRenderer: noopRenderer{},
		Domain:      "gno.land",
	})

	var buf bytes.Buffer
	require.NoError(t, OverviewView(data).Render(&buf))
	require.Contains(t, buf.String(), `data-name="Config Load"`,
		"type card data-name must fold in method names so the filter can find methods")
}

// Only the stdlib row leaves the site.
func TestOverviewView_StdlibImportLeavesTheSite(t *testing.T) {
	t.Parallel()
	u, err := weburl.Parse("/r/demo/foo")
	require.NoError(t, err)

	data := BuildOverview(OverviewInput{
		URL:         u,
		Files:       []string{"foo.gno"},
		Doc:         &doc.JSONDocumentation{Imports: []string{"gno.land/p/demo/avl", "strings"}},
		DocRenderer: noopRenderer{},
		Domain:      "gno.land",
	})

	var buf bytes.Buffer
	require.NoError(t, OverviewView(data).Render(&buf))
	out := buf.String()

	require.Contains(t, out, `href="`+stdlibSourceBase+`strings" class="b-pkg-list__link" rel="noopener noreferrer"`)
	require.Contains(t, out, `<a href="/p/demo/avl" class="b-pkg-list__link">`,
		"an on-site import carries no rel and no external icon")
	require.Equal(t, 1, strings.Count(out, `<use href="#ico-external-link">`),
		"only the stdlib row is marked outbound")
}

// countingRenderer counts the docs and signatures it renders.
type countingRenderer struct {
	noopRenderer
	docs, sigs int
}

func (r *countingRenderer) RenderDocumentation(w io.Writer, src []byte) error {
	r.docs++
	return r.noopRenderer.RenderDocumentation(w, src)
}

func (r *countingRenderer) RenderSource(w io.Writer, name string, src []byte) error {
	r.sigs++
	return r.noopRenderer.RenderSource(w, name, src)
}

func TestOverviewView_RendersOnlyShownSymbols(t *testing.T) {
	t.Parallel()
	u, err := weburl.Parse("/r/demo/foo")
	require.NoError(t, err)

	funcs := make([]*doc.JSONFunc, maxOverviewSymbols+100)
	for i := range funcs {
		funcs[i] = &doc.JSONFunc{Name: "Hello", Signature: "func Hello()", Doc: "Hello greets."}
	}
	r := &countingRenderer{}
	data := BuildOverview(OverviewInput{
		URL:         u,
		DocRenderer: r,
		Domain:      "gno.land",
		Doc:         &doc.JSONDocumentation{Funcs: funcs},
	})
	require.NoError(t, OverviewView(data).Render(io.Discard))

	// One signature and one doc per function the page shows.
	require.Equal(t, maxOverviewSymbols, r.sigs)
	require.Equal(t, maxOverviewSymbols, r.docs)
}

// blankRenderer renders every doc to nothing, as a doc of link definitions does.
type blankRenderer struct{ noopRenderer }

func (blankRenderer) RenderDocumentation(io.Writer, []byte) error { return nil }

func TestOverviewView_DocRenderingToNothingIsAbsent(t *testing.T) {
	t.Parallel()
	u, err := weburl.Parse("/r/demo/foo")
	require.NoError(t, err)

	linkDef := "[spec]: https://example.com/spec"
	data := BuildOverview(OverviewInput{
		URL:         u,
		DocRenderer: blankRenderer{},
		Domain:      "gno.land",
		Doc: &doc.JSONDocumentation{
			PackageDoc: linkDef,
			Funcs:      []*doc.JSONFunc{{Name: "Hello", Signature: "func Hello()", Doc: linkDef}},
			Types:      []*doc.JSONType{{Name: "Config", Type: "struct{}", Kind: "struct", Doc: linkDef}},
			Values: []*doc.JSONValueDecl{{
				Signature: "const Max = 1", Const: true, Doc: linkDef,
				Values: []*doc.JSONValue{{Name: "Max"}},
			}},
		},
	})

	var buf bytes.Buffer
	require.NoError(t, OverviewView(data).Render(&buf))
	out := buf.String()
	require.Equal(t, 3, strings.Count(out, `<article class="b-pkg-symbol"`), "one card per symbol")
	require.NotContains(t, out, `id="overview"`, "no section for a package doc rendering to nothing")
	require.NotContains(t, out, "b-pkg-symbol__doc", "no doc block for a symbol doc rendering to nothing")
}

// failingRenderer fails every render.
type failingRenderer struct{}

func (failingRenderer) RenderDocumentation(io.Writer, []byte) error {
	return errors.New("render failed")
}

func (failingRenderer) RenderSource(io.Writer, string, []byte) error {
	return errors.New("render failed")
}

func TestOverviewView_FailedRenderWritesEscapedSource(t *testing.T) {
	t.Parallel()
	u, err := weburl.Parse("/r/demo/foo")
	require.NoError(t, err)

	data := BuildOverview(OverviewInput{
		URL:         u,
		DocRenderer: failingRenderer{},
		Domain:      "gno.land",
		Doc: &doc.JSONDocumentation{
			Funcs: []*doc.JSONFunc{{Name: "Hello", Signature: "func Hello() <b>", Doc: "<i>greets</i>"}},
		},
	})

	var buf bytes.Buffer
	require.NoError(t, OverviewView(data).Render(&buf))
	out := buf.String()
	require.Contains(t, out, "func Hello() &lt;b&gt;")
	require.Contains(t, out, "&lt;i&gt;greets&lt;/i&gt;")
}
