
package mdui

import (
	"strings"
)

// NewBuilder creates and returns a new Builder instance
func NewBuilder() *Builder {
	return &Builder{
		elements: []string{},
	}
}

// AddRaw adds raw Markdown content to the Builder
func (b *Builder) AddRaw(markdown string) *Builder {
	b.elements = append(b.elements, markdown)
	return b
}

// AddHeading adds a heading to the Builder
func (b *Builder) AddHeading(level int, text string) *Builder {
	heading := Heading(level, text)
	b.elements = append(b.elements, heading)
	return b
}

// AddParagraph adds a paragraph to the Builder
func (b *Builder) AddParagraph(text string) *Builder {
	paragraph := Paragraph(text)
	b.elements = append(b.elements, paragraph)
	return b
}

// AddButton adds a button (link) to the Builder
func (b *Builder) AddButton(text, href string) *Builder {
	button := Button(text, href)
	b.elements = append(b.elements, button)
	return b
}

// AddImage adds an image to the Builder
func (b *Builder) AddImage(src, alt string) *Builder {
	image := Image(src, alt)
	b.elements = append(b.elements, image)
	return b
}

// AddDivider adds a horizontal rule to the Builder
func (b *Builder) AddDivider() *Builder {
	divider := Divider()
	b.elements = append(b.elements, divider)
	return b
}

// AddList adds a list (ordered or unordered) to the Builder
func (b *Builder) AddList(items []string, hrefs []string, ordered bool) *Builder {
	list := List(items, hrefs, ordered)
	b.elements = append(b.elements, list)
	return b
}

// AddTable adds a table to the Builder
func (b *Builder) AddTable(headers []string, rows [][]string) *Builder {
	table := Table(headers, rows)
	b.elements = append(b.elements, table)
	return b
}

// AddNavbar adds a navigation bar to the Builder
func (b *Builder) AddNavbar(links []NavLink) *Builder {
	navbar := Navbar(links)
	b.elements = append(b.elements, navbar)
	return b
}

// AddQuote adds a blockquote to the Builder
func (b *Builder) AddQuote(text string) *Builder {
	quote := Quote(text)
	b.elements = append(b.elements, quote)
	return b
}

// AddBold adds bold text to the Builder
func (b *Builder) AddBold(text string) *Builder {
	bold := Bold(text)
	b.elements = append(b.elements, bold)
	return b
}

// AddItalic adds italic text to the Builder
func (b *Builder) AddItalic(text string) *Builder {
	italic := Italic(text)
	b.elements = append(b.elements, italic)
	return b
}

// AddStrikethrough adds strikethrough text to the Builder
func (b *Builder) AddStrikethrough(text string) *Builder {
	strikethrough := Strikethrough(text)
	b.elements = append(b.elements, strikethrough)
	return b
}

// AddCodeBlock adds a code block to the Builder
func (b *Builder) AddCodeBlock(code string) *Builder {
	codeBlock := CodeBlock(code)
	b.elements = append(b.elements, codeBlock)
	return b
}

// AddAlert adds an alert box with a specified type to the Builder
func (b *Builder) AddAlert(content, alertType string) *Builder {
	alert := Alert(content, alertType)
	b.elements = append(b.elements, alert)
	return b
}

// AddCollapsible adds a collapsible section to the Builder
func (b *Builder) AddCollapsible(title, content string) *Builder {
	collapsible := Collapsible(title, content)
	b.elements = append(b.elements, collapsible)
	return b
}

// AddFootnote adds a footnote to the Builder
func (b *Builder) AddFootnote(label, text string) *Builder {
	footnote := Footnote(label, text)
	b.elements = append(b.elements, footnote)
	return b
}

// AddBadge adds a badge to the Builder
func (b *Builder) AddBadge(label, color string) *Builder {
	badge := Badge(label, color)
	b.elements = append(b.elements, badge)
	return b
}

// AddBadgeWithIcon adds a badge with an icon to the Builder
func (b *Builder) AddBadgeWithIcon(label, color, icon string) *Builder {
	badgeWithIcon := BadgeWithIcon(label, color, icon)
	b.elements = append(b.elements, badgeWithIcon)
	return b
}

// AddKeyboardShortcut adds a keyboard shortcut to the Builder
func (b *Builder) AddKeyboardShortcut(keys ...string) *Builder {
	shortcut := KeyboardShortcut(keys...)
	b.elements = append(b.elements, shortcut)
	return b
}

// AddBlockquoteWithCitation adds a blockquote with an optional citation to the Builder
func (b *Builder) AddBlockquoteWithCitation(quote, citation string) *Builder {
	blockquoteWithCitation := BlockquoteWithCitation(quote, citation)
	b.elements = append(b.elements, blockquoteWithCitation)
	return b
}

// AddTableOfContents adds a table of contents to the Builder
func (b *Builder) AddTableOfContents(headings []string) *Builder {
	toc := TableOfContents(headings)
	b.elements = append(b.elements, toc)
	return b
}

// Render generates and returns the final Markdown content
func (b *Builder) Render() string {
	return strings.Join(b.elements, "\n")
}

// AddIfElseRender adds conditional content to the Builder based on a condition
func (b *Builder) AddIfElseRender(condition bool, ifTrue string, ifFalse string) *Builder {
	result := IfElseRender(condition, ifTrue, ifFalse)
	b.elements = append(b.elements, result)
	return b
}

func (b *Builder) AddBreakLine() *Builder {
	lineBreak := BreakLine()
	b.elements = append(b.elements, lineBreak)
	return b
}