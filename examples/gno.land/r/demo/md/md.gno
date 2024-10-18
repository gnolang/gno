package md

import (
	"strings"

	"gno.land/p/demo/md"
	"gno.land/p/demo/table"
)

func Render(path string) string {
	title := "Exploring markdown functions in the gno.land/p/md library"

	var sb strings.Builder
	sb.WriteString(md.H1(md.Italic(md.Bold(title))) + md.LineBreak(1))

	// Bold
	sb.WriteString(md.Bold("This is bold text") + md.LineBreak(1))

	// Italic
	sb.WriteString(md.Italic("This is italic text") + md.LineBreak(1))

	// Strikethrough
	sb.WriteString(md.Strikethrough("This text is strikethrough") + md.LineBreak(1))

	// Headers
	sb.WriteString(md.H1("This is a level 1 header") + md.LineBreak(1))
	sb.WriteString(md.H2("This is a level 2 header") + md.LineBreak(1))
	sb.WriteString(md.H3("This is a level 3 header") + md.LineBreak(1))
	sb.WriteString(md.H4("This is a level 4 header") + md.LineBreak(1))
	sb.WriteString(md.H5("This is a level 5 header") + md.LineBreak(1))
	sb.WriteString(md.H6("This is a level 6 header") + md.LineBreak(1))

	// Bullet List
	sb.WriteString(md.BulletList([]string{"Item 1", "Item 2", "Item 3"}) + md.LineBreak(1))

	// Ordered List
	sb.WriteString(md.OrderedList([]string{"First", "Second", "Third"}) + md.LineBreak(1))

	// Todo List
	sb.WriteString(md.TodoList([]string{"Task 1", "Task 2"}, []bool{true, false}) + md.LineBreak(1))

	// Blockquote
	sb.WriteString(md.Blockquote("This is a blockquote.\nIt can span multiple lines.") + md.LineBreak(1))

	// Inline Code
	sb.WriteString(md.InlineCode("fmt.Println() // inline code") + md.LineBreak(1))

	// Code Block
	sb.WriteString(md.CodeBlock("package main\n\nfunc main() {\n\t// Your code here\n}") + md.LineBreak(1))

	// Horizontal Rule
	sb.WriteString(md.HorizontalRule())

	// Language Code Block
	sb.WriteString(md.LanguageCodeBlock("go", "package main\n\nfunc main() {}") + md.LineBreak(1))

	// Link
	sb.WriteString(md.Link("Gnoland official docs", "https://docs.gno.land") + md.LineBreak(1))

	// Image
	sb.WriteString(md.Image("Gnoland Logo", "https://gnolang.github.io/blog/2024-05-21_the-gnome/src/banner.png") + md.LineBreak(1))

	// Footnote
	sb.WriteString(md.Footnote("1", "This is a footnote.") + md.LineBreak(1))

	// Paragraph
	sb.WriteString(md.Paragraph("This is a paragraph of text.") + md.LineBreak(1))

	// Table
	tb, _ := table.New([]string{"Feature", "Description"}, [][]string{
		{"Bold", "Make text bold using " + md.Bold("double asterisks")},
		{"Italic", "Make text italic using " + md.Italic("single asterisks")},
		{"Strikethrough", "Cross out text using " + md.Strikethrough("double tildes")},
	})

	sb.WriteString(md.Table(tb) + md.LineBreak(1))

	// EscapeMarkdown
	text := "- Escape special chars like *, _, and ` in markdown"

	// Show non-escaped version
	sb.WriteString(md.H3("Text Without Escape:") + md.LineBreak(1))
	sb.WriteString(text + md.LineBreak(2))

	// Show escaped version
	sb.WriteString(md.H3("Text With Escape:") + md.LineBreak(1))
	sb.WriteString(md.EscapeMarkdown(text) + md.LineBreak(2))

	return sb.String()
}
