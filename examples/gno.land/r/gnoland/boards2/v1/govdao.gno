package boards2

// TODO: This file must be moved to a sub-realm, requires PR #5037

import (
	"errors"
	"strings"

	"gno.land/p/moul/md"
	"gno.land/p/nt/ufmt"

	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/v3/impl"
	"gno.land/r/gov/dao/v3/memberstore"
)

// ProposeRealmMembersAdd creates a new GovDAO proposal to add one or more users to Boards2 realm.
// All users are added as `owners` when proposal passes.
// Only tier1 and tier2 GovDAO members can vote on this type of proposal.
func ProposeRealmMembersAdd(users ...address) dao.ProposalRequest {
	if len(users) == 0 {
		panic("proposal requires at least one user account")
	}

	header := ufmt.Sprintf(
		"Proposal adds the following realm `owners` to %s:",
		md.Link("Boards2", "/r/gnoland/boards2/v1"),
	)

	var desc strings.Builder
	desc.WriteString(md.Paragraph(header))
	for _, addr := range users {
		if !addr.IsValid() {
			panic("invalid address: " + addr.String())
		}

		// Make sure current user is not a member yet
		if gPerms.HasUser(addr) {
			panic("user is already a member: " + addr.String())
		}

		desc.WriteString(md.BulletItem(userLink(addr)))
	}

	cb := func(realm) error {
		for _, addr := range users {
			if gPerms.HasUser(addr) {
				return errors.New("user is already a member: " + addr.String())
			}

			gPerms.SetUserRoles(addr, RoleOwner)
		}
		return nil
	}

	return dao.NewProposalRequestWithFilter(
		"Add Boards2 realm members",
		desc.String(),
		dao.NewSimpleExecutor(cb, ""),
		impl.FilterByTier{Tier: memberstore.T2},
	)
}

// ProposeRealmMembersRemove creates a new GovDAO proposal to remove one or more Boards2 realm members.
// Only tier1 and tier2 GovDAO members can vote on this type of proposal.
func ProposeRealmMembersRemove(members ...address) dao.ProposalRequest {
	if len(members) == 0 {
		panic("proposal requires at least one member account")
	}

	header := ufmt.Sprintf(
		"Proposal removes the following realm members from %s:",
		md.Link("Boards2", "/r/gnoland/boards2/v1"),
	)

	var desc strings.Builder
	desc.WriteString(md.Paragraph(header))
	for _, addr := range members {
		if !addr.IsValid() {
			panic("invalid address: " + addr.String())
		}

		// Make sure current user is a member
		if !gPerms.HasUser(addr) {
			panic("user is not a member: " + addr.String())
		}

		desc.WriteString(md.BulletItem(userLink(addr)))
	}

	cb := func(realm) error {
		for _, addr := range members {
			if !gPerms.HasUser(addr) {
				return errors.New("user is not a member: " + addr.String())
			}

			gPerms.RemoveUser(addr)
		}
		return nil
	}

	return dao.NewProposalRequestWithFilter(
		"Remove Boards2 realm members",
		desc.String(),
		dao.NewSimpleExecutor(cb, ""),
		impl.FilterByTier{Tier: memberstore.T2},
	)
}
