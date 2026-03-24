package memberstore

import (
	"strconv"
	"testing"

	"gno.land/p/nt/urequire/v0"
)

func TestPower(t *testing.T) {
	ms := NewMembersByTier()
	addMembers(ms, 100, T1)
	addMembers(ms, 100, T2)
	addMembers(ms, 100, T3)

	tiers.Iterate("", "", func(key string, value interface{}) bool {
		po := value.(Tier).PowerHandler(ms, tiers)
		if key == T1 && po != 3.0 {
			t.Fatal("wrong value for T1")
		}
		if key == T2 && po != 2.0 {
			t.Fatal("wrong value for T2")
		}
		if key == T3 && po != 1.0 {
			t.Fatal("wrong value for T3")
		}

		return false
	})

	ms = NewMembersByTier()
	addMembers(ms, 100, T1)
	addMembers(ms, 50, T2)
	addMembers(ms, 10, T3)

	tiers.Iterate("", "", func(key string, value interface{}) bool {
		po := value.(Tier).PowerHandler(ms, tiers)
		if key == T1 && po != 3.0 {
			t.Fatal("wrong value for T1")
		}
		if key == T2 && po != 2.0 {
			t.Fatal("wrong value for T2")
		}
		if key == T3 && po != 1.0 {
			t.Fatal("wrong value for T3")
		}

		return false
	})

	ms = NewMembersByTier()
	addMembers(ms, 100, T1)
	addMembers(ms, 200, T2)
	addMembers(ms, 100, T3)

	tiers.Iterate("", "", func(key string, value interface{}) bool {
		po := value.(Tier).PowerHandler(ms, tiers)
		if key == T1 && po != 3.0 {
			t.Fatal("wrong value for T1")
		}
		if key == T2 && po != 1.0 {
			t.Fatal("wrong value for T2")
		}
		if key == T3 && po != 1.0 {
			t.Fatal("wrong value for T3")
		}

		return false
	})

	ms = NewMembersByTier()
	addMembers(ms, 100, T1)
	addMembers(ms, 200, T2)
	addMembers(ms, 1000, T3)

	tiers.Iterate("", "", func(key string, value interface{}) bool {
		po := value.(Tier).PowerHandler(ms, tiers)
		if key == T1 && po != 3.0 {
			t.Fatal("wrong value for T1")
		}
		if key == T2 && po != 1.0 {
			t.Fatal("wrong value for T2")
		}
		if key == T3 && po != 0.1 {
			t.Fatal("wrong value for T3")
		}

		return false
	})
}

func TestCreateMembers(t *testing.T) {
	ms := NewMembersByTier()
	println("adding members...")
	addMembers(ms, 10, "T1")
	println("added T1")
	addMembers(ms, 100, "T2")
	println("added T2")
	addMembers(ms, 1000, "T3")
	println("added T3")

	m, tier := ms.GetMember(address("11T3"))
	urequire.Equal(t, "T3", tier)

	m, tier = ms.GetMember(address("2000T1"))
	urequire.Equal(t, "", tier)
	if m != nil {
		t.Fatal("member must be nil if not found")
	}

	tier = ms.RemoveMember(address("1T1"))
	urequire.Equal(t, "T1", tier)
}

func addMembers(ms MembersByTier, c int, tier string) {
	// mt := avl.NewTree() XXX
	ms.SetTier(tier)
	for i := 0; i < c; i++ {
		addr := address(strconv.Itoa(i) + tier)
		if err := ms.SetMember(tier, addr, &Member{}); err != nil {
			panic(err.Error())
		}
	}
}
