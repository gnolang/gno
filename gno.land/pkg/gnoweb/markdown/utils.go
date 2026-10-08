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
// with that tag ending (`/>` or `>`) on this line within maxLen bytes. It
// stops at a `<` outside a quoted value, so each attempt reads one tag.
func scanGnoTag(src, prefix []byte, maxLen int, attr func(key, val []byte)) (size int, selfClosing bool) {
	n := len(prefix)
	if len(src) <= n || !bytes.EqualFold(src[:n], prefix) {
		return 0, false
	}
	if c := src[n]; c != '/' && c != '>' && !util.IsSpace(c) {
		return 0, false // e.g. <gno-buttons>
	}
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

func isGnoTagAttrNameEnd(c byte) bool {
	return c == '/' || c == '>' || c == '=' || util.IsSpace(c)
}

// gnoTagLineParser opens a paragraph on a line that starts with a body-less
// inline gno-* tag. Without it, a line holding only `<gno-button … />` is a
// CommonMark type-7 HTML block, which takes the line (and every line up to
// the next blank one) before the tag's inline parser runs, and safe mode
// strips it. It delegates to goldmark's own paragraph parser, so the
// paragraph behaves like any other; the inline parser decides the rest.
type gnoTagLineParser struct {
	parser.BlockParser
	prefix []byte
	maxLen int
}

var _ parser.BlockParser = (*gnoTagLineParser)(nil)

func newGnoTagLineParser(prefix []byte, maxLen int) *gnoTagLineParser {
	return &gnoTagLineParser{BlockParser: parser.NewParagraphParser(), prefix: prefix, maxLen: maxLen}
}

func (*gnoTagLineParser) Trigger() []byte { return []byte{'<'} }

func (p *gnoTagLineParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, _ := reader.PeekLine()
	if size, _ := scanGnoTag(util.TrimLeftSpace(line), p.prefix, p.maxLen, nil); size == 0 {
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

// nodeText returns the text content of a node, recursively.
func nodeText(src []byte, n ast.Node) []byte {
	var buf bytes.Buffer
	writeNodeText(src, &buf, n)
	return buf.Bytes()
}

// writeNodeText writes the text content of a node to a buffer.
func writeNodeText(src []byte, dst io.Writer, n ast.Node) {
	switch n := n.(type) {
	case *ast.Text:
		_, _ = dst.Write(n.Segment.Value(src))
	case *ast.String:
		_, _ = dst.Write(n.Value)
	default:
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			writeNodeText(src, dst, c)
		}
	}
}

var titleCaser = cases.Title(language.AmericanEnglish)

func titleCase(s string) string {
	return titleCaser.String(s)
}
