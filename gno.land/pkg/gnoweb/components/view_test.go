package components

import (
	"io"
	"strings"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/doc"
	"github.com/stretchr/testify/assert"
)

func TestSourceView(t *testing.T) {
	tests := []struct {
		name     string
		data     SourceData
		expected int
	}{
		{
			name: "Basic Test",
			data: SourceData{
				PkgPath:      "example/path",
				Files:        []string{"README.md", "main.gno", "test_file.gno", "example_test.gno"},
				FileName:     "main.gno",
				FileSize:     "1KB",
				FileLines:    100,
				FileCounter:  1,
				FileDownload: "example/path/main.gno",
				FileSource:   NewReaderComponent(strings.NewReader("testdata")),
			},
			expected: 4,
		},
		{
			name: "No Files",
			data: SourceData{
				PkgPath:      "example/path",
				Files:        []string{},
				FileName:     "main.gno",
				FileSize:     "1KB",
				FileLines:    100,
				FileCounter:  1,
				FileDownload: "example/path/main.gno",
				FileSource:   NewReaderComponent(strings.NewReader("testdata")),
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view := SourceView(tt.data)

			assert.NotNil(t, view, "expected view to be non-nil")

			tocItemsCount := len(tt.data.Files)
			assert.Equal(t, tt.expected, tocItemsCount, "expected %d TOC items, got %d", tt.expected, tocItemsCount)
			assert.Equal(t, SourceViewType, view.Type, "expected view type %s, got %s", SourceViewType, view.Type)

			assert.NoError(t, view.Render(io.Discard))
		})
	}
}

func TestStatusErrorComponent(t *testing.T) {
	message := "Test Error"
	view := StatusErrorComponent(message)

	assert.NotNil(t, view, "expected view to be non-nil")

	expectedTitle := "Error: " + message
	templateComponent, ok := view.Component.(*TemplateComponent)
	assert.True(t, ok, "expected TemplateComponent type in view.Component")

	statusData, ok := templateComponent.data.(StatusData)
	assert.True(t, ok, "expected StatusData type in component data")

	assert.Equal(t, expectedTitle, statusData.Title, "expected title %s, got %s", expectedTitle, statusData.Title)

	assert.NoError(t, view.Render(io.Discard))
}

func TestStatusNoRenderComponent(t *testing.T) {
	pkgPath := "example/path"
	view := StatusNoRenderComponent(pkgPath)

	assert.NotNil(t, view, "expected view to be non-nil")

	templateComponent, ok := view.Component.(*TemplateComponent)
	assert.True(t, ok, "expected TemplateComponent type in view.Component")

	statusData, ok := templateComponent.data.(StatusData)
	assert.True(t, ok, "expected StatusData type in component data")

	expectedURL := pkgPath + "$source"
	assert.Equal(t, expectedURL, statusData.ButtonURL, "expected ButtonURL %s, got %s", expectedURL, statusData.ButtonURL)

	assert.NoError(t, view.Render(io.Discard))
}

func TestRedirectView(t *testing.T) {
	data := RedirectData{
		To:            "example/path",
		WithAnalytics: true,
	}
	view := RedirectView(data)

	assert.NotNil(t, view, "expected view to be non-nil")

	templateComponent, ok := view.Component.(*TemplateComponent)
	assert.True(t, ok, "expected TemplateComponent type in view.Component")

	redirectData, ok := templateComponent.data.(RedirectData)
	assert.True(t, ok, "expected RedirectData type in component data")

	assert.Equal(t, data.To, redirectData.To, "expected redirect to %s, got %s", data.To, redirectData.To)
	assert.Equal(t, data.WithAnalytics, redirectData.WithAnalytics, "expected WithAnalytics to be %v, got %v", data.WithAnalytics, redirectData.WithAnalytics)

	assert.NoError(t, view.Render(io.Discard))
}

func TestViewRender(t *testing.T) {
	component := NewTemplateComponent("ui/toc_generic", nil)
	view := &View{
		Type:      "test-view",
		Component: component,
	}

	writer := &mockWriter{}
	err := view.Render(writer)
	assert.NoError(t, err, "expected no error")

	assert.Equal(t, "rendered", writer.written, "expected 'rendered', got %s", writer.written)
}

type mockWriter struct {
	written string
}

func (m *mockWriter) Write(p []byte) (n int, err error) {
	m.written = "rendered"
	return len(p), nil
}

func TestRealmView(t *testing.T) {
	content := NewReaderComponent(strings.NewReader("testdata"))
	tocItems := &RealmTOCData{
		Items: []*TocItem{
			{Title: "Introduction", ID: "introduction"},
		},
	}
	data := RealmData{
		ComponentContent: content,
		TocItems:         tocItems,
	}

	view := RealmView(data)

	assert.NotNil(t, view, "expected view to be non-nil")

	templateComponent, ok := view.Component.(*TemplateComponent)
	assert.True(t, ok, "expected TemplateComponent type in view.Component")

	realmViewParams, ok := templateComponent.data.(realmViewParams)
	assert.True(t, ok, "expected realmViewParams type in component data")

	assert.Equal(t, content, realmViewParams.Article.ComponentContent, "expected component content to match")

	assert.NoError(t, view.Render(io.Discard))
}

func TestRealmViewTOCXSSPrevention(t *testing.T) {
	tests := []struct {
		name           string
		tocTitle       string
		tocID          string
		mustNotContain []string
		mustContain    []string
	}{
		{
			name:           "HTML entities in TOC title",
			tocTitle:       "&lt;script&gt;alert('XSS')&lt;/script&gt; Heading",
			tocID:          "heading",
			mustNotContain: []string{"<script>alert", "<script>"},
			mustContain:    []string{"&amp;lt;script&amp;gt;"},
		},
		{
			name:           "Image tag with onerror via entities",
			tocTitle:       "&lt;img src=x onerror=alert(1)&gt;",
			tocID:          "img",
			mustNotContain: []string{},
			mustContain:    []string{"&amp;lt;img"},
		},
		{
			name:           "SVG with onload via entities",
			tocTitle:       "&lt;svg onload=alert(1)&gt;",
			tocID:          "svg",
			mustNotContain: []string{},
			mustContain:    []string{"&amp;lt;svg"},
		},
		{
			name:           "Numeric HTML entities",
			tocTitle:       "&#60;script&#62;alert(1)&#60;/script&#62;",
			tocID:          "numeric",
			mustNotContain: []string{"<script"},
			mustContain:    []string{"&amp;#60;script"},
		},
		{
			name:           "Normal heading with ampersand",
			tocTitle:       "API & SDK",
			tocID:          "api-sdk",
			mustNotContain: []string{},
			mustContain:    []string{"API &amp; SDK"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a RealmView with potentially malicious TOC item
			content := NewReaderComponent(strings.NewReader("test content"))
			tocItems := &RealmTOCData{
				Items: []*TocItem{
					{Title: tt.tocTitle, ID: tt.tocID},
				},
			}
			data := RealmData{
				ComponentContent: content,
				TocItems:         tocItems,
			}

			view := RealmView(data)

			var buf strings.Builder
			err := view.Render(&buf)
			assert.NoError(t, err, "expected no error rendering view")

			rendered := buf.String()

			tocStart := strings.Index(rendered, `<ul class="b-toc">`)
			tocEnd := strings.LastIndex(rendered, `</ul>`)
			if tocStart == -1 || tocEnd == -1 {
				t.Fatal("could not find TOC in rendered HTML")
			}
			tocHTML := rendered[tocStart : tocEnd+5]

			for _, danger := range tt.mustNotContain {
				if strings.Contains(tocHTML, danger) {
					t.Errorf("Found unescaped dangerous pattern %q in TOC HTML.\n"+
						"TOC HTML: %s",
						danger, tocHTML)
				}
			}

			for _, safe := range tt.mustContain {
				if !strings.Contains(tocHTML, safe) {
					t.Errorf("Expected escaped pattern %q not found in TOC HTML.\n"+
						"Title: %s\nTOC HTML: %s",
						safe, tt.tocTitle, tocHTML)
				}
			}
		})
	}
}

func TestHelpView(t *testing.T) {
	functions := []*doc.JSONFunc{
		{Name: "Func1", Params: []*doc.JSONField{{Name: "param1"}}},
		{Name: "Func2", Params: []*doc.JSONField{{Name: "param1"}, {Name: "param2"}}},
	}
	data := HelpData{
		SelectedFunc: "Func1",
		Functions:    functions,
		RealmName:    "TestRealm",
	}

	view := HelpView(data)

	assert.NotNil(t, view, "expected view to be non-nil")

	templateComponent, ok := view.Component.(*TemplateComponent)
	assert.True(t, ok, "expected TemplateComponent type in view.Component")

	helpViewParams, ok := templateComponent.data.(helpViewParams)
	assert.True(t, ok, "expected helpViewParams type in component data")

	assert.Equal(t, data.RealmName, helpViewParams.RealmName, "expected realm name %s, got %s", data.RealmName, helpViewParams.RealmName)

	assert.NoError(t, view.Render(io.Discard))
}

func TestDirectoryView(t *testing.T) {
	pkgPath := "example/path"
	files := []string{"file1.gno", "file2.gno"}
	fileCounter := 2
	linkType := DirLinkTypeSource
	mode := ViewModePackage
	hasRenderMap := map[string]bool{"file1.gno": true, "file2.gno": false}

	view := DirectoryView(pkgPath, files, fileCounter, linkType, mode, hasRenderMap)

	assert.NotNil(t, view, "expected view to be non-nil")

	templateComponent, ok := view.Component.(*TemplateComponent)
	assert.True(t, ok, "expected TemplateComponent type in view.Component")

	dirData, ok := templateComponent.data.(DirData)
	assert.True(t, ok, "expected DirData type in component data")

	assert.Equal(t, pkgPath, dirData.PkgPath, "expected PkgPath %s, got %s", pkgPath, dirData.PkgPath)
	assert.Equal(t, len(files), len(dirData.FilesLinks), "expected %d files, got %d", len(files), len(dirData.FilesLinks))
	assert.Equal(t, fileCounter, dirData.FileCounter, "expected FileCounter %d, got %d", fileCounter, dirData.FileCounter)
	assert.Equal(t, mode, dirData.Mode, "expected Mode %v, got %v", mode, dirData.Mode)

	assert.NoError(t, view.Render(io.Discard))
}

func TestDirLinkType_LinkPrefix(t *testing.T) {
	cases := []struct {
		name     string
		linkType DirLinkType
		pkgPath  string
		expected string
	}{
		{
			name:     "Source link type",
			linkType: DirLinkTypeSource,
			pkgPath:  "/r/test/pkg",
			expected: "/r/test/pkg$source&file=",
		},
		{
			name:     "File link type",
			linkType: DirLinkTypeFile,
			pkgPath:  "/r/test/pkg",
			expected: "",
		},
		{
			name:     "Invalid link type",
			linkType: DirLinkType(999),
			pkgPath:  "/r/test/pkg",
			expected: "",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := tc.linkType.LinkPrefix(tc.pkgPath)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestBuildFilesLinks(t *testing.T) {
	hasRenderMap := map[string]bool{"file1.gno": true, "file2.gno": false}

	cases := []struct {
		name         string
		files        []string
		linkType     DirLinkType
		pkgPath      string
		hasRenderMap map[string]bool
		expected     FilesLinks
	}{
		{
			name:         "Source link type with multiple files",
			files:        []string{"file1.gno", "file2.gno"},
			linkType:     DirLinkTypeSource,
			pkgPath:      "/r/test/pkg",
			hasRenderMap: hasRenderMap,
			expected: FilesLinks{
				{Link: "/r/test/pkg$source&file=file1.gno", Name: "file1.gno", SourceLink: "file1.gno$source", HasRender: true},
				{Link: "/r/test/pkg$source&file=file2.gno", Name: "file2.gno", SourceLink: "file2.gno$source", HasRender: false},
			},
		},
		{
			name:         "File link type with multiple files",
			files:        []string{"file1.gno", "file2.gno"},
			linkType:     DirLinkTypeFile,
			pkgPath:      "/r/test/pkg",
			hasRenderMap: hasRenderMap,
			expected: FilesLinks{
				{Link: "file1.gno", Name: "file1.gno", SourceLink: "file1.gno$source", HasRender: true},
				{Link: "file2.gno", Name: "file2.gno", SourceLink: "file2.gno$source", HasRender: false},
			},
		},
		{
			name:         "Empty files list",
			files:        []string{},
			linkType:     DirLinkTypeSource,
			pkgPath:      "/r/test/pkg",
			hasRenderMap: nil,
			expected:     FilesLinks{},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := buildFilesLinks(tc.files, tc.linkType, tc.pkgPath, tc.hasRenderMap)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestUserView(t *testing.T) {
	data := UserData{
		Username:   "testuser",
		Handlename: "Test User",
		Bio:        "This is a test user.",
		Links: []UserLink{
			{Type: UserLinkTypeLink, URL: "https://example.com"},
			{Type: UserLinkTypeGithub, URL: "https://github.com/testuser", Title: "GitHub"},
		},
		Contributions: []UserContribution{
			{
				Title: "Realm Contribution",
				Type:  UserContributionTypeRealm,
			},
			{
				Title: "Package Contribution",
				Type:  UserContributionTypePackage,
			},
		},
		PackageCount: 2,
		RealmCount:   1,
		PureCount:    1,
		Content:      NewReaderComponent(strings.NewReader("Test content")),
	}

	view := UserView(data)

	assert.NotNil(t, view, "expected view to be non-nil")

	templateComponent, ok := view.Component.(*TemplateComponent)
	assert.True(t, ok, "expected TemplateComponent type in view.Component")

	userData, ok := templateComponent.data.(UserData)
	assert.True(t, ok, "expected UserData type in component data")

	// Assert that link title for UserLinkTypeLink was set to the host
	assert.Equal(t, "example.com", userData.Links[0].Title, "expected link title to be host")

	// Assert counts
	assert.Equal(t, 1, userData.RealmCount, "expected 1 realm")
	assert.Equal(t, 2, userData.PackageCount, "expected 2 packages")
	assert.Equal(t, 1, userData.PureCount, "expected 1 pure package")

	assert.NoError(t, view.Render(io.Discard))
}
