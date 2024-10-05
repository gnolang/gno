package markdown

import (
	"gno.land/p/demo/ufmt"
	"strings"
	"errors"
)

// MarkdownBuilder is a structure for building markdown text
type MarkdownBuilder struct {
	body []string
}

// NewBuilder returns a new instance of MarkdownBuilder
func NewBuilder() *MarkdownBuilder {
	return &MarkdownBuilder{
		body: make([]string, 0),
	}
}

// String returns the complete markdown text as a single string
func (m *MarkdownBuilder) String() string {
	return strings.Join(m.body, "\n")
}

// PlainText adds plain text to the markdown
func (m *MarkdownBuilder) PlainText(text string) *MarkdownBuilder {
	m.body = append(m.body, text)
	return m
}

// Bold adds bold text to the markdown
func (m *MarkdownBuilder) Bold(text string) *MarkdownBuilder {
	m.body = append(m.body, ufmt.Sprintf("**%s**", text))
	return m
}

// Italic adds italicized text to the markdown
func (m *MarkdownBuilder) Italic(text string) *MarkdownBuilder {
	m.body = append(m.body, ufmt.Sprintf("*%s*", text))
	return m
}

// H1 adds a level 1 header to the markdown
func (m *MarkdownBuilder) H1(text string) *MarkdownBuilder {
	m.body = append(m.body, ufmt.Sprintf("# %s", text))
	return m
}

// H2 adds a level 2 header to the markdown
func (m *MarkdownBuilder) H2(text string) *MarkdownBuilder {
	m.body = append(m.body, ufmt.Sprintf("## %s", text))
	return m
}

// H3 adds a level 3 header to the markdown
func (m *MarkdownBuilder) H3(text string) *MarkdownBuilder {
	m.body = append(m.body, ufmt.Sprintf("### %s", text))
	return m
}

// UnorderedList adds an unordered list to the markdown
func (m *MarkdownBuilder) UnorderedList(items []string) *MarkdownBuilder {
	for _, item := range items {
		m.body = append(m.body, ufmt.Sprintf("- %s", item))
	}
	return m
}

// OrderedList adds an ordered list to the markdown
func (m *MarkdownBuilder) OrderedList(items []string) *MarkdownBuilder {
	for i, item := range items {
		m.body = append(m.body, ufmt.Sprintf("%d. %s", i+1, item))
	}
	return m
}

// TodoItem represents a single todo item with its completion status
type TodoItem struct {
	Item string 
	Done bool  
}

// TodoItem adds a single todo item with a checkbox to the markdown
func (m *MarkdownBuilder) TodoItem(item string, done bool) *MarkdownBuilder {
	if done {
		m.body = append(m.body, ufmt.Sprintf("- [x] %s", item))
	} else {
		m.body = append(m.body, ufmt.Sprintf("- [ ] %s", item))
	}
	return m
}

// TodoList adds a list of todo items with checkboxes to the markdown
func (m *MarkdownBuilder) TodoList(items []TodoItem) *MarkdownBuilder {
	for _, item := range items {
		m.TodoItem(item.Item, item.Done)
	}
	return m
}

// Blockquote adds a blockquote to the markdown
func (m *MarkdownBuilder) Blockquote(text string) *MarkdownBuilder {
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		m.body = append(m.body, ufmt.Sprintf("> %s", line))
	}
	return m
}

// InlineCode adds inline code to the markdown document
func (m *MarkdownBuilder) InlineCode(code string) *MarkdownBuilder {
	m.body = append(m.body, ufmt.Sprintf("`%s`", code))
	return m
}

// CodeBlock adds a code block to the markdown document, optionally specifying the language
func (m *MarkdownBuilder) CodeBlock(code string, language string) *MarkdownBuilder {
	if language != "" {
		m.body = append(m.body, ufmt.Sprintf("```%s\n%s\n```", language, code))
	} else {
		m.body = append(m.body, ufmt.Sprintf("```\n%s\n```", code))
	}
	return m
}

// LineBreak adds a specified number of line breaks to the markdown document
func (m *MarkdownBuilder) LineBreak(count uint) *MarkdownBuilder {
	if count > 0 {
		lineBreaks := strings.Repeat("\n", int(count))
		m.body = append(m.body, lineBreaks) 
	}
	return m
}

// HorizontalRule adds a horizontal rule to the markdown document
func (m *MarkdownBuilder) HorizontalRule() *MarkdownBuilder {
	m.body = append(m.body, "---")
	return m
}

// Link adds a hyperlink to the markdown document
func (m *MarkdownBuilder) Link(text, url string) *MarkdownBuilder {
	m.body = append(m.body, ufmt.Sprintf("[%s](%s)", text, url))
	return m
}

// Image adds an image to the markdown document
func (m *MarkdownBuilder) Image(altText, url string) *MarkdownBuilder {
	m.body = append(m.body, ufmt.Sprintf("![%s](%s)", altText, url))
	return m
}

// Footnote adds a footnote to the markdown document
func (m *MarkdownBuilder) Footnote(reference, text string) *MarkdownBuilder {
	m.body = append(m.body, ufmt.Sprintf("[%s]: %s", reference, text))
	return m
}

// TableSet defines the structure for a markdown table
type TableSet struct {
	Header []string
	Rows   [][]string // Using a slice of slices to represent rows
}

// ValidateColumns checks if the number of columns in each row matches the number of columns in the header
func (t *TableSet) ValidateColumns() error {
	numCols := len(t.Header)
	for _, row := range t.Rows {
		if len(row) != numCols {
			return ufmt.Errorf("row %v does not match header length %d", row, numCols)
		}
	}
	return nil
}

// NewTableSet creates a new TableSet instance, ensuring the header and rows match in size
func NewTableSet(header []string, rows [][]string) (*TableSet, error) {
	t := &TableSet{
		Header: header,
		Rows:   rows,
	}
	if err := t.ValidateColumns(); err != nil {
		return nil, err
	}
	return t, nil
}

// AddTable adds a markdown table to the document using the specified TableSet
func (m *MarkdownBuilder) AddTable(t *TableSet) *MarkdownBuilder {
	var builder strings.Builder

	builder.WriteString("| " + strings.Join(t.Header, " | ") + " |\n")
	builder.WriteString("| " + strings.Repeat("---|", len(t.Header)) + "\n")

	for _, row := range t.Rows {
		builder.WriteString("| " + strings.Join(row, " | ") + " |\n")
	}

	m.body = append(m.body, builder.String())
	return m
}
