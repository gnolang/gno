package markdown

import (
	"bytes"
	"errors"
	"html/template"
	"io"
	"unicode"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"golang.org/x/net/html"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// HTMLEscapeString escapes special characters in HTML content
func HTMLEscapeString(s string) string {
	return template.HTMLEscapeString(s)
}

// ParseHTMLTokens parses an HTML stream and returns a slice of html.Token.
// It stops at EOF or on error.
func ParseHTMLTokens(r io.Reader) ([]html.Token, error) {
	tokenizer := html.NewTokenizer(r)
	tokenizer.AllowCDATA(false)

	toks := []html.Token{}
	for {
		// Check for any html comment
		tokenizer.Next()
		tok := tokenizer.Token()
		if tok.Type == html.ErrorToken {
			err := tokenizer.Err()
			if err != nil && errors.Is(err, io.EOF) {
				return toks, nil
			}

			return nil, err
		}

		toks = append(toks, tok)
	}
}

// scanGnoTag reads a `<name …>` tag at the start of src without allocating;
// prefix is "<name", matched case-insensitively. It calls attr, when non-nil,
// for each attribute in source order with its raw value (no entity decoding),
// aliasing src, and returns the tag's length, or 0 when src does not start
// with that tag ending (`/>` or `>`) on this line within maxLen bytes. For the
// body-less inline gno-* tags: <gno-icon /> and <gno-button />.
func scanGnoTag(src, prefix []byte, maxLen int, attr func(key, val []byte)) (size int, selfClosing bool) {
	if !hasGnoTagPrefix(src, prefix) {
		return 0, false
	}
	n := len(prefix)
	src = src[:min(len(src), maxLen)]
	if eol := bytes.IndexByte(src, '\n'); eol >= 0 {
		src = src[:eol] // a tag spans one line
	}

	for i := n; i < len(src); {
		switch c := src[i]; {
		case util.IsSpace(c):
			i++
			continue
		case c == '>':
			return i + 1, false
		case c == '<':
			return 0, false // the next tag: each attempt reads one tag
		case c == '/':
			if i+1 < len(src) && src[i+1] == '>' {
				return i + 2, true
			}
			i++
			continue
		}

		// Attribute name, then an optional `= value`.
		start := i
		for i++; i < len(src) && !isGnoTagAttrNameEnd(src[i]); i++ {
		}
		key := src[start:i]
		for i < len(src) && util.IsSpace(src[i]) {
			i++
		}
		var val []byte
		if i < len(src) && src[i] == '=' {
			for i++; i < len(src) && util.IsSpace(src[i]); i++ {
			}
			if i == len(src) {
				return 0, false
			}
			if q := src[i]; q == '"' || q == '\'' {
				end := bytes.IndexByte(src[i+1:], q)
				if end < 0 {
					return 0, false
				}
				val = src[i+1 : i+1+end]
				i += end + 2
			} else {
				start := i
				for i < len(src) && !util.IsSpace(src[i]) && src[i] != '>' {
					i++
				}
				val = src[start:i]
			}
		}
		if attr != nil {
			attr(key, val)
		}
	}
	return 0, false
}

// hasGnoTagPrefix reports whether src starts with prefix ("<name"),
// case-insensitively, followed by a byte that ends a tag name: whitespace,
// `/` or `>`, so `<gno-icons>` is not `<gno-icon`.
func hasGnoTagPrefix(src, prefix []byte) bool {
	n := len(prefix)
	if len(src) <= n || !bytes.EqualFold(src[:n], prefix) {
		return false
	}
	c := src[n]
	return c == '/' || c == '>' || util.IsSpace(c)
}

func isGnoTagAttrNameEnd(c byte) bool {
	return c == '/' || c == '>' || c == '=' || util.IsSpace(c)
}

// gnoTagLineParser opens a paragraph on a line that starts with a body-less
// inline gno-* tag. Without it, a line holding only `<gno-icon … />` is a
// CommonMark type-7 HTML block, which takes the line (and every line up to
// the next blank one) before the tag's inline parser runs, and safe mode
// strips it. It delegates to goldmark's own paragraph parser, so the
// paragraph behaves like any other; the inline parser decides the rest.
//
// It opens on the tag name alone, not on a tag the inline parser would
// accept: a tag too long or malformed for scanGnoTag must still not turn the
// line into an HTML block, which would hide the lines after it.
type gnoTagLineParser struct {
	parser.BlockParser
	prefix []byte
}

var _ parser.BlockParser = (*gnoTagLineParser)(nil)

func newGnoTagLineParser(prefix []byte) *gnoTagLineParser {
	return &gnoTagLineParser{BlockParser: parser.NewParagraphParser(), prefix: prefix}
}

func (*gnoTagLineParser) Trigger() []byte { return []byte{'<'} }

func (p *gnoTagLineParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, _ := reader.PeekLine()
	if !hasGnoTagPrefix(util.TrimLeftSpace(line), p.prefix) {
		return nil, parser.NoChildren
	}
	return p.BlockParser.Open(parent, reader, pc)
}

func ExtractAttr(attrs []html.Attribute, key string) (val string, ok bool) {
	for _, attr := range attrs {
		if key == attr.Key {
			return attr.Val, true
		}
	}

	return "", false
}

// GetWordArticle returns "a" or "an" based on the first letter of the word
func GetWordArticle(word string) string {
	if len(word) == 0 {
		return "a"
	}

	// Check if the first letter is a vowel (a, e, i, o, u)
	firstChar := unicode.ToLower(rune(word[0]))
	if firstChar == 'a' || firstChar == 'e' || firstChar == 'i' || firstChar == 'o' || firstChar == 'u' {
		return "an"
	}
	return "a"
}

// nodeText returns the text content of a node, recursively, with backslash
// escapes resolved.
func nodeText(src []byte, n ast.Node) []byte {
	var w nodeTextWriter
	w.walk(src, n)
	return w.buf.Bytes()
}

// nodeTextWriter collects a node's text. An icon that renders adds its
// decoded label, the name it gives the heading or link it sits in, set apart
// from the text next to it by a space; an icon that renders nothing, or sits
// in an image's alt text, adds nothing.
type nodeTextWriter struct {
	buf   bytes.Buffer
	space bool // a label was just written: the next text needs a space
}

func (w *nodeTextWriter) walk(src []byte, n ast.Node) {
	switch n := n.(type) {
	case *ast.Text:
		w.text(util.UnescapePunctuations(n.Segment.Value(src)))
	case *ast.String:
		w.text(util.UnescapePunctuations(n.Value))
	case *Icon:
		label := n.accessibleName()
		if label == nil || inImage(n) {
			return
		}
		if l := w.buf.Len(); l > 0 && !util.IsSpace(w.buf.Bytes()[l-1]) {
			w.buf.WriteByte(' ')
		}
		w.buf.Write(label)
		w.space = true
	default:
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			w.walk(src, c)
		}
	}
}

func (w *nodeTextWriter) text(b []byte) {
	if len(b) == 0 {
		return
	}
	if w.space && !util.IsSpace(b[0]) {
		w.buf.WriteByte(' ')
	}
	w.space = false
	w.buf.Write(b)
}

var titleCaser = cases.Title(language.AmericanEnglish)

func titleCase(s string) string {
	return titleCaser.String(s)
}

// trimTagLine strips 0-3 leading spaces (CM §4.5 indent tolerance) and
// trailing ASCII whitespace from a line holding a gno-* block tag. A line
// indented further keeps a leading space, so it never parses as one tag.
func trimTagLine(line []byte) []byte {
	i := 0
	for i < len(line) && i < 3 && line[i] == ' ' {
		i++
	}
	return util.TrimRightSpace(line[i:])
}
