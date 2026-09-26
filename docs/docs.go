// Package docs embeds the gno.land prose documentation (the Markdown files
// under this directory) and exposes them as an fs.FS so consumers like
// gnoweb can render them without a copy step. Putting the embed package
// here keeps the source-of-truth and the embed at the same location, so
// edits to any .md file land in the next build automatically.
//
// The whitepaper (.tex/.pdf/.aux/.toc) and the Makefile are deliberately
// excluded from the embed.
package docs

import (
	"embed"
	"io/fs"
)

//go:embed README.md CONSTITUTION.md LAWS.md MANIFESTO.md
//go:embed users/*.md builders/*.md resources/*.md
//go:embed images _assets
var content embed.FS

// FS returns the documentation filesystem rooted at the docs/ directory.
// Paths are relative, e.g. "builders/getting-started.md", "images/logo.png".
func FS() fs.FS { return content }
