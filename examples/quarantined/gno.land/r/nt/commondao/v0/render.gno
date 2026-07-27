package commondao

import (
	"strconv"
	"strings"
	"time"

	"gno.land/p/jeronimoalbi/pager"
	"gno.land/p/moul/addrset"
	"gno.land/p/moul/md"
	"gno.land/p/moul/mdtable"
	"gno.land/p/moul/realmpath"
	"gno.land/p/nt/commondao/v0"
	"gno.land/p/nt/mux/v0"
	"gno.land/p/nt/ufmt/v0"

	"gno.land/r/sys/users"
)

const dateFormat = "Mon, 02 Jan 2006 03:04pm MST"

// rawTextDefinition marks proposal definitions whose Body returns raw,
// user-supplied text that the renderer must escape. Definitions that
// assemble their own markdown escape any embedded user fields themselves
// and are rendered as trusted markdown.
type rawTextDefinition interface {
	isRawTextDefinition()
}

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

func renderHome(res *mux.ResponseWriter, req *mux.Request) {
	res.Write(md.H1("Common DAO"))
	res.Write(md.HorizontalRule())
	res.Write(ufmt.Sprintf(
		md.Paragraph("This realm can be used to create CommonDAO instances based on %s package."),
		md.Link("commondao", "/p/nt/commondao/v0/"),
	))

	pages, err := pager.New(req.RawPath, daos.Size(), pager.WithPageSize(10))
	if err != nil {
		res.Write(err.Error())
		return
	}

	var items []string
	daos.ReverseIterateByOffset(pages.Offset(), pages.PageSize(), func(_ string, v any) bool {
		dao := v.(*commondao.CommonDAO)
		o := getOptions(dao.ID())
		if o.AllowRender && o.AllowListing {
			items = append(items, md.Link(dao.Name(), daoURL(dao.ID())))
		}
		return false
	})

	if len(items) == 0 {
		return
	}

	res.Write(md.Paragraph("Here is a list of some of the DAOs that were created:"))
	res.Write(md.Paragraph(md.BulletList(items)))

	if pages.HasPages() {
		res.Write(md.Paragraph(pager.Picker(pages)))
	}
}

func renderDAO(res *mux.ResponseWriter, req *mux.Request) {
	dao := mustGetDAOFromRequest(req)

	// Render header messages
	if dao.IsDeleted() {
		res.Write(md.Blockquote("⚠ This DAO has been dissolved"))
	}

	// Render header
	res.Write(md.H1(md.EscapeText(dao.Name())))
	if desc := dao.Description(); desc != "" {
		res.Write(md.Paragraph(md.EscapeText(desc)))
	}

	// Render main menu
	menu := []string{
		md.Link("View Proposals", daoProposalsURL(dao.ID())),
		md.Link("View Settings", settingsURL(dao.ID())),
	}

	if parentDAO := dao.Parent(); parentDAO != nil {
		menu = append(menu, md.Link("Go to Parent DAO", daoURL(parentDAO.ID())))
	}

	res.Write(md.Paragraph(strings.Join(menu, " • ")))
	res.Write(md.HorizontalRule())

	// Render council
	council := dao.Council()
	if council.Size() == 0 {
		res.Write(md.Paragraph(md.Bold("⚠ The DAO has no council members")))
	} else {
		renderCouncil(res, req.RawPath, council)
	}

	renderTreasury(res, dao)

	// Render organization tree
	if dao.ChildrenCount() > 0 {
		r := parseRealmPath(req.RawPath)
		dissolvedVisible := r.Query.Has("dissolved") || dao.IsDeleted()
		res.Write(md.H2("Tree"))

		// Render toggle only when DAO is not dissolved,
		// otherwise render the whole tree when DAO is dissolved
		if !dao.IsDeleted() {
			var toggleLink string
			if dissolvedVisible {
				r.Query.Del("dissolved")
				toggleLink = md.Link("hide", r.String())
			} else {
				r.Query.Add("dissolved", "")
				toggleLink = md.Link("show", r.String())
			}

			res.Write(md.Paragraph("Dissolved: " + toggleLink))
		}

		renderTree(res, dao, "", dissolvedVisible)
	}

	// Render latest proposals
	if dao.ActiveProposalsSize() > 0 {
		res.Write(md.H2("Latest Proposals"))
		dao.IterateActiveProposals(0, 3, true, func(p *commondao.Proposal) bool {
			renderProposalsListItem(res, dao, p)
			return false
		})
	}

	// Render proposal creation links
	if !dao.IsDeleted() {
		renderCreateProposalSection(dao.ID(), res)
	}
}

func renderCreateProposalSection(daoID uint64, res *mux.ResponseWriter) {
	var (
		cols []string
		o    = getOptions(daoID)
	)

	{
		cols = append(cols, md.Paragraph(textProposalLink(daoID))+
			md.Paragraph(
				"This type of proposal is also known as text proposal which can be used for example "+
					"to get consensus on initiatives without actually making any change on-chain.",
			),
		)
	}

	if o.AllowChildren {
		cols = append(cols, md.Paragraph(newSubDAOLink(daoID))+
			md.Paragraph(
				"This type of proposal is used to create SubDAOs, "+
					"which are used to create tree based DAOs.",
			),
		)
	}

	{
		cols = append(cols, md.Paragraph(dissolveSubDAOLink(daoID))+
			md.Paragraph(
				"This type of proposal can be used to dissolve DAOs and SubDAOs.",
			)+
			md.Paragraph(
				"Dissolving a DAO can't be undone, once the dissolution proposal passes "+
					"and is executed DAO will be readonly.",
			),
		)
	}

	{
		cols = append(cols, md.Paragraph(treasurySpendLink(daoID))+
			md.Paragraph(
				"This type of proposal sends coins from the DAO's own treasury "+
					"when it passes and is executed.",
			),
		)

		cols = append(cols, md.Paragraph(treasuryClawbackLink(daoID))+
			md.Paragraph(
				"This type of proposal sweeps a descendant DAO's treasury to "+
					"the descendant's parent DAO.",
			),
		)

		cols = append(cols, md.Paragraph(treasuryFreezeLink(daoID))+
			md.Paragraph(
				"This type of proposal freezes or unfreezes a descendant DAO's "+
					"treasury. While frozen, no treasury spend can execute.",
			),
		)
	}

	{
		cols = append(cols, md.Paragraph(updateCouncilLink(daoID))+
			md.Paragraph(
				"This type of proposal can be used to add new council members to this DAO "+
					"and also to remove existing ones.",
			)+
			md.Paragraph(
				"A single proposal allows new council members to be added and any number of existing ones "+
					"removed within the same proposal.",
			),
		)
	}

	res.Write(md.H2("Create Proposal"))
	res.Write(md.Paragraph("These are the proposal supported by this DAO:"))
	res.Write(md.Columns(cols, false))
}

func renderTreasury(res *mux.ResponseWriter, dao *commondao.CommonDAO) {
	res.Write(md.H2("Treasury"))
	res.Write(md.BulletItem("Address: " + string(dao.Address())))

	balance := treasuryBalance(dao)
	if balance.IsZero() {
		res.Write(md.BulletItem("Balance: (empty)"))
	} else {
		res.Write(md.BulletItem("Balance: " + balance.String()))
	}

	if dao.IsTreasuryFrozen() {
		res.Write(md.Paragraph(md.Bold("⚠ The treasury is frozen: spends cannot execute until a proper ancestor DAO unfreezes it")))
	}
}

func renderCouncil(res *mux.ResponseWriter, path string, council *addrset.ReadonlySet) {
	pages, err := pager.New(path, council.Size(), pager.WithPageQueryParam("members"), pager.WithPageSize(8))
	if err != nil {
		res.Write(err.Error())
		return
	}

	table := mdtable.Table{Headers: []string{"Council"}}
	council.IterateByOffset(pages.Offset(), pages.PageSize(), func(addr address) bool {
		table.Append([]string{userLink(addr)})
		return false
	})

	res.Write(md.Paragraph(table.String()))

	if pages.HasPages() {
		res.Write(md.Paragraph(pager.Picker(pages)))
	}
}

func renderTree(res *mux.ResponseWriter, dao *commondao.CommonDAO, indent string, showDissolved bool) {
	daoLink := md.Link(dao.Name(), daoURL(dao.ID()))
	if dao.IsDeleted() {
		// Strikethough dissolved DAO names
		daoLink = md.Strikethrough(daoLink)
	}

	res.Write(indent + md.BulletItem(daoLink))

	indent += "  "
	dao.IterateChildren(func(subDAO *commondao.CommonDAO) bool {
		if showDissolved || !subDAO.IsDeleted() {
			renderTree(res, subDAO, indent, showDissolved)
		}
		return false
	})
}

func renderSettings(res *mux.ResponseWriter, req *mux.Request) {
	dao := mustGetDAOFromRequest(req)
	o := getOptions(dao.ID())

	// Render header
	res.Write(md.H1(md.EscapeText(dao.Name()) + ": Settings"))

	// Render main menu
	res.Write(md.Paragraph(goToDAOLink(dao.ID())))
	res.Write(md.HorizontalRule())

	// Render options
	table := mdtable.Table{Headers: []string{"Options", "Values"}}
	table.Append([]string{"Allow listing", strconv.FormatBool(o.AllowListing)})
	table.Append([]string{"Allow render", strconv.FormatBool(o.AllowRender)})
	table.Append([]string{"Allow SubDAOs", strconv.FormatBool(o.AllowChildren)})
	table.Append([]string{"Enable early execution", strconv.FormatBool(o.AllowExecution)})
	table.Append([]string{"Max active proposals", strconv.Itoa(dao.MaxActiveProposals())})

	res.Write(md.H2("Options"))
	res.Write(table.String())
}

func renderProposalsList(res *mux.ResponseWriter, req *mux.Request) {
	dao := mustGetDAOFromRequest(req)

	// Render header
	res.Write(md.H1(md.EscapeText(dao.Name()) + ": Proposals"))

	// Render main menu
	res.Write(md.Paragraph(goToDAOLink(dao.ID())))
	res.Write(md.HorizontalRule())

	// Render proposals
	if dao.ActiveProposalsSize() == 0 && dao.FinishedProposalsSize() == 0 {
		res.Write(md.Paragraph(md.Bold("⚠ The DAO has no proposals")))
		return
	}

	size := dao.ActiveProposalsSize()
	iterate := dao.IterateActiveProposals
	renderFinished := req.Query.Has("finished")
	if renderFinished {
		size = dao.FinishedProposalsSize()
		iterate = dao.IterateFinishedProposals
	}

	pages, err := pager.New(req.RawPath, size, pager.WithPageSize(8))
	if err != nil {
		res.Write(err.Error())
		return
	}

	var viewLink, sortLink string

	r := parseRealmPath(req.RawPath)
	if renderFinished {
		r.Query.Del("finished")
		viewLink = md.Link("active", r.String())
	} else {
		r.Query.Add("finished", "")
		viewLink = md.Link("finished", r.String())
	}

	r = parseRealmPath(req.RawPath)
	reverseSort := r.Query.Get("order") != "asc"
	if reverseSort {
		r.Query.Set("order", "asc")
		sortLink = md.Link("oldest", r.String())
	} else {
		r.Query.Set("order", "desc")
		sortLink = md.Link("newest", r.String())
	}

	res.Write(md.Paragraph("View: " + viewLink + " • Sort by: " + sortLink))

	if size == 0 {
		if renderFinished {
			res.Write(md.Paragraph("Currently there are no finished proposals"))
		} else {
			res.Write(md.Paragraph("Currently there are no active proposals"))
		}
	} else {
		iterate(pages.Offset(), pages.PageSize(), reverseSort, func(p *commondao.Proposal) bool {
			renderProposalsListItem(res, dao, p)
			return false
		})
	}

	// Render pager
	if pages.HasPages() {
		res.Write(md.HorizontalRule())
		res.Write(pager.Picker(pages))
	}
}

func renderProposalsListItem(res *mux.ResponseWriter, dao *commondao.CommonDAO, p *commondao.Proposal) {
	def := p.Definition()
	record := p.VotingRecord()
	o := getOptions(dao.ID())

	// Render title
	res.Write(ufmt.Sprintf("**[#%d %s](%s)**  \n", p.ID(), md.EscapeText(def.Title()), proposalURL(dao.ID(), p.ID())))

	// Render details
	res.Write(ufmt.Sprintf("Created by %s  \n", userLink(p.Creator())))
	res.Write(ufmt.Sprintf("Voting ends on %s  \n", p.VotingDeadline().UTC().Format(dateFormat)))

	// Render status
	status := []string{
		ufmt.Sprintf("Votes: **%d**", record.Size()),
		ufmt.Sprintf("Status: **%s**", string(p.Status())),
	}

	// Render actions
	if isVotingPeriodActive(p) {
		status = append(status, voteLink(dao.ID(), p.ID()))
	}

	if isExecutionAllowed(p) && (o.AllowExecution || p.HasVotingDeadlinePassed()) {
		status = append(status, executeLink(dao.ID(), p.ID()))
	}

	res.Write(md.Paragraph(strings.Join(status, " • ")))
}

func renderProposal(res *mux.ResponseWriter, req *mux.Request) {
	dao := mustGetDAOFromRequest(req)
	p := mustGetProposalFromRequest(req, dao)

	// Check that proposal has no issues
	if err := p.Validate(); err != nil {
		res.Write(md.Blockquote("⚠ **ERROR**: " + err.Error()))
	}

	votingActive := isVotingPeriodActive(p)
	if votingActive {
		res.Write(
			md.Blockquote("Voting ends on " + md.Bold(p.VotingDeadline().UTC().Format(dateFormat))),
		)
	}

	def := p.Definition()

	// Render header
	res.Write(md.H1("#" + strconv.FormatUint(p.ID(), 10) + " " + md.EscapeText(def.Title())))

	// Render main menu
	items := []string{goToDAOLink(dao.ID())}
	o := getOptions(dao.ID())
	if votingActive {
		items = append(items, voteLink(dao.ID(), p.ID()))
	}

	if isExecutionAllowed(p) && (o.AllowExecution || p.HasVotingDeadlinePassed()) {
		items = append(items, executeLink(dao.ID(), p.ID()))
	}

	res.Write(md.Paragraph(strings.Join(items, " • ")))
	res.Write(md.HorizontalRule())

	// Render details
	res.Write(md.H2("Details"))
	res.Write(md.BulletItem("Proposer: " + userLink(p.Creator())))
	res.Write(md.BulletItem("Submit Time: " + p.CreatedAt().UTC().Format(time.RFC1123)))

	record := p.VotingRecord()
	if p.Status() == commondao.StatusActive {
		if p.ExpectedOutcome() == commondao.OutcomePassed {
			res.Write(md.BulletItem("Expected Outcome: **pass** ☑"))
		} else {
			res.Write(md.BulletItem("Expected Outcome: **fail** ☒"))
		}
	}

	statusItem := "Status: " + md.Bold(string(p.Status()))
	if reason := p.StatusReason(); reason != "" {
		statusItem += " • " + md.Italic(md.EscapeText(reason))
	}
	res.Write(md.BulletItem(statusItem))

	// Render proposal body. Definitions that build their own markdown
	// escape their embedded user fields themselves; only raw-text bodies
	// are escaped here.
	if body := def.Body(); body != "" {
		res.Write(md.H2("Description"))
		if _, raw := def.(rawTextDefinition); raw {
			body = md.EscapeText(body)
		}
		res.Write(md.Paragraph(body))
	}

	// Render voting stats and votes
	if record.Size() > 0 {
		renderProposalStats(res, record)
		renderProposalVotes(res, req.RawPath, dao, p)
	}
}

func renderProposalStats(res *mux.ResponseWriter, record commondao.ReadonlyVotingRecord) {
	totalCount := float64(record.Size())
	table := mdtable.Table{Headers: []string{"Vote Choices", "Percentage of Votes"}}

	record.IterateVotesCount(func(c commondao.VoteChoice, voteCount int) bool {
		percentage := float64(voteCount*100) / totalCount

		table.Append([]string{md.EscapeText(string(c)), strconv.FormatFloat(percentage, 'f', 2, 64) + "%"})
		return false
	})

	res.Write(md.H2("Stats"))
	res.Write(md.Paragraph(table.String()))
}

func renderProposalVotes(res *mux.ResponseWriter, path string, dao *commondao.CommonDAO, p *commondao.Proposal) {
	res.Write(md.H2("Votes")) // Render title here so it appears before any pager errors

	record := p.VotingRecord()
	pages, err := pager.New(path, record.Size(), pager.WithPageQueryParam("votes"), pager.WithPageSize(5))
	if err != nil {
		res.Write(err.Error())
		return
	}

	table := mdtable.Table{Headers: []string{"Users", "Votes"}}
	record.Iterate(pages.Offset(), pages.PageSize(), false, func(v commondao.Vote) bool {
		voteDetails := md.Link(string(v.Choice), voteURL(dao.ID(), p.ID(), v.Address))
		if v.Reason != "" {
			voteDetails += " with a reason"
		}

		table.Append([]string{userLink(v.Address), voteDetails})
		return false
	})

	res.Write(ufmt.Sprintf("Total number of votes: **%d**\n", record.Size()))
	res.Write(md.Paragraph(table.String()))

	if pages.HasPages() {
		res.Write(md.Paragraph(pager.Picker(pages)))
	}
}

func renderProposalVote(res *mux.ResponseWriter, req *mux.Request) {
	member := address(req.GetVar("address"))
	if !member.IsValid() {
		res.Write("invalid address")
		return
	}

	dao := mustGetDAOFromRequest(req)
	p := mustGetProposalFromRequest(req, dao)
	v, found := p.VotingRecord().GetVote(member)
	if !found {
		res.Write("vote not found")
		return
	}

	links := []string{
		goToDAOLink(dao.ID()),
		goToProposalLink(dao.ID(), p.ID()),
	}

	res.Write(ufmt.Sprintf("# Vote: Proposal #%d\n", p.ID()))
	res.Write(md.Paragraph(strings.Join(links, " • ")))
	res.Write(md.HorizontalRule())

	res.Write(md.H2("Details"))
	res.Write(md.BulletItem("User: " + userLink(v.Address)))
	res.Write(md.BulletItem("Vote: " + md.EscapeText(string(v.Choice))))

	if v.Reason != "" {
		res.Write(md.H2("Reason"))
		res.Write(md.EscapeText(v.Reason))
	}
}

func mustGetDAOFromRequest(req *mux.Request) *commondao.CommonDAO {
	rawID := req.GetVar("daoID")
	daoID, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil {
		panic("invalid DAO ID")
	}

	o := getOptions(daoID)
	if o == nil || !o.AllowRender {
		panic("rendering is not enabled")
	}

	return mustGetDAO(daoID)
}

func mustGetProposalFromRequest(req *mux.Request, dao *commondao.CommonDAO) *commondao.Proposal {
	rawID := req.GetVar("proposalID")
	proposalID, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil {
		panic("invalid proposal ID")
	}

	p := dao.GetProposal(proposalID)
	if p == nil {
		panic("proposal not found")
	}
	return p
}

func parseRealmPath(path string) *realmpath.Request {
	r := realmpath.Parse(path)
	r.Realm = string(realmLink)
	return r
}

func voteLink(daoID, proposalID uint64) string {
	return md.Link("Vote", realmLink.Call(
		"Vote",
		"daoID", strconv.FormatUint(daoID, 10),
		"proposalID", strconv.FormatUint(proposalID, 10),
		"vote", "",
		"reason", "",
	))
}

func executeLink(daoID, proposalID uint64) string {
	return md.Link("Execute", realmLink.Call(
		"Execute",
		"daoID", strconv.FormatUint(daoID, 10),
		"proposalID", strconv.FormatUint(proposalID, 10),
	))
}

func textProposalLink(daoID uint64) string {
	return ufmt.Sprintf("[General Proposal](%s)", realmLink.Call(
		"CreateTextProposal",
		"daoID", strconv.FormatUint(daoID, 10),
		"title", "",
		"body", "",
		"votingDays", "7",
	))
}

func updateCouncilLink(daoID uint64) string {
	return md.Link("Update Council", realmLink.Call(
		"CreateCouncilUpdateProposal",
		"daoID", strconv.FormatUint(daoID, 10),
		"newMembers", "",
		"removeMembers", "",
	))
}

func newSubDAOLink(daoID uint64) string {
	return md.Link("New SubDAO", realmLink.Call(
		"CreateSubDAOProposal",
		"daoID", strconv.FormatUint(daoID, 10),
		"name", "",
		"description", "",
		"members", "",
	))
}

func dissolveSubDAOLink(daoID uint64) string {
	return md.Link("Dissolve DAO", realmLink.Call(
		"CreateDissolutionProposal",
		"daoID", strconv.FormatUint(daoID, 10),
		"destination", "",
	))
}

func treasuryClawbackLink(daoID uint64) string {
	return md.Link("Treasury Clawback", realmLink.Call(
		"CreateTreasuryClawbackProposal",
		"daoID", strconv.FormatUint(daoID, 10),
		"targetID", "",
	))
}

func treasuryFreezeLink(daoID uint64) string {
	return md.Link("Treasury Freeze", realmLink.Call(
		"CreateTreasuryFreezeProposal",
		"daoID", strconv.FormatUint(daoID, 10),
		"targetID", "",
		"frozen", "true",
	))
}

func treasurySpendLink(daoID uint64) string {
	return md.Link("Treasury Spend", realmLink.Call(
		"CreateTreasurySpendProposal",
		"daoID", strconv.FormatUint(daoID, 10),
		"to", "",
		"denom", "ugnot",
		"amount", "",
	))
}

func goToDAOLink(daoID uint64) string {
	return md.Link("Go to DAO", daoURL(daoID))
}

func goToProposalLink(daoID, proposalID uint64) string {
	return md.Link("Go to Proposal", proposalURL(daoID, proposalID))
}

func userLink(addr address) string {
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
	// Early passed proposals can be executed right away; active proposals
	// can be finalized once their voting deadline passes.
	if p.Status() == commondao.StatusPassed {
		return true
	}
	return p.Status() == commondao.StatusActive && !time.Now().Before(p.VotingDeadline())
}
