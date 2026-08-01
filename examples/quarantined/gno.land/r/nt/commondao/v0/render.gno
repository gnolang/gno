package commondao

import (
	"strconv"
	"strings"
	"time"

	"gno.land/p/jeronimoalbi/pager"
	"gno.land/p/moul/md"
	"gno.land/p/moul/mdtable"
	"gno.land/p/moul/realmpath"
	"gno.land/p/nt/addrset/v0"
	"gno.land/p/nt/commondao/v0"
	"gno.land/p/nt/markdown/sanitize/v0"
	"gno.land/p/nt/mux/v0"
	"gno.land/p/nt/ufmt/v0"

	"gno.land/r/sys/users"
)

const dateFormat = "Mon, 02 Jan 2006 03:04pm MST"

// trustedMarkdownBody marks proposal definitions whose Body returns
// markdown the definition assembled itself (escaping any embedded user
// fields with md.EscapeText). Those bodies are rendered verbatim. Every
// other body — including the execution kind, whose Body is the proposer's
// raw text — is treated as user-supplied and escaped by the renderer. A
// definition from an imported package cannot implement this realm-local
// marker either, so it is escaped too. The render layer is the single escape
// point: definitions return raw text and opt out only when they render their
// own markup.
type trustedMarkdownBody interface {
	isTrustedMarkdownBody()
}

func Render(path string) string {
	router := mux.NewRouter()
	router.HandleFunc("", renderHome)
	router.HandleFunc("{daoID}", renderDAO)
	router.HandleFunc("{daoID}/settings", renderSettings)
	router.HandleFunc("{daoID}/bylaws", renderBylaws)
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
		if isListed(dao.ID()) {
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
		// A dissolved DAO with no live proper ancestor can never move funds
		// again (its own council is gone and no ancestor can claw back), so
		// warn that late deposits to its treasury address are unrecoverable.
		if !hasLiveProperAncestor(dao) {
			res.Write(md.Blockquote("⚠ No live ancestor DAO remains: coins sent to this treasury address can no longer be recovered."))
		}
	}

	// Render header (Charter: purpose is required, description optional)
	res.Write(md.H1(md.EscapeText(dao.Name())))
	res.Write(md.Paragraph(md.Bold("Purpose:") + " " + md.EscapeText(dao.Purpose())))
	if desc := dao.Description(); desc != "" {
		res.Write(md.Paragraph(md.EscapeText(desc)))
	}

	// Render main menu
	menu := []string{
		md.Link("View Proposals", daoProposalsURL(dao.ID())),
		md.Link("View Settings", settingsURL(dao.ID())),
	}

	if set := bylawsView(dao.ID()); set != nil && set.Size() > 0 {
		menu = append(menu, md.Link("View Bylaws & Mandates", bylawsURL(dao.ID())))
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
		renderCreateProposalSection(dao, res)
	}
}

func renderCreateProposalSection(dao *commondao.CommonDAO, res *mux.ResponseWriter) {
	daoID := dao.ID()

	// One entry per link-encodable catalog kind, in display order (the
	// execution kind has no entry: its closure cannot be encoded in a
	// help link). Only the kinds registered on the DAO are rendered.
	entries := []struct {
		kind string
		text string
	}{
		{
			kindText,
			md.Paragraph(textProposalLink(daoID)) +
				md.Paragraph(
					"This type of proposal is also known as text proposal which can be used for example "+
						"to get consensus on initiatives without actually making any change on-chain.",
				),
		},
		{
			kindCouncilUpdate,
			md.Paragraph(updateCouncilLink(daoID)) +
				md.Paragraph(
					"This type of proposal can be used to add new council members to this DAO "+
						"and also to remove existing ones.",
				) +
				md.Paragraph(
					"A single proposal allows new council members to be added and any number of existing ones "+
						"removed within the same proposal.",
				),
		},
		{
			kindAncestorCouncilUpdate,
			md.Paragraph(ancestorCouncilUpdateLink(daoID)) +
				md.Paragraph(
					"This type of proposal lets this DAO, as an ancestor, add or "+
						"remove council members of one of its descendant DAOs — the "+
						"rescue path for a descendant whose council is stuck or empty.",
				),
		},
		{
			kindDissolve,
			md.Paragraph(dissolveSubDAOLink(daoID)) +
				md.Paragraph(
					"This type of proposal can be used to dissolve DAOs and SubDAOs.",
				) +
				md.Paragraph(
					"Dissolving a DAO can't be undone, once the dissolution proposal passes "+
						"and is executed DAO will be readonly.",
				),
		},
		{
			kindTreasurySpend,
			md.Paragraph(treasurySpendLink(daoID)) +
				md.Paragraph(
					"This type of proposal sends coins from the DAO's own treasury "+
						"when it passes and is executed.",
				),
		},
		{
			kindTreasuryClawback,
			md.Paragraph(treasuryClawbackLink(daoID)) +
				md.Paragraph(
					"This type of proposal sweeps a descendant DAO's treasury to "+
						"the descendant's parent DAO.",
				),
		},
		{
			kindTreasuryFreeze,
			md.Paragraph(treasuryFreezeLink(daoID)) +
				md.Paragraph(
					"This type of proposal freezes or unfreezes a descendant DAO's "+
						"treasury. While frozen, no funds can leave the treasury.",
				),
		},
		{
			kindSubDAO,
			md.Paragraph(newSubDAOLink(daoID)) +
				md.Paragraph(
					"This type of proposal is used to create SubDAOs, "+
						"which are used to create tree based DAOs.",
				),
		},
		{
			kindManageKinds,
			md.Paragraph(registerKindLink(daoID)+" • "+deregisterKindLink(daoID)) +
				md.Paragraph(
					"This type of proposal registers one of the realm's catalog "+
						"proposal kinds on this DAO by name, or deregisters a kind. "+
						"Deregistering a kind blocks new proposals of that kind while "+
						"in-flight ones still vote and execute. The manage-kinds kind "+
						"itself cannot be deregistered.",
				),
		},
		{
			kindAmendBylaws,
			md.Paragraph(amendBylawsLink(daoID)) +
				md.Paragraph(
					"This type of proposal adds, amends or removes one of the DAO's "+
						"bylaws and mandates documents with a diff patch pinned to the "+
						"current document text. Build the payload with the "+
						"AmendBylawsPayload query function.",
				),
		},
	}

	var cols []string
	for _, e := range entries {
		if dao.HasKind(e.kind) {
			cols = append(cols, e.text)
		}
	}

	if len(cols) == 0 {
		return
	}

	res.Write(md.H2("Create Proposal"))
	res.Write(md.Paragraph("These are the proposal supported by this DAO:"))
	res.Write(md.Columns(cols, false))
}

func renderTreasury(res *mux.ResponseWriter, dao *commondao.CommonDAO) {
	res.Write(md.H2("Treasury"))
	res.Write(md.BulletItem("Address: " + md.EscapeText(string(dao.Address()))))

	balance := treasuryBalance(dao)
	if balance.IsZero() {
		res.Write(md.BulletItem("Balance: (empty)"))
	} else {
		res.Write(md.BulletItem("Balance: " + md.EscapeText(balance.String())))
	}

	if dao.IsTreasuryFrozen() {
		res.Write(md.Paragraph(md.Bold("⚠ The treasury is frozen: no funds can leave until a proper ancestor DAO unfreezes it")))
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

	// Render header
	res.Write(md.H1(md.EscapeText(dao.Name()) + ": Settings"))

	// Render main menu
	res.Write(md.Paragraph(goToDAOLink(dao.ID())))
	res.Write(md.HorizontalRule())

	// Render info
	table := mdtable.Table{Headers: []string{"Setting", "Value"}}
	table.Append([]string{"Listed", strconv.FormatBool(isListed(dao.ID()))})
	table.Append([]string{"Max active proposals", strconv.Itoa(dao.MaxActiveProposals())})
	table.Append([]string{"Proposal kinds", md.EscapeText(strings.Join(dao.KindNames(), ", "))})

	res.Write(md.H2("Info"))
	res.Write(table.String())
}

func renderBylaws(res *mux.ResponseWriter, req *mux.Request) {
	dao := mustGetDAOFromRequest(req)

	// Render header
	res.Write(md.H1(md.EscapeText(dao.Name()) + ": Bylaws & Mandates"))

	// Render main menu
	res.Write(md.Paragraph(goToDAOLink(dao.ID())))
	res.Write(md.HorizontalRule())

	set := bylawsView(dao.ID())
	if set == nil || set.Size() == 0 {
		res.Write(md.Paragraph("This DAO has no bylaws or mandates documents."))
		return
	}

	paths := set.List("")
	pages, err := pager.New(req.RawPath, len(paths), pager.WithPageSize(5))
	if err != nil {
		res.Write(err.Error())
		return
	}

	// Documents are multi-line plaintext with council-controlled content:
	// the path is escaped inline and the text goes through sanitize.Block,
	// which preserves paragraph structure (an inline escape would fold the
	// whole document to one line) while escaping block-level hazards. The
	// hash is what an amendment payload must pin (AmendBylawsPayload reads
	// it for you).
	for i := pages.Offset(); i < len(paths) && i < pages.Offset()+pages.PageSize(); i++ {
		path := paths[i]
		text, _ := set.Get(path)
		res.Write(md.H2(md.EscapeText(path)))
		res.Write(md.Paragraph("sha256: " + set.Hash(path)))
		res.Write(sanitize.Block(text))
	}

	if pages.HasPages() {
		res.Write(md.Paragraph(pager.Picker(pages)))
	}
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

	if isExecutionAllowed(p) {
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
	if votingActive {
		items = append(items, voteLink(dao.ID(), p.ID()))
	}

	if isExecutionAllowed(p) {
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

	// Render proposal body. Bodies are raw user text and escaped here by
	// default; only definitions that assemble their own markdown (marked
	// trustedMarkdownBody) are rendered verbatim.
	if body := def.Body(); body != "" {
		res.Write(md.H2("Description"))
		if _, trusted := def.(trustedMarkdownBody); !trusted {
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
		// A changed vote leaves a zeroed counter behind; skip it so the
		// table shows only choices with live votes.
		if voteCount == 0 {
			return false
		}

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
		voteDetails := md.Link(string(v.Choice()), voteURL(dao.ID(), p.ID(), v.Address()))
		if v.Reason() != "" {
			voteDetails += " with a reason"
		}

		table.Append([]string{userLink(v.Address()), voteDetails})
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
	res.Write(md.BulletItem("User: " + userLink(v.Address())))
	res.Write(md.BulletItem("Vote: " + md.EscapeText(string(v.Choice()))))

	if v.Reason() != "" {
		res.Write(md.H2("Reason"))
		res.Write(md.EscapeText(v.Reason()))
	}
}

func mustGetDAOFromRequest(req *mux.Request) *commondao.CommonDAO {
	rawID := req.GetVar("daoID")
	daoID, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil {
		panic("invalid DAO ID")
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

func ancestorCouncilUpdateLink(daoID uint64) string {
	return md.Link("Ancestor Council Update", realmLink.Call(
		"CreateAncestorCouncilUpdateProposal",
		"daoID", strconv.FormatUint(daoID, 10),
		"targetID", "",
		"newMembers", "",
		"removeMembers", "",
	))
}

func newSubDAOLink(daoID uint64) string {
	return md.Link("New SubDAO", realmLink.Call(
		"CreateSubDAOProposal",
		"daoID", strconv.FormatUint(daoID, 10),
		"name", "",
		"purpose", "",
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

func registerKindLink(daoID uint64) string {
	return md.Link("Register Proposal Kind", realmLink.Call(
		"CreateRegisterKindProposal",
		"daoID", strconv.FormatUint(daoID, 10),
		"kindName", "",
	))
}

func deregisterKindLink(daoID uint64) string {
	return md.Link("Deregister Proposal Kind", realmLink.Call(
		"CreateDeregisterKindProposal",
		"daoID", strconv.FormatUint(daoID, 10),
		"kindName", "",
	))
}

func amendBylawsLink(daoID uint64) string {
	return md.Link("Amend Bylaws", realmLink.Call(
		"CreateAmendBylawsProposal",
		"daoID", strconv.FormatUint(daoID, 10),
		"payload", "",
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
