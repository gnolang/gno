package moul_md

import "gno.land/p/moul/md"

// This package demonstrates the usage of the gno.land/p/moul/md package.

func Render(path string) string {
	sections := []string{
		renderIntro(),
		renderTextFormatting(),
		renderHeaders(),
		renderLists(),
		renderLinksAndImages(),
		renderCode(),
		renderBlockquotes(),
		renderAdvancedFeatures(),
		renderUtilityFunctions(),
		renderPracticalExample(),
		renderBestPractices(),
		renderSeeAlso(),
	}

	output := ""
	for _, section := range sections {
		output += section
	}
	return output
}

func renderIntro() string {
	return md.H1("Package Demo: `p/moul/md` ") +
		md.Paragraph("This document demonstrates the features of the "+md.Link("gno.land/p/moul/md", "https://gno.land/p/moul/md")+" package, showing both the source code and the rendered result for each feature.") +
		"\n" +
		md.Blockquote(md.Bold("Note:")+" The md package provides helper functions to generate markdown programmatically, making it easier to create dynamic documentation and content.") +
		"\n"
}

func renderTextFormatting() string {
	return md.H2("Text Formatting") + "\n" +
		// Bold
		renderExample(
			"Bold Text",
			`md.Bold("Important text")`,
			md.Bold("Important text"),
		) +
		// Italic
		renderExample(
			"Italic Text",
			`md.Italic("Emphasized text")`,
			md.Italic("Emphasized text"),
		) +
		// Strikethrough
		renderExample(
			"Strikethrough Text",
			`md.Strikethrough("Deprecated feature")`,
			md.Strikethrough("Deprecated feature"),
		) +
		// Combined
		renderExample(
			"Combined Formatting",
			`md.Bold(md.Italic("Very important"))`,
			md.Bold(md.Italic("Very important")),
		)
}

func renderHeaders() string {
	code := `md.H1("Main Title")
md.H2("Section")
md.H3("Subsection")
md.H4("Sub-subsection")
md.H5("Minor heading")
md.H6("Smallest heading")`

	result := md.H1("Main Title") +
		md.H2("Section") +
		md.H3("Subsection") +
		md.H4("Sub-subsection") +
		md.H5("Minor heading") +
		md.H6("Smallest heading")

	return md.H2("Headers") +
		md.Paragraph("The package supports all six levels of markdown headers:") +
		"\n" +
		md.LanguageCodeBlock("go", code) +
		"\nResult:\n" +
		result +
		"\n"
}

func renderLists() string {
	output := md.H2("Lists") + "\n"

	// Bullet Lists
	bulletCode := `items := []string{
    "First item",
    "Second item",
    "Third item with\nmultiple lines",
}
md.BulletList(items)`

	output += md.H3("Bullet Lists") +
		md.LanguageCodeBlock("go", bulletCode) +
		"\nResult:\n" +
		md.BulletList([]string{
			"First item",
			"Second item",
			"Third item with\nmultiple lines",
		}) + "\n"

	// Ordered Lists
	orderedCode := `steps := []string{
    "First step",
    "Second step",
    "Third step with\nadditional details",
}
md.OrderedList(steps)`

	output += md.H3("Ordered Lists") +
		md.LanguageCodeBlock("go", orderedCode) +
		"\nResult:\n" +
		md.OrderedList([]string{
			"First step",
			"Second step",
			"Third step with\nadditional details",
		}) + "\n"

	// Todo Lists
	todoCode := `tasks := []string{
    "Completed task",
    "Another completed task",
    "Pending task",
    "Another pending task",
}
completed := []bool{true, true, false, false}
md.TodoList(tasks, completed)`

	output += md.H3("Todo Lists") +
		md.LanguageCodeBlock("go", todoCode) +
		"\nResult:\n" +
		md.TodoList(
			[]string{
				"Completed task",
				"Another completed task",
				"Pending task",
				"Another pending task",
			},
			[]bool{true, true, false, false},
		) + "\n"

	// Nested Lists
	nestedCode := `md.BulletItem("Parent item") +
md.Nested(
    md.BulletItem("Nested item 1") +
    md.BulletItem("Nested item 2") +
    md.Nested(
        md.BulletItem("Deeply nested"),
        "    ",
    ),
    "  ",
)`

	nestedResult := md.BulletItem("Parent item") +
		md.Nested(
			md.BulletItem("Nested item 1")+
				md.BulletItem("Nested item 2")+
				md.Nested(
					md.BulletItem("Deeply nested"),
					"    ",
				),
			"  ",
		)

	output += md.H3("Nested Lists") +
		md.LanguageCodeBlock("go", nestedCode) +
		"\nResult:\n" +
		nestedResult + "\n"

	return output
}

func renderLinksAndImages() string {
	return md.H2("Links and Images") + "\n" +
		// Regular Links
		renderExample(
			"Regular Links",
			`md.Link("Gno Homepage", "https://gno.land")`,
			md.Link("Gno Homepage", "https://gno.land"),
		) +
		// User Links
		renderExample(
			"User Links",
			`md.UserLink("moul")`,
			md.UserLink("moul"),
		) +
		renderExample(
			"",
			`md.UserLink("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")`,
			md.UserLink("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
		) +
		// Images
		md.H3("Images") +
		md.LanguageCodeBlock("go", `md.Image("Gno Logo", "/public/imgs/gnoland.svg")`) +
		"\nResult:\n" +
		md.Image("Gno Logo", "/public/imgs/gnoland.svg") + "\n" +
		// Clickable Images
		md.H3("Clickable Images") +
		md.LanguageCodeBlock("go", `md.InlineImageWithLink("Click me!", "/public/imgs/gnoland.svg", "https://gno.land")`) +
		"\nResult:\n" +
		md.InlineImageWithLink("Click me!", "/public/imgs/gnoland.svg", "https://gno.land") + "\n"
}

func renderCode() string {
	return md.H2("Code") + "\n" +
		// Inline Code
		renderExample(
			"Inline Code",
			`"Use " + md.InlineCode("gno test") + " to run tests"`,
			"Use "+md.InlineCode("gno test")+" to run tests",
		) +
		// Simple Code Block
		md.H3("Simple Code Block") +
		md.LanguageCodeBlock("go", `md.CodeBlock("func main() {\n    println(\"Hello, Gno!\")\n}")`) +
		"\nResult:\n" +
		md.CodeBlock("func main() {\n    println(\"Hello, Gno!\")\n}") + "\n" +
		// Language Code Block
		md.H3("Language-specific Code Block") +
		md.LanguageCodeBlock("go", `md.LanguageCodeBlock("go", "package main\n\nfunc main() {\n    println(\"Hello!\")\n}")`) +
		"\nResult:\n" +
		md.LanguageCodeBlock("go", "package main\n\nfunc main() {\n    println(\"Hello!\")\n}") + "\n"
}

func renderBlockquotes() string {
	return md.H2("Blockquotes") +
		md.LanguageCodeBlock("go", `md.Blockquote("This is an important note.")`) +
		"\nResult:\n" +
		md.Blockquote("This is an important note.") + "\n" +
		md.LanguageCodeBlock("go", `md.Blockquote("Multi-line quotes\nare also supported\nwith proper formatting.")`) +
		"\nResult:\n" +
		md.Blockquote("Multi-line quotes\nare also supported\nwith proper formatting.") + "\n"
}

func renderAdvancedFeatures() string {
	output := md.H2("Advanced Features") + "\n"

	// Collapsible Sections
	output += md.H3("Collapsible Sections") +
		md.LanguageCodeBlock("go", `md.CollapsibleSection("Click to expand", "Hidden content here!")`) +
		"\nResult:\n" +
		md.CollapsibleSection("Click to expand", "Hidden content here!") + "\n"

	// Two Columns
	columnsCode := `md.Columns([]string{
    "Left column content",
    "Right column content",
}, false)`

	output += md.H3("Two Columns") +
		md.LanguageCodeBlock("go", columnsCode) +
		"\nResult:\n" +
		md.Columns([]string{
			"Left column content",
			"Right column content",
		}, false) + "\n"

	// Multiple Columns
	multiColumnsCode := `md.ColumnsN([]string{
    "Item 1", "Item 2", "Item 3",
    "Item 4", "Item 5", "Item 6",
}, 3, true)`

	output += md.H3("Multiple Columns (3 per row)") +
		md.LanguageCodeBlock("go", multiColumnsCode) +
		"\nResult:\n" +
		md.ColumnsN([]string{
			"Item 1", "Item 2", "Item 3",
			"Item 4", "Item 5", "Item 6",
		}, 3, true) + "\n"

	return output
}

func renderUtilityFunctions() string {
	return md.H2("Utility Functions") + "\n" +
		// Horizontal Rule
		renderExampleWithRawResult(
			"Horizontal Rule",
			`md.HorizontalRule()`,
			md.HorizontalRule(),
		) +
		// Footnotes
		md.H3("Footnotes") +
		md.LanguageCodeBlock("go", `"This has a footnote[1].\n\n" + md.Footnote("1", "Footnote content here.")`) +
		"\nResult:\n" +
		"This has a footnote[1].\n\n" +
		md.Footnote("1", "Footnote content here.") + "\n" +
		// Paragraphs
		renderExampleWithRawResult(
			"Paragraphs",
			`md.Paragraph("This ensures proper spacing.")`,
			md.Paragraph("This ensures proper spacing."),
		)
}

func renderPracticalExample() string {
	code := `// Example function that generates a rich user profile with balanced columns
func RenderProfile(username, name, bio string, avatar string, tasks []string, completed []bool) string {
    // First row: Avatar/Basic Info | Bio
    avatarSection := md.Image(name + " avatar", avatar) + "\n\n" +
        md.H3("Basic Info") +
        md.BulletList([]string{
            md.Bold("Name:") + " " + name,
            md.Bold("Username:") + " " + md.UserLink(username),
            md.Bold("Status:") + " 🟢 Active",
            md.Bold("Joined:") + " January 2024",
        })
    
    bioSection := md.H3("Bio") +
        md.Blockquote(bio)
    
    // Second row: Tasks | Links
    tasksSection := md.H3("Current Tasks") +
        md.TodoList(tasks, completed)
    
    linksSection := md.H3("Links") +
        md.BulletList([]string{
            md.Link("GitHub", "https://github.com/" + username),
            md.Link("Portfolio", "https://example.com/" + username),
            md.Link("LinkedIn", "https://linkedin.com/in/" + username),
        })
    
    // Combine with main title and two sets of columns
    return md.H1("User Profile: " + name) +
        md.HorizontalRule() +
        md.Columns([]string{avatarSection, bioSection}, true) +
        "\n" +
        md.Columns([]string{tasksSection, linksSection}, true)
}

// Usage:
profile := RenderProfile(
    "johndoe",
    "John Doe", 
    "Passionate Gno developer building the future of smart contracts. Love working with blockchain technology and contributing to open source.",
    "/public/imgs/gnoland.svg",
    []string{"Complete Gno tutorial", "Build first realm", "Deploy to testnet", "Write documentation"},
    []bool{true, true, false, false},
)`

	// Generate actual result - First row
	avatarSection := md.Image("John Doe avatar", "/public/imgs/gnoland.svg") + "\n\n" +
		md.H3("Basic Info") +
		md.BulletList([]string{
			md.Bold("Name:") + " John Doe",
			md.Bold("Username:") + " " + md.UserLink("johndoe"),
			md.Bold("Status:") + " 🟢 Active",
			md.Bold("Joined:") + " January 2024",
		})

	bioSection := md.H3("Bio") +
		md.Blockquote("Passionate Gno developer building the future of smart contracts. Love working with blockchain technology and contributing to open source.")

	// Second row
	tasksSection := md.H3("Current Tasks") +
		md.TodoList(
			[]string{"Complete Gno tutorial", "Build first realm", "Deploy to testnet", "Write documentation"},
			[]bool{true, true, false, false},
		)

	linksSection := md.H3("Links") +
		md.BulletList([]string{
			md.Link("GitHub", "https://github.com/johndoe"),
			md.Link("Portfolio", "https://example.com/johndoe"),
			md.Link("LinkedIn", "https://linkedin.com/in/johndoe"),
		})

	result := md.H1("User Profile: John Doe") +
		md.HorizontalRule() +
		md.Columns([]string{avatarSection, bioSection}, true) +
		"\n" +
		md.Columns([]string{tasksSection, linksSection}, true)

	return md.H2("Practical Example") +
		md.Paragraph("Here's a complete example showing how to build a rich user profile page using columns, images, and various md features:") +
		"\n" +
		md.LanguageCodeBlock("go", code) +
		"\nResult when called with the above parameters:\n" +
		result + "\n"
}

func renderBestPractices() string {
	practices := []string{
		"Prioritize code readability over performance in Render() - use clear variable names, logical grouping, and readable concatenation for humans",
		"Use the md package for all programmatic markdown generation in Render() methods",
		"Combine functions for complex formatting: " + md.InlineCode("md.Bold(md.Italic(\"text\"))"),
		"Use " + md.InlineCode("md.Paragraph()") + " to ensure proper spacing between elements",
		"Leverage " + md.InlineCode("md.ColumnsN()") + " for responsive layouts",
	}

	return md.H2("Best Practices") +
		md.OrderedList(practices) +
		"\n"
}

func renderSeeAlso() string {
	links := []string{
		md.Link("gno.land/p/moul/md", "https://gno.land/p/moul/md") + " - The md package source code",
		md.Link("Markdown on Gno", "https://gno.land/r/docs/markdown") + " - Comprehensive guide on Gno Flavored Markdown syntax",
	}

	return md.H2("See Also") +
		md.BulletList(links)
}

// Helper function to render an example with code and result
func renderExample(title, code, result string) string {
	output := ""
	if title != "" {
		output += md.H3(title)
	}
	output += md.LanguageCodeBlock("go", code) + "\n"
	output += md.Paragraph("Result: "+result) + "\n"
	return output
}

// Helper function for examples where result needs raw output
func renderExampleWithRawResult(title, code, result string) string {
	output := md.H3(title)
	output += md.LanguageCodeBlock("go", code) + "\n"
	output += "Result:\n" + result + "\n"
	return output
}
