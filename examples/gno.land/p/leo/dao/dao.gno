package dao

import (
	"errors"
	"std"

	"gno.land/p/leo/utils"
)

type IDAO interface {
	GetMembers() []std.Address

	AddMember(std.Address) error

	DeleteMember(std.Address) error

	IsMember(std.Address) bool

	AssertUserIsMember(std.Address)
}

var (
	ErrUserAlreadyIsMember = errors.New("user is already a member")
	ErrUserIsNotMember     = errors.New("user is not a member")
)

type DAO struct {
	members []std.Address
}

func New() *DAO {
	return &DAO{}
}

func (d *DAO) GetMembers() []std.Address {
	return d.members
}

func (d *DAO) AddMember(usr std.Address) error {
	if d.IsMember(usr) {
		return ErrUserAlreadyIsMember
	}
	d.members = append(d.members, usr)
	return nil
}

func (d *DAO) DeleteMember(usr std.Address) error {
	for k, v := range d.members {
		if v == usr {
			d.members = append(d.members[:k], d.members[k+1:]...)
			return nil
		}
	}
	return ErrUserIsNotMember
}

func (d *DAO) IsMember(usr std.Address) bool {
	return utils.AddressSliceContains(d.members, func(_ int, e std.Address) bool {
		return e == usr
	})
}

func (d *DAO) AsserUserIsMember(usr std.Address) {
	if !d.IsMember(usr) {
		panic(ErrUserIsNotMember)
	}
}

func (d *DAO) AsserUserIsNotMember(usr std.Address) {
	if d.IsMember(usr) {
		panic(ErrUserAlreadyIsMember)
	}
}
