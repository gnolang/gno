package commondao

import "std"

// JoinDao allows a member to join the DAO
// It checks if the member is already part of the DAO and adds them if not
func JoinDao(cur realm, caller std.Address) error {
	if DAO.Members().Has(caller) {
		panic("you are already part of this friend group")
	}

	DAO.Members().Add(caller)
	return nil
}
