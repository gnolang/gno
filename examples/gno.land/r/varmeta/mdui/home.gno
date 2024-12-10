package mdui

import (
	"gno.land/p/varmeta/mdui"
)

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
1. [Explore Components](/r/varmeta/mdui:components) - Discover a variety of components available in MDUI.
2. [Check out Examples](/r/varmeta/mdui:examples) - Learn how to build real-world projects with MDUI.
3. [Try It Out](/r/varmeta/mdui:examples) - See how easy it is to create your first page.
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

