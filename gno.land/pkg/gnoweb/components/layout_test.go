package components

import (
	"bytes"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
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
		Analytics: AnalyticsData{
			Enabled:    true,
			AssetsPath: "/assets",
		},
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
	assert.Len(t, enrichedData.Links.Dev, 4, "expected dev links with State and Actions for home mode")
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
			name: "State active when state present",
			query: url.Values{
				"state": []string{""},
			},
			label:    "State",
			expected: true,
		},
		{
			name: "Content inactive when state present",
			query: url.Values{
				"state": []string{""},
			},
			label:    "Content",
			expected: false,
		},
		{
			name:     "Unknown label returns false",
			query:    url.Values{},
			label:    "Unknown",
			expected: false,
		},
	}

	for _, tc := range cases {
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
	assert.Len(t, links, 4, "expected Content, State, Source, and Actions links")
	assert.Equal(t, "Content", links[0].Label)
	assert.Equal(t, "State", links[1].Label)
	assert.Equal(t, "Source", links[2].Label)
	assert.Equal(t, "Actions", links[3].Label)
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

func TestStaticHeaderDevLinks_SourceKeepsOpenFile(t *testing.T) {
	t.Parallel()

	u := weburl.GnoURL{
		Path:     "/r/test/pkg",
		WebQuery: url.Values{"source": {""}, "file": {"admin.gno"}},
	}

	links := StaticHeaderDevLinks(u, ViewModeRealm, false)
	source := links[2]
	require.Equal(t, "Source", source.Label)
	assert.Contains(t, source.URL, "file=admin.gno", "the Source tab must not drop the open file")
	assert.True(t, source.IsActive)

	// With no file open, Source still points at the package overview.
	u.WebQuery = url.Values{}
	links = StaticHeaderDevLinks(u, ViewModeRealm, false)
	assert.NotContains(t, links[2].URL, "file=")
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
	assert.Len(t, enriched.Links.Dev, 4, "expected Content, State, Source, and Actions links")
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

// The chip and data-network must survive the full render, and the off-mainnet
// escalation must reach the class attribute. A Go-only test would still pass
// with a dead CSS selector, so this pins the markup side only.
func TestIndexLayout_NetworkPropagation(t *testing.T) {
	t.Parallel()

	render := func(kind NetworkKind, chainID string) string {
		var buf bytes.Buffer
		err := IndexLayout(IndexData{
			HeadData:    HeadData{ChainId: chainID},
			HeaderData:  HeaderData{ChainId: chainID},
			BodyView:    NewTemplateView(StatusViewType, "status", StatusData{}),
			NetworkKind: kind,
		}).Render(&buf)
		require.NoError(t, err)
		return buf.String()
	}

	testnet := render(NetworkTestnet, "pearl-1")
	assert.Contains(t, testnet, `data-network="testnet"`)
	assert.Contains(t, testnet, `class="network-chip"`)
	assert.Contains(t, testnet, "pearl-1")

	local := render(NetworkLocal, "dev")
	assert.Contains(t, local, `class="network-chip"`)

	// Mainnet keeps the header it had before the chip existed.
	mainnet := render(NetworkMainnet, "gnoland-1")
	assert.Contains(t, mainnet, `data-network="mainnet"`)
	assert.NotContains(t, mainnet, "network-chip")
}

func TestNewRealmNotice(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, text, short string
		wantErr           bool
	}{
		{name: "empty", text: "", wantErr: true},
		{name: "spaces", text: "   ", wantErr: true},
		{name: "heading only", text: "# heading only", wantErr: true},
		{name: "first line empty", text: "\nsecond line", wantErr: true},
		{name: "raw html only", text: "<b></b>", wantErr: true},
		{name: "zero-width space", text: "\u200b", wantErr: true},
		{name: "zero-width entity", text: "&#8203;", wantErr: true},
		{name: "nbsp entity", text: "&nbsp;", wantErr: true},
		{name: "braille blank", text: "\u2800", wantErr: true},
		{name: "empty link", text: "[](https://example.com)", wantErr: true},
		{name: "image only", text: "![warning](https://example.com/w.png)", wantErr: true},
		{name: "text", text: "Community realm"},
		{name: "text and short", text: "Community realm, long", short: "Community realm"},
		{name: "short renders nothing", text: "Community realm", short: "<b></b>", wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			n, err := NewRealmNotice(tc.text, tc.short)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.True(t, n.Enabled())
			assert.Equal(t, tc.short != "", n.Short.Enabled())
		})
	}
}

func TestNewRealmNotice_DropsImages(t *testing.T) {
	t.Parallel()

	n, err := NewRealmNotice("Read ![logo](https://example.com/l.png) the code", "")
	require.NoError(t, err)
	var buf strings.Builder
	require.NoError(t, n.Text.Render(&buf))
	assert.Equal(t, "Read  the code", buf.String())
}

func TestRealmNotice_Lines(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 0, RealmNotice{}.Lines())
	short, err := NewRealmNotice("Community realm, long", "Community realm")
	require.NoError(t, err)
	assert.Equal(t, 1, short.Lines())
	custom, err := NewRealmNotice("Operator text", "")
	require.NoError(t, err)
	assert.Equal(t, 2, custom.Lines())
}

// realmNoticeLayout renders a realm page with the given notice and banner.
func realmNoticeLayout(t *testing.T, notice RealmNotice, banner string) string {
	t.Helper()

	bannerData, err := NewBannerData(banner, "")
	require.NoError(t, err)
	data := IndexData{
		HeadData:   HeadData{Title: "Test"},
		HeaderData: HeaderData{Notice: notice},
		Mode:       ViewModeRealm,
		Banner:     bannerData,
		BodyView: &View{
			Type:      "test-view",
			Component: NewReaderComponent(strings.NewReader("testdata")),
		},
	}

	var buf strings.Builder
	require.NoError(t, IndexLayout(data).Render(&buf))
	return buf.String()
}

// noticeRow returns the realm-notice row's markup, or "" if there is none.
func noticeRow(out string) string {
	start := strings.Index(out, `<div class="b-header-notice"`)
	if start < 0 {
		return ""
	}
	end := strings.Index(out[start:], "</div>")
	return out[start : start+end]
}

func TestIndexLayout_RealmNotice(t *testing.T) {
	t.Parallel()

	t.Run("no notice leaves the page unchanged", func(t *testing.T) {
		t.Parallel()
		out := realmNoticeLayout(t, RealmNotice{}, "")
		assert.Empty(t, noticeRow(out))
		assert.Contains(t, out, "</nav>\n</header>")
		assert.Contains(t, out, `<html lang="en">`)
		assert.NotContains(t, out, "aria-describedby")
	})

	t.Run("default notice is the header's one-line second row", func(t *testing.T) {
		t.Parallel()
		notice, err := NewRealmNotice("**Community realm**, deployed by its author.", "**Community realm.**")
		require.NoError(t, err)
		out := realmNoticeLayout(t, notice, "Maintenance")

		header := strings.Index(out, `<header class="b-header">`)
		require.NotEqual(t, -1, header)
		assert.Less(t, strings.Index(out, `class="b-banner"`), header, "banner must render above the header")
		assert.Less(t, strings.Index(out, "</nav>"), strings.Index(out, `<div class="b-header-notice"`), "row follows the nav")
		assert.Less(t, strings.Index(out, `<div class="b-header-notice"`), strings.Index(out, "</header>\n<main"), "row is inside the header")

		row := noticeRow(out)
		assert.Contains(t, row, `role="note"`)
		assert.Contains(t, row, `aria-label="Community realm notice"`)
		assert.Contains(t, row, `<p id="realm-notice"`)
		assert.Contains(t, row, `aria-hidden="true"`)
		assert.Contains(t, row, `<use href="#ico-info-circle"></use>`)
		assert.Contains(t, row, `<span class="short"><strong>Community realm.</strong></span>`)
		assert.Contains(t, row, `<span class="long"><strong>Community realm</strong>, deployed by its author.</span>`)
		assert.Contains(t, out, `<html lang="en" data-realm-notice-lines="1">`)
		assert.Contains(t, out, `aria-describedby="realm-notice"`)
	})

	t.Run("operator text shows as-is and reserves two lines", func(t *testing.T) {
		t.Parallel()
		notice, err := NewRealmNotice("Operator <script>alert(1)</script> & co", "")
		require.NoError(t, err)
		out := realmNoticeLayout(t, notice, "")

		row := noticeRow(out)
		assert.Contains(t, row, "<span>Operator <!-- raw HTML omitted -->alert(1)<!-- raw HTML omitted --> &amp; co</span>")
		assert.NotContains(t, row, "<script>")
		assert.NotContains(t, row, `class="short"`)
		assert.Contains(t, out, `<html lang="en" data-realm-notice-lines="2">`)
	})
}

// headFixture renders the index layout head with the given build version.
func headFixture(t *testing.T, version string) string {
	t.Helper()

	data := IndexData{
		HeadData: HeadData{
			Title:         "Test",
			AssetsPath:    "/public/",
			ChromaPath:    "/public/_chroma/style.css",
			AssetsVersion: version,
		},
		Mode: ViewModeHome,
		BodyView: &View{
			Type:      "test-view",
			Component: NewReaderComponent(strings.NewReader("testdata")),
		},
	}

	var buf strings.Builder
	require.NoError(t, IndexLayout(data).Render(&buf))
	return buf.String()
}

// Assets the head requests on its own carry the version: an edge cache keyed on
// the URL would otherwise serve a stale favicon or chroma stylesheet across
// releases for as long as its TTL allows.
func TestIndexLayout_AssetVersioning(t *testing.T) {
	output := headFixture(t, "20260920120000")

	for _, href := range []string{
		`href="/public/favicon.ico?v=20260920120000"`,
		`href="/public/_chroma/style.css?v=20260920120000"`,
		`href="/public/main.css?v=20260920120000"`,
	} {
		assert.Contains(t, output, href)
	}
	assert.NotContains(t, output, "/public//", "an asset URL must not carry a doubled slash")
}

// A preload is claimed only by a request for the very same URL, and the request
// for a font is issued by the @font-face rule in the built stylesheet. The
// preload href therefore has to be spelled exactly as the stylesheet spells it:
// append a version to one side only and the preload is never claimed, so the
// font is fetched twice on every cold load. Versioning a font means changing the
// URL the stylesheet emits, which is a build concern rather than a template one.
func TestIndexLayout_FontPreloadsMatchStylesheet(t *testing.T) {
	css, err := os.ReadFile(filepath.Join("..", "public", "main.css"))
	require.NoError(t, err, "the built stylesheet is the source of truth for font URLs")

	matches := regexp.MustCompile(`url\(["']?([^)"']+\.woff2)["']?\)`).FindAllSubmatch(css, -1)
	require.NotEmpty(t, matches, "no woff2 @font-face URL found in the built stylesheet")

	output := headFixture(t, "20260920120000")

	var matched int
	for _, m := range matches {
		// main.css is served from AssetsPath, so a relative url() in it resolves there.
		stylesheetURL := "/public/" + strings.TrimPrefix(string(m[1]), "./")
		if !strings.Contains(output, `href="`+stylesheetURL) {
			continue // the head does not preload this font
		}
		assert.Contains(t, output, `href="`+stylesheetURL+`"`,
			"preload must match the stylesheet URL exactly, with nothing appended")
		matched++
	}
	require.NotZero(t, matched, "expected the head to preload at least one stylesheet font")
}
