package dao

import (
	"std"

	"gno.land/p/demo/avl"
)

type DAO struct {
	Proposals        *avl.Tree
	Whitelist        *avl.Tree
	AwaitingApproval *avl.Tree
	Hangouts         *avl.Tree
}

// Create a new DAO
// Initializes proposal, whitelist and waitlist
func NewDAO(caller std.Address) *DAO {
	newDao := &DAO{
		Proposals:        avl.NewTree(),
		Whitelist:        avl.NewTree(),
		AwaitingApproval: avl.NewTree(),
		Hangouts:         avl.NewTree(),
	}

	newDao.whitelistMemberNoApproval(caller)
	newDao.defaultAdmin(caller)

	return newDao
}
