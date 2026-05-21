// Package v3al2store stores a single *Holder in its realm. The
// Holder lives in /r/tests/vm/v3al2store's storage; when a foreign
// /r/ accesses Token.Reflect(), Layer-2 borrow fires and shifts
// m.Realm to here. Used by the Layer-2 walk filetest.
package v3al2store

import "gno.land/p/nt/v3al2helper"

// Token is package-level — allocated in /r/tests/vm/v3al2store's
// realm at init, so its receiver Object carries this realm's PkgID
// for Layer-2 detection.
var Token = &v3al2helper.Holder{Tag: "stored-in-l2store"}
