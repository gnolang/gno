package govdao

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
)

var theLaw *Law
var members MembersByTier

func init() {
	theLaw = &Law{
		VotingPercentageNeeded: 66,
		Payment:                90000,
	}

	theLaw.CheckMemberAddProposal = memberAddChecker
	theLaw.CheckMemberPromoteProposal = memberPromoteChecker
	theLaw.CheckMemberWithdrawnProposal = memberWithdrawnChecker
	theLaw.CheckProposalWithdrawn = proposalWithdrawnCheker

	theLaw.Tiers = TiersByName{avl.NewTree()}
	theLaw.Tiers.Set(T1, Tier{
		MinSize: func(membersByTier MembersByTier, tiersByName TiersByName) int {
			return 70
		},
		MaxSize: func(membersByTier MembersByTier, tiersByName TiersByName) int {
			return 0
		},
		BasePower: 3,
		PowerHandler: func(membersByTier MembersByTier, tiersByName TiersByName) float64 {
			t, _ := tiersByName.GetTier(T1)
			return t.BasePower
		},
	})

	theLaw.Tiers.Set(T2, Tier{
		MaxSize: func(membersByTier MembersByTier, tiersByName TiersByName) int {
			return membersByTier.GetTierSize(T1) * 4
		},
		MinSize: func(membersByTier MembersByTier, tiersByName TiersByName) int {
			return membersByTier.GetTierSize(T1) / 4
		},
		BasePower: 2,
		PowerHandler: func(membersByTier MembersByTier, tiersByName TiersByName) float64 {
			t1ms := float64(membersByTier.GetTierSize(T1))
			t1, _ := tiersByName.GetTier(T1)
			t2ms := float64(membersByTier.GetTierSize(T2))
			t2, _ := tiersByName.GetTier(T2)

			t1p := t1.BasePower * t1ms
			t2p := t2.BasePower * t2ms

			// capped to 2/3 of tier 1
			t1ptreshold := t1p * (2.0 / 3.0)
			if t2p > t1ptreshold {
				return t1ptreshold / t2ms
			}

			return t2.BasePower
		},
	})

	theLaw.Tiers.Set(T3, Tier{
		MaxSize: func(membersByTier MembersByTier, tiersByName TiersByName) int {
			return 0
		},
		MinSize: func(membersByTier MembersByTier, tiersByName TiersByName) int {
			return 0
		},
		BasePower: 1,
		PowerHandler: func(membersByTier MembersByTier, tiersByName TiersByName) float64 {
			t1ms := float64(membersByTier.GetTierSize(T1))
			t1, _ := tiersByName.GetTier(T1)
			t3ms := float64(membersByTier.GetTierSize(T3))
			t3, _ := tiersByName.GetTier(T3)

			t1p := t1.BasePower * t1ms
			t3p := t3.BasePower * t3ms

			//capped to 1/3 of tier 1
			t1ptreshold := t1p * (1.0 / 3.0)
			if t3p > t1ptreshold {
				return t1ptreshold / t3ms
			}

			return t3.BasePower
		},
	})

	members = NewMembersByTier()
	members.SetMember(T1, moul, &Member{})
}

func memberAddChecker(membersByTier MembersByTier, proposer std.Address, proposedTier string, proposedAddr std.Address) error {
	// TODO: implement logic to check add member proposals
	_, proposerTier := membersByTier.GetMember(proposer)
	// TODO: check invitation points
	switch proposerTier {
	case T3:
		return ufmt.Errorf("proposer has an invalid tier to add members to the DAO: %s", proposerTier)
	case T2:
		if proposedTier != T3 {
			return ufmt.Errorf("proposer with tier %s cannot add members with tier %s.", proposerTier, proposedTier)
		}
		return nil
	case T1:
		return nil
	case "":
		return ufmt.Errorf("user with address %v is not a member", proposer)
	default:
		return ufmt.Errorf("unknown proposer tier: %v", proposerTier)
	}

	return nil
}

func memberPromoteChecker(membersByTier MembersByTier, proposer std.Address, proposedTier string, proposedAddr std.Address) error {
	// TODO: implement logic to check promote member proposals
	//TODO check if the proposed member is on a lower tier

	_, proposerTier := membersByTier.GetMember(proposer)

	switch proposerTier {
	case T1:
		return nil
	case T2:
		if proposedTier == T1 {
			return ufmt.Errorf("proposer with tier %s cannot promote members with tier %s.", proposerTier, proposedTier)
		}

		return nil
	case T3:
		return ufmt.Errorf("proposer with tier %s cannot promote members with tier %s.", proposerTier, proposedTier)
	case "":
		return ufmt.Errorf("user with address %s is not a member", proposer)
	default:
		return ufmt.Errorf("unknown proposer tier: %s", proposerTier)
	}

	return nil
}

func memberWithdrawnChecker(membersByTier MembersByTier, proposer std.Address, proposedAddr std.Address) error {
	return nil // anyone can create a withdraw member proposal
}

func proposalWithdrawnCheker(membersByTier MembersByTier, proposer std.Address) error {
	// TODO: implement logic to check withdraw proposals
	return nil
}

var lawStringifier func(law *Law, allMembers MembersByTier) string = func(law *Law, allMembers MembersByTier) string {
	out := "Law Definition:\n\n"
	out += ufmt.Sprintf("The total voting power needed to accept a proposal is %d%%.\n\n", law.VotingPercentageNeeded)
	out += ufmt.Sprintf("The law defines the existence of %d tiers:\n", law.Tiers.Size())
	law.Tiers.Iterate("", "", func(tn string, value interface{}) bool {
		t := value.(Tier)
		capped := t.PowerHandler(allMembers, law.Tiers)
		out += ufmt.Sprintf("- %s with a base power per member of %v.\n", tn, t.BasePower)
		out += ufmt.Sprintf("\tThe actual capped power for this tier taking into account actual members is %v.\n", capped)

		tierSize := allMembers.GetTierSize(tn)
		out += ufmt.Sprintf("\tDesired max amount of member is %d and the desired min is %d. There are today %d members on that tier.\n",
			t.MaxSize(allMembers, law.Tiers), t.MinSize(allMembers, law.Tiers), tierSize)

		return false
	})

	return out
}

const (
	T1 = "T1"
	T2 = "T2"
	T3 = "T3"
)

var moul std.Address = "g1manfred47kzduec920z88wfr64ylksmdcedlf5" // @moul
