// PKGPATH: gno.land/r/caller
// MAXALLOC: 100000000
package caller

import (
	tests "gno.land/r/tests/vm"
)

// The source-side readonly check in doOpConvert (Case 1) closes
// the type-pun launder (zrealm_launder_g_typepun.gno): converting
// a foreign-readonly /p/-value to a parallel /p/-type with mutator
// methods would otherwise bypass the readonly write check via
// borrow rule 2. Anonymous-struct reads of foreign values panic at
// the conversion site.

func main() {
	x := (struct {
		A int
		B *tests.Foo
	})(*tests.MyFoo)

	println("read A:", x.A)
}

// Error:
// illegal conversion of readonly or externally stored value
