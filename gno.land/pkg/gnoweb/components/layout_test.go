package components

import (
	"io"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/stretchr/testify/assert"
)

func TestIndexLayout(t *testing.T) {
	data := IndexData{
		HeadData: HeadData{
			Title: "Test Title",
		},
		BodyView: &View{
			Type:      "test-view",
			Component: NewReaderComponent(strings.NewReader("testdata")),
		},
	}

	component := IndexLayout(data)

	assert.NotNil(t, component, "expected component to be non-nil")

	templateComponent, ok := component.(*TemplateComponent)
	assert.True(t, ok, "expected TemplateComponent type in component")

	layoutParams, ok := templateComponent.data.(indexLayoutParams)
	assert.True(t, ok, "expected indexLayoutParams type in component data")

	assert.Equal(t, FullLayout, layoutParams.Layout, "expected layout %s, got %s", FullLayout, layoutParams.Layout)

	assert.NoError(t, component.Render(io.Discard))
}

func TestEnrichFooterData(t *testing.T) {
	data := FooterData{
		Analytics:  true,
		AssetsPath: "/assets",
	}

	enrichedData := EnrichFooterData(data)

	assert.NotEmpty(t, enrichedData.Sections, "expected sections to be populated")

	expectedSections := []string{"Footer navigation", "Social media", "Legal"}
	for i, section := range enrichedData.Sections {
		assert.Equal(t, expectedSections[i], section.Title, "expected section title %s, got %s", expectedSections[i], section.Title)
	}
}

func TestEnrichHeaderData(t *testing.T) {
	data := HeaderData{
		RealmURL: weburl.GnoURL{
			WebQuery: map[string][]string{},
		},
		Breadcrumb: BreadcrumbData{
			Parts: []BreadcrumbPart{{Name: "p/demo/grc/grc20"}},
		},
	}

	enrichedData := EnrichHeaderData(data, true)

	assert.NotEmpty(t, enrichedData.Links.General, "expected general links to be populated")
	assert.NotEmpty(t, enrichedData.Links.Dev, "expected dev links to be populated")
}
