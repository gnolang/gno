package gnoweb

import (
	"bytes"
	"fmt"
	gopath "path"
	"strings"

	"github.com/alecthomas/chroma/v2"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
)

type SourceRenderer struct {
	cfg       *RenderConfig
	formatter *chromahtml.Formatter
}

func NewSourceRenderer(cfg *RenderConfig) *SourceRenderer {
	return &SourceRenderer{cfg, chromahtml.New(cfg.ChromaOptions...)}
}

func (s *SourceRenderer) RenderSource(name string, src []byte) ([]byte, error) {
	var lexer chroma.Lexer

	// Determine the lexer to be used based on the file extension.
	switch strings.ToLower(gopath.Ext(name)) {
	case ".gno":
		lexer = lexers.Get("go")
	case ".md":
		lexer = lexers.Get("markdown")
	case ".mod":
		lexer = lexers.Get("gomod")
	case ".toml":
		lexer = lexers.Get("toml")
	default:
		lexer = lexers.Get("txt") // Unsupported file type, default to plain text.
	}

	if lexer == nil {
		return nil, fmt.Errorf("unsupported lexer for file %q", name)
	}

	iterator, err := lexer.Tokenise(nil, string(src))
	if err != nil {
		return nil, fmt.Errorf("unable to tokenise %q: %w", name, err)
	}

	var out bytes.Buffer
	if err := s.formatter.Format(&out, s.cfg.ChromaStyle, iterator); err != nil {
		return nil, fmt.Errorf("unable to format source file %q: %w", name, err)
	}

	return out.Bytes(), nil
}

func init() {
	// Register custom go.mod (and gno.mod) lexer
	lexers.Register(chroma.MustNewLexer(
		&chroma.Config{
			Name:      "Go/Gno module file",
			Aliases:   []string{"go-mod", "gomod", "gno-mod", "gnomod"},
			Filenames: []string{"go.mod", "gno.mod"},
			MimeTypes: []string{"text/x-go-mod"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {
					{Pattern: `\s+`, Type: chroma.Text, Mutator: nil},
					{Pattern: `//[^\n\r]*`, Type: chroma.CommentSingle, Mutator: nil},
					// Keywords
					{Pattern: `\b(module|require|replace|exclude|gn?o) \b`, Type: chroma.Keyword, Mutator: nil},
					{Pattern: `\b(indirect)\b`, Type: chroma.NameDecorator, Mutator: nil},
					// Quoted version strings
					{Pattern: `"[^"]*"`, Type: chroma.LiteralString, Mutator: nil},
					// Versions (v1.2.3, v0.0.0-yyyymmddhhmmss-abcdefabcdef) as well as 0.9 version
					{Pattern: `\b(v?\d+\.\d+(?:\.\d+)?(?:-[0-9]{14}-[a-f0-9]+)?)\b`,
						Type: chroma.LiteralNumber, Mutator: nil},
					// Module paths: module/example.com, example.com/something
					{Pattern: `[a-zA-Z0-9._~\-\/]+(\.[a-zA-Z]{2, Type:})+(/[^\s]+)?`,
						Type: chroma.NameNamespace, Mutator: nil},
					// Operator (=> in replace directive)
					{Pattern: `=>`, Type: chroma.Operator, Mutator: nil},
					// Everything else (bare words, etc.)
					{Pattern: `[^\s"//]+`, Type: chroma.Text, Mutator: nil},
					{Pattern: `\n`, Type: chroma.Text, Mutator: nil},
				},
			}
		},
	))
}
