package mdui

import (
	"strings"

	"gno.land/p/demo/mdui"
)

// Constants for common paths
const (
	BasePath         = "/r/demo/mdui"
	ComponentsPath   = "/r/demo/mdui:components"
	ExamplesPath     = "/r/demo/mdui:examples"
	BuilderGuidePath = "/r/demo/mdui:builder-guide"
)

// Common Navbar Links
var NavbarLinks = map[string]string{
	"Home":          BasePath,
	"Components":    ComponentsPath,
	"Examples":      ExamplesPath,
	"Builder Guide": BuilderGuidePath,
}

// RenderHomePage renders the Home page using MarkdownBuilder
func RenderHomePage() string {
	builder := mdui.NewBuilder()
	builder.AddNavbar(NavbarLinks)

	builder.AddHeading(1, "Welcome to MDUI - Markdown-based UI Development").
		AddParagraph(`
MDUI is a versatile UI framework built on the simplicity and power of Markdown.
It enables developers to create user interfaces with ease by using Markdown syntax, making web development more accessible and streamlined.
Explore our component library and see how MDUI can simplify your development workflow.
`).
		AddDivider().
		AddHeading(2, "Getting Started with MDUI").
		AddParagraph(`
Ready to start using MDUI? Follow these simple steps to get up and running:
1. [Explore Components](/r/demo/mdui:components) - Discover a variety of components available in MDUI.
2. [Check out Examples](/r/demo/mdui:examples) - Learn how to build real-world projects with MDUI.
3. [Try It Out](/r/demo/mdui:examples) - See how easy it is to create your first page.
`).
		AddButton("Start Building Now", ComponentsPath).
		AddDivider().
		AddHeading(2, "Why Choose MDUI?").
		AddList([]string{
			"Simple and intuitive Markdown-based syntax.",
			"Reusable components for rapid development.",
			"Easy integration with existing Markdown projects.",
			"Perfect for quickly creating functional prototypes.",
		}, nil, false).
		AddDivider()

	return builder.Render()
}

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
			Example:     mdui.Navbar(map[string]string{"Home": "/home", "About": "/about", "Contact": "/contact"}),
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
			Title:       "Emoji",
			Description: "Adds an emoji to the content using a shortcode.",
			Example:     mdui.Emoji("smile"),
			SampleCode:  `mdui.Emoji("smile")`,
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

// RenderExamplesPage renders the Example page with sample components demonstrating the full range of MDUI functionalities
func RenderExamplesPage() string {
	builder := mdui.NewBuilder()
	builder.AddNavbar(NavbarLinks)

	// Introduction Section
	builder.AddHeading(1, "MDUI Example Page - Showcase All Components").
		AddParagraph(`
Welcome to the MDUI example page! This page demonstrates all the components available in MDUI, showcasing the power of Markdown for building user interfaces.
Explore the various elements below to see how you can create beautiful and functional web pages using MDUI.
`).
		AddDivider()

	// Hero Section with Image and Badge
	builder.AddImage("https://media.istockphoto.com/id/1359861587/vector/blue-modern-abstract-wide-banner-with-geometric-shapes-dark-blue-abstract-background.jpg?s=612x612&w=0&k=20&c=xDMI5_-xyh_IHm2ca07x5xFTR7kRoFZJhxrq10FQra0=", "Hero Image").
		AddHeading(2, "Welcome to MDUI").
		AddBadge("New Feature", "green").
		AddParagraph("MDUI enables developers to build modern, clean, and responsive user interfaces effortlessly.")

	// Key Features List
	builder.AddHeading(2, "Key Features").
		AddList([]string{
			"Responsive design with Markdown simplicity",
			"Reusable components for rapid development",
			"Customizable themes and styling",
			"Easy integration with existing projects",
		}, nil, false).
		AddDivider()

	// Alerts Section
	builder.AddHeading(2, "Alert Messages").
		AddAlert("This is an informational message.", "info").
		AddAlert("Be cautious about this warning.", "warning").
		AddAlert("This is a dangerous action. Proceed with care.", "danger").
		AddAlert("Congratulations, your action was successful!", "success").
		AddDivider()

	// Code Block Example
	builder.AddHeading(2, "Code Example").
		AddCodeBlock(`func main() {
    fmt.Println("Hello, MDUI!")
}`).
		AddDivider()

	// Table Example
	builder.AddHeading(2, "Feature Comparison Table").
		AddTable([]string{"Feature", "MDUI", "Other Frameworks"}, [][]string{
			{"Ease of Use", "High", "Medium"},
			{"Flexibility", "High", "Variable"},
			{"Learning Curve", "Low", "Medium"},
		}).
		AddDivider()

	// Collapsible Section
	builder.AddHeading(2, "More Details").
		AddCollapsible("Expand to see more", `
Here is some more detailed information about MDUI and its components.
You can use collapsible sections to keep the layout clean and reveal information when needed.
`).
		AddDivider()

	// Blockquote with Citation
	builder.AddHeading(2, "Inspirational Quote").
		AddBlockquoteWithCitation("Simplicity is the ultimate sophistication.", "Leonardo da Vinci").
		AddDivider()

	// Keyboard Shortcut Example
	builder.AddHeading(2, "Keyboard Shortcut").
		AddParagraph("Press the following shortcut to refresh the page:").
		AddKeyboardShortcut("Ctrl", "R").
		AddDivider()

	// Emoji Example
	builder.AddHeading(2, "Add Some Fun with Emojis").
		AddParagraph("MDUI supports popular emojis to enhance your content:").
		AddEmoji("smile").
		AddEmoji("rocket").
		AddEmoji("thumbs_up").
		AddDivider()

	// Table of Contents
	builder.AddHeading(2, "Table of Contents").
		AddTableOfContents([]string{"Introduction", "Key Features", "Alert Messages", "Code Example", "Feature Comparison Table", "More Details", "Inspirational Quote", "Keyboard Shortcut", "Emojis"}).
		AddDivider()

	// Conditional Rendering Example
	builder.AddHeading(2, "Conditional Content").
		AddParagraph("MDUI also supports conditional content rendering. See the result below based on a true/false condition.").
		AddIfElseRender(true, "This content is displayed because the condition is true.", "This content would display if the condition were false").
		AddDivider()

	// Badges Section with Icons
	builder.AddHeading(2, "Badge Examples").
		AddParagraph("Show off your badges with or without icons:").
		AddBadge("Beta", "blue").
		AddBadgeWithIcon("Go", "blue", "go").
		AddBadgeWithIcon("JavaScript", "yellow", "javascript").
		AddBadge("Build Passing", "green").
		AddDivider()

	// Final Message and Button
	builder.AddHeading(2, "Ready to Start Building with MDUI?").
		AddParagraph("MDUI makes it simple to create UIs with Markdown. Try it out by clicking the button below:").
		AddButton("Start Using MDUI", "/start")

	return builder.Render()
}

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

// Render entry point to route to different pages based on path
func Render(path string) string {
	switch {
	case strings.Contains(path, "components"):
		return RenderComponentsPage()
	case strings.Contains(path, "examples"):
		return RenderExamplesPage()
	case strings.Contains(path, "builder-guide"):
		return RenderBuilderGuidePage()
	default:
		return RenderHomePage()
	}
}
