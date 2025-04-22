package commondao

import (
	"std"
	"strconv"
	"strings"
	"time"

	"gno.land/p/demo/mux"
	"gno.land/p/demo/ufmt"
	"gno.land/p/jeronimoalbi/pager"
	"gno.land/p/moul/realmpath"
	"gno.land/p/moul/txlink"
	"gno.land/p/nt/commondao"

	"gno.land/r/sys/users"
)

const dateFormat = "Mon, 02 Jan 2006 03:04pm MST"

func Render(path string) string {
	router := mux.NewRouter()
	router.HandleFunc("", renderHome)
	router.HandleFunc("{daoID}", renderDAO)
	router.HandleFunc("{daoID}/settings", renderSettings)
	router.HandleFunc("{daoID}/proposals", renderProposalsList)
	router.HandleFunc("{daoID}/proposals/{proposalID}", renderProposal)
	router.HandleFunc("{daoID}/proposals/{proposalID}/vote/{address}", renderProposalVote)
	return router.Render(path)
}

func renderHome(res *mux.ResponseWriter, _ *mux.Request) {
	// TODO: Render a header text for users
	// TODO: Render generic links to vote or execute proposals
	res.Write("# Common DAO\n")
}

func renderDAO(res *mux.ResponseWriter, req *mux.Request) {
	dao := mustGetDAOFromRequest(req)
	o := getOptions(dao.ID())

	// Render header
	res.Write("# ")
	res.Write(dao.Name())
	res.Write("\n")

	if s := dao.Description(); s != "" {
		res.Write(s)
		res.Write("\n\n")
	}

	// Render main menu
	menu := []string{
		ufmt.Sprintf("[List Proposals](%s)", daoProposalsURL(dao)),
		ufmt.Sprintf("[View Settings](%s)", settingsURL(dao)),
	}

	if o.AllowMembersUpdate {
		menu = append(menu, updateMembersLink(dao.ID()))
	}

	if parentDAO := dao.Parent(); parentDAO != nil {
		menu = append(menu, ufmt.Sprintf("[Go to Parent DAO](%s)", daoURL(parentDAO)))
	}

	res.Write(strings.Join(menu, " • "))
	res.Write("\n\n---\n")

	// Render members
	members := dao.Members()
	if members.Size() == 0 {
		res.Write("**⚠ The DAO has no members**\n\n")
	} else {
		renderMembers(res, req.RawPath, members)
	}

	// Render organization tree
	if dao.Children().Len() > 0 {
		res.Write("## Tree\n")
		renderTree(res, dao, "")
	}

	// Render latest proposals
	proposals := dao.ActiveProposals()
	if proposals.Size() == 0 {
		return
	}

	res.Write("## Latest Proposals\n")
	proposals.Iterate(0, 3, true, func(p *commondao.Proposal) bool {
		renderProposalsListItem(res, dao, p)
		return false
	})
}

func renderMembers(res *mux.ResponseWriter, path string, members commondao.MemberStorage) {
	pages, err := pager.New(path, members.Size(), pager.WithPageQueryParam("members"), pager.WithPageSize(8))
	if err != nil {
		res.Write(err.Error())
		return
	}

	res.Write("| Members |\n| --- |\n")
	members.IterateByOffset(pages.Offset(), pages.PageSize(), func(addr std.Address) bool {
		res.Write("| ")
		res.Write(userLink(addr))
		res.Write(" |\n")
		return false
	})

	if pages.HasPages() {
		res.Write("\n")
		res.Write(pager.Picker(pages))
		res.Write("\n")
	}

	res.Write("\n")
}

func renderTree(res *mux.ResponseWriter, dao *commondao.CommonDAO, indent string) {
	res.Write(ufmt.Sprintf("%s- [%s](%s)\n", indent, dao.Name(), daoURL(dao)))

	indent += "  "
	dao.Children().ForEach(func(_ int, v any) bool {
		if subDAO, ok := v.(*commondao.CommonDAO); ok {
			renderTree(res, subDAO, indent)
		}
		return false
	})
}

func renderSettings(res *mux.ResponseWriter, req *mux.Request) {
	dao := mustGetDAOFromRequest(req)
	o := getOptions(dao.ID())

	// Render header
	res.Write("# ")
	res.Write(dao.Name())
	res.Write(": Settings\n")

	// Render main menu
	res.Write(ufmt.Sprintf("[Go to DAO](%s)", daoURL(dao)))
	res.Write("\n\n---\n")

	// Render options
	res.Write("## Options\n")
	res.Write("| Option | Value |\n| --- | --- |\n")
	res.Write(ufmt.Sprintf("| Allow SubDAOs | %s |\n", strconv.FormatBool(o.AllowChildren)))
	res.Write(ufmt.Sprintf("| Enable Voting | %s |\n", strconv.FormatBool(o.AllowVoting)))
	res.Write(ufmt.Sprintf("| Enable Proposal Execution | %s |\n", strconv.FormatBool(o.AllowExecution)))
	res.Write(ufmt.Sprintf("| Allow Updating Members of this DAO | %s |\n", strconv.FormatBool(o.AllowMembersUpdate)))
}

func renderProposalsList(res *mux.ResponseWriter, req *mux.Request) {
	dao := mustGetDAOFromRequest(req)

	// Render header
	res.Write("# ")
	res.Write(dao.Name())
	res.Write(": Proposals\n")

	// Render main menu
	res.Write(ufmt.Sprintf("[Go to DAO](%s)", daoURL(dao)))
	res.Write("\n\n---\n")

	// Render proposals
	if dao.ActiveProposals().Size() == 0 && dao.FinishedProposals().Size() == 0 {
		res.Write("**⚠ The DAO has no proposals**\n")
		return
	}

	proposals := dao.ActiveProposals()
	renderFinished := req.Query.Has("finished")
	if renderFinished {
		proposals = dao.FinishedProposals()
	}

	pages, err := pager.New(req.RawPath, proposals.Size(), pager.WithPageSize(8))
	if err != nil {
		res.Write(err.Error())
		return
	}

	res.Write("View: ")
	r := realmpath.Parse(req.RawPath)
	if renderFinished {
		r.Query.Del("finished")
		res.Write(ufmt.Sprintf("[active](%s)", r.String()))
	} else {
		r.Query.Add("finished", "")
		res.Write(ufmt.Sprintf("[finished](%s)", r.String()))
	}

	res.Write(" • Sort by: ")

	r = realmpath.Parse(req.RawPath)
	reverseSort := r.Query.Get("order") != "asc"
	if reverseSort {
		r.Query.Set("order", "asc")
		res.Write(ufmt.Sprintf("[oldest](%s)", r.String()))
	} else {
		r.Query.Set("order", "desc")
		res.Write(ufmt.Sprintf("[newest](%s)", r.String()))
	}

	res.Write("\n\n")

	proposals.Iterate(pages.Offset(), pages.PageSize(), reverseSort, func(p *commondao.Proposal) bool {
		renderProposalsListItem(res, dao, p)
		return false
	})

	// Render pager
	if pages.HasPages() {
		res.Write("\n\n---\n")
		res.Write(pager.Picker(pages))
		res.Write("\n")
	}
}

func renderProposalsListItem(res *mux.ResponseWriter, dao *commondao.CommonDAO, p *commondao.Proposal) {
	def := p.Definition()
	record := p.VotingRecord()
	o := getOptions(dao.ID())

	// Render title
	res.Write(ufmt.Sprintf("**[#%d %s](%s)**  \n", p.ID(), def.Title(), proposalURL(dao, p)))

	// Render details
	res.Write(ufmt.Sprintf("Created by %s  \n", userLink(p.Creator())))
	res.Write(ufmt.Sprintf("Voting ends on %s  \n", p.VotingDeadline().UTC().Format(dateFormat)))

	// Render status
	status := []string{
		ufmt.Sprintf("**%d votes**", record.Size()),
	}

	// Render actions
	if o.AllowVoting && isVotingPeriodActive(p) {
		status = append(status, voteLink(dao.ID(), p.ID()))
	}

	if o.AllowExecution && isExecutionAllowed(p) {
		status = append(status, executeLink(dao.ID(), p.ID()))
	}

	res.Write(strings.Join(status, " • "))
	res.Write("\n\n")
}

func renderProposal(res *mux.ResponseWriter, req *mux.Request) {
	dao := mustGetDAOFromRequest(req)
	p := mustGetProposalFromRequest(req, dao)

	// Check that proposal has no issues
	if err := p.Validate(); err != nil {
		res.Write("> ⚠ **ERROR**: ")
		res.Write(err.Error())
		res.Write("\n")
	}

	votingActive := isVotingPeriodActive(p)
	if votingActive {
		res.Write(ufmt.Sprintf("> Voting ends on **%s**\n", p.VotingDeadline().UTC().Format(dateFormat)))
	}

	def := p.Definition()

	// Render header
	res.Write("# #")
	res.Write(strconv.FormatUint(p.ID(), 10))
	res.Write(" ")
	res.Write(def.Title())
	res.Write("\n")

	// Render main menu
	res.Write(ufmt.Sprintf("[Go to DAO](%s)", daoURL(dao)))

	o := getOptions(dao.ID())
	if o.AllowVoting && votingActive {
		res.Write(" • ")
		res.Write(voteLink(dao.ID(), p.ID()))
	}

	if o.AllowExecution && isExecutionAllowed(p) {
		res.Write(" • ")
		res.Write(executeLink(dao.ID(), p.ID()))
	}

	res.Write("\n\n---\n")

	// Render details
	res.Write("## Details\n")
	res.Write(ufmt.Sprintf("- Proposer: %s\n", userLink(p.Creator())))
	res.Write(ufmt.Sprintf("- Submit Time: %s\n", p.CreatedAt().UTC().Format(time.RFC1123)))

	record := p.VotingRecord()
	if p.Status() == commondao.StatusActive {
		passes, _ := def.Tally(record.Readonly(), commondao.NewMemberSet(dao.Members()))
		if passes {
			res.Write("- Expected Outcome: **pass** ☑\n")
		} else {
			res.Write("- Expected Outcome: **fail** ☒\n")
		}
	}

	res.Write(ufmt.Sprintf("- Status: **%s**", string(p.Status())))
	if reason := p.StatusReason(); reason != "" {
		res.Write(ufmt.Sprintf(" • _%s_", reason))
	}
	res.Write("\n")

	// Render proposal body
	if body := def.Body(); body != "" {
		res.Write("## Description\n")
		res.Write(body)
		res.Write("\n\n")
	}

	if record.Size() > 0 {
		// Render voting stats
		res.Write("## Stats\n")
		renderProposalStats(res, record)

		// Render votes
		res.Write("## Votes\n")
		res.Write(ufmt.Sprintf("Total number of votes: **%d**\n", record.Size()))
		renderProposalVotes(res, req.RawPath, dao, p)
	}
}

func renderProposalStats(res *mux.ResponseWriter, record *commondao.VotingRecord) {
	totalCount := float64(record.Size())

	res.Write("| Vote Choice | Percentage of Votes |\n| --- | --- |\n")
	record.IterateVotesCount(func(c commondao.VoteChoice, voteCount int) bool {
		percentage := float64(voteCount*100) / totalCount
		res.Write(
			ufmt.Sprintf("| %s | %s%% |\n", string(c), strconv.FormatFloat(percentage, 'f', 2, 64)),
		)
		return false
	})
}

func renderProposalVotes(res *mux.ResponseWriter, path string, dao *commondao.CommonDAO, p *commondao.Proposal) {
	record := p.VotingRecord()
	pages, err := pager.New(path, record.Size(), pager.WithPageQueryParam("votes"), pager.WithPageSize(5))
	if err != nil {
		res.Write(err.Error())
		return
	}

	res.Write("| User | Vote |\n| --- | --- |\n")
	record.Iterate(pages.Offset(), pages.PageSize(), false, func(v commondao.Vote) bool {
		res.Write("| ")
		res.Write(userLink(v.Address))
		res.Write(" | ")
		res.Write(ufmt.Sprintf("[%s](%s)", string(v.Choice), voteURL(dao, p, v.Address)))

		if v.Reason != "" {
			res.Write(" with a reason")
		}

		res.Write(" |\n")
		return false
	})

	if pages.HasPages() {
		res.Write("\n")
		res.Write(pager.Picker(pages))
		res.Write("\n")
	}

	res.Write("\n")
}

func renderProposalVote(res *mux.ResponseWriter, req *mux.Request) {
	member := std.Address(req.GetVar("address"))
	if !member.IsValid() {
		res.Write("Invalid address")
		return
	}

	dao := mustGetDAOFromRequest(req)
	p := mustGetProposalFromRequest(req, dao)
	v, found := p.VotingRecord().GetVote(member)
	if !found {
		res.Write("Vote not found")
		return
	}

	res.Write(ufmt.Sprintf("# Vote: Proposal #%d\n", p.ID()))
	res.Write(ufmt.Sprintf("[Go to DAO](%s) • ", daoURL(dao)))
	res.Write(ufmt.Sprintf("[Go to Proposal](%s)", proposalURL(dao, p)))
	res.Write("\n\n---\n")

	res.Write("## Details\n")
	res.Write(ufmt.Sprintf("- User: %s\n", userLink(v.Address)))
	res.Write(ufmt.Sprintf("- Vote: %s\n", string(v.Choice)))

	if v.Reason != "" {
		res.Write("## Reason\n")
		res.Write(v.Reason)
	}
}

func mustGetDAOFromRequest(req *mux.Request) *commondao.CommonDAO {
	rawID := req.GetVar("daoID")
	daoID, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil {
		panic("Invalid DAO ID")
	}

	o := getOptions(daoID)
	if o == nil || o.Private {
		panic("Forbidden")
	}

	return mustGetDAO(daoID)
}

func mustGetProposalFromRequest(req *mux.Request, dao *commondao.CommonDAO) *commondao.Proposal {
	rawID := req.GetVar("proposalID")
	proposalID, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil {
		panic("Invalid proposal ID")
	}

	p := dao.GetProposal(proposalID)
	if p == nil {
		panic("Proposal not found")
	}

	return p
}

func voteLink(daoID, proposalID uint64) string {
	return ufmt.Sprintf("[Vote](%s)", txlink.Call(
		"Vote",
		"daoID", strconv.FormatUint(daoID, 10),
		"proposalID", strconv.FormatUint(proposalID, 10),
		"vote", "",
		"reason", "",
	))
}

func executeLink(daoID, proposalID uint64) string {
	return ufmt.Sprintf("[Execute](%s)", txlink.Call(
		"Execute",
		"daoID", strconv.FormatUint(daoID, 10),
		"proposalID", strconv.FormatUint(proposalID, 10),
	))
}

func updateMembersLink(daoID uint64) string {
	return ufmt.Sprintf("[Update Members](%s)", txlink.Call(
		"CreateMembersUpdateProposal",
		"daoID", strconv.FormatUint(daoID, 10),
		"newMembers", "",
		"removeMembers", "",
	))
}

func userLink(addr std.Address) string {
	user := users.ResolveAddress(addr)
	if user != nil {
		return user.RenderLink("")
	}
	return addr.String()
}

func isVotingPeriodActive(p *commondao.Proposal) bool {
	return p.Status() == commondao.StatusActive && time.Now().Before(p.VotingDeadline())
}

func isExecutionAllowed(p *commondao.Proposal) bool {
	return p.Status() == commondao.StatusActive && !time.Now().Before(p.VotingDeadline())
}
