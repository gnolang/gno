package markdown

import (
	"strings"
	"unicode"

	"github.com/yuin/goldmark/ast"
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
)

// RealmMeta is what rendering a document yields besides its HTML.
type RealmMeta struct {
	Toc         Toc
	Description string
}

// Description returns the first paragraph long enough to summarise the page,
// as plain text. It repeats only what the page already shows, so a page cannot
// carry a summary a reader cannot see.
func Description(doc ast.Node, src []byte) string {
	for n := doc.FirstChild(); n != nil; n = n.NextSibling() {
		if n.Kind() != ast.KindParagraph {
			continue
		}
		text := strings.Join(strings.Fields(visibleText(src, n)), " ")
		if len([]rune(text)) >= descriptionMinRunes {
			return truncateRunes(text, descriptionMaxRunes)
		}
	}
	return ""
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
			b.Write(nodeText(src, c))
		}
	}
	walk(n)
	return b.String()
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
