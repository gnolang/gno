package mdui

import (
	"strings"

	"gno.land/p/demo/mdui"
)

// RenderHomePage - Renders the Home page using MarkdownBuilder
func RenderHomePage() string {
	builder := mdui.NewMarkdownBuilder()

	// Navbar with links
	navbarLinks := map[string]string{
		"Home":          "/r/demo/mdui",
		"Components":    "/r/demo/mdui:components",
		"Examples":      "/r/demo/mdui:examples",
		"Builder Guide": "/r/demo/mdui:builder-guide",
	}
	builder.AddNavbar(navbarLinks)

	// Heading and Introduction
	builder.AddHeading(1, "Welcome to MDUI - Markdown-based UI Development").
		AddParagraph(`
MDUI is a versatile UI framework built on the simplicity and power of Markdown.
It enables developers to create user interfaces with ease by using Markdown syntax, making web development more accessible and streamlined.
MDUI is ideal for anyone who wants to quickly build functional, beautiful web pages using familiar Markdown concepts.
Explore our component library and see how MDUI can simplify your development workflow.
`).
		AddDivider()

	// Getting Started Section
	builder.AddHeading(2, "Getting Started with MDUI").
		AddParagraph(`
Ready to start using MDUI? Follow these simple steps to get up and running:
1. [Explore Components](/r/demo/mdui:components) - Discover a variety of components available in MDUI.
2. [Check out Examples](/r/demo/mdui:examples) - Learn how to build real-world projects with MDUI.
3. [Try It Out](/r/demo/mdui:examples) - See how easy it is to create your first page.
`).
		AddButton("Start Building Now", "/r/demo/mdui:components").
		AddDivider()

	// Benefits Section
	builder.AddHeading(2, "Why Choose MDUI?").
		AddList([]string{
			"Simple and intuitive Markdown-based syntax.",
			"Reusable components for rapid development.",
			"Easy integration with existing Markdown projects.",
			"Perfect for quickly creating functional prototypes.",
		}, false).
		AddDivider()

	return builder.Render()
}

// RenderComponentsPage - Renders the Components page with a list of available components, examples, and sample code
func RenderComponentsPage() string {
	// Navbar with updated links
	navbarLinks := map[string]string{
		"Home":          "/r/demo/mdui",
		"Components":    "/r/demo/mdui:components",
		"Examples":      "/r/demo/mdui:examples",
		"Builder Guide": "/r/demo/mdui:builder-guide",
	}
	navbar := mdui.Navbar(navbarLinks)

	// Main title and introduction
	headingMain := mdui.Heading(1, "MDUI Components Overview")
	descriptionMain := mdui.Paragraph(`
MDUI offers a variety of components that help developers build modern user interfaces quickly and efficiently.
Below is a list of components available in the MDUI package, along with examples of how to use them and sample code.
`)

	// List of components with descriptions, examples, and sample code
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
			Title:       "Code Block",
			Description: "A component for rendering code blocks in Markdown.",
			Example:     mdui.CodeBlock("func main() {\n    fmt.Println(\"Hello, world!\")\n}"),
			SampleCode:  `mdui.CodeBlock("func main() {\n    fmt.Println(\"Hello, world!\")\n}")`,
		},
		{
			Title:       "Divider",
			Description: "A horizontal line divider for separating content.",
			Example:     mdui.Divider(),
			SampleCode:  `mdui.Divider()`,
		},
		{
			Title:       "Paragraph",
			Description: "Formats a block of text as a paragraph.",
			Example:     mdui.Paragraph("This is a sample paragraph."),
			SampleCode:  `mdui.Paragraph("This is a sample paragraph.")`,
		},
		{
			Title:       "Quote",
			Description: "A blockquote component for highlighting quoted text.",
			Example:     mdui.Quote("This is a quoted text."),
			SampleCode:  `mdui.Quote("This is a quoted text.")`,
		},
		{
			Title:       "List (Unordered)",
			Description: "Generates an unordered Markdown list.",
			Example:     mdui.List([]string{"Item 1", "Item 2", "Item 3"}, false),
			SampleCode:  `mdui.List([]string{"Item 1", "Item 2", "Item 3"}, false)`,
		},
		{
			Title:       "List (Ordered)",
			Description: "Generates an ordered Markdown list.",
			Example:     mdui.List([]string{"Step 1", "Step 2", "Step 3"}, true),
			SampleCode:  `mdui.List([]string{"Step 1", "Step 2", "Step 3"}, true)`,
		},
		{
			Title:       "Table",
			Description: "Generates a table with headers and rows for tabular data.",
			Example:     mdui.Table([]string{"Name", "Age"}, [][]string{{"Alice", "30"}, {"Bob", "25"}}),
			SampleCode:  `mdui.Table([]string{"Name", "Age"}, [][]string{{"Alice", "30"}, {"Bob", "25"}})`,
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
			Title:       "Alert",
			Description: "An alert block using Markdown with a specified type (info, warning, danger).",
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
			Title:       "Badge with Icon",
			Description: "Creates a badge with an icon using Shields.io.",
			Example:     mdui.BadgeWithIcon("Go Lang", "blue", "go"),
			SampleCode:  `mdui.BadgeWithIcon("Go Lang", "blue", "go")`,
		},
	}

	// Generate Markdown for each component
	componentMarkdown := ""
	for _, component := range componentList {
		componentMarkdown += mdui.Heading(2, component.Title) + "\n"
		componentMarkdown += mdui.Paragraph(component.Description) + "\n"
		componentMarkdown += mdui.Paragraph("**Rendered Example:**") + "\n"
		componentMarkdown += component.Example + "\n\n"
		componentMarkdown += mdui.Paragraph("**Sample Code:**") + "\n"
		componentMarkdown += mdui.CodeBlock(component.SampleCode) + "\n"
		componentMarkdown += mdui.Divider() + "\n"
	}

	// Combine everything into the components page
	return navbar + "\n" +
		headingMain + "\n" +
		descriptionMain + "\n" +
		componentMarkdown
}

func RenderExamplesPage() string {
	// Navbar with links
	navbarLinks := map[string]string{
		"Home":          "/r/demo/mdui",
		"Components":    "/r/demo/mdui:components",
		"Examples":      "/r/demo/mdui:examples",
		"Builder Guide": "/r/demo/mdui:builder-guide",
	}
	navbar := mdui.Navbar(navbarLinks)

	// Main title and introduction
	headingMain := mdui.Heading(1, "MDUI Example Page - Showcase All Components")
	introParagraph := mdui.Paragraph(`
Welcome to the MDUI example page! This page demonstrates all the components available in MDUI, showcasing the power of Markdown for building user interfaces.
Explore the various elements below to see how you can create beautiful and functional web pages using MDUI.
`)

	// Hero section
	heroImage := "https://png.pngtree.com/thumb_back/fh260/background/20230711/pngtree-d-rendering-of-a-landing-page-concept-featuring-an-abstract-blue-image_3838290.jpg"
	heroTitle := mdui.Heading(2, "Welcome to the World of MDUI")
	heroText := mdui.Paragraph(`
MDUI enables developers to build modern, clean, and responsive user interfaces effortlessly.
This example showcases all the components available in the MDUI library.
`)

	// Using the Badge component
	badge := mdui.Badge("New Feature", "green")

	// Features section using List
	featuresHeading := mdui.Heading(2, "Key Features")
	featuresList := mdui.List([]string{
		"Responsive design with Markdown simplicity",
		"Reusable components for rapid development",
		"Customizable themes and styling",
		"Easy integration with existing projects",
	}, false)

	// Alert section for different alert types
	alertSectionHeading := mdui.Heading(2, "Alert Messages")
	alertInfo := mdui.Alert("This is an informational message.", "info")
	alertWarning := mdui.Alert("Be cautious about this warning.", "warning")
	alertDanger := mdui.Alert("This is a dangerous action. Proceed with care.", "danger")

	// Code block example
	codeBlockHeading := mdui.Heading(2, "Code Example")
	codeBlock := mdui.CodeBlock(`func main() {
    fmt.Println("Hello, MDUI!")
}`)

	// Table example
	tableHeading := mdui.Heading(2, "Feature Comparison Table")
	table := mdui.Table([]string{"Feature", "MDUI", "Other Frameworks"}, [][]string{
		{"Ease of Use", "High", "Medium"},
		{"Flexibility", "High", "Variable"},
		{"Learning Curve", "Low", "Medium"},
	})

	// Collapsible section
	collapsible := mdui.Collapsible("Show more details", `
Here is some more detailed information about MDUI and its components.
You can use collapsible sections to keep the layout clean and reveal information when needed.
`)

	// Quote example
	quoteHeading := mdui.Heading(2, "Inspirational Quote")
	quote := mdui.Quote("Simplicity is the ultimate sophistication. - Leonardo da Vinci")

	// Footnote example
	footnote := mdui.Footnote("1", "MDUI is a fictional Markdown UI framework for demonstration purposes.")

	// Divider for separating sections
	divider := mdui.Divider()

	// Organize all elements to create a comprehensive example
	content := navbar + "\n" +
		headingMain + "\n" +
		introParagraph + "\n" +
		divider + "\n" +
		mdui.Image(heroImage, "Landing Page Background") + "\n" +
		heroTitle + " " + badge + "\n" +
		heroText + "\n" +
		divider + "\n" +
		featuresHeading + "\n" +
		featuresList + "\n" +
		divider + "\n" +
		alertSectionHeading + "\n" +
		alertInfo + "\n" +
		alertWarning + "\n" +
		alertDanger + "\n" +
		divider + "\n" +
		codeBlockHeading + "\n" +
		codeBlock + "\n" +
		divider + "\n" +
		tableHeading + "\n" +
		table + "\n" +
		divider + "\n" +
		collapsible + "\n" +
		divider + "\n" +
		quoteHeading + "\n" +
		quote + "\n" +
		divider + "\n" +
		footnote + "\n"

	return content
}

// RenderBuilderGuidePage - Renders a guide page explaining how to use the MarkdownBuilder
func RenderBuilderGuidePage() string {
	builder := mdui.NewMarkdownBuilder()

	// Navbar
	navbarLinks := map[string]string{
		"Home":          "/r/demo/mdui",
		"Components":    "/r/demo/mdui:components",
		"Examples":      "/r/demo/mdui:examples",
		"Builder Guide": "/r/demo/mdui:builder-guide",
	}
	builder.AddNavbar(navbarLinks)

	// Title and introduction
	builder.AddHeading(1, "How to Use the MarkdownBuilder").
		AddParagraph(`
The MarkdownBuilder in MDUI is a powerful tool that allows developers to construct dynamic web pages with a simple, intuitive, and chainable API.
By leveraging the builder pattern, you can quickly and easily generate complex Markdown-based UIs without manually managing strings or formatting.
This guide will show you how to use the MarkdownBuilder to create content step by step.
`).
		AddDivider()

	// Section: Builder Overview
	builder.AddHeading(2, "1. Builder Overview").
		AddParagraph(`
The MarkdownBuilder is an object that accumulates UI elements like headings, paragraphs, buttons, tables, and more.
You can chain methods to add elements, and at the end, call the Render method to output the final Markdown.
`).
		AddCodeBlock(`
builder := mdui.NewMarkdownBuilder()

// Add components using chained methods
builder.AddHeading(1, "Welcome to My Page").
    AddParagraph("This is a sample paragraph.").
    AddButton("Click Here", "/click")

// Render the final Markdown
result := builder.Render()
`).
		AddDivider()

	// Section: Adding Components
	builder.AddHeading(2, "2. Adding Components").
		AddParagraph(`
You can use the MarkdownBuilder to add various UI components like headings, paragraphs, lists, buttons, images, and more. Below are some examples:
`).
		AddCodeBlock(`
// Adding a heading
builder.AddHeading(2, "This is a Heading")

// Adding a paragraph
builder.AddParagraph("This is a paragraph of text.")

// Adding a button
builder.AddButton("Learn More", "/learn-more")

// Adding a list
builder.AddList([]string{"Item 1", "Item 2", "Item 3"}, false)
`).
		AddDivider()

	// Section: Complete Example
	builder.AddHeading(2, "3. Complete Example").
		AddParagraph(`
Below is a complete example that demonstrates how to use the MarkdownBuilder to create a simple webpage:
`).
		AddCodeBlock(`
builder := mdui.NewMarkdownBuilder()

// Build the content
builder.AddHeading(1, "Welcome to MDUI").
    AddParagraph("MDUI is a framework for building UIs with Markdown.").
    AddDivider().
    AddHeading(2, "Features").
    AddList([]string{
        "Simple to use",
        "Supports various UI components",
        "Easy to integrate with existing projects",
    }, false).
    AddButton("Get Started", "/start")

// Render the final Markdown
result := builder.Render()
`).
		AddDivider()

	// Section: Rendering Content
	builder.AddHeading(2, "4. Rendering Content").
		AddParagraph(`
Once you've added all your components, you can call the Render method to output the final Markdown string.
This string can then be used as content in your application.
`).
		AddCodeBlock(`
result := builder.Render()
fmt.Println(result)
`).
		AddDivider()

	// Section: Conclusion
	builder.AddHeading(2, "Conclusion").
		AddParagraph(`
The MarkdownBuilder makes it easy to generate complex UIs with Markdown using a clean and intuitive API.
With just a few method calls, you can add headings, paragraphs, buttons, tables, and more, and chain them together to build dynamic pages effortlessly.
Try out the MarkdownBuilder in your next project and see how it simplifies your workflow!
`).
		AddButton("Start Using MarkdownBuilder", "/r/demo/mdui:examples")

	return builder.Render()
}

// Render - The main entry point for displaying different pages
func Render(path string) string {
	// Routing logic for multiple pages
	if strings.Contains(path, "components") {
		return RenderComponentsPage()
	} else if strings.Contains(path, "examples") {
		return RenderExamplesPage()
	} else if strings.Contains(path, "builder-guide") {
		return RenderBuilderGuidePage()
	}
	return RenderHomePage() // Default to the home page
}
