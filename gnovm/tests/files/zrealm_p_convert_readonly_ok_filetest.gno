// PKGPATH: gno.land/r/caller
// MAXALLOC: 100000000
package caller

import (
	tests "gno.land/r/tests/vm"
)

// interrealm v2 Phase 4: dropping the source-side readonly check in
// doOpConvert (old Case 1) lets legitimate read-only patterns
// work — converting a foreign /r/-owned value into an anonymous
// struct type for inspection no longer panics. The N_Readonly
// taint survives the conversion (ConvertTo doesn't touch tv.N),
// so any subsequent write through the converted value is still
// caught at the write site (see exploit/realm_exploiter.gno for
// the negative case).

func main() {
	// Cross-realm read into a local anonymous struct type.
	// Pre-interrealm-v2 Phase 4 this panicked with "illegal conversion
	// of readonly or externally stored value." Now it succeeds;
	// reads work, but writes via the converted value still fail.
	x := (struct {
		A int
		B *tests.Foo
	})(*tests.MyFoo)

	println("read A:", x.A)
}

// Output:
// read A: 1
