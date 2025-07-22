package commondao

import "std"

func JoinDao(cur realm, caller std.Address) error {
	if DAO.Members().Has(caller) {
		panic("you are already part of this friend group")
	}

	DAO.Members().Add(caller)
	return nil
}
