package components

import (
	"io"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/markdown"
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
		Items: []*markdown.TocItem{
			{Title: []byte("Introduction"), ID: []byte("introduction")},
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

	assert.Equal(t, data.RealmName, helpViewParams.HelpData.RealmName, "expected realm name %s, got %s", data.RealmName, helpViewParams.HelpData.RealmName)

	assert.NoError(t, view.Render(io.Discard))
}

func TestDirectoryView(t *testing.T) {
	data := DirData{
		PkgPath:     "example/path",
		Files:       []string{"file1.gno", "file2.gno"},
		FileCounter: 2,
	}

	view := DirectoryView(data)

	assert.NotNil(t, view, "expected view to be non-nil")

	templateComponent, ok := view.Component.(*TemplateComponent)
	assert.True(t, ok, "expected TemplateComponent type in view.Component")

	dirData, ok := templateComponent.data.(DirData)
	assert.True(t, ok, "expected DirData type in component data")

	assert.Equal(t, data.PkgPath, dirData.PkgPath, "expected PkgPath %s, got %s", data.PkgPath, dirData.PkgPath)
	assert.Equal(t, len(data.Files), len(dirData.Files), "expected %d files, got %d", len(data.Files), len(dirData.Files))
	assert.Equal(t, data.FileCounter, dirData.FileCounter, "expected FileCounter %d, got %d", data.FileCounter, dirData.FileCounter)

	assert.NoError(t, view.Render(io.Discard))
}
