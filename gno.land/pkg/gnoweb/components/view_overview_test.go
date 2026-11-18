package components

import (
	"bytes"
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

// The Source row is an outbound link to a URL the deployer chose: it must keep
// its rel, its external icon and its accessible "not verified" note, and be
// absent when the package declares none.
func TestOverviewView_SourceRow(t *testing.T) {
	t.Parallel()
	u, err := weburl.Parse("/r/demo/foo")
	require.NoError(t, err)

	render := func(gnomod string) string {
		data := BuildOverview(OverviewInput{
			URL:         u,
			Files:       []string{"foo.gno", "gnomod.toml"},
			Doc:         &doc.JSONDocumentation{},
			Sources:     map[string][]byte{"gnomod.toml": []byte(gnomod)},
			DocRenderer: noopRenderer{},
			Domain:      "gno.land",
		})
		var buf bytes.Buffer
		require.NoError(t, OverviewView(data).Render(&buf))
		return buf.String()
	}

	out := render("module = \"gno.land/r/demo/foo\"\n\n[source]\nrepository = \"https://github.com/o/r\"\npath = \"r/foo\"\nrevision = \"3cc494ec4d1f\"\n")
	require.Contains(t, out, `<a href="https://github.com/o/r/tree/3cc494ec4d1f/r/foo" rel="noopener nofollow ugc"`)
	require.Contains(t, out, `aria-describedby="pkg-source-note">github.com/o/r@3cc494ec4<svg class="c-icon" aria-hidden="true"><use href="#ico-external-link"></use></svg></a>`)
	require.Contains(t, out, `<span id="pkg-source-note" class="u-sr-only">`)

	for _, mod := range []string{
		"module = \"gno.land/r/demo/foo\"\n",
		"module = \"gno.land/r/demo/foo\"\n\n[source]\nrepository = \"javascript:alert(1)\"\n",
	} {
		out := render(mod)
		require.NotContains(t, out, ">Source</dt>")
		require.NotContains(t, out, "javascript:")
	}
}
