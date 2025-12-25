package mictofun

import (
	"testing"
)

// ============================================================================
// DOCUMENT BUILDER TESTS
// ============================================================================

func TestNew(t *testing.T) {
	doc := New()
	if doc == nil {
		t.Fatal("New() returned nil")
	}
	if doc.String() != "" {
		t.Errorf("New document should be empty, got: %q", doc.String())
	}
}

func TestHeadings(t *testing.T) {
	tests := []struct {
		name     string
		input    func() string
		expected string
	}{
		{"H1", func() string { return New().H1("Title").String() }, "# Title\n"},
		{"H2", func() string { return New().H2("Title").String() }, "## Title\n"},
		{"H3", func() string { return New().H3("Title").String() }, "### Title\n"},
		{"H4", func() string { return New().H4("Title").String() }, "#### Title\n"},
		{"H5", func() string { return New().H5("Title").String() }, "##### Title\n"},
		{"H6", func() string { return New().H6("Title").String() }, "###### Title\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input()
			if got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestTextFormatting(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Bold", Bold("test"), "**test**"},
		{"Italic", Italic("test"), "*test*"},
		{"Strike", Strike("test"), "~~test~~"},
		{"Code", Code("test"), "`test`"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.input != tt.expected {
				t.Errorf("got %q, want %q", tt.input, tt.expected)
			}
		})
	}
}

func TestParagraph(t *testing.T) {
	doc := New().P("Hello world").String()
	expected := "Hello world\n"
	if doc != expected {
		t.Errorf("got %q, want %q", doc, expected)
	}
}

func TestQuote(t *testing.T) {
	doc := New().Quote("Hello\nWorld").String()
	expected := "> Hello\n> World\n"
	if doc != expected {
		t.Errorf("got %q, want %q", doc, expected)
	}
}

func TestCodeBlock(t *testing.T) {
	doc := New().CodeBlock("fmt.Println()", "go").String()
	expected := "```go\nfmt.Println()\n```\n"
	if doc != expected {
		t.Errorf("got %q, want %q", doc, expected)
	}
}

func TestHR(t *testing.T) {
	doc := New().HR().String()
	expected := "\n---\n"
	if doc != expected {
		t.Errorf("got %q, want %q", doc, expected)
	}
}

// ============================================================================
// LIST TESTS
// ============================================================================

func TestBulletList(t *testing.T) {
	doc := New().Bullet("Item 1", "Item 2", "Item 3").String()
	expected := "- Item 1\n- Item 2\n- Item 3\n"
	if doc != expected {
		t.Errorf("got %q, want %q", doc, expected)
	}
}

func TestNumberedList(t *testing.T) {
	doc := New().Numbered("First", "Second", "Third").String()
	expected := "1. First\n2. Second\n3. Third\n"
	if doc != expected {
		t.Errorf("got %q, want %q", doc, expected)
	}
}

func TestTodoList(t *testing.T) {
	doc := New().Todo([]string{"Done", "Pending"}, []bool{true, false}).String()
	expected := "- [x] Done\n- [ ] Pending\n"
	if doc != expected {
		t.Errorf("got %q, want %q", doc, expected)
	}
}

func TestMultilineBullet(t *testing.T) {
	doc := New().Bullet("Line1\nLine2", "Single").String()
	expected := "- Line1\n  Line2\n- Single\n"
	if doc != expected {
		t.Errorf("got %q, want %q", doc, expected)
	}
}

// ============================================================================
// LINK & IMAGE TESTS
// ============================================================================

func TestLink(t *testing.T) {
	result := Link("Click here", "https://gno.land")
	expected := "[Click here](https://gno.land)"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestImage(t *testing.T) {
	result := Image("Logo", "https://gno.land/logo.png")
	expected := "![Logo](https://gno.land/logo.png)"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestUserLink(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"moul", "[@moul](/u/moul)"},
		{"g1abc123", "[g1abc123](/u/g1abc123)"},
	}

	for _, tt := range tests {
		result := UserLink(tt.input)
		if result != tt.expected {
			t.Errorf("UserLink(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestRealmLink(t *testing.T) {
	result := RealmLink("r/demo/foo", "Demo Realm")
	expected := "[Demo Realm](/r/demo/foo)"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

// ============================================================================
// TABLE TESTS
// ============================================================================

func TestBasicTable(t *testing.T) {
	table := NewTable().
		Headers("Name", "Age").
		Row("Alice", "30").
		Row("Bob", "25")

	expected := "| Name | Age |\n| --- | --- |\n| Alice | 30 |\n| Bob | 25 |\n"
	if table.String() != expected {
		t.Errorf("got %q, want %q", table.String(), expected)
	}
}

func TestTableAlignment(t *testing.T) {
	table := NewTable().
		Headers("Left", "Center", "Right").
		SetAlign(AlignLeft, AlignCenter, AlignRight).
		Row("L", "C", "R")

	result := table.String()
	if !containsString(result, ":---:") {
		t.Error("Expected center alignment marker :---:")
	}
	if !containsString(result, "---:") {
		t.Error("Expected right alignment marker ---:")
	}
}

func TestEmptyTable(t *testing.T) {
	table := NewTable()
	if table.String() != "" {
		t.Errorf("Empty table should return empty string, got %q", table.String())
	}
}

func TestTableCellEscaping(t *testing.T) {
	table := NewTable().
		Headers("Data").
		Row("Value with | pipe")

	result := table.String()
	if containsString(result, "| pipe |") {
		t.Error("Pipe character should be escaped in table cells")
	}
}

// ============================================================================
// ALERT TESTS
// ============================================================================

func TestAlerts(t *testing.T) {
	tests := []struct {
		name     string
		fn       func(string) string
		marker   string
	}{
		{"Note", Note, "[!NOTE]"},
		{"Tip", Tip, "[!TIP]"},
		{"Important", Important, "[!IMPORTANT]"},
		{"Warning", Warning, "[!WARNING]"},
		{"Caution", Caution, "[!CAUTION]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn("Test content")
			if !containsString(result, tt.marker) {
				t.Errorf("Expected %s marker in output", tt.marker)
			}
			if !containsString(result, "> Test content") {
				t.Error("Expected content to be quoted")
			}
		})
	}
}

func TestAlertMultiline(t *testing.T) {
	result := Note("Line1\nLine2")
	expected := "> [!NOTE]\n> Line1\n> Line2\n"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

// ============================================================================
// SECTION TESTS
// ============================================================================

func TestCollapsible(t *testing.T) {
	result := Collapsible("FAQ", "Some answer")
	if !containsString(result, "<details>") {
		t.Error("Expected <details> tag")
	}
	if !containsString(result, "<summary>FAQ</summary>") {
		t.Error("Expected <summary> tag with title")
	}
	if !containsString(result, "Some answer") {
		t.Error("Expected content in output")
	}
}

func TestColumns(t *testing.T) {
	result := Columns("Col1", "Col2")
	if !containsString(result, "<gno-columns>") {
		t.Error("Expected <gno-columns> tag")
	}
	if !containsString(result, "|||") {
		t.Error("Expected column separator |||")
	}
}

func TestCenter(t *testing.T) {
	result := Center("Centered text")
	if !containsString(result, `align="center"`) {
		t.Error("Expected center alignment")
	}
}

// ============================================================================
// FLUENT CHAINING TEST
// ============================================================================

func TestFluentChaining(t *testing.T) {
	doc := New().
		H1("Welcome").
		P("Hello " + Bold("world") + "!").
		H2("Features").
		Bullet("Fast", "Simple", "Powerful").
		HR().
		Note("This is a note.").
		P("Footer")

	result := doc.String()

	checks := []string{
		"# Welcome",
		"**world**",
		"## Features",
		"- Fast",
		"---",
		"[!NOTE]",
		"Footer",
	}

	for _, check := range checks {
		if !containsString(result, check) {
			t.Errorf("Expected %q in output", check)
		}
	}
}

func TestComponentComposition(t *testing.T) {
	table := NewTable().
		Headers("Name", "Value").
		Row("Key", "123")

	doc := New().
		H1("Report").
		Add(table).
		P("End of report")

	result := doc.String()

	if !containsString(result, "# Report") {
		t.Error("Expected heading in output")
	}
	if !containsString(result, "| Name | Value |") {
		t.Error("Expected table in output")
	}
	if !containsString(result, "End of report") {
		t.Error("Expected footer in output")
	}
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
