// Package innerowner persists an *Inner that lives in this realm.
// Foreign-realm callers use GetInner() to obtain a pointer to it.
// When they pass it as the receiver of a /p/ method, PushFrameCall's
// borrow rule 2 fires and m.Realm shifts to this realm.
package innerowner

import "gno.land/p/demo/tests/p_closurecap"

var inn *p_closurecap.Inner

func init() {
	inn = &p_closurecap.Inner{N: 0}
}

// GetInner returns the persisted Inner. Non-crossing: the caller's
// realm stays current, but the returned *Inner carries PkgID =
// innerowner from its init-time allocation.
func GetInner() *p_closurecap.Inner {
	return inn
}
