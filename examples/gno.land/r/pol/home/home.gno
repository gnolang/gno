package home

import (
	"std"
	"strconv"
	"net/url"

	"gno.land/p/moul/txlink"
	"gno.land/p/pol/dao"
	sdao "gno.land/p/pol/dao/simple"
)

var polDao *sdao.SimpleDAO
var voteId int = 0
var counter uint = 0

// Init

func init() {
	polDao = sdao.NewSimpleDAO()
}

// Exported functions

func JoinDao() {
	err := polDao.AddMember(dao.Member(std.PreviousRealm().Address()))
	if err != nil {
		panic(err)
	}
}

func LeaveDao() {
	err := polDao.ExcludeMember(dao.Member(std.PreviousRealm().Address()))
	if err != nil {
		panic(err)
	}
}

func Proposal(title string, description string) {
	// unescape the strings
	title, err := url.QueryUnescape(title)
	if err != nil {
		panic(err)
	}
	description, err = url.QueryUnescape(description)
	if err != nil {
		panic(err)
	}

	// create a member & check if its in the dao
	actor := dao.Member(std.PreviousRealm().Address())
	polDao.AssertUserIsMember(actor)

	// create the proposal
	// POTENTIAL ISSUE: I had to reverse it title & description ? ill see why
	voteId += 1
	prop := sdao.NewSimpleProposal(actor, description, title, func()error {
		counter += 1
		return nil
	})
	polDao.Propose(prop)
}

func Vote(vote string, id string) {
	polDao.AssertUserIsMember(dao.Member(std.PreviousRealm().Address()))

	// get id & vote
	propi, err := strconv.Atoi(vote)
	if err != nil {
		panic(err)
	}
	prop := dao.ProposalChoice(uint(propi))
	propID, err := strconv.Atoi(id)
	if err != nil {
		panic(err)
	}
	allProps := polDao.GetProposals()
	if propID >= len(allProps) {
		panic("propID is greater than all proposals")
	}

	// make the user vote
	err = allProps[propID].Vote(prop)
	if err != nil {
		panic(err)
	}
}

func EndVote(id string) {
	actor := dao.Member(std.PreviousRealm().Address())
	polDao.AssertUserIsMember(actor)

	// get the id
	propID, err := strconv.Atoi(id)
	if err != nil {
		panic(err)
	}
	allProps := polDao.GetProposals()
	if propID >= len(allProps) {
		panic("propID is greater than all proposals")
	}

	// end the vote (only if its the user)
	prop := allProps[propID]
	if prop.GetAuthor() != actor {
		panic("the actor is not the author")
	}
	err = prop.ExecuteIfAccepted()
	if err != nil {
		panic(err)
	}
}

// Render functions

func renderIntroduction() string {
	out := "# Welcome to the `pol` gno page !\n"
	out += "Hi my name is Paul, I'm a developer and welcome to my page !\n\n"
	out += "### Go check out my super game: [gnominers](/r/pol/gnominers) !\n\n"
	out += "### Check out my own [DAO implementation](/r/pol/home:dao) !"
	return out
}

func renderDAOStuff() string {
	// display the infos
	out := "# The pol's DAO\n"
	out += "Go back to [home](/r/pol/home)\n\n"
	out += "## About\n"
	out += "Here is my DAO, based on the `SimpleDAO` in `/p/pol/dao`, go check it out !\n\n"
	joinLink := txlink.Call("JoinDao")
	leaveLink := txlink.Call("LeaveDao")
	dummyVoteLink := txlink.Call("Proposal", "title", "Dummy vote #" + strconv.Itoa(voteId), "description", "Do we increment the counter ?")
	out += "You can [Join](" + joinLink + ") or [Leave](" + leaveLink + ") the sdao.\n"
	out += "You can also post a [Dummy Vote](" + dummyVoteLink + ") to test. (if you joined)\n\n"
	out += "The counter is now at " + strconv.Itoa(int(counter)) + ".\n\n"

	// display the members
	out += "|Member List|\n|-|\n"
	for _, m := range polDao.GetMembers() {
		out += "|" + m.Address().String() + "|\n"
	}
	out += "\n"

	// display the dao proposals
	out += "## Propositions\n"
	out += "---\n"
	props := polDao.GetProposals()
	for propID := len(props) - 1; propID >= 0; propID-- {
		iprop := props[propID]
		prop, ok := iprop.(*sdao.SimpleProposal)
		if !ok {
			continue
		}
		if prop.IsEnded() {
			out += "> [ended] " + prop.Title + "\n\n---\n\n"
			continue
		}
		out += "### " + prop.Title + "\n"
		out += prop.Description + "\n\n"
		out += "> by " + iprop.GetAuthor().Address().String() + "\n\n"
		out += strconv.Itoa(int(prop.HowManyVoters())) + " votes -- "
		var lnk string
		for k, v := range polDao.GetChoicesNames() {
			lnk = txlink.Call("Vote", "vote", strconv.Itoa(int(k)), "id",  strconv.Itoa(propID))
			out += "[" + v + "](" + lnk + ") "
		}
		lnk = txlink.Call("EndVote", "id", strconv.Itoa(propID))
		out += "([End Vote](" + lnk + "))\n\n---\n\n" 
	}
	out += "\n"
	return out
}

func Render(path string) string {
	if path == "dao" {
		return renderDAOStuff()
	}
	return renderIntroduction()
}
