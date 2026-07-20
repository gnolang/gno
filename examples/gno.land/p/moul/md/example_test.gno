package md

import "strings"

func ExampleHeaders() {
	println(H1("Header 1"))
	println(H2("Header 2"))
	println(H3("Header 3"))
	println(H4("Header 4"))
	println(H5("Header 5"))
	println(H6("Header 6"))

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
}

func ExampleStyles() {
	println(Bold("bold"))
	println(Italic("italic"))
	println(Strikethrough("strikethrough"))

	// Output:
	// **bold**
	// *italic*
	// ~~strikethrough~~
}

func ExampleLists() {
	println(BulletList([]string{
		"Item 1",
		"Item 2\nMore details for item 2",
	}))
	println(OrderedList([]string{"Step 1", "Step 2"}))
	println(TodoList([]string{"Task 1", "Task 2\nSubtask 2"}, []bool{true, false}))
	println(Nested(BulletList([]string{"Parent Item", OrderedList([]string{"Child 1", "Child 2"})}), "  "))

	// Output:
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
}

func ExampleBlocks() {
	print(Paragraph("This is a paragraph."))
	// Blockquote adds multiple newlines which Output doesn't allow, so trim
	println(strings.TrimSpace(Blockquote("This is a blockquote\nSpanning multiple lines")))

	// Output:
	// This is a paragraph.
	//
	// > This is a blockquote
	// > Spanning multiple lines
}

func ExampleCode() {
	println(InlineCode("inline `code`"))
	println(CodeBlock("line1\nline2"))
	println(LanguageCodeBlock("go", "func main() {\nprintln(\"Hello, world!\")\n}"))
	println(HorizontalRule())

	// Output:
	// `` inline `code` ``
	// ```
	// line1
	// line2
	// ```
	//
	// ```go
	// func main() {
	// println("Hello, world!")
	// }
	// ```
	//
	// ---
}

func ExampleReferences() {
	println(Link("Gno", "http://gno.land"))
	println(Image("Alt Text", "http://example.com/image.png"))
	println(InlineImageWithLink("Alt Text", "http://example.com/image.png", "http://example.com"))
	println(FootnoteDefinition("ref", "This is a footnote"))
	// LinkReferenceDefinition adds multiple newlines which Output doesn't allow, so trim
	println(strings.TrimSpace(LinkReferenceDefinition("r-example", "/r/example", "")))

	// Output:
	//
	// [Gno](http://gno.land)
	// ![Alt Text](http://example.com/image.png)
	// [![Alt Text](http://example.com/image.png)](http://example.com)
	// [^ref]:
	//     This is a footnote
	//
	// [r-example]: /r/example
}

func ExampleColumns() {
	println("4 columns in one gno-columns tag:")
	println(Columns([]string{
		"Column1\ncontent1",
		"Column2\ncontent2",
		"Column3\ncontent3",
		"Column4\ncontent4",
	}, true))

	// Should be automatically placed in multiple column tags
	println("3 cols per row without padding:")
	println(ColumnsN([]string{
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
	println(ColumnsN([]string{
		"Column1\ncontent1",
		"Column2\ncontent2",
	}, 4, true))

	// Output:
	// 4 columns in one gno-columns tag:
	// <gno-columns>
	// Column1
	// content1
	// <gno-columns-sep>
	// Column2
	// content2
	// <gno-columns-sep>
	// Column3
	// content3
	// <gno-columns-sep>
	// Column4
	// content4
	// </gno-columns>
	//
	// 3 cols per row without padding:
	// <gno-columns>
	// Row1Column1
	// content1
	// <gno-columns-sep>
	// Row1Column2
	// content2
	// <gno-columns-sep>
	// Row1Column3
	// content3
	// </gno-columns>
	// <gno-columns>
	// Row2Column1
	// content1
	// <gno-columns-sep>
	// Row2Column2
	// content2
	// <gno-columns-sep>
	// Row2Column3
	// content3
	// </gno-columns>
	// <gno-columns>
	// Row3Column1
	// content1
	// <gno-columns-sep>
	// Row3Column2
	// content2
	// <gno-columns-sep>
	// Row3Column3
	// content3
	// </gno-columns>
	//
	// 2 padded to 4:
	// <gno-columns>
	// Column1
	// content1
	// <gno-columns-sep>
	// Column2
	// content2
	// <gno-columns-sep>
	//
	// <gno-columns-sep>
	//
	// </gno-columns>
}
