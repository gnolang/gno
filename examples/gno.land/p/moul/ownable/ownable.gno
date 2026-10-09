// Package ownable proposes an abstraction around the ownership which provides
// various kind of owners and some composition options.
package ownable

import "std"

type Owner interface {
	Addr() std.Address
	Kind() Kind
	owner()
}

type UserOwner struct {
	addr std.Address
}

type RealmOwner struct {
	realm std.Realm
}

type MembstoreOwner struct {
	// membstore.Membstore
}

type DAOOwner struct {
	// dao.DAO
}

func And(a, b ...Owner) Owner { panic("not implemented") }
func Or(a, b ...Owner) Owner  { panic("not implemented") }
