package mdui

import (
	"gno.land/p/varmeta/mdui"
)

// RenderBuilderGuidePage renders a guide page explaining how to use the MarkdownBuilder
func RenderBuilderGuidePage() string {
	builder := mdui.NewBuilder()
	builder.AddNavbar(NavbarLinks)

	builder.AddHeading(1, "How to Use the MarkdownBuilder").
		AddParagraph(`
The MarkdownBuilder in MDUI is a powerful tool that allows developers to construct dynamic web pages with a simple, intuitive, and chainable API.
By leveraging the builder pattern, you can quickly and easily generate complex Markdown-based UIs without manually managing strings or formatting.
`).
		AddDivider().
		AddHeading(2, "1. Builder Overview").
		AddParagraph(`
The MarkdownBuilder accumulates UI elements like headings, paragraphs, buttons, tables, and more.
You can chain methods to add elements, and at the end, call the Render method to output the final Markdown.
`).
		AddCodeBlock(`
builder := mdui.NewBuilder()
// Add components using chained methods
builder.AddHeading(1, "Welcome to My Page").AddParagraph("This is a sample paragraph.").AddButton("Click Here", "/click")
// Render the final Markdown
result := builder.Render()
`).
		AddDivider().
		AddHeading(2, "2. Adding Components").
		AddParagraph(`
Use MarkdownBuilder to add various UI components like headings, paragraphs, lists, buttons, images, and more.
`).
		AddCodeBlock(`
// Adding a heading
builder.AddHeading(2, "This is a Heading")
// Adding a paragraph
builder.AddParagraph("This is a paragraph of text.")
// Adding a button
builder.AddButton("Learn More", "/learn-more")
// Adding a list
builder.AddList([]string{"Item 1", "Item 2", "Item 3"}, nil, false)
`).
		AddDivider().
		AddHeading(2, "3. Complete Example").
		AddParagraph("Below is a complete example using MarkdownBuilder:").
		AddCodeBlock(`
builder := mdui.NewBuilder()
// Build the content
builder.AddHeading(1, "Welcome to MDUI").
    AddParagraph("MDUI is a framework for building UIs with Markdown.").
    AddDivider().
	AddRaw("## Adding raw markdown").
    AddHeading(2, "Features").
    AddList([]string{
        "Simple to use",
        "Supports various UI components",
        "Easy to integrate with existing projects",
    }, nil, false).
    AddButton("Get Started", "/start")
// Render the final Markdown
result := builder.Render()
`).
		AddDivider().
		AddHeading(2, "4. Rendering Content").
		AddParagraph(`
Once you've added all your components, you can call the Render method to output the final Markdown string.
This string can then be used as content in your application.
`).
		AddCodeBlock(`result := builder.Render(); fmt.Println(result)`).
		AddDivider().
		AddHeading(2, "Conclusion").
		AddParagraph(`
The MarkdownBuilder makes it easy to generate complex UIs with Markdown using a clean and intuitive API.
With just a few method calls, you can add headings, paragraphs, buttons, tables, and more, and chain them together to build dynamic pages effortlessly.
`).
		AddButton("Start Using MarkdownBuilder", ExamplesPath)

	return builder.Render()
}