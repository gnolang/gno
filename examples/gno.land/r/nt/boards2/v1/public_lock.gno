package boards2

import "std"

// LockRealm locks the realm making it readonly.
// Warning, realm can't be unlocked once locked.
// This only locks the realm leaving realm membership open for modifications.
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

// LockRealmMembers locks the realm and realm membership making them readonly.
// Warning, realm can't be unlocked once locked.
func LockRealmMembers() {
	assertRealmMembersAreNotLocked()

	caller := std.PreviousRealm().Address()
	gPerms.WithPermission(caller, PermissionRealmMembersLock, nil, func(Args) {
		gLocked.realm = true
		gLocked.realmMembers = true

		std.Emit("RealmMembersLocked", "caller", caller.String())
	})
}

// AreRealmMembersLocked checks if realm members has been locked.
func AreRealmMembersLocked() bool {
	return gLocked.realmMembers
}

func assertRealmIsNotLocked() { // TODO: Add filtests for locked realm case to all public functions
	if gLocked.realm {
		panic("realm is locked")
	}
}

func assertRealmMembersAreNotLocked() { // TODO: Add filtests for locked members case to all public member functions
	if gLocked.realmMembers {
		panic("realm and membership are locked")
	}
}
