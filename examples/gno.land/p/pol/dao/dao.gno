package dao

import (
	"errors"
	"std"
)

type Member std.Address

func (m Member) Address() std.Address {
	return std.Address(m)
}

type IDAO interface {
	// Get the organisation members
	GetMembers() []Member

	// Add a member to an organisation
	AddMember(who Member) error

	// Propose the exclusion of a member
	ExcludeMember(who Member) error

	// Check if member is member
	IsMember(who Member) bool

	// Assert if the user is a member
	AssertUserIsMember(who Member) 
}

type DAO struct {
	IDAO
	ProposalStore
	members []Member
}

func (d *DAO) GetMembers() []Member {
	return d.members
}

func (d *DAO) AddMember(who Member) error {
	if d.IsMember(who) {
		return errors.New("already a member")
	}
	d.members = append(d.members, who)
	return nil
}

func (d *DAO) ExcludeMember(who Member) error {
	for k, v := range d.members {
		if v == who {
			d.members = append(d.members[:k], d.members[k + 1:]...)
			return nil
		}
	}
	return errors.New("member not found")
}

func (d *DAO) IsMember(who Member) bool {
	for _, v := range d.members {
		if v == who {
			return true
		}
	}
	return false
}

func (d *DAO) AssertUserIsMember(who Member) {
	if who.Address().String() == "" || !d.IsMember(who) {
		panic("user is not in DAO")
	}
}

func NewDAO() DAO {
	return DAO{members: []Member{}}
}

