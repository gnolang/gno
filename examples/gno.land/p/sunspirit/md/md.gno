package md

import (
	"strings"

	"gno.land/p/nt/ufmt"
)

// Builder helps to build a Markdown string from individual elements
type Builder struct {
	elements []string
}

// NewBuilder creates a new Builder instance
func NewBuilder() *Builder {
	return &Builder{}
}

// Add adds a Markdown element to the builder
func (m *Builder) Add(md ...string) *Builder {
	m.elements = append(m.elements, md...)
	return m
}

// Render returns the final Markdown string joined with the specified separator
func (m *Builder) Render(separator string) string {
	return strings.Join(m.elements, separator)
}

// Bold returns bold text for markdown
func Bold(text string) string {
	return ufmt.Sprintf("**%s**", text)
}

// Italic returns italicized text for markdown
func Italic(text string) string {
	return ufmt.Sprintf("*%s*", text)
}

// Strikethrough returns strikethrough text for markdown
func Strikethrough(text string) string {
	return ufmt.Sprintf("~~%s~~", text)
}

// H1 returns a level 1 header for markdown
func H1(text string) string {
	return ufmt.Sprintf("# %s\n", text)
}

// H2 returns a level 2 header for markdown
func H2(text string) string {
	return ufmt.Sprintf("## %s\n", text)
}

// H3 returns a level 3 header for markdown
func H3(text string) string {
	return ufmt.Sprintf("### %s\n", text)
}

// H4 returns a level 4 header for markdown
func H4(text string) string {
	return ufmt.Sprintf("#### %s\n", text)
}

// H5 returns a level 5 header for markdown
func H5(text string) string {
	return ufmt.Sprintf("##### %s\n", text)
}

// H6 returns a level 6 header for markdown
func H6(text string) string {
	return ufmt.Sprintf("###### %s\n", text)
}

// BulletList returns an bullet list for markdown
func BulletList(items []string) string {
	var sb strings.Builder
	for _, item := range items {
		sb.WriteString(ufmt.Sprintf("- %s\n", item))
	}
	return sb.String()
}

// OrderedList returns an ordered list for markdown
func OrderedList(items []string) string {
	var sb strings.Builder
	for i, item := range items {
		sb.WriteString(ufmt.Sprintf("%d. %s\n", i+1, item))
	}
	return sb.String()
}

// TodoList returns a list of todo items with checkboxes for markdown
func TodoList(items []string, done []bool) string {
	var sb strings.Builder

	for i, item := range items {
		checkbox := " "
		if done[i] {
			checkbox = "x"
		}
		sb.WriteString(ufmt.Sprintf("- [%s] %s\n", checkbox, item))
	}
	return sb.String()
}

// Blockquote returns a blockquote for markdown
func Blockquote(text string) string {
	lines := strings.Split(text, "\n")
	var sb strings.Builder
	for _, line := range lines {
		sb.WriteString(ufmt.Sprintf("> %s\n", line))
	}

	return sb.String()
}

// InlineCode returns inline code for markdown
func InlineCode(code string) string {
	return ufmt.Sprintf("`%s`", code)
}

// CodeBlock creates a markdown code block
func CodeBlock(content string) string {
	return ufmt.Sprintf("```\n%s\n```", content)
}

// LanguageCodeBlock creates a markdown code block with language-specific syntax highlighting
func LanguageCodeBlock(language, content string) string {
	return ufmt.Sprintf("```%s\n%s\n```", language, content)
}

// LineBreak returns the specified number of line breaks for markdown
func LineBreak(count uint) string {
	if count > 0 {
		return strings.Repeat("\n", int(count)+1)
	}
	return ""
}

// HorizontalRule returns a horizontal rule for markdown
func HorizontalRule() string {
	return "---\n"
}

// Link returns a hyperlink for markdown
func Link(text, url string) string {
	return ufmt.Sprintf("[%s](%s)", text, url)
}

// Image returns an image for markdown
func Image(altText, url string) string {
	return ufmt.Sprintf("![%s](%s)", altText, url)
}

// Footnote returns a footnote for markdown
func Footnote(reference, text string) string {
	return ufmt.Sprintf("[%s]: %s", reference, text)
}

// Paragraph wraps the given text in a Markdown paragraph
func Paragraph(content string) string {
	return ufmt.Sprintf("%s\n", content)
}

// MdTable is an interface for table types that can be converted to Markdown format
type MdTable interface {
	String() string
}

// Table takes any MdTable implementation and returns its markdown representation
func Table(table MdTable) string {
	return table.String()
}

// EscapeMarkdown escapes special markdown characters in a string
func EscapeMarkdown(text string) string {
	return ufmt.Sprintf("``%s``", text)
}
