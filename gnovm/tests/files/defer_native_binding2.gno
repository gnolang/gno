// PKGPATH: gno.land/r/test/defernative
package defernative

import "chain/params"

// A deferred native that has not been called normally earlier in the
// transaction must still run, and its state change must persist.
func set(cur realm) {
	defer params.SetString("k", "deferred")
}

func main(cur realm) {
	set(cross(cur))
	v, ok := params.GetString("k")
	println(v, ok)
}

// Output:
// deferred true
