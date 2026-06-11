// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"gno.land/r/tests/vm/crossrealm_b"
)

// delete on a nil map is a no-op per the Go spec, even when the nil map
// was obtained from an external realm. Note a nil value never carries
// the readonly taint (TypedValue.IsReadonly requires a non-nil value),
// so this pins the cross-realm no-op path, not taint interaction.
func main(cur realm) {
	var m map[string]int // nil
	crossrealm_b.SetObject(cross, m)
	rm := crossrealm_b.Object.(map[string]int)
	delete(rm, "admin")
	println("ok, len:", len(rm))
}

// Output:
// ok, len: 0
