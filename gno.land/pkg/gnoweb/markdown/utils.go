package markdown

import (
	"io"

	"golang.org/x/net/html"
)

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
			if err != nil && err == io.EOF {
				return toks, nil
			}

			return nil, err
		}

		toks = append(toks, tok)
	}
}
