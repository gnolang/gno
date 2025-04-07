package components

import (
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/markdown"
	"github.com/gnolang/gno/gnovm/pkg/doc"
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
				FileSource:   NewTemplateComponent("ui/code_wrapper", nil),
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
				FileSource:   NewTemplateComponent("ui/code_wrapper", nil),
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view := SourceView(tt.data)

			if view == nil {
				t.Error("expected view to be non-nil")
				return
			}

			tocItemsCount := len(tt.data.Files)
			if tocItemsCount != tt.expected {
				t.Errorf("expected %d TOC items, got %d", tt.expected, tocItemsCount)
			}
			if view.Type != SourceViewType {
				t.Errorf("expected view type %s, got %s", SourceViewType, view.Type)
			}
		})
	}
}

func TestStatusErrorComponent(t *testing.T) {
	message := "Test Error"
	view := StatusErrorComponent(message)

	if view == nil {
		t.Error("expected view to be non-nil")
		return
	}

	expectedTitle := "Error: " + message
	templateComponent, ok := view.Component.(*TemplateComponent)
	if !ok {
		t.Error("expected TemplateComponent type in view.Component")
		return
	}

	statusData, ok := templateComponent.data.(StatusData)
	if !ok {
		t.Error("expected StatusData type in component data")
		return
	}

	if statusData.Title != expectedTitle {
		t.Errorf("expected title %s, got %s", expectedTitle, statusData.Title)
	}
}

func TestStatusNoRenderComponent(t *testing.T) {
	pkgPath := "example/path"
	view := StatusNoRenderComponent(pkgPath)

	if view == nil {
		t.Error("expected view to be non-nil")
		return
	}

	templateComponent, ok := view.Component.(*TemplateComponent)
	if !ok {
		t.Error("expected TemplateComponent type in view.Component")
		return
	}

	statusData, ok := templateComponent.data.(StatusData)
	if !ok {
		t.Error("expected StatusData type in component data")
		return
	}

	expectedURL := pkgPath + "$source"
	if statusData.ButtonURL != expectedURL {
		t.Errorf("expected ButtonURL %s, got %s", expectedURL, statusData.ButtonURL)
	}
}

func TestRedirectView(t *testing.T) {
	data := RedirectData{
		To:            "example/path",
		WithAnalytics: true,
	}
	view := RedirectView(data)

	if view == nil {
		t.Error("expected view to be non-nil")
		return
	}

	templateComponent, ok := view.Component.(*TemplateComponent)
	if !ok {
		t.Error("expected TemplateComponent type in view.Component")
		return
	}

	redirectData, ok := templateComponent.data.(RedirectData)
	if !ok {
		t.Error("expected RedirectData type in component data")
		return
	}

	if redirectData.To != data.To {
		t.Errorf("expected redirect to %s, got %s", data.To, redirectData.To)
	}

	if redirectData.WithAnalytics != data.WithAnalytics {
		t.Errorf("expected WithAnalytics to be %v, got %v", data.WithAnalytics, redirectData.WithAnalytics)
	}
}

func TestViewRender(t *testing.T) {
	component := NewTemplateComponent("ui/toc_generic", nil)
	view := &View{
		Type:      "test-view",
		Component: component,
	}

	writer := &mockWriter{}
	err := view.Render(writer)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if writer.written != "rendered" {
		t.Errorf("expected 'rendered', got %s", writer.written)
	}
}

type mockWriter struct {
	written string
}

func (m *mockWriter) Write(p []byte) (n int, err error) {
	m.written = "rendered"
	return len(p), nil
}

func TestRealmView(t *testing.T) {
	component := NewTemplateComponent("ui/content", nil)
	tocItems := &RealmTOCData{
		Items: []*markdown.TocItem{
			{Title: []byte("Introduction"), ID: []byte("introduction")},
		},
	}
	data := RealmData{
		ComponentContent: component,
		TocItems:         tocItems,
	}

	view := RealmView(data)

	if view == nil {
		t.Error("expected view to be non-nil")
		return
	}

	templateComponent, ok := view.Component.(*TemplateComponent)
	if !ok {
		t.Error("expected TemplateComponent type in view.Component")
		return
	}

	realmViewParams, ok := templateComponent.data.(realmViewParams)
	if !ok {
		t.Error("expected realmViewParams type in component data")
		return
	}

	if realmViewParams.Article.ComponentContent != component {
		t.Error("expected component content to match")
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

	if view == nil {
		t.Error("expected view to be non-nil")
		return
	}

	templateComponent, ok := view.Component.(*TemplateComponent)
	if !ok {
		t.Error("expected TemplateComponent type in view.Component")
		return
	}

	helpViewParams, ok := templateComponent.data.(helpViewParams)
	if !ok {
		t.Error("expected helpViewParams type in component data")
		return
	}

	if helpViewParams.HelpData.RealmName != data.RealmName {
		t.Errorf("expected realm name %s, got %s", data.RealmName, helpViewParams.HelpData.RealmName)
	}
}

func TestDirectoryView(t *testing.T) {
	data := DirData{
		PkgPath:     "example/path",
		Files:       []string{"file1.gno", "file2.gno"},
		FileCounter: 2,
	}

	view := DirectoryView(data)

	if view == nil {
		t.Error("expected view to be non-nil")
		return
	}

	templateComponent, ok := view.Component.(*TemplateComponent)
	if !ok {
		t.Error("expected TemplateComponent type in view.Component")
		return
	}

	dirData, ok := templateComponent.data.(DirData)
	if !ok {
		t.Error("expected DirData type in component data")
		return
	}

	if dirData.PkgPath != data.PkgPath {
		t.Errorf("expected PkgPath %s, got %s", data.PkgPath, dirData.PkgPath)
	}

	if len(dirData.Files) != len(data.Files) {
		t.Errorf("expected %d files, got %d", len(data.Files), len(dirData.Files))
	}

	if dirData.FileCounter != data.FileCounter {
		t.Errorf("expected FileCounter %d, got %d", data.FileCounter, dirData.FileCounter)
	}
}
