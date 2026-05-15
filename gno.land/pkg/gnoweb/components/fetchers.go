package components

import (
	"context"
	"html/template"
)

// FileFetcher reads a single file from a package. `height = 0` queries
// the latest block; a positive value pins to that historical height so
// time-travel views render source consistent with the value snapshot.
type FileFetcher interface {
	Fetch(ctx context.Context, pkgPath, fileName string, height int64) ([]byte, error)
}

// SnippetHighlighter returns template.HTML so the result is treated as
// already-safe markup by html/template.
type SnippetHighlighter interface {
	Render(fileName string, source []byte) (template.HTML, error)
}
