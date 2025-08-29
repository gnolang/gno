package markdown

import (
	"bytes"
	"errors"
	"html/template"
	"io"
	"unicode"

	"github.com/yuin/goldmark/ast"
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
