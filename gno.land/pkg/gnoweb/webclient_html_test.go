package gnoweb

import (
	"bytes"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/gnolang/gno/gnovm/pkg/doc"
)

// JSONParam represents a function parameter in JSON format
type JSONParam struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// JSONResult represents a function result in JSON format
type JSONResult struct {
	Type string `json:"type"`
}

// --- Unit Tests ---

// TestDoc verifies that the Doc method returns proper JSON documentation.
func TestDoc(t *testing.T) {
	// Create a mock package with functions
	mockPkg := &MockPackage{
		Path: "test/pkg",
		Functions: []*doc.JSONFunc{
			{
				Name: "TestFunc",
				Params: []*doc.JSONField{
					{Name: "param1", Type: "string"},
				},
				Results: []*doc.JSONField{
					{Type: "string"},
				},
			},
		},
	}
	client := NewMockWebClient(mockPkg)

	jdoc, err := client.Doc("test/pkg")
	if err != nil {
		t.Fatalf("Doc returned an error: %v", err)
	}
	if len(jdoc.Funcs) != 1 || jdoc.Funcs[0].Name != "TestFunc" {
		t.Error("documentation does not contain the expected functions")
	}
}

// TestSourceFile verifies source file rendering with and without formatting.
func TestSourceFile(t *testing.T) {
	// Create a mock package with a source file
	mockPkg := &MockPackage{
		Path: "test/pkg",
		Files: map[string]string{
			"test.gno": "package main\n\nfunc main() { println(\"Hello\") }",
		},
	}
	client := NewMockWebClient(mockPkg)

	// Test formatted mode
	var buf bytes.Buffer
	meta, err := client.SourceFile(&buf, "test/pkg", "test.gno", false)
	if err != nil {
		t.Fatalf("SourceFile (formatted) returned an error: %v", err)
	}
	if meta.Lines == 0 {
		t.Error("number of lines should not be zero")
	}
	output := buf.String()
	if !strings.Contains(output, "package main") {
		t.Error("formatted output does not contain expected source content")
	}

	// Test raw mode
	buf.Reset()
	metaRaw, err := client.SourceFile(&buf, "test/pkg", "test.gno", true)
	if err != nil {
		t.Fatalf("SourceFile (raw) returned an error: %v", err)
	}
	rawOutput := buf.String()
	if !strings.Contains(rawOutput, "package main") {
		t.Error("raw output does not contain expected source content")
	}
	if meta.Lines != metaRaw.Lines {
		t.Error("number of lines should be identical in raw and formatted mode")
	}
}

// TestSources verifies the retrieval of file list.
func TestSources(t *testing.T) {
	// Create a mock package with multiple files
	mockPkg := &MockPackage{
		Path: "test/pkg",
		Files: map[string]string{
			"file1.gno": "package main",
			"file2.gno": "package main",
		},
	}
	client := NewMockWebClient(mockPkg)

	files, err := client.Sources("test/pkg")
	if err != nil {
		t.Fatalf("Sources returned an error: %v", err)
	}
	if len(files) != 2 || files[0] != "file1.gno" {
		t.Errorf("unexpected file list: %v", files)
	}
}

// TestRenderRealm verifies realm rendering.
func TestRenderRealm(t *testing.T) {
	// Create a mock package with a Render function
	mockPkg := &MockPackage{
		Path:   "test/pkg",
		Domain: "test",
		Functions: []*doc.JSONFunc{
			{
				Name: "Render",
				Params: []*doc.JSONField{
					{Type: "string"},
				},
				Results: []*doc.JSONField{
					{Type: "string"},
				},
			},
		},
	}
	client := NewMockWebClient(mockPkg)

	url := &weburl.GnoURL{Path: "test/pkg"}
	var buf bytes.Buffer
	meta, err := client.RenderRealm(&buf, url)
	if err != nil {
		t.Fatalf("RenderRealm returned an error: %v", err)
	}
	rendered := buf.String()
	if !strings.Contains(rendered, "[test]/test/pkg:") {
		t.Error("realm rendering does not contain expected format")
	}
	if meta == nil {
		t.Error("metadata should not be nil")
	}
}

// TestRenderMd verifies markdown file rendering.
func TestRenderMd(t *testing.T) {
	// Create a mock package with a markdown file
	mockPkg := &MockPackage{
		Path: "test/pkg",
		Files: map[string]string{
			"readme.md": "# Test Markdown\n\nThis is a test.",
		},
	}
	client := NewMockWebClient(mockPkg)

	url := &weburl.GnoURL{Path: "test/pkg"}
	var buf bytes.Buffer
	meta, err := client.RenderMd(&buf, url, "readme.md")
	if err != nil {
		t.Fatalf("RenderMd returned an error: %v", err)
	}
	rendered := buf.String()
	if !strings.Contains(rendered, "<h1") {
		t.Error("markdown rendering should contain <h1> tag")
	}
	if meta == nil {
		t.Error("metadata should not be nil")
	}
}

// TestParseMarkdown verifies markdown parsing and rendering.
func TestParseMarkdown(t *testing.T) {
	// Create a mock package with a markdown file
	mockPkg := &MockPackage{
		Path: "test/pkg",
		Files: map[string]string{
			"test.md": "# Hello\n\nThis is a *test* markdown.",
		},
	}
	client := NewMockWebClient(mockPkg)

	markdownContent := []byte("# Hello\n\nThis is a *test* markdown.")
	var buf bytes.Buffer
	node, err := client.ParseMarkdown(&buf, markdownContent)
	if err != nil {
		t.Fatalf("ParseMarkdown returned an error: %v", err)
	}
	if node == nil {
		t.Error("AST should not be nil")
	}
	rendered := buf.String()
	if !strings.Contains(rendered, "<h1") {
		t.Error("HTML rendering should contain <h1> tag")
	}
}

// TestFormatSource verifies source code syntax highlighting.
func TestFormatSource(t *testing.T) {
	// Create a mock package with a source file
	mockPkg := &MockPackage{
		Path: "test/pkg",
		Files: map[string]string{
			"test.gno": "package main\n\nfunc main() { println(\"Hello\") }",
		},
	}
	client := NewMockWebClient(mockPkg)

	source := []byte("package main\n\nfunc main() { println(\"Hello\") }")
	var buf bytes.Buffer
	err := client.FormatSource(&buf, "test.gno", source)
	if err != nil {
		t.Fatalf("FormatSource returned an error: %v", err)
	}
	formatted := buf.String()
	if !strings.Contains(formatted, "chroma-") {
		t.Error("formatted code should contain Chroma CSS classes (e.g., 'chroma-')")
	}
}

// TestWriteFormatterCSS verifies that the generated CSS is not empty.
func TestWriteFormatterCSS(t *testing.T) {
	// Create a mock package with a source file
	mockPkg := &MockPackage{
		Path: "test/pkg",
		Files: map[string]string{
			"test.gno": "package main\n\nfunc main() { println(\"Hello\") }",
		},
	}
	client := NewMockWebClient(mockPkg)

	var buf bytes.Buffer
	err := client.WriteFormatterCSS(&buf)
	if err != nil {
		t.Fatalf("WriteFormatterCSS returned an error: %v", err)
	}
	css := buf.String()
	if len(css) == 0 {
		t.Error("generated CSS should not be empty")
	}
}
