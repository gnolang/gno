package boards2

import "std"

// LockRealm locks the realm making it readonly.
// Warning, realm can't be unlocked once locked.
func LockRealm() {
	assertRealmIsNotLocked()

	caller := std.PreviousRealm().Address()
	gPerms.WithPermission(caller, PermissionRealmLock, nil, func(Args) {
		gLocked.realm = true

		std.Emit("RealmLocked", "caller", caller.String())
	})
}

// IsRealmLocked checks if boards realm has been locked.
func IsRealmLocked() bool {
	return gLocked.realm
}

func assertRealmIsNotLocked() { // TODO: Add filtests for locked realm case to all public functions?
	if gLocked.realm {
		panic("realm is locked")
	}
}
