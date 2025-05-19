package components

import (
	"io"
	"net/url"
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

	enrichedData := EnrichHeaderData(data, HeaderModeTemplateHome)

	assert.NotEmpty(t, enrichedData.Links.General, "expected general links to be populated")
	assert.NotEmpty(t, enrichedData.Links.Dev, "expected dev links to be populated")
}

func TestIsActive(t *testing.T) {
	cases := []struct {
		name     string
		query    url.Values
		label    string
		expected bool
	}{
		{
			name:     "Content active when no source or help",
			query:    url.Values{},
			label:    "Content",
			expected: true,
		},
		{
			name: "Content inactive when source present",
			query: url.Values{
				"source": []string{""},
			},
			label:    "Content",
			expected: false,
		},
		{
			name: "Content inactive when help present",
			query: url.Values{
				"help": []string{""},
			},
			label:    "Content",
			expected: false,
		},
		{
			name: "Source active when source present",
			query: url.Values{
				"source": []string{""},
			},
			label:    "Source",
			expected: true,
		},
		{
			name: "Actions active when help present",
			query: url.Values{
				"help": []string{""},
			},
			label:    "Actions",
			expected: true,
		},
		{
			name:     "Unknown label returns false",
			query:    url.Values{},
			label:    "Unknown",
			expected: false,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := isActive(tc.query, tc.label)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestStaticHeaderDevLinks_WithRealmMode(t *testing.T) {
	t.Parallel()

	u := weburl.GnoURL{
		Path: "/r/test/pkg",
	}

	// Test realm mode (default case)
	links := StaticHeaderDevLinks(u, HeaderModeTemplateRealm)
	assert.Len(t, links, 3)
	assert.Equal(t, "Content", links[0].Label)
	assert.Equal(t, "Source", links[1].Label)
	assert.Equal(t, "Actions", links[2].Label)
}

func TestEnrichHeaderData_WithRealmMode(t *testing.T) {
	t.Parallel()

	data := HeaderData{
		RealmURL: weburl.GnoURL{
			Path: "/r/test/pkg",
		},
	}

	// Test realm mode
	enriched := EnrichHeaderData(data, HeaderModeTemplateRealm)
	assert.Equal(t, "/r/test/pkg", enriched.RealmPath)
	assert.Empty(t, enriched.Links.General)
	assert.Len(t, enriched.Links.Dev, 3)
}

func TestEnrichHeaderData_WithExplorerMode(t *testing.T) {
	t.Parallel()

	data := HeaderData{
		RealmURL: weburl.GnoURL{
			Path: "/r/test/pkg",
		},
	}

	// Test explorer mode
	enriched := EnrichHeaderData(data, HeaderModeTemplateExplorer)
	assert.Equal(t, "/r/test/pkg", enriched.RealmPath)
	assert.Empty(t, enriched.Links.General)
	assert.Empty(t, enriched.Links.Dev)
}

func TestHeaderModeTemplatePredicates(t *testing.T) {
	cases := []struct {
		mode         HeaderModeTemplate
		name         string
		wantExplorer bool
		wantRealm    bool
		wantPackage  bool
		wantHome     bool
	}{
		{
			mode:         HeaderModeTemplateExplorer,
			name:         "Explorer",
			wantExplorer: true,
		},
		{
			mode:      HeaderModeTemplateRealm,
			name:      "Realm",
			wantRealm: true,
		},
		{
			mode:        HeaderModeTemplatePackage,
			name:        "Package",
			wantPackage: true,
		},
		{
			mode:     HeaderModeTemplateHome,
			name:     "Home",
			wantHome: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.wantExplorer, tc.mode.IsTemplateExplorer(), "IsTemplateExplorer")
			assert.Equal(t, tc.wantRealm, tc.mode.IsTemplateRealm(), "IsTemplateRealm")
			assert.Equal(t, tc.wantPackage, tc.mode.IsTemplateModePackage(), "IsTemplateModePackage")
			assert.Equal(t, tc.wantHome, tc.mode.IsTemplateHome(), "IsTemplateHome")
		})
	}
}
