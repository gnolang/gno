package main

import "gno.land/p/moul/umd"

func main() {
	println(umd.H1("Header 1"))
	println(umd.H2("Header 2"))
	println(umd.H3("Header 3"))
	println(umd.H4("Header 4"))
	println(umd.H5("Header 5"))
	println(umd.H6("Header 6"))
	println(umd.Bold("bold"))
	println(umd.Italic("italic"))
	println(umd.Strikethrough("strikethrough"))
	println(umd.BulletList([]string{
		"Item 1",
		"Item 2\nMore details for item 2",
	}))
	println(umd.OrderedList([]string{"Step 1", "Step 2"}))
	println(umd.TodoList([]string{"Task 1", "Task 2\nSubtask 2"}, []bool{true, false}))
	println(umd.Nested(umd.BulletList([]string{"Parent Item", umd.OrderedList([]string{"Child 1", "Child 2"})}), "  "))
	println(umd.Blockquote("This is a blockquote\nSpanning multiple lines"))
	println(umd.InlineCode("inline `code`"))
	println(umd.CodeBlock("line1\nline2"))
	println(umd.LanguageCodeBlock("go", "func main() {\nprintln(\"Hello, world!\")\n}"))
	println(umd.HorizontalRule())
	println(umd.Link("Gno", "http://gno.land"))
	println(umd.Image("Alt Text", "http://example.com/image.png"))
	println(umd.InlineImageWithLink("Alt Text", "http://example.com/image.png", "http://example.com"))
	println(umd.Footnote("ref", "This is a footnote"))
	println(umd.Paragraph("This is a paragraph."))
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
