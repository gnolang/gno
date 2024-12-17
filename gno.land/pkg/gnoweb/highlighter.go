package gnoweb

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
)

type Highlighter interface {
	Format(w io.Writer, fileName string, file []byte) error
}

type ChromaHighlighter struct {
	*html.Formatter
	style *chroma.Style
}

func NewChromaHighlighter(formater *html.Formatter, style *chroma.Style) Highlighter {
	return &ChromaHighlighter{Formatter: formater, style: style}
}

func (f *ChromaHighlighter) Format(w io.Writer, fileName string, src []byte) error {
	var lexer chroma.Lexer

	switch strings.ToLower(filepath.Ext(fileName)) {
	case ".gno":
		lexer = lexers.Get("go")
	case ".md":
		lexer = lexers.Get("markdown")
	case ".mod":
		lexer = lexers.Get("gomod")
	default:
		lexer = lexers.Get("txt") // file kind not supported, fallback on `.txt`
	}

	if lexer == nil {
		return fmt.Errorf("unsupported lexer for file %q", fileName)
	}

	iterator, err := lexer.Tokenise(nil, string(src))
	if err != nil {
		return fmt.Errorf("unable to tokenise %q: %w ", fileName, err)
	}

	if err := f.Formatter.Format(w, f.style, iterator); err != nil {
		return fmt.Errorf("unable to format source file %q: %w", fileName, err)
	}

	return nil
}

type noopHighlighter struct{}

func (f *noopHighlighter) Format(w io.Writer, fileName string, src []byte) error {
	_, err := w.Write(src)
	return err
}
