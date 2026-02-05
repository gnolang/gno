// Package md provides a fluent API for generating Markdown in Gno.
//
// The package offers two approaches for building Markdown:
//
// 1. Builder pattern with method chaining for complex documents:
//
//	doc := md.New().
//		H1("Welcome").
//		P("This is a paragraph.").
//		Bullet("Item 1", "Item 2", "Item 3")
//	output := doc.String()
//
// 2. Standalone functions for simple composition:
//
//	output := md.H1("Title") + md.P("Some text with " + md.Bold("emphasis"))
//
// Both approaches can be mixed as needed. The package supports all common
// Markdown elements: headers, paragraphs, lists (bullet and numbered),
// tables, code blocks, blockquotes, horizontal rules, links, images,
// and inline formatting (bold, italic, code, strikethrough).
package md

import (
	"strconv"
	"strings"
)

// Document is a Markdown document builder that supports method chaining.
type Document struct {
	buf strings.Builder
}

// New creates a new empty Markdown document.
func New() *Document {
	return &Document{}
}

// String returns the complete Markdown document as a string.
func (d *Document) String() string {
	return d.buf.String()
}

// Write appends raw text to the document without any formatting.
func (d *Document) Write(s string) *Document {
	d.buf.WriteString(s)
	return d
}

// Writeln appends raw text followed by a newline.
func (d *Document) Writeln(s string) *Document {
	d.buf.WriteString(s)
	d.buf.WriteString("\n")
	return d
}

// LF adds a single newline (line feed) to the document.
func (d *Document) LF() *Document {
	d.buf.WriteString("\n")
	return d
}

// BR adds a Markdown line break (two spaces followed by newline).
func (d *Document) BR() *Document {
	d.buf.WriteString("  \n")
	return d
}

// -----------------------------------------------------------------------------
// Headers
// -----------------------------------------------------------------------------

// H1 appends a level-1 header.
func (d *Document) H1(text string) *Document {
	d.buf.WriteString("# ")
	d.buf.WriteString(text)
	d.buf.WriteString("\n\n")
	return d
}

// H2 appends a level-2 header.
func (d *Document) H2(text string) *Document {
	d.buf.WriteString("## ")
	d.buf.WriteString(text)
	d.buf.WriteString("\n\n")
	return d
}

// H3 appends a level-3 header.
func (d *Document) H3(text string) *Document {
	d.buf.WriteString("### ")
	d.buf.WriteString(text)
	d.buf.WriteString("\n\n")
	return d
}

// H4 appends a level-4 header.
func (d *Document) H4(text string) *Document {
	d.buf.WriteString("#### ")
	d.buf.WriteString(text)
	d.buf.WriteString("\n\n")
	return d
}

// H5 appends a level-5 header.
func (d *Document) H5(text string) *Document {
	d.buf.WriteString("##### ")
	d.buf.WriteString(text)
	d.buf.WriteString("\n\n")
	return d
}

// H6 appends a level-6 header.
func (d *Document) H6(text string) *Document {
	d.buf.WriteString("###### ")
	d.buf.WriteString(text)
	d.buf.WriteString("\n\n")
	return d
}

// -----------------------------------------------------------------------------
// Paragraphs and text blocks
// -----------------------------------------------------------------------------

// P appends a paragraph with proper spacing.
func (d *Document) P(text string) *Document {
	d.buf.WriteString(text)
	d.buf.WriteString("\n\n")
	return d
}

// Quote appends a blockquote.
func (d *Document) Quote(text string) *Document {
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		d.buf.WriteString("> ")
		d.buf.WriteString(line)
		d.buf.WriteString("\n")
	}
	d.buf.WriteString("\n")
	return d
}

// HR appends a horizontal rule.
func (d *Document) HR() *Document {
	d.buf.WriteString("---\n\n")
	return d
}

// -----------------------------------------------------------------------------
// Code
// -----------------------------------------------------------------------------

// CodeBlock appends a fenced code block with optional language specification.
func (d *Document) CodeBlock(code string, lang string) *Document {
	d.buf.WriteString("```")
	d.buf.WriteString(lang)
	d.buf.WriteString("\n")
	d.buf.WriteString(code)
	if !strings.HasSuffix(code, "\n") {
		d.buf.WriteString("\n")
	}
	d.buf.WriteString("```\n\n")
	return d
}

// -----------------------------------------------------------------------------
// Lists
// -----------------------------------------------------------------------------

// Bullet appends an unordered (bullet) list.
func (d *Document) Bullet(items ...string) *Document {
	for _, item := range items {
		d.buf.WriteString("- ")
		d.buf.WriteString(item)
		d.buf.WriteString("\n")
	}
	d.buf.WriteString("\n")
	return d
}

// Numbered appends an ordered (numbered) list.
func (d *Document) Numbered(items ...string) *Document {
	for i, item := range items {
		d.buf.WriteString(strconv.Itoa(i + 1))
		d.buf.WriteString(". ")
		d.buf.WriteString(item)
		d.buf.WriteString("\n")
	}
	d.buf.WriteString("\n")
	return d
}

// Checklist appends a task list with checkboxes.
// The checked parameter indicates which items are checked (by index).
func (d *Document) Checklist(items []string, checked []bool) *Document {
	for i, item := range items {
		if i < len(checked) && checked[i] {
			d.buf.WriteString("- [x] ")
		} else {
			d.buf.WriteString("- [ ] ")
		}
		d.buf.WriteString(item)
		d.buf.WriteString("\n")
	}
	d.buf.WriteString("\n")
	return d
}

// -----------------------------------------------------------------------------
// Tables
// -----------------------------------------------------------------------------

// Table appends a Markdown table.
// Headers define the column names.
// Rows is a slice of rows, where each row is a slice of cell values.
func (d *Document) Table(headers []string, rows [][]string) *Document {
	if len(headers) == 0 {
		return d
	}

	// Calculate column widths for pretty formatting
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Header row
	d.buf.WriteString("|")
	for i, h := range headers {
		d.buf.WriteString(" ")
		d.buf.WriteString(h)
		d.buf.WriteString(strings.Repeat(" ", widths[i]-len(h)))
		d.buf.WriteString(" |")
	}
	d.buf.WriteString("\n")

	// Separator row
	d.buf.WriteString("|")
	for _, w := range widths {
		d.buf.WriteString(" ")
		d.buf.WriteString(strings.Repeat("-", w))
		d.buf.WriteString(" |")
	}
	d.buf.WriteString("\n")

	// Data rows
	for _, row := range rows {
		d.buf.WriteString("|")
		for i := 0; i < len(headers); i++ {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			d.buf.WriteString(" ")
			d.buf.WriteString(cell)
			d.buf.WriteString(strings.Repeat(" ", widths[i]-len(cell)))
			d.buf.WriteString(" |")
		}
		d.buf.WriteString("\n")
	}
	d.buf.WriteString("\n")

	return d
}

// -----------------------------------------------------------------------------
// Standalone element functions (return strings)
// These can be used for composition or with ufmt.Sprintf
// -----------------------------------------------------------------------------

// H1 returns a level-1 header string.
func H1(text string) string {
	return "# " + text + "\n\n"
}

// H2 returns a level-2 header string.
func H2(text string) string {
	return "## " + text + "\n\n"
}

// H3 returns a level-3 header string.
func H3(text string) string {
	return "### " + text + "\n\n"
}

// H4 returns a level-4 header string.
func H4(text string) string {
	return "#### " + text + "\n\n"
}

// H5 returns a level-5 header string.
func H5(text string) string {
	return "##### " + text + "\n\n"
}

// H6 returns a level-6 header string.
func H6(text string) string {
	return "###### " + text + "\n\n"
}

// P returns a paragraph string.
func P(text string) string {
	return text + "\n\n"
}

// HR returns a horizontal rule string.
func HR() string {
	return "---\n\n"
}

// Quote returns a blockquote string.
func Quote(text string) string {
	lines := strings.Split(text, "\n")
	result := ""
	for _, line := range lines {
		result += "> " + line + "\n"
	}
	return result + "\n"
}

// CodeBlock returns a fenced code block string.
func CodeBlock(code, lang string) string {
	suffix := "\n"
	if strings.HasSuffix(code, "\n") {
		suffix = ""
	}
	return "```" + lang + "\n" + code + suffix + "```\n\n"
}

// Bullet returns an unordered list string.
func Bullet(items ...string) string {
	result := ""
	for _, item := range items {
		result += "- " + item + "\n"
	}
	return result + "\n"
}

// Numbered returns an ordered list string.
func Numbered(items ...string) string {
	result := ""
	for i, item := range items {
		result += strconv.Itoa(i+1) + ". " + item + "\n"
	}
	return result + "\n"
}

// -----------------------------------------------------------------------------
// Inline formatting functions
// These return formatted strings without trailing newlines
// -----------------------------------------------------------------------------

// Bold returns text wrapped in bold markers.
func Bold(text string) string {
	return "**" + text + "**"
}

// Italic returns text wrapped in italic markers.
func Italic(text string) string {
	return "_" + text + "_"
}

// BoldItalic returns text wrapped in bold and italic markers.
func BoldItalic(text string) string {
	return "***" + text + "***"
}

// InlineCode returns text wrapped in inline code markers.
func InlineCode(text string) string {
	return "`" + text + "`"
}

// Strikethrough returns text wrapped in strikethrough markers.
func Strikethrough(text string) string {
	return "~~" + text + "~~"
}

// Link returns a Markdown link.
func Link(text, url string) string {
	return "[" + text + "](" + url + ")"
}

// Image returns a Markdown image.
func Image(alt, url string) string {
	return "![" + alt + "](" + url + ")"
}

// LinkWithTitle returns a Markdown link with a title attribute.
func LinkWithTitle(text, url, title string) string {
	return "[" + text + "](" + url + " \"" + title + "\")"
}

// ImageWithTitle returns a Markdown image with a title attribute.
func ImageWithTitle(alt, url, title string) string {
	return "![" + alt + "](" + url + " \"" + title + "\")"
}

// -----------------------------------------------------------------------------
// Table helper
// -----------------------------------------------------------------------------

// Table returns a Markdown table string.
func Table(headers []string, rows [][]string) string {
	if len(headers) == 0 {
		return ""
	}

	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	result := "|"
	for i, h := range headers {
		result += " " + h + strings.Repeat(" ", widths[i]-len(h)) + " |"
	}
	result += "\n|"
	for _, w := range widths {
		result += " " + strings.Repeat("-", w) + " |"
	}
	result += "\n"

	for _, row := range rows {
		result += "|"
		for i := 0; i < len(headers); i++ {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			result += " " + cell + strings.Repeat(" ", widths[i]-len(cell)) + " |"
		}
		result += "\n"
	}

	return result + "\n"
}
