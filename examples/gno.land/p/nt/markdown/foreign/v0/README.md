> **v0 - Unaudited**
> This is an initial version of this package that has not yet been formally audited.
> A fully audited version will be published as a subsequent release.
> Use in production at your own risk.

# `foreign` - Foreign markdown sandbox

Realm-side helper that wraps externally-built markdown in a `<gno-foreign>` sandbox block. gnoweb renders the wrapped body inside its own goldmark sub-instance, so markdown you did not author cannot reach out and alter the surrounding page. Use it when flowing in markdown returned by another realm's interface method, fetched from chain storage owned by another realm, or otherwise outside your control.

## Usage

```go
package myrealm

import "gno.land/p/nt/markdown/foreign/v0"

func Render(path string) string {
    body := otherRealm.Render(path) // markdown you did not author
    return "## Included content\n\n" + foreign.Foreign(body)
}
```

With a caller-supplied label shown as a strip above the body:

```go
foreign.ForeignWithLabel("Pulled from /r/foo", body)
```

## API

```go
func Foreign(body string) string
func ForeignWithLabel(label, body string) string
func MaxBlocksPerRender() int
```

## Notes

- `body` is normalized before wrapping: `\r\n` and bare `\r` become `\n`, and any line that looks like a `gno-foreign` opener or closer (bare, attribute-bearing, or any case) has its leading `<` escaped to `&lt;`. Foreign content therefore cannot terminate the sandbox early or open a nested one.
- `ForeignWithLabel` sanitizes the label: bidi/zero-width characters are stripped, NUL is dropped, other control characters and the Unicode line separators (U+2028/U+2029/U+0085) become spaces, `&` `<` `>` `"` become HTML entities, and surrounding whitespace is trimmed. A label that is empty after sanitization behaves exactly like `Foreign` (no label strip, no default text).
- `MaxBlocksPerRender()` re-exports gnoweb's per-render cap on `<gno-foreign>` blocks (the same value the renderer reads). Past the cap, later blocks fall through to raw HTML and are dropped, so keep a page's foreign total under it.
- The renderer-side contract lives in `gno.land/pkg/gnoweb/markdown/ext_foreign.go`.
- To clean user-supplied (rather than realm-supplied) markdown at the leaf level, see [`sanitize`](../../sanitize/v0).
