package components

import (
	"bytes"
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
