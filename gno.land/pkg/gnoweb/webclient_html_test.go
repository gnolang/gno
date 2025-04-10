package gnoweb

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/gnolang/gno/gnovm/pkg/doc"
	"github.com/yuin/goldmark/parser"
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

// errorWriter is a writer that always fails
type errorWriter struct{}

func (w *errorWriter) Write(p []byte) (int, error) {
	return 0, fmt.Errorf("write error")
}

// dummyGnoURL creates a dummy GnoURL for testing
func dummyGnoURL(path string) *weburl.GnoURL {
	return &weburl.GnoURL{Path: path}
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

// TestSourceFileNilWriter verifies that SourceFile returns metadata when writer is nil
func TestSourceFileNilWriter(t *testing.T) {
	// Create a mock package with a source file
	mockPkg := &MockPackage{
		Path: "test/pkg",
		Files: map[string]string{
			"test.gno": "package main\n\nfunc main() { println(\"Hello\") }",
		},
	}
	client := NewMockWebClient(mockPkg)

	// Test with nil writer
	meta, err := client.SourceFile(nil, "test/pkg", "test.gno", false)
	if err != nil {
		t.Fatalf("SourceFile returned an error: %v", err)
	}
	if meta == nil {
		t.Fatal("metadata should not be nil")
	}
	// Only check Lines if meta is not nil
	if meta.Lines != 3 {
		t.Errorf("expected 3 lines, got %d", meta.Lines)
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

func TestRenderMd_SourceFileError(t *testing.T) {
	// Create a mock package without the markdown file
	mockPkg := &MockPackage{
		Path:  "test/pkg",
		Files: map[string]string{},
	}
	client := NewMockWebClient(mockPkg)

	url := &weburl.GnoURL{Path: "test/pkg"}
	var buf bytes.Buffer
	_, err := client.RenderMd(&buf, url, "nonexistent.md")
	if err == nil {
		t.Error("expected error when file does not exist")
	}
}

func TestRenderMd_ParseError(t *testing.T) {
	// Create a mock package with a markdown file
	mockPkg := &MockPackage{
		Path: "test/pkg",
		Files: map[string]string{
			"readme.md": "# Test Markdown\n\nThis is a test.",
		},
	}
	client := NewMockWebClient(mockPkg)

	url := &weburl.GnoURL{Path: "test/pkg"}
	// Use errorWriter to simulate write error
	errorWriter := &errorWriter{}
	_, err := client.RenderMd(errorWriter, url, "readme.md")
	if err == nil {
		t.Error("expected error when writing fails")
	}
}

func TestRenderMd_WriteError(t *testing.T) {
	// Create a mock package with a markdown file
	mockPkg := &MockPackage{
		Path: "test/pkg",
		Files: map[string]string{
			"readme.md": "# Test Markdown\n\nThis is a test.",
		},
	}
	client := NewMockWebClient(mockPkg)

	url := &weburl.GnoURL{Path: "test/pkg"}
	// Use errorWriter to simulate write failure
	errorWriter := &errorWriter{}
	_, err := client.RenderMd(errorWriter, url, "readme.md")
	if err == nil {
		t.Error("expected error when writing fails")
	}
}

// TestRenderMdNotFound verifies error handling when markdown file is not found
func TestRenderMdNotFound(t *testing.T) {
	// Create a mock package
	mockPkg := &MockPackage{
		Path:  "test/pkg",
		Files: map[string]string{
			// No markdown files
		},
	}
	client := NewMockWebClient(mockPkg)

	var buf bytes.Buffer
	url := dummyGnoURL("test/pkg")
	_, err := client.RenderMd(&buf, url, "nonexistent.md")
	if err == nil {
		t.Error("expected an error when file is not found")
	}
}

// TestRenderMdWriteError verifies error handling when writing rendered content fails
func TestRenderMdWriteError(t *testing.T) {
	// Create a mock package with a markdown file
	mockPkg := &MockPackage{
		Path: "test/pkg",
		Files: map[string]string{
			"test.md": "# Test\n\nThis is a test.",
		},
	}
	client := NewMockWebClient(mockPkg)

	// Use errorWriter to simulate write failure
	errorWriter := &errorWriter{}
	url := dummyGnoURL("test/pkg")
	_, err := client.RenderMd(errorWriter, url, "test.md")
	if err == nil {
		t.Error("expected an error when writing fails")
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

// TestParseMarkdownWithContext verifies markdown parsing with context options
func TestParseMarkdownWithContext(t *testing.T) {
	// Create a mock package
	mockPkg := &MockPackage{
		Path: "test/pkg",
	}
	client := NewMockWebClient(mockPkg)

	markdownContent := []byte("# Hello\n\nThis is a *test* markdown.")
	var buf bytes.Buffer

	// Create a dummy context option
	ctxOpt := parser.WithContext(parser.NewContext())

	node, err := client.ParseMarkdown(&buf, markdownContent, ctxOpt)
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

// TestParseMarkdownError verifies error handling in markdown parsing
func TestParseMarkdownError(t *testing.T) {
	// Create a mock package
	mockPkg := &MockPackage{
		Path: "test/pkg",
	}
	client := NewMockWebClient(mockPkg)

	// Create a writer that will fail
	errorWriter := &errorWriter{}
	markdownContent := []byte("# Hello\n\nThis is a *test* markdown.")

	_, err := client.ParseMarkdown(errorWriter, markdownContent)
	if err == nil {
		t.Error("expected an error when writer fails")
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
