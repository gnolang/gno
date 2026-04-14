package main

import "gno.land/p/moul/md"

func main() {
	println(md.H1("Header 1"))
	println(md.H2("Header 2"))
	println(md.H3("Header 3"))
	println(md.H4("Header 4"))
	println(md.H5("Header 5"))
	println(md.H6("Header 6"))
	println(md.Bold("bold"))
	println(md.Italic("italic"))
	println(md.Strikethrough("strikethrough"))
	println(md.BulletList([]string{
		"Item 1",
		"Item 2\nMore details for item 2",
	}))
	println(md.OrderedList([]string{"Step 1", "Step 2"}))
	println(md.TodoList([]string{"Task 1", "Task 2\nSubtask 2"}, []bool{true, false}))
	println(md.Nested(md.BulletList([]string{"Parent Item", md.OrderedList([]string{"Child 1", "Child 2"})}), "  "))
	println(md.Blockquote("This is a blockquote\nSpanning multiple lines"))
	println(md.InlineCode("inline `code`"))
	println(md.CodeBlock("line1\nline2"))
	println(md.LanguageCodeBlock("go", "func main() {\nprintln(\"Hello, world!\")\n}"))
	println(md.HorizontalRule())
	println(md.Link("Gno", "http://gno.land"))
	println(md.Image("Alt Text", "http://example.com/image.png"))
	println(md.InlineImageWithLink("Alt Text", "http://example.com/image.png", "http://example.com"))
	println(md.Footnote("ref", "This is a footnote"))
	println(md.Paragraph("This is a paragraph."))

	println("4 columns in one gno-columns tag:")
	println(md.Columns([]string{
		"Column1\ncontent1",
		"Column2\ncontent2",
		"Column3\ncontent3",
		"Column4\ncontent4",
	}, true))

	// Should be automatically placed in multiple column tags
	println("3 cols per row without padding:")
	println(md.ColumnsN([]string{
		"Row1Column1\ncontent1",
		"Row1Column2\ncontent2",
		"Row1Column3\ncontent3",
		"Row2Column1\ncontent1",
		"Row2Column2\ncontent2",
		"Row2Column3\ncontent3",
		"Row3Column1\ncontent1",
		"Row3Column2\ncontent2",
		"Row3Column3\ncontent3",
	}, 3, false))

	// Should be padded, up to 4 cols
	println("2 padded to 4:")
	println(md.ColumnsN([]string{
		"Column1\ncontent1",
		"Column2\ncontent2",
	}, 4, true))

}

// Output:
// # Header 1
//
// ## Header 2
//
// ### Header 3
//
// #### Header 4
//
// ##### Header 5
//
// ###### Header 6
//
// **bold**
// *italic*
// ~~strikethrough~~
// - Item 1
// - Item 2
//   More details for item 2
//
// 1. Step 1
// 2. Step 2
//
// - [x] Task 1
// - [ ] Task 2
//   Subtask 2
//
//   - Parent Item
//   - 1. Child 1
//     2. Child 2
//
//
// > This is a blockquote
// > Spanning multiple lines
//
// `inline \`code\``
// ```
// line1
// line2
// ```
// ```go
// func main() {
// println("Hello, world!")
// }
// ```
// ---
//
// [Gno](http://gno.land)
// ![Alt Text](http://example.com/image.png)
// [![Alt Text](http://example.com/image.png)](http://example.com)
// [ref]: This is a footnote
// This is a paragraph.
//
//
// 4 columns in one gno-columns tag:
// <gno-columns>
// Column1
// content1
// |||
// Column2
// content2
// |||
// Column3
// content3
// |||
// Column4
// content4
// </gno-columns>
//
// 3 cols per row without padding:
// <gno-columns>
// Row1Column1
// content1
// |||
// Row1Column2
// content2
// |||
// Row1Column3
// content3
// </gno-columns>
// <gno-columns>
// Row2Column1
// content1
// |||
// Row2Column2
// content2
// |||
// Row2Column3
// content3
// </gno-columns>
// <gno-columns>
// Row3Column1
// content1
// |||
// Row3Column2
// content2
// |||
// Row3Column3
// content3
// </gno-columns>
//
// 2 padded to 4:
// <gno-columns>
// Column1
// content1
// |||
// Column2
// content2
// |||
//
// |||
//
// </gno-columns>
