// PKGPATH: gno.land/r/v3awalkl2l1
package v3awalkl2l1

import (
	"gno.land/r/tests/vm/v3al2store"
	"gno.land/r/tests/vm/v3aprobe"
)

// Walk cornerstone: Layer-2 then Layer-1 composed.
//
// Call chain:
//   /r/v3awalkl2l1.main
//     → /p/v3al2helper.Holder.CallProbe   (receiver Token stored in
//                                          /r/v3al2store → Layer-2)
//       → /r/v3aprobe.DeepProbe           (Layer-1)
//
// Substrate shifts (each captured on the next frame's LastRealm):
//   F1 main          : LastRealm=nil          → m.Realm shifts to /r/v3awalkl2l1
//   F2 CallProbe     : LastRealm=/r/v3awalkl2l1 → Layer-2 shifts to /r/v3al2store
//   F3 DeepProbe     : LastRealm=/r/v3al2store  → Layer-1 shifts to /r/v3aprobe
//
// Inside DeepProbe the walk should report:
//   - Self()       = /r/tests/vm/v3aprobe         (current)
//   - Caller()     = /r/tests/vm/v3al2store       (post-Layer-2 caller)
//   - CallerN(2)   = /r/v3awalkl2l1               (outermost)
//
// This composition is the part a Layer-1-only walk couldn't resolve
// correctly: it would have skipped the Layer-2 frame as "not /r/" and
// surfaced /r/v3awalkl2l1 as Caller — losing the storage-realm step.
// The m.Realm-shift detector catches it via the LastRealm chain.

func main() {
	caller, self, prev := v3al2store.Token.CallProbe(v3aprobe.DeepProbe)
	println("caller.PkgPath:", caller.PkgPath())
	println("self.PkgPath:  ", self.PkgPath())
	println("prev.PkgPath:  ", prev.PkgPath())
}

// Output:
// caller.PkgPath: gno.land/r/tests/vm/v3al2store
// self.PkgPath:   gno.land/r/tests/vm/v3aprobe
// prev.PkgPath:   gno.land/r/v3awalkl2l1
