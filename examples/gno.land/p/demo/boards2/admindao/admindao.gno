package admindao

import (
	"errors"
	"std"

	"gno.land/p/demo/avl"
)

// TODO: Add support for proposals
// TODO: Add support for events

// ErrMemberExists indicates that a member is already part of the DAO.
var ErrMemberExists = errors.New("member already exist")

// AdminDAO defines a Boards administration DAO.
type AdminDAO struct {
	parent  *AdminDAO
	members *avl.Tree // string(std.Address) -> struct{}
}

// New creates a new admin DAO.
func New(options ...Option) *AdminDAO {
	dao := &AdminDAO{members: avl.NewTree()}
	for _, apply := range options {
		apply(dao)
	}
	return dao
}

// Parent returns the parent DAO.
// Null can be returned when DAO has no parent assigned.
func (dao AdminDAO) Parent() *AdminDAO {
	return dao.parent
}

// Members returns the list of DAO members.
func (dao AdminDAO) Members() []std.Address {
	var members []std.Address
	dao.members.Iterate("", "", func(key string, _ interface{}) bool {
		members = append(members, std.Address(key))
		return false
	})
	return members
}

// AddMember adds a new member to the DAO.
func (dao *AdminDAO) AddMember(user std.Address) error {
	if dao.IsMember(user) {
		return ErrMemberExists
	}
	dao.members.Set(user.String(), struct{}{})
	return nil
}

// IsMember checks if a user is a member of the DAO.
func (dao AdminDAO) IsMember(user std.Address) bool {
	return dao.members.Has(user.String())
}
