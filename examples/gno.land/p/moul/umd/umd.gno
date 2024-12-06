package umd

import (
	"strconv"
	"strings"
)

// Example: Bold("foo") => "**foo**"
func Bold(text string) string {
	return "**" + text + "**"
}

// Example: Italic("foo") => "*foo*"
func Italic(text string) string {
	return "*" + text + "*"
}

// Example: Strikethrough("foo") => "~~foo~~"
func Strikethrough(text string) string {
	return "~~" + text + "~~"
}

// Example: H1("foo") => "# foo\n"
func H1(text string) string {
	return "# " + text + "\n"
}

// Example: H2("foo") => "## foo\n"
func H2(text string) string {
	return "## " + text + "\n"
}

// Example: H3("foo") => "### foo\n"
func H3(text string) string {
	return "### " + text + "\n"
}

// Example: BulletList([]string{"foo", "bar"}) => "- foo\n- bar\n"
func BulletList(items []string) string {
	var sb strings.Builder
	for _, item := range items {
		sb.WriteString("- " + EscapeText(item) + "\n")
	}
	return sb.String()
}

// Example: OrderedList([]string{"foo", "bar"}) => "1. foo\n2. bar\n"
func OrderedList(items []string) string {
	var sb strings.Builder
	for i, item := range items {
		sb.WriteString(strconv.Itoa(i+1) + ". " + EscapeText(item) + "\n")
	}
	return sb.String()
}

// Example: TodoList([]string{"foo", "bar"}, []bool{true, false}) => "- [x] foo\n- [ ] bar\n"
func TodoList(items []string, done []bool) string {
	var sb strings.Builder
	for i, item := range items {
		checkbox := " "
		if done[i] {
			checkbox = "x"
		}
		sb.WriteString("- [" + checkbox + "] " + EscapeText(item) + "\n")
	}
	return sb.String()
}

// Example: Blockquote("foo\nbar") => "> foo\n> bar\n"
func Blockquote(text string) string {
	lines := strings.Split(text, "\n")
	var sb strings.Builder
	for _, line := range lines {
		sb.WriteString("> " + line + "\n")
	}
	return sb.String()
}

// Example: InlineCode("foo") => "`foo`"
func InlineCode(code string) string {
	return "`" + strings.ReplaceAll(code, "`", "\\`") + "`"
}

// Example: CodeBlock("foo") => "```\nfoo\n```"
func CodeBlock(content string) string {
	return "```\n" + content + "\n```"
}

// Example: LanguageCodeBlock("go", "foo") => "```go\nfoo\n```"
func LanguageCodeBlock(language, content string) string {
	return "```" + language + "\n" + content + "\n```"
}

// Example: HorizontalRule() => "---\n"
func HorizontalRule() string {
	return "---\n"
}

// Example: Link("foo", "http://example.com") => "[foo](http://example.com)"
func Link(text, url string) string {
	return "[" + EscapeText(text) + "](" + EscapeText(url) + ")"
}

// Example: Image("foo", "http://example.com") => "![foo](http://example.com)"
func Image(altText, url string) string {
	return "![" + EscapeText(altText) + "](" + EscapeText(url) + ")"
}

// Example: Footnote("foo", "bar") => "[foo]: bar"
func Footnote(reference, text string) string {
	return "[" + EscapeText(reference) + "]: " + EscapeText(text)
}

// Example: Paragraph("foo") => "foo\n"
func Paragraph(content string) string {
	return content + "\n"
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
	)
	return replacer.Replace(text)
}
