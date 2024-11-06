package mdui

import (
	"strconv"
	"strings"
)

type Builder struct {
	elements []string
}

// NewBuilder creates and returns a new Builder instance
func NewBuilder() *Builder {
	return &Builder{
		elements: []string{},
	}
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
func (b *Builder) AddList(items []string, ordered bool) *Builder {
	list := List(items, ordered)
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
func (b *Builder) AddNavbar(links map[string]string) *Builder {
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

// AddEmoji adds an emoji to the Builder
func (b *Builder) AddEmoji(name string) *Builder {
	emoji := Emoji(name)
	b.elements = append(b.elements, emoji)
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

// BreakLine generates a Markdown line break (two spaces followed by a newline)
func BreakLine() string {
	return "  \n" // Two spaces followed by a newline
}

// IfElseRender generates different Markdown output based on a condition
// If the condition is true, it will render the "ifTrue" content.
// Otherwise, it renders the "ifFalse" content.
func IfElseRender(condition bool, ifTrue string, ifFalse string) string {
	if condition {
		return ifTrue
	} else {
		return ifFalse
	}
}

// Navbar generates a Markdown navigation menu from a map of links
func Navbar(links map[string]string) string {
	nav := ""
	for text, href := range links {
		if nav != "" {
			nav += " | " // Add a pipe separator between links
		}
		nav += "[" + text + "](" + href + ")"
	}
	return nav + "\n"
}

// Heading creates a Markdown heading based on the level
func Heading(level int, text string) string {
	if level < 1 || level > 6 {
		level = 1
	}
	headingPrefix := strings.Repeat("#", level)
	return headingPrefix + " " + text + "\n"
}

// Button generates a Markdown link styled as a button
func Button(text, href string) string {
	return "[" + text + "](" + href + ")"
}

// Image generates Markdown for an image
func Image(src, alt string) string {
	return "![" + alt + "](" + src + ")"
}

// CodeBlock wraps code in Markdown code block syntax
func CodeBlock(code string) string {
	return "```go\n" + code + "\n```\n"
}

// Divider renders a Markdown horizontal rule
func Divider() string {
	return "---\n"
}

// Paragraph formats a text paragraph in Markdown
func Paragraph(text string) string {
	return text + "\n"
}

// Quote generates a Markdown blockquote
func Quote(text string) string {
	return "> " + text + "\n"
}

// List generates a Markdown list (ordered or unordered)
func List(items []string, ordered bool) string {
	list := ""
	for i, item := range items {
		if ordered {
			list += strconv.Itoa(i+1) + ". " + item + "\n"
		} else {
			list += "- " + item + "\n"
		}
	}
	return list
}

// Link generates a Markdown link
func Link(text, href string) string {
	return "[" + text + "](" + href + ")"
}

// Table generates a Markdown table
func Table(headers []string, rows [][]string) string {
	table := "| " + strings.Join(headers, " | ") + " |\n"
	table += "|" + strings.Repeat("---|", len(headers)) + "\n"
	for _, row := range rows {
		table += "| " + strings.Join(row, " | ") + " |\n"
	}
	return table
}

// Bold formats text in bold
func Bold(text string) string {
	return "**" + text + "**"
}

// Italic formats text in italic
func Italic(text string) string {
	return "_" + text + "_"
}

// Strikethrough adds a strikethrough to the text
func Strikethrough(text string) string {
	return "~~" + text + "~~"
}

// Alert creates a Markdown-styled alert block with a specified type (info, warning, danger)
func Alert(content, alertType string) string {
	var prefix string
	switch alertType {
	case "info":
		prefix = "**ℹ️ Info:** " // Info icon
	case "warning":
		prefix = "**⚠️ Warning:** " // Warning icon
	case "danger":
		prefix = "**❌ Danger:** " // Danger icon
	default:
		prefix = "**ℹ️ Info:** " // Default to info icon
	}
	return "> " + prefix + content + "\n"
}

// Collapsible creates a collapsible section with a title and content
func Collapsible(title, content string) string {
	return "<details>\n<summary>" + title + "</summary>\n\n" + content + "\n\n</details>\n"
}

// Footnote generates a Markdown footnote
func Footnote(label, text string) string {
	return "[^" + label + "]: " + text + "\n"
}

// Badge generates a Markdown badge (often used in documentation)
func Badge(label, color string) string {
	return "![](https://img.shields.io/badge/" + strings.ReplaceAll(label, " ", "%20") + "-" + color + ")"
}

// TableOfContents generates a simple table of contents based on an array of headings
func TableOfContents(headings []string) string {
	toc := "## Table of Contents\n"
	for i, heading := range headings {
		toc += strconv.Itoa(i+1) + ". [" + heading + "](#" + strings.ToLower(strings.ReplaceAll(heading, " ", "-")) + ")\n"
	}
	return toc
}

// KeyboardShortcut formats a keyboard shortcut in Markdown using code block style
func KeyboardShortcut(keys ...string) string {
	return "`" + strings.Join(keys, " + ") + "`"
}

// Emoji adds an emoji to the content using a shortcode
func Emoji(name string) string {
	emojiMap := map[string]string{
		"smile":     "😊",
		"heart":     "❤️",
		"thumbs_up": "👍",
		// Add more emoji mappings as needed
	}
	if emoji, exists := emojiMap[name]; exists {
		return emoji
	}
	return name
}

// BlockquoteWithCitation generates a Markdown blockquote with an optional citation
func BlockquoteWithCitation(quote, citation string) string {
	if citation != "" {
		return "> " + quote + "\n> \n> — " + citation + "\n"
	}
	return "> " + quote + "\n"
}

// BadgeWithIcon generates a Markdown badge with an icon using Shields.io
func BadgeWithIcon(label, color, icon string) string {
	return "![](https://img.shields.io/badge/" + icon + "-" + strings.ReplaceAll(label, " ", "%20") + "-" + color + "?style=flat&logo=" + icon + ")"
}
