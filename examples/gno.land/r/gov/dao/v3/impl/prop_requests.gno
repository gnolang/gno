package impl

import (
	"std"

	"gno.land/p/demo/ufmt"
	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/v3/memberstore"
)

type ChangeLawMetadata struct {
	Law *Law
}

func (a *ChangeLawMetadata) IsMetadata() {}
func (a *ChangeLawMetadata) String() string {
	return ufmt.Sprintf("A new Law is proposed:\n %v", a.Law)
}

func changeLawCallback(m dao.Metadata) error {
	amm, ok := m.(*ChangeLawMetadata)
	if !ok {
		panic("metadata type for changing the law is wrong")
	}

	law = amm.Law

	return nil
}

func NewChangeLawRequest(law *Law) dao.ProposalRequest {
	member, _ := memberstore.Get().GetMember(std.OriginCaller())
	if member == nil {
		panic("proposer is not a member")
	}

	return dao.NewProposalRequestBuilder("Change Law Proposal").
	WithDescription("This proposal is looking to change the actual govDAO Law").
	WithCallback(changeLawCallback).
	WithMetadata(&ChangeLawMetadata{
		Law: law,
	}).Build()
}

type AddMemberMetadata struct {
	Member    *memberstore.Member
	Portfolio string
	Tier      string
	Addr      std.Address
	Caller    *memberstore.Member
}

func (a *AddMemberMetadata) IsMetadata() {}
func (a *AddMemberMetadata) String() string {
	return ufmt.Sprintf("A new member with address %v is proposed to be on tier %v. Provided Portfolio information:\n\n%v", a.Addr, a.Tier, a.Portfolio)
}

func addMemberCallback(m dao.Metadata) error {
	amm, ok := m.(*AddMemberMetadata)
	if !ok {
		panic("metadata type for adding a member is wrong")
	}

	amm.Caller.RemoveInvitationPoint()

	return memberstore.Get().SetMember(amm.Tier, amm.Addr, amm.Member)
}

func NewAddMemberRequest(addr std.Address, tier string, portfolio string) dao.ProposalRequest {
	_, ok := memberstore.Tiers.GetTier(tier)
	if !ok {
		panic("provided tier does not exists")
	}

	if tier != memberstore.T1 && tier != memberstore.T2 {
		panic("Only T1 and T2 members must be added by proposal. To add a T3 member use AddMember function directly.")
	}

	if portfolio == "" {
		panic("A portfolio for the proposed member is required")
	}

	member, _ := memberstore.Get().GetMember(std.OriginCaller())
	if member == nil {
		panic("proposer is not a member")
	}

	if member.InvitationPoints <= 0 {
		panic("proposer does not have enough invitation points for inviting new people to the board")
	}

	return dao.NewProposalRequestBuilder("New Member Prorposal").
	WithDescription("This proposal is looking to add a new member to the board.").
	WithCallback(addMemberCallback).
	WithMetadata(&AddMemberMetadata{
		Member:    memberByTier(tier),
		Portfolio: portfolio,
		Tier:      tier,
		Addr:      addr,
		Caller:    member,
	}).Build()
}

type RemoveMemberMetadata struct {
	Addr   std.Address
	Tier   string
	Reason string
}

func (a *RemoveMemberMetadata) IsMetadata() {}
func (a *RemoveMemberMetadata) String() string {
	return ufmt.Sprintf("Member with address %v will be withdrawn.\n\n REASON: %v.", a.Addr, a.Reason)
}

func removeMemberCallback(m dao.Metadata) error {
	rmm, ok := m.(*RemoveMemberMetadata)
	if !ok {
		panic("metadata type for removing a member is wrong")
	}

	memberstore.Get().RemoveMember(rmm.Addr)

	return nil
}

func NewWithdrawMemberRequest(addr std.Address, reason string) dao.ProposalRequest {
	member, tier := memberstore.Get().GetMember(addr)
	if member == nil {
		panic("user we want to remove not found")
	}

	if tier == memberstore.T1 && reason == "" {
		panic("T1 user removals must contains a reason.")
	}

	return dao.NewProposalRequestBuilder("Member Withdrawn").
	WithDescription("This proposal is looking to remove a member from the board.").
	WithCallback(removeMemberCallback).
	WithMetadata(&RemoveMemberMetadata{
		Addr:   addr,
		Tier:   tier,
		Reason: reason,
	}).Build()
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

func promoteMemberCallback(m dao.Metadata) error {
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

func NewPromoteMemberRequest(addr std.Address, fromTier string, toTier string) dao.ProposalRequest {
	return dao.NewProposalRequestBuilder("Member Promotion").
	WithDescription("This proposal is looking to promote a member to an upper tier.").
	WithCallback(promoteMemberCallback).
	WithMetadata(&PromoteMemberMetadata{
		Addr:     addr,
		FromTier: fromTier,
		ToTier:   toTier,
		Member:   memberByTier(toTier),
	}).Build()
}

func memberByTier(tier string) *memberstore.Member {
	switch tier {
	case memberstore.T1:
		t, _ := memberstore.Tiers.GetTier(memberstore.T1)
		return &memberstore.Member{
			InvitationPoints: t.InvitationPoints,
		}
	case memberstore.T2:
		t, _ := memberstore.Tiers.GetTier(memberstore.T2)
		return &memberstore.Member{
			InvitationPoints: t.InvitationPoints,
		}
	case memberstore.T3:
		t, _ := memberstore.Tiers.GetTier(memberstore.T3)
		return &memberstore.Member{
			InvitationPoints: t.InvitationPoints,
		}
	default:
		panic("member not found by the specified tier")
	}
}
