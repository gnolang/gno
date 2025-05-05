package impl

import (
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/avl/pager"
	"gno.land/p/demo/mux"
	"gno.land/p/demo/ufmt"
	"gno.land/r/gov/dao"
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

	r.HandleFunc("/", func(rw *mux.ResponseWriter, req *mux.Request) {
		rw.Write(ren.renderActiveProposals(req.RawPath, d))
	})

	r.HandleFunc("", func(rw *mux.ResponseWriter, req *mux.Request) {
		rw.Write(ren.renderActiveProposals(req.RawPath, d))
	})

	r.HandleFunc("{pid}", func(rw *mux.ResponseWriter, req *mux.Request) {
		rw.Write(ren.renderProposal(req.GetVar("pid"), d))
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
	out := "# Active Proposals:\n"

	page, err := ren.pssPager.GetPageByPath(url)
	if err != nil {
		out += ufmt.Sprintf("Error getting selected page: %v", err.Error())
		return out
	}

	for _, item := range page.Items {
		out += ren.renderProposal(item.Key, d)
	}

	out += page.Picker("")

	return out
}

func (ren *render) renderProposal(sPid string, d *GovDAO) string {
	pid, err := strconv.Atoi(sPid)
	if err != nil {
		panic(err.Error())
	}
	ps := d.pss.GetStatus(dao.ProposalID(pid))

	p := cross(dao.MustGetProposal)(dao.ProposalID(pid))

	out := ""
	out += ufmt.Sprintf("## Proposal with id: %v", sPid)
	out += StringifyProposal(p)
	out += "\n\n"
	out += ps.String()
	out += ufmt.Sprintf("- [Go to votes list](%v:%v/votes).", ren.relativeRealmPath, sPid)
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
	out += ufmt.Sprintf("## Voters for Proposal with id: %v\n\n", sPid)
	out += StringifyVotes(ps)

	return out
}
