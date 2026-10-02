package markdown

import (
	"bufio"
	"bytes"
	stdhtml "html"
	"strings"
	"unicode"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer/html"
)

const (
	// descriptionMaxRunes caps the summary where search results and link
	// previews cut it anyway.
	descriptionMaxRunes = 160
	// descriptionMinRunes skips a paragraph too short to summarise anything.
	// Index pages open on a date and dashboards on a one-word status, and
	// either as a summary is worse than none; measured against deployed
	// realms, a real one-line summary clears this.
	descriptionMinRunes = 40
	// titleMaxRunes caps a heading lifted into <title>, leaving room for the
	// domain before a search result cuts it.
	titleMaxRunes = 60
)

// RealmMeta is what rendering a document yields besides its HTML.
type RealmMeta struct {
	Toc         Toc
	Title       string
	Description string
}

// Lead returns what the top of a document says about it, as plain text: the
// title is a leading h1, and the description the first paragraph long enough
// to summarise between that h1, or the document start, and the next heading
// or rule. Text further down belongs to a section, a list or a post someone
// else wrote, not to the page, so it never names the page. A paragraph with
// no visible text, such as a banner image, does not count as leading.
func Lead(doc ast.Node, src []byte) (title, description string) {
	n := doc.FirstChild()
	for n != nil && n.Kind() == ast.KindParagraph && plainText(src, n) == "" {
		n = n.NextSibling()
	}
	if h, ok := n.(*ast.Heading); ok && h.Level == 1 {
		title = truncateRunes(plainText(src, n), titleMaxRunes)
		n = n.NextSibling()
	}
	for ; n != nil; n = n.NextSibling() {
		switch n.Kind() {
		case ast.KindHeading, ast.KindThematicBreak:
			return title, ""
		case ast.KindParagraph:
			if text := plainText(src, n); len([]rune(text)) >= descriptionMinRunes {
				return title, truncateRunes(text, descriptionMaxRunes)
			}
		}
	}
	return title, ""
}

// plainText is the visible text of n on one line.
func plainText(src []byte, n ast.Node) string {
	return oneLine(visibleText(src, n))
}

// oneLine collapses every run of whitespace in s to one space.
func oneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// visibleText is nodeText minus the parts a reader does not read. An image's
// alt text is the one that matters: it travels with the node but shows only
// when the image fails, so lifting it into the description would let a page
// publish a summary nobody sees, which is the abuse #3910 named.
func visibleText(src []byte, n ast.Node) string {
	var b strings.Builder
	var walk func(ast.Node)
	walk = func(n ast.Node) {
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			// An image nests under a link when it is the link's content, so
			// the skip has to follow the tree down, not just its first row.
			if c.Kind() == ast.KindImage {
				continue
			}
			if c.ChildCount() > 0 {
				walk(c)
				continue
			}
			text := nodeText(src, c)
			// goldmark resolves escapes and entities in text when it writes
			// HTML, but writes a code span as typed; the summary follows.
			if c.Kind() == ast.KindText && n.Kind() != ast.KindCodeSpan {
				text = resolveText(text)
			}
			b.Write(text)
		}
	}
	walk(n)
	return b.String()
}

// resolveText resolves backslash escapes and entity references the way goldmark
// does when it writes a text node, in one pass, so an escaped entity (`\&amp;`)
// stays literal as it does on the page. The writer emits HTML; unescaping it
// gives the plain text the template escapes again into the attribute.
func resolveText(text []byte) []byte {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	html.DefaultWriter.Write(w, text)
	_ = w.Flush()
	return []byte(stdhtml.UnescapeString(buf.String()))
}

// TruncateTitle caps an operator-written title at the length of one lifted
// from a page's own h1, so both cut in the same place.
func TruncateTitle(s string) string { return truncateLine(s, titleMaxRunes) }

// TruncateDescription caps an operator-written summary at the length of one
// lifted from a page's own text, so both cut in the same place.
func TruncateDescription(s string) string { return truncateLine(s, descriptionMaxRunes) }

// truncateLine puts s on one line and caps it at limit runes.
func truncateLine(s string, limit int) string {
	return truncateRunes(oneLine(s), limit)
}

// truncateRunes cuts on a word boundary so the summary never ends mid-word.
func truncateRunes(s string, limit int) string {
	runes := []rune(s)
	if len(runes) <= limit {
		return s
	}
	cut := string(runes[:limit])
	if i := strings.LastIndexFunc(cut, unicode.IsSpace); i > 0 {
		cut = cut[:i]
	}
	return cut + "…"
}
