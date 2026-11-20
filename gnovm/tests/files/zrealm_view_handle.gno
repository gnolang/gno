// PKGPATH: gno.land/r/demo/handleuser
package handleuser

import "gno.land/r/tests/vm/viewdefault"

// The owner writes its stored handle in its own entry function; the caller's
// reference to the same object is a view.

func main(cur realm) {
	viewdefault.HandleSet(cross(cur), "h", "h1")
	println("owner sees:", viewdefault.Read("h"))
}

// Output:
// owner sees: h1
