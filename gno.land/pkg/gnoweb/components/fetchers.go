package components

import (
	"context"
	"html/template"
)

// FileFetcher reads a single file from a package, used by the state
// orchestrator to pull source snippets referenced by walker output.
type FileFetcher interface {
	Fetch(ctx context.Context, pkgPath, fileName string) ([]byte, error)
}

// SnippetHighlighter returns template.HTML so the result is treated as
// already-safe markup by html/template.
type SnippetHighlighter interface {
	Render(fileName string, source []byte) (template.HTML, error)
}
