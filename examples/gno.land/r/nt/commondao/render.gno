package commondao

import (
	"errors"
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/mux"
	"gno.land/p/demo/ufmt"
	"gno.land/p/jeronimoalbi/pager"
	"gno.land/p/moul/realmpath"
	"gno.land/p/moul/txlink"
	"gno.land/p/nt/commondao"
)

func Render(path string) string {
	// TODO: Add view to render a single proposal (allow/disallow must be configurable)

	router := mux.NewRouter()
	router.HandleFunc("", renderHome)
	router.HandleFunc("{daoID}", renderDAO)
	router.HandleFunc("{daoID}/proposals", renderDAOProposals)
	// TODO: Implement proposal view
	return router.Render(path)
}

func renderHome(res *mux.ResponseWriter, _ *mux.Request) {
	// TODO: Render a header text for users
	// TODO: Render generic links to vote or execute proposals
	res.Write("# Common DAO\n")
}

func renderDAO(res *mux.ResponseWriter, req *mux.Request) {
	// TODO: Support config option to allow/disallow rendering a DAO

	dao, err := getRequestDAO(req)
	if err != nil {
		res.Write(err.Error())
		return
	}

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
		renderProposalListItem(res, dao, p)
		return false
	})
}

func renderMembers(res *mux.ResponseWriter, path string, members commondao.MemberStorage) {
	p, err := pager.New(path, members.Size(), pager.WithPageQueryParam("members"), pager.WithPageSize(8))
	if err != nil {
		res.Write(err.Error())
		return
	}

	res.Write("| Members |\n| --- |\n")
	members.IterateByOffset(p.Offset(), p.PageSize(), func(addr std.Address) bool {
		res.Write("| ")
		res.Write(addr.String())
		res.Write(" |\n")
		return false
	})

	if p.HasPages() {
		res.Write("\n")
		res.Write(pager.Picker(p))
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

func renderDAOProposals(res *mux.ResponseWriter, req *mux.Request) {
	dao, err := getRequestDAO(req)
	if err != nil {
		res.Write(err.Error())
		return
	}

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

	p, err := pager.New(req.RawPath, proposals.Size(), pager.WithPageSize(8))
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

	proposals.Iterate(p.Offset(), p.PageSize(), reverseSort, func(p *commondao.Proposal) bool {
		renderProposalListItem(res, dao, p)
		return false
	})

	// Render pager
	if p.HasPages() {
		res.Write("\n\n---\n")
		res.Write(pager.Picker(p))
		res.Write("\n")
	}
}

func renderProposalListItem(res *mux.ResponseWriter, dao *commondao.CommonDAO, p *commondao.Proposal) {
	def := p.Definition()
	record := p.VotingRecord()

	// Render title
	res.Write(ufmt.Sprintf("**[#%d %s](%s)**  \n", p.ID(), def.Title(), proposalURL(dao, p)))

	// Render details
	res.Write(ufmt.Sprintf("Created by %s  \n", p.Creator())) // TODO: Add users link
	res.Write(ufmt.Sprintf("Voting ends on %s  \n", p.VotingDeadline().Format("Mon, 02 Jan 2006 03:04pm MST")))

	// Render status
	res.Write(ufmt.Sprintf("**%d votes** • ", record.Size()))

	// Render actions
	res.Write(voteLink(dao.ID(), p.ID()))
	res.Write(" • ")
	res.Write(executeLink(dao.ID(), p.ID()))
	res.Write("\n\n")
}

func getRequestDAO(req *mux.Request) (*commondao.CommonDAO, error) {
	rawID := req.GetVar("daoID")
	daoID, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil {
		return nil, errors.New("Value is an invalid DAO ID")
	}

	var (
		dao  *commondao.CommonDAO
		path = req.Query.Get("path")
	)

	if path != "" {
		dao = getSubDAO(daoID, path)
	} else {
		dao = getDAO(daoID)
	}

	if dao == nil {
		return nil, errors.New("DAO not found")
	}
	return dao, nil
}

func voteLink(daoID, proposalID uint64) string {
	return ufmt.Sprintf("[vote](%s)", txlink.Call(
		"Vote",
		"daoID", strconv.FormatUint(daoID, 10), // TODO: SubDAOs have no ID, only slug
		"proposalID", strconv.FormatUint(proposalID, 10),
		"vote", "",
		"reason", "",
	))
}

func executeLink(daoID, proposalID uint64) string {
	return ufmt.Sprintf("[execute](%s)", txlink.Call(
		"Execute",
		"daoID", strconv.FormatUint(daoID, 10), // TODO: SubDAOs have no ID, only slug
		"proposalID", strconv.FormatUint(proposalID, 10),
	))
}
