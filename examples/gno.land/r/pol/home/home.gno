package home

import (
	"errors"
	"net/url"
	"std"
	"strconv"
	"strings"

	"gno.land/p/moul/md"
	"gno.land/p/moul/mdtable"
	"gno.land/p/moul/txlink"
	"gno.land/p/pol/dao"
	sdao "gno.land/p/pol/dao/simple"
	"gno.land/r/pol/gnominers"

	// "gno.land/p/pol/uufmt"
	_ "gno.land/p/pol/nfmt"
	// "gno.land/p/pol/unsafe"
)

var myPage = std.CurrentRealm().PkgPath()[:len(std.CurrentRealm().PkgPath())-len("home")]
var here = std.CurrentRealm().PkgPath()[len(std.ChainDomain()):]
var polDao *sdao.SimpleDAO
var voteId int = 0
var counter uint = 0

const br = "\n\n"

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

func ProposeSomething(prop *sdao.SimpleProposal) error {
	if !strings.HasPrefix(std.PreviousRealm().PkgPath(), myPage) {
		return errors.New("needs to be from a r/pol base")
	}
	polDao.Propose(prop)
	return nil
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
	if prop.GetAuthor() != "" && prop.GetAuthor() != actor {
		panic("the actor is not the author")
	}
	err = prop.ExecuteIfAccepted()
	if err != nil {
		panic(err)
	}
}

// Render functions

func renderIntroduction() string {
	out := md.H1("Welcome to the `pol` gno page !")
	out += "Hi my name is Paul, I'm a developer and welcome to my page !" + br
	out += md.H3("Go check out the latest version of my super game: " + md.Link("GnoMiners", gnominers.LatestRelease()) + " !")
	out += md.H3("Check out this " + md.Link("DAO implementation", here + ":dao") + " that uses " + md.InlineCode("p/pol/dao") + " !") + br
	return out
}

func renderDAOStuff() string {
	// display the infos
	out := md.H1("The pol's DAO")
	out += "Go back to " + md.Link("home", here) + br

	out += "<gno-columns>" + br
	out += md.H2("About")
	out += "Here is my DAO, based on the " + md.InlineCode("SimpleDAO")
	out += " in " + md.InlineCode("/p/pol/dao") + ", go check it out !" + br

	out += "You can " + md.Link("Join", txlink.Call("JoinDao"))
	out += " or " + md.Link("Leave", txlink.Call("LeaveDao")) + " the dao. " + br

	dummyVoteLink := txlink.Call("Proposal", "title", "Dummy vote #" + strconv.Itoa(voteId), "description", "Do we increment the counter ?")
	out += "You can also post a " + md.Link("Dummy Vote", dummyVoteLink) + " to test. (if you joined)" + br
	out += "The counter is now at " + strconv.Itoa(int(counter)) + "." + br

	// display the members
	t := mdtable.Table{ Headers: []string{ "Member List" } }
	for _, m := range polDao.GetMembers() {
		t.Append([]string{m.Address().String()})
	}
	out += t.String()
	out += "|||" + br

	// display the dao proposals
	out += md.H2("Propositions")
	out += md.HorizontalRule()
	props := polDao.GetProposals()
	for propID := len(props) - 1; propID >= 0; propID-- {
		iprop := props[propID]
		prop, ok := iprop.(*sdao.SimpleProposal)
		if !ok {
			continue
		}
		if prop.IsEnded() {
			out += md.Blockquote("[ended] " + prop.Title) + md.HorizontalRule()
			continue
		}

		out += md.H3(prop.Title)
		out += prop.Description + br
		addr := iprop.GetAuthor().Address().String()
		if addr == "" {
			addr = "nobody"
		}
		out += md.Blockquote("by " + addr) + br
		out += strconv.Itoa(int(prop.HowManyVoters())) + " votes -- "

		var lnk string
		for k, v := range iprop.GetChoicesNames() {
			lnk = txlink.Call("Vote", "vote", strconv.Itoa(int(k)), "id",  strconv.Itoa(propID))
			out += md.Link(v, lnk) + " "
		}
		lnk = txlink.Call("EndVote", "id", strconv.Itoa(propID))
		out += "(" + md.Link("End Vote", lnk) + ")" + br + md.HorizontalRule()
	}
	out += br
	out += "</gno-columns>"
	return out
}

func Render(path string) string {
	if path == "dao" {
		return renderDAOStuff()
	}
	return renderIntroduction()
}
