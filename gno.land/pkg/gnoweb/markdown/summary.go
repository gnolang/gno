package markdown

import (
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// summaryParser reads a document as CommonMark under the realm renderer's
// emphasis guard: the source is untrusted, and Summary only reads it.
var summaryParser = goldmark.New(goldmark.WithExtensions(ExtEmphasis)).Parser()

// summaryMaxBytes is how much of a document Summary parses. Its leading
// paragraphs sit at the top, and a summary is a few hundred runes, so a long
// document costs no more than a short one.
const summaryMaxBytes = 8 << 10

// Summary is the opening of a markdown document as plain text: its leading
// paragraphs, up to the first heading, list, quote or code block, cut after
// the last whole sentence that fits limit runes, or on a word when even the
// first sentence does not fit. It reads text the way Lead does, so an image's
// alt text, raw HTML and invisible format characters never reach it.
func Summary(src string, limit int) string {
	if len(src) > summaryMaxBytes {
		src = src[:summaryMaxBytes]
	}
	b := []byte(src)
	doc := summaryParser.Parse(text.NewReader(b))
	var parts []string
	for n := doc.FirstChild(); n != nil && n.Kind() == ast.KindParagraph; n = n.NextSibling() {
		if t := plainText(b, n); t != "" {
			parts = append(parts, t)
		}
	}
	return truncateSentences(strings.Join(parts, " "), limit)
}

// truncateSentences keeps the whole sentences of s that fit limit runes. A
// sentence ends on '.', '!' or '?' followed by a space or the end of s; an
// abbreviation such as "e.g." can end one early, which only shortens the
// summary. When not even the first sentence fits, s is cut on a word.
func truncateSentences(s string, limit int) string {
	end, n := -1, 0
	for i, r := range s {
		if n == limit {
			if end > 0 {
				return s[:end]
			}
			return truncateRunes(s, limit)
		}
		n++
		if (r == '.' || r == '!' || r == '?') && (i+1 == len(s) || s[i+1] == ' ') {
			end = i + 1
		}
	}
	return s
}
