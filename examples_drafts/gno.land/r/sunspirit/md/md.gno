package md

import (
	"gno.land/p/sunspirit/md"
	"gno.land/p/sunspirit/table"
)

func Render(path string) string {
	title := "A simple, flexible, and easy-to-use library for creating markdown documents in gno.land"

	mdBuilder := md.NewBuilder().
		Add(md.H1(md.Italic(md.Bold(title)))).

		// Bold Text section
		Add(
			md.H3(md.Bold("1. Bold Text")),
			md.Paragraph("To make text bold, use the `md.Bold()` function:"),
			md.Bold("This is bold text"),
		).

		// Italic Text section
		Add(
			md.H3(md.Bold("2. Italic Text")),
			md.Paragraph("To make text italic, use the `md.Italic()` function:"),
			md.Italic("This is italic text"),
		).

		// Strikethrough Text section
		Add(
			md.H3(md.Bold("3. Strikethrough Text")),
			md.Paragraph("To add strikethrough, use the `md.Strikethrough()` function:"),
			md.Strikethrough("This text is strikethrough"),
		).

		// Headers section
		Add(
			md.H3(md.Bold("4. Headers (H1 to H6)")),
			md.Paragraph("You can create headers (H1 to H6) using the `md.H1()` to `md.H6()` functions:"),
			md.H1("This is a level 1 header"),
			md.H2("This is a level 2 header"),
			md.H3("This is a level 3 header"),
			md.H4("This is a level 4 header"),
			md.H5("This is a level 5 header"),
			md.H6("This is a level 6 header"),
		).

		// Bullet List section
		Add(
			md.H3(md.Bold("5. Bullet List")),
			md.Paragraph("To create bullet lists, use the `md.BulletList()` function:"),
			md.BulletList([]string{"Item 1", "Item 2", "Item 3"}),
		).

		// Ordered List section
		Add(
			md.H3(md.Bold("6. Ordered List")),
			md.Paragraph("To create ordered lists, use the `md.OrderedList()` function:"),
			md.OrderedList([]string{"First", "Second", "Third"}),
		).

		// Todo List section
		Add(
			md.H3(md.Bold("7. Todo List")),
			md.Paragraph("You can create a todo list using the `md.TodoList()` function, which supports checkboxes:"),
			md.TodoList([]string{"Task 1", "Task 2"}, []bool{true, false}),
		).

		// Blockquote section
		Add(
			md.H3(md.Bold("8. Blockquote")),
			md.Paragraph("To create blockquotes, use the `md.Blockquote()` function:"),
			md.Blockquote("This is a blockquote.\nIt can span multiple lines."),
		).

		// Inline Code section
		Add(
			md.H3(md.Bold("9. Inline Code")),
			md.Paragraph("To insert inline code, use the `md.InlineCode()` function:"),
			md.InlineCode("fmt.Println() // inline code"),
		).

		// Code Block section
		Add(
			md.H3(md.Bold("10. Code Block")),
			md.Paragraph("For multi-line code blocks, use the `md.CodeBlock()` function:"),
			md.CodeBlock("package main\n\nfunc main() {\n\t// Your code here\n}"),
		).

		// Horizontal Rule section
		Add(
			md.H3(md.Bold("11. Horizontal Rule")),
			md.Paragraph("To add a horizontal rule (separator), use the `md.HorizontalRule()` function:"),
			md.LineBreak(1),
			md.HorizontalRule(),
		).

		// Language-specific Code Block section
		Add(
			md.H3(md.Bold("12. Language-specific Code Block")),
			md.Paragraph("To create language-specific code blocks, use the `md.LanguageCodeBlock()` function:"),
			md.LanguageCodeBlock("go", "package main\n\nfunc main() {}"),
		).

		// Hyperlink section
		Add(
			md.H3(md.Bold("13. Hyperlink")),
			md.Paragraph("To create a hyperlink, use the `md.Link()` function:"),
			md.Link("Gnoland official docs", "https://docs.gno.land"),
		).

		// Image section
		Add(
			md.H3(md.Bold("14. Image")),
			md.Paragraph("To insert an image, use the `md.Image()` function:"),
			md.LineBreak(1),
			md.Image("Gnoland Logo", "https://gnolang.github.io/blog/2024-05-21_the-gnome/src/banner.png"),
		).

		// Footnote section
		Add(
			md.H3(md.Bold("15. Footnote")),
			md.Paragraph("To create footnotes, use the `md.Footnote()` function:"),
			md.LineBreak(1),
			md.Footnote("1", "This is a footnote."),
		).

		// Table section
		Add(
			md.H3(md.Bold("16. Table")),
			md.Paragraph("To create a table, use the `md.Table()` function. Here's an example of a table:"),
		)

	// Create a table using the table package
	tb, _ := table.New([]string{"Feature", "Description"}, [][]string{
		{"Bold", "Make text bold using " + md.Bold("double asterisks")},
		{"Italic", "Make text italic using " + md.Italic("single asterisks")},
		{"Strikethrough", "Cross out text using " + md.Strikethrough("double tildes")},
	})
	mdBuilder.Add(md.Table(tb))

	// Escaping Markdown section
	mdBuilder.Add(
		md.H3(md.Bold("17. Escaping Markdown")),
		md.Paragraph("Sometimes, you need to escape special Markdown characters (like *, _, and `). Use the `md.EscapeMarkdown()` function for this:"),
	)

	// Example of escaping markdown
	text := "- Escape special chars like *, _, and ` in markdown"
	mdBuilder.Add(
		md.H4("Text Without Escape:"),
		text,
		md.LineBreak(1),
		md.H4("Text With Escape:"),
		md.EscapeMarkdown(text),
	)

	return mdBuilder.Render(md.LineBreak(1))
}
