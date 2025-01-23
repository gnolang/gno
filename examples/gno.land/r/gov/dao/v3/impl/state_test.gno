package govdao

import (
	"fmt"
	"std"
	"strconv"
	"testing"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/testutils"
	"gno.land/r/gov/dao/ruler"
	"gno.land/r/gov/dao/v3/types"
)

func TestPowerDefinedInLaw(t *testing.T) {
	ms := types.NewMembersByTier()
	addMembers(ms, 100, T1)
	addMembers(ms, 100, T2)
	addMembers(ms, 100, T3)

	theLaw.Tiers.Iterate("", "", func(key string, value interface{}) bool {
		po := value.(types.Tier).PowerHandler(ms, theLaw.Tiers)
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

	ms = types.NewMembersByTier()
	addMembers(ms, 100, T1)
	addMembers(ms, 50, T2)
	addMembers(ms, 10, T3)

	theLaw.Tiers.Iterate("", "", func(key string, value interface{}) bool {
		po := value.(types.Tier).PowerHandler(ms, theLaw.Tiers)
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

	ms = types.NewMembersByTier()
	addMembers(ms, 100, T1)
	addMembers(ms, 200, T2)
	addMembers(ms, 100, T3)

	theLaw.Tiers.Iterate("", "", func(key string, value interface{}) bool {
		po := value.(types.Tier).PowerHandler(ms, theLaw.Tiers)
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

	ms = types.NewMembersByTier()
	addMembers(ms, 100, T1)
	addMembers(ms, 200, T2)
	addMembers(ms, 1000, T3)

	theLaw.Tiers.Iterate("", "", func(key string, value interface{}) bool {
		po := value.(types.Tier).PowerHandler(ms, theLaw.Tiers)
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

func TestLawString(t *testing.T) {
	ms := types.NewMembersByTier()
	addMembers(ms, 10, T1)
	addMembers(ms, 20, T2)
	addMembers(ms, 100, T3)

	println(lawStringifier(theLaw, ms))
}

func TestProposeRunSeveralProposals(t *testing.T) {
	std.TestSetOrigCaller(moul)

	pn := MemberProposalAdd(testutils.TestAddress("antonio"), T3)
	if pn != 0 {
		t.Fatal("wrong proposal number", pn)
	}

	pn = LawProposalChange(theLaw) // propose the same law just to check the proposal output
	if pn != 1 {
		t.Fatal("wrong proposal number", pn)
	}

	pn = NewProposal(&types.Proposal{
		Title:       "Custom Proposal",
		Description: "This is a custom proposal",
		Executor: types.NewExecutor(func(m types.Metadata) error {
			if ruler.CurrentDAO != std.PrevRealm().PkgPath() {
				panic("trying to execute with the wrong realm!: " + std.PrevRealm().PkgPath())
			}
			return nil
		}, &propMetadata{}),
	})
	if pn != 2 {
		t.Fatal("wrong proposal number", pn)
	}

	std.TestSetRealm(std.NewUserRealm(moul))

	Vote(0, "YES")
	Vote(1, "NO")

	fmt.Println(Render(""))
}

type propMetadata struct {
}

func (m *propMetadata) IsMetadata() {}
func (m *propMetadata) String() string {
	return "proposal metadata!"
}

func addMembers(ms types.MembersByTier, c int, tier string) {
	mt := avl.NewTree()
	for i := 0; i < c; i++ {
		addr := std.Address(strconv.Itoa(i))
		mt.Set(string(addr), &types.Member{})
	}

	ms.Set(tier, mt)
}
