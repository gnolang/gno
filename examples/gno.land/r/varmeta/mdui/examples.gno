package mdui

import (
	"gno.land/p/varmeta/mdui"
)

// RenderExamplesPage renders the Example page with sample components varmetanstrating the full range of MDUI functionalities
func RenderExamplesPage() string {
	builder := mdui.NewBuilder()
	builder.AddNavbar(NavbarLinks)

	// Introduction Section
	builder.AddHeading(1, "MDUI Example Page - Showcase All Components").
		AddParagraph(`
Welcome to the MDUI example page! This page varmetanstrates all the components available in MDUI, showcasing the power of Markdown for building user interfaces.
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

	// Table of Contents
	builder.AddHeading(2, "Table of Contents").
		AddTableOfContents([]string{"Introduction", "Key Features", "Alert Messages", "Code Example", "Feature Comparison Table", "More Details", "Inspirational Quote", "Keyboard Shortcut"}).
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