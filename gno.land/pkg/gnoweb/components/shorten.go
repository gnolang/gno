package components

import (
	"path"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

// TruncMiddle shortens s to its first head and last tail runes joined by an
// ellipsis ("g17zyd…9cxg"), or returns it unchanged when that would not make
// it shorter. Negative bounds count as zero.
func TruncMiddle(s string, head, tail int) string {
	head, tail = max(head, 0), max(tail, 0)
	r := []rune(s)
	if len(r) <= head+tail+1 {
		return s
	}
	return string(r[:head]) + "…" + string(r[len(r)-tail:])
}

// PackageName returns a package path's name, with its version if it has
// one: "gno.land/r/demo/foo/v2" → "foo/v2", "gno.land/r/demo/foo" → "foo".
func PackageName(pkgPath string) string {
	base := path.Base(pkgPath)
	name := gno.LastPathElement(pkgPath)
	if name != base {
		return name + "/" + base
	}
	return name
}
