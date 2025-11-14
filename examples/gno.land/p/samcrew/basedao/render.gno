package basedao

import (
	"strconv"
	"strings"

	"gno.land/p/nt/mux"
)

// WELL KNOWN PATHS
const (
	HOME_PATH             = ""
	MEMBERS_PATH          = "members"
	PROPOSALS_PATH        = "proposals"
	HOME_NO_PROFILE_PATH  = "noprofile"
	CONFIG_PATH           = "config"
	PROPOSAL_HISTORY_PATH = "history"
	MEMBER_DETAIL_PATH    = "member/{address}"
	PROPOSAL_DETAIL_PATH  = "proposal/{id}"
	ROLE_DETAIL_PATH      = "role/{name}"
	FALLBACK_DISPLAY_NAME = "Anon"
)

func (d *DAOPrivate) initRenderingRouter() {
	if d.RenderRouter == nil {
		d.RenderRouter = mux.NewRouter()
	}
}

func (d *DAOPrivate) InitDefaultRendering() {
	d.RenderRouter.HandleFunc(HOME_PATH, d.MuxHomePage)
	d.RenderRouter.HandleFunc(MEMBERS_PATH, d.MuxMembersPage)
	d.RenderRouter.HandleFunc(PROPOSALS_PATH, d.MuxProposalsPage)
	d.RenderRouter.HandleFunc(HOME_NO_PROFILE_PATH, d.MuxHomePage)
	d.RenderRouter.HandleFunc(CONFIG_PATH, d.MuxConfigPage)
	d.RenderRouter.HandleFunc(PROPOSAL_HISTORY_PATH, d.MuxProposalHistoryPage)
	d.RenderRouter.HandleFunc(MEMBER_DETAIL_PATH, d.MuxMemberDetailPage)
	d.RenderRouter.HandleFunc(PROPOSAL_DETAIL_PATH, d.MuxProposalDetailPage)
	d.RenderRouter.HandleFunc(ROLE_DETAIL_PATH, d.MuxRoleDetailPage)
}

func (d *DAOPrivate) Render(path string) string {
	if d.RenderFn != nil {
		return d.RenderFn(path, d)
	}
	return d.RenderRouter.Render(path)
}

// DEFAULT_HANDLERS

func (d *DAOPrivate) MuxHomePage(res *mux.ResponseWriter, req *mux.Request) {
	res.Write(d.HomePageView(req.RawPath))
}

func (d *DAOPrivate) MuxMembersPage(res *mux.ResponseWriter, req *mux.Request) {
	res.Write(d.MembersPageView(req.RawPath))
}

func (d *DAOPrivate) MuxProposalsPage(res *mux.ResponseWriter, req *mux.Request) {
	res.Write(d.ProposalsPageView(req.RawPath))
}

func (d *DAOPrivate) MuxConfigPage(res *mux.ResponseWriter, req *mux.Request) {
	res.Write(d.ConfigPageView())
}

func (d *DAOPrivate) MuxProposalHistoryPage(res *mux.ResponseWriter, req *mux.Request) {
	res.Write(d.ProposalHistoryPageView(req.RawPath))
}

func (d *DAOPrivate) MuxMemberDetailPage(res *mux.ResponseWriter, req *mux.Request) {
	res.Write(d.MemberDetailPageView(req.GetVar("address")))
}

func (d *DAOPrivate) MuxProposalDetailPage(res *mux.ResponseWriter, req *mux.Request) {
	id, err := strconv.ParseUint(req.GetVar("id"), 10, 64)
	if err != nil {
		panic(err)
	}
	res.Write(d.ProposalDetailPageView(id))
}

func (d *DAOPrivate) MuxRoleDetailPage(res *mux.ResponseWriter, req *mux.Request) {
	res.Write(d.RoleDetailPageView(req.RawPath, req.GetVar("name")))
}

// UTILITIES

func getLinkPath(pkgPath string) string {
	slashIdx := strings.IndexRune(pkgPath, '/')
	if slashIdx != 1 {
		return pkgPath[slashIdx:]
	}
	return ""
}
