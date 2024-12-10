package mdui

import (
	"strconv"
	"strings"
)

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

func Navbar(links []NavLink) string {
	nav := ""
	for i, link := range links {
		if i > 0 {
			nav += " | "
		}
		nav += "[" + link.Text + "](" + link.Href + ")"
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
	return "```\n" + code + "\n```\n"
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

// List generates a Markdown list (ordered or unordered) with each item optionally as a link
func List(items []string, hrefs []string, ordered bool) string {
	list := ""
	for i, item := range items {
		var listItem string
		if i < len(hrefs) && hrefs[i] != "" {
			// Create a link if an href is available
			listItem = "[" + item + "](" + hrefs[i] + ")"
		} else {
			// Use plain text if no corresponding href
			listItem = item
		}
		if ordered {
			list += strconv.Itoa(i+1) + ". " + listItem + "\n"
		} else {
			list += "- " + listItem + "\n"
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

// Alert creates a Markdown-styled alert block with a specified type or custom prefix
func Alert(content, alertType string) string {
	var prefix string
	switch alertType {
	case "info":
		prefix = "**ℹ️ Info:** " // Info icon
	case "warning":
		prefix = "**⚠️ Warning:** " // Warning icon
	case "danger":
		prefix = "**❌ Danger:** " // Danger icon
	case "success":
		prefix = "**✅ Success:** " // Success icon
	default:
		prefix = "**" + strings.Title(alertType) + ":** " // Custom prefix based on type
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
