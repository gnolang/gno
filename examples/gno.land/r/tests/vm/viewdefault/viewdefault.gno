// Package viewdefault is a fixture for view-by-default: a reference that
// crosses a realm boundary is read-only, and only this realm's own entry
// functions write its stored objects.
package viewdefault

import (
	"encoding/base64"

	"gno.land/p/demo/tests/viewgrant/v0"
	"gno.land/p/nt/avl/v0"
)

var (
	view   = avl.NewTree() // handed out as a view
	handle = avl.NewTree() // written only by this realm's entry functions
	inbox  *avl.Tree       // a caller's tree, stored by Store
)

func init() {
	view.Set("k", "v0")
	handle.Set("k", "h0")
}

// GetView returns a reference the caller can read but not write through.
func GetView() *avl.Tree { return view }

// ViewSetter hands out a method value bound here on the view: called from
// another realm it writes nothing, because this realm's code did not call it.
func ViewSetter() func(string, any) bool { return view.Set }

// SetView writes the stored view from this realm's own crossing entry.
func SetView(cur realm, k string, v any) { view.Set(k, v) }

// CopyToView runs this realm's own method value from library code it called,
// writing a key nothing else writes so a probe can tell it landed.
func CopyToView(cur realm) {
	handle.Iterate("", "", func(k string, v any) bool {
		view.Set("own", "copy")
		return true
	})
}

// GetHandle returns a reference the caller can read; only this realm writes it.
func GetHandle() *avl.Tree { return handle }

// HandleSet writes the stored handle from this realm's own crossing entry.
func HandleSet(cur realm, k string, v any) { handle.Set(k, v) }

// Store keeps a caller's tree and later writes to it.
func Store(cur realm, t *avl.Tree) { inbox = t }

// StoreAndWrite writes through a caller's tree inside the same call.
func StoreAndWrite(cur realm, t *avl.Tree) {
	inbox = t
	t.Set("from", "viewdefault")
}

// RunHook calls a caller's function value from this realm's own code.
func RunHook(cur realm, fn func()) { fn() }

// IterateHandle runs a caller's callback from library code this realm calls.
func IterateHandle(cur realm, fn func(string, any) bool) { handle.Iterate("", "", fn) }

// HandleIterate is IterateHandle with the callback invoked from library code.
func HandleIterate(cur realm, fn func(string, any) bool) { handle.Iterate("", "", fn) }

// CallDeferred runs a caller's callback from a defer in this realm's code.
func CallDeferred(cur realm, fn func(string, any) bool) { defer fn("d", "deferred") }

// NewViewWriter hands out a fresh /p/-built writer onto the stored view.
func NewViewWriter() *viewgrant.Writer { return viewgrant.NewWriter(view) }

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

// secret is handed out as a view; Encoder hands out a stdlib method value
// that writes its first argument.
var secret = []byte("secret-bytes")

func Secret() []byte                 { return secret }
func SecretString() string           { return string(secret) }
func Encoder() func(dst, src []byte) { return base64.StdEncoding.Encode }
