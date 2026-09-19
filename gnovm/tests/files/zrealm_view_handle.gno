// PKGPATH: gno.land/r/demo/handleuser
package handleuser

import "gno.land/r/tests/vm/viewdefault"

// The owner granted with mutable(handle): the same /p/ method write borrows
// to the owner and commits to the owner's storage.

func main(cur realm) {
	t := viewdefault.GetHandle()
	t.Set("h", "h1")
	println("owner sees:", viewdefault.Read("h"))
}

// Output:
// owner sees: h1
