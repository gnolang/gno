package memberstore

import (
	"std"
	"strconv"
	"testing"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/urequire"
)

func TestCreateMembers(t *testing.T) {
	ms := NewMembersByTier()
	println("adding members...")
	addMembers(ms, 1000, "T1")
	println("added T1")
	addMembers(ms, 1000, "T2")
	println("added T2")
	addMembers(ms, 1000, "T3")
	println("added T3")

	m , tier := ms.GetMember(std.Address("11T3"))
	urequire.Equal(t,"T3",tier)

	m, tier = ms.GetMember(std.Address("2000T1"))
	urequire.Equal(t,"",tier)
	if m != nil {
		t.Fatal("member must be nil if not found")
	}

	tier = ms.RemoveMember(std.Address("1T1"))
	urequire.Equal(t,"T1",tier)
}



func addMembers(ms MembersByTier, c int, tier string) {
	mt := avl.NewTree()
	for i := 0; i < c; i++ {
		addr := std.Address(strconv.Itoa(i) + tier)
		if err := ms.SetMember(tier, addr, &Member{}); err != nil {
			panic(err.Error())
		}
	}
}