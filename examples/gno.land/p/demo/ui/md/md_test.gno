package md

import (
	"testing"
)

// -----------------------------------------------------------------------------
// Document Builder Tests
// -----------------------------------------------------------------------------

func TestNew(t *testing.T) {
	doc := New()
	if doc == nil {
		t.Fatal("New() returned nil")
	}
	if doc.String() != "" {
		t.Errorf("New document should be empty, got: %q", doc.String())
	}
}

func TestDocument_Write(t *testing.T) {
	doc := New().Write("hello").Write(" world")
	expected := "hello world"
	if doc.String() != expected {
		t.Errorf("Write failed: expected %q, got %q", expected, doc.String())
	}
}

func TestDocument_Writeln(t *testing.T) {
	doc := New().Writeln("line1").Writeln("line2")
	expected := "line1\nline2\n"
	if doc.String() != expected {
		t.Errorf("Writeln failed: expected %q, got %q", expected, doc.String())
	}
}

func TestDocument_LF(t *testing.T) {
	doc := New().Write("a").LF().Write("b")
	expected := "a\nb"
	if doc.String() != expected {
		t.Errorf("LF failed: expected %q, got %q", expected, doc.String())
	}
}

func TestDocument_BR(t *testing.T) {
	doc := New().Write("line1").BR().Write("line2")
	expected := "line1  \nline2"
	if doc.String() != expected {
		t.Errorf("BR failed: expected %q, got %q", expected, doc.String())
	}
}

func TestDocument_Headers(t *testing.T) {
	tests := []struct {
		name     string
		fn       func(*Document, string) *Document
		level    string
		expected string
	}{
		{"H1", (*Document).H1, "Title", "# Title\n\n"},
		{"H2", (*Document).H2, "Title", "## Title\n\n"},
		{"H3", (*Document).H3, "Title", "### Title\n\n"},
		{"H4", (*Document).H4, "Title", "#### Title\n\n"},
		{"H5", (*Document).H5, "Title", "##### Title\n\n"},
		{"H6", (*Document).H6, "Title", "###### Title\n\n"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			doc := New()
			tc.fn(doc, tc.level)
			if doc.String() != tc.expected {
				t.Errorf("%s failed: expected %q, got %q", tc.name, tc.expected, doc.String())
			}
		})
	}
}

func TestDocument_P(t *testing.T) {
	doc := New().P("This is a paragraph.")
	expected := "This is a paragraph.\n\n"
	if doc.String() != expected {
		t.Errorf("P failed: expected %q, got %q", expected, doc.String())
	}
}

func TestDocument_Quote(t *testing.T) {
	doc := New().Quote("This is quoted text.")
	expected := "> This is quoted text.\n\n"
	if doc.String() != expected {
		t.Errorf("Quote failed: expected %q, got %q", expected, doc.String())
	}
}

func TestDocument_Quote_Multiline(t *testing.T) {
	doc := New().Quote("Line 1\nLine 2")
	expected := "> Line 1\n> Line 2\n\n"
	if doc.String() != expected {
		t.Errorf("Quote multiline failed: expected %q, got %q", expected, doc.String())
	}
}

func TestDocument_HR(t *testing.T) {
	doc := New().HR()
	expected := "---\n\n"
	if doc.String() != expected {
		t.Errorf("HR failed: expected %q, got %q", expected, doc.String())
	}
}

func TestDocument_CodeBlock(t *testing.T) {
	doc := New().CodeBlock("fmt.Println(\"Hello\")", "go")
	expected := "```go\nfmt.Println(\"Hello\")\n```\n\n"
	if doc.String() != expected {
		t.Errorf("CodeBlock failed: expected %q, got %q", expected, doc.String())
	}
}

func TestDocument_CodeBlock_NoLang(t *testing.T) {
	doc := New().CodeBlock("some code", "")
	expected := "```\nsome code\n```\n\n"
	if doc.String() != expected {
		t.Errorf("CodeBlock no lang failed: expected %q, got %q", expected, doc.String())
	}
}

func TestDocument_CodeBlock_TrailingNewline(t *testing.T) {
	doc := New().CodeBlock("code\n", "")
	expected := "```\ncode\n```\n\n"
	if doc.String() != expected {
		t.Errorf("CodeBlock trailing newline failed: expected %q, got %q", expected, doc.String())
	}
}

func TestDocument_Bullet(t *testing.T) {
	doc := New().Bullet("Item 1", "Item 2", "Item 3")
	expected := "- Item 1\n- Item 2\n- Item 3\n\n"
	if doc.String() != expected {
		t.Errorf("Bullet failed: expected %q, got %q", expected, doc.String())
	}
}

func TestDocument_Bullet_Empty(t *testing.T) {
	doc := New().Bullet()
	expected := "\n"
	if doc.String() != expected {
		t.Errorf("Bullet empty failed: expected %q, got %q", expected, doc.String())
	}
}

func TestDocument_Numbered(t *testing.T) {
	doc := New().Numbered("First", "Second", "Third")
	expected := "1. First\n2. Second\n3. Third\n\n"
	if doc.String() != expected {
		t.Errorf("Numbered failed: expected %q, got %q", expected, doc.String())
	}
}

func TestDocument_Numbered_Empty(t *testing.T) {
	doc := New().Numbered()
	expected := "\n"
	if doc.String() != expected {
		t.Errorf("Numbered empty failed: expected %q, got %q", expected, doc.String())
	}
}

func TestDocument_Checklist(t *testing.T) {
	items := []string{"Do this", "Done that", "And this"}
	checked := []bool{false, true, false}
	doc := New().Checklist(items, checked)
	expected := "- [ ] Do this\n- [x] Done that\n- [ ] And this\n\n"
	if doc.String() != expected {
		t.Errorf("Checklist failed: expected %q, got %q", expected, doc.String())
	}
}

func TestDocument_Checklist_MoreItemsThanChecked(t *testing.T) {
	items := []string{"Item 1", "Item 2", "Item 3"}
	checked := []bool{true}
	doc := New().Checklist(items, checked)
	expected := "- [x] Item 1\n- [ ] Item 2\n- [ ] Item 3\n\n"
	if doc.String() != expected {
		t.Errorf("Checklist partial checked failed: expected %q, got %q", expected, doc.String())
	}
}

func TestDocument_Table(t *testing.T) {
	headers := []string{"Name", "Age"}
	rows := [][]string{
		{"Alice", "30"},
		{"Bob", "25"},
	}
	doc := New().Table(headers, rows)
	expected := "| Name  | Age |\n| ----- | --- |\n| Alice | 30  |\n| Bob   | 25  |\n\n"
	if doc.String() != expected {
		t.Errorf("Table failed: expected %q, got %q", expected, doc.String())
	}
}

func TestDocument_Table_EmptyHeaders(t *testing.T) {
	doc := New().Table([]string{}, [][]string{})
	if doc.String() != "" {
		t.Errorf("Table with empty headers should return empty, got: %q", doc.String())
	}
}

func TestDocument_Table_MissingCells(t *testing.T) {
	headers := []string{"A", "B", "C"}
	rows := [][]string{
		{"1"},
		{"1", "2"},
	}
	doc := New().Table(headers, rows)
	// Should handle missing cells gracefully
	if doc.String() == "" {
		t.Error("Table with missing cells should not return empty")
	}
}

func TestDocument_Chaining(t *testing.T) {
	doc := New().
		H1("Title").
		P("Introduction").
		H2("Section 1").
		Bullet("Point A", "Point B").
		HR().
		Quote("A famous quote")

	// Verify it contains expected elements
	output := doc.String()
	if output == "" {
		t.Error("Chained document should not be empty")
	}
	// Basic content checks
	if len(output) < 50 {
		t.Errorf("Chained document seems too short: %q", output)
	}
}

// -----------------------------------------------------------------------------
// Standalone Function Tests
// -----------------------------------------------------------------------------

func TestStandalone_Headers(t *testing.T) {
	tests := []struct {
		fn       func(string) string
		input    string
		expected string
	}{
		{H1, "Title", "# Title\n\n"},
		{H2, "Title", "## Title\n\n"},
		{H3, "Title", "### Title\n\n"},
		{H4, "Title", "#### Title\n\n"},
		{H5, "Title", "##### Title\n\n"},
		{H6, "Title", "###### Title\n\n"},
	}

	for _, tc := range tests {
		result := tc.fn(tc.input)
		if result != tc.expected {
			t.Errorf("expected %q, got %q", tc.expected, result)
		}
	}
}

func TestStandalone_P(t *testing.T) {
	result := P("A paragraph")
	expected := "A paragraph\n\n"
	if result != expected {
		t.Errorf("P failed: expected %q, got %q", expected, result)
	}
}

func TestStandalone_HR(t *testing.T) {
	result := HR()
	expected := "---\n\n"
	if result != expected {
		t.Errorf("HR failed: expected %q, got %q", expected, result)
	}
}

func TestStandalone_Quote(t *testing.T) {
	result := Quote("Quoted text")
	expected := "> Quoted text\n\n"
	if result != expected {
		t.Errorf("Quote failed: expected %q, got %q", expected, result)
	}
}

func TestStandalone_Quote_Multiline(t *testing.T) {
	result := Quote("Line 1\nLine 2")
	expected := "> Line 1\n> Line 2\n\n"
	if result != expected {
		t.Errorf("Quote multiline failed: expected %q, got %q", expected, result)
	}
}

func TestStandalone_CodeBlock(t *testing.T) {
	result := CodeBlock("code here", "python")
	expected := "```python\ncode here\n```\n\n"
	if result != expected {
		t.Errorf("CodeBlock failed: expected %q, got %q", expected, result)
	}
}

func TestStandalone_CodeBlock_TrailingNewline(t *testing.T) {
	result := CodeBlock("code\n", "")
	expected := "```\ncode\n```\n\n"
	if result != expected {
		t.Errorf("CodeBlock trailing newline failed: expected %q, got %q", expected, result)
	}
}

func TestStandalone_Bullet(t *testing.T) {
	result := Bullet("A", "B", "C")
	expected := "- A\n- B\n- C\n\n"
	if result != expected {
		t.Errorf("Bullet failed: expected %q, got %q", expected, result)
	}
}

func TestStandalone_Numbered(t *testing.T) {
	result := Numbered("First", "Second")
	expected := "1. First\n2. Second\n\n"
	if result != expected {
		t.Errorf("Numbered failed: expected %q, got %q", expected, result)
	}
}

func TestStandalone_Table(t *testing.T) {
	headers := []string{"Col1", "Col2"}
	rows := [][]string{{"a", "b"}}
	result := Table(headers, rows)
	if result == "" {
		t.Error("Table should not return empty")
	}
	// Check for key elements
	if len(result) < 20 {
		t.Errorf("Table output seems too short: %q", result)
	}
}

func TestStandalone_Table_Empty(t *testing.T) {
	result := Table([]string{}, [][]string{})
	if result != "" {
		t.Errorf("Table with empty headers should return empty, got: %q", result)
	}
}

// -----------------------------------------------------------------------------
// Inline Formatting Tests
// -----------------------------------------------------------------------------

func TestInline_Bold(t *testing.T) {
	result := Bold("text")
	expected := "**text**"
	if result != expected {
		t.Errorf("Bold failed: expected %q, got %q", expected, result)
	}
}

func TestInline_Italic(t *testing.T) {
	result := Italic("text")
	expected := "_text_"
	if result != expected {
		t.Errorf("Italic failed: expected %q, got %q", expected, result)
	}
}

func TestInline_BoldItalic(t *testing.T) {
	result := BoldItalic("text")
	expected := "***text***"
	if result != expected {
		t.Errorf("BoldItalic failed: expected %q, got %q", expected, result)
	}
}

func TestInline_InlineCode(t *testing.T) {
	result := InlineCode("code")
	expected := "`code`"
	if result != expected {
		t.Errorf("InlineCode failed: expected %q, got %q", expected, result)
	}
}

func TestInline_Strikethrough(t *testing.T) {
	result := Strikethrough("deleted")
	expected := "~~deleted~~"
	if result != expected {
		t.Errorf("Strikethrough failed: expected %q, got %q", expected, result)
	}
}

func TestInline_Link(t *testing.T) {
	result := Link("Click here", "https://example.com")
	expected := "[Click here](https://example.com)"
	if result != expected {
		t.Errorf("Link failed: expected %q, got %q", expected, result)
	}
}

func TestInline_Image(t *testing.T) {
	result := Image("Alt text", "https://example.com/img.png")
	expected := "![Alt text](https://example.com/img.png)"
	if result != expected {
		t.Errorf("Image failed: expected %q, got %q", expected, result)
	}
}

func TestInline_LinkWithTitle(t *testing.T) {
	result := LinkWithTitle("Click", "https://example.com", "A title")
	expected := "[Click](https://example.com \"A title\")"
	if result != expected {
		t.Errorf("LinkWithTitle failed: expected %q, got %q", expected, result)
	}
}

func TestInline_ImageWithTitle(t *testing.T) {
	result := ImageWithTitle("Alt", "https://example.com/img.png", "Title")
	expected := "![Alt](https://example.com/img.png \"Title\")"
	if result != expected {
		t.Errorf("ImageWithTitle failed: expected %q, got %q", expected, result)
	}
}

// -----------------------------------------------------------------------------
// Integration/Composition Tests
// -----------------------------------------------------------------------------

func TestComposition_InlineInDocument(t *testing.T) {
	doc := New().
		H1("Welcome").
		P("This is " + Bold("bold") + " and " + Italic("italic") + " text.").
		P("Check out " + Link("this link", "https://gno.land"))

	output := doc.String()
	if output == "" {
		t.Error("Composed document should not be empty")
	}

	// Verify inline elements are present
	expectedParts := []string{"**bold**", "_italic_", "[this link](https://gno.land)"}
	for _, part := range expectedParts {
		found := false
		for i := 0; i <= len(output)-len(part); i++ {
			if output[i:i+len(part)] == part {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find %q in output", part)
		}
	}
}

func TestComposition_StandaloneConcat(t *testing.T) {
	output := H1("Title") + P("Intro") + Bullet("A", "B") + HR()

	expectedParts := []string{"# Title", "Intro", "- A", "- B", "---"}
	for _, part := range expectedParts {
		found := false
		for i := 0; i <= len(output)-len(part); i++ {
			if output[i:i+len(part)] == part {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find %q in output", part)
		}
	}
}

func TestRealWorldExample_BlogPost(t *testing.T) {
	doc := New().
		H1("My Blog Post").
		P("Published on 2024-01-15").
		H2("Introduction").
		P("Welcome to my blog! This post covers " + Bold("important topics") + ".").
		H2("Main Points").
		Bullet(
			"First point with "+Link("a link", "/page1"),
			"Second point with "+InlineCode("code"),
			"Third point",
		).
		H2("Code Example").
		CodeBlock("func main() {\n\tfmt.Println(\"Hello\")\n}", "go").
		HR().
		Quote("This is a notable quote from someone.").
		P("Thanks for reading!")

	output := doc.String()
	if len(output) < 200 {
		t.Errorf("Blog post example seems too short: %d chars", len(output))
	}
}

func TestRealWorldExample_DataTable(t *testing.T) {
	headers := []string{"ID", "Name", "Status", "Action"}
	rows := [][]string{
		{"1", "Alice", "Active", Link("View", "/users/1")},
		{"2", "Bob", "Inactive", Link("View", "/users/2")},
		{"3", "Charlie", "Active", Link("View", "/users/3")},
	}

	doc := New().
		H2("User List").
		Table(headers, rows).
		P(Italic("Last updated: 2024-01-15"))

	output := doc.String()
	if output == "" {
		t.Error("Data table example should not be empty")
	}
}
