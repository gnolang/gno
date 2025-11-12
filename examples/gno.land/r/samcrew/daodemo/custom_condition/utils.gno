package custom_condition

import (
	"strings"
	"unicode"

	"gno.land/p/mason/md"
	"gno.land/p/moul/txlink"
	"gno.land/p/nt/ufmt"
	"gno.land/p/samcrew/basedao"
	"gno.land/p/samcrew/daokit"
)

// initDemoProposals creates initial proposals to demonstrate custom condition functionality
func initDemoProposals() {
	demoProposals := []struct {
		address  string
		roles    []string
		proposer string
		title    string
		desc     string
	}{
		{
			"g1demomember001234567890abcdefghijklmnopqr1",
			[]string{},
			"g1demo1234567890abcdefghijklmnopqrstuvwxy2",
			"Add Alice as member with no role",
			"This proposal will add Alice to our DAO without roles.",
		},
		{
			"g1demomember001234567890abcdefghijklmnopqr2",
			[]string{"Role"},
			"g1demo1234567890abcdefghijklmnopqrstuvwxy3",
			"Add Bob as member with Role",
			"This proposal will add Bob to our DAO with the 'Role' role.",
		},
		{
			"g1demomember001234567890abcdefghijklmnopqr3",
			[]string{},
			"g1demo1234567890abcdefghijklmnopqrstuvwxy2",
			"Add Charlie as member with no role",
			"This proposal will add Charlie to our DAO without roles.",
		},
	}

	for _, prop := range demoProposals {
		action := basedao.NewAddMemberAction(&basedao.ActionAddMember{
			Address: address(prop.address),
			Roles:   prop.roles,
		})

		req := daokit.ProposalRequest{
			Title:       prop.title,
			Description: prop.desc,
			Action:      action,
		}
		daoPrivate.Core.Propose(prop.proposer, req)
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
//   - roles: Comma-separated string of roles to assign to the member (e.g., "admin,moderator" or "voter")
func ProposeAddMember(cur realm, address address, roles string) {
	rs := strings.Split(roles, ",")
	for i, s := range rs {
		rs[i] = strings.TrimFunc(s, unicode.IsSpace)
	}

	payload := basedao.ActionAddMember{
		Address: address,
		Roles:   rs,
	}
	action := basedao.NewAddMemberAction(&payload)

	proposal := daokit.ProposalRequest{
		Title:       ufmt.Sprintf("Add member %s with roles %s", address, strings.Join(rs, ", ")),
		Description: ufmt.Sprintf("This proposal will add %s as a member with roles: %s", address, strings.Join(rs, ", ")),
		Action:      action,
	}

	Propose(cur, proposal)
}

func renderDemo() string {
	s := ""
	s += "# ⚙️ Custom Condition Demo\n\n"
	s += "This showcases how to create custom governance rules.\n\n"
	s += "## ℹ️ How it Works\n\n"
	s += "1. **Join the DAO** using the " + md.Link("AddMember", txlink.Call("AddMember")) + " function.\n\n"
	s += "2. **Create a Proposal** using " + md.Link("ProposeAddMember", txlink.Call("ProposeAddMember")) + " function with these parameters:\n\n"
	s += "- `address`: The address of the member to add\n"
	s += "- `roles`: Comma-separated roles (e.g., \"Role\")\n\n"
	s += "3. " + md.Link("Vote", txlink.Call("Vote")) + " on proposals using their ID.\n\n"
	s += "4. " + md.Link("Execute", txlink.Call("Execute")) + " proposals that have passed (met the required voting conditions).\n\n"
	s += "   📝 **Note**: Only proposals that have met the governance conditions can be executed. You need at least 1 vote from members with no roles to execute a proposal.\n\n"
	s += "## 🎯 Custom Condition: NoRole\n\n"
	s += "This DAO uses a initial custom condition called **NoRole** that requires:\n"
	s += "- At least **1 vote** from members who have **no assigned roles**\n"
	s += "## 📋 Available Actions\n\n"
	s += "### Member Management\n"
	s += "- " + md.Link("🔗 Add Yourself as Member", txlink.Call("AddMember")) + " - Join the DAO to participate in governance\n"
	s += "- " + md.Link("🔗 Propose Add Member", txlink.Call("ProposeAddMember")) + " - Create a proposal to add a new member with specific roles\n\n"
	s += "### Other Demos\n"
	s += "- " + md.Link("🔗 Simple DAO Demo", "/r/samcrew/daodemo/simple_dao") + " - Basic DAO functionality\n"
	s += "- " + md.Link("🔗 Custom Resource Demo", "/r/samcrew/daodemo/custom_resource") + " - Custom DAO actions\n\n"
	s += "*This is a demonstration DAO built with gno.land and daokit*"
	return s
}
