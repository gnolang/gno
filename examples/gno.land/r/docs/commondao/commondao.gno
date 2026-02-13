package commondao

import (
	"std"
	"strconv"

	"gno.land/p/moul/md"
	"gno.land/p/moul/mdtable"
	"gno.land/p/moul/txlink"
	"gno.land/p/nt/commondao"
)

const (
	chaegnal = "g1dzzx5qwcmawqnhhd76ehkteas5v2dg42hqnd7g"
	leon     = "g125em6arxsnj49vx35f0n0z34putv5ty3376fg5"
	manfred  = "g1manfred47kzduec920z88wfr64ylksmdcedlf5"
)

// Creates a DAO with Active and Finished proposal storage
// Also, it creates a member storage
// Other options can be added, all of which can be found in 'p/nt/commondao/commondao_options'
var DAO = commondao.New(
	commondao.WithName("CommonDAO Documentation Realm"), // adds a name to the common DAO
	commondao.WithDescription("An example DAO"),         // adds a description to the common DAO
	commondao.WithMember(chaegnal),                      // adds the creator of the DAO to its member storage
	commondao.WithMember(leon),
	commondao.WithMember(manfred),
)

func Render(url string) string {
	out := md.H1("A CommonDao Example")

	if url == "join" {
		out += displayHidden()
		return out
	}

	out += md.Paragraph("This realm is meant to demonstrate how to use the [/p/nt/commondao](/p/nt/commondao) pure package.")
	out += md.Paragraph("The commondao package is meant to act as a DAO factory to use in any realms easily.")

	out += md.Paragraph("In the source code, you'll find out how to create a basic DAO and where to look for more informations about more advanced options")

	out += displayDaoOverview()

	return out
}

func displayMembers() string {
	table := mdtable.Table{
		Headers: []string{"Address"},
	}

	members := DAO.Members()

	// Iterates through the members of the DAO and adds them to the table
	if members.Size() == 0 {
		return md.Paragraph("The DAO has no members yet.")
	} else {
		members.IterateByOffset(0, members.Size(), func(adr std.Address) bool {
			table.Append([]string{adr.String()})
			return false
		})
	}

	return table.String()
}

// dislayActiveProposals displays the active proposals of the DAO
func displayActiveProposals() string {
	table := mdtable.Table{
		Headers: []string{"Address", "Title", "Description", "", "Votes", "Action"},
	}

	DAO.ActiveProposals().Iterate(0, DAO.ActiveProposals().Size(), false, func(p *commondao.Proposal) bool {
		yay := p.VotingRecord().VoteCount(commondao.ChoiceYes)
		nay := p.VotingRecord().VoteCount(commondao.ChoiceNo)

		voteYay := md.Link("Yay", txlink.Call("Vote",
			"vote", "true",
			"proposalID", strconv.FormatUint(p.ID(), 10),
			"reason", "",
		))

		voteNay := md.Link("Nay", txlink.Call("Vote",
			"vote", "false",
			"proposalID", strconv.FormatUint(p.ID(), 10),
			"reason", "",
		))

		table.Append([]string{
			p.Creator().String(),
			p.Definition().Title(),
			p.Definition().Body(),
			"Yay: " + strconv.Itoa(yay) + " / Nay: " + strconv.Itoa(nay),
			voteYay + " | " + voteNay,
			md.Link("Check", txlink.Call("ExecutionPassedProposals",
				"proposalID", strconv.FormatUint(p.ID(), 10),
			)),
		})

		return false
	})

	return table.String()
}

// displayDaoOverview displays the overview of the DAO
func displayDaoOverview() string {
	out := md.H2("DAO Overview")
	out += md.Bold("Title:") + " " + DAO.Name() + "\n\n"
	out += md.Bold("Description:") + " " + DAO.Description() + "\n\n"

	out += md.H3("DAO's members")
	out += displayMembers()

	out += md.H3("DAO's Active Proposals")
	if DAO.ActiveProposals().Size() == 0 {
		out += md.Paragraph("The DAO has no active proposals.")
	} else {
		out += displayActiveProposals()
	}

	return out
}

// displayHidden checks if the URL contains the hidden parameter and displays a message if it does
func displayHidden() string {
	out := md.H2("You found the secret of this realm!")

	out += md.Paragraph("Well done on reading the documentation of the CommonDAO package!")
	out += md.Paragraph("You can request to joins the DAO by clicking the button below:")
	out += md.Link("Request to join the DAO", txlink.Call("CreateProposal",
		"title", "Join the DAO",
		"desc", "I want to join the DAO as a member.",
	))

	return out
}
