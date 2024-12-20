package govdao

import (
	"errors"
	"std"

	"gno.land/p/demo/ufmt"
)

// changeTheLawCallback implements the function needed for an executor,
// making possible to create a proposal for changing the DAO's law.
func changeTheLawCallback(m Metadata) error {
	mm, ok := m.(*lawMetadata)
	if !ok {
		return errors.New("wrong metadata type for changing the law")
	}

	theLaw = mm.NewLaw

	return nil
}

type lawMetadata struct {
	NewLaw *Law
}

func (a *lawMetadata) IsMetadata()    {}
func (a *lawMetadata) String() string { return lawStringifier(a.NewLaw, members) }

type lawStringMetadata struct {
	newStringTemplate func(law *Law, allMembers MembersByTier) string
}

func (a *lawStringMetadata) IsMetadata()    {}
func (a *lawStringMetadata) String() string { return a.newStringTemplate(theLaw, members) }

func changeTheLawStringCallback(m Metadata) error {
	mm, ok := m.(*lawStringMetadata)
	if !ok {
		return errors.New("wrong metadata type for changing the law string template")
	}

	lawStringifier = mm.newStringTemplate

	return nil
}

// addNewMemberCallback implements the function needed for an executor,
// making possible to create a proposal for adding new members to the DAO
func addNewMemberCallback(m Metadata) error {
	// TODO:
	/*
		T1 criteria includes expertise in categories, significant contributions, and demonstration of value alignment. T2 criteria includes expertise in categories, continuing contributions. T3 criteria includes expertise in categories, continuing contributions.
		T1 and T2 members are added via individual proposals, each with a markdown resume/portfolio application document.
	*/
	/*
		Target is minimum 70 members after 7 years.
		If the minimum is not reached, AND 2 years has passed, AND no members are added for the quarter, AND there are Qualified Candidates, 1 Qualified Candidate may be elected by GovDAO, but with T1 abstaining.
	*/
	mm, ok := m.(*addMemberMetadata)
	if !ok {
		return errors.New("wrong metadata type for adding or modifying a new member")
	}

	members.SetMember(mm.Tier, mm.Addr, mm.Member)

	return nil
}

type addMemberMetadata struct {
	Member *Member
	Tier   string
	Addr   std.Address
}

func (a *addMemberMetadata) IsMetadata() {}
func (a *addMemberMetadata) String() string {
	return ufmt.Sprintf("A new member with address %v is proposed to be on tier %v.", a.Addr, a.Tier)
}

func promoteNewMemberCallback(m Metadata) error {
	mm, ok := m.(*addMemberMetadata)
	if !ok {
		return errors.New("wrong metadata type for adding or modifying a new member")
	}

	members.SetMember(mm.Tier, mm.Addr, mm.Member)

	return nil
}

type promoteMemberMetadata struct {
	FromTier string
	ToTier   string
	Addr     std.Address
}

func (a *promoteMemberMetadata) IsMetadata() {}
func (a *promoteMemberMetadata) String() string {
	return ufmt.Sprintf("A new member with address %v will be promoted from tier %v to tier %v.", a.Addr, a.FromTier, a.ToTier)
}
