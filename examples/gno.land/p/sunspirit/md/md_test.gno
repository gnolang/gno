package md

import (
	"testing"

	"gno.land/p/nt/uassert/v0"
	"gno.land/p/sunspirit/table"
)

func TestNewBuilder(t *testing.T) {
	mdBuilder := NewBuilder()

	uassert.Equal(t, len(mdBuilder.elements), 0, "Expected 0 elements")
}

func TestAdd(t *testing.T) {
	mdBuilder := NewBuilder()

	header := H1("Hi")
	body := Paragraph("This is a test")

	mdBuilder.Add(header, body)

	uassert.Equal(t, len(mdBuilder.elements), 2, "Expected 2 element")
	uassert.Equal(t, mdBuilder.elements[0], header, "Expected element %s, got %s", header, mdBuilder.elements[0])
	uassert.Equal(t, mdBuilder.elements[1], body, "Expected element %s, got %s", body, mdBuilder.elements[1])
}

func TestRender(t *testing.T) {
	mdBuilder := NewBuilder()

	header := H1("Hello")
	body := Paragraph("This is a test")

	seperator := "\n"
	expected := header + seperator + body

	output := mdBuilder.Add(header, body).Render(seperator)

	uassert.Equal(t, output, expected, "Expected rendered string %s, got %s", expected, output)
}

func Test_Bold(t *testing.T) {
	uassert.Equal(t, Bold("Hello"), "**Hello**")
}

func Test_Italic(t *testing.T) {
	uassert.Equal(t, Italic("Hello"), "*Hello*")
}

func Test_Strikethrough(t *testing.T) {
	uassert.Equal(t, Strikethrough("Hello"), "~~Hello~~")
}

func Test_H1(t *testing.T) {
	uassert.Equal(t, H1("Header 1"), "# Header 1\n")
}

func Test_H2(t *testing.T) {
	uassert.Equal(t, H2("Header 2"), "## Header 2\n")
}

func Test_H3(t *testing.T) {
	uassert.Equal(t, H3("Header 3"), "### Header 3\n")
}

func Test_H4(t *testing.T) {
	uassert.Equal(t, H4("Header 4"), "#### Header 4\n")
}

func Test_H5(t *testing.T) {
	uassert.Equal(t, H5("Header 5"), "##### Header 5\n")
}

func Test_H6(t *testing.T) {
	uassert.Equal(t, H6("Header 6"), "###### Header 6\n")
}

func Test_BulletList(t *testing.T) {
	items := []string{"Item 1", "Item 2", "Item 3"}
	result := BulletList(items)
	expected := "- Item 1\n- Item 2\n- Item 3\n"
	uassert.Equal(t, result, expected)
}

func Test_OrderedList(t *testing.T) {
	items := []string{"Item 1", "Item 2", "Item 3"}
	result := OrderedList(items)
	expected := "1. Item 1\n2. Item 2\n3. Item 3\n"
	uassert.Equal(t, result, expected)
}

func Test_TodoList(t *testing.T) {
	items := []string{"Task 1", "Task 2"}
	done := []bool{true, false}
	result := TodoList(items, done)
	expected := "- [x] Task 1\n- [ ] Task 2\n"
	uassert.Equal(t, result, expected)
}

func Test_Blockquote(t *testing.T) {
	text := "This is a blockquote.\nIt has multiple lines."
	result := Blockquote(text)
	expected := "> This is a blockquote.\n> It has multiple lines.\n"
	uassert.Equal(t, result, expected)
}

func Test_InlineCode(t *testing.T) {
	result := InlineCode("code")
	uassert.Equal(t, result, "`code`")
}

func Test_LanguageCodeBlock(t *testing.T) {
	result := LanguageCodeBlock("python", "print('Hello')")
	expected := "```python\nprint('Hello')\n```"
	uassert.Equal(t, result, expected)
}

func Test_CodeBlock(t *testing.T) {
	result := CodeBlock("print('Hello')")
	expected := "```\nprint('Hello')\n```"
	uassert.Equal(t, result, expected)
}

func Test_LineBreak(t *testing.T) {
	result := LineBreak(2)
	expected := "\n\n\n"
	uassert.Equal(t, result, expected)

	result = LineBreak(0)
	expected = ""
	uassert.Equal(t, result, expected)
}

func Test_HorizontalRule(t *testing.T) {
	result := HorizontalRule()
	uassert.Equal(t, result, "---\n")
}

func Test_Link(t *testing.T) {
	result := Link("Google", "http://google.com")
	uassert.Equal(t, result, "[Google](http://google.com)")
}

func Test_Image(t *testing.T) {
	result := Image("Alt text", "http://image.url")
	uassert.Equal(t, result, "![Alt text](http://image.url)")
}

func Test_Footnote(t *testing.T) {
	result := Footnote("1", "This is a footnote.")
	uassert.Equal(t, result, "[1]: This is a footnote.")
}

func Test_Paragraph(t *testing.T) {
	result := Paragraph("This is a paragraph.")
	uassert.Equal(t, result, "This is a paragraph.\n")
}

func Test_Table(t *testing.T) {
	tb, err := table.New([]string{"Header1", "Header2"}, [][]string{
		{"Row1Col1", "Row1Col2"},
		{"Row2Col1", "Row2Col2"},
	})
	uassert.NoError(t, err)

	result := Table(tb)
	expected := "| Header1 | Header2 |\n| ---|---|\n| Row1Col1 | Row1Col2 |\n| Row2Col1 | Row2Col2 |\n"
	uassert.Equal(t, result, expected)
}

func Test_EscapeMarkdown(t *testing.T) {
	result := EscapeMarkdown("- This is `code`")
	uassert.Equal(t, result, "``- This is `code```")
}
