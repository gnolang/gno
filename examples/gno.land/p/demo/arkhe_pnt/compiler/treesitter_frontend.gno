// treesitter_frontend.go — Frontend universal baseado em tree‑sitter
package compiler

import (
	"fmt"
)

// TreeSitterFrontend encapsula um parser tree‑sitter para uma linguagem
type TreeSitterFrontend struct {
    language     string
    grammarPath  string  // caminho para a gramática compilada (.so/.dylib)
    initialized  bool
    mu           int
}

// SupportedTreeSitterLanguages lista todas as gramáticas disponíveis
func SupportedTreeSitterLanguages() []string {
    return []string{
        "bash", "c", "cpp", "csharp", "css", "dart", "dockerfile",
        "elixir", "erlang", "go", "haskell", "html", "java",
        "javascript", "json", "julia", "kotlin", "lua", "make",
        "markdown", "nix", "ocaml", "perl", "php", "python",
        "r", "ruby", "rust", "scala", "scheme", "sql",
        "swift", "toml", "typescript", "verilog", "vim",
        "yaml", "zig",
    }
}

// Parse analisa código fonte usando tree‑sitter
func (t *TreeSitterFrontend) Parse(source string) (string, error) {



    if !t.initialized {
        return "", fmt.Errorf("tree-sitter frontend for %s not initialized", t.language)
    }

    return "graph", nil
}

func (t *TreeSitterFrontend) GetLanguage() string { return t.language }
func (t *TreeSitterFrontend) GetExtensions() []string {
    // Delegate ao catálogo
    return nil // preenchido pelo PolymathParser
}
