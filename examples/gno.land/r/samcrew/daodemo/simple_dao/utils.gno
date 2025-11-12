package simple_dao

import (
	"strings"
	"unicode"

	"gno.land/p/mason/md"
	"gno.land/p/moul/txlink"
	"gno.land/p/nt/ufmt"
	"gno.land/p/samcrew/basedao"
	"gno.land/p/samcrew/daocond"
	"gno.land/p/samcrew/daokit"
)

// Creates a number of initial proposals and votes.
// It divides the proposals into 4 parts: AddMember, RemoveMember, AssignRole, EditProfile.
// It initializes votes for each proposal type.
func initProposals(nb int) {
	const financeOfficer = "g1demo1234567890abcdefghijklmnopqrstuvwxyz"
	const nbTypes = 4
	proposalsPerType := nb / nbTypes
	offsets := []int{0, proposalsPerType, 2 * proposalsPerType, 3 * proposalsPerType, nb} // AddMember, Remove, AssignRole, EditProfile

	lastAddr := financeOfficer

	for t, start := range offsets[:nbTypes] {
		end := offsets[t+1]
		for i := start; i < end; i++ {
			addr := ufmt.Sprintf("g1fakemember000000000000000000000000000%02d", i)
			var action daokit.Action
			var title string
			var description string

			// CREATE
			switch t {
			case 0: // AddMember
				action = basedao.NewAddMemberAction(&basedao.ActionAddMember{Address: address(addr), Roles: []string{}})
				title = ufmt.Sprintf("Add member %s with no role", addr)
				description = ufmt.Sprintf("This proposal will add %s as a new member of the DAO without any specific roles. They will be able to participate in governance and vote on proposals.", addr)
			case 1: // RemoveMember
				action = basedao.NewRemoveMemberAction(address(addr))
				title = ufmt.Sprintf("Remove member %s", addr)
				description = ufmt.Sprintf("This proposal will remove %s from the DAO membership. They will no longer be able to participate in governance or vote on proposals.", addr)
			case 2: // AssignRole
				role := "public-relationships"
				action = basedao.NewAssignRoleAction(&basedao.ActionAssignRole{Address: address(financeOfficer), Role: role})
				title = ufmt.Sprintf("Assign role %s for %s", role, addr)
				description = ufmt.Sprintf("This proposal will assign the '%s' role to the finance officer. This role grants additional responsibilities and permissions within the DAO.", role)
			case 3: // EditProfile
				action = basedao.NewEditProfileAction([2]string{"Bio", "A demo DAO showcasing governance and community decision-making."})
				title = "Edit the DAO's profile"
				description = "This proposal will update the DAO's bio with a simple description of our purpose as a demonstration DAO."
			}
			req := daokit.ProposalRequest{
				Title:       title,
				Action:      action,
				Description: description,
			}
			id := daoPrivate.Core.Propose(lastAddr, req)

			// VOTE
			quart := i * 4 / nb
			switch quart {
			case 0: // VoteNo
				daoPrivate.Core.Vote(financeOfficer, id, daocond.VoteNo)
			case 1: // VoteYes
				daoPrivate.Core.Vote(financeOfficer, id, daocond.VoteYes)
			case 2: // VoteAbstain
				daoPrivate.Core.Vote(financeOfficer, id, daocond.VoteAbstain)
			case 3: // VoteYes + Execute
				daoPrivate.Core.Vote(financeOfficer, id, daocond.VoteYes)
				daoPrivate.Core.Execute(id)
			}

			lastAddr = addr
		}
	}
}

// Bypass limitation by adding yourself to the DAO.
// It is necessary to be part of the DAO to create a Proposal.
func AddMember(cur realm) {
	id := daoPrivate.CallerID()
	daoPrivate.Members.AddMember(id, make([]string, 0))
}

// Creates a Proposal to add a new member to the DAO with specified roles.
// This function exist to let users try the userflow of daokit with a simple MsgCall (maketx call) instead of a MsgRun.
// See why a run is necessary for creating a proposal -> https://docs.gno.land/users/interact-with-gnokey#run.
// Parameters:
//   - address: The std.Address of the member to be added
//   - roles: Comma-separated roles (e.g., "public-relationships,finance-officer" or "finance-officer")
func ProposeAddMember(cur realm, address address, roles string) {
	rs := strings.Split(roles, ",")
	for i, s := range rs {
		rs[i] = strings.TrimFunc(s, unicode.IsSpace)
	}
	req := daokit.ProposalRequest{
		Title: ufmt.Sprintf("Add member %s with roles %s", address, strings.Join(rs, ", ")),
		Action: basedao.NewAddMemberAction(&basedao.ActionAddMember{
			Address: address,
			Roles:   rs,
		}),
	}
	localDAO.Propose(req)
}

func renderDemo() string {
	s := ""
	s += "# 🏛️ Simple DAO Actions\n\n"
	s += "Welcome to Simple DAO! This is a demonstration of basic DAO functionality.\n\n"
	s += "## ℹ️ How it Works\n\n"
	s += "1. **Join the DAO** using the " + md.Link("AddMember", txlink.Call("AddMember")) + " function.\n\n"
	s += "2. **Create a pre-made Proposal** using " + md.Link("ProposeAddMember", txlink.Call("ProposeAddMember")) + "  function using this parameters:\n\n"
	s += "- `address`: The address of the member to add\n"
	s += "- `roles`: Comma-separated roles (e.g., \"public-relationships,finance-officer\" or \"finance-officer\")\n\n"
	s += "3. " + md.Link("Vote", txlink.Call("Vote")) + " on proposals using their ID.\n\n"
	s += "4. " + md.Link("Execute", txlink.Call("Execute")) + " proposals that have passed (met the required voting conditions).\n\n"
	s += "   📝 **Note**: Only proposals that have met the governance conditions can be executed. You need both **40% of member approval** AND at least **1 finance-officer** to execute a proposal.\n\n"
	s += "## 📋 Available Actions\n\n"
	s += "### Member Management\n"
	s += "- " + md.Link("🔗 Add Yourself as Member", txlink.Call("AddMember")) + " - Join the DAO to participate in governance\n"
	s += "- " + md.Link("🔗 Propose Add Member", txlink.Call("ProposeAddMember")) + " - Create a proposal to add a new member with specific roles\n\n"
	s += "### Other Demos\n"
	s += "- " + md.Link("🔗 Custom Resource Example", "/r/samcrew/daodemo/custom_resource") + " - See a demo of custom resource implementation\n"
	s += "- " + md.Link("🔗 Custom Condition Example", "/r/samcrew/daodemo/custom_condition") + " - See a demo of custom condition implementation\n\n"
	s += "*This is a demonstration DAO built with gno.land and daokit*"
	return s
}
