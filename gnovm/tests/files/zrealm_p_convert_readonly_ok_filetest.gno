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
// cannot copy gno.land/r/tests/vm.Foo outside realm gno.land/r/tests/vm (use a pointer)

// TypeCheckError:
// gno.land/r/caller/zrealm_p_convert_readonly_ok.gno:17:7: cannot hold a value of gno.land/r/tests/vm.Foo outside its realm (use a pointer); gno.land/r/caller/zrealm_p_convert_readonly_ok.gno:20:5: cannot hold a value of gno.land/r/tests/vm.Foo outside its realm (use a pointer)
