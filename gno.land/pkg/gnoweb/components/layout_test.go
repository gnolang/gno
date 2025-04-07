package components

import (
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
)

func TestIndexLayout(t *testing.T) {
	data := IndexData{
		HeadData: HeadData{
			Title: "Test Title",
		},
		BodyView: &View{Type: "test-view"},
	}

	component := IndexLayout(data)

	if component == nil {
		t.Error("expected component to be non-nil")
		return
	}

	templateComponent, ok := component.(*TemplateComponent)
	if !ok {
		t.Error("expected TemplateComponent type in component")
		return
	}

	layoutParams, ok := templateComponent.data.(indexLayoutParams)
	if !ok {
		t.Error("expected indexLayoutParams type in component data")
		return
	}

	if layoutParams.Layout != FullLayout {
		t.Errorf("expected layout %s, got %s", FullLayout, layoutParams.Layout)
	}
}

func TestEnrichFooterData(t *testing.T) {
	data := FooterData{
		Analytics:  true,
		AssetsPath: "/assets",
	}

	enrichedData := EnrichFooterData(data)

	if len(enrichedData.Sections) == 0 {
		t.Error("expected sections to be populated")
	}

	expectedSections := []string{"Footer navigation", "Social media", "Legal"}
	for i, section := range enrichedData.Sections {
		if section.Title != expectedSections[i] {
			t.Errorf("expected section title %s, got %s", expectedSections[i], section.Title)
		}
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

	if len(enrichedData.Links.General) == 0 {
		t.Error("expected general links to be populated")
	}

	if len(enrichedData.Links.Dev) == 0 {
		t.Error("expected dev links to be populated")
	}
}
