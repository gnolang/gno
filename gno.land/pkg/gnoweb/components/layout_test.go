package components

import (
	"net/url"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/stretchr/testify/assert"
)

func TestIndexLayout(t *testing.T) {
	tests := []struct {
		name     string
		mode     ViewMode
		viewType ViewType
	}{
		{
			name:     "Home mode",
			mode:     ViewModeHome,
			viewType: "test-view",
		},
		{
			name:     "Realm mode",
			mode:     ViewModeRealm,
			viewType: "test-view",
		},
		{
			name:     "Package mode",
			mode:     ViewModePackage,
			viewType: "test-view",
		},
		{
			name:     "Explorer mode",
			mode:     ViewModeExplorer,
			viewType: "test-view",
		},
		{
			name:     "User mode",
			mode:     ViewModeUser,
			viewType: "test-view",
		},
		{
			name:     "Directory view",
			mode:     ViewModePackage,
			viewType: DirectoryViewType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := IndexData{
				HeadData: HeadData{
					Title: "Test Title",
				},
				Mode: tt.mode,
				BodyView: &View{
					Type:      tt.viewType,
					Component: NewReaderComponent(strings.NewReader("testdata")),
				},
			}

			component := IndexLayout(data)
			assert.NotNil(t, component, "expected component to be non-nil")

			templateComponent, ok := component.(*TemplateComponent)
			assert.True(t, ok, "expected TemplateComponent type in component")

			_, ok = templateComponent.data.(indexLayoutParams)
			assert.True(t, ok, "expected indexLayoutParams type in component data")
		})
	}
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

	enrichedData := EnrichHeaderData(data, ViewModeHome)

	assert.NotEmpty(t, enrichedData.Links.General, "expected general links to be populated")
	assert.Len(t, enrichedData.Links.Dev, 3, "expected dev links with Actions for home mode")
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
	links := StaticHeaderDevLinks(u, ViewModeRealm)
	assert.Len(t, links, 3, "expected Content, Source, and Actions links")
	assert.Equal(t, "Content", links[0].Label)
	assert.Equal(t, "Source", links[1].Label)
	assert.Equal(t, "Actions", links[2].Label)
}

func TestStaticHeaderDevLinks_WithPackageMode(t *testing.T) {
	t.Parallel()

	u := weburl.GnoURL{
		Path: "/r/test/pkg",
	}

	// Test package mode
	links := StaticHeaderDevLinks(u, ViewModePackage)
	assert.Len(t, links, 2, "expected Content and Source links only")
	assert.Equal(t, "Content", links[0].Label)
	assert.Equal(t, "Source", links[1].Label)
}

func TestStaticHeaderDevLinks_WithExplorerMode(t *testing.T) {
	t.Parallel()

	u := weburl.GnoURL{
		Path: "/r/test/pkg",
	}

	// Test explorer mode
	links := StaticHeaderDevLinks(u, ViewModeExplorer)
	assert.Empty(t, links, "expected no links in explorer mode")
}

func TestEnrichHeaderData_WithRealmMode(t *testing.T) {
	t.Parallel()

	data := HeaderData{
		RealmURL: weburl.GnoURL{
			Path: "/r/test/pkg",
		},
	}

	// Test realm mode
	enriched := EnrichHeaderData(data, ViewModeRealm)
	assert.Equal(t, "/r/test/pkg", enriched.RealmPath)
	assert.Empty(t, enriched.Links.General)
	assert.Len(t, enriched.Links.Dev, 3, "expected Content, Source, and Actions links")
}

func TestEnrichHeaderData_WithExplorerMode(t *testing.T) {
	t.Parallel()

	data := HeaderData{
		RealmURL: weburl.GnoURL{
			Path: "/r/test/pkg",
		},
	}

	// Test explorer mode
	enriched := EnrichHeaderData(data, ViewModeExplorer)
	assert.Equal(t, "/r/test/pkg", enriched.RealmPath)
	assert.Empty(t, enriched.Links.General)
	assert.Empty(t, enriched.Links.Dev, "expected no dev links in explorer mode")
}

func TestViewModePredicates(t *testing.T) {
	cases := []struct {
		mode         ViewMode
		name         string
		wantExplorer bool
		wantRealm    bool
		wantPackage  bool
		wantHome     bool
		wantUser     bool
	}{
		{
			mode:         ViewModeExplorer,
			name:         "Explorer",
			wantExplorer: true,
		},
		{
			mode:      ViewModeRealm,
			name:      "Realm",
			wantRealm: true,
		},
		{
			mode:        ViewModePackage,
			name:        "Package",
			wantPackage: true,
		},
		{
			mode:     ViewModeHome,
			name:     "Home",
			wantHome: true,
		},
		{
			mode:     ViewModeUser,
			name:     "User",
			wantUser: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.wantExplorer, tc.mode.IsExplorer(), "IsExplorer")
			assert.Equal(t, tc.wantRealm, tc.mode.IsRealm(), "IsRealm")
			assert.Equal(t, tc.wantPackage, tc.mode.IsPackage(), "IsPackage")
			assert.Equal(t, tc.wantHome, tc.mode.IsHome(), "IsHome")
			assert.Equal(t, tc.wantUser, tc.mode.IsUser(), "IsUser")
		})
	}
}
