package valhalla

import (
	"errors"
	
	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/avl"
)

var MemberList avl.Tree

func AddMember(memberKey string) error {
	assertIsAdmin()
	if IsMember(memberKey) {
		return errors.New("member already exists")
	}
	MemberList.Set(memberKey, true)
	return nil
}

func RemoveMember(memberKey string) error {
	assertIsAdmin()
	if !IsMember(memberKey) {
		return errors.New("member does not exist")
	}
	MemberList.Remove(memberKey)
	return nil
}

func IsMember(memberKey string) bool {
	return MemberList.Has(memberKey)
}

func MemberCount() int {
	return MemberList.Size()
}

func Render() {
	println(
		ufmt.Sprintf(
			"Members:\n\n%v", MemberList,
		),
	)
}
