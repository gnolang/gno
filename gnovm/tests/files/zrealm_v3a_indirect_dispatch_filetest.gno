// PKGPATH: gno.land/r/v3aindirect
package v3aindirect

import (
	"gno.land/p/demo/tests/v3aclhelper"
)

// Phase A.2/A.3 verification: indirect dispatch of a /p/ helper from a
// /r/ realm. The /p/ helper queries runtime.Caller() internally; the
// answer should resolve to the EOA (via v2's transition-walk semantics)
// no matter whether the helper is invoked directly or via a function
// value indirection.
//
// What we verify: the existing v2 substrate already implements v3a's
// "identity-query is stable across direct vs indirect dispatch"
// guarantee. Layer-1 borrow + the override walk converge on the
// same answer.

func main(cur realm) {
	// Direct call: /p/helper.CallerPath() resolves to the EOA realm
	// (empty pkgPath since EOA is a user, not a code realm).
	directPath := v3aclhelper.CallerPath()

	// Indirect call: store the same /p/ helper in a function value,
	// then invoke. The function value's OriginRealm is /p/v3aclhelper
	// (where the function was declared). Dispatch via the value still
	// routes to the same code; runtime.Caller() inside resolves the
	// same answer.
	fn := v3aclhelper.CallerPath
	indirectPath := fn()

	println("direct == indirect:", directPath == indirectPath)
	println("path is empty (eoa):", directPath == "")
}

// Output:
// direct == indirect: true
// path is empty (eoa): true
