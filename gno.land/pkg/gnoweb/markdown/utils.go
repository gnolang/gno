package markdown

import (
	"bytes"
	"errors"
	"html/template"
	"io"
	"unicode"

	"html/template"

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

// ExtractText returns the text content of a node, recursively.
func ExtractText(node ast.Node, source []byte) []byte {
	if node == nil {
		return nil
	}

	var buf bytes.Buffer

	type textNoder interface {
		Text([]byte) []byte
	}
	type softBreaker interface {
		SoftLineBreak() bool
	}

	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		// If the node can be rendered as text, use its method
		if tn, ok := child.(textNoder); ok {
			buf.Write(tn.Text(source))
		} else if t, ok := child.(*ast.Text); ok {
			// Fallback: raw text from the segment
			buf.Write(t.Segment.Value(source))
		} else {
			// Else, descend into its subtree
			buf.Write(ExtractText(child, source))
		}

		// Soft line break -> "\n"
		if sb, ok := child.(softBreaker); ok && sb.SoftLineBreak() {
			buf.WriteByte('\n')
		}
	}

	return buf.Bytes()
}

var titleCaser = cases.Title(language.AmericanEnglish)

func titleCase(s string) string {
	return titleCaser.String(s)
}
