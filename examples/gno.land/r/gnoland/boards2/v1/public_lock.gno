package boards2

import (
	"chain"
	"strconv"

	"gno.land/p/gnoland/boards"
)

// LockRealm locks the realm making it readonly.
//
// WARNING: Realm can't be unlocked once locked.
//
// Realm can also be locked without locking realm members.
// Realm members can be locked when locking the realm or afterwards.
// This is relevant for two reasons, one so that members can be modified after the lock.
// The other is for realm owners, who can delete threads and comments after the lock.
func LockRealm(cur realm, lockRealmMembers bool) {
	assertRealmMembersAreNotLocked()

	// If realm members are not being locked assert that realm is not locked.
	// Members can be locked after locking the realm, in a second `LockRealm` call.
	if !lockRealmMembers {
		assertRealmIsNotLocked()
	}

	caller := cur.Previous().Address()
	gPerms.WithPermission(caller, PermissionRealmLock, boards.Args{}, func() {
		gLocked.realm = true
		gLocked.realmMembers = lockRealmMembers

		chain.Emit(
			"RealmLocked",
			"caller", caller.String(),
			"lockRealmMembers", strconv.FormatBool(lockRealmMembers),
		)
	})
}

// IsRealmLocked checks if boards realm has been locked.
func IsRealmLocked() bool {
	return gLocked.realm
}

// AreRealmMembersLocked checks if realm members have been locked.
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
		panic("realm and members are locked")
	}
}
