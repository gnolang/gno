// Package mictofun provides a fluent, chainable API for generating Markdown content.
//
// Unlike other markdown libraries that use separate functions, mictofun uses method
// chaining to create readable and composable markdown documents.
//
// Example usage:
//
//	import "gno.land/p/demo/ui/mictofun"
//
//	func Render(path string) string {
//	    return mictofun.New().
//	        H1("Welcome").
//	        P("This is a " + mictofun.Bold("powerful") + " markdown library.").
//	        H2("Features").
//	        Bullet("Fluent API", "Chainable methods", "Component composition").
//	        HR().
//	        P("Created by " + mictofun.UserLink("mictofun")).
//	        String()
//	}
package mictofun

import (
	"strconv"
	"strings"
)

// Doc is the main markdown document builder with fluent API.
type Doc struct {
	parts []string
}

// New creates a new markdown document builder.
func New() *Doc {
	return &Doc{
		parts: make([]string, 0),
	}
}

// String returns the final markdown string.
func (d *Doc) String() string {
	return strings.Join(d.parts, "\n")
}

// write appends content to the document.
func (d *Doc) write(content string) *Doc {
	d.parts = append(d.parts, content)
	return d
}

// Add adds a component that implements fmt.Stringer interface.
func (d *Doc) Add(component interface{ String() string }) *Doc {
	return d.write(component.String())
}

// AddRaw adds raw markdown content directly.
func (d *Doc) AddRaw(markdown string) *Doc {
	return d.write(markdown)
}

// Ln adds one or more empty lines.
func (d *Doc) Ln(count ...int) *Doc {
	n := 1
	if len(count) > 0 && count[0] > 0 {
		n = count[0]
	}
	return d.write(strings.Repeat("\n", n))
}

// ============================================================================
// HEADINGS
// ============================================================================

// H1 adds a level 1 heading.
func (d *Doc) H1(text string) *Doc {
	return d.write("# " + text + "\n")
}

// H2 adds a level 2 heading.
func (d *Doc) H2(text string) *Doc {
	return d.write("## " + text + "\n")
}

// H3 adds a level 3 heading.
func (d *Doc) H3(text string) *Doc {
	return d.write("### " + text + "\n")
}

// H4 adds a level 4 heading.
func (d *Doc) H4(text string) *Doc {
	return d.write("#### " + text + "\n")
}

// H5 adds a level 5 heading.
func (d *Doc) H5(text string) *Doc {
	return d.write("##### " + text + "\n")
}

// H6 adds a level 6 heading.
func (d *Doc) H6(text string) *Doc {
	return d.write("###### " + text + "\n")
}

// ============================================================================
// TEXT FORMATTING (Inline helpers - return string for embedding)
// ============================================================================

// Bold returns bold formatted text.
func Bold(text string) string {
	return "**" + text + "**"
}

// Italic returns italic formatted text.
func Italic(text string) string {
	return "*" + text + "*"
}

// Strike returns strikethrough formatted text.
func Strike(text string) string {
	return "~~" + text + "~~"
}

// Code returns inline code formatted text.
func Code(text string) string {
	return "`" + strings.ReplaceAll(text, "`", "\\`") + "`"
}

// ============================================================================
// PARAGRAPH & TEXT BLOCKS
// ============================================================================

// P adds a paragraph.
func (d *Doc) P(text string) *Doc {
	return d.write(text + "\n")
}

// Quote adds a blockquote. Handles multiline text.
func (d *Doc) Quote(text string) *Doc {
	lines := strings.Split(text, "\n")
	var sb strings.Builder
	for _, line := range lines {
		sb.WriteString("> " + line + "\n")
	}
	return d.write(sb.String())
}

// CodeBlock adds a fenced code block with optional language.
func (d *Doc) CodeBlock(content string, lang ...string) *Doc {
	language := ""
	if len(lang) > 0 {
		language = lang[0]
	}
	escaped := strings.ReplaceAll(content, "```", "\\`\\`\\`")
	return d.write("```" + language + "\n" + escaped + "\n```\n")
}

// HR adds a horizontal rule.
func (d *Doc) HR() *Doc {
	return d.write("\n---\n")
}

// ============================================================================
// LISTS
// ============================================================================

// Bullet adds a bullet list from items.
func (d *Doc) Bullet(items ...string) *Doc {
	var sb strings.Builder
	for _, item := range items {
		lines := strings.Split(item, "\n")
		sb.WriteString("- " + lines[0] + "\n")
		for _, line := range lines[1:] {
			sb.WriteString("  " + line + "\n")
		}
	}
	return d.write(sb.String())
}

// Numbered adds a numbered list from items.
func (d *Doc) Numbered(items ...string) *Doc {
	var sb strings.Builder
	for i, item := range items {
		lines := strings.Split(item, "\n")
		sb.WriteString(strconv.Itoa(i+1) + ". " + lines[0] + "\n")
		for _, line := range lines[1:] {
			sb.WriteString("   " + line + "\n")
		}
	}
	return d.write(sb.String())
}

// Todo adds a todo list with checkboxes.
func (d *Doc) Todo(items []string, done []bool) *Doc {
	var sb strings.Builder
	for i, item := range items {
		checkbox := " "
		if i < len(done) && done[i] {
			checkbox = "x"
		}
		lines := strings.Split(item, "\n")
		sb.WriteString("- [" + checkbox + "] " + lines[0] + "\n")
		for _, line := range lines[1:] {
			sb.WriteString("  " + line + "\n")
		}
	}
	return d.write(sb.String())
}

// NestedBullet adds items with a prefix for nested lists.
func (d *Doc) NestedBullet(prefix string, items ...string) *Doc {
	var sb strings.Builder
	for _, item := range items {
		sb.WriteString(prefix + "- " + item + "\n")
	}
	return d.write(sb.String())
}

// ============================================================================
// LINKS & IMAGES
// ============================================================================

// Link returns a markdown link.
func Link(text, url string) string {
	return "[" + EscapeText(text) + "](" + url + ")"
}

// Image returns a markdown image.
func Image(alt, url string) string {
	return "![" + EscapeText(alt) + "](" + url + ")"
}

// ImageWithLink returns an image wrapped in a link.
func ImageWithLink(alt, imageURL, linkURL string) string {
	return "[" + Image(alt, imageURL) + "](" + linkURL + ")"
}

// Img adds an image to the document.
func (d *Doc) Img(alt, url string) *Doc {
	return d.write(Image(alt, url) + "\n")
}

// ============================================================================
// GNO.LAND SPECIFIC HELPERS
// ============================================================================

// UserLink returns a link to a gno.land user profile.
// For usernames starting with "g1", displays the address.
// For other usernames, adds @ prefix.
func UserLink(user string) string {
	if strings.HasPrefix(user, "g1") {
		return "[" + EscapeText(user) + "](/u/" + user + ")"
	}
	return "[@" + EscapeText(user) + "](/u/" + user + ")"
}

// RealmLink returns a link to a gno.land realm.
func RealmLink(path string, text ...string) string {
	displayText := path
	if len(text) > 0 {
		displayText = text[0]
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return "[" + EscapeText(displayText) + "](" + path + ")"
}

// ============================================================================
// UTILITIES
// ============================================================================

// EscapeText escapes special markdown characters in text.
func EscapeText(text string) string {
	replacer := strings.NewReplacer(
		`*`, `\*`,
		`_`, `\_`,
		`[`, `\[`,
		`]`, `\]`,
		`(`, `\(`,
		`)`, `\)`,
		`~`, `\~`,
		`>`, `\>`,
		`|`, `\|`,
		`-`, `\-`,
		`+`, `\+`,
		`.`, `\.`,
		`!`, `\!`,
		"`", "\\`",
	)
	return replacer.Replace(text)
}

// Footnote returns a footnote reference.
func Footnote(ref, text string) string {
	return "[" + EscapeText(ref) + "]: " + text
}
