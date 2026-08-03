// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_ui_thread_13_filetest

// A non-numeric thread/reply ID segment in the URL is echoed back in the error
// message; it must be escaped so a crafted URL can't inject markup/HTML into
// the rendered page (a reflected-injection vector, live under gnoweb -html).
package z_ui_thread_13_filetest

import (
	"strings"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	boards2.CreateBoard(cross(cur), "test-board", false, false)
}

func main(cur realm) {
	// board exists, so the route reaches renderThread; "9<x" fails Atoi and is
	// echoed — escaped, so the raw "9<x" never appears (md.EscapeText → "9\<x").
	out := boards2.Render("test-board/9<x")
	ok := strings.Contains(out, "Invalid thread ID") &&
		!strings.Contains(out, "9<x")
	println(ok)
}

// Output:
// true
