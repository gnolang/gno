package markdown

import (
	"testing"
	"gno.land/p/demo/uassert"
)

func TestNewMarkdown(t *testing.T) {
	m := NewBuilder()
	expected := ""
	uassert.Equal(t, m.String(), expected)
}

func TestPlainText(t *testing.T) {
	m := NewBuilder().PlainText("Hello, World!")
	expected := "Hello, World!"
	uassert.Equal(t, m.String(), expected)
}

func TestBold(t *testing.T) {
	m := NewBuilder().Bold("Bold Text")
	expected := "**Bold Text**"
	uassert.Equal(t, m.String(), expected)
}

func TestItalic(t *testing.T) {
	m := NewBuilder().Italic("Italic Text")
	expected := "*Italic Text*"
	uassert.Equal(t, m.String(), expected)
}

func TestH1(t *testing.T) {
	m := NewBuilder().H1("Header 1")
	expected := "# Header 1"
	uassert.Equal(t, m.String(), expected)
}

func TestH2(t *testing.T) {
	m := NewBuilder().H2("Header 2")
	expected := "## Header 2"
	uassert.Equal(t, m.String(), expected)
}

func TestUnorderedList(t *testing.T) {
	m := NewBuilder().UnorderedList([]string{"Item 1", "Item 2"})
	expected := "- Item 1\n- Item 2"
	uassert.Equal(t, m.String(), expected)
}

func TestOrderedList(t *testing.T) {
	m := NewBuilder().OrderedList([]string{"First", "Second"})
	expected := "1. First\n2. Second"
	uassert.Equal(t, m.String(), expected)
}

func TestTodoItem(t *testing.T) {
	m := NewBuilder().TodoItem("Buy milk", true)
	expected := "- [x] Buy milk"
	uassert.Equal(t, m.String(), expected)
}

func TestTodoList(t *testing.T) {
	m := NewBuilder().TodoList([]TodoItem{
		{Item: "Task 1", Done: true},
		{Item: "Task 2", Done: false},
	})
	expected := "- [x] Task 1\n- [ ] Task 2"
	uassert.Equal(t, m.String(), expected)
}

func TestBlockquote(t *testing.T) {
	m := NewBuilder().Blockquote("This is a quote.")
	expected := "> This is a quote."
	uassert.Equal(t, m.String(), expected)
}

func TestInlineCode(t *testing.T) {
	m := NewBuilder().InlineCode("fmt.Println()")
	expected := "`fmt.Println()`"
	uassert.Equal(t, m.String(), expected)
}

func TestCodeBlock(t *testing.T) {
	m := NewBuilder().CodeBlock("package main\n\nfunc main() {}", "go")
	expected := "```go\npackage main\n\nfunc main() {}\n```"
	uassert.Equal(t, m.String(), expected)
}

func TestLineBreak(t *testing.T) {
	m := NewBuilder().PlainText("First line").LineBreak(1).PlainText("Second line")
	expected := "First line\n\n\nSecond line"
	uassert.Equal(t, m.String(), expected)
}

func TestHorizontalRule(t *testing.T) {
	m := NewBuilder().HorizontalRule()
	expected := "---"
	uassert.Equal(t, m.String(), expected)
}

func TestLink(t *testing.T) {
	m := NewBuilder().Link("Google", "https://google.com")
	expected := "[Google](https://google.com)"
	uassert.Equal(t, m.String(), expected)
}

func TestImage(t *testing.T) {
	m := NewBuilder().Image("Alt text", "http://example.com/image.png")
	expected := "![Alt text](http://example.com/image.png)"
	uassert.Equal(t, m.String(), expected)
}

func TestFootnote(t *testing.T) {
	m := NewBuilder().Footnote("1", "This is a footnote")
	expected := "[1]: This is a footnote"
	uassert.Equal(t, m.String(), expected)
}

func TestTable(t *testing.T) {
	m := NewBuilder()
	
	table, err := NewTableSet(
		[]string{"Header1", "Header2"},
		[][]string{
			{"Row1Col1", "Row1Col2"},
			{"Row2Col1", "Row2Col2"},
		},
	)
	if err != nil {
		t.Fatalf("Failed to create table set: %v", err)
	}

	m.AddTable(table)
	
	expected := "| Header1 | Header2 |\n" +
		"| ---|---|\n" +
		"| Row1Col1 | Row1Col2 |\n" +
		"| Row2Col1 | Row2Col2 |\n"
	
	uassert.Equal(t, m.String(), expected)
}


