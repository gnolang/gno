package components

import (
	"net/url"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	expectedSections := []string{"Footer navigation", "Social media"}
	for i, section := range enrichedData.Sections {
		assert.Equal(t, expectedSections[i], section.Title, "expected section title %s, got %s", expectedSections[i], section.Title)
	}

	assert.NotEmpty(t, enrichedData.LegalNotice, "expected legal notice to be populated")
	assert.Contains(t, enrichedData.LegalNotice, "NewTendermint", "expected legal notice to mention NewTendermint")

	assert.Len(t, enrichedData.LegalLinks, 3, "expected 3 legal links")
	expectedLabels := []string{"Gno GPL License", "Gno.land Network Interaction Terms", "Gno.land Contributor License Agreement"}
	for i, link := range enrichedData.LegalLinks {
		assert.Equal(t, expectedLabels[i], link.Label, "expected legal link label %s, got %s", expectedLabels[i], link.Label)
		assert.NotEmpty(t, link.URL, "expected legal link URL to be non-empty")
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
	links := StaticHeaderDevLinks(u, ViewModeRealm, false)
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
	links := StaticHeaderDevLinks(u, ViewModePackage, false)
	assert.Len(t, links, 2, "expected Content and Source links only")
	assert.Equal(t, "Content", links[0].Label)
	assert.Equal(t, "Source", links[1].Label)
}

func TestStaticHeaderDevLinks_StaticContent(t *testing.T) {
	t.Parallel()

	u := weburl.GnoURL{
		Path: "/r/test/pkg",
	}

	links := StaticHeaderDevLinks(u, ViewModeRealm, true)
	require.Len(t, links, 1, "static content should only have Content link")
	assert.Equal(t, "Content", links[0].Label)
}

func TestStaticHeaderDevLinks_WithExplorerMode(t *testing.T) {
	t.Parallel()

	u := weburl.GnoURL{
		Path: "/r/test/pkg",
	}

	// Test explorer mode
	links := StaticHeaderDevLinks(u, ViewModeExplorer, false)
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

func TestIndexLayout_ThemePropagation(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		theme         string
		wantAttr      string
		wantNoDataTag bool
	}{
		{
			name:     "success: dark theme rendered in HTML",
			theme:    "dark",
			wantAttr: `data-theme="dark"`,
		},
		{
			name:     "success: light theme rendered in HTML",
			theme:    "light",
			wantAttr: `data-theme="light"`,
		},
		{
			name:          "edge: empty theme omits data-theme attribute",
			theme:         "",
			wantNoDataTag: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			data := IndexData{
				HeadData: HeadData{
					Title: "Test",
				},
				Mode:  ViewModeHome,
				Theme: tc.theme,
				BodyView: &View{
					Type:      "test-view",
					Component: NewReaderComponent(strings.NewReader("testdata")),
				},
			}

			component := IndexLayout(data)

			var buf strings.Builder
			err := component.Render(&buf)
			require.NoError(t, err, "expected no render error")

			output := buf.String()
			if tc.wantNoDataTag {
				assert.NotContains(t, output, `data-theme=`, "expected no data-theme attribute")
			} else {
				assert.Contains(t, output, tc.wantAttr, "expected HTML to contain %s", tc.wantAttr)
			}
		})
	}
}

func TestNewBannerData(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name            string
		input           string
		globalURL       string
		wantEnabled     bool
		wantHasURL      bool
		wantContains    string
		wantNotContains string
	}{
		{
			name:        "empty is disabled",
			input:       "",
			wantEnabled: false,
		},
		{
			name:         "plain text",
			input:        "Beta",
			wantEnabled:  true,
			wantContains: "Beta",
		},
		{
			name:         "markdown link gets target blank",
			input:        "[Beta](https://example.com)",
			wantEnabled:  true,
			wantContains: `<a href="https://example.com" target="_blank" rel="noopener noreferrer">Beta</a>`,
		},
		{
			name:         "bold and italic",
			input:        "This is **bold** and *italic*",
			wantEnabled:  true,
			wantContains: "<strong>bold</strong>",
		},
		{
			name:         "content after newline discarded",
			input:        "line one\nline two",
			wantEnabled:  true,
			wantContains: "line one",
		},
		{
			name:        "truncated over max length",
			input:       strings.Repeat("a", MaxBannerLength+50),
			wantEnabled: true,
		},
		{
			name:        "HTML block stripped",
			input:       `<script>alert("xss")</script>`,
			wantEnabled: false,
		},
		{
			name:         "javascript URL sanitized",
			input:        `[click](javascript:alert(1))`,
			wantEnabled:  true,
			wantContains: `href=""`,
		},
		{
			name:            "global URL strips inline links",
			input:           "[click](https://other.com)",
			globalURL:       "https://gno.land",
			wantEnabled:     true,
			wantHasURL:      true,
			wantContains:    "click",
			wantNotContains: `href="https://other.com"`,
		},
		{
			name:        "global javascript URL rejected",
			input:       "Hello",
			globalURL:   "javascript:alert(1)",
			wantEnabled: true,
			wantHasURL:  false,
		},
		{
			name:        "global ftp URL rejected",
			input:       "Hello",
			globalURL:   "ftp://bad.com",
			wantEnabled: true,
			wantHasURL:  false,
		},
		{
			name:        "heading block stripped",
			input:       "# Big Heading",
			wantEnabled: false,
		},
		{
			name:        "blockquote stripped",
			input:       "> quoted text",
			wantEnabled: false,
		},
		{
			name:        "thematic break stripped",
			input:       "---",
			wantEnabled: false,
		},
		{
			name:        "list item stripped",
			input:       "- list entry",
			wantEnabled: false,
		},
		{
			name:         "leading whitespace trimmed before parsing",
			input:        "    code line",
			wantEnabled:  true,
			wantContains: "code line",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			banner, err := NewBannerData(tc.input, tc.globalURL)
			require.NoError(t, err)
			assert.Equal(t, tc.wantEnabled, banner.Enabled())
			assert.Equal(t, tc.wantHasURL, banner.HasURL())

			var buf strings.Builder
			require.NoError(t, banner.Render(&buf))
			rendered := buf.String()

			if tc.wantContains != "" {
				assert.Contains(t, rendered, tc.wantContains)
			}
			if tc.wantNotContains != "" {
				assert.NotContains(t, rendered, tc.wantNotContains)
			}
			if banner.Enabled() {
				assert.NotContains(t, rendered, "<p>")
			}
		})
	}
}

func TestIndexLayout_Banner(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name            string
		markdown        string
		url             string
		wantBanner      bool
		wantContains    string
		wantNotContains string
	}{
		{
			name:       "no banner when empty",
			markdown:   "",
			wantBanner: false,
		},
		{
			name:         "plain text renders in div",
			markdown:     "Maintenance",
			wantBanner:   true,
			wantContains: "Maintenance",
		},
		{
			name:         "markdown link renders inline",
			markdown:     "[Beta](https://example.com)",
			wantBanner:   true,
			wantContains: `href="https://example.com"`,
		},
		{
			name:         "global URL wraps banner in anchor",
			markdown:     "Beta release",
			url:          "https://gno.land",
			wantBanner:   true,
			wantContains: `<a href="https://gno.land"`,
		},
		{
			name:            "global URL overrides inline links",
			markdown:        "[click here](https://other.com)",
			url:             "https://gno.land",
			wantBanner:      true,
			wantContains:    `<a href="https://gno.land"`,
			wantNotContains: `href="https://other.com"`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			banner, err := NewBannerData(tc.markdown, tc.url)
			require.NoError(t, err)

			data := IndexData{
				HeadData: HeadData{Title: "Test"},
				Mode:     ViewModeHome,
				Banner:   banner,
				BodyView: &View{
					Type:      "test-view",
					Component: NewReaderComponent(strings.NewReader("testdata")),
				},
			}

			var buf strings.Builder
			err = IndexLayout(data).Render(&buf)
			require.NoError(t, err)

			output := buf.String()
			if !tc.wantBanner {
				assert.NotContains(t, output, "b-banner")
			} else {
				assert.Contains(t, output, "b-banner")
				assert.Contains(t, output, tc.wantContains)
				if tc.wantNotContains != "" {
					assert.NotContains(t, output, tc.wantNotContains)
				}
			}
		})
	}
}
