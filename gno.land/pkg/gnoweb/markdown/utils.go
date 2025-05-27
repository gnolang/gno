package markdown

import (
	"errors"
	"io"
	"strings"

	"html/template"

	"golang.org/x/net/html"
)

// HTMLEscapeString escapes special characters in HTML content
func HTMLEscapeString(s string) string {
	return template.HTMLEscapeString(s)
}

// ParseHTMLToken parse line for tokens
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

func ExtractAttr(line string, attr string) string {
	tokens, err := ParseHTMLTokens(strings.NewReader(line))
	if err != nil {
		return ""
	}

	for _, tok := range tokens {
		if tok.Type == html.StartTagToken || tok.Type == html.SelfClosingTagToken {
			for _, a := range tok.Attr {
				if a.Key == attr {
					return a.Val
				}
			}
		}
	}
	return ""
}
