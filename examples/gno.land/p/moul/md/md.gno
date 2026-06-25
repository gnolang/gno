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
//
// # Sanitization contract
//
// Some helpers in this package sanitize their user-derived arguments
// INTERNALLY (via p/nt/markdown/sanitize/v0). When using these, pass
// raw user input — do NOT pre-wrap with sanitize.*, or you will
// double-wrap (the escapers are not idempotent and double-wrap is a
// bug):
//
//	Link, UserLink, Image, InlineImageWithLink, FootnoteDefinition,
//	LinkReferenceDefinition, CollapsibleSection (title only),
//	InlineCode, CodeBlock, LanguageCodeBlock, Blockquote
//
// The other helpers DO NOT sanitize — they are pure builders that
// wrap their input in markdown chrome. User-derived input reaches
// the output unmodified, so callers MUST wrap with sanitize.* at the
// call site:
//
//	Bold, Italic, Strikethrough, H1-H6, BulletList, BulletItem,
//	OrderedList, TodoList, TodoItem, Nested, Paragraph, Columns,
//	ColumnsN, HorizontalRule
//
// Examples:
//
//	// Sanitizing helper — pass raw:
//	out += md.Link(post.Title, post.URL)                                 // good
//	out += md.Link(sanitize.InlineText(post.Title), sanitize.URL(post.URL)) // BAD: double-wrap
//
//	// Non-sanitizing helper — wrap once:
//	out += md.H2(sanitize.InlineText(post.Title))                        // good
//	out += md.H2(post.Title)                                             // BAD: raw user input
//	out += md.H2(sanitize.InlineText(sanitize.InlineText(post.Title)))   // BAD: double-wrap
//
// Composition: outputs of sanitizing helpers are safe markdown chrome
// and can be embedded inside non-sanitizing helpers freely:
//
//	out += md.H2(md.Link(post.Title, post.URL))   // good — H2 doesn't re-escape Link's output
//
// The reverse is unsafe: do NOT embed a non-sanitizing helper's
// markdown chrome inside a sanitizing helper's arg, or the inner
// markdown gets re-escaped:
//
//	out += md.Link(md.Bold(post.Title), post.URL) // BAD: Link's internal sanitize
//	                                              // escapes the ** chars from md.Bold
package md

import (
	"strconv"
	"strings"

	"gno.land/p/nt/markdown/sanitize/v0"
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

// Blockquote returns the text as a CommonMark blockquote.
// Example: Blockquote("foo\nbar") => "\n> foo\n> bar\n\n"
//
// Delegates to sanitize.Blockquote, which cleans the content (bidi-strip,
// CR/CRLF/U+2028/U+2029/NEL line-ending normalize, LRD strip, ref-link
// escape, block-marker escape, fence auto-close), line-prefixes each
// line with "> ", and wraps with "\n" / "\n\n" so the quote opens
// cleanly and cannot pull appended chrome into the quote via CM §5.2
// lazy continuation. Callers do NOT need to pre-wrap the input.
func Blockquote(text string) string {
	return sanitize.Blockquote(text)
}

// InlineCode wraps the given text as a CommonMark inline code span.
// Example: InlineCode("foo") => "`foo`"
//
// Delegates to sanitize.InlineCode, which cleans the input (bidi-strip,
// CR/CRLF + NEL + U+2028/U+2029 folded to single space, NUL→U+FFFD)
// and picks a backtick-run length that outscans any internal backticks.
// Callers do NOT need to pre-wrap the input.
func InlineCode(code string) string {
	return sanitize.InlineCode(code)
}

// CodeBlock creates a markdown code block.
// Example: CodeBlock("foo") => "```\nfoo\n```"
//
// Delegates to sanitize.CodeBlock, which cleans the content (bidi-strip,
// CR/CRLF normalize, NEL/U+2028/U+2029 fold, NUL→U+FFFD) and picks a
// fence wide enough to outscan any backticks in content. Callers do NOT
// need to pre-wrap content with another sanitize helper.
func CodeBlock(content string) string {
	return sanitize.CodeBlock(content)
}

// LanguageCodeBlock creates a markdown code block with language-specific syntax highlighting.
// Example: LanguageCodeBlock("go", "foo") => "```go\nfoo\n```"
//
// Delegates to sanitize.LanguageCodeBlock, which validates the language
// tag (charset ^[a-zA-Z0-9_+-]{1,32}$, falling back to a tagless fence
// if invalid) and cleans the content as CodeBlock does. Callers do NOT
// need to pre-wrap either argument.
func LanguageCodeBlock(language, content string) string {
	return sanitize.LanguageCodeBlock(language, content)
}

// HorizontalRule returns a horizontal rule for markdown.
// Example: HorizontalRule() => "---\n"
func HorizontalRule() string {
	return "---\n"
}

// Link returns a hyperlink for markdown.
// Example: Link("foo", "http://example.com") => "[foo](http://example.com)"
//
// The text and url args are sanitized internally — text via
// sanitize.InlineText, url via sanitize.URL (allowlists http/https/
// mailto/relative/fragment; rejects javascript:, data:, blob:, etc.).
// Callers do NOT need to pre-wrap either argument; double-wrapping is
// a bug (the inline-text escaper is non-idempotent).
//
// If the URL fails the scheme allowlist, the href is rendered empty —
// the link becomes inert rather than carrying a malicious destination.
func Link(text, url string) string {
	return "[" + sanitize.InlineText(text) + "](" + sanitize.URL(url) + ")"
}

// UserLink returns a user profile link for markdown.
// Example: UserLink("moul") => "[@moul](/u/moul)"
// Example: UserLink("g1blah...") => "[g1blah...](/u/g1blah...)"
//
// Validates the user identifier — if it matches the gno bech32 address
// pattern (g1...), produces an address-style link; otherwise tries the
// r/sys/users-charset username and produces an @-style link. Returns
// "" if the identifier matches neither — callers should treat "" as
// "skip the user mention" rather than emit a broken link.
func UserLink(user string) string {
	if addr := sanitize.BechString(user, "g"); addr != "" {
		return "[" + addr + "](/u/" + addr + ")"
	}
	if name := sanitize.UserName(user); name != "" {
		return "[@" + name + "](/u/" + name + ")"
	}
	return ""
}

// InlineImageWithLink creates an inline image wrapped in a hyperlink for markdown.
// Example: InlineImageWithLink("alt text", "image-url", "link-url") => "[![alt text](image-url)](link-url)"
//
// altText and imageUrl are sanitized via Image (sanitize.InlineText +
// sanitize.ImageURL); linkUrl is sanitized via sanitize.URL. Callers
// do NOT need to pre-wrap any argument.
func InlineImageWithLink(altText, imageUrl, linkUrl string) string {
	return "[" + Image(altText, imageUrl) + "](" + sanitize.URL(linkUrl) + ")"
}

// Image returns an image for markdown.
// Example: Image("foo", "http://example.com") => "![foo](http://example.com)"
//
// altText is sanitized via sanitize.InlineText, url via
// sanitize.ImageURL (allowlists http/https/relative + data:image/*;
// rejects mailto:, javascript:, data:text/html, etc.). Callers do NOT
// need to pre-wrap either argument.
func Image(altText, url string) string {
	return "![" + sanitize.InlineText(altText) + "](" + sanitize.ImageURL(url) + ")"
}

// FootnoteDefinition emits a GFM footnote definition — `[^name]: body` —
// for a footnote that is referenced elsewhere in the document by
// `[^name]`.
//
// Example: FootnoteDefinition("note1", "Long form of the citation.")
// renders as:
//
//	[^note1]:
//	    Long form of the citation.
//
// The `name` is validated as a FootnoteLabel (^[A-Za-z0-9_-]{1,64}$);
// `text` is user-supplied multi-paragraph prose, sanitized via Block.
// An invalid name or empty body returns "".
//
// Delegates to sanitize.FootnoteDefinition. Callers do NOT need to
// pre-wrap either argument.
func FootnoteDefinition(name, text string) string {
	return sanitize.FootnoteDefinition(name, text)
}

// LinkReferenceDefinition emits a CommonMark link reference definition
// (CM §4.7) — `[label]: url "title"` — for a reference link that is
// invoked elsewhere by `[text][label]` or by the shortcut form
// `[label]`.
//
// Example: LinkReferenceDefinition("r/docs/help", "/r/docs/help", "")
// renders as:
//
//	[r/docs/help]: /r/docs/help
//
// The `label` is validated as a FootnoteLabel (^[A-Za-z0-9_-]{1,64}$);
// `url` is sanitized via URL (allowlist); `title` is sanitized via
// LinkTitle. An invalid label or rejected URL returns "".
//
// Realms should choose a namespaced label using dashes
// (e.g. `r-myrealm-help`) so that shortcut-reference invocations from
// user content can't collide with bare words a user is likely to write.
// `/` is not in the FootnoteLabel charset.
//
// Delegates to sanitize.LinkReferenceDefinition. Callers do NOT need to
// pre-wrap any argument.
func LinkReferenceDefinition(label, url, title string) string {
	return sanitize.LinkReferenceDefinition(label, url, title)
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
//
// The title argument is sanitized via sanitize.HTMLEscape (it lands in
// an HTML element body inside <summary>, not a markdown context — so
// HTML entity escaping is the correct policy, not markdown backslash
// escaping). The content argument is passed through unchanged because
// <details> with a blank-line-separated body allows markdown inside
// (CM §4.6); callers must pre-wrap content with sanitize.Block if it
// derives from user input.
func CollapsibleSection(title, content string) string {
	return "<details><summary>" + sanitize.HTMLEscape(title) + "</summary>\n\n" + content + "\n</details>\n"
}

// EscapeURL escapes characters in a URL for use in markdown link syntax.
//
// Deprecated: use sanitize.URL (for link href) or sanitize.ImageURL
// (for image src) directly. EscapeURL previously only percent-encoded
// ( and ) and did not validate the URL scheme — a security footgun
// (EscapeURL("javascript:alert(1)") returned a working XSS payload).
// It now delegates to sanitize.URL, which allowlists schemes (rejects
// javascript:, data:text/html, vbscript:, blob:, etc.) and
// percent-encodes all unsafe bytes (including the ( ) this function
// handled). The other helpers in this package (Link, UserLink, Image,
// InlineImageWithLink) sanitize their URL args internally now, so you
// rarely need to call any URL-escape helper directly.
//
// Behavior change vs. the original implementation: invalid schemes
// now return "" instead of passing through with ( ) escaped. Non-ASCII
// bytes get percent-encoded to standard RFC 3986 wire form instead of
// passing through as raw UTF-8.
func EscapeURL(url string) string {
	return sanitize.URL(url)
}

// EscapeText escapes special Markdown characters in regular text for
// use in inline contexts.
//
// Deprecated: use sanitize.InlineText directly. EscapeText was
// INCOMPLETE — it missed \, #, <, & and did not strip bidi/zero-width
// characters, replace NUL with U+FFFD, or normalize line endings.
// User input could inject backslash escapes (\* cancels neighboring
// escapes), autolinks (<https://x>), raw HTML (<script>), HTML entity
// references (&amp;), and bidi spoofing (RLO before an address). It
// now delegates to sanitize.InlineText. The other helpers in this
// package (Link, UserLink, Image, InlineImageWithLink, CollapsibleSection)
// sanitize their text args internally now, so you rarely need to call
// any inline-text-escape helper directly.
//
// Behavior change vs. the original implementation: \, #, <, & are now
// escaped; | is no longer escaped (the original was over-escaping —
// outside GFM table-cell context, | is markdown-inert; for table
// cells use sanitize.TableCell which adds the | escape on top of the
// inline-text set). Bidi controls are stripped, NUL becomes U+FFFD,
// and line endings normalize.
func EscapeText(text string) string {
	return sanitize.InlineText(text)
}

// Columns returns a formatted row of columns using the Gno syntax.
// If you want a specific number of columns per row (<=4), use ColumnsN.
// Check /r/docs/markdown#columns for more info.
// If padded=true & the final <gno-columns> tag is missing column content, an empty
// column element will be placed to keep the cols per row constant.
// Padding works only with colsPerRow > 0.
//
// Example:
//
//	Columns([]string{"A", "B"}, false)
//	// Returns:
//	// <gno-columns>
//	// A
//	// <gno-columns-sep/>
//	// B
//	// </gno-columns>
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
			sb.WriteString("<gno-columns-sep/>\n")
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
//
// Example:
//
//	ColumnsN([]string{"A", "B", "C"}, 2, false)
//	// Returns:
//	// <gno-columns>
//	// A
//	// <gno-columns-sep/>
//	// B
//	// </gno-columns>
//	// <gno-columns>
//	// C
//	// </gno-columns>
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
