// Package md provides helper functions for generating Markdown content programmatically.
//
// It includes utilities for text formatting, creating lists, blockquotes, code blocks,
// links, images, and more.
//
// Highlights:
// - Supports basic Markdown syntax such as bold, italic, strikethrough, headers, and lists.
// - Manages multiline support in lists (e.g., bullet, ordered, and todo lists).
// - Includes advanced helpers like inline images with links and nested list prefixes.
//
// For a comprehensive example of how to use these helpers, see:
// https://gno.land/r/docs/moul_md
package md

import (
	"strconv"
	"strings"
)

// Bold returns bold text for markdown.
// Example: Bold("foo") => "**foo**"
func Bold(text string) string {
	return "**" + text + "**"
}

// Italic returns italicized text for markdown.
// Example: Italic("foo") => "*foo*"
func Italic(text string) string {
	return "*" + text + "*"
}

// Strikethrough returns strikethrough text for markdown.
// Example: Strikethrough("foo") => "~~foo~~"
func Strikethrough(text string) string {
	return "~~" + text + "~~"
}

// H1 returns a level 1 header for markdown.
// Example: H1("foo") => "# foo\n"
func H1(text string) string {
	return "# " + text + "\n"
}

// H2 returns a level 2 header for markdown.
// Example: H2("foo") => "## foo\n"
func H2(text string) string {
	return "## " + text + "\n"
}

// H3 returns a level 3 header for markdown.
// Example: H3("foo") => "### foo\n"
func H3(text string) string {
	return "### " + text + "\n"
}

// H4 returns a level 4 header for markdown.
// Example: H4("foo") => "#### foo\n"
func H4(text string) string {
	return "#### " + text + "\n"
}

// H5 returns a level 5 header for markdown.
// Example: H5("foo") => "##### foo\n"
func H5(text string) string {
	return "##### " + text + "\n"
}

// H6 returns a level 6 header for markdown.
// Example: H6("foo") => "###### foo\n"
func H6(text string) string {
	return "###### " + text + "\n"
}

// BulletList returns a bullet list for markdown.
// Example: BulletList([]string{"foo", "bar"}) => "- foo\n- bar\n"
func BulletList(items []string) string {
	var sb strings.Builder
	for _, item := range items {
		sb.WriteString(BulletItem(item))
	}
	return sb.String()
}

// BulletItem returns a bullet item for markdown.
// Example: BulletItem("foo") => "- foo\n"
func BulletItem(item string) string {
	var sb strings.Builder
	lines := strings.Split(item, "\n")
	sb.WriteString("- " + lines[0] + "\n")
	for _, line := range lines[1:] {
		sb.WriteString("  " + line + "\n")
	}
	return sb.String()
}

// OrderedList returns an ordered list for markdown.
// Example: OrderedList([]string{"foo", "bar"}) => "1. foo\n2. bar\n"
func OrderedList(items []string) string {
	var sb strings.Builder
	for i, item := range items {
		lines := strings.Split(item, "\n")
		sb.WriteString(strconv.Itoa(i+1) + ". " + lines[0] + "\n")
		for _, line := range lines[1:] {
			sb.WriteString("   " + line + "\n")
		}
	}
	return sb.String()
}

// TodoList returns a list of todo items with checkboxes for markdown.
// Example: TodoList([]string{"foo", "bar\nmore bar"}, []bool{true, false}) => "- [x] foo\n- [ ] bar\n  more bar\n"
func TodoList(items []string, done []bool) string {
	var sb strings.Builder
	for i, item := range items {
		sb.WriteString(TodoItem(item, done[i]))
	}
	return sb.String()
}

// TodoItem returns a todo item with checkbox for markdown.
// Example: TodoItem("foo", true) => "- [x] foo\n"
func TodoItem(item string, done bool) string {
	var sb strings.Builder
	checkbox := " "
	if done {
		checkbox = "x"
	}
	lines := strings.Split(item, "\n")
	sb.WriteString("- [" + checkbox + "] " + lines[0] + "\n")
	for _, line := range lines[1:] {
		sb.WriteString("  " + line + "\n")
	}
	return sb.String()
}

// Nested prefixes each line with a given prefix, enabling nested lists.
// Example: Nested("- foo\n- bar", "  ") => "  - foo\n  - bar\n"
func Nested(content, prefix string) string {
	lines := strings.Split(content, "\n")
	for i := range lines {
		if strings.TrimSpace(lines[i]) != "" {
			lines[i] = prefix + lines[i]
		}
	}
	return strings.Join(lines, "\n")
}

// Blockquote returns a blockquote for markdown.
// Example: Blockquote("foo\nbar") => "> foo\n> bar\n"
func Blockquote(text string) string {
	lines := strings.Split(text, "\n")
	var sb strings.Builder
	for _, line := range lines {
		sb.WriteString("> " + line + "\n")
	}
	return sb.String()
}

// InlineCode returns inline code for markdown.
// Example: InlineCode("foo") => "`foo`"
func InlineCode(code string) string {
	return "`" + strings.ReplaceAll(code, "`", "\\`") + "`"
}

// CodeBlock creates a markdown code block.
// Example: CodeBlock("foo") => "```\nfoo\n```"
func CodeBlock(content string) string {
	return "```\n" + strings.ReplaceAll(content, "```", "\\```") + "\n```"
}

// LanguageCodeBlock creates a markdown code block with language-specific syntax highlighting.
// Example: LanguageCodeBlock("go", "foo") => "```go\nfoo\n```"
func LanguageCodeBlock(language, content string) string {
	return "```" + language + "\n" + strings.ReplaceAll(content, "```", "\\```") + "\n```"
}

// HorizontalRule returns a horizontal rule for markdown.
// Example: HorizontalRule() => "---\n"
func HorizontalRule() string {
	return "---\n"
}

// Link returns a hyperlink for markdown.
// Example: Link("foo", "http://example.com") => "[foo](http://example.com)"
func Link(text, url string) string {
	return "[" + EscapeText(text) + "](" + url + ")"
}

// UserLink returns a user profile link for markdown.
// For usernames, it adds @ prefix to the display text.
// Example: UserLink("moul") => "[@moul](/u/moul)"
// Example: UserLink("g1blah") => "[g1blah](/u/g1blah)"
func UserLink(user string) string {
	if strings.HasPrefix(user, "g1") {
		return "[" + EscapeText(user) + "](/u/" + user + ")"
	}
	return "[@" + EscapeText(user) + "](/u/" + user + ")"
}

// InlineImageWithLink creates an inline image wrapped in a hyperlink for markdown.
// Example: InlineImageWithLink("alt text", "image-url", "link-url") => "[![alt text](image-url)](link-url)"
func InlineImageWithLink(altText, imageUrl, linkUrl string) string {
	return "[" + Image(altText, imageUrl) + "](" + linkUrl + ")"
}

// Image returns an image for markdown.
// Example: Image("foo", "http://example.com") => "![foo](http://example.com)"
func Image(altText, url string) string {
	return "![" + EscapeText(altText) + "](" + url + ")"
}

// Footnote returns a footnote for markdown.
// Example: Footnote("foo", "bar") => "[foo]: bar"
func Footnote(reference, text string) string {
	return "[" + EscapeText(reference) + "]: " + text
}

// Paragraph wraps the given text in a Markdown paragraph.
// Example: Paragraph("foo") => "foo\n"
func Paragraph(content string) string {
	return content + "\n\n"
}

// CollapsibleSection creates a collapsible section for markdown using
// HTML <details> and <summary> tags.
// Example:
// CollapsibleSection("Click to expand", "Hidden content")
// =>
// <details><summary>Click to expand</summary>
//
// Hidden content
// </details>
func CollapsibleSection(title, content string) string {
	return "<details><summary>" + EscapeText(title) + "</summary>\n\n" + content + "\n</details>\n"
}

// EscapeText escapes special Markdown characters in regular text where needed.
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
		".", `\.`,
		"!", `\!`,
		"`", "\\`",
	)
	return replacer.Replace(text)
}

// Columns returns a formatted row of columns using the Gno syntax.
// If you want a specific number of columns per row (<=4), use ColumnsN.
// Check /r/docs/markdown#columns for more info.
// If padded=true & the final <gno-columns> tag is missing column content, an empty
// column element will be placed to keep the cols per row constant.
// Padding works only with colsPerRow > 0.
func Columns(contentByColumn []string, padded bool) string {
	if len(contentByColumn) == 0 {
		return ""
	}
	maxCols := 4
	if padded && len(contentByColumn)%maxCols != 0 {
		missing := maxCols - len(contentByColumn)%maxCols
		contentByColumn = append(contentByColumn, make([]string, missing)...)
	}

	var sb strings.Builder
	sb.WriteString("<gno-columns>\n")

	for i, column := range contentByColumn {
		if i > 0 {
			sb.WriteString("|||\n")
		}
		sb.WriteString(column + "\n")
	}

	sb.WriteString("</gno-columns>\n")
	return sb.String()
}

const maxColumnsPerRow = 4

// ColumnsN splits content into multiple rows of N columns each and formats them.
// If colsPerRow <= 0, all items are placed in one <gno-columns> block.
// If padded=true & the final <gno-columns> tag is missing column content, an empty
// column element will be placed to keep the cols per row constant.
// Padding works only with colsPerRow > 0.
// Note: On standard-size screens, gnoweb handles a max of 4 cols per row.
func ColumnsN(content []string, colsPerRow int, padded bool) string {
	if len(content) == 0 {
		return ""
	}
	if colsPerRow <= 0 {
		return Columns(content, padded)
	}

	var sb strings.Builder
	// Case 2: Multiple blocks with max 4 columns
	for i := 0; i < len(content); i += colsPerRow {
		end := i + colsPerRow
		if end > len(content) {
			end = len(content)
		}
		row := content[i:end]

		// Add padding if needed
		if padded && len(row) < colsPerRow {
			row = append(row, make([]string, colsPerRow-len(row))...)
		}

		sb.WriteString(Columns(row, false))
	}
	return sb.String()
}

// FormInputType defines the type of a form input element.
// It is used to differentiate between types such as text, number, email, etc.
type FormInputType string

const (
	// FormText represents a plain text input.
	// Example: "hello"
	FormText FormInputType = "text"

	// FormNumber represents a numeric input.
	// Example: 42
	FormNumber = "number"

	// FormEmail represents an input that expects an email address.
	// Must match regex: '^[\w-\.]+@([\w-]+\.)+[\w-]{2,4}$'
	FormEmail = "email"

	// FormTel represents a telephone number input.
	// Example: "+33 6 12 34 56 78"
	FormTel = "tel"

	// FormPassword represents a password input.
	// The value is hidden from display.
	FormPassword = "password"
)

// Used to check for the validity of a FormInput in O(n)
var validFormInputSet = map[FormInputType]struct{}{
	FormText: {}, FormNumber: {}, FormEmail: {}, FormTel: {}, FormPassword: {},
}

// FormInput represents a single form input field.
// Name is the input name, PlaceHolder is the placeholder text,
// and Type determines the input type (text, number, etc.).
type FormInput struct {
	Name        string
	PlaceHolder string
	Type        FormInputType
}

// Form represents an entire form block with a path and multiple input fields.
// The form can be rendered as a string in a structured format.
type Form struct {
	Path   string
	Inputs []FormInput
}

func escapeValue(text string) string {
	replacer := strings.NewReplacer(
		`<`, "&#x3c",
		`=`, "&#x3d",
		`>`, "&#x3e",
		`"`, "&#x22",
		`\`, "&#x5c",
		`/`, "&#x2f",
	)
	return replacer.Replace(text)
}

// SetPath defines the form submission path.
// This path determines the endpoint to which the form will send data.
// If unset, the form uses the current page path.
func (f *Form) SetPath(path string) {
	f.Path = path
}

// AddInput adds a generic input to the form.
// The inputType can be as follows:
// - `FormText` - adds a text input field to the form.
// - `FormNumber` - adds a numeric input field to the form.
// - `FormEmail` - adds an email input field to the form.
// - `FormTel` - adds a telephone number input field to the form.
// - `FormPassword` - adds a password input field to the form.
func (f *Form) AddInput(inputType FormInputType, name, placeholder string) {
	f.Inputs = append(f.Inputs, FormInput{
		Name:        name,
		PlaceHolder: placeholder,
		Type:        inputType,
	})
}

// String returns the form as a Gno-formatted string.
// It generates the <gno-form> block with all included inputs.
// Example output:
// <gno-form path="/submit">
//
//	<gno-input name="email" placeholder="Your email" type="email" />
//
// </gno-form>
func (f Form) String() string {
	// Check if no inputs
	if len(f.Inputs) == 0 {
		return ""
	}
	var sb strings.Builder
	
	// Write the begin of the form
	str := "<gno-form"
	if f.Path != "" {
		str += " path=\"" + escapeValue(f.Path) + "\""
	}
	str += ">\n"
	sb.WriteString(str)

	// Generate input lines
	for _, v := range f.Inputs {
		// Use text by default
		x := v.Type
		if _, ok := validFormInputSet[v.Type]; !ok {
			x = FormText
		}

		// Write the line
		sb.WriteString(`  <gno-input name="` +
			escapeValue(v.Name) +
			`" type="` +
			string(x) +
			`" placeholder="` +
			escapeValue(v.PlaceHolder) +
			`" />` + "\n",
		)
	}

	// Write the end of the form
	sb.WriteString("</gno-form>\n")
	return sb.String()
}
