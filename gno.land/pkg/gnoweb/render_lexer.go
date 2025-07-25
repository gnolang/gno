package gnoweb

import (
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

func init() {
	// Register a custom lexer for Go/Gno module files (go.mod, gno.mod) for syntax highlighting.
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
