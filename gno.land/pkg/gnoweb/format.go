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

// FormatSource defines the interface for formatting source code.
type FormatSource interface {
	Format(w io.Writer, fileName string, file []byte) error
}

// ChromaSourceHighlighter implements the Highlighter interface using the Chroma library.
type ChromaSourceHighlighter struct {
	*html.Formatter
	style *chroma.Style
}

// NewChromaSourceHighlighter constructs a new ChromaHighlighter with the given formatter and style.
func NewChromaSourceHighlighter(formatter *html.Formatter, style *chroma.Style) FormatSource {
	return &ChromaSourceHighlighter{Formatter: formatter, style: style}
}

// Format applies syntax highlighting to the source code using Chroma.
func (f *ChromaSourceHighlighter) Format(w io.Writer, fileName string, src []byte) error {
	var lexer chroma.Lexer

	// Determine the lexer to be used based on the file extension.
	switch strings.ToLower(filepath.Ext(fileName)) {
	case ".gno":
		lexer = lexers.Get("go")
	case ".md":
		lexer = lexers.Get("markdown")
	case ".mod":
		lexer = lexers.Get("gomod")
	default:
		lexer = lexers.Get("txt") // Unsupported file type, default to plain text.
	}

	if lexer == nil {
		return fmt.Errorf("unsupported lexer for file %q", fileName)
	}

	iterator, err := lexer.Tokenise(nil, string(src))
	if err != nil {
		return fmt.Errorf("unable to tokenise %q: %w", fileName, err)
	}

	if err := f.Formatter.Format(w, f.style, iterator); err != nil {
		return fmt.Errorf("unable to format source file %q: %w", fileName, err)
	}

	return nil
}

// noopFormat is a no-operation highlighter that writes the source code as-is.
type noopFormat struct{}

// Format writes the source code to the writer without any formatting.
func (f *noopFormat) Format(w io.Writer, fileName string, src []byte) error {
	_, err := w.Write(src)
	return err
}
