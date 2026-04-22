package components

import (
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
