package impl

import (
	"errors"
	"std"

	"gno.land/p/demo/ufmt"
	"gno.land/r/gov/dao/proxy"
	"gno.land/r/gov/dao/v3/memberstore"
)

type ChangeLawMetadata struct {
	Law *Law
}

func (a *ChangeLawMetadata) IsMetadata() {}
func (a *ChangeLawMetadata) String() string {
	return ufmt.Sprintf("A new Law is proposed:\n %v", a.Law)
}

func changeLawCallback(m proxy.Metadata) error {
	amm, ok := m.(*ChangeLawMetadata)
	if !ok {
		panic("metadata type for changing the law is wrong")
	}

	law = amm.Law

	return nil
}

func NewChangeLawRequest(law *Law) proxy.ProposalRequest {
	member, _ := memberstore.Get().GetMember(std.GetOrigCaller())
	if member == nil {
		panic("proposer is not a member")
	}

	return proxy.ProposalRequest{
		Title:       "Change Law Proposal",
		Description: "This prpoposal is looking to change the actual govDAO Law",
		Callback: changeLawCallback,
		Metadata: &ChangeLawMetadata{
			Law: law,
		},
	}
}

type AddMemberMetadata struct {
	Member *memberstore.Member
	Tier   string
	Addr   std.Address
	Caller *memberstore.Member
}

func (a *AddMemberMetadata) IsMetadata() {}
func (a *AddMemberMetadata) String() string {
	return ufmt.Sprintf("A new member with address %v is proposed to be on tier %v.", a.Addr, a.Tier)
}

func addMemberCallback(m proxy.Metadata) error {
	amm, ok := m.(*AddMemberMetadata)
	if !ok {
		panic("metadata type for adding a member is wrong")
	}

	amm.Caller.RemoveInvitationPoint()

	return memberstore.Get().SetMember(amm.Tier, amm.Addr, amm.Member)
}

func NewAddMemberRequest(addr std.Address, tier string) proxy.ProposalRequest {
	_,ok := memberstore.Tiers.GetTier(tier)
	if !ok {
		panic("provided tier does not exists")
	}
	
	member, _ := memberstore.Get().GetMember(std.GetOrigCaller())
	if member == nil {
		panic("proposer is not a member")
	}

	if member.InvitationPoints <= 0 {
		panic("proposer does not have enough invitation points for inviting new people to the board")
	}

	return proxy.ProposalRequest{
		Title:       "New Member porposal",
		Description: "This prpoposal is looking to add a new member to the board.",
		Callback: addMemberCallback,
		Metadata: &AddMemberMetadata{
			Member: memberByTier(tier),
			Tier: tier,
			Addr: addr,
		},
	}
}

type RemoveMemberMetadata struct {
	Addr   std.Address
	Reason string
}

func (a *RemoveMemberMetadata) IsMetadata() {}
func (a *RemoveMemberMetadata) String() string {
	return ufmt.Sprintf("Member with address %v will be withdrawn. REASON: %v.", a.Addr, a.Reason)
}

func removeMemberCallback(m proxy.Metadata) error {
	rmm, ok := m.(*RemoveMemberMetadata)
	if !ok {
		panic("metadata type for removing a member is wrong")
	}

	member, tier := memberstore.Get().GetMember(std.GetOrigCaller())
	if member == nil {
		panic("user we tried to remove not found")
	}

	if tier == memberstore.T1 && rmm.Reason == "" {
		return errors.New("T1 user removals must contains a reason.")
	}

	memberstore.Get().RemoveMember(rmm.Addr)

	return nil
}

func NewWithdrawMemberRequest(addr std.Address, reason string) proxy.ProposalRequest {
	return proxy.ProposalRequest{
		Title:       "Member Withdrawn",
		Description: "This prpoposal is looking to remove a member from the board.",
		Callback: removeMemberCallback,
		Metadata: &RemoveMemberMetadata{
			Addr: addr,
			Reason: reason,
		},
	}
}

type PromoteMemberMetadata struct {
	FromTier string
	ToTier   string
	Addr     std.Address
	Member   *memberstore.Member
}

func (a *PromoteMemberMetadata) IsMetadata() {}
func (a *PromoteMemberMetadata) String() string {
	return ufmt.Sprintf("A new member with address %v will be promoted from tier %v to tier %v.", a.Addr, a.FromTier, a.ToTier)
}

func promoteMemberCallback(m proxy.Metadata) error {
	pmm, ok := m.(*PromoteMemberMetadata)
	if !ok {
		panic("metadata type for promoting a member is wrong")
	}

	prevTier := memberstore.Get().RemoveMember(pmm.Addr)
	if prevTier == "" {
		panic("member not found, so cannot be promoted")
	}

	if prevTier != pmm.FromTier {
		panic("previous tier changed from the one indicated in the proposal")
	}

	return memberstore.Get().SetMember(pmm.ToTier, pmm.Addr, pmm.Member)
}

func NewPromoteMemberRequest(addr std.Address, fromTier string, toTier string) proxy.ProposalRequest {
	return proxy.ProposalRequest{
		Title:       "Member Promotion",
		Description: "This prpoposal is looking to promote a member to an upper tier.",
		Callback: promoteMemberCallback,
		Metadata: &PromoteMemberMetadata{
			Addr: addr,
			FromTier: fromTier,
			ToTier: toTier,
			Member: memberByTier(toTier),
		},
	}
}

func memberByTier(tier string) *memberstore.Member {
	switch tier {
	case memberstore.T1:
		return &memberstore.Member{
			InvitationPoints:3,
		}
	case memberstore.T2:
		return &memberstore.Member{
			InvitationPoints:2,
		}
	case memberstore.T3:
		return &memberstore.Member{
			InvitationPoints:1,
		}
	default:
		return &memberstore.Member{}
	}
}