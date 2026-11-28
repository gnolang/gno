> **v0 - Unaudited**
> This is an initial version of this package that has not yet been formally audited.
> A fully audited version will be published as a subsequent release.
> Use in production at your own risk.

# `fqname` - Fully qualified identifiers

Parse, construct, and link fully qualified Gno identifiers of the form `<pkgpath>.<name>` (e.g. `gno.land/p/nt/avl/v0.Tree`).

## Usage

```go
import "gno.land/p/nt/fqname/v0"

// Split a fully qualified name
pkgpath, name := fqname.Parse("gno.land/p/nt/avl/v0.Tree")
// pkgpath == "gno.land/p/nt/avl/v0", name == "Tree"

// Rebuild one from its parts
id := fqname.Construct("gno.land/r/demo/foo20", "Token")
// id == "gno.land/r/demo/foo20.Token"

// Render as a Markdown link (gno.land paths become clickable)
link := fqname.RenderLink("gno.land/r/demo/foo20", "Token")
// link == "[gno.land/r/demo/foo20](/r/demo/foo20).Token"
```

## API

```go
// Parse splits a fully qualified identifier into (pkgpath, name).
// If no name is present (no dot after the last slash), name is "".
func Parse(fqname string) (pkgpath, name string)

// Construct joins pkgpath and name with a dot. If name is empty, returns pkgpath.
func Construct(pkgpath, name string) string

// RenderLink formats a fully qualified identifier as Markdown.
// Paths starting with "gno.land" are turned into a link to the package;
// other paths are returned as plain text. The slug is dot-appended and
// markdown-escaped.
func RenderLink(pkgPath, slug string) string
```

## Notes

- `Parse` treats everything after the dot following the last slash as the name, so nested selectors like `Pkg.Type.Method` round-trip as a single name.
- `RenderLink` only links `gno.land`-rooted paths; foreign domains (e.g. `github.com/...`) are returned unmodified except for the dot-joined slug.
- `RenderLink` markdown-escapes the `slug`, but NOT `pkgPath`: a `]` or `)` in an untrusted `pkgPath` breaks out of the link. Pass validated package paths, or sanitize with [`gno.land/p/nt/markdown/sanitize/v0`](../../markdown/sanitize/v0) before rendering untrusted input.
