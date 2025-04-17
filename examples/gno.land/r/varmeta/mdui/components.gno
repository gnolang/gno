package mdui

import (
	"gno.land/p/varmeta/mdui"
)

// RenderComponentsPage renders the Components page with a list of available components, examples, and sample code
func RenderComponentsPage() string {
	builder := mdui.NewBuilder()
	builder.AddNavbar(NavbarLinks)

	builder.AddHeading(1, "MDUI Components Overview").
		AddParagraph(`
MDUI offers a variety of components that help developers build modern user interfaces quickly and efficiently.
Below is a list of components available in the MDUI package, along with examples of how to use them and sample code.
`)

	// Comprehensive list of components
	componentList := []struct {
		Title       string
		Description string
		Example     string
		SampleCode  string
	}{
		{
			Title:       "Navbar",
			Description: "A navigation menu that can be used to create a list of links for site navigation.",
			Example:     mdui.Navbar(NavbarLinks),
			SampleCode:  `mdui.Navbar(map[string]string{"Home": "/home", "About": "/about", "Contact": "/contact"})`,
		},
		{
			Title:       "Heading",
			Description: "A heading component to create titles of different levels (from H1 to H6).",
			Example:     mdui.Heading(2, "Sample Heading"),
			SampleCode:  `mdui.Heading(2, "Sample Heading")`,
		},
		{
			Title:       "Paragraph",
			Description: "Formats a block of text as a paragraph.",
			Example:     mdui.Paragraph("This is a sample paragraph."),
			SampleCode:  `mdui.Paragraph("This is a sample paragraph.")`,
		},
		{
			Title:       "Button",
			Description: "A clickable button component, implemented as a link in Markdown.",
			Example:     mdui.Button("Click Here", "/click-here"),
			SampleCode:  `mdui.Button("Click Here", "/click-here")`,
		},
		{
			Title:       "Image",
			Description: "An image component for displaying pictures with optional alt text.",
			Example:     mdui.Image("https://img-cdn.pixlr.com/image-generator/history/65bb506dcb310754719cf81f/ede935de-1138-4f66-8ed7-44bd16efc709/medium.webp", "Sample Image"),
			SampleCode:  `mdui.Image("https://img-cdn.pixlr.com/image-generator/history/65bb506dcb310754719cf81f/ede935de-1138-4f66-8ed7-44bd16efc709/medium.webp", "Sample Image")`,
		},
		{
			Title:       "Divider",
			Description: "A horizontal line divider for separating content.",
			Example:     mdui.Divider(),
			SampleCode:  `mdui.Divider()`,
		},
		{
			Title:       "List (Unordered with Links)",
			Description: "Generates an unordered Markdown list with each item as a link.",
			Example:     mdui.List([]string{"Item 1", "Item 2", "Item 3"}, []string{"#item1", "#item2", "#item3"}, false),
			SampleCode:  `mdui.List([]string{"Item 1", "Item 2", "Item 3"}, []string{"#item1", "#item2", "#item3"}, false)`,
		},
		{
			Title:       "List (Ordered with Links)",
			Description: "Generates an ordered Markdown list with each item as a link.",
			Example:     mdui.List([]string{"Step 1", "Step 2", "Step 3"}, []string{"#step1", "#step2", "#step3"}, true),
			SampleCode:  `mdui.List([]string{"Step 1", "Step 2", "Step 3"}, []string{"#step1", "#step2", "#step3"}, true)`,
		},
		{
			Title:       "Table",
			Description: "Generates a table with headers and rows for tabular data.",
			Example:     mdui.Table([]string{"Name", "Age"}, [][]string{{"Alice", "30"}, {"Bob", "25"}}),
			SampleCode:  `mdui.Table([]string{"Name", "Age"}, [][]string{{"Alice", "30"}, {"Bob", "25"}})`,
		},
		{
			Title:       "Quote",
			Description: "A blockquote component for highlighting quoted text.",
			Example:     mdui.Quote("This is a quoted text."),
			SampleCode:  `mdui.Quote("This is a quoted text.")`,
		},
		{
			Title:       "Bold",
			Description: "Formats text in bold.",
			Example:     mdui.Bold("This is bold text"),
			SampleCode:  `mdui.Bold("This is bold text")`,
		},
		{
			Title:       "Italic",
			Description: "Formats text in italic.",
			Example:     mdui.Italic("This is italic text"),
			SampleCode:  `mdui.Italic("This is italic text")`,
		},
		{
			Title:       "Strikethrough",
			Description: "Adds a strikethrough to the text.",
			Example:     mdui.Strikethrough("This text is crossed out"),
			SampleCode:  `mdui.Strikethrough("This text is crossed out")`,
		},
		{
			Title:       "Code Block",
			Description: "A component for rendering code blocks in Markdown.",
			Example:     mdui.CodeBlock("func main() {\n    fmt.Println(\"Hello, world!\")\n}"),
			SampleCode:  `mdui.CodeBlock("func main() {\n    fmt.Println(\"Hello, world!\")\n}")`,
		},
		{
			Title:       "Alert",
			Description: "An alert block using Markdown with a specified type (info, warning, danger, success).",
			Example:     mdui.Alert("This action cannot be undone.", "danger"),
			SampleCode:  `mdui.Alert("This action cannot be undone.", "danger")`,
		},
		{
			Title:       "Collapsible",
			Description: "A collapsible section that expands when clicked, using plain Markdown.",
			Example:     mdui.Collapsible("Show more", "Here is the additional content."),
			SampleCode:  `mdui.Collapsible("Show more", "Here is the additional content.")`,
		},
		{
			Title:       "Footnote",
			Description: "Generates a footnote reference and definition.",
			Example:     mdui.Footnote("1", "This is a footnote."),
			SampleCode:  `mdui.Footnote("1", "This is a footnote.")`,
		},
		{
			Title:       "Badge",
			Description: "Creates a badge using the Shields.io service.",
			Example:     mdui.Badge("Build Passing", "green"),
			SampleCode:  `mdui.Badge("Build Passing", "green")`,
		},
		{
			Title:       "Badge with Icon",
			Description: "Creates a badge with an icon using Shields.io.",
			Example:     mdui.BadgeWithIcon("Go Lang", "blue", "go"),
			SampleCode:  `mdui.BadgeWithIcon("Go Lang", "blue", "go")`,
		},
		{
			Title:       "Table of Contents",
			Description: "Generates a simple table of contents based on an array of headings.",
			Example:     mdui.TableOfContents([]string{"Introduction", "Features", "Conclusion"}),
			SampleCode:  `mdui.TableOfContents([]string{"Introduction", "Features", "Conclusion"})`,
		},
		{
			Title:       "Keyboard Shortcut",
			Description: "Formats a keyboard shortcut in Markdown using code block style.",
			Example:     mdui.KeyboardShortcut("Ctrl", "Alt", "Del"),
			SampleCode:  `mdui.KeyboardShortcut("Ctrl", "Alt", "Del")`,
		},
		{
			Title:       "Blockquote with Citation",
			Description: "Generates a Markdown blockquote with an optional citation.",
			Example:     mdui.BlockquoteWithCitation("To be or not to be, that is the question.", "William Shakespeare"),
			SampleCode:  `mdui.BlockquoteWithCitation("To be or not to be, that is the question.", "William Shakespeare")`,
		},
		{
			Title:       "BreakLine",
			Description: "Adds a line break (two spaces followed by a newline) in Markdown.",
			Example:     mdui.BreakLine(),
			SampleCode:  `mdui.BreakLine()`,
		},
		{
			Title:       "IfElseRender",
			Description: "Conditionally renders content based on a Boolean condition.",
			Example:     mdui.IfElseRender(true, "Condition is True", "Condition is False"),
			SampleCode:  `mdui.IfElseRender(true, "Condition is True", "Condition is False")`,
		},
	}

	// Generate Markdown for each component in the list
	for _, component := range componentList {
		builder.AddHeading(2, component.Title).
			AddParagraph(component.Description).
			AddParagraph("**Rendered Example:**").
			AddParagraph(component.Example).
			AddParagraph("**Sample Code:**").
			AddCodeBlock(component.SampleCode).
			AddDivider()
	}

	return builder.Render()
}