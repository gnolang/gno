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

	// Explain how to make text bold
	sb.WriteString(md.H3(md.Bold("1. Bold Text")) + md.LineBreak(1))
	sb.WriteString(md.Paragraph("To make text bold, use the `md.Bold()` function:") + md.LineBreak(1))
	sb.WriteString(md.Bold("This is bold text") + md.LineBreak(2))

	// Explain how to make text italic
	sb.WriteString(md.H3(md.Bold("2. Italic Text")) + md.LineBreak(1))
	sb.WriteString(md.Paragraph("To make text italic, use the `md.Italic()` function:") + md.LineBreak(1))
	sb.WriteString(md.Italic("This is italic text") + md.LineBreak(2))

	// Explain how to add strikethrough
	sb.WriteString(md.H3(md.Bold("3. Strikethrough Text")) + md.LineBreak(1))
	sb.WriteString(md.Paragraph("To add strikethrough, use the `md.Strikethrough()` function:") + md.LineBreak(1))
	sb.WriteString(md.Strikethrough("This text is strikethrough") + md.LineBreak(2))

	// Explain headers (levels 1-6)
	sb.WriteString(md.H3(md.Bold("4. Headers (H1 to H6)")) + md.LineBreak(1))
	sb.WriteString(md.Paragraph("You can create headers (H1 to H6) using the `md.H1()` to `md.H6()` functions:") + md.LineBreak(1))
	sb.WriteString(md.H1("This is a level 1 header") + md.LineBreak(1))
	sb.WriteString(md.H2("This is a level 2 header") + md.LineBreak(1))
	sb.WriteString(md.H3("This is a level 3 header") + md.LineBreak(1))
	sb.WriteString(md.H4("This is a level 4 header") + md.LineBreak(1))
	sb.WriteString(md.H5("This is a level 5 header") + md.LineBreak(1))
	sb.WriteString(md.H6("This is a level 6 header") + md.LineBreak(2))

	// Explain bullet list
	sb.WriteString(md.H3(md.Bold("5. Bullet List")) + md.LineBreak(1))
	sb.WriteString(md.Paragraph("To create bullet lists, use the `md.BulletList()` function:") + md.LineBreak(1))
	sb.WriteString(md.BulletList([]string{"Item 1", "Item 2", "Item 3"}) + md.LineBreak(2))

	// Explain ordered list
	sb.WriteString(md.H3(md.Bold("6. Ordered List")) + md.LineBreak(1))
	sb.WriteString(md.Paragraph("To create ordered lists, use the `md.OrderedList()` function:") + md.LineBreak(1))
	sb.WriteString(md.OrderedList([]string{"First", "Second", "Third"}) + md.LineBreak(2))

	// Explain todo list
	sb.WriteString(md.H3(md.Bold("7. Todo List")) + md.LineBreak(1))
	sb.WriteString(md.Paragraph("You can create a todo list using the `md.TodoList()` function, which supports checkboxes:") + md.LineBreak(1))
	sb.WriteString(md.TodoList([]string{"Task 1", "Task 2"}, []bool{true, false}) + md.LineBreak(2))

	// Explain blockquote
	sb.WriteString(md.H3(md.Bold("8. Blockquote")) + md.LineBreak(1))
	sb.WriteString(md.Paragraph("To create blockquotes, use the `md.Blockquote()` function:") + md.LineBreak(1))
	sb.WriteString(md.Blockquote("This is a blockquote.\nIt can span multiple lines.") + md.LineBreak(2))

	// Explain inline code
	sb.WriteString(md.H3(md.Bold("9. Inline Code")) + md.LineBreak(1))
	sb.WriteString(md.Paragraph("To insert inline code, use the `md.InlineCode()` function:") + md.LineBreak(1))
	sb.WriteString(md.InlineCode("fmt.Println() // inline code") + md.LineBreak(2))

	// Explain code block
	sb.WriteString(md.H3(md.Bold("10. Code Block")) + md.LineBreak(1))
	sb.WriteString(md.Paragraph("For multi-line code blocks, use the `md.CodeBlock()` function:") + md.LineBreak(1))
	sb.WriteString(md.CodeBlock("package main\n\nfunc main() {\n\t// Your code here\n}") + md.LineBreak(2))

	// Explain horizontal rule
	sb.WriteString(md.H3(md.Bold("11. Horizontal Rule")) + md.LineBreak(1))
	sb.WriteString(md.Paragraph("To add a horizontal rule (separator), use the `md.HorizontalRule()` function:") + md.LineBreak(1))
	sb.WriteString(md.HorizontalRule() + md.LineBreak(2))

	// Explain language-specific code block
	sb.WriteString(md.H3(md.Bold("12. Language-specific Code Block")) + md.LineBreak(1))
	sb.WriteString(md.Paragraph("To create language-specific code blocks, use the `md.LanguageCodeBlock()` function:") + md.LineBreak(1))
	sb.WriteString(md.LanguageCodeBlock("go", "package main\n\nfunc main() {}") + md.LineBreak(2))

	// Explain link
	sb.WriteString(md.H3(md.Bold("13. Hyperlink")) + md.LineBreak(1))
	sb.WriteString(md.Paragraph("To create a hyperlink, use the `md.Link()` function:") + md.LineBreak(1))
	sb.WriteString(md.Link("Gnoland official docs", "https://docs.gno.land") + md.LineBreak(2))

	// Explain image
	sb.WriteString(md.H3(md.Bold("14. Image")) + md.LineBreak(1))
	sb.WriteString(md.Paragraph("To insert an image, use the `md.Image()` function:") + md.LineBreak(1))
	sb.WriteString(md.Image("Gnoland Logo", "https://gnolang.github.io/blog/2024-05-21_the-gnome/src/banner.png") + md.LineBreak(2))

	// Explain footnote
	sb.WriteString(md.H3(md.Bold("15. Footnote")) + md.LineBreak(1))
	sb.WriteString(md.Paragraph("To create footnotes, use the `md.Footnote()` function:") + md.LineBreak(1))
	sb.WriteString(md.Footnote("1", "This is a footnote.") + md.LineBreak(2))

	// Explain paragraph
	sb.WriteString(md.H3(md.Bold("16. Paragraph")) + md.LineBreak(1))
	sb.WriteString(md.Paragraph("To create a paragraph of text, use the `md.Paragraph()` function:") + md.LineBreak(1))
	sb.WriteString(md.Paragraph("This is a paragraph of text.") + md.LineBreak(2))

	// Explain table
	sb.WriteString(md.H3(md.Bold("17. Table")) + md.LineBreak(1))
	sb.WriteString(md.Paragraph("To create a table, use the `md.Table()` function. Here's an example of a table:") + md.LineBreak(1))

	// Create a table using the table package
	tb, _ := table.New([]string{"Feature", "Description"}, [][]string{
		{"Bold", "Make text bold using " + md.Bold("double asterisks")},
		{"Italic", "Make text italic using " + md.Italic("single asterisks")},
		{"Strikethrough", "Cross out text using " + md.Strikethrough("double tildes")},
	})

	sb.WriteString(md.Table(tb) + md.LineBreak(2))

	// Explain how to escape markdown
	sb.WriteString(md.H3(md.Bold("18. Escaping Markdown")) + md.LineBreak(1))
	sb.WriteString(md.Paragraph("Sometimes, you need to escape special Markdown characters (like *, _, and `). Use the `md.EscapeMarkdown()` function for this:") + md.LineBreak(1))

	// Example of escaping markdown
	text := "- Escape special chars like *, _, and ` in markdown"
	sb.WriteString(md.H4("Text Without Escape:") + md.LineBreak(1))
	sb.WriteString(text + md.LineBreak(2))

	sb.WriteString(md.H4("Text With Escape:") + md.LineBreak(1))
	sb.WriteString(md.EscapeMarkdown(text) + md.LineBreak(2))

	return sb.String()
}

