// Package viewdefault is a fixture for the view-by-default rule: a reference
// that crosses a realm boundary is read-only unless its owner granted a
// handle with mutable(x).
package viewdefault

import "gno.land/p/nt/avl/v0"

var (
	view   = avl.NewTree()          // handed out as a view
	handle = avl.NewTree()          // handed out with a grant
	inbox  *avl.Tree                // a caller's tree, stored by Store
)

func init() {
	view.Set("k", "v0")
	handle.Set("k", "h0")
}

// GetView returns a reference the caller can read but not write through.
func GetView() *avl.Tree { return view }

// GetHandle returns a reference the caller may write through.
func GetHandle() *avl.Tree { return mutable(handle) }

// GrantForeign tries to grant on an object this realm does not own.
func GrantForeign(cur realm, t *avl.Tree) *avl.Tree { return mutable(t) }

// Store keeps a caller's tree and later writes to it.
func Store(cur realm, t *avl.Tree) { inbox = t }

func WriteInbox(cur realm) { inbox.Set("from", "viewdefault") }

func Read(k string) string {
	if v := view.Get(k); v != nil {
		return v.(string)
	}
	if v := handle.Get(k); v != nil {
		return v.(string)
	}
	return ""
}
