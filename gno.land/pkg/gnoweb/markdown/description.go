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

// Title returns the text of the first top-level h1, as plain text, read the
// way Description reads a paragraph.
func Title(doc ast.Node, src []byte) string {
	for n := doc.FirstChild(); n != nil; n = n.NextSibling() {
		if h, ok := n.(*ast.Heading); !ok || h.Level != 1 {
			continue
		}
		if text := plainText(src, n); text != "" {
			return truncateRunes(text, titleMaxRunes)
		}
	}
	return ""
}

// Description returns the first paragraph long enough to summarise the page,
// as plain text. It repeats only what the page already shows, so a page cannot
// carry a summary a reader cannot see.
func Description(doc ast.Node, src []byte) string {
	for n := doc.FirstChild(); n != nil; n = n.NextSibling() {
		if n.Kind() != ast.KindParagraph {
			continue
		}
		text := plainText(src, n)
		if len([]rune(text)) >= descriptionMinRunes {
			return truncateRunes(text, descriptionMaxRunes)
		}
	}
	return ""
}

// plainText is the visible text of n on one line.
func plainText(src []byte, n ast.Node) string {
	return strings.Join(strings.Fields(visibleText(src, n)), " ")
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

// TruncateDescription caps a summary at the same length as one lifted from a
// page's own text, so an operator-written description and a derived one cut in
// the same place.
func TruncateDescription(s string) string {
	return truncateRunes(strings.Join(strings.Fields(s), " "), descriptionMaxRunes)
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
