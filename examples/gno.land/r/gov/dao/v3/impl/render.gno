package impl

import (
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/avl/pager"
	"gno.land/p/demo/mux"
	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/helplink"
	"gno.land/r/gov/dao"
	"gno.land/r/sys/users"
)

type render struct {
	relativeRealmPath string
	router            *mux.Router
	pssPager          *pager.Pager
}

func NewRender(d *GovDAO) *render {
	ren := &render{
		pssPager: pager.NewPager(d.pss.Tree, 3, true),
	}

	r := mux.NewRouter()

	r.HandleFunc("", func(rw *mux.ResponseWriter, req *mux.Request) {
		rw.Write(ren.renderActiveProposals(req.RawPath, d))
	})

	r.HandleFunc("{pid}", func(rw *mux.ResponseWriter, req *mux.Request) {
		rw.Write(ren.renderProposalPage(req.GetVar("pid"), d))
	})

	r.HandleFunc("{pid}/votes", func(rw *mux.ResponseWriter, req *mux.Request) {
		rw.Write(ren.renderVotesForProposal(req.GetVar("pid"), d))
	})

	ren.router = r

	return ren
}

func (ren *render) Render(pkgPath string, path string) string {
	relativePath, found := strings.CutPrefix(pkgPath, std.ChainDomain())
	if !found {
		panic(ufmt.Sprintf(
			"realm package with unexpected name found: %v in chain domain %v",
			pkgPath, std.ChainDomain()))
	}
	ren.relativeRealmPath = relativePath
	return ren.router.Render(path)
}

func (ren *render) renderActiveProposals(url string, d *GovDAO) string {
	out := "# GovDAO Proposals\n"

	page := ren.pssPager.MustGetPageByPath(url)
	for _, item := range page.Items {
		out += ren.renderProposalListItem(item.Key, d)
		out += "---\n\n"
	}

	out += page.Picker("")

	return out
}

func (ren *render) renderProposalPage(sPid string, d *GovDAO) string {
	pid, err := strconv.Atoi(sPid)
	if err != nil {
		panic(err.Error())
	}
	ps := d.pss.GetStatus(dao.ProposalID(pid))
	p := dao.MustGetProposal(dao.ProposalID(pid))
	out := ufmt.Sprintf("## Prop #%v - %v\n", sPid, p.Title())

	out += p.Description()
	out += "\n\n"

	out += "Author: " + tryResolveAddr(p.Author()) + "\n\n"

	out += ps.String()
	out += "\n"
	out += ufmt.Sprintf("[Detailed voting list](%v:%v/votes)", ren.relativeRealmPath, sPid)
	out += "\n\n---\n\n"

	out += renderActionBar(sPid)

	return out
}

func (ren *render) renderProposalListItem(sPid string, d *GovDAO) string {
	pid, err := strconv.Atoi(sPid)
	if err != nil {
		panic(err.Error())
	}

	ps := d.pss.GetStatus(dao.ProposalID(pid))
	p := dao.MustGetProposal(dao.ProposalID(pid))
	out := ufmt.Sprintf("### [Prop #%v - %v](%v:%v)\n", sPid, p.Title(), ren.relativeRealmPath, sPid)
	out += ufmt.Sprintf("Author: %s\n\n", p.Author())

	out += "Status: " + getPropStatus(ps)
	out += "\n\n"

	out += "Tiers eligible to vote: "
	out += strings.Join(ps.TiersAllowedToVote, ", ")

	out += "\n\n"
	return out
}

func (ren *render) renderVotesForProposal(sPid string, d *GovDAO) string {
	pid, err := strconv.Atoi(sPid)
	if err != nil {
		panic(err.Error())
	}
	ps := d.pss.GetStatus(dao.ProposalID(pid))

	out := ""
	out += ufmt.Sprintf("# Proposal #%v - Vote List\n\n", sPid)
	out += StringifyVotes(ps)

	return out
}

func isPropActive(ps *proposalStatus) bool {
	return !ps.Accepted && !ps.Denied
}

func getPropStatus(ps *proposalStatus) string {
	if ps.Accepted {
		return "ACCEPTED"
	} else if ps.Denied {
		return "REJECTED"
	}
	return "ACTIVE"
}

func renderActionBar(sPid string) string {
	out := "### Actions\n"

	proxy := helplink.Realm("gno.land/r/gov/dao")
	out += proxy.Func("Vote YES", "MustVoteOnProposalSimple", "pid", sPid, "option", "YES") + " | "
	out += proxy.Func("Vote NO", "MustVoteOnProposalSimple", "pid", sPid, "option", "NO") + " | "
	out += proxy.Func("Vote ABSTAIN", "MustVoteOnProposalSimple", "pid", sPid, "option", "ABSTAIN")

	out += "\n\n"
	out += "WARN: Please double check transaction data before voting."
	return out
}

func tryResolveAddr(addr std.Address) string {
	userData := users.ResolveAddress(addr)
	if userData == nil {
		return addr.String()
	}
	return userData.RenderLink("")
}
